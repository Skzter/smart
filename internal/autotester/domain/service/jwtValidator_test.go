package service

import (
	"context"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTValidatorValidate(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-123")

	v, err := NewJWTValidator()
	assert.NoError(t, err)

	sign := func(claims jwt.Claims, secret string) string {
		t.Helper()
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, signErr := tok.SignedString([]byte(secret))
		assert.NoError(t, signErr)
		return s
	}

	now := time.Now()

	validToken := sign(jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
	}, "test-secret-123")

	expiredToken := sign(jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(-10 * time.Minute)),
	}, "test-secret-123")

	missingExpToken := sign(jwt.MapClaims{
		"sub": "user123",
	}, "test-secret-123")

	invalidSignatureToken := sign(jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
	}, "wrong-secret")

	tests := []struct {
		name      string
		token     string
		wantValid bool
		wantErr   bool
	}{
		{"empty token -> error + invalid", "", false, true},
		{"whitespace token -> error + invalid", "   ", false, true},
		{"valid token -> ok", validToken, true, false},
		{"expired token -> error + invalid", expiredToken, false, true},
		{"missing exp -> error + invalid", missingExpToken, false, true},
		{"invalid signature -> error + invalid", invalidSignatureToken, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, e := v.Validate(context.Background(), tc.token)

			if tc.wantErr {
				assert.Error(t, e)
			} else {
				assert.NoError(t, e)
			}

			assert.Equal(t, tc.wantValid, res.Valid)
			assert.False(t, res.Revoked)
		})
	}
}
