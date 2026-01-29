package repository

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
)

// vorerst TODO
const token = "token"

func TestNewAutotesterAPIRepository(t *testing.T) {
	tests := []struct {
		name      string
		logger    *slog.Logger
		client    *http.Client
		baseURL   string
		expectErr bool
	}{
		{
			name:      "success",
			logger:    slog.New(slog.DiscardHandler),
			client:    &http.Client{},
			baseURL:   "http://example.com",
			expectErr: false,
		},
		{
			name:      "nil-logger",
			logger:    nil,
			client:    &http.Client{},
			baseURL:   "http://example.com",
			expectErr: true,
		},
		{
			name:      "nil-client",
			logger:    slog.Default(),
			client:    nil,
			baseURL:   "http://example.com",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewAutotesterAPIRepository(test.logger, test.client, test.baseURL)
			if test.expectErr {
				require.Error(t, err)
				require.Nil(t, repo)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, repo)
		})
	}
}

func TestGetTemplate(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name            string
		token           string
		statusCode      int
		responseBody    string
		expectErr       bool
		expectedContent string
	}{
		{
			name:            "success",
			token:           token,
			statusCode:      http.StatusOK,
			responseBody:    `{"template":"template text"}`,
			expectErr:       false,
			expectedContent: "template text",
		},
		{
			name:            "non-200",
			token:           token,
			statusCode:      http.StatusInternalServerError,
			responseBody:    `error`,
			expectErr:       true,
			expectedContent: "",
		},
		{
			name:            "invalid-json",
			token:           token,
			statusCode:      http.StatusOK,
			responseBody:    `{"template":`,
			expectErr:       true,
			expectedContent: "",
		},
		{
			name:            "empty-token",
			token:           "",
			statusCode:      http.StatusUnauthorized,
			responseBody:    `unauthorized`,
			expectErr:       true,
			expectedContent: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/template" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer "+test.token {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer srv.Close()

			client := srv.Client()
			repo, err := NewAutotesterAPIRepository(logger, client, srv.URL)
			require.NoError(t, err)

			res, err := repo.GetTemplate(context.Background(), test.token)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, test.expectedContent, res.Content)
		})
	}
}

func TestValidatePrompt(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name         string
		token        string
		statusCode   int
		responseBody string
		expectErr    bool
		expectedMsg  string
	}{
		{
			name:         "success - valid prompt",
			token:        token,
			statusCode:   http.StatusOK,
			responseBody: `{"message":{"body":""},"userId":"user-123","chatId":"chat-456"}`,
			expectErr:    false,
			expectedMsg:  "",
		},
		{
			name:         "success - invalid prompt with feedback",
			token:        token,
			statusCode:   http.StatusOK,
			responseBody: `{"message":{"body":"Please provide more context"},"userId":"user-123","chatId":"chat-456"}`,
			expectErr:    false,
			expectedMsg:  "Please provide more context",
		},
		{
			name:         "non-200 feedback",
			token:        token,
			statusCode:   http.StatusBadRequest,
			responseBody: `error`,
			expectErr:    true,
			expectedMsg:  "",
		},
		{
			name:         "invalid-json",
			token:        token,
			statusCode:   http.StatusOK,
			responseBody: `{"message":`,
			expectErr:    true,
			expectedMsg:  "",
		},
		{
			name:         "empty-token",
			token:        "",
			statusCode:   http.StatusUnauthorized,
			responseBody: `unauthorized`,
			expectErr:    true,
			expectedMsg:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/validate" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer "+test.token {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer srv.Close()

			client := srv.Client()
			repo, err := NewAutotesterAPIRepository(logger, client, srv.URL)
			require.NoError(t, err)

			req := &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			}

			res, err := repo.ValidatePrompt(context.Background(), req, test.token)
			if test.expectErr {
				require.Error(t, err)
				require.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, test.expectedMsg, res.Result.Body)
			require.Equal(t, "user-123", res.UserId)
			require.Equal(t, "chat-456", res.ChatId)
		})
	}
}
func TestGenerateTest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name         string
		token        string
		statusCode   int
		responseBody string
		expectErr    bool
		expectedTest string
	}{
		{
			name:         "success",
			token:        token,
			statusCode:   http.StatusOK,
			responseBody: `{"message":{"id":"msg-1","role":"assistant","body":"generated test code","createdAt":"2025-12-11T10:00:00Z"},"userId":"user-123","chatId":"chat-456"}`,
			expectErr:    false,
			expectedTest: "generated test code",
		},
		{
			name:         "non-200",
			token:        token,
			statusCode:   http.StatusBadRequest,
			responseBody: `error`,
			expectErr:    true,
			expectedTest: "",
		},
		{
			name:         "invalid-json",
			token:        token,
			statusCode:   http.StatusOK,
			responseBody: `{"message":`,
			expectErr:    true,
			expectedTest: "",
		},
		{
			name:         "empty-token",
			token:        "",
			statusCode:   http.StatusUnauthorized,
			responseBody: `unauthorized`,
			expectErr:    true,
			expectedTest: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/chat" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer "+test.token {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer srv.Close()

			client := srv.Client()
			repo, err := NewAutotesterAPIRepository(logger, client, srv.URL)
			require.NoError(t, err)

			req := &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			}

			res, err := repo.GenerateTest(context.Background(), req, test.token)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, test.expectedTest, res.Result.Body)
		})
	}
}

func TestSaveTest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name         string
		token        string
		statusCode   int
		responseBody string
		expectErr    bool
		expectUUID   bool
	}{
		{
			name:         "success",
			token:        token,
			statusCode:   http.StatusOK,
			responseBody: `{"testcaseId":"550e8400-e29b-41d4-a716-446655440000","action":"saved"}`,
			expectErr:    false,
			expectUUID:   true,
		},
		{
			name:         "non-200",
			token:        token,
			statusCode:   http.StatusInternalServerError,
			responseBody: `error`,
			expectErr:    true,
			expectUUID:   false,
		},
		{
			name:         "invalid-json",
			token:        token,
			statusCode:   http.StatusOK,
			responseBody: `{"testcaseId":`,
			expectErr:    true,
			expectUUID:   false,
		},
		{
			name:         "empty-token",
			token:        "",
			statusCode:   http.StatusUnauthorized,
			responseBody: `unauthorized`,
			expectErr:    true,
			expectUUID:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/saveLocal" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer "+test.token {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer srv.Close()

			client := srv.Client()
			repo, err := NewAutotesterAPIRepository(logger, client, srv.URL)
			require.NoError(t, err)

			req := &entity.SaveTestRequest{
				Code:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			}

			res, err := repo.SaveTest(context.Background(), req, token)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			if test.expectUUID {
				require.NotEmpty(t, res.TestId)
				require.Len(t, res.TestId, 36)
			}
		})
	}
}

func TestRunTest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name           string
		token          string
		statusCode     int
		responseBody   string
		expectErr      bool
		expectedStatus string
	}{
		{
			name:           "success",
			token:          token,
			statusCode:     http.StatusOK,
			responseBody:   `{"result":"passed"}`,
			expectErr:      false,
			expectedStatus: "passed",
		},
		{
			name:           "non-200",
			token:          token,
			statusCode:     http.StatusInternalServerError,
			responseBody:   `error`,
			expectErr:      true,
			expectedStatus: "",
		},
		{
			name:           "invalid-json",
			token:          token,
			statusCode:     http.StatusOK,
			responseBody:   `{"result":`,
			expectErr:      true,
			expectedStatus: "",
		},
		{
			name:           "empty-token",
			token:          "",
			statusCode:     http.StatusUnauthorized,
			responseBody:   `unauthorized`,
			expectErr:      true,
			expectedStatus: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/run" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer "+test.token {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer srv.Close()

			client := srv.Client()
			repo, err := NewAutotesterAPIRepository(logger, client, srv.URL)
			require.NoError(t, err)

			req := &entity.RunTestRequest{
				TestId: "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			}

			res, err := repo.RunTest(context.Background(), req, test.token)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, test.expectedStatus, res.Result)
		})
	}
}

// nolint:funlen
func TestReadTestLogStream(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name           string
		testId         string
		token          string
		statusCode     int
		streamEvents   []string
		expectedEvents []entity.LogEvent
		expectErr      bool
		cancelContext  bool
	}{
		{
			name:       "success-multiple-events",
			testId:     "uuid-123",
			token:      token,
			statusCode: http.StatusOK,
			streamEvents: []string{
				"event:progress\ndata:{\"content\":\"started\"}\n\n",
				"event:log\ndata:{\"content\":\"running test\"}\n\n",
			},
			expectedEvents: []entity.LogEvent{
				{Event: "progress", Data: "{\"content\":\"started\"}"},
				{Event: "log", Data: "{\"content\":\"running test\"}"},
			},
			expectErr: false,
		},
		{
			name:           "error-status-404",
			testId:         "invalid-id",
			token:          token,
			statusCode:     http.StatusNotFound,
			streamEvents:   []string{},
			expectedEvents: nil,
			expectErr:      true,
		},
		{
			name:       "robustness-invalid-lines",
			testId:     "uuid-456",
			token:      token,
			statusCode: http.StatusOK,
			streamEvents: []string{
				"event:progress\ndata:{\"content\":\"started\"}\n\n",
				"this is an invalid line without colon\n",
				": comment\n\n",
				"event:log\ndata:{\"content\":\"still running\"}\n\n",
			},
			expectedEvents: []entity.LogEvent{
				{Event: "progress", Data: "{\"content\":\"started\"}"},
				{Event: "log", Data: "{\"content\":\"still running\"}"},
			},
			expectErr: false,
		},
		{
			name:       "context-canceled",
			testId:     "uuid-ctx",
			token:      token,
			statusCode: http.StatusOK,
			streamEvents: []string{ // more than buffered channel
				"event:progress\ndata:{\"content\":\"started\"}\n\n",
				"event:log\ndata:{\"content\":\"running test\"}\n\n",
				"event:progress\ndata:{\"content\":\"started\"}\n\n",
			},
			expectErr:     true,
			cancelContext: true,
		},
		{
			name:       "spaces-and-multiline-data",
			testId:     "uuid-spaces",
			token:      token,
			statusCode: http.StatusOK,
			streamEvents: []string{
				"event:  trim-me  \ndata:   {\"key\": \"value\"}   \ndata: second-line \n\n",
			},
			expectedEvents: []entity.LogEvent{
				{Event: "trim-me", Data: "{\"key\": \"value\"}\nsecond-line"},
			},
			expectErr: false,
		},
		{
			name:       "very-large-data-line",
			testId:     "uuid-large",
			token:      token,
			statusCode: http.StatusOK,
			streamEvents: []string{
				"event:large\ndata:" + strings.Repeat("A", 70000) + "\n\n",
			},
			expectedEvents: []entity.LogEvent{
				{Event: "large", Data: strings.Repeat("A", 70000)},
			},
			expectErr: false,
		},
		{
			name:       "robustness-invalid-and-comments",
			testId:     "uuid-456",
			token:      token,
			statusCode: http.StatusOK,
			streamEvents: []string{
				": this is an SSE comment and should be ignored\n",
				"event:progress\ndata:{\"content\":\"started\"}\n\n",
				"invalid line without prefix and colon is ignored by logic\n",
				": yet another comment\n",
				"event:log\ndata:{\"content\":\"still running\"}\n\n",
			},
			expectedEvents: []entity.LogEvent{
				{Event: "progress", Data: "{\"content\":\"started\"}"},
				{Event: "log", Data: "{\"content\":\"still running\"}"},
			},
			expectErr: false,
		},
		{
			name:           "empty-token",
			testId:         "uuid-token",
			token:          "",
			statusCode:     http.StatusUnauthorized,
			streamEvents:   []string{},
			expectedEvents: nil,
			expectErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/api/v1/test/%s/stream", test.testId)
				if r.URL.Path != expectedPath {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				if r.Header.Get("Authorization") != "Bearer "+test.token {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if test.statusCode != http.StatusOK {
					w.WriteHeader(test.statusCode)
					return
				}

				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(http.StatusOK)

				flusher, ok := w.(http.Flusher)
				require.True(t, ok)

				for _, event := range test.streamEvents {
					_, _ = fmt.Fprint(w, event)
					flusher.Flush()
				}
			}))
			defer srv.Close()

			repo, err := NewAutotesterAPIRepository(logger, srv.Client(), srv.URL)
			require.NoError(t, err)

			eventsCh := make(chan *entity.LogEvent, 2)
			ctx, cancel := context.WithCancel(context.Background())

			if test.cancelContext {
				errCh := make(chan error, 1)
				go func() {
					errCh <- repo.ReadTestLogStream(ctx, test.testId, test.token, eventsCh)
				}()
				time.Sleep(50 * time.Millisecond)
				cancel()

				err = <-errCh
				close(eventsCh)
			} else {
				err = repo.ReadTestLogStream(ctx, test.testId, test.token, eventsCh)
				close(eventsCh)
				cancel()
			}

			if test.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			var received []entity.LogEvent
			for ev := range eventsCh {
				received = append(received, *ev)
			}

			require.Equal(t, len(test.expectedEvents), len(received))
			for i := range test.expectedEvents {
				require.Equal(t, test.expectedEvents[i].Event, received[i].Event)
				require.Equal(t, test.expectedEvents[i].Data, received[i].Data)
			}
		})
	}
}

func TestNewJSONRequest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name              string
		method            string
		url               string
		body              interface{}
		expectErr         bool
		expectContentType bool
	}{
		{
			name:              "get-without-body",
			method:            http.MethodGet,
			url:               "http://example.com/api",
			body:              nil,
			expectErr:         false,
			expectContentType: false,
		},
		{
			name:              "post-with-body",
			method:            http.MethodPost,
			url:               "http://example.com/api",
			body:              map[string]string{"key": "value"},
			expectErr:         false,
			expectContentType: true,
		},
		{
			name:              "invalid-url",
			method:            http.MethodGet,
			url:               ":",
			body:              nil,
			expectErr:         true,
			expectContentType: false,
		},
		{
			name:              "unmarshalable-body",
			method:            http.MethodPost,
			url:               "http://example.com/api",
			body:              make(chan int),
			expectErr:         true,
			expectContentType: false,
		},
		{
			name:              "put-with-complex-body",
			method:            http.MethodPut,
			url:               "http://example.com/api",
			body:              &entity.GenerateTestRequest{Prompt: "test", UserId: "u1", ChatId: "c1"},
			expectErr:         false,
			expectContentType: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &http.Client{}
			repo, err := NewAutotesterAPIRepository(logger, client, "http://example.com")
			require.NoError(t, err)

			concreteRepo := repo.(*autotesterAPIRepository)
			req, err := concreteRepo.newJSONRequest(context.Background(), test.method, test.url, test.body, token)

			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, req)
			require.Equal(t, test.method, req.Method)
			require.Equal(t, test.url, req.URL.String())
			require.Equal(t, "Bearer "+token, req.Header.Get("Authorization"))
			if test.expectContentType {
				require.Equal(t, "application/json", req.Header.Get("Content-Type"))
			} else {
				require.Empty(t, req.Header.Get("Content-Type"))
			}
		})
	}
}

// nolint:funlen
func TestDoAndDecode(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectErr    bool
		expectedData string
		useNilResult bool
		nilClient    bool
		nilLogger    bool
		nilReq       bool
	}{
		{
			name:         "success-with-data",
			statusCode:   http.StatusOK,
			responseBody: `{"data":"test value"}`,
			expectErr:    false,
			expectedData: "test value",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "non-200-status",
			statusCode:   http.StatusBadRequest,
			responseBody: `error message`,
			expectErr:    true,
			expectedData: "",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "invalid-json",
			statusCode:   http.StatusOK,
			responseBody: `{"data":`,
			expectErr:    true,
			expectedData: "",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "empty-response",
			statusCode:   http.StatusOK,
			responseBody: `{}`,
			expectErr:    false,
			expectedData: "",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "nil-result-success",
			statusCode:   http.StatusOK,
			responseBody: `{"data":"ignored"}`,
			expectErr:    false,
			expectedData: "",
			useNilResult: true,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "status-404",
			statusCode:   http.StatusNotFound,
			responseBody: `not found`,
			expectErr:    true,
			expectedData: "",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "status-500",
			statusCode:   http.StatusInternalServerError,
			responseBody: `server error`,
			expectErr:    true,
			expectedData: "",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "nil-client",
			statusCode:   http.StatusOK,
			responseBody: `{"data":"ignored"}`,
			expectErr:    true,
			expectedData: "",
			useNilResult: false,
			nilClient:    true,
			nilLogger:    false,
			nilReq:       false,
		},
		{
			name:         "nil-logger",
			statusCode:   http.StatusOK,
			responseBody: `{"data":"ignored"}`,
			expectErr:    true,
			expectedData: "",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    true,
			nilReq:       false,
		},
		{
			name:         "nil-req",
			statusCode:   http.StatusOK,
			responseBody: `{"data":"ignored"}`,
			expectErr:    true,
			expectedData: "",
			useNilResult: false,
			nilClient:    false,
			nilLogger:    false,
			nilReq:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.statusCode)
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer srv.Close()

			var client *http.Client
			var testLogger *slog.Logger
			var req *http.Request
			var err error

			if test.nilClient {
				client = nil
			} else {
				client = srv.Client()
			}

			if test.nilLogger {
				testLogger = nil
			} else {
				testLogger = logger
			}

			if test.nilReq {
				req = nil
			} else {
				req, err = http.NewRequest(http.MethodGet, srv.URL, nil)
				require.NoError(t, err)
			}

			type testResponse struct {
				Data string `json:"data"`
			}

			if test.useNilResult {
				err = doAndDecode[testResponse](client, testLogger, req, nil)
			} else {
				var result testResponse
				err = doAndDecode(client, testLogger, req, &result)
				if !test.expectErr && err == nil {
					require.Equal(t, test.expectedData, result.Data)
				}
			}

			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
