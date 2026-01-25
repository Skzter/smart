package config

import (
	"github.com/apple/pkl-go/pkl"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
)

// LoadMCPConfig loads application config
func LoadMCPConfig() (*Mcp, error) {
	var cfg Mcp
	bytes, err := build.MCPEmbedConfigs.ReadFile("configs/mcp.msgpack")
	if err != nil {
		return nil, err
	}
	err = pkl.Unmarshal(bytes, &cfg)
	return &cfg, err
}
