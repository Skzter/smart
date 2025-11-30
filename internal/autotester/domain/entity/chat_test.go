package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// nolint:funlen
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		obj     *Chat
		wantErr bool
	}{
		{
			name: "valid Chat",
			obj: &Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []Message{{Message: entity.Message{Id: "m1"}, Type: TypeValidation}},
			},
			wantErr: false,
		},
		{
			name: "empty id",
			obj: &Chat{
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []Message{{Message: entity.Message{Id: "m1"}, Type: TypeValidation}},
			},
			wantErr: true,
		},
		{
			name: "empty userId",
			obj: &Chat{
				Id:                       "chat123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []Message{{Message: entity.Message{Id: "m1"}, Type: TypeValidation}},
			},
			wantErr: true,
		},
		{
			name: "empty Messages",
			obj: &Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []Message{},
			},
			wantErr: true,
		},
		{
			name: "updatedAt zero",
			obj: &Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Time{},
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []Message{{Message: entity.Message{Id: "m1"}, Type: TypeValidation}},
			},
			wantErr: true,
		},
		{
			name: "createdAt zero",
			obj: &Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Time{},
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []Message{{Message: entity.Message{Id: "m1"}, Type: TypeValidation}},
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

func TestNewChat(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		messages []Message
	}{
		{
			name: "happy path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat := NewChat(tt.id, tt.messages)

			assert.NotNil(t, chat)
		})
	}
}

func TestAddMessage(t *testing.T) {
	tests := []struct {
		name     string
		initial  []Message
		add      *entity.Message
		expected []Message
	}{
		{
			name:    "add to empty",
			initial: nil,
			add:     &entity.Message{Id: "m1"},
			expected: []Message{
				{Message: entity.Message{Id: "m1"}, Type: TypeValidation},
			},
		},
		{
			name:    "append to existing",
			initial: []Message{{Message: entity.Message{Id: "m1"}, Type: TypeValidation}},
			add:     &entity.Message{Id: "m2"},
			expected: []Message{
				{Message: entity.Message{Id: "m1"}, Type: TypeValidation},
				{Message: entity.Message{Id: "m2"}, Type: TypeValidation},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewChat("u", tt.initial)
			// satisfy Validate precondition
			c.LastAutoPlaywrightPrompt = "p"

			c.AddMessage(tt.add, TypeValidation)

			assert.Len(t, c.Messages, len(tt.expected))
			for i := range tt.expected {
				assert.Equal(t, tt.expected[i], c.Messages[i])
				assert.Equal(t, tt.expected[i], c.Messages[i])
			}
		})
	}
}

func TestFilter(t *testing.T) {
	base := []Message{
		{Message: entity.Message{Id: "m0"}, Type: TypeAny},
		{Message: entity.Message{Id: "m1"}, Type: TypeValidation},
		{Message: entity.Message{Id: "m2"}, Type: TypeGeneration},
		{Message: entity.Message{Id: "m3"}, Type: TypeValidation},
		{Message: entity.Message{Id: "m4"}, Type: TypeAny},
		{Message: entity.Message{Id: "m5"}, Type: TypeValidation},
		{Message: entity.Message{Id: "m6"}, Type: TypeGeneration},
	}

	tests := []struct {
		name         string
		messages     []Message
		wantType     Type
		wantMessages []*entity.Message
	}{
		{
			name:         "no matches",
			messages:     []Message{{Message: entity.Message{Id: "only gen"}, Type: TypeGeneration}},
			wantType:     TypeValidation,
			wantMessages: []*entity.Message{},
		},
		{
			name:     "multiple matches",
			messages: base,
			wantType: TypeValidation,
			wantMessages: []*entity.Message{
				&base[0].Message,
				&base[1].Message,
				&base[3].Message,
				&base[4].Message,
				&base[5].Message,
			},
		},
		{
			name:         "single match",
			messages:     base[1:4],
			wantType:     TypeGeneration,
			wantMessages: []*entity.Message{&base[2].Message},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewChat("u", tt.messages)
			c.LastAutoPlaywrightPrompt = "p"

			got := c.Filter(tt.wantType)
			c.Filter(tt.wantType)
			assert.Len(t, got, len(tt.wantMessages))
			for i, msg := range tt.wantMessages {
				assert.Equal(t, msg, got[i])
			}
		})
	}
}
