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

	t.Run("context canceled returns error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond)

		_, err := r.Search(ctx, defaultReq)
		assert.Error(t, err)
	})
}
