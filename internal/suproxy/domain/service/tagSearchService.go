package service

import (
	"context"
	"fmt"
	"strconv"
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

	parquetFiles, err := t.s3.ListParquetFiles(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keys: %w", err)
	}
	var matchingKeys []string
	for _, file := range parquetFiles {
		keys := extractKeysFromFile(file)
		for _, key := range keys {
			// check if key from current file substring from tag (prompt from request), if it is append file to array
			if strings.Contains(tag, key) {
				matchingKeys = append(matchingKeys, file)
			}
		}
	}

	return matchingKeys, nil
}

func extractKeysFromFile(parquetFile string) []string {
	// filename is: supplierData/no-hotelid_missing-tourdates.parquet
	// cuts suffix/prefix so only true filename remains
	ParquetFileNoSuffix, ok := strings.CutSuffix(parquetFile, ".parquet")
	if !ok {
		return nil
	}
	ParquetFileKeysOnly, ok := strings.CutPrefix(ParquetFileNoSuffix, "supplierData/")
	if !ok {
		return nil
	}
	// keys are seperated with "-" in filename
	keys := strings.Split(ParquetFileKeysOnly, "-")
	validKeys := []string{}
	for _, key := range keys {
		// sometimes number in filename, so if it errors its a string and true key
		if _, err := strconv.Atoi(key); err != nil {
			validKeys = append(validKeys, key)
		}
	}
	return validKeys
}
