package flight_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/i-pul/search-flight/internal/domain"
	flighth "github.com/i-pul/search-flight/internal/handler/flight"
)

// capturingMock captures the FilterParams and SortParams passed by the handler.
type capturingMock struct {
	capturedFilter domain.FilterParams
	capturedSort   domain.SortParams
}

func (m *capturingMock) Search(_ context.Context, _ domain.SearchRequest, fp domain.FilterParams, sp domain.SortParams) (*domain.SearchResponse, error) {
	m.capturedFilter = fp
	m.capturedSort = sp
	return &domain.SearchResponse{}, nil
}

// ---- Validation: extra cases ------------------------------------------------

func TestSearchFlights_ValidationErrors_Extra(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{
			name:    "missing cabinClass",
			body:    `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1,"cabinClass":""}`,
			wantMsg: "cabinClass is required",
		},
		{
			name:    "invalid returnDate format",
			body:    `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1,"cabinClass":"economy","returnDate":"22-12-2025"}`,
			wantMsg: "returnDate must be in YYYY-MM-DD format",
		},
		{
			name:    "returnDate before departureDate",
			body:    `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1,"cabinClass":"economy","returnDate":"2025-12-10"}`,
			wantMsg: "returnDate must not be before departureDate",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := flighth.New(&mockUsecase{})
			ctx := call(h, tc.body)

			assert.Equal(t, 400, ctx.Response.StatusCode())
			m := decodeError(t, ctx.Response.Body())
			assert.Equal(t, "validation_error", m["error"])
			assert.Contains(t, m["message"], tc.wantMsg)
		})
	}
}

// ---- Filter time fields -----------------------------------------------------

func TestSearchFlights_FilterTimeFields(t *testing.T) {
	tests := []struct {
		name    string
		filters map[string]any
		check   func(t *testing.T, fp domain.FilterParams)
	}{
		{
			name:    "departAfter sets EarliestDepart",
			filters: map[string]any{"departAfter": "06:00"},
			check: func(t *testing.T, fp domain.FilterParams) {
				require.NotNil(t, fp.EarliestDepart)
				assert.Equal(t, 6, fp.EarliestDepart.Hour())
			},
		},
		{
			name:    "departBefore sets LatestDepart",
			filters: map[string]any{"departBefore": "20:00"},
			check: func(t *testing.T, fp domain.FilterParams) {
				require.NotNil(t, fp.LatestDepart)
				assert.Equal(t, 20, fp.LatestDepart.Hour())
			},
		},
		{
			name:    "arriveAfter sets EarliestArrive",
			filters: map[string]any{"arriveAfter": "08:30"},
			check: func(t *testing.T, fp domain.FilterParams) {
				require.NotNil(t, fp.EarliestArrive)
				assert.Equal(t, 8, fp.EarliestArrive.Hour())
				assert.Equal(t, 30, fp.EarliestArrive.Minute())
			},
		},
		{
			name:    "arriveBefore sets LatestArrive",
			filters: map[string]any{"arriveBefore": "22:00"},
			check: func(t *testing.T, fp domain.FilterParams) {
				require.NotNil(t, fp.LatestArrive)
				assert.Equal(t, 22, fp.LatestArrive.Hour())
			},
		},
		{
			name: "scalar filter fields forwarded",
			filters: map[string]any{
				"minPrice":    500000.0,
				"maxPrice":    1500000.0,
				"maxStops":    0,
				"maxDuration": 120,
				"airlines":    []string{"GA", "JT"},
			},
			check: func(t *testing.T, fp domain.FilterParams) {
				require.NotNil(t, fp.MinPrice)
				assert.Equal(t, 500000.0, *fp.MinPrice)
				require.NotNil(t, fp.MaxPrice)
				assert.Equal(t, 1500000.0, *fp.MaxPrice)
				require.NotNil(t, fp.MaxStops)
				assert.Equal(t, 0, *fp.MaxStops)
				require.NotNil(t, fp.MaxDuration)
				assert.Equal(t, 120, *fp.MaxDuration)
				assert.Equal(t, []string{"GA", "JT"}, fp.Airlines)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &capturingMock{}
			h := flighth.New(mock)

			body := writeJSON(t, map[string]any{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": "2025-12-15",
				"passengers":    1,
				"cabinClass":    "economy",
				"filters":       tc.filters,
			})

			ctx := call(h, body)
			require.Equal(t, 200, ctx.Response.StatusCode())
			tc.check(t, mock.capturedFilter)
		})
	}
}

// ---- Invalid time filter errors ---------------------------------------------

func TestSearchFlights_InvalidFilterErrors(t *testing.T) {
	fields := []struct {
		name  string
		field string
	}{
		{"departBefore bad time", "departBefore"},
		{"arriveAfter bad time", "arriveAfter"},
		{"arriveBefore bad time", "arriveBefore"},
	}

	for _, tc := range fields {
		t.Run(tc.name, func(t *testing.T) {
			body := writeJSON(t, map[string]any{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": "2025-12-15",
				"passengers":    1,
				"cabinClass":    "economy",
				"filters":       map[string]any{tc.field: "not-a-time"},
			})
			h := flighth.New(&mockUsecase{})
			ctx := call(h, body)
			assert.Equal(t, 400, ctx.Response.StatusCode())
			m := decodeError(t, ctx.Response.Body())
			assert.Equal(t, "invalid_filter", m["error"])
		})
	}
}

// ---- Sort options ------------------------------------------------------------

func TestSearchFlights_SortOptions(t *testing.T) {
	sortOptions := []string{
		"price_asc", "price_desc",
		"duration_asc", "duration_desc",
		"departure_time", "arrival_time",
		"best_value",
		"unknown_sort", // falls back to price_asc
		"",             // nil sort — falls back to price_asc
	}

	for _, opt := range sortOptions {
		opt := opt
		t.Run("sort="+opt, func(t *testing.T) {
			mock := &capturingMock{}
			h := flighth.New(mock)

			req := map[string]any{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": "2025-12-15",
				"passengers":    1,
				"cabinClass":    "economy",
			}
			if opt != "" {
				req["sort"] = map[string]any{"by": opt}
			}

			ctx := call(h, writeJSON(t, req))
			assert.Equal(t, 200, ctx.Response.StatusCode())
		})
	}
}
