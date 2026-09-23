package flightsearchusecase

import (
	"time"

	"github.com/reynerpantou/bookcabin/internal/model/flight"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

func (u *useCaseImpl) applyFilter(flights []flight.Flight, params *requestparamsmodel.RequestParams) []flight.Flight {
	airlines := make(map[string]struct{}, len(params.Airlines))
	for _, code := range params.Airlines {
		airlines[code] = struct{}{}
	}
	result := make([]flight.Flight, 0, len(flights))
	for _, f := range flights {
		if validateFilter(f, params, airlines) {
			result = append(result, f)
		}
	}
	return result

}

func validateFilter(f flight.Flight, params *requestparamsmodel.RequestParams, airlines map[string]struct{}) bool {
	filter := params.Filters
	switch {
	case f.AvailableSeats < params.Passengers:
		return false
	case filter.MinPrice != nil && f.Price.Amount < *filter.MinPrice:
		return false
	case filter.MaxPrice != nil && f.Price.Amount > *filter.MaxPrice:
		return false
	case filter.MaxStops != nil && f.Stops > *filter.MaxStops:
		return false
	case filter.MaxDurationMinutes != nil && f.Duration.TotalMinutes > *filter.MaxDurationMinutes:
		return false
	}
	if len(airlines) > 0 {
		if _, ok := airlines[f.Airline.Code]; !ok {
			return false
		}
	}
	return inTimeWindow(f.Departure.DateTime, filter.DepartureFrom, filter.DepartureTo) &&
		inTimeWindow(f.Arrival.DateTime, filter.ArrivalFrom, filter.ArrivalTo)
}

func inTimeWindow(datetime, from, to string) bool {
	if from == "" && to == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return false
	}
	clock := t.Format("15:04")
	return (from == "" || clock >= from) && (to == "" || clock <= to)
}
