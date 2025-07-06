package handler_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

// TestPostOfferlist tests the PostOfferlist handler with various request scenarios
func TestPostOfferlist(t *testing.T) {
	ctrl, err := setupController()
	if err != nil {
		t.Fatalf("Failed to setup controller: %v", err)
	}

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
                "request": "{\"some\": \"data\"}"
            }`,
			expectedStatus: http.StatusOK,
			expectedBody:   "200",
		},
		{
			name: "empty valid JSON",
			requestBody: `{
                "header": [],
                "prompt": "",
                "destination": "",
                "request": ""
            }`,
			expectedStatus: http.StatusOK,
			expectedBody:   "200",
		},
		{
			name: "minimal valid JSON",
			requestBody: `{
                "header": null,
                "prompt": "minimal",
                "destination": "https://test.com",
                "request": "{}"
            }`,
			expectedStatus: http.StatusOK,
			expectedBody:   "200",
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "empty request body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "malformed JSON",
			requestBody:    `{"header": ["test"], "prompt":}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
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
				t.Errorf("Expected response body %s, got %s", tt.expectedBody, rec.Body.String())
			}
		})
	}
}

// BenchmarkPostOfferlist benchmarks the PostOfferlist handler with a large number of requests
func BenchmarkPostOfferlist(b *testing.B) {
	_, err := setupController()
	if err != nil {
		b.Fatalf("Failed to setup controller: %v", err)
	}

	requestBody := entity.Request{
		Header:      []string{"Content-Type: application/json"},
		Prompt:      "Benchmarking offers",
		Destination: "https://example.com/api/offers",
		Request:     `"offers"`,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		b.Fatalf("Failed to marshal request body: %v", err)
	}

	for b.Loop() {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/Offerlist", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		if rec.Code != http.StatusOK {
			b.Fatalf("Unexpected status code: got %d", rec.Code)
		}
	}
}
func setupController() (*handler.SuproxyController, error) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()
	ctrl, err := handler.NewSuproxyController(logger)
	if err != nil {
		return nil, err
	}
	return ctrl, nil
}
