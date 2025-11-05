package service

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

// nolint:funlen
func TestNewTaglistSync(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	taglistError := errors.New("Taglist error")

	tests := []struct {
		name          string
		logger        *slog.Logger
		expectError   bool
		expectedError error
		wantError     bool
		mockResponse  []any
	}{
		{
			name:          "success - taglist loaded correctly",
			logger:        logger,
			expectError:   false,
			expectedError: nil,
			wantError:     false,
			mockResponse:  []any{[]string{"tag1", "tag2"}, nil},
		},
		{
			name:          "error - logger is nil",
			logger:        nil,
			expectError:   true,
			expectedError: nil,
			wantError:     true,
		},
		{
			name:          "error - taglistservice fails",
			logger:        logger,
			expectError:   true,
			expectedError: taglistError,
			wantError:     true,
			mockResponse:  []any{nil, taglistError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := mocks.NewMockTaglistStorage(t)
			if tt.mockResponse != nil {
				serv.On("GetTaglist", mock.Anything).Return(tt.mockResponse...)
			}
			sync, err := NewTaglistSync(tt.logger, serv)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected => %s, got => nil", "not nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected error to be nil, got => %s", err.Error())
				}
				if sync == nil {
					t.Fatal("expected sync to be not nil")
				}
			}
		})
	}
}
