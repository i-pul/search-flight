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

func TestName(t *testing.T) {
	assert.Equal(t, "BatikAir", New().Name())
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
		f := batikFlight{FlightNumber: "ID999", DepartureDateTime: "not-a-time"}
		_, err := adapt(f)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse departure")
	})

	t.Run("bad arrival time", func(t *testing.T) {
		f := batikFlight{
			FlightNumber:      "ID999",
			DepartureDateTime: "2025-12-15T06:00:00+0700",
			ArrivalDateTime:   "not-a-time",
		}
		_, err := adapt(f)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse arrival")
	})

	t.Run("arrival not after departure", func(t *testing.T) {
		f := batikFlight{
			FlightNumber:      "ID999",
			DepartureDateTime: "2025-12-15T08:00:00+0700",
			ArrivalDateTime:   "2025-12-15T06:00:00+0700",
		}
		_, err := adapt(f)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "arrival not after departure")
	})

	t.Run("invalid travel time falls back to computed duration", func(t *testing.T) {
		f := batikFlight{
			FlightNumber:      "ID999",
			DepartureDateTime: "2025-12-15T06:00:00+0700",
			ArrivalDateTime:   "2025-12-15T07:45:00+0700",
			TravelTime:        "not-parseable",
		}
		flight, err := adapt(f)
		require.NoError(t, err)
		assert.Equal(t, 105, flight.Duration.TotalMinutes) // 1h 45m = 105 min
	})
}

func TestParseBaggageInfo(t *testing.T) {
	t.Run("comma-separated splits into carry-on and checked", func(t *testing.T) {
		b := parseBaggageInfo("7kg cabin, 20kg checked")
		assert.Equal(t, "7kg cabin", b.CarryOn)
		assert.Equal(t, "20kg checked", b.Checked)
	})

	t.Run("no comma keeps whole string as carry-on", func(t *testing.T) {
		b := parseBaggageInfo("cabin only")
		assert.Equal(t, "cabin only", b.CarryOn)
		assert.Equal(t, "", b.Checked)
	})

	t.Run("max one split on first comma", func(t *testing.T) {
		b := parseBaggageInfo("a, b, c")
		assert.Equal(t, "a", b.CarryOn)
		assert.Equal(t, "b, c", b.Checked)
	})
}

func TestAirportCity(t *testing.T) {
	tests := []struct {
		iata string
		want string
	}{
		{"CGK", "Jakarta"},
		{"DPS", "Denpasar"},
		{"SUB", "Surabaya"},
		{"UPG", "Makassar"},
		{"SOC", "Surakarta"},
		{"JOG", "Yogyakarta"},
		{"XXX", "XXX"}, // unknown falls back to IATA code
	}
	for _, tc := range tests {
		t.Run(tc.iata, func(t *testing.T) {
			assert.Equal(t, tc.want, airportCity(tc.iata))
		})
	}
}
