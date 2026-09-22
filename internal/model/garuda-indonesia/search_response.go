package garudaindonesiamodel

type SearchResponse struct {
	Status  string   `json:"status"`
	Flights []Flight `json:"flights"`
}

type Flight struct {
	FlightID        string    `json:"flight_id"`
	Airline         string    `json:"airline"`
	AirlineCode     string    `json:"airline_code"`
	Departure       Airport   `json:"departure"`
	Arrival         Airport   `json:"arrival"`
	DurationMinutes int       `json:"duration_minutes"`
	Stops           int       `json:"stops"`
	Aircraft        string    `json:"aircraft"`
	Price           Price     `json:"price"`
	Segments        []Segment `json:"segments,omitempty"`
	AvailableSeats  int       `json:"available_seats"`
	FareClass       string    `json:"fare_class"`
	Baggage         Baggage   `json:"baggage"`
	Amenities       []string  `json:"amenities,omitempty"`
}

type Airport struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal"`
}

type Price struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type Baggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}

type Segment struct {
	FlightNumber    string         `json:"flight_number"`
	Departure       SegmentAirport `json:"departure"`
	Arrival         SegmentAirport `json:"arrival"`
	DurationMinutes int            `json:"duration_minutes"`
	LayoverMinutes  int            `json:"layover_minutes,omitempty"`
}

type SegmentAirport struct {
	Airport string `json:"airport"`
	Time    string `json:"time"`
}
