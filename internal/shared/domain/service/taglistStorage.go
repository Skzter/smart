package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// DefaultTagList provides a default tag list
func DefaultTagList() *entity.TagList {
	return &entity.TagList{
		Tags: []entity.Tag{
			{Name: "no_hotelid", Description: "The response does not contain a hotelid field."},
			{Name: "no_outbound_flight", Description: "The response lacks outbound flight information (flight.outboundflightsegments is missing or empty)."},
			{Name: "no_return_flight", Description: "The response lacks return flight information (flight.inboundflightsegments is missing or empty)."},
			{Name: "no_price_info", Description: "Price information is missing (price.amount or equivalent fields are missing or empty)."},
			{Name: "no_mealtype", Description: "Meal type information is missing, such as rooms.mealtype or accommodation.rooms.mealbookingcode."},
			{Name: "no_duration_info", Description: "Duration is missing (duration or overnightduration fields are not present)."},
			{Name: "no_travelers", Description: "No traveler information is present (travelers is missing or empty)."},
			{Name: "no_offerid", Description: "No unique offer ID (offerid) is present."},
			{Name: "invalid_dates", Description: "Date information is implausible or inconsistent (e.g., return date before departure)."},
			{Name: "no_paymenttypes", Description: "Available payment methods are missing (paymenttypes is missing or empty)."},
			{Name: "no_cancellation_info", Description: "Cancellation conditions are completely missing (services.cancellationrebookingfee is not present)."},
			{Name: "no_emission_info", Description: "Emissions data (emissions.*) is completely missing."},
			{Name: "no_room_description", Description: "Room descriptions or equipment information are missing (e.g., rooms.roomdescription, rooms.roomfacilities)."},
			{Name: "no_transfer_info", Description: "Transfer information is missing (e.g., transfer, privatetransfer)."},
			{Name: "no_departuredate", Description: "The field departuredate is missing or empty."},
			{Name: "no_returndate", Description: "The field returndate is missing or empty."},
			{Name: "missing_flight_price", Description: "Flight price information is missing (e.g., flight.price.amount or flight.outboundprice.amount)."},
			{Name: "missing_room_price", Description: "Room price information is missing (e.g., rooms.price.amount)."},
			{Name: "missing_roomtype", Description: "The room type (rooms.roomtype) is not specified."},
			{Name: "no_offerhash", Description: "The field accommodation.rooms.offerhash is missing, although it is needed for offer identification."},
			{Name: "missing_touroperatorcode", Description: "The touroperatorcode field is missing."},
			{Name: "invalid_room_assignment", Description: "No traveler-to-room assignment is given (rooms.travelerassignment is missing or empty)."},
			{Name: "no_availability_info", Description: "Availability object is missing or contains no relevant information (e.g., availability.flight, availability.unit)."},
			{Name: "missing_enginecodes", Description: "One or more fields like accommodation.rooms.enginecodes.roomid or mealid are missing."},
			{Name: "missing_duration_component", Description: "Duration subfields like overnightduration.nightsinhotel or overnightduration.checkinhotel are missing."},
			{Name: "no_mealbookingcode", Description: "The field accommodation.rooms.mealbookingcode is missing."},
			{Name: "no_checkindate_or_checkoutdate", Description: "accommodation.checkindate or accommodation.checkoutdate is missing."},
			{Name: "missing_airport_codes", Description: "One or more fields such as departureairport.code or destinationairport.code are missing."},
			{Name: "no_personbookingcode", Description: "The field travelers.personbookingcode is missing – this is often required for traveler identification or pricing."},
			{Name: "no_bag_information", Description: "Baggage-related information is missing, such as included luggage, hand luggage allowances, or relevant fields in flight.baggage or services.baggage."},
		},
	}
}

// TaglistStorage defines the interface for managing taglists.
type TaglistStorage interface {
	// StoreTaglist updates stored taglist
	StoreTaglist(context.Context, *entity.TagList) error
	// GetTaglist retrieves Taglist from storage
	GetTaglist(ctx context.Context) (*entity.TagList, error)
}

// taglistStorage provides access to the database through the configured repository.
type taglistStorage struct {
	logger *slog.Logger
	repo   repository.TaglistStorage
}

// NewTaglistStorage creates a new instance of TaglistService.
func NewTaglistStorage(logger *slog.Logger, repo repository.TaglistStorage) (TaglistStorage, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}
	return &taglistStorage{
		logger: logger,
		repo:   repo,
	}, nil
}

// StoreTaglist updates stored taglist
func (d *taglistStorage) StoreTaglist(ctx context.Context, taglist *entity.TagList) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	// Check if taglist exists
	exists, err := d.repo.TaglistExists(ctx)
	if err != nil {
		return fmt.Errorf("S3 Error: %w", err)
	}

	// Create or update taglist
	if !exists {
		err = d.repo.CreateTaglist(ctx, taglist)
	} else {
		err = d.repo.UpdateTaglist(ctx, taglist)
	}
	if err != nil {
		return fmt.Errorf("failed to save taglist: %w", err)
	}

	d.logger.Debug("Taglist saved successfully", "taglist", taglist)
	return nil
}

// GetTaglist retrieves Taglist from storage
func (d *taglistStorage) GetTaglist(ctx context.Context) (*entity.TagList, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, fmt.Errorf("context cannot be nil, %w", err)
	}

	list, err := d.repo.ReadTaglist(ctx)

	if err != nil {
		return nil, err
	}

	return list, nil
}
