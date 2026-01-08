//go:build integration

package tools

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Smoke test: this test currently exists mainly to verify that the started Testcontainers-based
// Autotester container is up and reachable from the test runtime.
func TestSetupIntegrationEnvironment_AutotesterContainerIsReachable(t *testing.T) {
	start := time.Now()
	baseURL := SetupIntegrationEnvironment(t)
	t.Logf("Autotester baseURL: %s", baseURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(baseURL + "/v1/template") // path without /api because prod version is build with replace_api_paths
	t.Logf("GET %s/v1/template (after %s)", baseURL, time.Since(start))
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	require.NoError(t, readErr)
	t.Logf("HTTP status: %s", resp.Status)
	t.Logf("Body (first %d bytes): %q", len(bodyBytes), string(bodyBytes))

	// 200 = Template vorhanden;
	require.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}
