package entity

import "encoding/json"

// SupplierResponse represents the response from a supplier.
// It contains the HTTP status code and a list of supplier offers.
type SupplierResponse struct {
	HTTPStatusCode int               `json:"httpstatuscode"`
	Data           SupplierOfferList `json:"data"`
}

// SupplierOfferList represents a list of supplier offers, each represented as a JSON object.
type SupplierOfferList struct {
	Items []json.RawMessage `json:"items"`
}
