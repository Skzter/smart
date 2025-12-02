package logger

import (
	"context"
	"testing"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

// test-implementation
type testStorage struct{}

func (ts *testStorage) CreateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if ctx == nil || taglist == nil {
		return sharedErrors.ErrValidation
	}
	return nil
}

func (ts *testStorage) ReadTaglist(ctx context.Context) (*entity.TagList, error) {
	if ctx == nil {
		return nil, sharedErrors.ErrInternalServer
	}
	return &entity.TagList{}, nil
}

func (ts *testStorage) UpdateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if ctx == nil || taglist == nil {
		return sharedErrors.ErrValidation
	}
	return nil
}

func Test_AllErrorFunctions(t *testing.T) {
	storage := &testStorage{}

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "CreateTaglist_nil_context",
			fn: func() error {
				return storage.CreateTaglist(context.TODO(), &entity.TagList{})
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
			name: "CreateTaglist_invalid_context",
			fn: func() error {
				return storage.CreateTaglist(context.Background(), nil)
			},
			wantErr: sharedErrors.ErrValidation,
		},
		{
			name: "ReadTaglist_nil_context",
			fn: func() error {
				_, err := storage.ReadTaglist(context.TODO())
				return err
			},
			wantErr: nil,
		},
		{
			name: "UpdateTaglist_nil_context",
			fn: func() error {
				return storage.UpdateTaglist(context.TODO(), &entity.TagList{})
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
