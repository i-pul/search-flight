package flight

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/i-pul/search-flight/internal/domain"
	flightrepo "github.com/i-pul/search-flight/internal/repository/flight"
	"github.com/i-pul/search-flight/internal/repository/flight/airasia"
	"github.com/i-pul/search-flight/internal/repository/flight/batikair"
	"github.com/i-pul/search-flight/internal/repository/flight/garuda"
	"github.com/i-pul/search-flight/internal/repository/flight/lionair"
)

func allRepos() []flightrepo.Repository {
	return []flightrepo.Repository{garuda.New(), lionair.New(), batikair.New(), airasia.New()}
}

var baseReq = domain.SearchRequest{
	Origin:        "CGK",
	Destination:   "DPS",
	DepartureDate: "2025-12-15",
	Passengers:    1,
	CabinClass:    "economy",
}

func TestFlightSearchUsecase(t *testing.T) {
	maxStops := 0

	tests := []struct {
		name   string
		filter domain.FilterParams
		sort   domain.SortParams
		check  func(t *testing.T, resp *domain.SearchResponse)
	}{
		{
			name:   "queries all 4 repositories",
			filter: domain.FilterParams{},
			sort:   domain.SortParams{By: domain.SortByPriceAsc},
			check: func(t *testing.T, resp *domain.SearchResponse) {
				assert.Equal(t, 4, resp.Metadata.ProvidersQueried)
				assert.NotZero(t, resp.Metadata.TotalResults)
				assert.Equal(t, "CGK", resp.SearchCriteria.Origin)
			},
		},
		{
			name:   "results sorted by price ascending",
			filter: domain.FilterParams{},
			sort:   domain.SortParams{By: domain.SortByPriceAsc},
			check: func(t *testing.T, resp *domain.SearchResponse) {
				for i := 1; i < len(resp.Flights); i++ {
					assert.GreaterOrEqual(t, resp.Flights[i].Price.Amount, resp.Flights[i-1].Price.Amount,
						"not sorted by price asc at index %d", i)
				}
			},
		},
		{
			name:   "max stops filter excludes connecting flights",
			filter: domain.FilterParams{MaxStops: &maxStops},
			sort:   domain.SortParams{},
			check: func(t *testing.T, resp *domain.SearchResponse) {
				for _, f := range resp.Flights {
					assert.Equal(t, 0, f.Stops, "flight %s has stops despite MaxStops=0", f.FlightNumber)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := New(allRepos(), DefaultScoreWeights(), 5*time.Second)
			resp, err := uc.Search(context.Background(), baseReq, tc.filter, tc.sort)
			require.NoError(t, err)
			tc.check(t, resp)
		})
	}
}

// slowRepo is a repository that blocks until its context is cancelled.
type slowRepo struct{}

func (s *slowRepo) Name() string { return "SlowProvider" }
func (s *slowRepo) Search(ctx context.Context, _ domain.SearchRequest) ([]domain.Flight, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestFlightSearchUsecase_Timeout(t *testing.T) {
	repos := []flightrepo.Repository{
		garuda.New(),  // fast provider
		&slowRepo{},   // never responds — gets killed by timeout
	}
	uc := New(repos, DefaultScoreWeights(), 600*time.Millisecond)

	resp, err := uc.Search(context.Background(), baseReq, domain.FilterParams{}, domain.SortParams{})
	require.NoError(t, err) // timeout is non-fatal; we get partial results
	assert.Equal(t, 2, resp.Metadata.ProvidersQueried)
	assert.Equal(t, 1, resp.Metadata.ProvidersFailed)   // slow repo timed out
	assert.Equal(t, 1, resp.Metadata.ProvidersSucceeded) // garuda returned
}
