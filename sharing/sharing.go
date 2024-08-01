// Package sharing provides the functions to generate and validate sharing codes.
package sharing

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	"checklist-api/db"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Claims is a struct that contains the claims for the JWT.
type Claims struct {
	ChecklistID string `json:"checklist_id"`
	UserID      string `json:"user_id"`
	jwt.StandardClaims
}

// generateSharingToken generates a sharing code for a checklist.
func generateSharingToken(checklistID string, userID string) (string, error) {
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &Claims{
		ChecklistID: checklistID,
		UserID:      userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

// ParseSharingToken parses a sharing code and returns the checklist ID and user ID.
func ParseSharingToken(token string) (*Claims, error) {
	tkn, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token is expired")
			}
			return nil, err
		}
		return nil, err
	}

	claims, ok := tkn.Claims.(*Claims)
	if !ok {
		return nil, err
	}

	return claims, nil
}

// GetShareCode creates a short code, and stores a sharing token in redis at that address
func GetShareCode(checklistID string, userID string) (string, error) {
	service, err := db.NewRedisService()
	if err != nil {
		return "error setting up redis service", err
	}

	hash := sha256.New()
	hash.Write([]byte(checklistID + userID + time.Now().String()))
	shortCode := fmt.Sprintf("%x", hash.Sum(nil))[0:11]

	token, err := generateSharingToken(checklistID, userID)
	if err != nil {
		return "error creating token", err
	}

	err = service.SetShortCodeWithJWT(shortCode, token)
	if err != nil {
		return "error saving token in redis", err
	}

	return shortCode, nil
}

// GetTokenFromShareCode retrieves the token from redis for that short code
func GetTokenFromShareCode(shareCode string) (string, error) {
	service, err := db.NewRedisService()
	if err != nil {
		return "error setting up redis service", err
	}

	token, err := service.GetJWTFromShortCode(shareCode)
	if err != nil {
		return "error getting token from redis service", err
	}

	return token, nil
}
