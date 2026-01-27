package entity

// RequestBody represents the parsed JSON body of the Request.
type RequestBody struct {
	DepartureAirportList []string            `json:"departureairportlist"`
	DepartureDate        string              `json:"departuredate"`
	ReturnDate           string              `json:"returndate"`
	Travelers            map[string]Traveler `json:"travelers"`
	TravelType           string              `json:"traveltype"`
}

// Traveler represents the number of adults and children travelers.
type Traveler struct {
	Adults   int `json:"adults"`
	Children int `json:"children"`
}
