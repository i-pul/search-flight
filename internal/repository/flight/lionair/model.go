package lionair

type lionResponse struct {
	Success bool     `json:"success"`
	Data    lionData `json:"data"`
}

type lionData struct {
	AvailableFlights []lionFlight `json:"available_flights"`
}

type lionFlight struct {
	ID         string        `json:"id"`
	Carrier    lionCarrier   `json:"carrier"`
	Route      lionRoute     `json:"route"`
	Schedule   lionSchedule  `json:"schedule"`
	FlightTime int           `json:"flight_time"`
	IsDirect   bool          `json:"is_direct"`
	StopCount  int           `json:"stop_count"`
	Layovers   []lionLayover `json:"layovers,omitempty"`
	Pricing    lionPricing   `json:"pricing"`
	SeatsLeft  int           `json:"seats_left"`
	PlaneType  string        `json:"plane_type"`
	Services   lionServices  `json:"services"`
}

type lionCarrier struct {
	Name string `json:"name"`
	IATA string `json:"iata"`
}

type lionRoute struct {
	From lionAirport `json:"from"`
	To   lionAirport `json:"to"`
}

type lionAirport struct {
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

type lionSchedule struct {
	Departure         string `json:"departure"`
	DepartureTimezone string `json:"departure_timezone"`
	Arrival           string `json:"arrival"`
	ArrivalTimezone   string `json:"arrival_timezone"`
}

type lionLayover struct {
	Airport         string `json:"airport"`
	DurationMinutes int    `json:"duration_minutes"`
}

type lionPricing struct {
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
	FareType string  `json:"fare_type"`
}

type lionServices struct {
	WiFiAvailable    bool        `json:"wifi_available"`
	MealsIncluded    bool        `json:"meals_included"`
	BaggageAllowance lionBaggage `json:"baggage_allowance"`
}

type lionBaggage struct {
	Cabin string `json:"cabin"`
	Hold  string `json:"hold"`
}
