package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Auth is an Interface which holds all methods which an AuthService implements
type Auth interface {
	MakeJWT(userId string) (string, error)
	ValidateJWT(tokenString string) (string, error)
	GetBearerToken(headers http.Header) (string, error)
}

type authService struct {
	logger *slog.Logger
	config *config.Config
}

// NewAuthService return the AuthService or error
func NewAuthService(logger *slog.Logger, config *config.Config) (Auth, error) {
	if err := assert.NotNil(logger, config); err != nil {
		return nil, err
	}

	return &authService{
		logger: logger,
		config: config,
	}, nil
}

// nolint: gochecknoglobals
var tokenSecret = []byte("hier wird mal ein toller secret key sein")

// MakeJWT takes userid and returns jwt on success
func (ts *authService) MakeJWT(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    ts.config.JwtTokenIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(ts.config.JwtExpirationTimeInHours) * time.Hour)),
		Subject:   userId,
	})
	signedString, err := token.SignedString(tokenSecret)
	if err != nil {
		return "", nil
	}
	return signedString, nil
}

// ValidateJWT takes existing token and validates expiration time, issuer and format
func (ts *authService) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		return tokenSecret, nil
	})
	if err != nil {
		return "", err
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return "", err
	}

	if issuer != ts.config.JwtTokenIssuer {
		return "", errors.New("invalid user")
	}

	return userId, nil
}

// GetBearerToken returns the bearer token in the Authorization Header of an http request
func (ts *authService) GetBearerToken(headers http.Header) (string, error) {
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
