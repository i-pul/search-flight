package flight

import (
	"context"
	"log/slog"
	"time"

	"github.com/i-pul/search-flight/internal/domain"
	flightrepo "github.com/i-pul/search-flight/internal/repository/flight"
	"github.com/i-pul/search-flight/internal/util"
	"golang.org/x/sync/errgroup"
)

type FlightUsecase interface {
	Search(ctx context.Context, req domain.SearchRequest, fp domain.FilterParams, sp domain.SortParams) (*domain.SearchResponse, error)
}

type FlightSearchUsecase struct {
	repos   []flightrepo.Repository
	weights ScoreWeights
	timeout time.Duration
	retry   RetryConfig
}

// RetryConfig controls per-provider retry behaviour inside the usecase.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

func New(repos []flightrepo.Repository, weights ScoreWeights, timeout time.Duration, retry RetryConfig) *FlightSearchUsecase {
	return &FlightSearchUsecase{repos: repos, weights: weights, timeout: timeout, retry: retry}
}

type legResult struct {
	flights []domain.Flight
	failed  int
}

// queryProviders queries all repositories concurrently and returns the aggregated flights
// and the count of failed providers.
func (u *FlightSearchUsecase) queryProviders(ctx context.Context, req domain.SearchRequest) legResult {
	type repoResult struct {
		flights []domain.Flight
		err     error
	}

	results := make([]repoResult, len(u.repos))
	eg, egCtx := errgroup.WithContext(ctx)
	for i, r := range u.repos {
		i, r := i, r
		eg.Go(func() error {
			flights, err := util.Retry(egCtx, u.retry.MaxAttempts, u.retry.BaseDelay, func() ([]domain.Flight, error) {
				return r.Search(egCtx, req)
			})
			results[i] = repoResult{flights: flights, err: err}
			return nil
		})
	}
	_ = eg.Wait()

	var all []domain.Flight
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
		all = append(all, res.flights...)
	}
	return legResult{flights: all, failed: failed}
}

func (u *FlightSearchUsecase) Search(
	ctx context.Context,
	req domain.SearchRequest,
	fp domain.FilterParams,
	sp domain.SortParams,
) (*domain.SearchResponse, error) {
	start := time.Now()

	slog.InfoContext(ctx, "search start", "request", req)

	tctx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	// Run outbound and return legs concurrently under the shared timeout.
	var outbound, ret legResult

	if req.ReturnDate != nil {
		returnReq := domain.SearchRequest{
			Origin:        req.Destination,
			Destination:   req.Origin,
			DepartureDate: *req.ReturnDate,
			Passengers:    req.Passengers,
			CabinClass:    req.CabinClass,
		}

		var eg errgroup.Group
		eg.Go(func() error { outbound = u.queryProviders(tctx, req); return nil })
		eg.Go(func() error { ret = u.queryProviders(tctx, returnReq); return nil })
		_ = eg.Wait()
	} else {
		outbound = u.queryProviders(tctx, req)
	}

	// Apply filters, scoring, sort to outbound flights.
	outbound.flights = ApplyFilters(outbound.flights, fp)
	if sp.By == domain.SortByBestValue {
		ComputeBestValueScores(outbound.flights, u.weights)
	}
	ApplySort(outbound.flights, sp)

	// Apply filters, scoring, sort to return flights.
	// Time-of-day filters are stripped because they were anchored to the outbound date.
	var returnFlights []domain.Flight
	if req.ReturnDate != nil {
		returnFP := stripTimeFilters(fp)
		ret.flights = ApplyFilters(ret.flights, returnFP)
		if sp.By == domain.SortByBestValue {
			ComputeBestValueScores(ret.flights, u.weights)
		}
		ApplySort(ret.flights, sp)
		returnFlights = ret.flights
	}

	succeeded := len(u.repos) - outbound.failed

	resp := &domain.SearchResponse{
		SearchCriteria: toCriteria(req),
		Metadata: domain.SearchMetadata{
			TotalResults:       len(outbound.flights),
			ProvidersQueried:   len(u.repos),
			ProvidersSucceeded: succeeded,
			ProvidersFailed:    outbound.failed,
			SearchTimeMs:       time.Since(start).Milliseconds(),
			CacheHit:           false,
		},
		Flights: outbound.flights,
	}

	if req.ReturnDate != nil {
		resp.Metadata.ReturnResults = len(returnFlights)
		resp.ReturnFlights = returnFlights
	}

	return resp, nil
}

// stripTimeFilters returns a copy of fp with departure/arrival time windows removed.
// Used for the return leg, whose date differs from the outbound date used to compute
// the original time windows.
func stripTimeFilters(fp domain.FilterParams) domain.FilterParams {
	fp.EarliestDepart = nil
	fp.LatestDepart = nil
	fp.EarliestArrive = nil
	fp.LatestArrive = nil
	return fp
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
