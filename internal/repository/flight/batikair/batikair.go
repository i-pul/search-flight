package batikair

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/i-pul/search-flight/internal/domain"
	"github.com/i-pul/search-flight/internal/mockdata"
	"github.com/i-pul/search-flight/internal/util"
)

type Repository struct{}

func New() *Repository { return &Repository{} }

func (r *Repository) Name() string { return "BatikAir" }

func (r *Repository) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	// Simulates a 200-400ms response time
	deadline := time.Now().Add(time.Duration(200+rand.Intn(201)) * time.Millisecond)

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
	var raw batikResponse
	if err := json.Unmarshal(mockdata.BatikAir, &raw); err != nil {
		return nil, fmt.Errorf("batikair: unmarshal: %w", err)
	}

	flights := make([]domain.Flight, 0, len(raw.Results))
	for _, f := range raw.Results {
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

func adapt(f batikFlight) (domain.Flight, error) {
	deptTime, err := util.ParseBatikAirTime(f.DepartureDateTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("batikair %s: parse departure: %w", f.FlightNumber, err)
	}
	arrTime, err := util.ParseBatikAirTime(f.ArrivalDateTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("batikair %s: parse arrival: %w", f.FlightNumber, err)
	}
	if !arrTime.After(deptTime) {
		return domain.Flight{}, fmt.Errorf("batikair %s: arrival not after departure", f.FlightNumber)
	}

	durationMinutes, err := util.ParseTravelTime(f.TravelTime)
	if err != nil {
		// Fall back to computing from timestamps
		durationMinutes = int(arrTime.Sub(deptTime).Minutes())
	}

	amenities := make([]string, 0, len(f.OnboardServices))
	for _, svc := range f.OnboardServices {
		amenities = append(amenities, strings.ToLower(svc))
	}

	aircraft := f.AircraftModel

	return domain.Flight{
		ID:       f.FlightNumber + "_BatikAir",
		Provider: "BatikAir",
		Airline: domain.Airline{
			Name: f.AirlineName,
			Code: f.AirlineIATA,
		},
		FlightNumber: f.FlightNumber,
		Departure: domain.FlightPoint{
			Airport:   f.Origin,
			City:      airportCity(f.Origin),
			Datetime:  deptTime.Format(time.RFC3339),
			Timestamp: deptTime.Unix(),
		},
		Arrival: domain.FlightPoint{
			Airport:   f.Destination,
			City:      airportCity(f.Destination),
			Datetime:  arrTime.Format(time.RFC3339),
			Timestamp: arrTime.Unix(),
		},
		Duration: domain.Duration{
			TotalMinutes: durationMinutes,
			Formatted:    util.FormatDuration(durationMinutes),
		},
		Stops: f.NumberOfStops,
		Price: domain.Price{
			Amount:   f.Fare.TotalPrice,
			Currency: f.Fare.CurrencyCode,
		},
		AvailableSeats: f.SeatsAvailable,
		CabinClass:     "economy", // class "Y" = economy
		Aircraft:       &aircraft,
		Amenities:      amenities,
		Baggage:        parseBaggageInfo(f.BaggageInfo),
	}, nil
}

func parseBaggageInfo(info string) domain.Baggage {
	parts := strings.SplitN(info, ",", 2)
	if len(parts) == 2 {
		return domain.Baggage{
			CarryOn: strings.TrimSpace(parts[0]),
			Checked: strings.TrimSpace(parts[1]),
		}
	}
	return domain.Baggage{CarryOn: info, Checked: ""}
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
