package handler_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	sharedentity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	mockrepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	suproxyentity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// TestPostOfferlist tests the PostOfferlist handler with various request scenarios
func TestPostOfferlist(t *testing.T) {
	ctrl := setupController(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "valid JSON request",
			requestBody: `{
				"header": ["Content-Type: application/json"],
				"prompt": "test prompt",
				"destination": "https://example.com",
				"request": "{\"httpstatuscode\":200,\"data\":{\"items\":[\"{}\"]}}"
			}`,
			expectedStatus: http.StatusOK,
			expectedBody:   "200",
		},
		{
			name: "empty valid JSON (no request field)",
			requestBody: `{
				"header": [],
				"prompt": "",
				"destination": "",
				"request": ""
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"empty json string"}`,
		},
		{
			name: "minimal valid JSON",
			requestBody: `{
				"header": null,
				"prompt": "minimal",
				"destination": "https://example.com",
				"request": "{\"httpstatuscode\":200,\"data\":{\"items\":[\"{}\"]}}"
			}`,
			expectedStatus: http.StatusOK,
			expectedBody:   "200",
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty request body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed JSON",
			requestBody:    `{"header": ["test"], "prompt":}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/v1/Offerlist", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			ctrl.PostOfferlist(ctx)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" && rec.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rec.Body.String())
			}
		})
	}
}

// BenchmarkPostOfferlist benchmarks the PostOfferlist handler with a large number of requests
func BenchmarkPostOfferlist(b *testing.B) {
	ctrl := setupController(b)

	requestBody := suproxyentity.Request{
		Header:      []string{"Content-Type: application/json"},
		Prompt:      "Benchmarking offers",
		Destination: "https://example.com/api/offers",
		Request:     `{"httpstatuscode":200,"data":{"items":["{}"]}}`,
	}

	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		b.Fatalf("Failed to marshal request body: %v", err)
	}

	for b.Loop() {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/Offerlist", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = req

		ctrl.PostOfferlist(ctx)

		if rec.Code != http.StatusOK {
			b.Fatalf("Unexpected status code: got %d", rec.Code)
		}
	}
}

// setupController initializes the SuproxyController for testing
func setupController(tb testing.TB) *handler.SuproxyController {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	cfg := &config.Config{
		Model: "gpt-4o",
		Prompts: &config.Prompts{
			ValidationPrompt: "test prompt",
		},
		Timeout:               20,
		MaxItemsPerValidation: 10,
	}

	validator := service.NewValidator(logger, cfg)

	mockConnector := mockrepo.NewOpenAI(tb)
	mockConnector.
		On("CreateRequest", mock.Anything, mock.Anything).
		Return(&sharedentity.Response{Text: `{"valid": true, "reason": []}`}, nil)

	validator.SetOpenAIService(mockConnector)

	ctrl, err := handler.NewSuproxyController(logger, validator)
	if err != nil {
		tb.Fatalf("Failed to create controller: %v", err)
	}
	return ctrl
}
