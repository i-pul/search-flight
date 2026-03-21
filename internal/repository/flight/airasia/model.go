package airasia

type airasiResponse struct {
	Status  string         `json:"status"`
	Flights []airasiFlight `json:"flights"`
}

type airasiFlight struct {
	FlightCode  string       `json:"flight_code"`
	Airline     string       `json:"airline"`
	FromAirport string       `json:"from_airport"`
	ToAirport   string       `json:"to_airport"`
	DepartTime  string       `json:"depart_time"`
	ArriveTime  string       `json:"arrive_time"`
	DurationHrs float64      `json:"duration_hours"`
	DirFlight   bool         `json:"direct_flight"`
	Stops       []airasiStop `json:"stops,omitempty"`
	PriceIDR    float64      `json:"price_idr"`
	Seats       int          `json:"seats"`
	CabinClass  string       `json:"cabin_class"`
	BaggageNote string       `json:"baggage_note"`
}

type airasiStop struct {
	Airport         string `json:"airport"`
	WaitTimeMinutes int    `json:"wait_time_minutes"`
}
