package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/infrastructure/database"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Auth is an Interface which holds all methods which an AuthService implements
type Auth interface {
	GenerateToken(ctx context.Context, userId string) (*entity.Token, error)
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

func (a *auth) GenerateToken(ctx context.Context, userId string) (*entity.Token, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	if err := assert.StringNotEmpty(userId); err != nil {
		return nil, err
	}
	dbToken, err := a.db.ReadToken(ctx, userId)
	switch err {
	case sql.ErrNoRows:
		token := rand.Text()
		expiresAt := time.Now().UTC().Add(time.Hour * time.Duration(a.config.TokenExpirationTimeHours))
		err := a.db.CreateToken(ctx, database.CreateTokenParams{
			UserID:    userId,
			Token:     token,
			ExpiresAt: expiresAt,
		})
		if err != nil {
			return nil, err
		}
		return &entity.Token{
			UserID:    userId,
			Token:     token,
			ExpiresAt: expiresAt,
		}, nil
	case nil:
		if dbToken.ExpiresAt.Before(time.Now().UTC()) || dbToken.RevokedAt.Valid {
			token := rand.Text()
			expiresAt := time.Now().UTC().Add(time.Hour * time.Duration(a.config.TokenExpirationTimeHours))
			err := a.db.UpdateToken(ctx, database.UpdateTokenParams{
				UserID:    dbToken.UserID,
				Token:     token,
				ExpiresAt: expiresAt,
				RevokedAt: sql.NullTime{},
			})
			if err != nil {
				return nil, err
			}
			return &entity.Token{
				UserID:    dbToken.UserID,
				Token:     token,
				ExpiresAt: expiresAt,
			}, nil
		}
		return &entity.Token{
			UserID:    dbToken.UserID,
			Token:     dbToken.Token,
			ExpiresAt: dbToken.ExpiresAt,
		}, nil
	default:
		return nil, err
	}
}

// GetBearerToken returns the bearer token in the Authorization Header of an http request
func (ts *auth) GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("authorization header missing")
	}
	splitHeader := strings.Split(header, " ")
	if len(splitHeader) < 2 || splitHeader[0] != "Bearer" {
		return "", fmt.Errorf("malformed authorization header")
	}
	return splitHeader[1], nil
}
