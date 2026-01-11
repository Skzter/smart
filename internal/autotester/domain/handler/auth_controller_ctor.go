package handler

import (
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
)

// NewAuthController returns a minimal controller instance that is sufficient
// for the auth validate endpoint. This avoids having to build the whole
// controller graph in unit tests.
func NewAuthController(logger *slog.Logger, jwtValidator service.JWTValidator) *AutotesterController {
	return &AutotesterController{
		logger:       logger,
		jwtValidator: jwtValidator,
	}
}
