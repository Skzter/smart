package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// nolint:funlen
func TestJWTValidatorValidate(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		wantValid   bool
		wantRevoked bool
		wantErr     bool
	}{
		{
			name:        "empty token -> error + invalid",
			token:       "",
			wantValid:   false,
			wantRevoked: false,
			wantErr:     true,
		},
		{
			name:        "whitespace token -> error + invalid",
			token:       "   ",
			wantValid:   false,
			wantRevoked: false,
			wantErr:     true,
		},
		{
			name:        "non-empty token -> valid (stub behavior)",
			token:       "abc",
			wantValid:   true,
			wantRevoked: false,
			wantErr:     false,
		},
		{
			name:        "non-empty token with spaces -> valid (stub behavior)",
			token:       "  abc  ",
			wantValid:   true,
			wantRevoked: false,
			wantErr:     false,
		},
	}

	v := NewJWTValidator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := v.Validate(context.Background(), tc.token)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.wantValid, res.Valid)
			assert.Equal(t, tc.wantRevoked, res.Revoked)
		})
	}
}
