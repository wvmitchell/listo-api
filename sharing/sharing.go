// Package sharing provides the functions to generate and validate sharing codes.
package sharing

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Claims is a struct that contains the claims for the JWT.
type Claims struct {
	ChecklistID string `json:"checklist_id"`
	jwt.StandardClaims
}

// GenerateSharingCode generates a sharing code for a checklist.
func GenerateSharingCode(checklistID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		ChecklistID: checklistID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ParseSharingCode parses a sharing code and returns the checklist ID.
func ParseSharingCode(token string) (string, error) {
	tkn, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return "", errors.New("token is expired")
			}
			return "", err
		}
		return "", err
	}

	claims, ok := tkn.Claims.(*Claims)
	if !ok {
		return "", err
	}

	return claims.ChecklistID, nil
}
