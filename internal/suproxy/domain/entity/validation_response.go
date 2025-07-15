package entity

import "encoding/json"

// SupplierResponse
type SupplierResponse struct {
	HTTPStatusCode int               `json:"httpstatuscode"`
	Data           SupplierOfferList `json:"data"`
}

// SupplierOfferList represents a list of supplier offers, each represented as a JSON object.
type SupplierOfferList struct {
	Items []json.RawMessage `json:"items"`
}
