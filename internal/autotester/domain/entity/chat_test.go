package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

func validChat() *Chat {
	return &Chat{
		Id:            "chat123",
		UserId:        "user123",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastTest:      "test123",
		SystemPrompt:  "sys prompt",
		InitialPrompt: "usr prompt",
		Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
	}
}

// nolint:funlen
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *Chat
		wantErr bool
	}{
		{
			name:    "valid Chat",
			obj:     validChat(),
			wantErr: false,
		},
		{
			name: "empty id",
			obj: &Chat{
				Id:            "",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty userId",
			obj: &Chat{
				Id:            "chat123",
				UserId:        "",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty initialPrompt",
			obj: &Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty systemprompt",
			obj: &Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty messages",
			obj: &Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{},
			},
			wantErr: true,
		},
		{
			name: "message with empty body",
			obj: &Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: ""}},
			},
			wantErr: true,
		},
		{
			name: "message with empty id",
			obj: &Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "message with empty role",
			obj: &Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "", Body: "msg"}},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.obj.Validate()

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
