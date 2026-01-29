package entity

import (
	"encoding/json"
	"fmt"
)

// RequestBody represents the parsed JSON body of the Request.
type RequestBody struct {
	DepartureAirportList []string  `json:"departureairportlist"`
	DepartureDate        string    `json:"departuredate"`
	ReturnDate           string    `json:"returndate"`
	Travelers            Travelers `json:"travelers"`
	TravelType           string    `json:"traveltype"`
}

// Traveler represents a single travelling person.
type Traveler struct {
	ID                string        `json:"id"`
	Type              string        `json:"type"` // adult, child, infant, ...
	Age               int           `json:"age"`
	PersonBookingCode string        `json:"personbookingcode"`
	Price             TravelerPrice `json:"price"`
}

// TravelerPrice represents the price information associated with a traveler.
type TravelerPrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Travelers represents a flexible list of Traveler objects that can be
// unmarshaled from multiple JSON formats (array, object, or map).
type Travelers []Traveler

// UnmarshalJSON implements custom unmarshaling for Travelers to support
// inconsistent JSON structures.
func (t *Travelers) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*t = nil
		return nil
	}

	// Case 1: array
	var arr []Traveler
	if err := json.Unmarshal(data, &arr); err == nil {
		*t = arr
		return nil
	}

	// Case 2: single object
	var single Traveler
	if err := json.Unmarshal(data, &single); err == nil {
		*t = []Traveler{single}
		return nil
	}

	// Case 3: object with numeric keys (map)
	var m map[string]Traveler
	if err := json.Unmarshal(data, &m); err == nil {
		out := make([]Traveler, 0, len(m))
		for _, v := range m {
			out = append(out, v)
		}
		*t = out
		return nil
	}

	return fmt.Errorf("unsupported travelers format: %s", string(data))
}
