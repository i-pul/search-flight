package airasia

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

// mustSearch retries up to maxAttempts to work around the 10% simulated failure rate.
func mustSearch(t *testing.T, r *Repository, req domain.SearchRequest) []domain.Flight {
	t.Helper()
	for range 20 {
		flights, err := r.Search(context.Background(), req)
		if err == nil {
			return flights
		}
	}
	t.Fatal("Search failed after 20 attempts")
	return nil
}

func TestSearch(t *testing.T) {
	r := New()

	t.Run("returns flights", func(t *testing.T) {
		flights := mustSearch(t, r, defaultReq)
		assert.NotEmpty(t, flights)
	})

	t.Run("QZ520 field normalization", func(t *testing.T) {
		flights := mustSearch(t, r, defaultReq)

		var qz520 *domain.Flight
		for i := range flights {
			if flights[i].FlightNumber == "QZ520" {
				qz520 = &flights[i]
				break
			}
		}
		require.NotNil(t, qz520, "QZ520 not found in results")

		tests := []struct {
			field string
			got   interface{}
			want  interface{}
		}{
			{"provider", qz520.Provider, "AirAsia"},
			{"airline code", qz520.Airline.Code, "QZ"},
			{"departure airport", qz520.Departure.Airport, "CGK"},
			{"arrival airport", qz520.Arrival.Airport, "DPS"},
			{"price amount", qz520.Price.Amount, float64(650000)},
			{"price currency", qz520.Price.Currency, "IDR"},
			{"stops", qz520.Stops, 0},
			{"duration minutes", qz520.Duration.TotalMinutes, 100},
			{"available seats", qz520.AvailableSeats, 67},
		}
		for _, tc := range tests {
			assert.Equal(t, tc.want, tc.got, tc.field)
		}
		assert.NotZero(t, qz520.Departure.Timestamp, "departure timestamp")
		assert.Greater(t, qz520.Arrival.Timestamp, qz520.Departure.Timestamp, "arrival after departure")
	})

	t.Run("QZ7250 with stop and duration", func(t *testing.T) {
		flights := mustSearch(t, r, defaultReq)

		for _, f := range flights {
			if f.FlightNumber == "QZ7250" {
				assert.Equal(t, 1, f.Stops, "QZ7250 stops")
				assert.Equal(t, 260, f.Duration.TotalMinutes, "QZ7250 duration (4.33h rounded)")
				return
			}
		}
		t.Error("QZ7250 not found in results")
	})

	t.Run("simulated failure rate is partial not total", func(t *testing.T) {
		failures := 0
		for i := 0; i < 50; i++ {
			_, err := r.Search(context.Background(), defaultReq)
			if err != nil {
				failures++
			}
		}
		assert.Less(t, failures, 50, "all 50 attempts failed — unexpected")
	})

	t.Run("context canceled returns error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond)

		_, err := r.Search(ctx, defaultReq)
		assert.Error(t, err)
	})
}
