package flight

import (
	"context"
	"log/slog"
	"time"

	"github.com/i-pul/search-flight/internal/domain"
	flightrepo "github.com/i-pul/search-flight/internal/repository/flight"
	"golang.org/x/sync/errgroup"
)

type FlightUsecase interface {
	Search(ctx context.Context, req domain.SearchRequest, fp domain.FilterParams, sp domain.SortParams) (*domain.SearchResponse, error)
}

type FlightSearchUsecase struct {
	repos   []flightrepo.Repository
	weights ScoreWeights
	timeout time.Duration
}

func New(repos []flightrepo.Repository, weights ScoreWeights, timeout time.Duration) *FlightSearchUsecase {
	return &FlightSearchUsecase{repos: repos, weights: weights, timeout: timeout}
}

func (u *FlightSearchUsecase) Search(
	ctx context.Context,
	req domain.SearchRequest,
	fp domain.FilterParams,
	sp domain.SortParams,
) (*domain.SearchResponse, error) {
	start := time.Now()

	slog.InfoContext(ctx, "search start", "request", req)

	type result struct {
		flights []domain.Flight
		err     error
	}

	results := make([]result, len(u.repos))

	tctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	eg, egCtx := errgroup.WithContext(tctx)
	for i, r := range u.repos {
		i, r := i, r
		eg.Go(func() error {
			flights, err := r.Search(egCtx, req)
			results[i] = result{flights: flights, err: err}
			return nil
		})
	}
	_ = eg.Wait()

	var allFlights []domain.Flight
	failed := 0
	for i, res := range results {
		if res.err != nil {
			slog.WarnContext(ctx, "provider failed",
				"provider", u.repos[i].Name(),
				"error", res.err,
			)
			failed++
			continue
		}
		slog.InfoContext(ctx, "provider ok",
			"provider", u.repos[i].Name(),
			"flights", len(res.flights),
		)
		allFlights = append(allFlights, res.flights...)
	}

	succeeded := len(u.repos) - failed

	allFlights = ApplyFilters(allFlights, fp)
	ComputeBestValueScores(allFlights, u.weights)
	ApplySort(allFlights, sp)

	return &domain.SearchResponse{
		SearchCriteria: toCriteria(req),
		Metadata: domain.SearchMetadata{
			TotalResults:       len(allFlights),
			ProvidersQueried:   len(u.repos),
			ProvidersSucceeded: succeeded,
			ProvidersFailed:    failed,
			SearchTimeMs:       time.Since(start).Milliseconds(),
			CacheHit:           false,
		},
		Flights: allFlights,
	}, nil
}

func toCriteria(req domain.SearchRequest) domain.SearchCriteria {
	return domain.SearchCriteria{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureDate: req.DepartureDate,
		ReturnDate:    req.ReturnDate,
		Passengers:    req.Passengers,
		CabinClass:    req.CabinClass,
	}
}
