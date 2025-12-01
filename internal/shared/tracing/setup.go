package tracing

import (
	"os"

	ddotel "github.com/DataDog/dd-trace-go/v2/ddtrace/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
)

// Environment returns the current runtime environment, falling back to development.
func Environment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	return "development"
}

// Setup configures OpenTelemetry/DataDog tracing for the provided service name.
// It returns a tracer and a shutdown function that callers should defer.
func Setup(serviceName string) (trace.Tracer, func() error, error) {
	env := Environment()

	if err := ensureEnv("DD_SERVICE", serviceName); err != nil {
		return nil, nil, err
	}
	if err := ensureEnv("DD_VERSION", build.Version); err != nil {
		return nil, nil, err
	}
	if err := ensureEnv("DD_ENV", env); err != nil {
		return nil, nil, err
	}

	tracerProvider := ddotel.NewTracerProvider()
	otel.SetTracerProvider(tracerProvider)

	shutdown := func() error {
		return tracerProvider.Shutdown()
	}

	tracer := otel.Tracer(os.Getenv("DD_SERVICE"))
	return tracer, shutdown, nil
}

func ensureEnv(key, value string) error {
	if os.Getenv(key) != "" {
		return nil
	}
	return os.Setenv(key, value)
}
