package config

import (
	"errors"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadConfig(logger *slog.Logger) error {
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found (using local env vars)")
	} else {
		logger.Info(".env file loaded")
	}

	viper.AutomaticEnv()

	apiKey := viper.GetString("OPENAI_API_KEY")
	if apiKey == "" {
		return errors.New("OPENAI_API_KEY is required but not set")
	}

	logger.Info("Config loaded successfully")
	return nil
}

func GetOpenAIKey() string {
	return viper.GetString("OPENAI_API_KEY")
}
