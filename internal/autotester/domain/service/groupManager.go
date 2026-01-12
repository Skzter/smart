// Package service provides domain services for the autotester application.
// This package contains business logic implementations that coordinate between
// domain entities and infrastructure concerns.
package service

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GroupManager defines the interface for managing groups within the autotester domain.
// It provides operations for creating, listing, and managing the relationship between
// groups and chats. Groups serve as organizational units that can contain multiple chats.
type GroupManager interface {
	// List retrieves all groups from storage.
	// Returns a slice of Group entities or an error if the operation fails.
	List(ctx context.Context) ([]*entity.Group, error)

	// Create creates a new group with the specified name, description, and creator.
	// The group ID is automatically generated using UUID and returned.
	Create(ctx context.Context, name string, description string, creator string) (string, error)

	// AddChatToGroup associates a chat with a group.
	// Returns ErrChatAlreadyInGroup if the chat is already associated with the group.
	AddChatToGroup(ctx context.Context, groupId string, chatId string) error

	// RemoveChatFromGroup removes the association between a chat and a group.
	// Returns ErrChatNotInGroup if the chat is not currently associated with the group.
	RemoveChatFromGroup(ctx context.Context, groupId string, chatId string) error
}

// NewGroupManager creates a new instance of GroupManager with the provided dependencies.
// It requires a chat storage service, group storage, logger, and tracer.
// Returns an error if any of the required dependencies are nil.
func NewGroupManager(chatStorage ChatStorageService, groupStorage GroupStorage, logger *slog.Logger, tracer trace.Tracer) (GroupManager, error) {
	if err := assert.NotNil(chatStorage, groupStorage, logger, tracer); err != nil {
		return nil, err
	}
	return &groupManager{
		chatStorage:  chatStorage,
		groupStorage: groupStorage,
		logger:       logger,
		tracer:       tracer,
	}, nil
}

// groupManager is the concrete implementation of the GroupManager interface.
// It coordinates between chat and group storage services to manage group operations
// and maintains observability through logging and tracing.
type groupManager struct {
	chatStorage  ChatStorageService
	groupStorage GroupStorage
	logger       *slog.Logger
	tracer       trace.Tracer
}

// List retrieves all groups from the group storage.
// It includes distributed tracing support and returns all available groups.
// Returns an error if the storage operation fails.
func (g *groupManager) List(ctx context.Context) ([]*entity.Group, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	ctx, span := g.tracer.Start(ctx, "groupManager.List")
	defer span.End()

	groups, err := g.groupStorage.LoadAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while loading groups")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return groups, nil
}

// Create creates a new group with the provided details.
// It generates a new UUID for the group ID and sets the creation timestamp to the current UTC time.
// The group is immediately persisted to storage. Returns an error if storage operation fails.
func (g *groupManager) Create(ctx context.Context, name string, description string, creator string) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return "", err
	}

	ctx, span := g.tracer.Start(ctx, "groupManager.Create")
	defer span.End()

	group := entity.Group{
		Id:          uuid.NewString(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   creator,
	}

	if err := g.groupStorage.New(ctx, &group); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing group")
		return "", err
	}

	span.SetStatus(codes.Ok, "")
	return group.Id, nil
}

// AddChatToGroup associates a chat with a specific group.
// It first loads the chat to verify it exists, then checks if the chat is already
// associated with the group to prevent duplicates. If the association doesn't exist,
// it adds the group ID to the chat's groups list and saves the updated chat.
// This function logs errors, the one returned can be used for display in frontend
func (g *groupManager) AddChatToGroup(ctx context.Context, groupId string, chatId string) error {
	if err := assert.NotNil(ctx); err != nil {
		g.logger.Error("assertion failed", "err", err)
		return errors.ErrInternalServer
	}

	ctx, span := g.tracer.Start(ctx, "groupManager.AddChatToGroup")
	defer span.End()

	chat, err := g.chatStorage.LoadChat(ctx, chatId)
	if err != nil {
		span.RecordError(err)
		g.logger.Error("error while loading chat", "err", err)
		span.SetStatus(codes.Error, "error while loading chat")
		return errors.ErrInternalServer
	}

	if slices.Contains(chat.Groups, groupId) {
		span.SetStatus(codes.Error, "chat is already in group")
		g.logger.Error("chat is already in group")
		return errors.ErrChatAlreadyInGroup
	}

	chat.Groups = append(chat.Groups, groupId)

	if err := g.chatStorage.SaveChat(ctx, chat); err != nil {
		span.RecordError(err)
		g.logger.Error("error while loading storing", "err", err)
		span.SetStatus(codes.Error, "error while storing modified chat")
		return errors.ErrInternalServer
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// RemoveChatFromGroup removes the association between a chat and a group.
// It loads the chat, finds the group ID in the chat's groups list, removes it,
// and saves the updated chat back to storage.
// Returns ErrChatNotInGroup if the chat is not currently associated with the group.
// This function logs errors, the one returned can be used for display in frontend
func (g *groupManager) RemoveChatFromGroup(ctx context.Context, groupId string, chatId string) error {
	if err := assert.NotNil(ctx); err != nil {
		g.logger.Error("assertion failed", "err", err)
		return err
	}

	ctx, span := g.tracer.Start(ctx, "groupManager.RemoveChatFromGroup")
	defer span.End()

	chat, err := g.chatStorage.LoadChat(ctx, chatId)
	if err != nil {
		span.RecordError(err)
		g.logger.Error("error while loading chat", "err", err)
		span.SetStatus(codes.Error, "error while loading chat")
		return err
	}

	index := slices.Index(chat.Groups, groupId)
	if index == -1 {
		span.RecordError(err)
		span.SetStatus(codes.Error, "deleting non-existant group from chat")
		return errors.ErrChatNotInGroup
	}

	chat.Groups = slices.Delete(chat.Groups, index, index+1)

	if err := g.chatStorage.SaveChat(ctx, chat); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while storing modified chat")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
