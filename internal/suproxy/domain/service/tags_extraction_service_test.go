package service

import (
	"reflect"
	"testing"
)

func TestKeywordTagExtractor(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		tags         []string
		expectedBody []string
		expectErr    bool
	}{
		{
			name:         "Empty prompt",
			prompt:       "",
			tags:         []string{"no-bag", "missing-date"},
			expectedBody: nil,
			expectErr:    true,
		},
		{
			name:         "No matching tags",
			prompt:       "This is a test prompt",
			tags:         []string{"no-bag", "missing-date"},
			expectedBody: nil,
			expectErr:    true,
		},
		{
			name:         "One matching tag",
			prompt:       "Find all the requests that had missing-date",
			tags:         []string{"no-bag", "missing-date"},
			expectedBody: []string{"missing-date"},
			expectErr:    false,
		},
		{
			name:         "Multiple matching tags",
			prompt:       "Which requests had no-bag and missing-date?",
			tags:         []string{"no-bag", "missing-date", "invalid-date"},
			expectedBody: []string{"no-bag", "missing-date"},
			expectErr:    false,
		},
		{
			name:         "Case insensitive match",
			prompt:       "Find all the requests that had No-Bag",
			tags:         []string{"no-bag", "missing-date", "invalid-date"},
			expectedBody: []string{"no-bag"},
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeywordTagExtractor(tt.prompt, tt.tags)
			if (err != nil) != tt.expectErr {
				t.Errorf("KeywordTagExtractor() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expectedBody) {
				t.Errorf("KeywordTagExtractor() = %v, expectedBody %v", got, tt.expectedBody)
			}
		})
	}
}
