package flightsearchusecase

import (
	"context"

	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"github.com/reynerpantou/bookcabin/internal/model/response"
)

func (u *useCaseImpl) FlightSearch(ctx context.Context, params requestparamsmodel.RequestParams) (response.FlightSearchResponse, error) {
	return response.FlightSearchResponse{}, nil
}
