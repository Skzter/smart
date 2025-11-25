package logger

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

/*
Hilfsfunktion: erstellt einen Logger, der in einen Buffer schreibt,
damit das Testen von Logs möglich ist.
*/
func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(handler)
}

func TestLogAndClassify_Validation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	// Ursprungsfehler = Validation-Typ (enthält "invalid")
	original := errors.New("invalid payload")

	mapped := LogAndClassify(logger, original, map[string]any{
		"endpoint": "/test",
	})

	if mapped != sharedErrors.ErrValidation {
		t.Errorf("expected %v, got %v", sharedErrors.ErrValidation, mapped)
	}

	log := buf.String()

	if !strings.Contains(log, "mapped_error") || !strings.Contains(log, "Validation") {
		t.Errorf("log does not contain mapped error: %s", log)
	}

	if !strings.Contains(log, "original_error") || !strings.Contains(log, "invalid payload") {
		t.Errorf("log does not contain original error: %s", log)
	}

	if !strings.Contains(log, "endpoint") {
		t.Errorf("log does not contain context data: %s", log)
	}
}

func TestLogAndClassify_Generation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	original := errors.New("failed to generate toolcall")

	mapped := LogAndClassify(logger, original)

	if mapped != sharedErrors.ErrGeneration {
		t.Errorf("expected %v, got %v", sharedErrors.ErrGeneration, mapped)
	}

	log := buf.String()
	if !strings.Contains(log, "Generate") {
		t.Errorf("expected generation category in log, got: %s", log)
	}
}

func TestLogAndClassify_Internal(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	original := errors.New("database timeout occurred")

	mapped := LogAndClassify(logger, original)

	if mapped != sharedErrors.ErrInternalServer {
		t.Errorf("expected %v, got %v", sharedErrors.ErrInternalServer, mapped)
	}

	log := buf.String()
	if !strings.Contains(log, "Interner Server Error") {
		t.Errorf("expected internal category in log, got: %s", log)
	}
}

func TestLogAndClassify_NoError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	result := LogAndClassify(logger, nil)

	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}

	if buf.Len() != 0 {
		t.Errorf("no log should be written if err is nil")
	}
}
