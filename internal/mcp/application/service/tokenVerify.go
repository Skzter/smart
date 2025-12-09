package service

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/modelcontextprotocol/go-sdk/auth"
)

type TokenInfo struct {
	UserID    string
	Scopes    []string
	ExpiresAt int64
}

// verifyJWT verifies JWT tokens and returns TokenInfo for the auth middleware.
// This function implements the TokenVerifier interface required by auth.RequireBearerToken.
func verifyJWT(ctx context.Context, tokenString string) (*auth.TokenInfo, error) {
	// Parse and validate the JWT token.
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Verify the signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		// Return standard error for invalid tokens.
		return nil, fmt.Errorf("%w: %v", auth.ErrInvalidToken, err)
	}

	// Extract claims and verify token validity.
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return &auth.TokenInfo{
			Scopes:     claims.Scopes,         // User permissions
			Expiration: claims.ExpiresAt.Time, // Token expiration time
		}, nil
	}

	return nil, fmt.Errorf("%w: invalid token claims", auth.ErrInvalidToken)
}

/*func VerifyJWT(ctx context.Context, tokenString string) (*TokenInfo, error) {

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		// Ensure signing method is correct
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}
		return jwtSecret, nil
	}, jwt.WithIssuer("check24-smart-mcp"),
	   jwt.WithAudience("check24-smart-mcp-clients"))

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return &TokenInfo{
		UserID:     claims.UserID,
		Scopes:     claims.Scopes,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}*/
