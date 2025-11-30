package main

import (
	"os"

	ddotel "github.com/DataDog/dd-trace-go/v2/ddtrace/opentelemetry"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
)

func getEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	return "development"
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	env := getEnvironment()
	if os.Getenv("DD_SERVICE") == "" {
		if err := os.Setenv("DD_SERVICE", "autotester"); err != nil {
			panic(err)
		}
	}
	if os.Getenv("DD_VERSION") == "" {
		if err := os.Setenv("DD_VERSION", build.Version); err != nil {
			panic(err)
		}
	}
	if os.Getenv("DD_ENV") == "" {
		if err := os.Setenv("DD_ENV", env); err != nil {
			panic(err)
		}
	}

	tracerProvider := ddotel.NewTracerProvider()
	defer func() {
		if err := tracerProvider.Shutdown(); err != nil {
			panic(err)
		}
	}()

	otel.SetTracerProvider(tracerProvider)

	tracer := otel.Tracer(os.Getenv("DD_SERVICE"))

	router, err := InitializeApp(cfg, tracer)
	if err != nil {
		panic(err)
	}

	if err := router.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
