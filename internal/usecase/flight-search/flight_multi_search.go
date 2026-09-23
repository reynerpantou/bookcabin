package flightsearchusecase

import (
	"context"
	"errors"
	"fmt"

	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"github.com/reynerpantou/bookcabin/internal/model/response"
	"github.com/reynerpantou/bookcabin/pkg/safewaitgroup"
)

func (u *useCaseImpl) FlightMultiSearch(ctx context.Context, params *requestparamsmodel.MultiRequestParams) (response.FlightMultiSearchResponse, error) {
	if params == nil || len(params.Legs) == 0 {
		return response.FlightMultiSearchResponse{}, fmt.Errorf("%w: at least one leg is required", ErrInvalidRequest)
	}
	results := make([]response.FlightSearchResponse, len(params.Legs))
	errs := make([]error, len(params.Legs))
	swg := safewaitgroup.NewSafeWaitGroup()
	for i := range params.Legs {
		errs[i] = errors.New("leg search did not complete")
		swg.Go(fmt.Sprintf("multi_search_leg_%d", i+1), func() {
			results[i], errs[i] = u.FlightSearch(ctx, &params.Legs[i])
		})
	}
	swg.Wait()
	for i, err := range errs {
		if err != nil {
			return response.FlightMultiSearchResponse{}, fmt.Errorf("leg %d: %w", i+1, err)
		}
	}
	return response.FlightMultiSearchResponse{Legs: results}, nil
}
