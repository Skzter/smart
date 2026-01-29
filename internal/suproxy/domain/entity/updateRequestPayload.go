package entity

// UpdateRequestPayload represents the HTTP request payload wrapper containing one or more request parameter sets.
type UpdateRequestPayload struct {
	Params []RequestBody `json:"params"`
}
