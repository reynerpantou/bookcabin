package response

type FlightSearchResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata       Metadata       `json:"metadata"`
	Flights        []Flight       `json:"flights"`
}

type SearchCriteria struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMS       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

type Flight struct {
	ID             string   `json:"id"`
	Provider       string   `json:"provider"`
	Airline        Airline  `json:"airline"`
	FlightNumber   string   `json:"flight_number"`
	Departure      Airport  `json:"departure"`
	Arrival        Airport  `json:"arrival"`
	Duration       Duration `json:"duration"`
	Stops          int      `json:"stops"`
	Price          Price    `json:"price"`
	AvailableSeats int      `json:"available_seats"`
	CabinClass     string   `json:"cabin_class"`
	Aircraft       *string  `json:"aircraft"`
	Amenities      []string `json:"amenities"`
	Baggage        Baggage  `json:"baggage"`
}

type Airline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Airport struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	DateTime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

type Duration struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

type Price struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}
