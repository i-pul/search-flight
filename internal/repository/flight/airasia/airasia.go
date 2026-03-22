package airasia

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/i-pul/search-flight/internal/domain"
	"github.com/i-pul/search-flight/internal/mockdata"
	"github.com/i-pul/search-flight/internal/util"
)

type Repository struct{}

func New() *Repository { return &Repository{} }

func (r *Repository) Name() string { return "AirAsia" }

func (r *Repository) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	// Simulates a 50-150ms response time
	deadline := time.Now().Add(time.Duration(50+rand.Intn(101)) * time.Millisecond)

	flights, err := r.search(req)

	if remaining := time.Until(deadline); remaining > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(remaining):
		}
	}

	return flights, err
}

func (r *Repository) search(req domain.SearchRequest) ([]domain.Flight, error) {
	// Simulate 90% success rate
	if rand.Float64() < 0.1 {
		return nil, fmt.Errorf("airasia: provider temporarily unavailable")
	}

	var raw airasiResponse
	if err := json.Unmarshal(mockdata.AirAsia, &raw); err != nil {
		return nil, fmt.Errorf("airasia: unmarshal: %w", err)
	}

	flights := make([]domain.Flight, 0, len(raw.Flights))
	for _, f := range raw.Flights {
		normalized, err := adapt(f)
		if err != nil {
			continue
		}
		if !req.Matches(normalized) {
			continue
		}
		flights = append(flights, normalized)
	}
	return flights, nil
}

func adapt(f airasiFlight) (domain.Flight, error) {
	deptTime, err := util.ParseRFC3339(f.DepartTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("airasia %s: parse departure: %w", f.FlightCode, err)
	}
	arrTime, err := util.ParseRFC3339(f.ArriveTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("airasia %s: parse arrival: %w", f.FlightCode, err)
	}
	if !arrTime.After(deptTime) {
		return domain.Flight{}, fmt.Errorf("airasia %s: arrival not after departure", f.FlightCode)
	}

	durationMinutes := int(math.Round(f.DurationHrs * 60))

	stops := 0
	if !f.DirFlight {
		stops = len(f.Stops)
	}

	return domain.Flight{
		ID:       f.FlightCode + "_AirAsia",
		Provider: "AirAsia",
		Airline: domain.Airline{
			Name: f.Airline,
			Code: strings.TrimRight(f.FlightCode, "0123456789"),
		},
		FlightNumber: f.FlightCode,
		Departure: domain.FlightPoint{
			Airport:   f.FromAirport,
			City:      airportCity(f.FromAirport),
			Datetime:  deptTime.Format(time.RFC3339),
			Timestamp: deptTime.Unix(),
			Timezone:  util.TimezoneAbbr(deptTime),
		},
		Arrival: domain.FlightPoint{
			Airport:   f.ToAirport,
			City:      airportCity(f.ToAirport),
			Datetime:  arrTime.Format(time.RFC3339),
			Timestamp: arrTime.Unix(),
			Timezone:  util.TimezoneAbbr(arrTime),
		},
		Duration: domain.Duration{
			TotalMinutes: durationMinutes,
			Formatted:    util.FormatDuration(durationMinutes),
		},
		Stops: stops,
		Price: domain.Price{
			Amount:    f.PriceIDR,
			Currency:  "IDR",
			Formatted: util.FormatPrice(f.PriceIDR, "IDR"),
		},
		AvailableSeats: f.Seats,
		CabinClass:     f.CabinClass,
		Aircraft:       nil,
		Amenities:      []string{},
		Baggage:        parseAirasiBaggage(f.BaggageNote),
	}, nil
}

func parseAirasiBaggage(note string) domain.Baggage {
	lower := strings.ToLower(note)
	if strings.Contains(lower, "checked bags") {
		return domain.Baggage{
			CarryOn: "Cabin baggage only",
			Checked: "Additional fee",
		}
	}
	return domain.Baggage{
		CarryOn: "Cabin baggage only",
		Checked: note,
	}
}

func airportCity(iata string) string {
	cities := map[string]string{
		"CGK": "Jakarta",
		"DPS": "Denpasar",
		"SUB": "Surabaya",
		"UPG": "Makassar",
		"SOC": "Surakarta",
		"JOG": "Yogyakarta",
	}
	if city, ok := cities[iata]; ok {
		return city
	}
	return iata
}
