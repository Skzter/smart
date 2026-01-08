package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupIntegrationEnvironment starts a Testcontainers container from an Autotester image for integration tests.
// TODO: It currently integrates against a real S3 backend (not ideal). Also, the production image rewrites API paths,
// which is not reflected in the MCP server, making local testing hard.
func SetupIntegrationEnvironment(t *testing.T) string {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	autotesterContainer := setupAutotesterContainer(t)

	host, _ := autotesterContainer.Host(ctx)
	port, _ := autotesterContainer.MappedPort(ctx, "8081")
	return "http://" + host + ":" + port.Port()
}

func setupAutotesterContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()

	autotesterContainer, err := testcontainers.Run(
		ctx,
		"gitlab.dit.htwk-leipzig.de:5050/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/autotester:latest",
		testcontainers.WithExposedPorts("8081/tcp"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("8081/tcp")),
	)
	require.NoError(t, err, "Failed to start Autotester container")
	testcontainers.CleanupContainer(t, autotesterContainer)

	return autotesterContainer
}
