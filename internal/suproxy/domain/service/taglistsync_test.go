package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

// nolint:funlen
func TestNewTaglistSync(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name              string
		logger            *slog.Logger
		expectError       bool
		expectedError     error
		wantError         bool
		mockResponse      []any
		mockResponseStore []any
	}{
		{
			name:         "success - taglist loaded correctly",
			logger:       logger,
			expectError:  false,
			wantError:    false,
			mockResponse: []any{&sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG1", Description: "TAG1"}, {Name: "TAG2", Description: "TAG2"}}}, nil},
		},
		{
			name:        "error - logger is nil",
			logger:      nil,
			expectError: true,
			wantError:   true,
		},
		{
			name:         "error - taglistservice fails",
			logger:       logger,
			expectError:  true,
			wantError:    true,
			mockResponse: []any{nil, errors.New("taglist error")},
		},
		{
			name:              "success - taglist was nil and is the default taglist",
			logger:            logger,
			expectError:       false,
			wantError:         false,
			mockResponse:      []any{nil, nil},
			mockResponseStore: []any{nil},
		},
		{
			name:              "error - taglist was empty, storing the default fails",
			logger:            logger,
			expectError:       true,
			wantError:         false,
			mockResponse:      []any{&sharedEntity.TagList{}, nil},
			mockResponseStore: []any{errors.New("storage error")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv := mocks.NewMockTaglistStorage(t)
			if tt.mockResponse != nil {
				serv.On("GetTaglist", mock.Anything).Return(tt.mockResponse...)
			}
			if tt.mockResponseStore != nil {
				serv.On("StoreTaglist", mock.Anything, mock.Anything).Return(tt.mockResponseStore...)
			}
			sync, err := NewTaglistSync(tt.logger, serv)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, sync)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sync)
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
			wantErr:      false,
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
func TestGetCurrentTaglist(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name        string
		initialList sharedEntity.TagList
		wantLen     int
	}{
		{
			name: "non-empty taglist returns correct copy",
			initialList: sharedEntity.TagList{Tags: []sharedEntity.Tag{
				{Name: "TAG1", Description: "TAG1"},
				{Name: "TAG2", Description: "TAG2"},
			}},
			wantLen: 2,
		},
		{
			name:        "empty taglist returns empty copy",
			initialList: sharedEntity.TagList{Tags: []sharedEntity.Tag{}},
			wantLen:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSrv := mocks.NewMockTaglistStorage(t)
			srv := taglistSync{
				logger:         logger,
				taglistService: mockSrv,
				tagList:        &tc.initialList,
			}

			got := srv.GetCurrentTaglist()

			// gleiche Länge & Werte prüfen
			if len(got.Tags) != tc.wantLen {
				t.Fatalf("expected len %d, got %d", tc.wantLen, len(got.Tags))
			}
			if !equalTags(got.Tags, tc.initialList.Tags) {
				t.Fatalf("expected tags %+v, got %+v", tc.initialList.Tags, got.Tags)
			}

			// sicherstellen, dass es eine Kopie ist
			if tc.wantLen > 0 && &got.Tags[0] == &tc.initialList.Tags[0] {
				t.Fatal("expected copy, but got reference to original")
			}

			// prüfen, dass Änderungen am Rückgabewert das Original nicht verändern
			if tc.wantLen > 0 {
				originalName := tc.initialList.Tags[0].Name
				got.Tags[0].Name = "CHANGED"
				if tc.initialList.Tags[0].Name != originalName {
					t.Fatal("modifying returned taglist changed the original")
				}
			}
		})
	}
}

// equalTags vergleicht zwei Slices von Tags unabhängig vom Pointer
func equalTags(a, b []sharedEntity.Tag) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Description != b[i].Description {
			return false
		}
	}
	return true
}
