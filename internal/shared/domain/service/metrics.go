package service

import (
	"os"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// Metric key suffixes - combined with service namespace for full metric names
const (
	MetricRequestSuccess  = "request.success"
	MetricRequestError    = "request.error"
	MetricRequestDuration = "request.duration"
	MetricStatusCode      = "status_code"
)

// MetricsService provides methods to track application metrics via DogStatsD.
type MetricsService interface {
	// IncRequestSuccess increments the successful request counter.
	IncRequestSuccess()
	// IncRequestError increments the error request counter with an error type tag.
	IncRequestError(errorType string)
	// RecordRequestDuration records the duration of a request.
	RecordRequestDuration(duration time.Duration)
	// RecordStatusCode records an HTTP status code.
	RecordStatusCode(statusCode int)
	// Close closes the StatsD client connection.
	Close() error
}

type metricsService struct {
	client *statsd.Client
}

// NewMetricsService creates a new MetricsService with DogStatsD client.
// serviceName is used as the namespace prefix (e.g., "suproxy" -> "smart.suproxy.").
// It reads DD_AGENT_HOST and DD_DOGSTATSD_PORT from environment variables.
func NewMetricsService(serviceName string) (MetricsService, error) {
	addr := os.Getenv("DD_AGENT_HOST")
	if addr == "" {
		addr = "localhost"
	}
	port := os.Getenv("DD_DOGSTATSD_PORT")
	if port == "" {
		port = "8125"
	}

	client, err := statsd.New(addr+":"+port,
		statsd.WithNamespace("smart."+serviceName+"."),
		statsd.WithTags([]string{
			"service:" + serviceName,
			"env:" + os.Getenv("DD_ENV"),
			"version:" + os.Getenv("DD_VERSION"),
		}),
	)
	if err != nil {
		return nil, err
	}

	return &metricsService{client: client}, nil
}

func (m *metricsService) IncRequestSuccess() {
	_ = m.client.Incr(MetricRequestSuccess, nil, 1)
}

func (m *metricsService) IncRequestError(errorType string) {
	_ = m.client.Incr(MetricRequestError, []string{"error_type:" + errorType}, 1)
}

func (m *metricsService) RecordRequestDuration(duration time.Duration) {
	_ = m.client.Timing(MetricRequestDuration, duration, nil, 1)
}

func (m *metricsService) RecordStatusCode(statusCode int) {
	_ = m.client.Histogram(MetricStatusCode, float64(statusCode), nil, 1)
}

func (m *metricsService) Close() error {
	return m.client.Close()
}
