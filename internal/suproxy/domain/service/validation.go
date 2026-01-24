package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// Common validation errors returned by the Validator.
var (
	ErrResponseNil       = errors.New("response is nil")
	ErrInvalidHTTPStatus = errors.New("invalid HTTP status")
	ErrEmptyOpenAIResult = errors.New("openai returned empty result")
)

func emptyOfferTag() sharedEntity.Tag {
	return sharedEntity.Tag{
		Name:        "empty_offer",
		Description: "",
	}
}

// staticly generated tags for non-valid offers
const (
	NoOffersInResponse = "no_offer"
	ReponseNot200      = "non_200"
)

// OpenAIValidationResult models the expected JSON structure of the OpenAI validation response
type OpenAIValidationResult struct {
	Valid  bool               `json:"valid"`
	Reason []sharedEntity.Tag `json:"reason"`
}

// Validator defines an interface for validating a SupplierResponse.
// Implementations should provide specific validation logic.
type Validator interface {
	Validate(ctx context.Context, offers *entity.SupplierResponse, tagList *sharedEntity.TagList) (*sharedEntity.TagList, error)
}

// Validator encapsulates the logic for validating supplier offer responses
// It sends up to MaxItems individual offer prompts to an OpenAI service for consistency checks
type validator struct {
	openAiService sharedService.OpenAI
	Logger        *slog.Logger
	cfg           *config.Suproxy
	tracer        trace.Tracer
}

// NewValidator creates a new validator service with logger and configuration
func NewValidator(logger *slog.Logger, cfg *config.Suproxy, service sharedService.OpenAI, tracer trace.Tracer) (Validator, error) {
	if err := assert.NotNil(logger, cfg, service); err != nil {
		return nil, err
	}

	return &validator{
		openAiService: service,
		Logger:        logger,
		cfg:           cfg,
		tracer:        tracer,
	}, nil
}

// Validate processes a supplier offer response, extracts individual offers (items), and sends up to MaxItems of them
// to an OpenAI service for validation
// nolint:funlen
func (v validator) Validate(ctx context.Context, offers *entity.SupplierResponse, tagList *sharedEntity.TagList) (*sharedEntity.TagList, error) {
	if err := assert.NotNil(ctx, offers); err != nil {
		return nil, err
	}

	ctx, span := v.tracer.Start(ctx, "validator.Validate")
	defer span.End()

	if offers.HTTPStatusCode != http.StatusOK {
		return &sharedEntity.TagList{
			Tags: []sharedEntity.Tag{
				{
					Name:        ReponseNot200,
					Description: "",
				},
			},
		}, nil
	}

	if len(offers.Data.Items) == 0 {
		return &sharedEntity.TagList{
			Tags: []sharedEntity.Tag{
				{
					Name:        NoOffersInResponse,
					Description: "",
				},
			},
		}, nil
	}

	v.Logger.Debug("Valid offerlist. Beginning LMM validation")

	newTags := make([]sharedEntity.Tag, 0, 10)

	sysPrompt := fmt.Sprintf(v.cfg.Prompts.ValidationPrompt, tagList.Format())

	for i, offer := range offers.Data.Items {
		v.Logger.Debug(fmt.Sprintf("checking offers: %d/%d", i, v.cfg.MaxItemsPerValidation))
		if i >= v.cfg.MaxItemsPerValidation {
			break
		}

		item := string(offer)

		item = strings.TrimSpace(item)
		if item == "" {
			return &sharedEntity.TagList{
				Tags: []sharedEntity.Tag{emptyOfferTag()},
			}, nil
		}

		req := sharedEntity.Request{
			Model:        v.cfg.Model,
			Messages:     []*sharedEntity.Message{{Role: sharedEntity.RoleUser, Body: item}},
			SystemPrompt: sysPrompt,
		}
		msg, err := v.openAiService.Request(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "OpenAI request failed")
			return nil, err
		}

		if strings.TrimSpace(msg.Body) == "" {
			err := fmt.Errorf("empty openai result for req: %v", req)
			span.RecordError(err)
			span.SetStatus(codes.Error, "empty OpenAI result")
			return nil, err
		}

		var validationResult OpenAIValidationResult

		err = json.Unmarshal([]byte(msg.Body), &validationResult)

		if err != nil {
			err := fmt.Errorf("invalid OpenAI response format at index %d: %v", i, err)
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid OpenAI response format")
			return nil, err
		}

		if !validationResult.Valid {
			for _, tag := range validationResult.Reason {
				if !slices.ContainsFunc(newTags, func(t sharedEntity.Tag) bool {
					return t.Name == tag.Name
				}) {
					newTags = append(newTags, tag)
				}
			}
		}
	}

	// Return no tags if all offers are valid
	if len(newTags) == 0 {
		span.SetStatus(codes.Ok, "")
		return &sharedEntity.TagList{
			Tags: []sharedEntity.Tag{},
		}, nil
	}
	span.SetStatus(codes.Ok, "")
	return &sharedEntity.TagList{Tags: newTags}, nil
}
