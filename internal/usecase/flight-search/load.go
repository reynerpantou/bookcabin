package flightsearchusecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/reynerpantou/bookcabin/internal/model/flight"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"github.com/reynerpantou/bookcabin/pkg/safewaitgroup"
)

type loadResponse struct {
	Flights            []flight.Flight
	ProvidersQueried   int
	ProvidersSucceeded int
	ProvidersFailed    int
	CacheHit           bool
}

func (u *useCaseImpl) load(ctx context.Context, params *requestparamsmodel.RequestParams) (loadResponse, error) {
	key := getCacheKey(params)
	if cached, found := u.cache.get(key); found {
		cached.CacheHit = true
		return cached, nil
	}
	res := u.fetchAll(ctx, params)
	if res.ProvidersSucceeded == 0 {
		return loadResponse{}, ErrAllProvidersFailed
	}
	res.Flights = u.applySearch(res.Flights, params)
	res.Flights = u.deduplicateCheapest(res.Flights)
	if res.ProvidersFailed == 0 {
		u.cache.set(key, res)
	}
	return res, nil
}

func (u *useCaseImpl) fetchAll(ctx context.Context, params *requestparamsmodel.RequestParams) loadResponse {
	ctx, cancel := context.WithTimeout(ctx, u.loadTimeout)
	defer cancel()
	results := make([][]flight.Flight, len(u.providers))
	errs := make([]error, len(u.providers))
	swg := safewaitgroup.NewSafeWaitGroup()
	for i, provider := range u.providers {
		errs[i] = errors.New("provider did not complete")
		swg.Go("load_provider_"+provider.name.String(), func() {
			results[i], errs[i] = provider.repository.Search(ctx, params)
		})
	}
	swg.Wait()
	res := loadResponse{
		ProvidersQueried: len(u.providers),
	}
	for i, p := range u.providers {
		if errs[i] != nil {
			res.ProvidersFailed++
			slog.Warn("provider search failed", "provider", p.name, "error", errs[i])
			continue
		}
		res.ProvidersSucceeded++
		res.Flights = append(res.Flights, results[i]...)
	}
	return res
}

func (u *useCaseImpl) applySearch(flights []flight.Flight, params *requestparamsmodel.RequestParams) []flight.Flight {
	result := make([]flight.Flight, 0, len(flights))
	for _, f := range flights {
		if validateSearch(f, params) {
			result = append(result, f)
		}
	}
	return result
}

func validateSearch(f flight.Flight, params *requestparamsmodel.RequestParams) bool {
	if params == nil {
		return false
	}
	if f.Departure.Airport != params.Origin ||
		f.Arrival.Airport != params.Destination ||
		f.CabinClass != params.CabinClass {
		return false
	}
	dep, err := time.Parse(time.RFC3339, f.Departure.DateTime)
	if err != nil {
		return false
	}
	return dep.Format(time.DateOnly) == params.DepartureDate
}

func (u *useCaseImpl) deduplicateCheapest(flights []flight.Flight) []flight.Flight {
	indexByKey := make(map[string]int, len(flights))
	result := make([]flight.Flight, 0, len(flights))
	for _, f := range flights {
		key := fmt.Sprintf("%s_%d", f.FlightNumber, f.Departure.Timestamp)
		if idx, ok := indexByKey[key]; ok {
			if f.Price.Amount < result[idx].Price.Amount {
				result[idx] = f
			}
			continue
		}
		indexByKey[key] = len(result)
		result = append(result, f)
	}
	return result
}
