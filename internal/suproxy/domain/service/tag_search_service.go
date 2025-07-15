package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	wrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

type TagSearchService interface {
	FindKeysByTag(ctx context.Context, tag string) ([]string, error)
}

type tagSearchService struct {
	logger slog.Logger
	s3     wrapper.S3StorageWrapper
}

func NewTagSearchService(logger slog.Logger, s3 wrapper.S3StorageWrapper) TagSearchService {
	return &tagSearchService{
		logger: logger,
		s3:     s3,
	}
}

func (t *tagSearchService) FindKeysByTag(ctx context.Context, tag string) ([]string, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		t.logger.Debug("Received empty tag after trimming")
		return nil, fmt.Errorf("tag is empty after trimming")
	}

	keys, err := t.s3.ListParquetFiles(ctx, "")
	if err != nil {
		t.logger.Error("Failed to fetch keys from S3", "error", err)
		return nil, fmt.Errorf("failed to fetch keys: %w", err)
	}

	var matchingKeys []string
	for _, key := range keys {
		if strings.Contains(key, tag) {
			t.logger.Debug("Matching key found", "key", key)
			matchingKeys = append(matchingKeys, key)
		}
	}

	t.logger.Info("Tag search completed", "tag", tag, "matches", len(matchingKeys))
	return matchingKeys, nil
}
