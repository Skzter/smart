package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupAutotesterAPIMock uses Testcontainers to spin up a Prism mock server.
// Testcontainers manages the Docker lifecycle
//
// Parameters:
//   - t: used for logging and to register t.Cleanup() so the container is killed automatically.
//   - apiSpecPath: the path to the OpenAPI specification
//
// Returns the base URL where the tests can reach the mocked API.
func SetupAutotesterAPIMock(t *testing.T, apiSpecPath string) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Port 4010 is the hardcoded internal default port for Prism.
	const prismPort = "4010/tcp"
	ctx := context.Background()

	absApiSpecPath, err := filepath.Abs(apiSpecPath)
	require.NoError(t, err, "failed to resolve absolute path for API spec")

	autotesterContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "stoplight/prism:4.14.1",
			// Arguments for the prism binary inside the container:
			// - mock: Starts the mock server mode.
			// - -h: Host flag to specify the binding address.
			// - 0.0.0.0: Binds to all network interfaces inside the container.
			// - --errors: Returns 400/500 errors if requests don't match the spec.
			// - /api/AutotesterAPI.yaml: The spec file Prism uses to generate the mock.
			Cmd: []string{
				"mock",
				"-h",
				"0.0.0.0",
				"--errors",
				"/api/AutotesterAPI.yaml",
			},
			ExposedPorts: []string{prismPort},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      absApiSpecPath,
					ContainerFilePath: "/api/AutotesterAPI.yaml",
				},
			},
			// Wait until Prism logs that it is ready.
			WaitingFor: wait.ForLog("Prism is listening"),
		},
		Started: true,
	})
	require.NoError(t, err, "failed to start Prism container")

	// Ensures the container is terminated automatically after the test and its subtests.
	testcontainers.CleanupContainer(t, autotesterContainer)

	host, err := autotesterContainer.Host(ctx)
	require.NoError(t, err)

	// Map the internal port 4010 to the dynamically assigned host port.
	port, err := autotesterContainer.MappedPort(ctx, prismPort)
	require.NoError(t, err)

	return fmt.Sprintf("http://%s:%s", host, port.Port())
}
