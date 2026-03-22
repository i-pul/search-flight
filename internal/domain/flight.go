package domain

import "strings"

type Flight struct {
	ID             string      `json:"id"`
	Provider       string      `json:"provider"`
	Airline        Airline     `json:"airline"`
	FlightNumber   string      `json:"flight_number"`
	Departure      FlightPoint `json:"departure"`
	Arrival        FlightPoint `json:"arrival"`
	Duration       Duration    `json:"duration"`
	Stops          int         `json:"stops"`
	Price          Price       `json:"price"`
	AvailableSeats int         `json:"available_seats"`
	CabinClass     string      `json:"cabin_class"`
	Aircraft       *string     `json:"aircraft"`
	Amenities      []string    `json:"amenities"`
	Baggage        Baggage     `json:"baggage"`
	BestValueScore float64     `json:"best_value_score"` // 0–100, higher = better value
}

type Airline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type FlightPoint struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
	Timezone  string `json:"timezone"` // e.g. "WIB", "WITA", "WIT"
}

type Duration struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"` // e.g. "4h 20m"
}

type Price struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Formatted string  `json:"formatted"` // e.g. "Rp 1.500.000"
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}

type SearchRequest struct {
	Origin        string  `json:"origin"`
	Destination   string  `json:"destination"`
	DepartureDate string  `json:"departureDate"`
	ReturnDate    *string `json:"returnDate"`
	Passengers    int     `json:"passengers"`
	CabinClass    string  `json:"cabinClass"`
}

func (req SearchRequest) Matches(f Flight) bool {
	if req.Origin != "" && !strings.EqualFold(f.Departure.Airport, req.Origin) {
		return false
	}
	if req.Destination != "" && !strings.EqualFold(f.Arrival.Airport, req.Destination) {
		return false
	}
	if req.DepartureDate != "" && len(f.Departure.Datetime) >= 10 {
		if f.Departure.Datetime[:10] != req.DepartureDate {
			return false
		}
	}
	if req.CabinClass != "" && !strings.EqualFold(f.CabinClass, req.CabinClass) {
		return false
	}
	if req.Passengers > 0 && f.AvailableSeats < req.Passengers {
		return false
	}
	return true
}

type SearchCriteria struct {
	Origin        string  `json:"origin"`
	Destination   string  `json:"destination"`
	DepartureDate string  `json:"departure_date"`
	ReturnDate    *string `json:"return_date,omitempty"`
	Passengers    int     `json:"passengers"`
	CabinClass    string  `json:"cabin_class"`
}

type SearchMetadata struct {
	TotalResults       int   `json:"total_results"`
	ReturnResults      int   `json:"return_results,omitempty"` // return leg count; omitted for one-way
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

type SearchResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata       SearchMetadata `json:"metadata"`
	Flights        []Flight       `json:"flights"`
	ReturnFlights  []Flight       `json:"return_flights,omitempty"` // populated for round-trip
}
