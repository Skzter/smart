package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT secret (in production, use environment variables).
// This should be a strong, randomly generated secret in real applications.
var jwtSecret = []byte("your-secret-key") //SECRET-KEY NOCH ÄNDERN!!!

// JWTClaims represents the claims in our JWT tokens.
// In a real application, you would include additional claims like issuer, audience, etc.
type JWTClaims struct {
	UserID string   `json:"user_id"` // User identifier
	Scopes []string `json:"scopes"`  // Permissions/roles for the user
	jwt.RegisteredClaims
}

func GenerateToken(userID string, scopes []string, expiration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "check24-smart-mcp",                            // MCP-SERVER SELBST
			Subject:   userID,                                         // Subject is the user ID
			Audience:  []string{"check24-smart-mcp-clients"},          // Intended audience ÜBERARBEITEN
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)), // Token expiration time
			IssuedAt:  jwt.NewNumericDate(time.Now()),                 // Token issuance time
			NotBefore: jwt.NewNumericDate(time.Now()),                 // Token valid from time
		},
	}
	// Create and sign the JWT token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
