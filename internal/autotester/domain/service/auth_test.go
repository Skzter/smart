package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mockRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/infrastructure/database"
)

func TestNewAuthService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()

	tests := []struct {
		name    string
		logger  *slog.Logger
		config  *config.Autotester
		db      *mockRepo.MockTokenDatabase
		tracer  trace.Tracer
		wantErr bool
	}{
		{
			name:    "success",
			logger:  logger,
			config:  cfg,
			db:      mockRepo.NewMockTokenDatabase(t),
			tracer:  otel.Tracer("test"),
			wantErr: false,
		},
		{
			name:    "error - nil arguments",
			logger:  nil,
			config:  nil,
			db:      nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := NewAuthService(tc.logger, tc.config, tc.db, tc.tracer)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, srv)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, srv)
			}
		})
	}
}

// nolint:funlen
func TestGenerateToken(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	userID := "testUser123"

	futureTime := time.Now().UTC().Add(time.Hour * 24)
	pastTime := time.Now().UTC().Add(-time.Hour * 24)

	// Use a sentinel context to explicitly test nil context
	type ctxKey struct{}
	nilCtxSentinel := context.WithValue(context.Background(), ctxKey{}, "nil")

	tests := []struct {
		name           string
		ctx            context.Context
		userID         string
		readResp       []any
		upsertResp     []any
		wantErr        bool
		expectedToken  string
		expectNilToken bool
	}{
		{
			name:   "success - valid token exists in db",
			userID: userID,
			readResp: []any{
				database.RefreshToken{
					ID:        1,
					UserID:    userID,
					Token:     "existing-token-123",
					CreatedAt: time.Now().UTC().Add(-time.Hour),
					UpdatedAt: time.Now().UTC().Add(-time.Hour),
					ExpiresAt: futureTime,
					RevokedAt: sql.NullTime{Valid: false},
				}, nil,
			},
			wantErr:       false,
			expectedToken: "existing-token-123",
		},
		{
			name:   "success - expired token, generates new one",
			userID: userID,
			readResp: []any{
				database.RefreshToken{
					ID:        1,
					UserID:    userID,
					Token:     "expired-token",
					CreatedAt: time.Now().UTC().Add(-time.Hour * 48),
					UpdatedAt: time.Now().UTC().Add(-time.Hour * 48),
					ExpiresAt: pastTime,
					RevokedAt: sql.NullTime{Valid: false},
				}, nil,
			},
			upsertResp: []any{
				database.RefreshToken{
					ID:        2,
					UserID:    userID,
					Token:     "new-token-456",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					ExpiresAt: futureTime,
					RevokedAt: sql.NullTime{Valid: false},
				}, nil,
			},
			wantErr:       false,
			expectedToken: "new-token-456",
		},
		{
			name:   "success - revoked token, generates new one",
			userID: userID,
			readResp: []any{
				database.RefreshToken{
					ID:        1,
					UserID:    userID,
					Token:     "revoked-token",
					CreatedAt: time.Now().UTC().Add(-time.Hour),
					UpdatedAt: time.Now().UTC().Add(-time.Hour),
					ExpiresAt: futureTime,
					RevokedAt: sql.NullTime{Valid: true, Time: time.Now().UTC()},
				}, nil,
			},
			upsertResp: []any{
				database.RefreshToken{
					ID:        3,
					UserID:    userID,
					Token:     "new-token-789",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					ExpiresAt: futureTime,
					RevokedAt: sql.NullTime{Valid: false},
				}, nil,
			},
			wantErr:       false,
			expectedToken: "new-token-789",
		},
		{
			name:   "success - no token in db, creates new one",
			userID: userID,
			readResp: []any{
				database.RefreshToken{}, sql.ErrNoRows,
			},
			upsertResp: []any{
				database.RefreshToken{
					ID:        4,
					UserID:    userID,
					Token:     "brand-new-token",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					ExpiresAt: futureTime,
					RevokedAt: sql.NullTime{Valid: false},
				}, nil,
			},
			wantErr:       false,
			expectedToken: "brand-new-token",
		},
		{
			name:   "error - database read error",
			userID: userID,
			readResp: []any{
				database.RefreshToken{}, sql.ErrConnDone,
			},
			wantErr:        true,
			expectNilToken: true,
		},
		{
			name:   "error - database upsert error",
			userID: userID,
			readResp: []any{
				database.RefreshToken{}, sql.ErrNoRows,
			},
			upsertResp: []any{
				database.RefreshToken{}, errors.New("db insert error"),
			},
			wantErr:        true,
			expectNilToken: true,
		},
		{
			name:           "error - nil context",
			ctx:            nilCtxSentinel,
			userID:         userID,
			wantErr:        true,
			expectNilToken: true,
		},
		{
			name:           "error - empty userId",
			userID:         "",
			wantErr:        true,
			expectNilToken: true,
		},
	}

	validateToken := func(t *testing.T, token *entity.Token, tc struct {
		name           string
		ctx            context.Context
		userID         string
		readResp       []any
		upsertResp     []any
		wantErr        bool
		expectedToken  string
		expectNilToken bool
	}) {
		if tc.expectNilToken {
			assert.Nil(t, token)
			return
		}

		assert.NotNil(t, token)
		assert.Equal(t, tc.userID, token.UserID)
		if tc.expectedToken != "" {
			assert.Equal(t, tc.expectedToken, token.Token)
		}
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.ctx
			passNilCtx := false

			// Check if this is the nil context test case
			switch ctx {
			case nilCtxSentinel:
				passNilCtx = true
				ctx = nil
			case nil:
				// Use t.Context() as default if ctx is not explicitly set
				ctx = t.Context()
			}

			mockDB := mockRepo.NewMockTokenDatabase(t)

			// Only set up mocks if we expect the methods to be called
			// (i.e., userID is not empty and we're not testing nil context)
			if tc.userID != "" && !passNilCtx {
				if tc.readResp != nil {
					mockDB.On("ReadToken",
						mock.Anything, tc.userID,
					).Return(tc.readResp...)
				}

				if tc.upsertResp != nil {
					mockDB.On("UpsertToken",
						mock.Anything, mock.Anything,
					).Return(tc.upsertResp...)
				}
			}

			authSrv := &auth{
				logger: logger,
				config: cfg,
				db:     mockDB,
				tracer: otel.Tracer("test"),
			}

			token, err := authSrv.GenerateToken(ctx, tc.userID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			validateToken(t, token, tc)
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	mockDB := mockRepo.NewMockTokenDatabase(t)

	authSrv := &auth{
		logger: logger,
		config: cfg,
		db:     mockDB,
		tracer: otel.Tracer("test"),
	}

	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "success - valid bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer my-secret-token-123"},
			},
			wantToken: "my-secret-token-123",
			wantErr:   false,
		},
		{
			name: "success - token with special characters",
			headers: http.Header{
				"Authorization": []string{"Bearer abc123-XYZ_456.789"},
			},
			wantToken: "abc123-XYZ_456.789",
			wantErr:   false,
		},
		{
			name:      "error - missing authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "error - empty authorization header",
			headers: http.Header{
				"Authorization": []string{""},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "error - malformed header (missing Bearer)",
			headers: http.Header{
				"Authorization": []string{"my-token-without-bearer"},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "error - malformed header (wrong scheme)",
			headers: http.Header{
				"Authorization": []string{"Basic dXNlcjpwYXNz"},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "error - malformed header (only Bearer)",
			headers: http.Header{
				"Authorization": []string{"Bearer"},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := authSrv.GetBearerToken(tc.headers)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantToken, token)
			}
		})
	}
}

func TestConvSqlNullTimeIntoTime(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		input    sql.NullTime
		expected *time.Time
	}{
		{
			name: "valid time",
			input: sql.NullTime{
				Valid: true,
				Time:  now,
			},
			expected: &now,
		},
		{
			name: "null time",
			input: sql.NullTime{
				Valid: false,
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convSqlNullTimeIntoTime(tc.input)

			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected.Unix(), result.Unix())
			}
		})
	}
}

// nolint:funlen
func TestValidateToken(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tr := otel.Tracer("test")

	now := time.Now().UTC()

	tests := []struct {
		name        string
		token       string
		dbResp      []any
		wantValid   bool
		wantRevoked bool
		wantErr     bool

		nilCtx bool
	}{
		{
			name:    "nil ctx -> error",
			token:   "some-token",
			wantErr: true,
			nilCtx:  true,
		},
		{
			name:    "empty token -> error",
			token:   "",
			wantErr: true,
		},
		{
			name:   "token not in db -> invalid",
			token:  "unknown-token",
			dbResp: []any{database.RefreshToken{}, sql.ErrNoRows},
		},
		{
			name:  "token expired in db -> invalid",
			token: "expired-token",
			dbResp: []any{
				database.RefreshToken{
					Token:     "expired-token",
					ExpiresAt: now.Add(-time.Minute),
					RevokedAt: sql.NullTime{Valid: false},
				}, nil,
			},
			wantValid: false,
		},
		{
			name:  "token revoked in db -> invalid + revoked",
			token: "revoked-token",
			dbResp: []any{
				database.RefreshToken{
					Token:     "revoked-token",
					ExpiresAt: now.Add(time.Hour),
					RevokedAt: sql.NullTime{Valid: true, Time: now},
				}, nil,
			},
			wantValid:   false,
			wantRevoked: true,
		},
		{
			name:  "token active in db -> valid",
			token: "active-token",
			dbResp: []any{
				database.RefreshToken{
					Token:     "active-token",
					ExpiresAt: now.Add(time.Hour),
					RevokedAt: sql.NullTime{Valid: false},
				}, nil,
			},
			wantValid: true,
		},
		{
			name:    "db error -> error",
			token:   "active-token",
			dbResp:  []any{database.RefreshToken{}, errors.New("db down")},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mockRepo.NewMockTokenDatabase(t)

			// DB is only called if ctx != nil AND token after TrimSpace is non-empty
			if !tc.nilCtx && strings.TrimSpace(tc.token) != "" {
				if tc.dbResp != nil {
					tok := strings.TrimSpace(tc.token)
					mockDB.
						On("ReadTokenByToken", mock.Anything, tok).
						Return(tc.dbResp...)
				}
			}

			authSrv := &auth{
				logger: logger,
				config: cfg,
				db:     mockDB,
				tracer: tr,
			}

			ctx := t.Context()
			if tc.nilCtx {
				ctx = nil
			}

			res, err := authSrv.ValidateToken(ctx, tc.token)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tc.wantValid, res.Valid)
			assert.Equal(t, tc.wantRevoked, res.Revoked)
		})
	}
}
