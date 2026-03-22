package batikair

type batikResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Results []batikFlight `json:"results"`
}

type batikFlight struct {
	FlightNumber      string      `json:"flightNumber"`
	AirlineName       string      `json:"airlineName"`
	AirlineIATA       string      `json:"airlineIATA"`
	Origin            string      `json:"origin"`
	Destination       string      `json:"destination"`
	DepartureDateTime string      `json:"departureDateTime"`
	ArrivalDateTime   string      `json:"arrivalDateTime"`
	TravelTime        string      `json:"travelTime"`
	NumberOfStops     int         `json:"numberOfStops"`
	Connections       []batikStop `json:"connections,omitempty"`
	Fare              batikFare   `json:"fare"`
	SeatsAvailable    int         `json:"seatsAvailable"`
	AircraftModel     string      `json:"aircraftModel"`
	BaggageInfo       string      `json:"baggageInfo"`
	OnboardServices   []string    `json:"onboardServices"`
}

type batikStop struct {
	StopAirport  string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"`
}

type batikFare struct {
	BasePrice    float64 `json:"basePrice"`
	Taxes        float64 `json:"taxes"`
	TotalPrice   float64 `json:"totalPrice"`
	CurrencyCode string  `json:"currencyCode"`
	Class        string  `json:"class"`
}
