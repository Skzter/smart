package errors

import (
	"errors"
)

// Frontend-facing errors for the users
// nolint: gochecknoglobals, staticcheck
var (
	//lint:ignore ST1005 deutsche sprache
	ErrInternalServer = errors.New("Interner Server Error - bitte nochmal versuchen")
	//lint:ignore ST1005 deutsche sprache
	ErrValidation = errors.New("Validation Error - bitte nochmal versuchen")
	//lint:ignore ST1005 deutsche sprache
	ErrGeneration = errors.New("Generate Error - bitte nochmal versuchen")

	ErrNilUserPrompt      = errors.New("request without user prompt")
	ErrNilSystemPrompt    = errors.New("request without system prompt")
	ErrNilModel           = errors.New("request without model")
	ErrNilRole            = errors.New("request without role")
	ErrEmptyResponseArray = errors.New("response contains no messages to choose from")
	ErrEmptyResponse      = errors.New("chosen response message is empty")
	ErrChatNotFound       = errors.New("chat not found")
)
