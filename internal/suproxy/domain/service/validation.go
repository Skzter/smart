package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	repository "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
)

type Validator struct {
	Connector              repository.OpenAI
	Logger                 *slog.Logger
	SystemPromptValidation string
}

type Response struct {
	HTTPStatusCode int  `json:"httpstatuscode"`
	Data           Data `json:"data"`
}

type Data struct {
	Items []Item `json:"items"`
}

type Item struct {
	Duration          int               `json:"duration"`
	DepartureDate     string            `json:"departuredate"`
	ReturnDate        string            `json:"returndate"`
	OvernightDuration OvernightDuration `json:"overnightduration"`

	Rooms              []Room        `json:"rooms"`
	Accommodation      Accommodation `json:"accommodation"`
	Flight             Flight        `json:"flight"`
	AlternativeFlights []interface{} `json:"alternativeflights"`
	Transfer           bool          `json:"transfer"`
	CarRental          bool          `json:"carrental"`
	PrivateTransfer    bool          `json:"privatetransfer"`
	RailAndFly         bool          `json:"railandfly"`
	CancellationFree   bool          `json:"cancellation_free"`
	RebookingFree      bool          `json:"rebooking_free"`

	Travelers []TravelerWithPrice `json:"travelers"`
	IsDirect  bool                `json:"isdirect"`
}

type OvernightDuration struct {
	CheckInHotel  string `json:"checkinhotel"`
	CheckOutHotel string `json:"checkouthotel"`
	NightsInHotel int    `json:"nightsinhotel"`
}

type Room struct {
	ID                 int      `json:"id"`
	MealType           string   `json:"mealtype"`
	OceanView          bool     `json:"oceanview"`
	PartialOceanView   bool     `json:"partialoceanview"`
	Price              Price    `json:"price"`
	RoomDescription    string   `json:"roomdescription"`
	RoomType           string   `json:"roomtype"`
	TravelerAssignment []string `json:"travelerassignment"`
	RoomFacilities     []string `json:"roomfacilities"`
	RoomTypes          []string `json:"roomtypes"`
}

type Accommodation struct {
	CheckInDate       string     `json:"checkindate"`
	CheckOutDate      string     `json:"checkoutdate"`
	HotelBookingCode  string     `json:"hotelbookingcode"`
	ObjectID          string     `json:"objectid"`
	OfferHash         string     `json:"offerhash"`
	OptionPossible    bool       `json:"optionpossible"`
	OTDSImportVersion string     `json:"otdsimportversion"`
	Rooms             []AccRoom  `json:"rooms"`
	SeasonOverlap     bool       `json:"seasonoverlap"`
	Travelers         []Traveler `json:"travelers"`
}

type AccRoom struct {
	ID                 int      `json:"id"`
	MealBookingCode    string   `json:"mealbookingcode"`
	OceanView          bool     `json:"oceanview"`
	OfferHash          string   `json:"offerhash"`
	PartialOceanView   bool     `json:"partialoceanview"`
	PriceCacheID       float64  `json:"pricecacheid"`
	RoomBookingCode    string   `json:"roombookingcode"`
	TravelerAssignment []string `json:"travelerassignment"`
}

type Traveler struct {
	Age  int    `json:"age"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

type TravelerWithPrice struct {
	Age   int    `json:"age"`
	ID    string `json:"id"`
	Type  string `json:"type"`
	Price Price  `json:"price"`
}

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Flight struct {
	Price         Price `json:"price"`
	OutboundPrice Price `json:"outboundprice"`
	InboundPrice  Price `json:"inboundprice"`
}

// itemsString converts data into a string for forwarding
func itemsString(items []Item) string {
	var result string
	for _, item := range items {
		result += fmt.Sprintf(
			"Duration: %d | DepartureDate: %s | ReturnDate: %s\n"+
				"OvernightDuration: %+v | Rooms: %+v | Accommodation: %+v | Flight: %+v\n"+
				"AlternativeFlights: %+v | Transfer: %t | CarRental: %t | PrivateTransfer: %t\n"+
				"RailAndFly: %t | CancellationFree: %t | RebookingFree: %t\n"+
				"Travelers: %+v | IsDirect: %t\n",
			item.Duration,
			item.DepartureDate,
			item.ReturnDate,
			item.OvernightDuration,
			item.Rooms,
			item.Accommodation,
			item.Flight,
			item.AlternativeFlights,
			item.Transfer,
			item.CarRental,
			item.PrivateTransfer,
			item.RailAndFly,
			item.CancellationFree,
			item.RebookingFree,
			item.Travelers,
			item.IsDirect,
		)
	}
	return strings.TrimSpace(result)
}

// NewValidator loads the configuration and returns the validator
func NewValidator(cfg *config.Config) *Validator {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	connector, err := repository.NewOpenAiRepository(logger, 5)
	if err != nil {
		log.Fatal(err)
	}

	return &Validator{
		Connector:              connector,
		Logger:                 logger,
		SystemPromptValidation: cfg.Prompts.ValidationPrompt,
	}
}

func (v *Validator) Validate(resp *Response) {
	if resp == nil {
		v.Logger.Warn("Validation skipped: Response is nil")
		return
	}

	if resp.HTTPStatusCode != 200 {
		v.Logger.Warn("Validation skipped: Invalid HTTP status", "status", resp.HTTPStatusCode)
		return
	}

	v.Logger.Info("Valid HTTP response. Forwarding to OpenAI...", "status", resp.HTTPStatusCode)

	req := entity.Request{
		Model:        "gpt-4o",
		Prompt:       itemsString(resp.Data.Items),
		SystemPrompt: v.SystemPromptValidation,
	}

	result, err := v.Connector.CreateRequest(context.Background(), req)
	if err != nil {
		v.Logger.Error("OpenAI request failed", "error", err)
		return
	}

	v.Logger.Info("OpenAI response received", "response", result.Text)
}
