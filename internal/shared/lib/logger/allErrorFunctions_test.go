package logger

import (
	"context"
	"io"
	"log/slog"
	"testing"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

type testStorage struct {
	logger *slog.Logger
}

func newTestStorage(logger *slog.Logger) *testStorage {
	return &testStorage{logger: logger}
}

func (ts *testStorage) CreateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if ctx == nil || taglist == nil {
		ts.logger.Error("CreateTaglist validation failed",
			"error", "nil context or taglist",
			"ctx_nil", ctx == nil,
			"taglist_nil", taglist == nil,
		)
		return sharedErrors.ErrValidation
	}

	ts.logger.Debug("CreateTaglist validation passed",
		"tags_count", len(taglist.Tags),
	)
	return nil
}

func (ts *testStorage) ReadTaglist(ctx context.Context) (*entity.TagList, error) {
	if ctx == nil {
		ts.logger.Error("ReadTaglist failed: nil context")
		return nil, sharedErrors.ErrInternalServer
	}

	ts.logger.Debug("ReadTaglist returning empty list")
	return &entity.TagList{}, nil
}

func (ts *testStorage) UpdateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if ctx == nil || taglist == nil {
		ts.logger.Error("UpdateTaglist validation failed",
			"error", "nil context or taglist",
		)
		return sharedErrors.ErrValidation
	}

	ts.logger.Info("UpdateTaglist called",
		"tags_count", len(taglist.Tags),
	)
	return nil
}

func Test_AllErrorFunctions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	storage := newTestStorage(logger)

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "CreateTaglist_success",
			fn: func() error {
				return storage.CreateTaglist(context.Background(), &entity.TagList{
					Tags: []entity.Tag{{Name: "test", Description: "test"}},
				})
			},
			wantErr: nil,
		},
		{
			name: "CreateTaglist_nil_taglist",
			fn: func() error {
				return storage.CreateTaglist(context.Background(), nil)
			},
			wantErr: sharedErrors.ErrValidation,
		},
		{
			name: "CreateTaglist_empty_tags",
			fn: func() error {
				return storage.CreateTaglist(context.Background(), &entity.TagList{
					Tags: []entity.Tag{},
				})
			},
			wantErr: nil,
		},
		{
			name: "ReadTaglist_success",
			fn: func() error {
				_, err := storage.ReadTaglist(context.Background())
				return err
			},
			wantErr: nil,
		},
		{
			name: "UpdateTaglist_success",
			fn: func() error {
				return storage.UpdateTaglist(context.Background(), &entity.TagList{
					Tags: []entity.Tag{{Name: "updated", Description: "updated"}},
				})
			},
			wantErr: nil,
		},
		{
			name: "UpdateTaglist_nil_taglist",
			fn: func() error {
				return storage.UpdateTaglist(context.Background(), nil)
			},
			wantErr: sharedErrors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()

			if tt.wantErr == nil && err != nil {
				t.Errorf("want nil, got %v", err)
			}
			if tt.wantErr != nil && err != tt.wantErr {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}
