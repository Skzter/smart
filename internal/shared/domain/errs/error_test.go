package errs

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

func TestErrorStruct(t *testing.T) {
	msg := "this is a message"
	err := errors.New("error it in")
	customErr := Error{
		Message:    msg,
		Underlying: err,
		Type:       Public,
	}

	tests := []struct {
		name             string
		error            Error
		wantedMessage    string
		wantedUnderlying error
	}{
		{
			name:             "Test Unwrap() and Error()",
			error:            customErr,
			wantedMessage:    msg,
			wantedUnderlying: err,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.error.Error()
			err := tt.error.Unwrap()
			if !errors.Is(err, tt.wantedUnderlying) {
				t.Fatalf("Error doesnt match: got => %v, wanted => %v", err, tt.wantedUnderlying)
			}
			if msg != tt.wantedMessage {
				t.Fatalf("Message doesnt match: got => %s, wanted => %s", msg, tt.wantedMessage)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	errPublic := errors.New("public error")
	publicCustomErr := &Error{
		Message:    fmt.Sprintf("wild error: %v", errPublic),
		Underlying: errPublic,
		Type:       Public,
	}

	tests := []struct {
		name         string
		givenError   error
		wantedError  error
		wantedStatus int
		wantErr      bool
	}{
		{
			name:         "nil error",
			givenError:   nil,
			wantedError:  nil,
			wantedStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "custom: repo error: empty response",
			givenError:   ErrEmptyResponse,
			wantedError:  ErrInternalServer,
			wantedStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "custom: service error: nil ctx",
			givenError:   &assert.NotNilError{Message: "assert failed"},
			wantedError:  ErrInternalServer,
			wantedStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "custom: public error",
			givenError:   publicCustomErr,
			wantedError:  publicCustomErr,
			wantedStatus: http.StatusBadRequest,
			wantErr:      true,
		},
		{
			name:         "public error",
			givenError:   errPublic,
			wantedError:  errPublic,
			wantedStatus: http.StatusBadRequest,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := HandleError(tt.givenError)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleError() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !errors.Is(err, tt.wantedError) {
				t.Errorf("gave back wrong error: got: %v, wanted error: %v", err, tt.wantedError)
			}
			if status != tt.wantedStatus {
				t.Errorf("gave back wrong code: got %d, wanted %d", status, tt.wantedStatus)
			}
		})
	}
}
