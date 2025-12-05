package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

func TestNewChat(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		messages []*Message
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
		initial  []*Message
		add      *entity.Message
		expected []*Message
	}{
		{
			name:    "add to empty",
			initial: nil,
			add:     &entity.Message{Id: "m1"},
			expected: []*Message{
				{Message: entity.Message{Id: "m1"}, Type: MessageTypeValidation},
			},
		},
		{
			name:    "append to existing",
			initial: []*Message{{Message: entity.Message{Id: "m1"}, Type: MessageTypeValidation}},
			add:     &entity.Message{Id: "m2"},
			expected: []*Message{
				{Message: entity.Message{Id: "m1"}, Type: MessageTypeValidation},
				{Message: entity.Message{Id: "m2"}, Type: MessageTypeValidation},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewChat("u", tt.initial)
			// satisfy Validate precondition
			c.LastAutoPlaywrightPrompt = "p"

			c.AddMessage(tt.add, MessageTypeValidation)

			assert.Len(t, c.Messages, len(tt.expected))
			for i := range tt.expected {
				assert.Equal(t, tt.expected[i], c.Messages[i])
			}
		})
	}
}

func TestFilter(t *testing.T) {
	base := []*Message{
		{Message: entity.Message{Id: "m0"}, Type: MessageTypeUser},
		{Message: entity.Message{Id: "m1"}, Type: MessageTypeValidation},
		{Message: entity.Message{Id: "m2"}, Type: MessageTypeGeneration},
		{Message: entity.Message{Id: "m3"}, Type: MessageTypeValidation},
		{Message: entity.Message{Id: "m4"}, Type: MessageTypeUser},
		{Message: entity.Message{Id: "m5"}, Type: MessageTypeValidation},
		{Message: entity.Message{Id: "m6"}, Type: MessageTypeGeneration},
	}

	tests := []struct {
		name         string
		messages     []*Message
		wantType     MessageType
		wantMessages []*entity.Message
	}{
		{
			name:         "no matches",
			messages:     []*Message{{Message: entity.Message{Id: "only gen"}, Type: MessageTypeGeneration}},
			wantType:     MessageTypeValidation,
			wantMessages: []*entity.Message{},
		},
		{
			name:     "multiple matches",
			messages: base,
			wantType: MessageTypeValidation,
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
			wantType:     MessageTypeGeneration,
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
