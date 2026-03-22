package lionair

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

func (r *Repository) Name() string { return "LionAir" }

func (r *Repository) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	// Simulates a 100-200ms response time
	deadline := time.Now().Add(time.Duration(100+rand.Intn(101)) * time.Millisecond)

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
	var raw lionResponse
	if err := json.Unmarshal(mockdata.LionAir, &raw); err != nil {
		return nil, fmt.Errorf("lionair: unmarshal: %w", err)
	}

	flights := make([]domain.Flight, 0, len(raw.Data.AvailableFlights))
	for _, f := range raw.Data.AvailableFlights {
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

func adapt(f lionFlight) (domain.Flight, error) {
	deptTime, err := util.ParseWithIANA(f.Schedule.Departure, f.Schedule.DepartureTimezone)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("lionair %s: parse departure: %w", f.ID, err)
	}
	arrTime, err := util.ParseWithIANA(f.Schedule.Arrival, f.Schedule.ArrivalTimezone)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("lionair %s: parse arrival: %w", f.ID, err)
	}
	if !arrTime.After(deptTime) {
		return domain.Flight{}, fmt.Errorf("lionair %s: arrival not after departure", f.ID)
	}

	stops := f.StopCount
	if f.IsDirect {
		stops = 0
	}

	// Derive amenities from service booleans
	amenities := make([]string, 0)
	if f.Services.WiFiAvailable {
		amenities = append(amenities, "wifi")
	}
	if f.Services.MealsIncluded {
		amenities = append(amenities, "meal")
	}

	planeType := f.PlaneType
	baggage := domain.Baggage{
		CarryOn: f.Services.BaggageAllowance.Cabin,
		Checked: f.Services.BaggageAllowance.Hold,
	}

	return domain.Flight{
		ID:       f.ID + "_LionAir",
		Provider: "LionAir",
		Airline: domain.Airline{
			Name: f.Carrier.Name,
			Code: f.Carrier.IATA,
		},
		FlightNumber: f.ID,
		Departure: domain.FlightPoint{
			Airport:   f.Route.From.Code,
			City:      f.Route.From.City,
			Datetime:  deptTime.Format(time.RFC3339),
			Timestamp: deptTime.Unix(),
		},
		Arrival: domain.FlightPoint{
			Airport:   f.Route.To.Code,
			City:      f.Route.To.City,
			Datetime:  arrTime.Format(time.RFC3339),
			Timestamp: arrTime.Unix(),
		},
		Duration: domain.Duration{
			TotalMinutes: f.FlightTime,
			Formatted:    util.FormatDuration(f.FlightTime),
		},
		Stops: stops,
		Price: domain.Price{
			Amount:   f.Pricing.Total,
			Currency: f.Pricing.Currency,
		},
		AvailableSeats: f.SeatsLeft,
		CabinClass:     strings.ToLower(f.Pricing.FareType),
		Aircraft:       &planeType,
		Amenities:      amenities,
		Baggage:        baggage,
	}, nil
}
