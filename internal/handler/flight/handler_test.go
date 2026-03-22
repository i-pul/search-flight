package flight_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/i-pul/search-flight/internal/domain"
	flighth "github.com/i-pul/search-flight/internal/handler/flight"
	"github.com/i-pul/search-flight/internal/middleware"
)

// mockUsecase is a controllable stub for FlightUsecase.
type mockUsecase struct {
	resp *domain.SearchResponse
	err  error
}

func (m *mockUsecase) Search(_ context.Context, _ domain.SearchRequest, _ domain.FilterParams, _ domain.SortParams) (*domain.SearchResponse, error) {
	return m.resp, m.err
}

// call invokes the handler (wrapped in Trace middleware) and returns the RequestCtx.
func call(h *flighth.Handler, body string) *fasthttp.RequestCtx {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBodyString(body)
	middleware.Trace(h.SearchFlights)(&ctx)
	return &ctx
}

func TestSearchFlights_InvalidJSON(t *testing.T) {
	h := flighth.New(&mockUsecase{})
	ctx := call(h, `not json`)

	assert.Equal(t, 400, ctx.Response.StatusCode())
	m := decodeError(t, ctx.Response.Body())
	assert.Equal(t, "invalid_request", m["error"])
}

func TestSearchFlights_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{
			name:    "missing origin",
			body:    `{"origin":"","destination":"DPS","departureDate":"2025-12-15","passengers":1}`,
			wantMsg: "origin must be a 3-letter IATA code",
		},
		{
			name:    "invalid origin length",
			body:    `{"origin":"CGKK","destination":"DPS","departureDate":"2025-12-15","passengers":1}`,
			wantMsg: "origin must be a 3-letter IATA code",
		},
		{
			name:    "missing destination",
			body:    `{"origin":"CGK","destination":"","departureDate":"2025-12-15","passengers":1}`,
			wantMsg: "destination must be a 3-letter IATA code",
		},
		{
			name:    "same origin and destination",
			body:    `{"origin":"CGK","destination":"CGK","departureDate":"2025-12-15","passengers":1}`,
			wantMsg: "origin and destination must be different",
		},
		{
			name:    "invalid departure date format",
			body:    `{"origin":"CGK","destination":"DPS","departureDate":"15-12-2025","passengers":1}`,
			wantMsg: "departureDate must be in YYYY-MM-DD format",
		},
		{
			name:    "zero passengers",
			body:    `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":0}`,
			wantMsg: "passengers must be at least 1",
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

func TestSearchFlights_InvalidFilter(t *testing.T) {
	body := writeJSON(t, map[string]any{
		"origin":        "CGK",
		"destination":   "DPS",
		"departureDate": "2025-12-15",
		"passengers":    1,
		"cabinClass":    "economy",
		"filters":       map[string]any{"departAfter": "not-a-time"},
	})
	h := flighth.New(&mockUsecase{})
	ctx := call(h, body)

	assert.Equal(t, 400, ctx.Response.StatusCode())
	m := decodeError(t, ctx.Response.Body())
	assert.Equal(t, "invalid_filter", m["error"])
}

func TestSearchFlights_UsecaseError(t *testing.T) {
	h := flighth.New(&mockUsecase{err: errors.New("provider down")})
	ctx := call(h, validBody(t))

	assert.Equal(t, 500, ctx.Response.StatusCode())
	m := decodeError(t, ctx.Response.Body())
	assert.Equal(t, "search_error", m["error"])
}

func TestSearchFlights_Success(t *testing.T) {
	fakeResp := &domain.SearchResponse{
		SearchCriteria: domain.SearchCriteria{
			Origin:      "CGK",
			Destination: "DPS",
		},
		Metadata: domain.SearchMetadata{
			TotalResults:       2,
			ProvidersQueried:   4,
			ProvidersSucceeded: 4,
		},
		Flights: []domain.Flight{
			{FlightNumber: "GA-101"},
			{FlightNumber: "GA-202"},
		},
	}

	h := flighth.New(&mockUsecase{resp: fakeResp})
	ctx := call(h, validBody(t))

	assert.Equal(t, 200, ctx.Response.StatusCode())
	assert.Equal(t, "application/json; charset=utf-8",
		string(ctx.Response.Header.ContentType()))

	var got domain.SearchResponse
	require.NoError(t, json.Unmarshal(ctx.Response.Body(), &got))
	assert.Equal(t, 2, got.Metadata.TotalResults)
	assert.Len(t, got.Flights, 2)
	assert.Equal(t, "GA-101", got.Flights[0].FlightNumber)
}

func TestSearchFlights_TraceIDHeader(t *testing.T) {
	h := flighth.New(&mockUsecase{resp: &domain.SearchResponse{}})
	ctx := call(h, validBody(t))

	traceID := string(ctx.Response.Header.Peek("X-Trace-Id"))
	assert.NotEmpty(t, traceID, "X-Trace-Id header must be set")
	// UUID v4: 8-4-4-4-12 hex chars separated by hyphens
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, traceID)
}
