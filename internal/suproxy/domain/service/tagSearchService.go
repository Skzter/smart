package service

import (
	"context"
	"fmt"
	"strings"

	wrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// TagSearchService defines the interface for searching S3 file keys by a given tag
type TagSearchService interface {
	FindKeysByTag(ctx context.Context, tag string) ([]string, error)
}

// tagSearchService is the concrete implementation of TagSearchService
type tagSearchService struct {
	s3 wrapper.S3StorageWrapper
}

// NewTagSearchService creates a new TagSearchService instance with S3 wrapper
func NewTagSearchService(s3 wrapper.S3StorageWrapper) TagSearchService {
	return &tagSearchService{
		s3: s3,
	}
}

// FindKeysByTag searches for all S3 file keys that contain the given tag string
// It trims whitespace, validates the tag, retrieves all keys from S3, and filters matching ones
func (t *tagSearchService) FindKeysByTag(ctx context.Context, tag string) ([]string, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return nil, fmt.Errorf("tag is empty after trimming")
	}

	keys, err := t.s3.ListParquetFiles(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keys: %w", err)
	}

	var matchingKeys []string
	for _, key := range keys {
		if strings.Contains(key, tag) {
			matchingKeys = append(matchingKeys, key)
		}
	}

	return matchingKeys, nil
}
