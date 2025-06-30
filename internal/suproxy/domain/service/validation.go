//This function checks the validity of request ans sends the itemsString to OpenAI API if the HTTP status code is 200 for furhter processing of infromation.
//It requieres unmarshalled json Data as input.
//The information that requires further processing is contained in the items field of the Data struct.

package service

import (
	"fmt"
	"strings"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

var SystemPromptValidation string = `Your role: data tester of webpage requests.The webpage is a travel portal. The search requests include the client's situation and wishes for their next destination. Response includes all current offers that satisfy the conditions of the requests.

Analyze a pair of request/response JSON files that contain http protocol data

What status code had been returned, is it a failed or succesful request?

For failed request: is any important information missing? If yes find a possible placeholder. Is there a wrong information such as IDs, etc?

For successful requests: Validate responses/requests in terms of input/output matching. Don't repeat the contents of the request to keep the answer short and precise. If any information is not matching: write it out in a short form (1-2 words) for tagging

Provide the answer in JSON format according to the structure below. Don't include any additional Text.

{
  "status": {
	"http_status_code": null,
	"status_message": "",
	"is_failure": null
  },
  "validation": {
	"validation_message": "Request and response information matches/doesn't match",
	"offers_returned": null
  },
  "issues": {
	"issue_tag": "",
	"offerid": null,
	"issue description": null,
	"possible_placeholdes": null
  }
}
`

type Response struct {
	HTTPStatusCode int  `json:"httpstatuscode"`
	Data           Data `json:"data"`
}

type Data struct {
	Items []Item `json:"items"`
}

type Item struct {
	Duration          int               `json:"duration"`
	DepartureDate     string            `json:"departuredate"`
	ReturnDate        string            `json:"returndate"`
	OvernightDuration OvernightDuration `json:"overnightduration"`

	Rooms              []Room        `json:"rooms"`
	Accommodation      Accommodation `json:"accommodation"`
	Flight             Flight        `json:"flight"`
	AlternativeFlights []interface{} `json:"alternativeflights"`
	Transfer           bool          `json:"transfer"`
	CarRental          bool          `json:"carrental"`
	PrivateTransfer    bool          `json:"privatetransfer"`
	RailAndFly         bool          `json:"railandfly"`
	CancellationFree   bool          `json:"cancellation_free"`
	RebookingFree      bool          `json:"rebooking_free"`

	Travelers []TravelerWithPrice `json:"travelers"` //

	IsDirect bool `json:"isdirect"` //
}

type OvernightDuration struct {
	CheckInHotel  string `json:"checkinhotel"`
	CheckOutHotel string `json:"checkouthotel"`
	NightsInHotel int    `json:"nightsinhotel"`
}

type Room struct {
	ID                 int      `json:"id"`
	MealType           string   `json:"mealtype"`
	OceanView          bool     `json:"oceanview"`
	PartialOceanView   bool     `json:"partialoceanview"`
	Price              Price    `json:"price"`
	RoomDescription    string   `json:"roomdescription"`
	RoomType           string   `json:"roomtype"`
	TravelerAssignment []string `json:"travelerassignment"`
	RoomFacilities     []string `json:"roomfacilities"`
	RoomTypes          []string `json:"roomtypes"`
}

type Accommodation struct {
	CheckInDate       string     `json:"checkindate"`
	CheckOutDate      string     `json:"checkoutdate"`
	HotelBookingCode  string     `json:"hotelbookingcode"`
	ObjectID          string     `json:"objectid"`
	OfferHash         string     `json:"offerhash"`
	OptionPossible    bool       `json:"optionpossible"`
	OTDSImportVersion string     `json:"otdsimportversion"`
	Rooms             []AccRoom  `json:"rooms"`
	SeasonOverlap     bool       `json:"seasonoverlap"`
	Travelers         []Traveler `json:"travelers"`
}

type AccRoom struct {
	ID                 int      `json:"id"`
	MealBookingCode    string   `json:"mealbookingcode"`
	OceanView          bool     `json:"oceanview"`
	OfferHash          string   `json:"offerhash"`
	PartialOceanView   bool     `json:"partialoceanview"`
	PriceCacheID       float64  `json:"pricecacheid"`
	RoomBookingCode    string   `json:"roombookingcode"`
	TravelerAssignment []string `json:"travelerassignment"`
}

type Traveler struct {
	Age  int    `json:"age"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

type TravelerWithPrice struct {
	Age   int    `json:"age"`
	ID    string `json:"id"`
	Type  string `json:"type"`
	Price Price  `json:"price"`
}

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Flight struct {
	Price         Price `json:"price"`
	OutboundPrice Price `json:"outboundprice"`
	InboundPrice  Price `json:"inboundprice"`
}

// itemsString convertiert Data im String für Weiterleitung.
func itemsString(items []Item) string {
	var result string
	for _, item := range items {
		result += fmt.Sprintf(
			"Duration: %d | DepartureDate: %s | ReturnDate: %s | OvernightDuration: %+v | Rooms: %+v | Accommodation: %+v | Flight: %+v | AlternativeFlights: %+v | Transfer: %t | CarRental: %t | PrivateTransfer: %t | RailAndFly: %t | CancellationFree: %t | RebookingFree: %t | Travelers: %+v | IsDirect: %t\n",
			item.Duration,
			item.DepartureDate,
			item.ReturnDate,
			item.OvernightDuration,
			item.Rooms,
			item.Accommodation,
			item.Flight,
			item.AlternativeFlights,
			item.Transfer,
			item.CarRental,
			item.PrivateTransfer,
			item.RailAndFly,
			item.CancellationFree,
			item.RebookingFree,
			item.Travelers,
			item.IsDirect,
		)
	}
	result = strings.TrimSpace(result)
	return result
}

func printItems(data *Data) {
	if data == nil || len(data.Items) == 0 {
		fmt.Println("No items found.")
		return
	}
	fmt.Println("Items:")

	for i := 0; i < 1; i++ {
		item := data.Items[i]
		fmt.Printf(
			"Duration: %d | DepartureDate: %s | ReturnDate: %s | OvernightDuration: %+v | Rooms: %+v | Accommodation: %+v | Flight: %+v | AlternativeFlights: %+v | Transfer: %t | CarRental: %t | PrivateTransfer: %t | RailAndFly: %t | CancellationFree: %t | RebookingFree: %t | Travelers: %+v | IsDirect: %t\n",
			item.Duration,
			item.DepartureDate,
			item.ReturnDate,
			item.OvernightDuration,
			item.Rooms,
			item.Accommodation,
			item.Flight,
			item.AlternativeFlights,
			item.Transfer,
			item.CarRental,
			item.PrivateTransfer,
			item.RailAndFly,
			item.CancellationFree,
			item.RebookingFree,
			item.Travelers,
			item.IsDirect,
		)
	}
}

// function validiert die Response und sendet die itemsString an OpenAI API, falls die HTTP-Statuscode 200 ist
func validate(resp *Response) {
	if resp == nil {
		fmt.Println("No response to validate.")
		return
	}

	if resp.HTTPStatusCode != 200 {
		fmt.Printf("Invalid HTTP code: %d\n", resp.HTTPStatusCode)
		return
	} else {
		fmt.Printf("Valid HTTP status code: %d\n", resp.HTTPStatusCode)
		//send itemsToString to function connecting to OpenAI API
		repository.ConnectAndRequestOpenAI(
			itemsString(resp.Data.Items),
			SystemPromptValidation,
			"gpt-3.5-turbo",
		)
	}

}
