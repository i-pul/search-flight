package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func baseMatchFlight() Flight {
	return Flight{
		Departure:      FlightPoint{Airport: "CGK", Datetime: "2025-12-15T06:00:00+07:00"},
		Arrival:        FlightPoint{Airport: "DPS"},
		CabinClass:     "economy",
		AvailableSeats: 5,
	}
}

func baseMatchReq() SearchRequest {
	return SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		CabinClass:    "economy",
		Passengers:    2,
	}
}

func TestMatches_FullMatch(t *testing.T) {
	assert.True(t, baseMatchReq().Matches(baseMatchFlight()))
}

func TestMatches_EmptyReqMatchesAnything(t *testing.T) {
	assert.True(t, SearchRequest{}.Matches(baseMatchFlight()))
}

func TestMatches_OriginMismatch(t *testing.T) {
	req := baseMatchReq()
	req.Origin = "SUB"
	assert.False(t, req.Matches(baseMatchFlight()))
}

func TestMatches_DestinationMismatch(t *testing.T) {
	req := baseMatchReq()
	req.Destination = "JOG"
	assert.False(t, req.Matches(baseMatchFlight()))
}

func TestMatches_DepartureDateMismatch(t *testing.T) {
	req := baseMatchReq()
	req.DepartureDate = "2025-12-20"
	assert.False(t, req.Matches(baseMatchFlight()))
}

func TestMatches_DepartureDateMatch(t *testing.T) {
	req := baseMatchReq()
	req.DepartureDate = "2025-12-15"
	assert.True(t, req.Matches(baseMatchFlight()))
}

func TestMatches_CabinClassMismatch(t *testing.T) {
	req := baseMatchReq()
	req.CabinClass = "business"
	assert.False(t, req.Matches(baseMatchFlight()))
}

func TestMatches_CabinClassCaseInsensitive(t *testing.T) {
	req := baseMatchReq()
	req.CabinClass = "ECONOMY"
	assert.True(t, req.Matches(baseMatchFlight()))
}

func TestMatches_InsufficientSeats(t *testing.T) {
	req := baseMatchReq()
	req.Passengers = 10 // only 5 available
	assert.False(t, req.Matches(baseMatchFlight()))
}

func TestMatches_ExactSeats(t *testing.T) {
	req := baseMatchReq()
	req.Passengers = 5
	assert.True(t, req.Matches(baseMatchFlight()))
}

func TestMatches_ShortDatetimeNotFiltered(t *testing.T) {
	f := baseMatchFlight()
	f.Departure.Datetime = "2025" // len < 10, date filter skipped
	req := baseMatchReq()
	req.DepartureDate = "2025-12-99" // would fail if compared
	assert.True(t, req.Matches(f))
}
