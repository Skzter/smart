package config

import (
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
	return &cfg, err
}
