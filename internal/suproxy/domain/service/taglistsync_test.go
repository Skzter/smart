package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

// nolint:funlen
func TestNewTaglistSync(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	taglistError := errors.New("Taglist error")

	tests := []struct {
		name          string
		logger        *slog.Logger
		expectError   bool
		expectedError error
		wantError     bool
		mockResponse  []any
	}{
		{
			name:          "success - taglist loaded correctly",
			logger:        logger,
			expectError:   false,
			expectedError: nil,
			wantError:     false,
			mockResponse:  []any{&sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG1", Description: "TAG1"}, {Name: "TAG2", Description: "TAG2"}}}, nil},
		},
		{
			name:          "error - logger is nil",
			logger:        nil,
			expectError:   true,
			expectedError: nil,
			wantError:     true,
		},
		{
			name:          "error - taglistservice fails",
			logger:        logger,
			expectError:   true,
			expectedError: taglistError,
			wantError:     true,
			mockResponse:  []any{nil, taglistError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := mocks.NewMockTaglistStorage(t)
			if tt.mockResponse != nil {
				serv.On("GetTaglist", mock.Anything).Return(tt.mockResponse...)
			}
			sync, err := NewTaglistSync(tt.logger, serv)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected => %s, got => nil", "not nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected error to be nil, got => %s", err.Error())
				}
				if sync == nil {
					t.Fatal("expected sync to be not nil")
				}
			}
		})
	}
}

func TestSyncTaglist(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	originalTaglist := sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG1", Description: "TAG1"}, {Name: "TAG2", Description: "TAG2"}}}
	differentTaglist := sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG3", Description: "TAG3"}, {Name: "TAG4", Description: "TAG4"}, {Name: "TAG5", Description: "TAG5"}}}
	tests := []struct {
		name         string
		ctx          context.Context
		taglist      sharedEntity.TagList
		mockResponse []any
		wantErr      bool
	}{
		{
			name:         "nil context",
			ctx:          nil,
			taglist:      originalTaglist,
			mockResponse: nil,
			wantErr:      true,
		},
		{
			name:         "given taglist is empty",
			ctx:          context.Background(),
			taglist:      sharedEntity.TagList{},
			mockResponse: nil,
			wantErr:      true,
		},
		{
			name:         "equal taglist",
			ctx:          context.Background(),
			taglist:      originalTaglist,
			mockResponse: nil,
			wantErr:      false,
		},
		{
			name:         "different taglist, successful upload",
			ctx:          context.Background(),
			taglist:      differentTaglist,
			mockResponse: []any{nil},
			wantErr:      false,
		},
		{
			name:         "different taglist, upload failed",
			ctx:          context.Background(),
			taglist:      differentTaglist,
			mockResponse: []any{errors.New("upload failed")},
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newTags := append([]sharedEntity.Tag(nil), originalTaglist.Tags...)
			mockTagListstorageSrv := mocks.NewMockTaglistStorage(t)
			srv := taglistSync{
				logger:         logger,
				taglistService: mockTagListstorageSrv,
				tagList:        &sharedEntity.TagList{Tags: newTags},
			}
			if tc.mockResponse != nil {
				mockTagListstorageSrv.On("StoreTaglist", mock.Anything, mock.Anything).Return(tc.mockResponse...)
			}
			err := srv.SyncTaglist(tc.ctx, &tc.taglist)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected nil, got => %s", err.Error())
				}
			}
		})
	}
}
