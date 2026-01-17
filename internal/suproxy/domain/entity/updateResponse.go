package entity

import "encoding/json"

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
		Items []json.RawMessage `json:"items"`

		// preserved or nulled by caller, not structurally modified here
		FilterOptions []json.RawMessage `json:"filterOptions"`
	} `json:"data"`
}
