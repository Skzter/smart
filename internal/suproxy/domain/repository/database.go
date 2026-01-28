package repository

import (
	"context"
	"fmt"
	"strings"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// DatabaseRepository is the domain-facing contract.
type DatabaseRepository interface {
	CreateRequest(ctx context.Context, dbEntry entity.DatabaseEntry) error
	ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error)
	UpdateRequest(ctx context.Context, key string, dbEntry entity.DatabaseEntry) error
	DeleteRequest(ctx context.Context, key string) error
	ListAllKeys(ctx context.Context) ([]string, error)
}

// ValidateDatabaseEntry enforces domain invariants.
func ValidateDatabaseEntry(dbEntry entity.DatabaseEntry) error {
	if err := validateRequest(dbEntry.Request); err != nil {
		return fmt.Errorf("request invalid: %w", err)
	}
	if err := validateResponse(dbEntry.Response); err != nil {
		return fmt.Errorf("response invalid: %w", err)
	}
	if err := validateTags(dbEntry.Tags); err != nil {
		return fmt.Errorf("tags invalid: %w", err)
	}
	return nil
}

func validateRequest(rq entity.Request) error {
	if len(rq.Header) == 0 {
		return fmt.Errorf("header must not be empty")
	}
	if err := assert.StringNotEmpty(rq.Tags); err != nil {
		return err
	}
	if err := assert.StringNotEmpty(rq.Destination); err != nil {
		return err
	}
	if err := assert.StringNotEmpty(rq.Body); err != nil {
		return err
	}
	return nil
}

func validateResponse(rp entity.Response) error {
	return assert.StringNotEmpty(rp.Response)
}

func validateTags(t *sharedEntity.TagList) error {
	if t == nil || len(t.Tags) == 0 {
		return fmt.Errorf("tags must not be empty")
	}
	return nil
}

// GenerateKey is domain logic.
func GenerateKey(tags *sharedEntity.TagList, unixTimestamp string) string {
	names := make([]string, len(tags.Tags))
	for i, tag := range tags.Tags {
		names[i] = strings.ToLower(strings.TrimSpace(tag.Name))
	}
	return fmt.Sprintf("%s-%s", strings.Join(names, "-"), unixTimestamp)
}
