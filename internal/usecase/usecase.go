package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/reynerpantou/bookcabin/common/config"
	"github.com/reynerpantou/bookcabin/internal/repository"
	flightsearchusecase "github.com/reynerpantou/bookcabin/internal/usecase/flight-search"
	"github.com/reynerpantou/bookcabin/pkg/safewaitgroup"
)

type UseCases struct {
	FlightSearch flightsearchusecase.UseCase
}

func NewUseCases(ctx context.Context, cfg *config.Config, repositories repository.Repositories) (UseCases, error) {
	useCases := UseCases{}
	swg := safewaitgroup.NewSafeWaitGroup()
	swg.Go("initalize_usecase_flight_search", func() {
		flightSearchUseCase, err := flightsearchusecase.NewUseCase(ctx, cfg, repositories)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to initialize flight search usecase",
				"error", err,
			)
			return
		}
		useCases.FlightSearch = flightSearchUseCase
	})
	swg.Wait()
	if (useCases == UseCases{}) {
		return UseCases{}, fmt.Errorf(
			"no usecases available",
		)
	}
	return useCases, nil
}
