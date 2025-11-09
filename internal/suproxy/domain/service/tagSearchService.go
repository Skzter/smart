package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	wrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
)

// TagSearchService defines the interface for searching S3 file keys by a given tag
type TagSearchService interface {
	FindKeysByTag(ctx context.Context, tag string) ([]string, error)
}

// tagSearchService is the concrete implementation of TagSearchService
type tagSearchService struct {
	config *config.Config
	s3     wrapper.S3StorageWrapper
}

// NewTagSearchService creates a new TagSearchService instance with S3 wrapper
func NewTagSearchService(cfg *config.Config, s3 wrapper.S3StorageWrapper) (TagSearchService, error) {
	if err := assert.NotNil(cfg, s3); err != nil {
		return nil, err
	}
	return &tagSearchService{
		config: cfg,
		s3:     s3,
	}, nil
}

// FindKeysByTag searches for parquet files whose extracted keys match the given tags.
// It scans all available parquet files, extracts meaningful keys from their filenames,
// and returns the list of files that contain keys present in the search tag.
func (t *tagSearchService) FindKeysByTag(ctx context.Context, tags string) ([]string, error) {
	tags = strings.TrimSpace(tags)
	if tags == "" {
		return nil, fmt.Errorf("tag is empty after trimming")
	}

	parquetFiles, err := t.s3.ListParquetFiles(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keys: %w", err)
	}
	var matchingKeys []string
	for _, file := range parquetFiles {
		keys := extractKeysFromFile(file, t.config.EntryPrefix)
		if isTagInKeys(keys, tags) {
			matchingKeys = append(matchingKeys, file)
		}
	}

	return matchingKeys, nil
}

// extractKeysFromFilename parses a parquet filename and extracts meaningful keys from it.
// It removes the configured prefix and ".parquet" suffix, then splits the remaining
// filename by the seperator "-" to extract individual keys. If the last segment is a numeric value,
// it is excluded as it's typically a timestamp.
func extractKeysFromFile(parquetFile string, prefix string) []string {
	ParquetFileNoSuffix, ok := strings.CutSuffix(parquetFile, ".parquet")
	if !ok {
		return nil
	}
	ParquetFileKeysOnly, ok := strings.CutPrefix(ParquetFileNoSuffix, prefix)
	if !ok {
		return nil
	}

	// keys are seperated with "-" in filename
	keys := strings.Split(ParquetFileKeysOnly, "-")

	// last key maybe be a number
	if _, err := strconv.Atoi(keys[len(keys)-1]); err == nil {
		// cuts of last element
		keys = keys[:len(keys)-1]
	}
	return keys
}

// isTagInKeys checks if the given tag contains any of keys as a substring
func isTagInKeys(keys []string, tag string) bool {
	for _, key := range keys {
		if strings.Contains(tag, key) {
			return true
		}
	}
	return false
}
