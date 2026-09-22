package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/reynerpantou/bookcabin/common/config"
	airasiarepository "github.com/reynerpantou/bookcabin/internal/repository/airasia"
	batikairrepository "github.com/reynerpantou/bookcabin/internal/repository/batik-air"
	garudaindonesiarepository "github.com/reynerpantou/bookcabin/internal/repository/garuda-indonesia"
	lionairrepository "github.com/reynerpantou/bookcabin/internal/repository/lion-air"
	"github.com/reynerpantou/bookcabin/pkg/safewaitgroup"
)

type Repositories struct {
	// airasia
	AirAsiaHTTP airasiarepository.HTTP
	// batik air
	BatikAirHTTP batikairrepository.HTTP
	// garuda indonesia
	GarudaIndonesiaHTTP garudaindonesiarepository.HTTP
	// lion air
	LionAirHTTP lionairrepository.HTTP
	// misc
}

func NewRepositories(ctx context.Context, cfg *config.Config) (Repositories, error) {
	repositories := Repositories{}
	swg := safewaitgroup.NewSafeWaitGroup()
	swg.Go("initalize_repository_airasia_http", func() {
		if !cfg.Airlines.AirAsia.Enabled {
			return
		}
		airAsiaHTTP, err := airasiarepository.NewHTTP(ctx, cfg.Airlines.AirAsia)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to initialize airline repository",
				"provider", "airasia-http",
				"error", err,
			)
			return
		}
		repositories.AirAsiaHTTP = airAsiaHTTP
	})
	swg.Go("initialize_repository_batik_air_http", func() {
		if !cfg.Airlines.BatikAir.Enabled {
			return
		}
		batikAirHTTP, err := batikairrepository.NewHTTP(ctx, cfg.Airlines.BatikAir)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to initialize airline repository",
				"provider", "batik-air-http",
				"error", err,
			)
			return
		}
		repositories.BatikAirHTTP = batikAirHTTP
	})
	swg.Go("initialize_repository_garuda_indonesia_http", func() {
		if !cfg.Airlines.GarudaIndonesia.Enabled {
			return
		}
		garudaIndonesiaHTTP, err := garudaindonesiarepository.NewHTTP(ctx, cfg.Airlines.GarudaIndonesia)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to initialize airline repository",
				"provider", "garuda-indonesia-http",
				"error", err,
			)
			return
		}
		repositories.GarudaIndonesiaHTTP = garudaIndonesiaHTTP
	})
	swg.Go("initialize_repository_lion_air_http", func() {
		if !cfg.Airlines.LionAir.Enabled {
			return
		}
		lionAirHTTP, err := lionairrepository.NewHTTP(ctx, cfg.Airlines.LionAir)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to initialize airline repository",
				"provider", "lion-air-http",
				"error", err,
			)
			return
		}
		repositories.LionAirHTTP = lionAirHTTP
	})
	swg.Wait()
	if (repositories == Repositories{}) {
		return Repositories{}, fmt.Errorf(
			"no airline repositories available",
		)
	}
	return repositories, nil
}
