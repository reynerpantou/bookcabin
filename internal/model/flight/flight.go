package flight

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
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Formatted string `json:"formatted"`
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}
