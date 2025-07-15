package entity

import "encoding/json"

// SupplierResponse
type SupplierResponse struct {
	HTTPStatusCode int               `json:"httpstatuscode"`
	Data           SupplierOfferList `json:"data"`
}

type SupplierOfferList struct {
	Items []json.RawMessage `json:"items"`
}
