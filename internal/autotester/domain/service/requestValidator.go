package service

/*
import (
	"context"
	//"fmt"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

type RequestValidationService interface {
	ValidateRequest(ctx context.Context, req sharedEntity.Request) (bool, error)
}

type requestValidationService struct {
	openAIService sharedService.OpenAI
	req           sharedEntity.Request
}
func NewRequestValidationService(openAIService sharedService.OpenAI, req sharedEntity.Request) RequestValidationService {
	return &requestValidationService{
		openAIService: openAIService,
		req:           req,
	}
}

func ValidateRequest(ctx context.Context, req sharedEntity.Request) (bool, error) {
	// Implement request validation logic here
	return true, nil
}



/*
 Files to connect:
- internal/autotester/domain/service/requestValidationService.go
- internal/autotester/domain/service/validatePrompt.go --> hier Service implementieren
- internal/autotester/domain/handler/requestHandler.go
- internal/autotester/domain/service/generatePrompt.go  --> Zielservice
*/
