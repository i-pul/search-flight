package garuda

type garudaResponse struct {
	Status  string         `json:"status"`
	Flights []garudaFlight `json:"flights"`
}

type garudaFlight struct {
	FlightID    string          `json:"flight_id"`
	Airline     string          `json:"airline"`
	AirlineCode string          `json:"airline_code"`
	Departure   garudaPoint     `json:"departure"`
	Arrival     garudaPoint     `json:"arrival"`
	Duration    int             `json:"duration_minutes"`
	Stops       int             `json:"stops"`
	Aircraft    string          `json:"aircraft"`
	Price       garudaPrice     `json:"price"`
	Seats       int             `json:"available_seats"`
	FareClass   string          `json:"fare_class"`
	Baggage     garudaBaggage   `json:"baggage"`
	Amenities   []string        `json:"amenities"`
	Segments    []garudaSegment `json:"segments,omitempty"`
}

type garudaPoint struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal"`
}

type garudaPrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type garudaBaggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}

type garudaSegment struct {
	FlightNumber    string      `json:"flight_number"`
	Departure       garudaPoint `json:"departure"`
	Arrival         garudaPoint `json:"arrival"`
	DurationMinutes int         `json:"duration_minutes"`
	LayoverMinutes  int         `json:"layover_minutes,omitempty"`
}
