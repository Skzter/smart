package entity

// ODTResponse represents the top-level response structure returned by the ODT API.
type ODTResponse struct {
	Data ODTData `json:"data"`
}

// ODTData contains the list of offer items returned by the ODT response.
type ODTData struct {
	Items []ODTItem `json:"items"`
}

// ODTItem represents a single travel offer including pricing, availability, and date details.
type ODTItem struct {
	IsDirect     bool    `json:"isdirect"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	Availability bool    `json:"availability"`
	Description  string  `json:"description"`
	OfferID      string  `json:"offerid"`

	DepartureDate string `json:"departuredate"`
	ReturnDate    string `json:"returndate"`
	CheckInHotel  string `json:"checkinhotel"`
	CheckOutHotel string `json:"checkouthotel"`

	OvernightDuration int `json:"overnightduration"`
}
