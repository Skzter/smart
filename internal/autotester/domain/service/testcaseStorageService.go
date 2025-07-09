package service

import "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"

type TestcaseStorageService interface {
	saveTestCase(entity.TestCase) error
}
