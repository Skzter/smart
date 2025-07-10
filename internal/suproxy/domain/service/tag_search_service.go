package service

/*
import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	//datenbank-service
)

type TagSearchService interface {
	FindKeysByTag(ctx context.Context, tag string) ([]string, error)
}

type tagSearchService struct {
	logger slog.Logger
	db     service.DatabaseService
}

func NewTagSearchService(logger slog.Logger, db service.DatabaseService) TagSearchService {
	return &tagSearchService{
		logger: logger,
		db:     db,
	}
}

func (t *tagSearchService) FindKeysByTag(ctx context.Context, tag string) ([]string, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		t.logger.Debug("Received empty tag after trimming")
		return nil, fmt.Errorf("tag is empty after trimming")
	}

	allKeys, err := t.db.GetAllKeys(ctx)
	if err != nil {
		t.logger.Error("Failed to fetch keys from database", "error", err)
		return nil, fmt.Errorf("failed to fetch keys: %w", err)
	}

	var matchingKeys []string
	for _, key := range allKeys {
		if strings.Contains(key, tag) {
			t.logger.Debug("Matching key found", "key", key)
			matchingKeys = append(matchingKeys, key)
		}
	}

	t.logger.Info("Tag search completed", "tag", tag, "matches", len(matchingKeys))
	return matchingKeys, nil
}

*/
