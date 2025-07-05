package entity

import "encoding/json"

type ValidationResponse struct {
	HTTPStatusCode int            `json:"httpstatuscode"`
	Data           ValidationData `json:"data"`
}

type ValidationData struct {
	Items []json.RawMessage `json:"items"`
}

type ValidationItem struct {
	Duration          int                         `json:"duration"`
	DepartureDate     string                      `json:"departuredate"`
	ReturnDate        string                      `json:"returndate"`
	OvernightDuration ValidationOvernightDuration `json:"overnightduration"`

	Rooms              []ValidationRoom        `json:"rooms"`
	Accommodation      ValidationAccommodation `json:"accommodation"`
	Flight             ValidationFlight        `json:"flight"`
	AlternativeFlights []interface{}           `json:"alternativeflights"`
	Transfer           bool                    `json:"transfer"`
	CarRental          bool                    `json:"carrental"`
	PrivateTransfer    bool                    `json:"privatetransfer"`
	RailAndFly         bool                    `json:"railandfly"`
	CancellationFree   bool                    `json:"cancellation_free"`
	RebookingFree      bool                    `json:"rebooking_free"`

	Travelers []ValidationTravelerWithPrice `json:"travelers"`
	IsDirect  bool                          `json:"isdirect"`
}

type ValidationOvernightDuration struct {
	CheckInHotel  string `json:"checkinhotel"`
	CheckOutHotel string `json:"checkouthotel"`
	NightsInHotel int    `json:"nightsinhotel"`
}

type ValidationRoom struct {
	ID                 int             `json:"id"`
	MealType           string          `json:"mealtype"`
	OceanView          bool            `json:"oceanview"`
	PartialOceanView   bool            `json:"partialoceanview"`
	Price              ValidationPrice `json:"price"`
	RoomDescription    string          `json:"roomdescription"`
	RoomType           string          `json:"roomtype"`
	TravelerAssignment []string        `json:"travelerassignment"`
	RoomFacilities     []string        `json:"roomfacilities"`
	RoomTypes          []string        `json:"roomtypes"`
}

type ValidationAccommodation struct {
	CheckInDate       string               `json:"checkindate"`
	CheckOutDate      string               `json:"checkoutdate"`
	HotelBookingCode  string               `json:"hotelbookingcode"`
	ObjectID          string               `json:"objectid"`
	OfferHash         string               `json:"offerhash"`
	OptionPossible    bool                 `json:"optionpossible"`
	OTDSImportVersion string               `json:"otdsimportversion"`
	Rooms             []ValidationAccRoom  `json:"rooms"`
	SeasonOverlap     bool                 `json:"seasonoverlap"`
	Travelers         []ValidationTraveler `json:"travelers"`
}

type ValidationAccRoom struct {
	ID                 int      `json:"id"`
	MealBookingCode    string   `json:"mealbookingcode"`
	OceanView          bool     `json:"oceanview"`
	OfferHash          string   `json:"offerhash"`
	PartialOceanView   bool     `json:"partialoceanview"`
	PriceCacheID       float64  `json:"pricecacheid"`
	RoomBookingCode    string   `json:"roombookingcode"`
	TravelerAssignment []string `json:"travelerassignment"`
}

type ValidationTraveler struct {
	Age  int    `json:"age"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ValidationTravelerWithPrice struct {
	Age   int             `json:"age"`
	ID    string          `json:"id"`
	Type  string          `json:"type"`
	Price ValidationPrice `json:"price"`
}

type ValidationPrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type ValidationFlight struct {
	Price         ValidationPrice `json:"price"`
	OutboundPrice ValidationPrice `json:"outboundprice"`
	InboundPrice  ValidationPrice `json:"inboundprice"`
}
