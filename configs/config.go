package config

import (
	"errors"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ConfigLoader struct {
	logger *slog.Logger
}

func NewConfigLoader(logger *slog.Logger) *ConfigLoader {
	return &ConfigLoader{
		logger: logger,
	}
}

func (cl ConfigLoader) LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		cl.logger.Warn("No .env file found (using local env vars)")
	} else {
		cl.logger.Info(".env file loaded")
	}

	viper.AutomaticEnv()

	apiKey := viper.GetString("OPENAI_API_KEY")
	if apiKey == "" {
		return errors.New("OPENAI_API_KEY is required but not set")
	}

	cl.logger.Info("Config loaded successfully")
	return nil
}

func GetOpenAIKey() string {
	return viper.GetString("OPENAI_API_KEY")
}
