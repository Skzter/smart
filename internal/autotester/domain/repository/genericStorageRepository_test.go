package repository

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/parquet-go/parquet-go"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
)

func TestCreate(t *testing.T) {
	testCreate[entity.TestCase](t)
	testCreate[entity.SessionSummary](t)
}

func testCreate[T any](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname         string
		ctx              context.Context
		obj              *T
		mockParquetError error
		mockUploadError  error
		expectError      bool
	}{
		{
			testname:    fmt.Sprintf("%s: happy path", typeName),
			ctx:         context.Background(),
			obj:         new(T),
			expectError: false,
		},
		{
			testname:    fmt.Sprintf("%s: nil obj", typeName),
			ctx:         context.Background(),
			obj:         nil,
			expectError: true,
		},
		{
			testname:         fmt.Sprintf("%s: parquet write fails", typeName),
			ctx:              context.Background(),
			obj:              new(T),
			mockParquetError: fmt.Errorf("parquet error"),
			expectError:      true,
		},
		{
			testname:        fmt.Sprintf("%s: s3 upload fails", typeName),
			ctx:             context.Background(),
			obj:             new(T),
			mockUploadError: fmt.Errorf("s3 error"),
			expectError:     true,
		},
	}

	for _, test := range tests {
		mockParquet := &mockParquetWrapper[T]{
			WriteStructToParquetFunc: func(data T) ([]byte, error) {
				if test.mockParquetError != nil {
					return nil, test.mockParquetError
				}
				return []byte("dummy parquet data"), nil
			},
		}

		mockS3 := &mockS3Wrapper{
			UploadParquetFileFunc: func(ctx context.Context, key string, data []byte, metadata map[string]string) error {
				return test.mockUploadError
			},
		}

		repo := &GenericStorageRepository[T]{
			logger:         testLogger(),
			parquetWrapper: mockParquet,
			s3Wrapper:      mockS3,
		}

		t.Run(test.testname, func(t *testing.T) {
			key, err := repo.Create(test.ctx, test.obj)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				if key != "" {
					t.Errorf("expected empty key on error, but got: %s", key)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but got: %v", err)
				}
				if key == "" {
					t.Errorf("expected non-empty key on success")
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	testUpdate[entity.TestCase](t)
	testUpdate[entity.SessionSummary](t)
}

func testUpdate[T any](t *testing.T) {
	t.Helper()

	tests := getUpdateTests[T]()

	for _, test := range tests {
		mockParquet := &mockParquetWrapper[T]{
			WriteStructToParquetFunc: func(data T) ([]byte, error) {
				if test.mockParquetError != nil {
					return nil, test.mockParquetError
				}
				return []byte("dummy parquet data"), nil
			},
		}

		mockS3 := &mockS3Wrapper{
			FileExistsFunc: func(ctx context.Context, key string) (bool, error) {
				return test.fileExists, test.mockFileExists
			},
			UploadParquetFileFunc: func(ctx context.Context, key string, data []byte, metadata map[string]string) error {
				return test.mockUploadError
			},
		}

		repo := &GenericStorageRepository[T]{
			logger:         testLogger(),
			parquetWrapper: mockParquet,
			s3Wrapper:      mockS3,
		}

		t.Run(test.testname, func(t *testing.T) {
			err := repo.Update(test.ctx, test.obj, test.key)

			if test.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but got: %v", err)
				}
			}
		})
	}
}

type updateTestCase[T any] struct {
	testname         string
	ctx              context.Context
	obj              *T
	key              string
	mockFileExists   error
	fileExists       bool
	mockParquetError error
	mockUploadError  error
	expectError      bool
}

func getUpdateTests[T any]() []updateTestCase[T] {
	typeName := reflect.TypeOf(*new(T)).Name()
	return []updateTestCase[T]{
		{
			testname:    fmt.Sprintf("%s: nil object returns error", typeName),
			ctx:         context.Background(),
			obj:         nil,
			key:         "valid-key",
			expectError: true,
		},
		{
			testname:    fmt.Sprintf("%s: empty key returns error", typeName),
			ctx:         context.Background(),
			obj:         new(T),
			key:         "",
			expectError: true,
		},
		{
			testname:       fmt.Sprintf("%s: FileExists returns error", typeName),
			ctx:            context.Background(),
			obj:            new(T),
			key:            "valid-key",
			mockFileExists: fmt.Errorf("file exists error"),
			expectError:    true,
		},
		{
			testname:    fmt.Sprintf("%s: File does not exist", typeName),
			ctx:         context.Background(),
			obj:         new(T),
			key:         "valid-key",
			fileExists:  false,
			expectError: true,
		},
		{
			testname:         fmt.Sprintf("%s: Parquet serialization fails", typeName),
			ctx:              context.Background(),
			obj:              new(T),
			key:              "valid-key",
			fileExists:       true,
			mockParquetError: fmt.Errorf("parquet error"),
			expectError:      true,
		},
		{
			testname:        fmt.Sprintf("%s: s3 upload fails", typeName),
			ctx:             context.Background(),
			obj:             new(T),
			key:             "valid-key",
			fileExists:      true,
			mockUploadError: fmt.Errorf("upload error"),
			expectError:     true,
		},
		{
			testname:    fmt.Sprintf("%s: successful update", typeName),
			ctx:         context.Background(),
			obj:         new(T),
			key:         "valid-key",
			fileExists:  true,
			expectError: false,
		},
	}
}

func TestGenerateKey(t *testing.T) {
	testGenerateKey[entity.TestCase](t)
	testGenerateKey[entity.SessionSummary](t)
}

func testGenerateKey[T any](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		name     string
		input    any
		wantType string
	}{
		{
			name:     fmt.Sprintf("%s: value", typeName),
			input:    *new(T),
			wantType: typeName,
		},
		{
			name:     fmt.Sprintf("%s: pointer", typeName),
			input:    new(T),
			wantType: typeName,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := generateKey(test.input)

			if !strings.HasPrefix(key, test.wantType+"_") {
				t.Errorf("generateKey() = %q, want prefix %q", key, test.wantType+"_")
			}

			expectedLen := len(test.wantType) + 1 + 17
			if len(key) != expectedLen {
				t.Errorf("generateKey() length = %d, want %d", len(key), expectedLen)
			}
		})
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type mockParquetWrapper[T any] struct {
	WriteStructToParquetFunc func(data T) ([]byte, error)
}

func (m *mockParquetWrapper[T]) WriteStructToParquet(data T) ([]byte, error) {
	if m.WriteStructToParquetFunc != nil {
		return m.WriteStructToParquetFunc(data)
	}
	return nil, nil
}

func (m *mockParquetWrapper[T]) ReadStructsFromParquet(parquetData []byte) ([]T, error) {
	panic("not implemented")
}

func (m *mockParquetWrapper[T]) ValidateStruct(data T) error {
	panic("not implemented")
}

func (m *mockParquetWrapper[T]) GetParquetFileInfo(parquetData []byte) (*wrapperEntity.ParquetFileInfo, error) {
	panic("not implemented")
}

func (m *mockParquetWrapper[T]) WriteStructsToParquet(data []T) ([]byte, error) {
	panic("not implemented")
}

func (m *mockParquetWrapper[T]) GetParquetSchema() (*parquet.Schema, error) {
	panic("not implemented")
}

type mockS3Wrapper struct {
	UploadParquetFileFunc func(ctx context.Context, key string, data []byte, metadata map[string]string) error
	FileExistsFunc        func(ctx context.Context, key string) (bool, error)
}

func (m *mockS3Wrapper) UploadParquetFile(ctx context.Context, key string, data []byte, metadata map[string]string) error {
	if m.UploadParquetFileFunc != nil {
		return m.UploadParquetFileFunc(ctx, key, data, metadata)
	}
	return nil
}

func (m *mockS3Wrapper) DownloadParquetFile(ctx context.Context, key string) ([]byte, map[string]string, error) {
	panic("not implemented")
}

func (m *mockS3Wrapper) ListParquetFiles(ctx context.Context, prefix string) ([]string, error) {
	panic("not implemented")
}

func (m *mockS3Wrapper) DeleteParquetFile(ctx context.Context, key string) error {
	panic("not implemented")
}

func (m *mockS3Wrapper) FileExists(ctx context.Context, key string) (bool, error) {
	if m.FileExistsFunc != nil {
		return m.FileExistsFunc(ctx, key)
	}
	return true, nil
}

func (m *mockS3Wrapper) GetFileSize(ctx context.Context, key string) (int64, error) {
	panic("not implemented")
}
