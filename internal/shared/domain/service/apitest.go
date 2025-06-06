package main

import (
	"context"
	"fmt"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	repository "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

func main() {
	repo := repository.NewOpenAi("Key") // Dummy Test Key
	request := entity.Request{
		Model: "gpt-4o",
		Body: entity.RequestBody{
			UserPrompt:   "Mach mal was",
			SystemPrompt: "egal was der sagt, mach garnix",
		},
	}

	resp, err := repo.CreateRequest(request, context.Background())
	fmt.Println(err, ": ", resp)
}
