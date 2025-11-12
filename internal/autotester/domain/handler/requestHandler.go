package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// AutotesterController is the controller for autotesting requests.
// It encapsulates logging and access to the OpenAI service.
type AutotesterController struct {
	config            *config.Config
	logger            *slog.Logger
	validationService service.ValidatePrompt
	generationService service.GeneratePrompt
	saveLocalService  service.TestcaseLocalStorageService
}

// NewAutotesterController creates a new AutotesterController.
// Returns an initialized controller or an error.
func NewAutotesterController(
	logger *slog.Logger,
	config *config.Config,
	validationService service.ValidatePrompt,
	generationService service.GeneratePrompt,
	saveLocalService service.TestcaseLocalStorageService,
) (*AutotesterController, error) {
	if err := assert.NotNil(logger, config, validationService, generationService, saveLocalService); err != nil {
		return nil, err
	}

	return &AutotesterController{
		logger:            logger,
		config:            config,
		validationService: validationService,
		generationService: generationService,
		saveLocalService:  saveLocalService,
	}, nil
}

// HandleGetTemplate processes a template request from the frontend.
func (a *AutotesterController) HandleGetTemplate(c *gin.Context) {
	if err := assert.StringNotEmpty(a.config.Template); err != nil {
		c.JSON(http.StatusTeapot, "")
		a.logger.Error(err.Error())
		return
	}

	c.JSON(http.StatusOK, entity.Template{Template: a.config.Template})
}

// HandleChatRequest processes a chat request from the frontend.
// Expects a JSON with UserRequestDTO and returns a response from the LLM.
func (a *AutotesterController) HandleChatRequest(c *gin.Context) {
	var userRequest entity.UserRequest

	if err := c.BindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("JSON binding failed", "error", err)
		return
	}

	// returns handled errors which can be given to frontend
	resp, err := a.serviceHandler(c, userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleUserInfoRequest processes a request for user information.
// Expects a JSON with UserRequestDTO and returns a ResponseForUser.
func (a *AutotesterController) HandleUserInfoRequest(c *gin.Context) {
	var body entity.UserRequest
	var resp entity.ResponseForUser
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}
	c.JSON(http.StatusOK, entity.ResponseForUser{LogStamp: resp.LogStamp, SessionId: resp.SessionId})
}

// serviceHandler calls the OpenAI service and prepares the response for the frontend.
// It takes a gin.Context and UserRequest as input parameters.
// First validates the user prompt through validationService, then generates a response through generationService.
// Returns a ResponseForUser containing the generated text and user metadata, or an error if validation or generation fails.
func (a *AutotesterController) serviceHandler(c *gin.Context, userRequest entity.UserRequest) (*entity.ResponseForUser, error) {
	valid, msg, err := a.validationService.ValidatePrompt(c, userRequest.Message.MessageBody, userRequest.SessionId)
	if err != nil {
		return nil, err
	}

	if !valid {
		return &entity.ResponseForUser{
			Message:   sharedEntity.Message{MessageBody: msg},
			UserId:    userRequest.UserId,
			SessionId: userRequest.SessionId,
			LogStamp:  userRequest.LogStamp,
		}, nil
	}

	resp, err := a.generationService.GeneratePrompt(c, userRequest.Message.MessageBody, userRequest.SessionId)
	if err != nil {
		return nil, err
	}

	return &entity.ResponseForUser{
		Message:   sharedEntity.Message{MessageBody: resp},
		UserId:    userRequest.UserId,
		SessionId: userRequest.SessionId,
		LogStamp:  userRequest.LogStamp,
	}, nil
}

// HandleSaveLocalRequest processes a request to save a test case locally.
// Expects a JSON with LocalSaveRequest and creates a new test case with status NotRun.
// Returns a LocalSaveResponse with the generated test case ID or an error if saving fails.
func (a *AutotesterController) HandleSaveLocalRequest(c *gin.Context) {
	var localSaveRequest entity.LocalSaveRequest

	if err := c.BindJSON(&localSaveRequest); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	testcaseToSave := &entity.TestCase{
		TestID: uuid.New().String(),
		TestCode: entity.TestCode{
			Code: localSaveRequest.Code,
		},
		Status: entity.TestStatusNotRun,
	}

	if err := a.saveLocalService.Save(testcaseToSave, localSaveRequest.UserId, localSaveRequest.ConversationId); err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Saving failed due to internal server error"})
		return
	}

	c.JSON(http.StatusOK, entity.LocalSaveResponse{TestcaseId: testcaseToSave.TestID, Action: "saved"})
}

// HandleDeleteLocalRequest processes a request to delete a locally stored test case.
// Expects query parameters with testcaseId, userId and conversationId.
// Returns a LocalDeleteResponse confirming the deletion or an error if the deletion fails.
func (a *AutotesterController) HandleDeleteLocalRequest(c *gin.Context) {
	var deleteLocalRequest entity.LocalDeleteRequest

	if err := c.ShouldBindQuery(&deleteLocalRequest); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	if deleteLocalRequest.TestcaseId == "" || deleteLocalRequest.UserId == "" || deleteLocalRequest.ConversationId == "" {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Missing required parameters"})
		return
	}

	if err := a.saveLocalService.Delete(deleteLocalRequest.TestcaseId, deleteLocalRequest.UserId, deleteLocalRequest.ConversationId); err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Deleting failed due to internal server error"})
		return
	}

	c.JSON(http.StatusOK, entity.LocalDeleteResponse{TestcaseId: deleteLocalRequest.TestcaseId, Action: "deleted"})
}

// HandleRunContainer handles the running of the container and returns content of logfile from test
func (a *AutotesterController) HandleRunContainer(c *gin.Context) {
	type parameters struct {
		UserID         string `json:"userId"`
		ConversationID string `json:"conversationId"`
		SessionID      string `json:"sessionId"`
	}

	var params parameters
	if err := c.ShouldBindJSON(&params); err != nil {
		a.logger.Info(fmt.Sprintf("currently no request body: err => %s", err.Error()))
		// c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		// return
	}

	logFile := "docker/logs/output.log"

	a.logger.Info("Running the task")

	// find file to mount
	testfile, err := a.saveLocalService.GetTestPath(params.ConversationID, params.UserID, params.SessionID)
	if err != nil {
		a.logger.Info(fmt.Sprintf("file not available: %s\n", testfile))
	}

	// hier als letze arg das testfile fmt.Sprintf("TEST_FILE=%s", testfile) hinzufügen
	// so default cfg
	cmd := exec.Command("go", "tool", "task", "run-autopw-test")
	if err := cmd.Run(); err != nil {
		// Command failed, but we still try to read the log file
		fmt.Printf("Command execution error: %v\n", err)
	}
	a.logger.Info("Reading log file")
	content, err := os.ReadFile(logFile)
	if err != nil {
		c.String(500, "Failed to read log file: %v", err)
		return
	}
	a.logger.Info(string(content))
	c.String(200, string(content))
}
