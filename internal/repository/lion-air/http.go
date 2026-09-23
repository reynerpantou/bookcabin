package lionairrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/reynerpantou/bookcabin/common/aviation"
	"github.com/reynerpantou/bookcabin/common/config"
	"github.com/reynerpantou/bookcabin/common/helper"
	randomutil "github.com/reynerpantou/bookcabin/common/random"
	timeutil "github.com/reynerpantou/bookcabin/common/time"
	"github.com/reynerpantou/bookcabin/internal/model/flight"
	lionairmodel "github.com/reynerpantou/bookcabin/internal/model/lion-air"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"golang.org/x/time/rate"
)

var errProviderUnavailable = errors.New("lion air: provider unavailable")

const lionTimeLayout = "2006-01-02T15:04:05"

type HTTP interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error)
}

type httpImpl struct {
	cfg                config.AirlineConfig
	mockSearchResponse lionairmodel.SearchResponse
	limiter            *rate.Limiter
}

func NewHTTP(ctx context.Context, cfg config.AirlineConfig) (HTTP, error) {
	bytes, err := os.ReadFile(cfg.Mock.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lion air mock file: %w", err)
	}
	var mockSearchResponse lionairmodel.SearchResponse
	if err = json.Unmarshal(bytes, &mockSearchResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lion air mock response: %w", err)
	}
	return &httpImpl{
		cfg:                cfg,
		mockSearchResponse: mockSearchResponse,
		limiter:            rate.NewLimiter(rate.Limit(cfg.RateLimit.RequestsPerSecond), cfg.RateLimit.Burst),
	}, nil
}

func (h *httpImpl) Search(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error) {
	if params == nil {
		return nil, fmt.Errorf("params is nil")
	}
	var flights []flight.Flight
	err := retry.New(
		retry.Context(ctx),
		retry.Attempts(uint(h.cfg.Retry.MaxAttempts)),
		retry.Delay(h.cfg.Retry.BaseDelay.Duration()),
		retry.MaxJitter(h.cfg.Retry.BaseDelay.Duration()),
		retry.DelayType(retry.CombineDelay(retry.BackOffDelay, retry.RandomDelay)),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, errProviderUnavailable) || errors.Is(err, context.DeadlineExceeded)
		}),
		retry.OnRetry(func(attempt uint, err error) {
			slog.WarnContext(ctx, "lion air: search attempt failed", "attempt", attempt+1, "error", err)

		}),
	).Do(func() error {
		if err := h.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("lion air: rate limit: %w", err)
		}
		var err error
		flights, err = h.searchOnce(ctx, params)
		return err
	})
	return flights, err
}

func (h *httpImpl) searchOnce(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error) {
	ctx, cancel := context.WithTimeout(
		ctx,
		h.cfg.Timeout.Duration(),
	)
	defer cancel()
	err := timeutil.SetRandomDelay(ctx, h.cfg.Mock.MinDelay.Duration(), h.cfg.Mock.MaxDelay.Duration())
	if err != nil {
		return nil, err
	}
	if !randomutil.IsSuccess(h.cfg.Mock.SuccessRate) {
		return nil, errProviderUnavailable
	}
	return mapToUnifiedFlights(h.mockSearchResponse)
}

func mapToUnifiedFlights(searchResponse lionairmodel.SearchResponse) ([]flight.Flight, error) {
	if !searchResponse.Success {
		return nil, fmt.Errorf("lion air: unsuccessful response")
	}
	flights := make([]flight.Flight, 0)
	for _, v := range searchResponse.Data.AvailableFlights {
		f, err := mapFlight(v)
		if err != nil {
			slog.Warn("lion air: skip invalid flight", "id", v.ID, "error", err)
			continue
		}
		flights = append(flights, f)
	}
	return flights, nil
}

func mapFlight(v lionairmodel.Flight) (flight.Flight, error) {
	dep, err := parseWithTimezone(v.Schedule.Departure, v.Schedule.DepartureTimezone)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse departure: %w", err)
	}
	arr, err := parseWithTimezone(v.Schedule.Arrival, v.Schedule.ArrivalTimezone)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse arrival: %w", err)
	}
	if !arr.After(dep) {
		return flight.Flight{}, fmt.Errorf("arrival %s not after departure %s", arr, dep)
	}
	cabinClass, err := aviation.NormalizeCabinClass(v.Pricing.FareType)
	if err != nil {
		return flight.Flight{}, err
	}
	totalMinutes := int(arr.Sub(dep).Minutes())
	currentFlight := flight.Flight{
		ID:       v.ID + "_" + aviation.LionAirProvider.String(),
		Provider: aviation.LionAirProvider.String(),
		Airline: flight.Airline{
			Name: v.Carrier.Name,
			Code: v.Carrier.IATA,
		},
		FlightNumber: v.ID,
		Departure: flight.Airport{
			Airport:   v.Route.From.Code,
			City:      aviation.GetCityFromAirportCode(v.Route.From.Code),
			DateTime:  dep.Format(time.RFC3339),
			Timestamp: dep.Unix(),
		},
		Arrival: flight.Airport{
			Airport:   v.Route.To.Code,
			City:      aviation.GetCityFromAirportCode(v.Route.To.Code),
			DateTime:  arr.Format(time.RFC3339),
			Timestamp: arr.Unix(),
		},
		Duration: flight.Duration{
			TotalMinutes: totalMinutes,
			Formatted:    timeutil.GetFormattedDuration(totalMinutes),
		},
		Stops: len(v.Layovers),
		Price: flight.Price{
			Amount:    v.Pricing.Total,
			Currency:  v.Pricing.Currency,
			Formatted: helper.GetFormattedCurrency(v.Pricing.Currency, v.Pricing.Total),
		},
		AvailableSeats: v.SeatsLeft,
		CabinClass:     cabinClass,
		Aircraft:       helper.StringPtr(v.PlaneType),
		Amenities:      mapAmenities(v.Services),
		Baggage: flight.Baggage{
			CarryOn: v.Services.BaggageAllowance.Cabin,
			Checked: v.Services.BaggageAllowance.Hold,
		},
	}
	return currentFlight, nil
}

func mapAmenities(s lionairmodel.Services) []string {
	amenities := make([]string, 0, 2)
	if s.WiFiAvailable {
		amenities = append(amenities, "wifi")
	}
	if s.MealsIncluded {
		amenities = append(amenities, "meal")
	}
	return amenities
}

func parseWithTimezone(value, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	return time.ParseInLocation(lionTimeLayout, value, loc)
}
