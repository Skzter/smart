//go:build unittest

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	otelTrace "go.opentelemetry.io/otel/trace"
)

// TestTracingSmokeTest verifies that the tracing infrastructure works correctly
// nolint:errcheck,funlen
func TestTracingSmokeTest(t *testing.T) {
	t.Run("span creation and completion", func(t *testing.T) {
		// Setup in-memory exporter to capture spans
		exporter := tracetest.NewInMemoryExporter()
		tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("smoke-test")
		ctx := context.Background()

		// Create and end a span
		_, span := tracer.Start(ctx, "test.Operation")
		span.End()

		// Verify span was recorded
		spans := exporter.GetSpans()
		require.Len(t, spans, 1, "expected exactly one span")
		assert.Equal(t, "test.Operation", spans[0].Name)
		assert.True(t, spans[0].EndTime.After(spans[0].StartTime), "end time should be after start time")
	})

	t.Run("error recording", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("smoke-test")
		ctx := context.Background()

		// Create span and record error
		_, span := tracer.Start(ctx, "test.ErrorOperation")
		testErr := errors.New("test error")
		span.RecordError(testErr)
		span.SetStatus(codes.Error, "operation failed")
		span.End()

		// Verify error was recorded
		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status.Code)
		assert.Equal(t, "operation failed", spans[0].Status.Description)

		// Verify error event was recorded
		require.NotEmpty(t, spans[0].Events)
		foundErrorEvent := false
		for _, event := range spans[0].Events {
			if event.Name == "exception" {
				foundErrorEvent = true
				break
			}
		}
		assert.True(t, foundErrorEvent, "expected error event to be recorded")
	})

	t.Run("success status", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("smoke-test")
		ctx := context.Background()

		_, span := tracer.Start(ctx, "test.SuccessOperation")
		span.SetStatus(codes.Ok, "")
		span.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Ok, spans[0].Status.Code)
	})

	t.Run("span attributes", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("smoke-test")
		ctx := context.Background()

		_, span := tracer.Start(ctx, "test.AttributeOperation")
		span.SetAttributes(
			attribute.String("bucket", "test-bucket"),
			attribute.Int("size_bytes", 1024),
		)
		span.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)

		attrs := spans[0].Attributes
		assert.Contains(t, attrs, attribute.String("bucket", "test-bucket"))
		assert.Contains(t, attrs, attribute.Int("size_bytes", 1024))
	})

	t.Run("span events", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("smoke-test")
		ctx := context.Background()

		_, span := tracer.Start(ctx, "test.EventOperation")
		span.AddEvent("file uploaded", otelTrace.WithAttributes(
			attribute.String("key", "data.parquet"),
		))
		span.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		require.Len(t, spans[0].Events, 1)
		assert.Equal(t, "file uploaded", spans[0].Events[0].Name)
	})

	t.Run("nested spans with context propagation", func(t *testing.T) {
		exporter := tracetest.NewInMemoryExporter()
		tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
		defer tp.Shutdown(context.Background())

		tracer := tp.Tracer("smoke-test")
		ctx := context.Background()

		// Parent span
		ctx, parentSpan := tracer.Start(ctx, "parent.Operation")

		// Child span - should inherit parent context
		_, childSpan := tracer.Start(ctx, "child.Operation")
		childSpan.End()

		parentSpan.End()

		spans := exporter.GetSpans()
		require.Len(t, spans, 2)

		// Find parent and child
		var parent, child *tracetest.SpanStub
		for i := range spans {
			switch spans[i].Name {
			case "parent.Operation":
				parent = &spans[i]
			case "child.Operation":
				child = &spans[i]
			}
		}

		require.NotNil(t, parent, "parent span not found")
		require.NotNil(t, child, "child span not found")

		// Verify parent-child relationship
		assert.Equal(t, parent.SpanContext.TraceID(), child.SpanContext.TraceID(), "spans should share trace ID")
		assert.Equal(t, parent.SpanContext.SpanID(), child.Parent.SpanID(), "child's parent should be parent span")
	})

	t.Run("global tracer provider fallback", func(t *testing.T) {
		// Verify otel.Tracer() returns a valid tracer even without explicit setup
		tracer := otel.Tracer("fallback-test")
		assert.NotNil(t, tracer)

		ctx := context.Background()
		_, span := tracer.Start(ctx, "noop.Operation")
		assert.NotNil(t, span)
		span.End() // Should not panic
	})
}

// TestTracingIntegrationPattern verifies the tracing pattern used in services
// nolint:errcheck
func TestTracingIntegrationPattern(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	defer tp.Shutdown(context.Background())

	tracer := tp.Tracer("service-test")

	// Simulate the pattern used in generatePrompt.go, openaiService.go, etc.
	simulateServiceCall := func(ctx context.Context, shouldFail bool) error {
		_, span := tracer.Start(ctx, "myService.DoWork")
		defer span.End()

		if shouldFail {
			err := errors.New("simulated failure")
			span.RecordError(err)
			span.SetStatus(codes.Error, "operation failed")
			return err
		}

		span.SetStatus(codes.Ok, "")
		return nil
	}

	t.Run("successful service call", func(t *testing.T) {
		exporter.Reset()
		err := simulateServiceCall(context.Background(), false)

		assert.NoError(t, err)
		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Ok, spans[0].Status.Code)
	})

	t.Run("failed service call", func(t *testing.T) {
		exporter.Reset()
		err := simulateServiceCall(context.Background(), true)

		assert.Error(t, err)
		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status.Code)
		assert.NotEmpty(t, spans[0].Events, "error event should be recorded")
	})
}
