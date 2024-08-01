// Package sharing provides the functions to generate and validate sharing codes.
package sharing

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var testJwtKey = []byte("test_jwt_key")

func TestGenerateShareToken(t *testing.T) {
	checklistID := "test_checklist_id"
	userID := "test_user_id"
	token, err := generateSharingToken(checklistID, userID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatalf("Generated token is empty")
	}
}

func TestParseShareToken(t *testing.T) {
	checklistID := "test_checklist_id"
	userID := "test_user_id"
	token, err := generateSharingToken(checklistID, userID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ParseSharingToken(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if claims.ChecklistID != checklistID {
		t.Fatalf("Expected checklistID %v, but got %v", checklistID, claims.ChecklistID)
	}
}

func TestParseInvalidToken(t *testing.T) {
	invalidToken := "invalid_token"

	_, err := ParseSharingToken(invalidToken)
	if err == nil {
		t.Fatalf("Expected an error for invalid token, but got nil")
	}
}

func TestTokenExpiration(t *testing.T) {
	// Generate a token with a short expiration time
	expirationTime := time.Now().Add(1 * time.Second)
	claims := &Claims{
		ChecklistID: "test_checklist_id",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(testJwtKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for the token to expire
	time.Sleep(2 * time.Second)

	_, err = ParseSharingToken(tokenString)
	if err == nil || err.Error() != "token is expired" {
		t.Fatalf("Expected 'token is expired' error, but got: %v", err)
	}
}

func TestGetShareCode(t *testing.T) {
	// GetShortCode provides a truncated sha256 hash 8 chars long that maps to a token in redis
	userID := "some-user-id"
	checklistID := "some-checklist-id"
	hash := sha256.New()
	hash.Write([]byte(checklistID + userID))
	expectedShortCode := fmt.Sprintf("%x", hash.Sum(nil))[0:11]

	result, err := GetShareCode(checklistID, userID)
	if err != nil {
		t.Fatalf("Expected short code but got: %v", err)
	}

	if expectedShortCode != result {
		t.Fatalf("Expected %s to equal %s", expectedShortCode, result)
	}
}

func TestGetTokenFromShareCode(t *testing.T) {
	userID := "some-user-id"
	checklistID := "some-checklist-id"
	shareCode, err := GetShareCode(checklistID, userID)

	if err != nil {
		t.Fatalf("Could not get share code: %v", err)
	}

	token, err := GetTokenFromShareCode(shareCode)
	if err != nil {
		t.Fatalf("Could not get token from share code: %v", err)
	}

	claims, err := ParseSharingToken(token)
	if err != nil {
		t.Fatalf("Could not parse token: %v", err)
	}

	if claims.ChecklistID != checklistID {
		t.Fatalf("Expected checklistID %s to equal %s", claims.ChecklistID, checklistID)
	}

	if claims.UserID != userID {
		t.Fatalf("Expected userID %s to equal %s", claims.UserID, userID)
	}
}
