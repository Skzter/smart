package entity

// ODTResponse represents the top-level response structure returned by the ODT API.
type ODTResponse struct {
	Data struct {
		Items []ODTItem `json:"items"`
	} `json:"data"`
}

// ODTItem represents a single travel offer including pricing, availability, and date details.
type ODTItem struct {
	OfferID            string            `json:"offerid"`
	TravelType         string            `json:"traveltype"`
	SupplierTravelType string            `json:"suppliertraveltype"`
	Description        string            `json:"description"`
	IsDirect           bool              `json:"isdirect"`
	Price              PriceInfo         `json:"price"`
	Availability       AvailabilityInfo  `json:"availability"`
	DepartureDate      string            `json:"departuredate"`
	ReturnDate         string            `json:"returndate"`
	Flight             FlightInfo        `json:"flight"`
	Accommodation      AccommodationInfo `json:"accommodation"`
	OvernightDuration  OvernightDuration `json:"overnightduration"`
}

// PriceInfo holds amount and currency for an offer's pricing.
type PriceInfo struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// AvailabilityInfo describes flight and unit availability counts.
type AvailabilityInfo struct {
	Flight string `json:"flight"`
	Unit   string `json:"unit"`
}

// FlightInfo captures airports and times for the outbound and inbound legs.
type FlightInfo struct {
	OutboundDepartureAirport   Airport `json:"outbounddepartureairport"`
	InboundDepartureAirport    Airport `json:"inbounddepartureairport"`
	OutboundDestinationAirport Airport `json:"outbounddestinationairport"`
	InboundDestinationAirport  Airport `json:"inbounddestinationairport"`
	OutboundDepartureTime      string  `json:"outbounddeparturetime"`
	InboundDepartureTime       string  `json:"inbounddeparturetime"`
	OutboundArrivalTime        string  `json:"outboundarrivaltime"`
	InboundArrivalTime         string  `json:"inboundarrivaltime"`
}

// Airport stores the IATA code for a flight airport.
type Airport struct {
	Code string `json:"code"`
}

// AccommodationInfo lists accommodation dates and rooms for the stay.
type AccommodationInfo struct {
	CheckInDate  string              `json:"checkindate"`
	CheckOutDate string              `json:"checkoutdate"`
	Rooms        []AccommodationRoom `json:"rooms"`
}

// AccommodationRoom describes the room and traveler assignments.
type AccommodationRoom struct {
	ID                 int               `json:"id"`
	TravelerAssignment []string          `json:"travelerassignment"`
	MealBookingCode    string            `json:"mealbookingcode"`
	RoomBookingCode    string            `json:"roombookingcode"`
	EngineCodes        map[string]string `json:"enginecodes"`
}

// OvernightDuration tracks the hotel check-in/out times and total nights.
type OvernightDuration struct {
	CheckInHotel  string `json:"checkinhotel"`
	CheckOutHotel string `json:"checkouthotel"`
	NightsInHotel int    `json:"nightsinhotel"`
}
