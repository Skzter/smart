package config

import (
	"os"

	"github.com/apple/pkl-go/pkl"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
)

// LoadConfig loads the configuration for the autotester from an embedded file.
func LoadConfig() (*Config, error) {
	var cfg Config
	bytes, err := build.AutotesterEmbedConfigs.ReadFile("configs/autotester.msgpack")
	if err != nil {
		return nil, err
	}
	err = pkl.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, err
	}

	// Overwrite prompt if file exists (for prompt optimizer)
	// In production/docker, this file might not exist, so we fallback to the PKL value.
	promptBytes, err := os.ReadFile("configs/prompts/autoplaywright_prompt.txt")
	if err == nil {
		cfg.Prompts.AutoPlaywrightPromptT = string(promptBytes)
	}

	return &cfg, nil
}
