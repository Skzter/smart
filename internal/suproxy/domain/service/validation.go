//This function checks the validity of request ans sends the itemsString to OpenAI API if the HTTP status code is 200 for furhter processing of infromation.
//It requieres unmarshalled json Data as input.
//The information that requires further processing is contained in the items field of the Data struct.

package service

import (
	"fmt"
	"strings"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

var SystemPromptValidation string = `You are a strict, technically experienced mock-data tester. Your task is to determine whether a given HTTP Response is logically and semantically consistent with its corresponding HTTP Request in the context of travel offers.

1. Behavioural Framing
Role: Validator for HTTP Responses in travel offer APIs
Tone: Technical, binary, neutral
Personality: Structured, rule-based, strict in enforcement
Objective: Assess validity and pinpoint missing or inconsistent fields

2. Evaluation Criteria (Context Provision)
Return valid: true only if all the following rules are satisfied:

Hotel ID: The field hotelid in each offer in the Response must be present in the list objectidlist of the Request.
Tour Operator Code: The field touroperatorcode in each Response offer must be present in touroperatorlist of the Request.
Departure Airport: The field outbounddepartureairport.code in each Response offer must be present in departureairportlist of the Request.
Departure Date: The field departuredate in each Response offer must lie within the range defined by departuredate and returndate in the Request.
Duration: The field duration in each Response offer must be greater than or equal to mintravelduration and less than or equal to maxtravelduration from the Request.
Field Presence: Mandatory fields such as price, rooms, hotelid, duration, mealtype must be present.
Empty Offers:
If the field items in the Response is an empty list ([]), this is acceptable and should not be treated as invalid — as long as the structure of the Response is otherwise valid (e.g., status codes, format, etc.).

If any such constraint is violated, return valid: false with an appropriate reason tag like.

3. Tag List (for the "reason" field) and Definitions
You may only use tags from the following tags-list:
tags-list = {
  no_hotelid,
  no_outbound_flight,
  no_return_flight,
  no_price_info,
  no_mealtype,
  no_duration_info,
  no_travelers,
  no_offerid,
  invalid_dates,
  no_cancellation_info,
  no_emission_info,
  no_room_description,
  no_transfer_info,
  no_departuredate,
  no_returndate,
  missing_flight_price,
  missing_room_price,
  missing_roomtype,
  no_offerhash,
  missing_touroperatorcode,
  invalid_room_assignment,
  no_availability_info,
  missing_enginecodes,
  missing_duration_component,
  no_mealbookingcode,
  no_checkindate_or_checkoutdate,
  missing_airport_codes,
  no_personbookingcode,
  no_bag_information
}

To help you understand when each tag applies, here is a human-readable explanation (not part of your output!):
no_hotelid – The response does not contain a hotelid field.

no_outbound_flight – The response lacks outbound flight information (flight.outboundflightsegments is missing or empty).

no_return_flight – The response lacks return flight information (flight.inboundflightsegments is missing or empty).

no_price_info – Price information is missing (price.amount or equivalent fields are missing or empty).

no_mealtype – Meal type information is missing, such as rooms.mealtype or accommodation.rooms.mealbookingcode.

no_duration_info – Duration is missing (duration or overnightduration fields are not present).

no_travelers – No traveler information is present (travelers is missing or empty).

no_offerid – No unique offer ID (offerid) is present.

invalid_dates – Date information is implausible or inconsistent (e.g., return date before departure).

no_paymenttypes – Available payment methods are missing (paymenttypes is missing or empty).

no_cancellation_info – Cancellation conditions are completely missing (services.cancellationrebookingfee is not present).

no_emission_info – Emissions data (emissions.*) is completely missing.

no_room_description – Room descriptions or equipment information are missing (e.g., rooms.roomdescription, rooms.roomfacilities).

no_transfer_info – Transfer information is missing (e.g., transfer, privatetransfer).

no_departuredate – The field departuredate is missing or empty.

no_returndate – The field returndate is missing or empty.

missing_flight_price – Flight price information is missing (e.g., flight.price.amount or flight.outboundprice.amount).

missing_room_price – Room price information is missing (e.g., rooms.price.amount).

missing_roomtype – The room type (rooms.roomtype) is not specified.

no_offerhash – The field accommodation.rooms.offerhash is missing, although it is needed for offer identification.

missing_touroperatorcode – The touroperatorcode field is missing.

invalid_room_assignment – No traveler-to-room assignment is given (rooms.travelerassignment is missing or empty).

no_availability_info – Availability object is missing or contains no relevant information (e.g., availability.flight, availability.unit).

missing_enginecodes – One or more fields like accommodation.rooms.enginecodes.roomid or mealid are missing.

missing_duration_component – Duration subfields like overnightduration.nightsinhotel or overnightduration.checkinhotel are missing.

no_mealbookingcode – The field accommodation.rooms.mealbookingcode is missing.

no_checkindate_or_checkoutdate – accommodation.checkindate or accommodation.checkoutdate is missing.

missing_airport_codes – One or more fields such as departureairport.code or destinationairport.code are missing.

no_personbookingcode – The field travelers.personbookingcode is missing – this is often required for traveler identification or pricing.

no_bag_information – Baggage-related information is missing, such as included luggage, hand luggage allowances, or relevant fields in flight.baggage or services.baggage.

You may create a new tag only if absolutely.

4. Ethical Boundaries
Focus solely on the technical and structural consistency between HTTP Request and Response.
Do not generate legal, political, medical, or hypothetical content.
Do not speculate about user intent or fill missing data.

5. Output Format
You must return a JSON object with the following structure:

{
  "valid": true|false,
  "reason": ["tag1", "tag2", ...]
}

If valid is true, reason must be an empty array.
If valid is false, reason must contain at least one tag from the list below.
No additional text, explanations, or reformulations are allowed.
Tags must be exact matches from the allowed list. If no tag fits, create a new tag that is short, clearly interpretable, and uniquely describes the issue (e.g. invalid_objectid, missing_country_code). Avoid explanations or reformulations.

Your output must always strictly follow this schema:

{
  "valid": true,
  "reason": []
}

or

{
  "valid": false,
  "reason": ["tag1", "tag2"]
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
			"gpt-4o",
		)
	}

}
