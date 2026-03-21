package flight

import (
	"sort"
	"strings"

	"github.com/i-pul/search-flight/internal/domain"
)

func ApplyFilters(flights []domain.Flight, p domain.FilterParams) []domain.Flight {
	out := make([]domain.Flight, 0, len(flights))
	for _, f := range flights {
		if p.MinPrice != nil && f.Price.Amount < *p.MinPrice {
			continue
		}
		if p.MaxPrice != nil && f.Price.Amount > *p.MaxPrice {
			continue
		}
		if p.MaxStops != nil && f.Stops > *p.MaxStops {
			continue
		}
		if p.MaxDuration != nil && f.Duration.TotalMinutes > *p.MaxDuration {
			continue
		}
		if len(p.Airlines) > 0 && !containsAirline(p.Airlines, f.Airline.Code) {
			continue
		}
		if p.EarliestDepart != nil && f.Departure.Timestamp < p.EarliestDepart.Unix() {
			continue
		}
		if p.LatestDepart != nil && f.Departure.Timestamp > p.LatestDepart.Unix() {
			continue
		}
		if p.EarliestArrive != nil && f.Arrival.Timestamp < p.EarliestArrive.Unix() {
			continue
		}
		if p.LatestArrive != nil && f.Arrival.Timestamp > p.LatestArrive.Unix() {
			continue
		}
		out = append(out, f)
	}
	return out
}

func ApplySort(flights []domain.Flight, p domain.SortParams) {
	sort.SliceStable(flights, func(i, j int) bool {
		a, b := flights[i], flights[j]
		switch p.By {
		case domain.SortByPriceAsc:
			return a.Price.Amount < b.Price.Amount
		case domain.SortByPriceDesc:
			return a.Price.Amount > b.Price.Amount
		case domain.SortByDurationAsc:
			return a.Duration.TotalMinutes < b.Duration.TotalMinutes
		case domain.SortByDurationDesc:
			return a.Duration.TotalMinutes > b.Duration.TotalMinutes
		case domain.SortByDepartureTime:
			return a.Departure.Timestamp < b.Departure.Timestamp
		case domain.SortByArrivalTime:
			return a.Arrival.Timestamp < b.Arrival.Timestamp
		default:
			return a.Price.Amount < b.Price.Amount
		}
	})
}

func containsAirline(airlines []string, code string) bool {
	for _, a := range airlines {
		if strings.EqualFold(a, code) {
			return true
		}
	}
	return false
}
