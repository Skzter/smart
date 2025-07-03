package repository

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
)

func TestNewGenericStorageRepository(t *testing.T) {
	testNewGenericStorageRepository[entity.TestCase](t)
	testNewGenericStorageRepository[entity.SessionSummary](t)
}

func testNewGenericStorageRepository[T Storable](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname             string
		logger               *slog.Logger
		s3Wrapper            *mocks.MockS3StorageWrapper
		parquetWrapper       *mocks.MockParquetFileWrapper[T]
		structValidationFunc func(*T) error
		wantErr              bool
	}{
		{
			testname:             typeName + ": happy path",
			logger:               slog.Default(),
			s3Wrapper:            &mocks.MockS3StorageWrapper{},
			parquetWrapper:       &mocks.MockParquetFileWrapper[T]{},
			structValidationFunc: func(*T) error { return nil },
			wantErr:              false,
		},
		{
			testname:             typeName + ": nil logger",
			logger:               nil,
			s3Wrapper:            &mocks.MockS3StorageWrapper{},
			parquetWrapper:       &mocks.MockParquetFileWrapper[T]{},
			structValidationFunc: func(*T) error { return nil },
			wantErr:              true,
		},
		{
			testname:             typeName + ": nil s3Wrapper",
			logger:               slog.Default(),
			s3Wrapper:            nil,
			parquetWrapper:       &mocks.MockParquetFileWrapper[T]{},
			structValidationFunc: func(*T) error { return nil },
			wantErr:              true,
		},
		{
			testname:             typeName + ": nil parquetWrapper",
			logger:               slog.Default(),
			s3Wrapper:            &mocks.MockS3StorageWrapper{},
			parquetWrapper:       nil,
			structValidationFunc: func(*T) error { return nil },
			wantErr:              true,
		},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			_, err := newGenericStorageRepository[T](
				test.logger,
				test.s3Wrapper,
				test.parquetWrapper,
				test.structValidationFunc,
			)
			if (err != nil) != test.wantErr {
				t.Errorf("[%s] newGenericStorageRepository() error = %v, wantErr %v", typeName, err, test.wantErr)
			}
		})
	}
}

func TestNewStorageRepository(t *testing.T) {
	testNewStorageRepository[entity.TestCase](t)
	testNewStorageRepository[entity.SessionSummary](t)
}

func testNewStorageRepository[T Storable](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname string
		logger   *slog.Logger
		wantErr  bool
	}{
		{
			testname: typeName + ": happy path",
			logger:   slog.Default(),
			wantErr:  false,
		},
		{
			testname: typeName + ": nil logger",
			logger:   nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			_, err := NewStorageRepository[T](test.logger)
			if (err != nil) != test.wantErr {
				t.Errorf("[%s] NewStorageRepository() error = %v, wantErr %v", typeName, err, test.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	testCreate[entity.TestCase](t)
	testCreate[entity.SessionSummary](t)
}

func testCreate[T Storable](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()
	ctx := context.Background()
	validObj := validTestObjPtr[T]()

	tests := []struct {
		testname         string
		ctx              context.Context
		obj              *T
		mockParquetError error
		mockUploadError  error
		expectError      bool
	}{
		{
			testname:    typeName + ": happy path",
			ctx:         ctx,
			obj:         validObj,
			expectError: false,
		},
		{
			testname:    typeName + ": nil obj",
			ctx:         ctx,
			obj:         nil,
			expectError: true,
		},
		{
			testname:         typeName + ": parquet write fails",
			ctx:              ctx,
			obj:              validObj,
			mockParquetError: fmt.Errorf("parquet error"),
			expectError:      true,
		},
		{
			testname:        typeName + ": s3 upload fails",
			ctx:             ctx,
			obj:             validObj,
			mockUploadError: fmt.Errorf("s3 error"),
			expectError:     true,
		},
		{
			testname:    typeName + ": validation fails",
			ctx:         ctx,
			obj:         new(T),
			expectError: true,
		},
	}

	for _, test := range tests {
		mockParquet := &mocks.MockParquetFileWrapper[T]{}
		mockS3 := &mocks.MockS3StorageWrapper{}

		if test.obj != nil {
			mockParquet.On("WriteStructToParquet", *test.obj).
				Return([]byte("dummy parquet data"), test.mockParquetError)
		}
		if test.obj != nil && test.mockParquetError == nil {
			mockS3.On("UploadParquetFile", test.ctx, mock.Anything, []byte("dummy parquet data"), mock.Anything).
				Return(test.mockUploadError)
		}

		validationFunc, err := GetValidationFunc[T]()
		if err != nil {
			t.Fatalf("setup failed, tried to test a type for which no ValidationFunc exists: %v", err)
		}

		repo, err := newGenericStorageRepository[T](
			slog.Default(),
			mockS3,
			mockParquet,
			validationFunc,
		)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		t.Run(test.testname, func(t *testing.T) {
			key, err := repo.Create(test.ctx, test.obj)
			if test.expectError {
				if err == nil {
					t.Errorf("[%s] Create() expected error but got none", typeName)
				}
				if key != "" {
					t.Errorf("[%s] Create() expected empty key on error, got: %s", typeName, key)
				}
			} else {
				if err != nil {
					t.Errorf("[%s] Create() unexpected error: %v", typeName, err)
				}
				if key == "" {
					t.Errorf("[%s] Create() expected non-empty key on success", typeName)
				}
			}
		})
	}
}

const validKey string = "valid-key"

func TestRead(t *testing.T) {
	testRead[entity.TestCase](t)
	testRead[entity.SessionSummary](t)
}

func getReadTestcases[T Storable]() []struct {
	testname         string
	key              string
	mockDownloadData []byte
	mockDownloadErr  error
	mockReadStructs  []T
	mockReadErr      error
	expectError      bool
	expectNilResult  bool
	validationFails  bool
} {
	typeName := reflect.TypeOf(*new(T)).Name()
	validObj := validTestObjPtr[T]()
	return []struct {
		testname         string
		key              string
		mockDownloadData []byte
		mockDownloadErr  error
		mockReadStructs  []T
		mockReadErr      error
		expectError      bool
		expectNilResult  bool
		validationFails  bool
	}{
		{
			testname:         typeName + ": happy path",
			key:              validKey,
			mockDownloadData: []byte("parquet"),
			mockReadStructs:  []T{*validObj},
			expectError:      false,
			expectNilResult:  false,
		},
		{
			testname:        typeName + ": empty key",
			key:             "",
			expectError:     true,
			expectNilResult: true,
		},
		{
			testname:        typeName + ": download error",
			key:             validKey,
			mockDownloadErr: fmt.Errorf("download error"),
			expectError:     true,
			expectNilResult: true,
		},
		{
			testname:         typeName + ": read parquet error",
			key:              validKey,
			mockDownloadData: []byte("parquet"),
			mockReadErr:      fmt.Errorf("read error"),
			expectError:      true,
			expectNilResult:  true,
		},
		{
			testname:         typeName + ": no data found",
			key:              validKey,
			mockDownloadData: []byte("parquet"),
			mockReadStructs:  []T{},
			expectError:      true,
			expectNilResult:  true,
		},
		{
			testname:         typeName + ": validation fails",
			key:              validKey,
			mockDownloadData: []byte("parquet"),
			mockReadStructs:  []T{*new(T)},
			expectError:      true,
			expectNilResult:  true,
		},
	}
}

func testRead[T Storable](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()
	ctx := context.Background()

	tests := getReadTestcases[T]()

	for _, test := range tests {
		mockParquet := &mocks.MockParquetFileWrapper[T]{}
		mockS3 := &mocks.MockS3StorageWrapper{}

		mockS3.On("DownloadParquetFile", ctx, test.key).
			Return(test.mockDownloadData, map[string]string{}, test.mockDownloadErr)

		mockParquet.On("ReadStructsFromParquet", test.mockDownloadData).
			Return(test.mockReadStructs, test.mockReadErr)

		validationFunc, err := GetValidationFunc[T]()
		if err != nil {
			t.Fatalf("setup failed, tried to test a type for which no ValidationFunc exists: %v", err)
		}

		repo, err := newGenericStorageRepository[T](
			slog.Default(),
			mockS3,
			mockParquet,
			validationFunc,
		)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		t.Run(test.testname, func(t *testing.T) {
			result, err := repo.Read(ctx, test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("[%s] Read() expected error but got none", typeName)
				}
				if !test.expectNilResult && result != nil {
					t.Errorf("[%s] Read() expected nil result on error, got: %+v", typeName, result)
				}
			} else {
				if err != nil {
					t.Errorf("[%s] Read() unexpected error: %v", typeName, err)
				}
				if result == nil {
					t.Errorf("[%s] Read() expected non-nil result on success", typeName)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	testUpdate[entity.TestCase](t)
	testUpdate[entity.SessionSummary](t)
}

func getUpdateTestcases[T Storable]() []struct {
	testname          string
	obj               *T
	key               string
	mockFileExists    bool
	mockFileExistsErr error
	mockParquetError  error
	mockUploadError   error
	expectError       bool
} {
	typeName := reflect.TypeOf(*new(T)).Name()
	validObj := validTestObjPtr[T]()

	return []struct {
		testname          string
		obj               *T
		key               string
		mockFileExists    bool
		mockFileExistsErr error
		mockParquetError  error
		mockUploadError   error
		expectError       bool
	}{
		{
			testname:       typeName + ": happy path",
			obj:            validObj,
			key:            validKey,
			mockFileExists: true,
			expectError:    false,
		},
		{
			testname:    typeName + ": nil obj",
			obj:         nil,
			key:         validKey,
			expectError: true,
		},
		{
			testname:    typeName + ": validation fails",
			obj:         new(T),
			key:         validKey,
			expectError: true,
		},
		{
			testname:    typeName + ": empty key",
			obj:         validObj,
			key:         "",
			expectError: true,
		},
		{
			testname:          typeName + ": file exists check error",
			obj:               validObj,
			key:               validKey,
			mockFileExistsErr: fmt.Errorf("exists error"),
			expectError:       true,
		},
		{
			testname:       typeName + ": file does not exist",
			obj:            validObj,
			key:            validKey,
			mockFileExists: false,
			expectError:    true,
		},
		{
			testname:         typeName + ": parquet write fails",
			obj:              validObj,
			key:              validKey,
			mockFileExists:   true,
			mockParquetError: fmt.Errorf("parquet error"),
			expectError:      true,
		},
		{
			testname:        typeName + ": s3 upload fails",
			obj:             validObj,
			key:             validKey,
			mockFileExists:  true,
			mockUploadError: fmt.Errorf("s3 error"),
			expectError:     true,
		},
	}
}

func testUpdate[T Storable](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()
	ctx := context.Background()

	tests := getUpdateTestcases[T]()

	for _, test := range tests {
		mockParquet := &mocks.MockParquetFileWrapper[T]{}
		mockS3 := &mocks.MockS3StorageWrapper{}

		mockS3.On("FileExists", ctx, test.key).
			Return(test.mockFileExists, test.mockFileExistsErr)

		if test.obj != nil && test.key != "" && test.mockFileExistsErr == nil && test.mockFileExists {
			mockParquet.On("WriteStructToParquet", *test.obj).
				Return([]byte("dummy parquet data"), test.mockParquetError)
		}

		if test.obj != nil && test.key != "" && test.mockFileExistsErr == nil && test.mockFileExists && test.mockParquetError == nil {
			mockS3.On("UploadParquetFile", ctx, test.key, []byte("dummy parquet data"), mock.Anything).
				Return(test.mockUploadError)
		}

		validationFunc, err := GetValidationFunc[T]()
		if err != nil {
			t.Fatalf("setup failed, tried to test a type for which no ValidationFunc exists: %v", err)
		}

		repo, err := newGenericStorageRepository[T](
			slog.Default(),
			mockS3,
			mockParquet,
			validationFunc,
		)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		t.Run(test.testname, func(t *testing.T) {
			err := repo.Update(ctx, test.obj, test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("[%s] Update() expected error but got none", typeName)
				}
			} else {
				if err != nil {
					t.Errorf("[%s] Update() unexpected error: %v", typeName, err)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testDelete[entity.TestCase](t)
	testDelete[entity.SessionSummary](t)
}

func testDelete[T Storable](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()
	ctx := context.Background()

	tests := []struct {
		testname      string
		key           string
		mockDeleteErr error
		expectError   bool
	}{
		{
			testname:    typeName + ": happy path",
			key:         validKey,
			expectError: false,
		},
		{
			testname:    typeName + ": empty key",
			key:         "",
			expectError: true,
		},
		{
			testname:      typeName + ": delete error",
			key:           validKey,
			mockDeleteErr: fmt.Errorf("delete error"),
			expectError:   true,
		},
	}

	for _, test := range tests {
		mockS3 := &mocks.MockS3StorageWrapper{}
		mockParquet := &mocks.MockParquetFileWrapper[T]{}

		mockS3.On("DeleteParquetFile", ctx, test.key).
			Return(test.mockDeleteErr)

		validationFunc, err := GetValidationFunc[T]()
		if err != nil {
			t.Fatalf("setup failed, tried to test a type for which no ValidationFunc exists: %v", err)
		}

		repo, err := newGenericStorageRepository[T](
			slog.Default(),
			mockS3,
			mockParquet,
			validationFunc,
		)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		t.Run(test.testname, func(t *testing.T) {
			err := repo.Delete(ctx, test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("[%s] Delete() expected error but got none", typeName)
				}
			} else {
				if err != nil {
					t.Errorf("[%s] Delete() unexpected error: %v", typeName, err)
				}
			}
		})
	}
}

func TestGetValidationFunc(t *testing.T) {
	testGetValidationFunc[entity.TestCase](t)
	testGetValidationFunc[entity.SessionSummary](t)
}

func testGetValidationFunc[T Storable](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname      string
		obj           *T
		wantErr       bool
		wantValidFunc bool
	}{
		{
			testname:      typeName + ": valid func for supported type",
			obj:           validTestObjPtr[T](),
			wantErr:       false,
			wantValidFunc: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			validationFunc, err := GetValidationFunc[T]()
			if test.wantErr {
				if err == nil {
					t.Errorf("[%s] GetValidationFunc() expected error but got none", typeName)
				}
				if validationFunc != nil {
					t.Errorf("[%s] GetValidationFunc() expected nil func on error", typeName)
				}
			} else {
				if err != nil {
					t.Errorf("[%s] GetValidationFunc() unexpected error: %v", typeName, err)
				}
				if validationFunc == nil {
					t.Errorf("[%s] GetValidationFunc() expected non-nil func", typeName)
				}
			}
		})
	}
}

func Test_testCaseValidationFunc(t *testing.T) {
	tests := []struct {
		name    string
		obj     entity.TestCase
		wantErr bool
	}{
		{
			name: "valid TestCase",
			obj: entity.TestCase{
				TestID:      "id",
				Description: "desc",
				TestCode:    entity.TestCode{Code: "code"},
				Status:      entity.TestStatusPassed,
			},
			wantErr: false,
		},
		{
			name: "empty TestID",
			obj: entity.TestCase{
				TestID:      "",
				Description: "desc",
				TestCode:    entity.TestCode{Code: "code"},
				Status:      entity.TestStatusPassed,
			},
			wantErr: true,
		},
		{
			name: "empty Description",
			obj: entity.TestCase{
				TestID:      "id",
				Description: "",
				TestCode:    entity.TestCode{Code: "code"},
				Status:      entity.TestStatusPassed,
			},
			wantErr: true,
		},
		{
			name: "empty Status",
			obj: entity.TestCase{
				TestID:      "id",
				Description: "desc",
				TestCode:    entity.TestCode{Code: "code"},
				Status:      "",
			},
			wantErr: true,
		},
		{
			name: "empty TestCode.Code",
			obj: entity.TestCase{
				TestID:      "id",
				Description: "desc",
				TestCode:    entity.TestCode{Code: ""},
				Status:      entity.TestStatusPassed,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := testCaseValidationFunc(&test.obj)
			if (err != nil) != test.wantErr {
				t.Errorf("testCaseValidationFunc() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestSessionSummaryValidationFunc(t *testing.T) {
	tests := []struct {
		name    string
		obj     entity.SessionSummary
		wantErr bool
	}{
		{
			name: "valid SessionSummary",
			obj: entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
			wantErr: false,
		},
		{
			name: "empty Summary",
			obj: entity.SessionSummary{
				Summary:   "",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "nil Messages",
			obj: entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  nil,
			},
			wantErr: true,
		},
		{
			name: "Messages contains nil",
			obj: entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{nil},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := sessionSummaryValidationFunc(&test.obj)
			if (err != nil) != test.wantErr {
				t.Errorf("sessionSummaryValidationFunc() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func validTestObjPtr[T any]() *T {
	var zero T
	switch any(zero).(type) {
	case entity.TestCase:
		v := entity.TestCase{
			TestID:      "id",
			Description: "desc",
			TestCode:    entity.TestCode{Code: "code"},
			Status:      entity.TestStatusPassed,
		}
		return any(&v).(*T)
	case entity.SessionSummary:
		v := entity.SessionSummary{
			Summary:   "summary",
			CreatedAt: time.Now(),
			Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
		}
		return any(&v).(*T)
	default:
		return nil
	}
}
