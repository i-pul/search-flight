package cached

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/i-pul/search-flight/internal/domain"
)

// Store is a per-provider in-process cache.
// It is a standalone component: it knows nothing about how flights are fetched;
// callers are responsible for checking the store before querying providers and
// writing successful results back.
type Store struct {
	cache *gocache.Cache
	ttl   time.Duration
}

// New returns a Store backed by the supplied go-cache instance.
func New(cache *gocache.Cache, ttl time.Duration) *Store {
	return &Store{cache: cache, ttl: ttl}
}

// Get returns cached flights for the given provider + request combination.
// The second return value is false when no entry exists or it has expired.
func (s *Store) Get(providerName string, req domain.SearchRequest) ([]domain.Flight, bool) {
	v, found := s.cache.Get(key(providerName, req))
	if !found {
		return nil, false
	}
	return v.([]domain.Flight), true
}

// Set stores flights under the given provider + request combination for the
// duration of the Store's TTL. Only call this after a successful provider query.
func (s *Store) Set(providerName string, req domain.SearchRequest, flights []domain.Flight) {
	s.cache.Set(key(providerName, req), flights, s.ttl)
}

func key(providerName string, req domain.SearchRequest) string {
	return fmt.Sprintf("%s:%s:%s:%s:%d:%s",
		providerName,
		req.Origin,
		req.Destination,
		req.DepartureDate,
		req.Passengers,
		req.CabinClass,
	)
}
