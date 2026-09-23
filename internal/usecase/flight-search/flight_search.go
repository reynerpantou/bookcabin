package flightsearchusecase

import (
	"context"
	"fmt"
	"time"

	"github.com/reynerpantou/bookcabin/internal/model/flight"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"github.com/reynerpantou/bookcabin/internal/model/response"
)

func (u *useCaseImpl) FlightSearch(ctx context.Context, params *requestparamsmodel.RequestParams) (response.FlightSearchResponse, error) {
	start := time.Now()
	if params.Origin == params.Destination {
		return response.FlightSearchResponse{}, fmt.Errorf("%w: origin and destination must differ", ErrInvalidRequest)
	}
	loadResp, err := u.load(ctx, params)
	if err != nil {
		return response.FlightSearchResponse{}, err
	}
	flights := u.applyFilter(loadResp.Flights, params)
	u.rank(flights, params.SortBy, params.SortOrder)
	return buildResponse(params, loadResp, flights, time.Since(start)), nil
}

func buildResponse(params *requestparamsmodel.RequestParams, loadResp loadResponse, flights []flight.Flight, endTime time.Duration) response.FlightSearchResponse {
	return response.FlightSearchResponse{
		SearchCriteria: response.SearchCriteria{
			Origin:        params.Origin,
			Destination:   params.Destination,
			DepartureDate: params.DepartureDate,
			Passengers:    params.Passengers,
			CabinClass:    params.CabinClass,
		},
		Metadata: response.Metadata{
			TotalResults:       len(flights),
			ProvidersQueried:   loadResp.ProvidersQueried,
			ProvidersSucceeded: loadResp.ProvidersSucceeded,
			ProvidersFailed:    loadResp.ProvidersFailed,
			SearchTimeMS:       endTime.Milliseconds(),
			CacheHit:           loadResp.CacheHit,
		},
		Flights: flights,
	}
}
