package flight

import (
	"testing"

	"github.com/i-pul/search-flight/internal/domain"
	"github.com/stretchr/testify/assert"
)

func makeFlight(price float64, durationMinutes, stops int) domain.Flight {
	return domain.Flight{
		Price:    domain.Price{Amount: price},
		Duration: domain.Duration{TotalMinutes: durationMinutes},
		Stops:    stops,
	}
}

var defaultW = DefaultScoreWeights()

func TestComputeBestValueScores_Empty(t *testing.T) {
	flights := []domain.Flight{}
	ComputeBestValueScores(flights, defaultW) // must not panic
}

func TestComputeBestValueScores_SingleFlight(t *testing.T) {
	flights := []domain.Flight{makeFlight(1500000, 90, 0)}
	ComputeBestValueScores(flights, defaultW)
	// Only flight: all factors normalise to 0, penalty = 0, score = 100
	assert.Equal(t, float64(100), flights[0].BestValueScore)
}

func TestComputeBestValueScores_BestAndWorstArePolar(t *testing.T) {
	flights := []domain.Flight{
		makeFlight(500000, 60, 0),   // best: cheapest, fastest, direct
		makeFlight(2000000, 180, 2), // worst: most expensive, slowest, most stops
	}
	ComputeBestValueScores(flights, defaultW)

	best := flights[0].BestValueScore
	worst := flights[1].BestValueScore

	assert.Greater(t, best, worst, "cheapest/fastest/direct should score higher")
	assert.Equal(t, float64(100), best)
	assert.Equal(t, float64(0), worst)
}

func TestComputeBestValueScores_PriceWeightDominates(t *testing.T) {
	// Two flights with same stops/duration — cheaper one should score higher
	flights := []domain.Flight{
		makeFlight(800000, 120, 0),
		makeFlight(1600000, 120, 0),
	}
	ComputeBestValueScores(flights, defaultW)
	assert.Greater(t, flights[0].BestValueScore, flights[1].BestValueScore)
}

func TestComputeBestValueScores_ScoresInRange(t *testing.T) {
	flights := []domain.Flight{
		makeFlight(500000, 60, 0),
		makeFlight(1000000, 90, 1),
		makeFlight(1500000, 120, 2),
		makeFlight(2000000, 150, 0),
	}
	ComputeBestValueScores(flights, defaultW)
	for _, f := range flights {
		assert.GreaterOrEqual(t, f.BestValueScore, float64(0))
		assert.LessOrEqual(t, f.BestValueScore, float64(100))
	}
}

func TestComputeBestValueScores_OrderedBestFirst(t *testing.T) {
	flights := []domain.Flight{
		makeFlight(2000000, 180, 2), // worst overall
		makeFlight(500000, 60, 0),   // best overall
		makeFlight(1200000, 120, 1), // middle
	}
	ComputeBestValueScores(flights, defaultW)

	// Verify score ordering: best > middle > worst
	assert.Greater(t, flights[1].BestValueScore, flights[2].BestValueScore)
	assert.Greater(t, flights[2].BestValueScore, flights[0].BestValueScore)
}

func TestComputeBestValueScores_WeightNormalisation(t *testing.T) {
	// Un-normalised {5, 3, 2} must produce the same scores as {0.5, 0.3, 0.2}
	flights1 := []domain.Flight{makeFlight(500000, 60, 0), makeFlight(2000000, 180, 2)}
	flights2 := []domain.Flight{makeFlight(500000, 60, 0), makeFlight(2000000, 180, 2)}

	ComputeBestValueScores(flights1, ScoreWeights{Price: 5, Duration: 3, Stops: 2})
	ComputeBestValueScores(flights2, defaultW)

	assert.Equal(t, flights2[0].BestValueScore, flights1[0].BestValueScore)
	assert.Equal(t, flights2[1].BestValueScore, flights1[1].BestValueScore)
}

func TestComputeBestValueScores_ZeroWeightsFallback(t *testing.T) {
	// All-zero weights must not panic and fall back to defaults
	flights := []domain.Flight{makeFlight(500000, 60, 0)}
	ComputeBestValueScores(flights, ScoreWeights{})
	assert.Equal(t, float64(100), flights[0].BestValueScore)
}

func TestComputeBestValueScores_CustomWeightsPriceOnly(t *testing.T) {
	// With price weight only, the cheaper flight must score higher
	// regardless of duration/stops advantage of the other flight
	flights := []domain.Flight{
		makeFlight(500000, 180, 2),  // cheap but slow with stops
		makeFlight(2000000, 60, 0),  // expensive but fast and direct
	}
	ComputeBestValueScores(flights, ScoreWeights{Price: 1, Duration: 0, Stops: 0})
	assert.Greater(t, flights[0].BestValueScore, flights[1].BestValueScore)
}
