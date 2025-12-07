package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint:funlen
func TestHandleGetRemoteTestcase(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		TestName         string
		QueryParams      map[string]string
		ExpectedStatus   int
		expectedResponse []*entity.TestcaseMetadata
		SetupMock        func(*mocks.MockTestcaseStorageService)
	}{
		{
			TestName:       "Get all testcases without filters",
			QueryParams:    map[string]string{},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
				{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
				{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
					{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
					{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
				}, nil)
			},
		},
		{
			TestName:       "Filter by author only",
			QueryParams:    map[string]string{"author": "9177b856-46a0-11f0-9fe2-0242ac120002"},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
				{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
					{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
				}, nil)
			},
		},
		{
			TestName:       "Filter by testcaseId only",
			QueryParams:    map[string]string{"testcaseId": "bdf1e5b4-47c2-11f0-b8a3-0242ac120002"},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
				}, nil)
			},
		},
		{
			TestName:       "Filter by createdAfter only",
			QueryParams:    map[string]string{"createdAfter": "2025-01-01T00:00:00Z"},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
				{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
					{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
				}, nil)
			},
		},
		{
			TestName:       "Filter by createdBefore only",
			QueryParams:    map[string]string{"createdBefore": "2025-12-31T23:59:59Z"},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
				}, nil)
			},
		},
		{
			TestName: "Filter by createdAfter and createdBefore (date range)",
			QueryParams: map[string]string{
				"createdAfter":  "2025-01-01T00:00:00Z",
				"createdBefore": "2025-06-30T23:59:59Z",
			},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
				}, nil)
			},
		},
		{
			TestName: "Filter by all parameters combined",
			QueryParams: map[string]string{
				"author":        "9177b856-46a0-11f0-9fe2-0242ac120002",
				"testcaseId":    "bdf1e5b4-47c2-11f0-b8a3-0242ac120002",
				"createdAfter":  "2025-01-01T00:00:00Z",
				"createdBefore": "2025-12-31T23:59:59Z",
			},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
				}, nil)
			},
		},
		{
			TestName:         "Service returns error",
			QueryParams:      map[string]string{},
			ExpectedStatus:   http.StatusInternalServerError,
			expectedResponse: nil,
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("storage service error"))
			},
		},
		{
			TestName:         "Invalid author format",
			QueryParams:      map[string]string{"author": "invalid-uuid"},
			ExpectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
			SetupMock:        func(m *mocks.MockTestcaseStorageService) {},
		},
		{
			TestName:         "Invalid testcaseId format",
			QueryParams:      map[string]string{"testcaseId": "not-a-uuid"},
			ExpectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
			SetupMock:        func(m *mocks.MockTestcaseStorageService) {},
		},
		{
			TestName:         "Invalid createdAfter format",
			QueryParams:      map[string]string{"createdAfter": "invalid-date"},
			ExpectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
			SetupMock:        func(m *mocks.MockTestcaseStorageService) {},
		},
		{
			TestName:         "Invalid createdBefore format",
			QueryParams:      map[string]string{"createdBefore": "not-a-date"},
			ExpectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
			SetupMock:        func(m *mocks.MockTestcaseStorageService) {},
		},
		{
			TestName:       "Filter with limit only",
			QueryParams:    map[string]string{"limit": "2"},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
				{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test1_1701456000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701456000", Updated: "1701460000", Name: "Test 1"},
					{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
				}, nil)
			},
		},
		{
			TestName:       "Filter with offset only",
			QueryParams:    map[string]string{"offset": "1"},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
				{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
					{Key: "testcase/test3_1701458000.parquet", Author: "9177b856-46a0-11f0-9fe2-0242ac120002", Created: "1701458000", Updated: "1701462000", Name: "Test 3"},
				}, nil)
			},
		},
		{
			TestName:       "Filter with limit and offset",
			QueryParams:    map[string]string{"limit": "1", "offset": "1"},
			ExpectedStatus: http.StatusOK,
			expectedResponse: []*entity.TestcaseMetadata{
				{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
			},
			SetupMock: func(m *mocks.MockTestcaseStorageService) {
				m.EXPECT().ReadAllMetadataWithFilter(mock.Anything, mock.Anything).Return([]*entity.TestcaseMetadata{
					{Key: "testcase/test2_1701457000.parquet", Author: "8177b856-46a0-11f0-9fe2-0242ac120003", Created: "1701457000", Updated: "1701461000", Name: "Test 2"},
				}, nil)
			},
		},
		{
			TestName:         "Invalid limit (less than 1)",
			QueryParams:      map[string]string{"limit": "0"},
			ExpectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
			SetupMock:        func(m *mocks.MockTestcaseStorageService) {},
		},
		{
			TestName:         "Invalid offset (negative)",
			QueryParams:      map[string]string{"offset": "-1"},
			ExpectedStatus:   http.StatusBadRequest,
			expectedResponse: nil,
			SetupMock:        func(m *mocks.MockTestcaseStorageService) {},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
			mockChatManager := mocks.NewMockChatManager(t)

			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockChatManager,
			)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			test.SetupMock(mockRemoteStorageServ)

			url := "/api/v1/tests"
			if len(test.QueryParams) > 0 {
				url += "?"
				first := true
				for key, value := range test.QueryParams {
					if !first {
						url += "&"
					}
					url += key + "=" + value
					first = false
				}
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller.HandleGetRemoteTestcase(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				var responseData []entity.TestcaseMetadata
				if err := json.Unmarshal(rec.Body.Bytes(), &responseData); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if len(responseData) != len(test.expectedResponse) {
					t.Errorf("Expected %d items, got %d", len(test.expectedResponse), len(responseData))
					return
				}

				for i, expected := range test.expectedResponse {
					actual := responseData[i]

					if actual.Key != expected.Key {
						t.Errorf("Item %d: Expected Key %s, got %s", i, expected.Key, actual.Key)
					}
					if actual.Author != expected.Author {
						t.Errorf("Item %d: Expected Author %s, got %s", i, expected.Author, actual.Author)
					}
					if actual.Created != expected.Created {
						t.Errorf("Item %d: Expected Created %s, got %s", i, expected.Created, actual.Created)
					}
					if actual.Updated != expected.Updated {
						t.Errorf("Item %d: Expected Updated %s, got %s", i, expected.Updated, actual.Updated)
					}
					if actual.Name != expected.Name {
						t.Errorf("Item %d: Expected Name %s, got %s", i, expected.Name, actual.Name)
					}
				}
			}
		})
	}
}
