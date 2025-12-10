package entity

import (
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
)

// RunTestTool coordinates locating and executing test files.
// SaveLocalService provides access to local test files.
// DockerService runs tests inside a Docker container and reads logs.
type RunTestTool struct {
	SaveLocalService service.TestcaseLocalStorageService
	DockerService    service.Docker
}
