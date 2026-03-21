package garuda

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

	t.Run("GA400 field normalization", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		var ga400 *domain.Flight
		for i := range flights {
			if flights[i].FlightNumber == "GA400" {
				ga400 = &flights[i]
				break
			}
		}
		require.NotNil(t, ga400, "GA400 not found in results")

		tests := []struct {
			field string
			got   interface{}
			want  interface{}
		}{
			{"provider", ga400.Provider, "GarudaIndonesia"},
			{"airline code", ga400.Airline.Code, "GA"},
			{"departure airport", ga400.Departure.Airport, "CGK"},
			{"arrival airport", ga400.Arrival.Airport, "DPS"},
			{"price amount", ga400.Price.Amount, float64(1250000)},
			{"price currency", ga400.Price.Currency, "IDR"},
			{"stops", ga400.Stops, 0},
			{"duration minutes", ga400.Duration.TotalMinutes, 110},
			{"duration formatted", ga400.Duration.Formatted, "1h 50m"},
		}
		for _, tc := range tests {
			assert.Equal(t, tc.want, tc.got, tc.field)
		}
		assert.NotZero(t, ga400.Departure.Timestamp, "departure timestamp")
		assert.Greater(t, ga400.Arrival.Timestamp, ga400.Departure.Timestamp, "arrival after departure")
	})

	t.Run("GA315 multi-segment stop count", func(t *testing.T) {
		flights, err := r.Search(context.Background(), defaultReq)
		require.NoError(t, err)

		for _, f := range flights {
			if f.FlightNumber == "GA315" {
				assert.Equal(t, 1, f.Stops, "GA315: stops derived from 2 segments")
				return
			}
		}
		t.Skip("GA315 not present")
	})

	t.Run("context canceled returns error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(5 * time.Millisecond)

		_, err := r.Search(ctx, defaultReq)
		assert.Error(t, err)
	})
}
