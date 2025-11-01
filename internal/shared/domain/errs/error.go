package errs

import (
	"errors"
	"net/http"
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

var (
	// ErrInternalServer is an frontend-facing error, when the request to openai fails or any other part of the repository
	ErrInternalServer = &Error{Type: Public, Message: "Interner Server Error - bitte nochmal versuchen"}
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

// HandleError handles the errors for the validate and generate functions
// takes the given error and checks if it is a validation or generation error and looks further if its from there or deeper in the system
// if source found decides if it can return this error or a generic error because the error message may expose sensitive data
func HandleError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	var customError *Error

	if errors.As(err, &customError) {
		switch customError.Type {
		case Private:
			return http.StatusInternalServerError, unwrap(ErrInternalServer)
		case Public:
			return http.StatusBadRequest, unwrap(customError)
		}
	}
	return http.StatusBadRequest, unwrap(err)
}

func unwrap(err error) string {
	unwrappedError := errors.Unwrap(err)
	// checks if the error need the unwrapped
	// Unwrap() returns nil if nothing needs to be unwrapped
	if unwrappedError != nil {
		err = unwrappedError
	}
	return err.Error()
}
