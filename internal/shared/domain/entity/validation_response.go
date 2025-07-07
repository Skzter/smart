package entity

import "encoding/json"

type SupplierOfferResponse struct {
	HTTPStatusCode int               `json:"httpstatuscode"`
	Data           SupplierOfferData `json:"data"`
}

type SupplierOfferData struct {
	Items []json.RawMessage `json:"items"`
}
