package flightsearchusecase

import (
	"context"
	"errors"
	"time"

	"github.com/reynerpantou/bookcabin/common/aviation"
	"github.com/reynerpantou/bookcabin/common/config"
	"github.com/reynerpantou/bookcabin/internal/model/flight"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"github.com/reynerpantou/bookcabin/internal/model/response"
	"github.com/reynerpantou/bookcabin/internal/repository"
)

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrAllProvidersFailed = errors.New("all flight providers failed")
)

type UseCase interface {
	FlightSearch(ctx context.Context, params *requestparamsmodel.RequestParams) (response.FlightSearchResponse, error)
}

type ProviderRepository interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error)
}

type provider struct {
	name       aviation.Provider
	repository ProviderRepository
}

type useCaseImpl struct {
	providers   []provider
	loadTimeout time.Duration
	cache       *cacheImpl
}

func NewUseCase(ctx context.Context, cfg *config.Config, repositories repository.Repositories) (UseCase, error) {
	candidates := []provider{
		{aviation.AirAsiaProvider, repositories.AirAsiaHTTP},
		{aviation.BatikAirProvider, repositories.BatikAirHTTP},
		{aviation.GarudaIndonesiaProvider, repositories.GarudaIndonesiaHTTP},
		{aviation.LionAirProvider, repositories.LionAirHTTP},
	}
	providers := make([]provider, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.repository == nil {
			continue
		}
		providers = append(providers, candidate)
	}
	if len(providers) == 0 {
		return nil, errors.New("no flight providers available")
	}
	cache := newLocalCache(
		ctx,
		cfg.FlightSearch.CacheTTL.Duration(),
		cfg.FlightSearch.CacheCleanupInterval.Duration(),
	)
	return &useCaseImpl{
		providers:   providers,
		loadTimeout: 2 * time.Second,
		cache:       cache,
	}, nil
}
