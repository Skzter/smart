package entity

type SupplierOfferResponse struct {
	HTTPStatusCode int    `json:"httpstatuscode"`
	Data           []byte `json:"data"`
}
