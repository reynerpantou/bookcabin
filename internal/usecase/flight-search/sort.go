package flightsearchusecase

import (
	"cmp"
	"slices"

	"github.com/reynerpantou/bookcabin/internal/model/flight"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

const (
	priceWeight    = 0.5
	durationWeight = 0.3
	stopsWeight    = 0.2
)

func (u *useCaseImpl) rank(flights []flight.Flight, sortBy, sortOrder string) {
	validate := validateSort(flights, sortBy)
	desc := sortOrder == requestparamsmodel.SortOrderDesc
	slices.SortStableFunc(flights, func(a, b flight.Flight) int {
		c := validate(a, b)
		if desc {
			c = -c
		}
		if c != 0 {
			return c
		}
		return tieBreak(a, b)
	})
}

func validateSort(flights []flight.Flight, sortBy string) func(a, b flight.Flight) int {
	switch sortBy {
	case requestparamsmodel.SortByPrice:
		return func(a, b flight.Flight) int { return cmp.Compare(a.Price.Amount, b.Price.Amount) }
	case requestparamsmodel.SortByDuration:
		return func(a, b flight.Flight) int { return cmp.Compare(a.Duration.TotalMinutes, b.Duration.TotalMinutes) }
	case requestparamsmodel.SortByDepartureTime:
		return func(a, b flight.Flight) int { return cmp.Compare(a.Departure.Timestamp, b.Departure.Timestamp) }
	case requestparamsmodel.SortByArrivalTime:
		return func(a, b flight.Flight) int { return cmp.Compare(a.Arrival.Timestamp, b.Arrival.Timestamp) }
	default: // best value based on price and convenience (duration & stop)
		costs := getBestValueCosts(flights)
		return func(a, b flight.Flight) int { return cmp.Compare(costs[a.ID], costs[b.ID]) }
	}
}

func tieBreak(a, b flight.Flight) int {
	if c := cmp.Compare(a.Price.Amount, b.Price.Amount); c != 0 {
		return c
	}
	if c := cmp.Compare(a.Departure.Timestamp, b.Departure.Timestamp); c != 0 {
		return c
	}
	return cmp.Compare(a.ID, b.ID)
}

func getBestValueCosts(flights []flight.Flight) map[string]float64 {
	costs := make(map[string]float64, len(flights))
	if len(flights) == 0 {
		return costs
	}
	minPrice, maxPrice := flights[0].Price.Amount, flights[0].Price.Amount
	minDur, maxDur := flights[0].Duration.TotalMinutes, flights[0].Duration.TotalMinutes
	maxStops := flights[0].Stops
	for _, f := range flights[1:] {
		minPrice, maxPrice = min(minPrice, f.Price.Amount), max(maxPrice, f.Price.Amount)
		minDur, maxDur = min(minDur, f.Duration.TotalMinutes), max(maxDur, f.Duration.TotalMinutes)
		maxStops = max(maxStops, f.Stops)
	}
	for _, f := range flights {
		costs[f.ID] = priceWeight*normalize(float64(f.Price.Amount), float64(minPrice), float64(maxPrice)) +
			durationWeight*normalize(float64(f.Duration.TotalMinutes), float64(minDur), float64(maxDur)) +
			stopsWeight*normalize(float64(f.Stops), 0, float64(maxStops))
	}
	return costs
}

func normalize(v, lo, hi float64) float64 {
	if hi <= lo {
		return 0
	}
	return (v - lo) / (hi - lo)
}
