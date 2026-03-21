package garuda

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/i-pul/search-flight/internal/domain"
	"github.com/i-pul/search-flight/internal/mockdata"
	"github.com/i-pul/search-flight/internal/util"
)

type Repository struct{}

func New() *Repository { return &Repository{} }

func (r *Repository) Name() string { return "GarudaIndonesia" }

func (r *Repository) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	// Simulates a 50-100ms response time
	deadline := time.Now().Add(time.Duration(50+rand.Intn(51)) * time.Millisecond)

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
	var raw garudaResponse
	if err := json.Unmarshal(mockdata.Garuda, &raw); err != nil {
		return nil, fmt.Errorf("garuda: unmarshal: %w", err)
	}

	flights := make([]domain.Flight, 0, len(raw.Flights))
	for _, f := range raw.Flights {
		normalized, err := adapt(f)
		if err != nil {
			// skip invalid flight data
			continue
		}
		if !req.Matches(normalized) {
			continue
		}
		flights = append(flights, normalized)
	}
	return flights, nil
}

func adapt(f garudaFlight) (domain.Flight, error) {
	deptTime, err := util.ParseRFC3339(f.Departure.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("garuda %s: parse departure: %w", f.FlightID, err)
	}
	arrTime, err := util.ParseRFC3339(f.Arrival.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("garuda %s: parse arrival: %w", f.FlightID, err)
	}
	if !arrTime.After(deptTime) {
		return domain.Flight{}, fmt.Errorf("garuda %s: arrival not after departure", f.FlightID)
	}

	stops := f.Stops
	if len(f.Segments) > 1 {
		stops = len(f.Segments) - 1
	}

	aircraft := f.Aircraft
	baggage := domain.Baggage{
		CarryOn: fmt.Sprintf("%d piece(s)", f.Baggage.CarryOn),
		Checked: fmt.Sprintf("%d piece(s)", f.Baggage.Checked),
	}

	return domain.Flight{
		ID:       f.FlightID + "_GarudaIndonesia",
		Provider: "GarudaIndonesia",
		Airline: domain.Airline{
			Name: f.Airline,
			Code: f.AirlineCode,
		},
		FlightNumber: f.FlightID,
		Departure: domain.FlightPoint{
			Airport:   f.Departure.Airport,
			City:      f.Departure.City,
			Datetime:  deptTime.Format(time.RFC3339),
			Timestamp: deptTime.Unix(),
		},
		Arrival: domain.FlightPoint{
			Airport:   f.Arrival.Airport,
			City:      f.Arrival.City,
			Datetime:  arrTime.Format(time.RFC3339),
			Timestamp: arrTime.Unix(),
		},
		Duration: domain.Duration{
			TotalMinutes: f.Duration,
			Formatted:    util.FormatDuration(f.Duration),
		},
		Stops:          stops,
		Price:          domain.Price{Amount: f.Price.Amount, Currency: f.Price.Currency},
		AvailableSeats: f.Seats,
		CabinClass:     f.FareClass,
		Aircraft:       &aircraft,
		Amenities:      f.Amenities,
		Baggage:        baggage,
	}, nil
}
