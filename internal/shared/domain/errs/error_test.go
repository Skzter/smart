package errs

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
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
		name              string
		givenError        error
		wantedErrorString string
		wantedStatus      int
		wantErr           bool
	}{
		{
			name:              "nil error",
			givenError:        nil,
			wantedErrorString: "",
			wantedStatus:      http.StatusOK,
			wantErr:           false,
		},
		{
			name:              "custom: repo error: empty response",
			givenError:        ErrEmptyResponse,
			wantedErrorString: ErrInternalServer.Error(),
			wantedStatus:      http.StatusInternalServerError,
			wantErr:           true,
		},
		{
			name:              "custom: public error",
			givenError:        publicCustomErr,
			wantedErrorString: publicCustomErr.Underlying.Error(),
			wantedStatus:      http.StatusBadRequest,
			wantErr:           true,
		},
		{
			name:              "public error",
			givenError:        errPublic,
			wantedErrorString: errPublic.Error(),
			wantedStatus:      http.StatusBadRequest,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := HandleError(tt.givenError)
			if msg != tt.wantedErrorString {
				t.Errorf("gave back wrong message: got %s, wanted %s", msg, tt.wantedErrorString)
			}
			if status != tt.wantedStatus {
				t.Errorf("gave back wrong code: got %d, wanted %d", status, tt.wantedStatus)
			}
		})
	}
}
