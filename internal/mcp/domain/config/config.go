package config

/* import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name    string         `yaml:"name"`
	Version string         `yaml:"version"`
	Schema  string         `yaml:"schema"`
	Models  []ModelConfig  `yaml:"models"`
	Context ContextConfig  `yaml:"context"`
	Servers []ServerConfig `yaml:"mcpServers"`
}

type ModelConfig struct {
	Name     string   `yaml:"name"`
	Provider string   `yaml:"provider"`
	Model    string   `yaml:"model"`
	Roles    []string `yaml:"roles"`

	DefaultCompletionOptions struct {
		MaxTokens int `yaml:"maxTokens"`
	} `yaml:"defaultCompletionOptions"`
}

type ContextConfig struct {
	Provider string `yaml:"provider"`
}

type ServerConfig struct {
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Cwd     string            `yaml:"cwd"`
	Env     map[string]string `yaml:"env"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse YAML: %w", err)
	}

	return &cfg, nil
}
*/
