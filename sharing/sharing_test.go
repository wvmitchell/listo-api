// Package sharing provides the functions to generate and validate sharing codes.
package sharing

import (
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

var testJwtKey = []byte("test_jwt_key")

func TestGenerateShareToken(t *testing.T) {
	checklistID := "test_checklist_id"
	token, err := GenerateSharingCode(checklistID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatalf("Generated token is empty")
	}
}

func TestParseShareToken(t *testing.T) {
	checklistID := "test_checklist_id"
	token, err := GenerateSharingCode(checklistID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	parsedChecklistID, err := ParseSharingCode(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if parsedChecklistID != checklistID {
		t.Fatalf("Expected checklistID %v, but got %v", checklistID, parsedChecklistID)
	}
}

func TestParseInvalidToken(t *testing.T) {
	invalidToken := "invalid_token"

	_, err := ParseSharingCode(invalidToken)
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

	_, err = ParseSharingCode(tokenString)
	if err == nil || err.Error() != "token is expired" {
		t.Fatalf("Expected 'token is expired' error, but got: %v", err)
	}
}
