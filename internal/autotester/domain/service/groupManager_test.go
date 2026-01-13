package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

func TestNewGroupManager(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")
	chatStorage := mocks.NewMockChatStorageService(t)
	groupStorage := mocks.NewMockGroupStorage(t)

	tests := []struct {
		name         string
		chatStorage  ChatStorageService
		groupStorage GroupStorage
		logger       *slog.Logger
		tracer       trace.Tracer
		wantErr      bool
	}{
		{
			name:         "all dependencies provided",
			chatStorage:  chatStorage,
			groupStorage: groupStorage,
			logger:       logger,
			tracer:       tracer,
			wantErr:      false,
		},
		{
			name:         "nil chat storage",
			chatStorage:  nil,
			groupStorage: groupStorage,
			logger:       logger,
			tracer:       tracer,
			wantErr:      true,
		},
		{
			name:         "nil group storage",
			chatStorage:  chatStorage,
			groupStorage: nil,
			logger:       logger,
			tracer:       tracer,
			wantErr:      true,
		},
		{
			name:         "nil logger",
			chatStorage:  chatStorage,
			groupStorage: groupStorage,
			logger:       nil,
			tracer:       tracer,
			wantErr:      true,
		},
		{
			name:         "nil tracer",
			chatStorage:  chatStorage,
			groupStorage: groupStorage,
			logger:       logger,
			tracer:       nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewGroupManager(tt.chatStorage, tt.groupStorage, tt.logger, tt.tracer)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, svc)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestGroupManager_List(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name         string
		setupMock    func(*mocks.MockChatStorageService, *mocks.MockGroupStorage)
		ctx          context.Context
		expectErr    bool
		expectGroups int
	}{
		{
			name: "successfully list all groups",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				groups := []*entity.Group{
					{Id: "1", Name: "Group 1", Description: "First group", CreatedBy: "user1"},
					{Id: "2", Name: "Group 2", Description: "Second group", CreatedBy: "user2"},
				}
				groupStorage.On("ListAll", mock.Anything).Return(groups, nil).Once()
			},
			ctx:          context.Background(),
			expectErr:    false,
			expectGroups: 2,
		},
		{
			name: "empty group list",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				groupStorage.On("ListAll", mock.Anything).Return([]*entity.Group{}, nil).Once()
			},
			ctx:          context.Background(),
			expectErr:    false,
			expectGroups: 0,
		},
		{
			name: "storage error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				groupStorage.On("ListAll", mock.Anything).Return(nil, errors.New("storage error")).Once()
			},
			ctx:       context.Background(),
			expectErr: true,
		},
		{
			name: "nil context returns error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				// No mock calls expected
			},
			ctx:       nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatStorage := mocks.NewMockChatStorageService(t)
			groupStorage := mocks.NewMockGroupStorage(t)
			tt.setupMock(chatStorage, groupStorage)

			svc, err := NewGroupManager(chatStorage, groupStorage, logger, tracer)
			assert.Nil(t, err)

			got, err := svc.List(tt.ctx)

			if tt.expectErr {
				assert.NotNil(t, err)
				assert.Nil(t, got)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, got)
				assert.Len(t, got, tt.expectGroups)
			}
		})
	}
}

func TestGroupManager_Create(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockChatStorageService, *mocks.MockGroupStorage)
		ctx         context.Context
		groupName   string
		description string
		creator     string
		expectErr   bool
	}{
		{
			name: "successfully create group",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				groupStorage.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Group) bool {
					return g.Name == "Test Group" &&
						g.Description == "Test Description" &&
						g.CreatedBy == "testuser" &&
						g.Id != "" &&
						!g.CreatedAt.IsZero()
				})).Return(nil).Once()
			},
			ctx:         context.Background(),
			groupName:   "Test Group",
			description: "Test Description",
			creator:     "testuser",
			expectErr:   false,
		},
		{
			name: "storage error during create",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				groupStorage.On("Create", mock.Anything, mock.Anything).Return(errors.New("storage error")).Once()
			},
			ctx:         context.Background(),
			groupName:   "Test Group",
			description: "Test Description",
			creator:     "testuser",
			expectErr:   true,
		},
		{
			name: "nil context returns error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				// No mock calls expected
			},
			ctx:         nil,
			groupName:   "Test Group",
			description: "Test Description",
			creator:     "testuser",
			expectErr:   true,
		},
		{
			name: "empty name still creates group",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				groupStorage.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Group) bool {
					return g.Name == "" &&
						g.Description == "Description" &&
						g.CreatedBy == "user"
				})).Return(nil).Once()
			},
			ctx:         context.Background(),
			groupName:   "",
			description: "Description",
			creator:     "user",
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatStorage := mocks.NewMockChatStorageService(t)
			groupStorage := mocks.NewMockGroupStorage(t)
			tt.setupMock(chatStorage, groupStorage)

			svc, err := NewGroupManager(chatStorage, groupStorage, logger, tracer)
			assert.Nil(t, err)

			groupId, err := svc.Create(tt.ctx, tt.groupName, tt.description, tt.creator)

			if tt.expectErr {
				assert.NotNil(t, err)
				assert.Empty(t, groupId)
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, groupId)
				assert.Nil(t, uuid.Validate(groupId))
			}
		})
	}
}

// nolint:funlen
func TestGroupManager_AddChatToGroup(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")
	chatId := "chat1"
	userId := "user1"
	groupId := "group1"

	tests := []struct {
		name      string
		setupMock func(*mocks.MockChatStorageService, *mocks.MockGroupStorage)
		ctx       context.Context
		groupId   string
		chatId    string
		expectErr bool
		errType   error
	}{
		{
			name: "successfully add chat to group",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{}, // Empty groups
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				chatStorage.On("SaveChat", mock.Anything, mock.MatchedBy(func(c *entity.Chat) bool {
					return c.Id == chatId && len(c.Groups) == 1 && c.Groups[0] == groupId
				})).Return(nil).Once()
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: false,
		},
		{
			name: "add chat to group when chat has other groups",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{"group2", "group3"}, // Already has other groups
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				chatStorage.On("SaveChat", mock.Anything, mock.MatchedBy(func(c *entity.Chat) bool {
					return c.Id == chatId && len(c.Groups) == 3 &&
						c.Groups[0] == "group2" && c.Groups[1] == "group3" && c.Groups[2] == groupId
				})).Return(nil).Once()
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: false,
		},
		{
			name: "chat already in group returns error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{groupId}, // Already in target group
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				// No SaveChat call expected
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
			errType:   sharedErrors.ErrChatAlreadyInGroup,
		},
		{
			name: "error loading chat",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(nil, errors.New("load error")).Once()
				// No SaveChat call expected
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
		},
		{
			name: "error saving chat",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{},
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				chatStorage.On("SaveChat", mock.Anything, mock.Anything).Return(errors.New("save error")).Once()
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
		},
		{
			name: "nil context returns error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				// No mock calls expected
			},
			ctx:       nil,
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatStorage := mocks.NewMockChatStorageService(t)
			groupStorage := mocks.NewMockGroupStorage(t)
			tt.setupMock(chatStorage, groupStorage)

			svc, err := NewGroupManager(chatStorage, groupStorage, logger, tracer)
			assert.Nil(t, err)

			err = svc.AddChatToGroup(tt.ctx, tt.groupId, tt.chatId)

			if tt.expectErr {
				assert.NotNil(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// nolint:funlen
func TestGroupManager_RemoveChatFromGroup(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")
	chatId := "chat1"
	userId := "user1"
	groupId := "group1"

	tests := []struct {
		name      string
		setupMock func(*mocks.MockChatStorageService, *mocks.MockGroupStorage)
		ctx       context.Context
		groupId   string
		chatId    string
		expectErr bool
		errType   error
	}{
		{
			name: "successfully remove chat from group",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{groupId}, // Chat is in the group
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				chatStorage.On("SaveChat", mock.Anything, mock.MatchedBy(func(c *entity.Chat) bool {
					// After slices.Delete, the slice should be empty
					return c.Id == chatId && len(c.Groups) == 0
				})).Return(nil).Once()
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: false,
		},
		{
			name: "remove chat from group with multiple groups",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{groupId, "group2", "group3"}, // Chat is in multiple groups
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				chatStorage.On("SaveChat", mock.Anything, mock.MatchedBy(func(c *entity.Chat) bool {
					// After removing groupId from [groupId, "group2", "group3"]
					// Result should be ["group2", "group3"]
					return c.Id == chatId && len(c.Groups) == 2 &&
						c.Groups[0] == "group2" && c.Groups[1] == "group3"
				})).Return(nil).Once()
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: false,
		},
		{
			name: "chat not in group returns error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{"group2", "group3"}, // Chat is not in target group
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				// No SaveChat call expected
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
			errType:   sharedErrors.ErrChatNotInGroup,
		},
		{
			name: "empty groups list returns error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{}, // No groups
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				// No SaveChat call expected
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
			errType:   sharedErrors.ErrChatNotInGroup,
		},
		{
			name: "error loading chat",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(nil, errors.New("load error")).Once()
				// No SaveChat call expected
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
		},
		{
			name: "error saving chat",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				chat := &entity.Chat{
					Id:     chatId,
					Author: userId,
					Groups: []string{groupId},
				}
				chatStorage.On("LoadChat", mock.Anything, chatId).Return(chat, nil).Once()
				chatStorage.On("SaveChat", mock.Anything, mock.Anything).Return(errors.New("save error")).Once()
			},
			ctx:       context.Background(),
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
		},
		{
			name: "nil context returns error",
			setupMock: func(chatStorage *mocks.MockChatStorageService, groupStorage *mocks.MockGroupStorage) {
				// No mock calls expected
			},
			ctx:       nil,
			groupId:   groupId,
			chatId:    chatId,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatStorage := mocks.NewMockChatStorageService(t)
			groupStorage := mocks.NewMockGroupStorage(t)
			tt.setupMock(chatStorage, groupStorage)

			svc, err := NewGroupManager(chatStorage, groupStorage, logger, tracer)
			assert.Nil(t, err)

			err = svc.RemoveChatFromGroup(tt.ctx, tt.groupId, tt.chatId)

			if tt.expectErr {
				assert.NotNil(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
