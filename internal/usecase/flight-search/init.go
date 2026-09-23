package flightsearchusecase

import (
	"context"

	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"github.com/reynerpantou/bookcabin/internal/model/response"
	"github.com/reynerpantou/bookcabin/internal/repository"
	airasiarepository "github.com/reynerpantou/bookcabin/internal/repository/airasia"
	batikairrepository "github.com/reynerpantou/bookcabin/internal/repository/batik-air"
	garudaindonesiarepository "github.com/reynerpantou/bookcabin/internal/repository/garuda-indonesia"
	lionairrepository "github.com/reynerpantou/bookcabin/internal/repository/lion-air"
)

type UseCase interface {
	FlightSearch(ctx context.Context, params *requestparamsmodel.RequestParams) (response.FlightSearchResponse, error)
}

type useCaseImpl struct {
	airAsiaHTTP         airasiarepository.HTTP
	batikAirHTTP        batikairrepository.HTTP
	garudaIndonesiaHTTP garudaindonesiarepository.HTTP
	lionAirHTTP         lionairrepository.HTTP
}

func NewUseCase(ctx context.Context, repositories repository.Repositories) (UseCase, error) {
	return &useCaseImpl{
		airAsiaHTTP:         repositories.AirAsiaHTTP,
		batikAirHTTP:        repositories.BatikAirHTTP,
		garudaIndonesiaHTTP: repositories.GarudaIndonesiaHTTP,
		lionAirHTTP:         repositories.LionAirHTTP,
	}, nil
}
