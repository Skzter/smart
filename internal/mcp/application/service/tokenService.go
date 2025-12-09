package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret is the HMAC secret used to sign JWTs.
// In production this must be provided via secure configuration (env/secret manager) and be strong and random.
var jwtSecret = []byte("your-secret-key") //SECRET-KEY NOCH ÄNDERN!!!

// JWTClaims represents the custom claims stored in our JWT tokens.
// It embeds jwt.RegisteredClaims and adds application-specific fields such as UserID and Scopes.
type JWTClaims struct {
	UserID string   `json:"user_id"` // User identifier
	Scopes []string `json:"scopes"`  // Permissions/roles assigned to the user
	jwt.RegisteredClaims
}

// GenerateToken creates and signs a JWT for the given user with the provided scopes and expiration.
// It returns the signed token string or an error if signing fails.
//
// Parameters:
//   - userID: unique identifier of the user for whom the token is issued.
//   - scopes: list of permission scopes to embed in the token.
//   - expiration: validity duration from the current time.
//
// Returns:
//   - signed JWT string on success
//   - error if token creation or signing fails
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
