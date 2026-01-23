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

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	config *config.Autotester
	db     repository.TokenDatabase
	tracer trace.Tracer
}

// NewAuthService return the AuthService or error
func NewAuthService(logger *slog.Logger, config *config.Autotester, db repository.TokenDatabase, tracer trace.Tracer) (Auth, error) {
	if err := assert.NotNil(logger, config, db, tracer); err != nil {
		return nil, err
	}

	return &auth{
		logger: logger,
		config: config,
		db:     db,
		tracer: tracer,
	}, nil
}

// GenerateToken generates a Token with a given userId
// if the user has a valid token in the db, it returns that token
// else it will generate a new one (if expired oder revoked)
// returns token, error
func (a *auth) GenerateToken(ctx context.Context, userId string) (*entity.Token, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	ctx, span := a.tracer.Start(ctx, "autotesterController.GenerateToken")
	defer span.End()
	if err := assert.StringNotEmpty(userId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing parameter")
		return nil, err
	}
	dbToken, err := a.db.ReadToken(ctx, userId)
	if err == nil {
		isExpired := dbToken.ExpiresAt.Before(time.Now().UTC())
		isRevoked := dbToken.RevokedAt.Valid
		if !isExpired && !isRevoked {
			return &entity.Token{
				UserID:    dbToken.UserID,
				Token:     dbToken.Token,
				CreatedAt: dbToken.CreatedAt,
				UpdatedAt: dbToken.UpdatedAt,
				ExpiresAt: dbToken.ExpiresAt,
				RevokedAt: convSqlNullTimeIntoTime(dbToken.RevokedAt),
			}, nil
		}
	} else if err != sql.ErrNoRows {
		span.RecordError(err)
		span.SetStatus(codes.Error, "read database error")
		return nil, err
	}
	token := rand.Text()
	expiresAt := time.Now().UTC().Add(a.config.TokenExpirationTime.GoDuration())
	dbToken, err = a.db.UpsertToken(ctx, database.UpsertTokenParams{
		UserID:    userId,
		Token:     token,
		ExpiresAt: expiresAt,
		RevokedAt: sql.NullTime{
			Valid: false,
		},
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "write database error")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return &entity.Token{
		UserID:    dbToken.UserID,
		Token:     dbToken.Token,
		CreatedAt: dbToken.CreatedAt,
		UpdatedAt: dbToken.UpdatedAt,
		ExpiresAt: dbToken.ExpiresAt,
		RevokedAt: convSqlNullTimeIntoTime(dbToken.RevokedAt),
	}, nil
}

// GetBearerToken returns the bearer token in the Authorization Header of an http request
func (a *auth) GetBearerToken(headers http.Header) (string, error) {
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

// convSqlNullTimeIntoTime converts database null time into go time
// returns nil if nil and time if not nil
func convSqlNullTimeIntoTime(sqltime sql.NullTime) *time.Time {
	if !sqltime.Valid {
		return nil
	} else {
		return &sqltime.Time
	}
}
