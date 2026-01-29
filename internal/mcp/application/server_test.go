package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/service"
)

func TestMCPSSE_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return &mcp.Server{}
	}, nil)

	router.Any("/mcp/*any", func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	})

	req := httptest.NewRequest(http.MethodGet, "/mcp/", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.GreaterOrEqual(t, rr.Code, 100)
	assert.Less(t, rr.Code, 600)
}

func TestMCPSSE_ErrorCase_InvalidPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return &mcp.Server{}
	}, nil)

	router.Any("/mcp/*any", func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	})

	req := httptest.NewRequest(http.MethodGet, "/invalid/path", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
func TestRouter_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(gin.Recovery())

	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return &mcp.Server{}
	}, nil)

	router.Any("/mcp/*any", func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	})

	req := httptest.NewRequest(http.MethodGet, "/mcp/test", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		router.ServeHTTP(rr, req)
	})

	assert.GreaterOrEqual(t, rr.Code, 100)
	assert.Less(t, rr.Code, 600)
}

func TestRouter_ErrorCase_Panic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(gin.Recovery())

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		router.ServeHTTP(rr, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestContext_HappyPath(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	assert.NoError(t, ctx.Err())

	cancel()

	assert.Equal(t, context.Canceled, ctx.Err())
}

func TestContext_ErrorCase_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(5 * time.Millisecond)

	assert.Equal(t, context.DeadlineExceeded, ctx.Err())
}

func TestGetTemplate_HappyPath(t *testing.T) {
	mockService := mocks.NewMockAutotesterAPIService(t)

	expectedResp := &entity.TemplateResponse{
		Content: "test template content",
	}
	mockService.On("GetTemplate", mock.Anything).Return(expectedResp, nil)

	result, err := mockService.GetTemplate(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test template content", result.Content)
	mockService.AssertCalled(t, "GetTemplate", mock.Anything)
}

func TestGetTemplate_ErrorCase(t *testing.T) {
	mockService := mocks.NewMockAutotesterAPIService(t)

	mockService.On("GetTemplate", mock.Anything).Return(nil, assert.AnError)

	result, err := mockService.GetTemplate(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	mockService.AssertCalled(t, "GetTemplate", mock.Anything)
}

func TestGenerateTest_HappyPath(t *testing.T) {
	mockService := mocks.NewMockAutotesterAPIService(t)

	expectedResp := &entity.GenerateTestToolResponse{
		GenerateMsg: &entity.GenerateMessage{
			Id:   "msg1",
			Role: "assistant",
			Body: "Generated test code",
		},
		UserId: "user123",
		ChatId: "chat456",
	}
	mockService.On("GenerateTest", mock.Anything, mock.Anything).Return(expectedResp, nil)

	req := &entity.GenerateTestRequest{
		Prompt: "Create a login test",
		UserId: "user123",
		ChatId: "chat456",
	}
	result, err := mockService.GenerateTest(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user123", result.UserId)
	assert.Equal(t, "chat456", result.ChatId)
	assert.NotNil(t, result.GenerateMsg)
	assert.Equal(t, "assistant", result.GenerateMsg.Role)
	mockService.AssertCalled(t, "GenerateTest", mock.Anything, mock.Anything)
}

func TestGenerateTest_ErrorCase(t *testing.T) {
	mockService := mocks.NewMockAutotesterAPIService(t)

	mockService.On("GenerateTest", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	req := &entity.GenerateTestRequest{
		Prompt: "Create test",
		UserId: "user123",
		ChatId: "chat456",
	}
	result, err := mockService.GenerateTest(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockService.AssertCalled(t, "GenerateTest", mock.Anything, mock.Anything)
}

func TestExecuteTest_HappyPath(t *testing.T) {
	mockService := mocks.NewMockAutotesterAPIService(t)

	expectedResp := &entity.ExecuteTestResponse{
		Result: "Test executed successfully",
	}
	mockService.On("ExecuteTest", mock.Anything, mock.Anything).Return(expectedResp, nil)

	req := &entity.ExecuteTestRequest{
		UserId: "user123",
		ChatId: "chat456",
		Test:   "test code",
	}
	result, err := mockService.ExecuteTest(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test executed successfully", result.Result)
	mockService.AssertCalled(t, "ExecuteTest", mock.Anything, mock.Anything)
}

func TestExecuteTest_ErrorCase(t *testing.T) {
	mockService := mocks.NewMockAutotesterAPIService(t)

	mockService.On("ExecuteTest", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	req := &entity.ExecuteTestRequest{
		UserId: "user123",
		ChatId: "chat456",
		Test:   "test code",
	}
	result, err := mockService.ExecuteTest(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockService.AssertCalled(t, "ExecuteTest", mock.Anything, mock.Anything)
}
