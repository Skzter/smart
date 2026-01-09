package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/infrastructure/database"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Auth is an Interface which holds all methods which an AuthService implements
type Auth interface {
	GenerateToken(ctx context.Context, userId string) (string, error)
	GetBearerToken(headers http.Header) (string, error)
}

type auth struct {
	logger *slog.Logger
	config *config.Config
	db     repository.TokenDatabase
}

// NewAuthService return the AuthService or error
func NewAuthService(logger *slog.Logger, config *config.Config, db repository.TokenDatabase) (Auth, error) {
	if err := assert.NotNil(logger, config, db); err != nil {
		return nil, err
	}

	return &auth{
		logger: logger,
		config: config,
		db:     db,
	}, nil
}

func (a *auth) GenerateToken(ctx context.Context, userId string) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return "", err
	}
	if err := assert.StringNotEmpty(userId); err != nil {
		return "", err
	}
	token, err := a.db.ReadToken(ctx, userId)
	switch err {
	case sql.ErrNoRows:
		token := rand.Text()
		err := a.db.CreateToken(ctx, database.CreateTokenParams{
			UserID:    userId,
			Token:     token,
			ExpiresAt: time.Now().UTC().Add(time.Hour * time.Duration(a.config.TokenExpirationTimeHours)),
		})
		if err != nil {
			return "", err
		}
		return token, nil
	case nil:
		if token.ExpiresAt.Before(time.Now()) || token.RevokedAt.Valid {
			token := rand.Text()
			err := a.db.UpdateToken(ctx, database.UpdateTokenParams{
				UserID:    userId,
				Token:     token,
				ExpiresAt: time.Now().UTC().Add(time.Hour * time.Duration(a.config.TokenExpirationTimeHours)),
				RevokedAt: sql.NullTime{},
			})
			if err != nil {
				return "", err
			}
			return token, nil
		}
		return token.Token, nil
	default:
		return "", err
	}
}

// GetBearerToken returns the bearer token in the Authorization Header of an http request
func (ts *auth) GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", errors.New("no header detected")
	}
	splitHeader := strings.Split(header, " ")
	if len(splitHeader) < 2 || splitHeader[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}
	return splitHeader[1], nil
}
