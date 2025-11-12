package entity

// TagList represents a list of tags with metadata.
type TagList struct {
	Tags []Tag `json:"tags"`
}

// Tag represents a Tag with Name and Description
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Format changes tags formats into a string
func (tl *TagList) Format() string {
	if tl == nil || len(tl.Tags) == 0 {
		tl = DefaultTagList()
	}

	formatted := ""
	for _, tag := range tl.Tags {
		formatted += tag.Name + " - " + tag.Description + "\n"
	}
	return formatted
}

// DefaultTagList provides a default tag list
func DefaultTagList() *TagList {
	return &TagList{
		Tags: []Tag{
			{Name: "no_hotelid", Description: "The response does not contain a hotelid field."},
			{Name: "no_outbound_flight", Description: "The response lacks outbound flight information (flight.outboundflightsegments is missing or empty)."},
			{Name: "no_return_flight", Description: "The response lacks return flight information (flight.inboundflightsegments is missing or empty)."},
			{Name: "no_price_info", Description: "Price information is missing (price.amount or equivalent fields are missing or empty)."},
			{Name: "no_mealtype", Description: "Meal type information is missing, such as rooms.mealtype or accommodation.rooms.mealbookingcode."},
			{Name: "no_duration_info", Description: "Duration is missing (duration or overnightduration fields are not present)."},
			{Name: "no_travelers", Description: "No traveler information is present (travelers is missing or empty)."},
			{Name: "no_offerid", Description: "No unique offer ID (offerid) is present."},
			{Name: "invalid_dates", Description: "Date information is implausible or inconsistent (e.g., return date before departure)."},
			{Name: "no_paymenttypes", Description: "Available payment methods are missing (paymenttypes is missing or empty)."},
			{Name: "no_cancellation_info", Description: "Cancellation conditions are completely missing (services.cancellationrebookingfee is not present)."},
			{Name: "no_emission_info", Description: "Emissions data (emissions.*) is completely missing."},
			{Name: "no_room_description", Description: "Room descriptions or equipment information are missing (e.g., rooms.roomdescription, rooms.roomfacilities)."},
			{Name: "no_transfer_info", Description: "Transfer information is missing (e.g., transfer, privatetransfer)."},
			{Name: "no_departuredate", Description: "The field departuredate is missing or empty."},
			{Name: "no_returndate", Description: "The field returndate is missing or empty."},
			{Name: "missing_flight_price", Description: "Flight price information is missing (e.g., flight.price.amount or flight.outboundprice.amount)."},
			{Name: "missing_room_price", Description: "Room price information is missing (e.g., rooms.price.amount)."},
			{Name: "missing_roomtype", Description: "The room type (rooms.roomtype) is not specified."},
			{Name: "no_offerhash", Description: "The field accommodation.rooms.offerhash is missing, although it is needed for offer identification."},
			{Name: "missing_touroperatorcode", Description: "The touroperatorcode field is missing."},
			{Name: "invalid_room_assignment", Description: "No traveler-to-room assignment is given (rooms.travelerassignment is missing or empty)."},
			{Name: "no_availability_info", Description: "Availability object is missing or contains no relevant information (e.g., availability.flight, availability.unit)."},
			{Name: "missing_enginecodes", Description: "One or more fields like accommodation.rooms.enginecodes.roomid or mealid are missing."},
			{Name: "missing_duration_component", Description: "Duration subfields like overnightduration.nightsinhotel or overnightduration.checkinhotel are missing."},
			{Name: "no_mealbookingcode", Description: "The field accommodation.rooms.mealbookingcode is missing."},
			{Name: "no_checkindate_or_checkoutdate", Description: "accommodation.checkindate or accommodation.checkoutdate is missing."},
			{Name: "missing_airport_codes", Description: "One or more fields such as departureairport.code or destinationairport.code are missing."},
			{Name: "no_personbookingcode", Description: "The field travelers.personbookingcode is missing – this is often required for traveler identification or pricing."},
			{Name: "no_bag_information", Description: "Baggage-related information is missing, such as included luggage, hand luggage allowances, or relevant fields in flight.baggage or services.baggage."},
		},
	}
}
