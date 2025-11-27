package handler

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
)

type dummyChatStorageService struct{}

var _ service.ChatStorageService = (*dummyChatStorageService)(nil)

func (d *dummyChatStorageService) SaveChat(ctx context.Context, chat *entity.Chat) error {
	return nil
}

func (d *dummyChatStorageService) LoadChat(ctx context.Context, userId string, chatId string) (*entity.Chat, error) {
	return nil, nil
}

func (d *dummyChatStorageService) LoadUserChats(ctx context.Context, userId string) ([]*entity.ChatSummary, error) {
	return nil, nil
}
