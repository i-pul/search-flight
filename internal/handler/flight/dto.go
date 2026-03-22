package flight

import (
	"fmt"
	"time"

	"github.com/i-pul/search-flight/internal/domain"
)

type searchHTTPRequest struct {
	Origin        string          `json:"origin"`
	Destination   string          `json:"destination"`
	DepartureDate string          `json:"departureDate"`
	ReturnDate    *string         `json:"returnDate,omitempty"`
	Passengers    int             `json:"passengers"`
	CabinClass    string          `json:"cabinClass"`
	Filters       *filtersRequest `json:"filters,omitempty"`
	Sort          *sortRequest    `json:"sort,omitempty"`
}

type filtersRequest struct {
	MinPrice     *float64 `json:"minPrice,omitempty"`
	MaxPrice     *float64 `json:"maxPrice,omitempty"`
	MaxStops     *int     `json:"maxStops,omitempty"`
	MaxDuration  *int     `json:"maxDuration,omitempty"`
	Airlines     []string `json:"airlines,omitempty"`
	DepartAfter  *string  `json:"departAfter,omitempty"`
	DepartBefore *string  `json:"departBefore,omitempty"`
	ArriveAfter  *string  `json:"arriveAfter,omitempty"`
	ArriveBefore *string  `json:"arriveBefore,omitempty"`
}

type sortRequest struct {
	By string `json:"by,omitempty"`
}

func validateSearchRequest(req searchHTTPRequest) error {
	if len(req.Origin) != 3 {
		return fmt.Errorf("origin must be a 3-letter IATA code")
	}
	if len(req.Destination) != 3 {
		return fmt.Errorf("destination must be a 3-letter IATA code")
	}
	if req.Origin == req.Destination {
		return fmt.Errorf("origin and destination must be different")
	}
	if _, err := time.Parse("2006-01-02", req.DepartureDate); err != nil {
		return fmt.Errorf("departureDate must be in YYYY-MM-DD format")
	}
	if req.Passengers < 1 {
		return fmt.Errorf("passengers must be at least 1")
	}
	if req.CabinClass == "" {
		return fmt.Errorf("cabinClass is required")
	}
	if req.ReturnDate != nil {
		if _, err := time.Parse("2006-01-02", *req.ReturnDate); err != nil {
			return fmt.Errorf("returnDate must be in YYYY-MM-DD format")
		}
		if *req.ReturnDate < req.DepartureDate {
			return fmt.Errorf("returnDate must not be before departureDate")
		}
	}
	return nil
}

func toSearchRequest(req searchHTTPRequest) domain.SearchRequest {
	return domain.SearchRequest{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureDate: req.DepartureDate,
		ReturnDate:    req.ReturnDate,
		Passengers:    req.Passengers,
		CabinClass:    req.CabinClass,
	}
}

func toFilterParams(f *filtersRequest, departureDate string) (domain.FilterParams, error) {
	if f == nil {
		return domain.FilterParams{}, nil
	}

	fp := domain.FilterParams{
		MinPrice:    f.MinPrice,
		MaxPrice:    f.MaxPrice,
		MaxStops:    f.MaxStops,
		MaxDuration: f.MaxDuration,
		Airlines:    f.Airlines,
	}

	wib := time.FixedZone("WIB", 7*3600)

	if f.DepartAfter != nil {
		t, err := parseTimeOnDate(departureDate, *f.DepartAfter, wib)
		if err != nil {
			return fp, fmt.Errorf("invalid departAfter: %w", err)
		}
		fp.EarliestDepart = &t
	}
	if f.DepartBefore != nil {
		t, err := parseTimeOnDate(departureDate, *f.DepartBefore, wib)
		if err != nil {
			return fp, fmt.Errorf("invalid departBefore: %w", err)
		}
		fp.LatestDepart = &t
	}
	if f.ArriveAfter != nil {
		t, err := parseTimeOnDate(departureDate, *f.ArriveAfter, wib)
		if err != nil {
			return fp, fmt.Errorf("invalid arriveAfter: %w", err)
		}
		fp.EarliestArrive = &t
	}
	if f.ArriveBefore != nil {
		t, err := parseTimeOnDate(departureDate, *f.ArriveBefore, wib)
		if err != nil {
			return fp, fmt.Errorf("invalid arriveBefore: %w", err)
		}
		fp.LatestArrive = &t
	}

	return fp, nil
}

func toSortParams(s *sortRequest) domain.SortParams {
	if s == nil {
		return domain.SortParams{By: domain.SortByPriceAsc}
	}
	switch domain.SortBy(s.By) {
	case domain.SortByPriceAsc, domain.SortByPriceDesc,
		domain.SortByDurationAsc, domain.SortByDurationDesc,
		domain.SortByDepartureTime, domain.SortByArrivalTime,
		domain.SortByBestValue:
		return domain.SortParams{By: domain.SortBy(s.By)}
	default:
		return domain.SortParams{By: domain.SortByPriceAsc}
	}
}

func parseTimeOnDate(date, hhmm string, loc *time.Location) (time.Time, error) {
	combined := date + "T" + hhmm + ":00"
	t, err := time.ParseInLocation("2006-01-02T15:04:05", combined, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("time %q on date %q: %w", hhmm, date, err)
	}
	return t, nil
}
