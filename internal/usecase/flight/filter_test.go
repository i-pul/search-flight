package flight

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/i-pul/search-flight/internal/domain"
)

// makeFlights returns a slice of test flight fixtures with varying attributes.
func makeFlights() []domain.Flight {
	wib := time.FixedZone("WIB", 7*3600)
	return []domain.Flight{
		{
			ID:             "QZ520_AirAsia",
			Provider:       "AirAsia",
			Airline:        domain.Airline{Name: "AirAsia", Code: "QZ"},
			FlightNumber:   "QZ520",
			Departure:      domain.FlightPoint{Airport: "CGK", Timestamp: time.Date(2025, 12, 15, 4, 45, 0, 0, wib).Unix()},
			Arrival:        domain.FlightPoint{Airport: "DPS", Timestamp: time.Date(2025, 12, 15, 7, 25, 0, 0, wib).Unix()},
			Duration:       domain.Duration{TotalMinutes: 100},
			Stops:          0,
			Price:          domain.Price{Amount: 650000, Currency: "IDR"},
			BestValueScore: 80,
		},
		{
			ID:             "JT650_LionAir",
			Provider:       "LionAir",
			Airline:        domain.Airline{Name: "Lion Air", Code: "JT"},
			FlightNumber:   "JT650",
			Departure:      domain.FlightPoint{Airport: "CGK", Timestamp: time.Date(2025, 12, 15, 16, 20, 0, 0, wib).Unix()},
			Arrival:        domain.FlightPoint{Airport: "DPS", Timestamp: time.Date(2025, 12, 15, 21, 10, 0, 0, wib).Unix()},
			Duration:       domain.Duration{TotalMinutes: 230},
			Stops:          1,
			Price:          domain.Price{Amount: 780000, Currency: "IDR"},
			BestValueScore: 50,
		},
		{
			ID:             "GA400_GarudaIndonesia",
			Provider:       "GarudaIndonesia",
			Airline:        domain.Airline{Name: "Garuda Indonesia", Code: "GA"},
			FlightNumber:   "GA400",
			Departure:      domain.FlightPoint{Airport: "CGK", Timestamp: time.Date(2025, 12, 15, 6, 0, 0, 0, wib).Unix()},
			Arrival:        domain.FlightPoint{Airport: "DPS", Timestamp: time.Date(2025, 12, 15, 8, 50, 0, 0, wib).Unix()},
			Duration:       domain.Duration{TotalMinutes: 110},
			Stops:          0,
			Price:          domain.Price{Amount: 1250000, Currency: "IDR"},
			BestValueScore: 65,
		},
		{
			ID:             "QZ7250_AirAsia",
			Provider:       "AirAsia",
			Airline:        domain.Airline{Name: "AirAsia", Code: "QZ"},
			FlightNumber:   "QZ7250",
			Departure:      domain.FlightPoint{Airport: "CGK", Timestamp: time.Date(2025, 12, 15, 15, 15, 0, 0, wib).Unix()},
			Arrival:        domain.FlightPoint{Airport: "DPS", Timestamp: time.Date(2025, 12, 15, 20, 35, 0, 0, wib).Unix()},
			Duration:       domain.Duration{TotalMinutes: 260},
			Stops:          1,
			Price:          domain.Price{Amount: 485000, Currency: "IDR"},
			BestValueScore: 90,
		},
	}
}

func TestApplyFilters(t *testing.T) {
	wib := time.FixedZone("WIB", 7*3600)
	minPrice := float64(700000)
	maxPrice := float64(750000)
	maxStops := 0
	maxDur := 120
	earliest := time.Date(2025, 12, 15, 12, 0, 0, 0, wib)
	latestDepart := time.Date(2025, 12, 15, 10, 0, 0, 0, wib)
	earliestArrive := time.Date(2025, 12, 15, 9, 0, 0, 0, wib)
	latestArrive := time.Date(2025, 12, 15, 8, 0, 0, 0, wib)

	tests := []struct {
		name      string
		filter    domain.FilterParams
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "no filter returns all",
			filter:    domain.FilterParams{},
			wantCount: 4,
		},
		{
			name:      "min price 700000",
			filter:    domain.FilterParams{MinPrice: &minPrice},
			wantCount: 2,
			wantIDs:   []string{"JT650_LionAir", "GA400_GarudaIndonesia"},
		},
		{
			name:      "max price 750000",
			filter:    domain.FilterParams{MaxPrice: &maxPrice},
			wantCount: 2,
			wantIDs:   []string{"QZ520_AirAsia", "QZ7250_AirAsia"},
		},
		{
			name:      "max stops 0 (direct only)",
			filter:    domain.FilterParams{MaxStops: &maxStops},
			wantCount: 2,
			wantIDs:   []string{"QZ520_AirAsia", "GA400_GarudaIndonesia"},
		},
		{
			name:      "max duration 120 min",
			filter:    domain.FilterParams{MaxDuration: &maxDur},
			wantCount: 2,
			wantIDs:   []string{"QZ520_AirAsia", "GA400_GarudaIndonesia"},
		},
		{
			name:      "single airline GA",
			filter:    domain.FilterParams{Airlines: []string{"GA"}},
			wantCount: 1,
			wantIDs:   []string{"GA400_GarudaIndonesia"},
		},
		{
			name:      "multiple airlines QZ and JT",
			filter:    domain.FilterParams{Airlines: []string{"QZ", "JT"}},
			wantCount: 3,
		},
		{
			name:      "depart after 12:00",
			filter:    domain.FilterParams{EarliestDepart: &earliest},
			wantCount: 2,
			wantIDs:   []string{"JT650_LionAir", "QZ7250_AirAsia"},
		},
		{
			name:      "depart before 10:00",
			filter:    domain.FilterParams{LatestDepart: &latestDepart},
			wantCount: 2,
			wantIDs:   []string{"QZ520_AirAsia", "GA400_GarudaIndonesia"},
		},
		{
			name:      "arrive after 09:00",
			filter:    domain.FilterParams{EarliestArrive: &earliestArrive},
			wantCount: 2,
			wantIDs:   []string{"JT650_LionAir", "QZ7250_AirAsia"},
		},
		{
			name:      "arrive before 08:00 keeps only QZ520",
			filter:    domain.FilterParams{LatestArrive: &latestArrive},
			wantCount: 1,
			wantIDs:   []string{"QZ520_AirAsia"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplyFilters(makeFlights(), tc.filter)
			assert.Len(t, got, tc.wantCount)

			if len(tc.wantIDs) > 0 {
				gotIDs := make([]string, len(got))
				for i, f := range got {
					gotIDs[i] = f.ID
				}
				assert.ElementsMatch(t, tc.wantIDs, gotIDs)
			}
		})
	}
}

func TestApplySort(t *testing.T) {
	tests := []struct {
		name   string
		sortBy domain.SortBy
		check  func(t *testing.T, flights []domain.Flight)
	}{
		{
			name:   "price ascending",
			sortBy: domain.SortByPriceAsc,
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.GreaterOrEqual(t, flights[i].Price.Amount, flights[i-1].Price.Amount)
				}
			},
		},
		{
			name:   "price descending",
			sortBy: domain.SortByPriceDesc,
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.LessOrEqual(t, flights[i].Price.Amount, flights[i-1].Price.Amount)
				}
			},
		},
		{
			name:   "duration ascending",
			sortBy: domain.SortByDurationAsc,
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.GreaterOrEqual(t, flights[i].Duration.TotalMinutes, flights[i-1].Duration.TotalMinutes)
				}
			},
		},
		{
			name:   "duration descending",
			sortBy: domain.SortByDurationDesc,
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.LessOrEqual(t, flights[i].Duration.TotalMinutes, flights[i-1].Duration.TotalMinutes)
				}
			},
		},
		{
			name:   "depart time ascending",
			sortBy: domain.SortByDepartureTime,
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.GreaterOrEqual(t, flights[i].Departure.Timestamp, flights[i-1].Departure.Timestamp)
				}
			},
		},
		{
			name:   "arrival time ascending",
			sortBy: domain.SortByArrivalTime,
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.GreaterOrEqual(t, flights[i].Arrival.Timestamp, flights[i-1].Arrival.Timestamp)
				}
			},
		},
		{
			name:   "best value descending score",
			sortBy: domain.SortByBestValue,
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.LessOrEqual(t, flights[i].BestValueScore, flights[i-1].BestValueScore)
				}
			},
		},
		{
			name:   "default (empty) falls back to price ascending",
			sortBy: "",
			check: func(t *testing.T, flights []domain.Flight) {
				for i := 1; i < len(flights); i++ {
					assert.GreaterOrEqual(t, flights[i].Price.Amount, flights[i-1].Price.Amount)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flights := makeFlights()
			ApplySort(flights, domain.SortParams{By: tc.sortBy})
			tc.check(t, flights)
		})
	}
}
