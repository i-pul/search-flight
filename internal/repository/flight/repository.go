package flight

import (
	"context"

	"github.com/i-pul/search-flight/internal/domain"
)

// Repository is the contract every airline data source must satisfy.
type Repository interface {
	Name() string
	Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error)
}
