package entity

import (
	"encoding/json"
)

// UpdateResponse is an internal representation of the response payload used to update offers and recompute aggregated fields while preserving the original response structure.
type UpdateResponse struct {
	Data struct {
		// recalculated on every update
		ResultCount           int     `json:"resultcount"`
		AvailableOffers       int     `json:"availableoffers"`
		CalculatedResultCount int     `json:"calculatedresultcount"`
		MinPrice              float64 `json:"minprice"`
		MaxPrice              float64 `json:"maxprice"`

		// updated one-by-one, then reassembled
		Items []Item `json:"items"`

		// preserved or nulled by caller, not structurally modified here
		FilterOptions []json.RawMessage `json:"filterOptions"`
	} `json:"data"`
}

// Item represents an individual offer item in the response.
type Item struct {
	IsDirect          bool    `json:"isdirect"`
	Price             float64 `json:"price"`
	Currency          string  `json:"currency"`
	Availability      bool    `json:"availability"`
	Description       string  `json:"description"`
	OfferID           string  `json:"offerid"`
	DepartureDate     string  `json:"departuredate"`
	ReturnDate        string  `json:"returndate"`
	CheckInHotel      string  `json:"checkinhotel"`
	CheckOutHotel     string  `json:"checkouthotel"`
	OvernightDuration int     `json:"overnightduration"`
}
