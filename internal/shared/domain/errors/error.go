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
)
