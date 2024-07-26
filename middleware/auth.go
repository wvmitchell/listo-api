// Package middleware provides the middleware for the application.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
)

var (
	auth0domain   string
	auth0audience string
	jwksURL       string
)

func initAuth() {
	auth0domain = os.Getenv("AUTH0_DOMAIN")
	auth0audience = os.Getenv("AUTH0_AUDIENCE")
	jwksURL = fmt.Sprintf("https://%s/.well-known/jwks.json", auth0domain)
}

func fetchJWKS() (jwk.Set, error) {
	set, err := jwk.Fetch(context.Background(), jwksURL)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse JWKS: %w", err)
	}

	return set, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	keySet, err := fetchJWKS()
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, found := token.Header["kid"].(string)
		if !found {
			return nil, fmt.Errorf("kid not found in token header")
		}

		key, found := keySet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("key not found in key set")
		}

		var rawKey interface{}
		if err := key.Raw(&rawKey); err != nil {
			return nil, fmt.Errorf("failed to get raw key: %w", err)
		}

		return rawKey, err
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		aud, ok := claims["aud"].([]interface{})
		if !ok {
			// if aud is not an array, convert it to an array for uniformity
			aud = []interface{}{claims["aud"]}
		}

		validAudience := false
		for _, audience := range aud {
			if audience == auth0audience {
				validAudience = true
				break
			}
		}

		if !validAudience {
			return nil, fmt.Errorf("invalid audience")
		}

		if claims["iss"] != fmt.Sprintf("https://%s/", auth0domain) {
			return nil, fmt.Errorf("invalid issuer")
		}
	} else {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

// AuthMiddleware is a middleware that checks the Authorization header,
// validates the claims made, and makes those claims available via the gin context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		initAuth()
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			fmt.Print("auth header required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := verifyToken(tokenString)
		if err != nil {
			fmt.Print("token not verified: ", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Print("claims invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		}

		c.Set("sub", claims["sub"])
		c.Next()
	}
}
