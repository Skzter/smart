package entity

type SupplierOfferList struct {
	HTTPStatusCode int
	Offers         []SupplierOffer
}

type SupplierOffer struct {
	Data []byte
}
