package service

/*
import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type TagExtractionService interface {
	KeywordTagExtractor(prompt string, tags []string) ([]string, error)
}

// KeywordTagExtractor extracts tags from a given prompt based on a list of tags.

type tagExtractionService struct {
	logger *slog.Logger
}

func NewTagExtractionService(logger *slog.Logger) TagExtractionService {
	return &tagExtractionService{
		logger: logger,
	}
}

func (k *tagExtractionService) KeywordTagExtractor(prompt string) ([]string, error) {

	tags[] = dbService.GetAllTags()

	if prompt == "" {
		return nil, errors.New("empty prompt")
	}

	// Check if any existing tags are provided in the prompt
	var extractedTags []string
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(prompt), strings.ToLower(tag)) {
			extractedTags = append(extractedTags, tag)
		}
	}

	if len(extractedTags) == 0 {
		return nil, fmt.Errorf("no tags found in prompt: %s", prompt)
	}

	return extractedTags, nil
}

*/
