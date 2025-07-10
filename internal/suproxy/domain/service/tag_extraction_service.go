package service

import (
	"errors"
	"fmt"
	"strings"
)

func KeywordTagExtractor(prompt string, tags []string) ([]string, error) {
	if prompt == "" {
		return nil, errors.New("empty prompt")
	}

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
