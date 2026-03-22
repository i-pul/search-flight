package batikair

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

	t.Run("ID6514 field normalization", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		var id6514 *domain.Flight
		for i := range flights {
			if flights[i].FlightNumber == "ID6514" {
				id6514 = &flights[i]
				break
			}
		}
		require.NotNil(t, id6514, "ID6514 not found in results")

		tests := []struct {
			field string
			got   interface{}
			want  interface{}
		}{
			{"provider", id6514.Provider, "BatikAir"},
			{"airline code", id6514.Airline.Code, "ID"},
			{"departure airport", id6514.Departure.Airport, "CGK"},
			{"arrival airport", id6514.Arrival.Airport, "DPS"},
			{"price amount", id6514.Price.Amount, float64(1100000)},
			{"price currency", id6514.Price.Currency, "IDR"},
			{"stops", id6514.Stops, 0},
			{"duration minutes", id6514.Duration.TotalMinutes, 105},
			{"duration formatted", id6514.Duration.Formatted, "1h 45m"},
			{"available seats", id6514.AvailableSeats, 32},
		}
		for _, tc := range tests {
			assert.Equal(t, tc.want, tc.got, tc.field)
		}
		assert.NotZero(t, id6514.Departure.Timestamp, "departure timestamp")
		assert.Greater(t, id6514.Arrival.Timestamp, id6514.Departure.Timestamp, "arrival after departure")
	})

	t.Run("ID7042 with stop and duration", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		for _, f := range flights {
			if f.FlightNumber == "ID7042" {
				assert.Equal(t, 1, f.Stops, "ID7042 stops")
				assert.Equal(t, 185, f.Duration.TotalMinutes, "ID7042 duration (3h 5m)")
				return
			}
		}
		t.Error("ID7042 not found in results")
	})

	t.Run("baggage fields parsed from string", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		for _, f := range flights {
			assert.NotEmpty(t, f.Baggage.CarryOn, "flight %s: empty CarryOn", f.FlightNumber)
			assert.NotEmpty(t, f.Baggage.Checked, "flight %s: empty Checked", f.FlightNumber)
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
