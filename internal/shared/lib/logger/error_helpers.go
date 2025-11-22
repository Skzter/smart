package logger

// Paket soll zentrale Helper für ein einheitliches Fehler-Logging bereit stellen.
// Klassifiziert interne Fehler automatisch in die drei definierten Kategorien
// (Validation, Generation, Internal) und loggt dabei sowohl die Originalmeldung
// als auch die gemappte öffentliche Fehlermeldung. Dadurch müssen Logging-Regeln
// nicht an hunderten Stellen manuell implementiert werden, sondern sind zentral

import (
	"log/slog"
	"strings"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

func getValidationKeywords() []string {
	return []string{
		"json", "decode", "unmarshal", "validation", "invalid", "missing", "required",
		"payload", "request", "field",
	}
}

func getGenerationKeywords() []string {
	return []string{
		"generate", "llm", "openai", "toolcall", "content", "model", "completion",
	}
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func classifyError(err error) error {
	if err == nil {
		return nil
	}

	s := strings.ToLower(err.Error())

	if containsAny(s, getValidationKeywords()...) {
		return errors.ErrValidation
	}

	generationKeywords := getGenerationKeywords()
	if containsAny(s, generationKeywords...) || (strings.Contains(s, "response") && strings.Contains(s, "parse")) {
		return errors.ErrGeneration
	}

	return errors.ErrInternalServer
}

// LogAndClassify logs the original error with contextual information
// parameter are the logger, original error and contextual information
// returns the classified error
func LogAndClassify(logger *slog.Logger, original error, ctx ...map[string]any) error {
	if original == nil {
		return nil
	}

	mapped := classifyError(original)

	attrs := []any{
		slog.String("mapped_error", mapped.Error()),
		slog.String("original_error", original.Error()),
	}

	if len(ctx) > 0 && ctx[0] != nil {
		for k, v := range ctx[0] {
			attrs = append(attrs, slog.Any(k, v))
		}
	}

	logger.Error("error occured",
		slog.Group("context", attrs...),
	)

	return mapped
}
