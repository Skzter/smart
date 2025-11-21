package logger

// Paket soll zentrale Helper für ein einheitliches Fehler-Logging bereit stellen.
// Klassifiziert interne Fehler automatisch in die drei definierten Kategorien
// (Validation, Generation, Internal) und loggt dabei sowohl die Originalmeldung
// als auch die gemappte öffentliche Fehlermeldung. Dadurch müssen Logging-Regeln
// nicht an hunderten Stellen manuell implementiert werden, sondern sind zentral

/*

import (
	"log/slog"
	"strings"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

var (
	validationKeywords = []string{
		"json", "decode", "unmarshal", "validation", "invalid", "missing", "required",
		"payload", "request", "field",
	}
	generationKeywords = []string{
		"generate", "llm", "openai", "toolcall", "content", "model", "completion",
	}
)

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

	if containsAny(s, validationKeywords...) {
		return errors.ErrValidation
	}

	if containsAny(s, generationKeywords...) || (strings.Contains(s, "response") && strings.Contains(s, "parse")) {
		return errors.ErrGeneration
	}

	return errors.ErrInternalServer
}

func LogAndClassify(original error, ctx ...map[string]any) error {

...

}

*/
