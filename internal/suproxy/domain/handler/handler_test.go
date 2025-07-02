package handler_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

// TestPostOfferlist tests the PostOfferlist handler with various inputs
func TestPostOfferlist(t *testing.T) {
	router := setupRouter()

	tests := map[string]struct {
		input        entity.Request
		expectStatus int
	}{
		"valid request": {
			input: entity.Request{
				Header:      []string{"Content-Type: application/json"},
				Prompt:      "Please provide a list of offers",
				Destination: "https://example.com/api/offers",
				Request:     `{"query": "offers"}`,
			},
			expectStatus: http.StatusOK,
		},
		"empty body": {
			input:        entity.Request{},
			expectStatus: http.StatusOK,
		},
		"missing destination": {
			input: entity.Request{
				Header:  []string{"Content-Type: application/json"},
				Prompt:  "Prompt without destination",
				Request: `{"data": "something"}`,
			},
			expectStatus: http.StatusOK,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jsonBody, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req, err := http.NewRequest("POST", "/api/v1/Offerlist", strings.NewReader(string(jsonBody)))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectStatus {
				t.Errorf("Expected status %d, got %d", tc.expectStatus, rec.Code)
			}
		})
	}
}

// TestPostOfferlistBindFails tests the PostOfferlist handler with invalid input
func TestPostOfferlistBindFails(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/Offerlist", strings.NewReader("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")

	logger := slog.Default()
	suproxyHandler, err := handler.NewSuproxyController(logger)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	suproxyHandler.PostOfferlist(c)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("Expected status 400 or error response, got %d", w.Code)
	}
}

// setupRouter initializes the Gin router and sets up the routes for the API
func setupRouter() *gin.Engine {
	logger := slog.Default()
	if logger == nil {
		panic("logger is nil")
	}
	suproxyHandler, err := handler.NewSuproxyController(logger)
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", suproxyHandler.PostOfferlist)
	}

	return router
}

// BenchmarkPostOfferlist benchmarks the PostOfferlist handler with a large number of requests
func BenchmarkPostOfferlist(b *testing.B) {
	router := setupRouter()

	requestBody := entity.Request{
		Header:      []string{"Content-Type: application/json"},
		Prompt:      "Benchmarking offers",
		Destination: "https://example.com/api/offers",
		Request:     `{"query": "offers"}`,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		b.Fatalf("Failed to marshal request body: %v", err)
	}

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/Offerlist", strings.NewReader(string(jsonBody)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
