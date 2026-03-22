package flight

import (
	"math"

	"github.com/i-pul/search-flight/internal/domain"
)

// ScoreWeights holds the relative importance of each factor in the best-value
// score. Values must be positive; they are automatically normalised so they
// always sum to 1, making it safe to supply any positive combination.
type ScoreWeights struct {
	Price    float64 // lower price is better
	Duration float64 // shorter duration is better
	Stops    float64 // fewer stops is better
}

// DefaultScoreWeights returns the default scoring weights:
// price 50%, duration 30%, stops 20%.
func DefaultScoreWeights() ScoreWeights {
	return ScoreWeights{Price: 0.50, Duration: 0.30, Stops: 0.20}
}

// normalise returns a copy of w where the three weights sum exactly to 1.
// Panics only when all three weights are zero (undefined behaviour).
func (w ScoreWeights) normalise() ScoreWeights {
	total := w.Price + w.Duration + w.Stops
	if total == 0 {
		return DefaultScoreWeights()
	}
	return ScoreWeights{
		Price:    w.Price / total,
		Duration: w.Duration / total,
		Stops:    w.Stops / total,
	}
}

// ComputeBestValueScores assigns a BestValueScore (0–100) to every flight in
// the slice based on a weighted combination of price, duration, and stops.
// Scores are relative to the provided set, so they should be computed after
// filtering with the final result candidate set.
func ComputeBestValueScores(flights []domain.Flight, w ScoreWeights) {
	if len(flights) == 0 {
		return
	}

	w = w.normalise()

	// Find min/max for each factor across the result set.
	minPrice, maxPrice := flights[0].Price.Amount, flights[0].Price.Amount
	minDur, maxDur := flights[0].Duration.TotalMinutes, flights[0].Duration.TotalMinutes
	maxStops := flights[0].Stops

	for _, f := range flights[1:] {
		if f.Price.Amount < minPrice {
			minPrice = f.Price.Amount
		}
		if f.Price.Amount > maxPrice {
			maxPrice = f.Price.Amount
		}
		if f.Duration.TotalMinutes < minDur {
			minDur = f.Duration.TotalMinutes
		}
		if f.Duration.TotalMinutes > maxDur {
			maxDur = f.Duration.TotalMinutes
		}
		if f.Stops > maxStops {
			maxStops = f.Stops
		}
	}

	priceRange := maxPrice - minPrice
	durRange := float64(maxDur - minDur)
	stopsRange := float64(maxStops)

	for i, f := range flights {
		var normPrice, normDur, normStops float64

		if priceRange > 0 {
			normPrice = (f.Price.Amount - minPrice) / priceRange
		}
		if durRange > 0 {
			normDur = float64(f.Duration.TotalMinutes-minDur) / durRange
		}
		if stopsRange > 0 {
			normStops = float64(f.Stops) / stopsRange
		}

		penalty := normPrice*w.Price + normDur*w.Duration + normStops*w.Stops
		// Round to one decimal place for a clean display value.
		flights[i].BestValueScore = math.Round((1-penalty)*1000) / 10
	}
}
