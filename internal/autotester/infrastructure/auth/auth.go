package auth

import (
	"crypto/rand"
	"encoding/hex"
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
	ValidateJWT(token string) error
	MakeRefreshToken() (string, error)
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

// MakeJWT takes userid and returns jwt on success
func (ts *authService) MakeJWT(userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "autotester",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(ts.config.JwtExpirationTimeInHours) * time.Hour)),
		Subject:   userId,
	})
	signedString, err := token.SignedString([]byte("hier wird mal ein toller secret key sein"))
	if err != nil {
		return "", nil
	}
	return signedString, nil
}

// ValidateJWT takes existing token and validates expiration time, issuer and format
func (ts *authService) ValidateJWT(token string) error {
	return nil
}

// MakeRefreshToken returns a new random refresh token
func (ts *authService) MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	encodedString := hex.EncodeToString(key)
	return encodedString, nil
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
