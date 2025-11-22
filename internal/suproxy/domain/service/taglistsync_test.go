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
		mockResponse      []any
		mockResponseStore []any
	}{
		{
			name:         "success - taglist loaded correctly",
			logger:       logger,
			expectError:  false,
			mockResponse: []any{&sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG1", Description: "TAG1"}, {Name: "TAG2", Description: "TAG2"}}}, nil},
		},
		{
			name:        "error - logger is nil",
			logger:      nil,
			expectError: true,
		},
		{
			name:              "error - taglistservice fails but still continues",
			logger:            logger,
			expectError:       false,
			mockResponse:      []any{nil, errors.New("taglist error")},
			mockResponseStore: []any{nil},
		},
		{
			name:              "success - taglist was nil and is the default taglist",
			logger:            logger,
			expectError:       false,
			mockResponse:      []any{nil, nil},
			mockResponseStore: []any{nil},
		},
		{
			name:              "error - taglist was empty, storing the default fails",
			logger:            logger,
			expectError:       false,
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

// nolint:funlen
func TestSyncTaglist(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	memory := sharedEntity.TagList{
		Tags: []sharedEntity.Tag{
			{Name: "TAG1", Description: "TAG1"},
			{Name: "TAG2", Description: "TAG2"},
		},
	}

	tests := []struct {
		name            string
		ctx             context.Context
		incoming        sharedEntity.TagList
		s3Response      sharedEntity.TagList
		s3Error         error
		storeError      error
		expectS3Load    bool
		expectStoreCall bool
		wantErr         bool
	}{
		{
			name:     "nil context",
			ctx:      nil,
			incoming: sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAGX"}}},
			wantErr:  true,
		},
		{
			name: "tag already in memory → no S3 load, no upload",
			ctx:  context.Background(),
			incoming: sharedEntity.TagList{Tags: []sharedEntity.Tag{
				{Name: "TAG1", Description: "TAG1"},
			}},
			expectS3Load:    false,
			expectStoreCall: false,
			wantErr:         false,
		},
		{
			name:            "tag already in S3 → S3 load, upload because merged from S3",
			ctx:             context.Background(),
			incoming:        sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG3", Description: "TAG3"}}},
			s3Response:      sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG3", Description: "TAG3"}}},
			expectS3Load:    true,
			expectStoreCall: true,
			wantErr:         false,
		},
		{
			name:            "new tag → upload",
			ctx:             context.Background(),
			incoming:        sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG999"}}},
			s3Response:      memory,
			expectS3Load:    true,
			expectStoreCall: true,
			wantErr:         false,
		},
		{
			name:            "upload fails",
			ctx:             context.Background(),
			incoming:        sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG999"}}},
			s3Response:      memory,
			expectS3Load:    true,
			expectStoreCall: true,
			storeError:      errors.New("upload failed"),
			wantErr:         true,
		},
		{
			name:         "s3 get fails",
			ctx:          context.Background(),
			incoming:     sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG_ERR"}}},
			s3Error:      errors.New("s3 get failed"),
			expectS3Load: true,
			wantErr:      true,
		},
		{
			name: "all tags already in memory → early return, no S3 load",
			ctx:  context.Background(),
			incoming: sharedEntity.TagList{Tags: []sharedEntity.Tag{
				{Name: "TAG1", Description: "TAG1"},
			}},
			expectS3Load:    false,
			expectStoreCall: false,
			wantErr:         false,
		},
		{
			name: "incoming tag not in memory, but found in S3 → S3 load, upload",
			ctx:  context.Background(),
			incoming: sharedEntity.TagList{Tags: []sharedEntity.Tag{
				{Name: "TAG3", Description: "TAG3"},
			}},
			s3Response: sharedEntity.TagList{Tags: []sharedEntity.Tag{
				{Name: "TAG1", Description: "TAG1"},
				{Name: "TAG2", Description: "TAG2"},
				{Name: "TAG3", Description: "TAG3"},
			}},
			expectS3Load:    true,
			expectStoreCall: true,
			wantErr:         false,
		},
		{
			name: "empty S3 response but new incoming tag → upload happens",
			ctx:  context.Background(),
			incoming: sharedEntity.TagList{Tags: []sharedEntity.Tag{
				{Name: "TAG_NEW", Description: "New tag"},
			}},
			s3Response:      sharedEntity.TagList{Tags: []sharedEntity.Tag{}},
			expectS3Load:    true,
			expectStoreCall: true,
			wantErr:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSrv := mocks.NewMockTaglistStorage(t)

			srv := taglistSync{
				logger:         logger,
				taglistService: mockSrv,
				tagList:        &sharedEntity.TagList{Tags: append([]sharedEntity.Tag(nil), memory.Tags...)},
			}

			// --- mock S3 load ---
			if tc.expectS3Load {
				mockSrv.On("GetTaglist", mock.Anything).Return(&tc.s3Response, tc.s3Error)
			}

			// --- mock upload ---
			if tc.expectStoreCall {
				mockSrv.On("StoreTaglist", mock.Anything, mock.Anything).Return(tc.storeError)
			}

			err := srv.SyncTaglist(tc.ctx, &tc.incoming)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
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
