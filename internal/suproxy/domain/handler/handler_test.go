package handler

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
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	suproxyentity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
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
			requestBody: `
				{"request": 
					"{\"apimode\":\"live\",\"id\":\"a0950be9-76ad-4fcb-932d-37660d10b1f8\",\"method\":\"search.getOfferList\",\"params\":[],\"requestSource\":\"pr.offerlist\"}",
				"header":{
					"Authorization": "Bearer asdfjsafjaölfaöfsal"
				},
				"destination":"https:://example.com",
				"prompt":""}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"httpstatuscode":200,"data":{"items":[{"offerid": 2814668548,"departuredate": "2025-07-07T00:00:0+0000",}]}}`,
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
		Header:      map[string]string{"Content-Type": "application/json"},
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
func setupController(tb testing.TB) *SuproxyController {
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

	mockConnector := mocks.NewMockOpenAIService(tb)
	mockConnector.
		On("Request", mock.Anything, mock.Anything).
		Return(&sharedentity.Response{Text: `{"valid": true, "reason": []}`}, nil)

	ctrl, err := NewSuproxyController(logger, cfg)
	if err != nil {
		tb.Fatalf("Failed to create controller: %v", err)
	}
	return ctrl
}
