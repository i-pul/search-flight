package lionair

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/i-pul/search-flight/internal/domain"
)

var defaultReq = domain.SearchRequest{
	Origin:        "CGK",
	Destination:   "DPS",
	DepartureDate: "2025-12-15",
	Passengers:    1,
	CabinClass:    "economy",
}

func TestName(t *testing.T) {
	assert.Equal(t, "LionAir", New().Name())
}

func TestSearch(t *testing.T) {
	r := New()

	t.Run("returns flights", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)
		assert.NotEmpty(t, flights)
	})

	t.Run("JT740 field normalization", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		var jt740 *domain.Flight
		for i := range flights {
			if flights[i].FlightNumber == "JT740" {
				jt740 = &flights[i]
				break
			}
		}
		require.NotNil(t, jt740, "JT740 not found in results")

		tests := []struct {
			field string
			got   interface{}
			want  interface{}
		}{
			{"provider", jt740.Provider, "LionAir"},
			{"airline code", jt740.Airline.Code, "JT"},
			{"departure airport", jt740.Departure.Airport, "CGK"},
			{"arrival airport", jt740.Arrival.Airport, "DPS"},
			{"price amount", jt740.Price.Amount, float64(950000)},
			{"price currency", jt740.Price.Currency, "IDR"},
			{"stops", jt740.Stops, 0},
			{"duration minutes", jt740.Duration.TotalMinutes, 105},
			{"duration formatted", jt740.Duration.Formatted, "1h 45m"},
			{"available seats", jt740.AvailableSeats, 45},
		}
		for _, tc := range tests {
			assert.Equal(t, tc.want, tc.got, tc.field)
		}
		assert.NotZero(t, jt740.Departure.Timestamp, "departure timestamp")
		assert.Greater(t, jt740.Arrival.Timestamp, jt740.Departure.Timestamp, "arrival after departure")
	})

	t.Run("JT650 with stop", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		for _, f := range flights {
			if f.FlightNumber == "JT650" {
				assert.Equal(t, 1, f.Stops, "JT650 stops")
				return
			}
		}
		t.Error("JT650 not found in results")
	})

	t.Run("cabin class normalized to lowercase", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		for _, f := range flights {
			assert.Equal(t, "economy", f.CabinClass, "flight %s cabin class", f.FlightNumber)
		}
	})

	t.Run("no match returns empty slice", func(t *testing.T) {
		req := domain.SearchRequest{
			Origin:        "SUB",
			Destination:   "JOG",
			DepartureDate: "2025-12-15",
			Passengers:    1,
			CabinClass:    "economy",
		}
		flights, err := r.Search(context.Background(), req)
		require.NoError(t, err)
		assert.Empty(t, flights)
	})

	t.Run("context canceled returns error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond)

		_, err := r.Search(ctx, defaultReq)
		assert.Error(t, err)
	})
}

func TestAdapt_Errors(t *testing.T) {
	t.Run("bad departure time", func(t *testing.T) {
		f := lionFlight{
			ID: "JT999",
			Schedule: lionSchedule{
				Departure:         "not-a-time",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T10:00:00",
				ArrivalTimezone:   "Asia/Makassar",
			},
		}
		_, err := adapt(f)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse departure")
	})

	t.Run("bad arrival time", func(t *testing.T) {
		f := lionFlight{
			ID: "JT999",
			Schedule: lionSchedule{
				Departure:         "2025-12-15T06:00:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "not-a-time",
				ArrivalTimezone:   "Asia/Makassar",
			},
		}
		_, err := adapt(f)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse arrival")
	})

	t.Run("arrival not after departure", func(t *testing.T) {
		f := lionFlight{
			ID: "JT999",
			Schedule: lionSchedule{
				Departure:         "2025-12-15T08:00:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T06:00:00",
				ArrivalTimezone:   "Asia/Jakarta",
			},
		}
		_, err := adapt(f)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "arrival not after departure")
	})

	t.Run("IsDirect sets stops to 0", func(t *testing.T) {
		f := lionFlight{
			ID: "JT999",
			Schedule: lionSchedule{
				Departure:         "2025-12-15T06:00:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T08:00:00",
				ArrivalTimezone:   "Asia/Makassar",
			},
			IsDirect:  true,
			StopCount: 3,
		}
		flight, err := adapt(f)
		require.NoError(t, err)
		assert.Equal(t, 0, flight.Stops)
	})

	t.Run("WiFi and meals included in amenities", func(t *testing.T) {
		f := lionFlight{
			ID: "JT999",
			Schedule: lionSchedule{
				Departure:         "2025-12-15T06:00:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T08:00:00",
				ArrivalTimezone:   "Asia/Makassar",
			},
			IsDirect: true,
			Services: lionServices{
				WiFiAvailable: true,
				MealsIncluded: true,
			},
		}
		flight, err := adapt(f)
		require.NoError(t, err)
		assert.Contains(t, flight.Amenities, "wifi")
		assert.Contains(t, flight.Amenities, "meal")
	})

	t.Run("no WiFi no meals results in empty amenities", func(t *testing.T) {
		f := lionFlight{
			ID: "JT999",
			Schedule: lionSchedule{
				Departure:         "2025-12-15T06:00:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T08:00:00",
				ArrivalTimezone:   "Asia/Makassar",
			},
			IsDirect: true,
		}
		flight, err := adapt(f)
		require.NoError(t, err)
		assert.Empty(t, flight.Amenities)
	})
}
