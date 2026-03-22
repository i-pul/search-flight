package domain

import "time"

type FilterParams struct {
	MinPrice       *float64
	MaxPrice       *float64
	MaxStops       *int
	EarliestDepart *time.Time
	LatestDepart   *time.Time
	EarliestArrive *time.Time
	LatestArrive   *time.Time
	Airlines       []string
	MaxDuration    *int // in minutes
}

type SortBy string

const (
	SortByPriceAsc      SortBy = "price_asc"
	SortByPriceDesc     SortBy = "price_desc"
	SortByDurationAsc   SortBy = "duration_asc"
	SortByDurationDesc  SortBy = "duration_desc"
	SortByDepartureTime SortBy = "departure_time"
	SortByArrivalTime   SortBy = "arrival_time"
)

type SortParams struct {
	By SortBy
}
