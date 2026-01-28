package repository

import (
	"testing"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name string
		tags *sharedEntity.TagList
		ts   string
		want string
	}{
		{
			name: "normalizes tags",
			tags: &sharedEntity.TagList{
				Tags: []sharedEntity.Tag{
					{Name: "Foo"},
					{Name: " Bar "},
				},
			},
			ts:   "123",
			want: "foo-bar-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateKey(tt.tags, tt.ts); got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestValidateDatabaseEntry(t *testing.T) {
	valid := entity.DatabaseEntry{
		Request: entity.Request{
			Header:      map[string]string{"h": "v"},
			Tags:        "tag",
			Destination: "dest",
			Body:        "body",
		},
		Response: entity.Response{Response: "ok"},
		Tags: &sharedEntity.TagList{
			Tags: []sharedEntity.Tag{{Name: "x"}},
		},
	}

	tests := []struct {
		name    string
		entry   entity.DatabaseEntry
		wantErr bool
	}{
		{"valid", valid, false},
		{"missing tags", entity.DatabaseEntry{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabaseEntry(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
