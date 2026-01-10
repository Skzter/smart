//go:build integration

package tools

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Smoke test: this test verifies that the started Testcontainers-based
// Prism mock container is up and reachable, and that it survives subtests.
//
// NOTE: This test is intended only for Proof of Concept (PoC) to verify the
// infrastructure setup and should be removed after acceptance.
func TestSetupIntegrationEnvironment_AutotesterContainerIsReachable(t *testing.T) {
	baseURL := SetupAutotesterAPIMock(t, "../../../../api/AutotesterAPI.yaml")
	t.Logf("Autotester baseURL: %s", baseURL)

	tests := []struct {
		name string
	}{
		{name: "First request"},
		{name: "Second request"},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			resp, err := client.Get(baseURL + "/api/v1/template")
			t.Logf("GET %s/api/v1/template (after %s)", baseURL, time.Since(start))

			require.NoError(t, err)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
			require.NoError(t, readErr)
			t.Logf("HTTP status: %s", resp.Status)
			t.Logf("Body (first %d bytes): %q", len(bodyBytes), string(bodyBytes))

			// 200 = Template vorhanden
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
