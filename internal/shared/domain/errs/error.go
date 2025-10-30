package errs

import (
	"errors"
)

// ErrorType for referencing Public or Private Errors -> Public for frontend, Private is backend only
type ErrorType int

// Enum for Error Types
const (
	Public ErrorType = iota
	Private
)

// Error is custom error type for validation service
type Error struct {
	Underlying error
	Message    string
	Type       ErrorType
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Underlying
}

// Public errors for the user
//
//nolint:gochecknoglobals,staticcheck
var (
	// ErrInternalServer is an frontend-facing error, when the request to openai fails or any other part of the repository
	//lint:ignore ST1005 "Satzanfang" vom deutschen Error großgeschrieben
	ErrInternalServer error = errors.New("Interner Server Error - bitte nochmal versuchen")
)

// Errors for the Repository
var (
	ErrNilUserPrompt      = &Error{Type: Private, Message: "request without user prompt"}
	ErrNilSystemPrompt    = &Error{Type: Private, Message: "request without system prompt"}
	ErrNilModel           = &Error{Type: Private, Message: "request without model"}
	ErrEmptyResponseArray = &Error{Type: Private, Message: "REPO openai error: response contains no messages to choose from"}
	ErrEmptyResponse      = &Error{Type: Private, Message: "REPO openai error: chosen response message is empty"}
	ErrOpenAI             = &Error{Type: Private, Message: "REPO openai error: request to the server failed"}
)
