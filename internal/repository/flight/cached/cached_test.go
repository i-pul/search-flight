package cached_test

import (
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/i-pul/search-flight/internal/domain"
	"github.com/i-pul/search-flight/internal/repository/flight/cached"
)

func newStore(ttl time.Duration) *cached.Store {
	return cached.New(gocache.New(ttl, ttl*2), ttl)
}

func baseReq() domain.SearchRequest {
	return domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}
}

func TestGet_MissOnEmptyStore(t *testing.T) {
	s := newStore(time.Minute)
	flights, ok := s.Get("Garuda", baseReq())
	assert.False(t, ok)
	assert.Nil(t, flights)
}

func TestSetAndGet_ReturnsStoredFlights(t *testing.T) {
	s := newStore(time.Minute)
	req := baseReq()
	want := []domain.Flight{{ID: "F1"}, {ID: "F2"}}

	s.Set("Garuda", req, want)

	got, ok := s.Get("Garuda", req)
	require.True(t, ok)
	assert.Equal(t, want, got)
}

func TestGet_DifferentProvidersSeparateEntries(t *testing.T) {
	s := newStore(time.Minute)
	req := baseReq()

	garuda := []domain.Flight{{ID: "G1"}}
	lion := []domain.Flight{{ID: "L1"}}

	s.Set("Garuda", req, garuda)
	s.Set("LionAir", req, lion)

	gotG, okG := s.Get("Garuda", req)
	require.True(t, okG)
	assert.Equal(t, garuda, gotG)

	gotL, okL := s.Get("LionAir", req)
	require.True(t, okL)
	assert.Equal(t, lion, gotL)
}

func TestGet_DifferentRequestsSeparateEntries(t *testing.T) {
	s := newStore(time.Minute)

	reqDPS := baseReq()
	reqSUB := baseReq()
	reqSUB.Destination = "SUB"

	flightsDPS := []domain.Flight{{ID: "DPS1"}}
	flightsSUB := []domain.Flight{{ID: "SUB1"}}

	s.Set("Garuda", reqDPS, flightsDPS)
	s.Set("Garuda", reqSUB, flightsSUB)

	gotDPS, ok := s.Get("Garuda", reqDPS)
	require.True(t, ok)
	assert.Equal(t, flightsDPS, gotDPS)

	gotSUB, ok := s.Get("Garuda", reqSUB)
	require.True(t, ok)
	assert.Equal(t, flightsSUB, gotSUB)
}

func TestGet_MissAfterTTLExpires(t *testing.T) {
	s := newStore(5 * time.Millisecond)
	req := baseReq()

	s.Set("Garuda", req, []domain.Flight{{ID: "F1"}})

	time.Sleep(30 * time.Millisecond)

	_, ok := s.Get("Garuda", req)
	assert.False(t, ok, "entry should have expired")
}

func TestSet_OverwritesPreviousEntry(t *testing.T) {
	s := newStore(time.Minute)
	req := baseReq()

	s.Set("Garuda", req, []domain.Flight{{ID: "old"}})
	s.Set("Garuda", req, []domain.Flight{{ID: "new"}})

	got, ok := s.Get("Garuda", req)
	require.True(t, ok)
	assert.Equal(t, []domain.Flight{{ID: "new"}}, got)
}
