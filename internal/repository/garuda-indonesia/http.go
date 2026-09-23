package garudaindonesiarepository

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
	garudaindonesiamodel "github.com/reynerpantou/bookcabin/internal/model/garuda-indonesia"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
	"golang.org/x/time/rate"
)

var errProviderUnavailable = errors.New("garuda indonesia: provider unavailable")

type HTTP interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error)
}

type httpImpl struct {
	cfg                config.AirlineConfig
	mockSearchResponse garudaindonesiamodel.SearchResponse
	limiter            *rate.Limiter
}

func NewHTTP(ctx context.Context, cfg config.AirlineConfig) (HTTP, error) {
	bytes, err := os.ReadFile(cfg.Mock.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read garuda indonesia mock file: %w", err)
	}
	var mockSearchResponse garudaindonesiamodel.SearchResponse
	if err = json.Unmarshal(bytes, &mockSearchResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal garuda indonesia mock response: %w", err)
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
			slog.WarnContext(ctx, "garuda indonesia: search attempt failed", "attempt", attempt+1, "error", err)

		}),
	).Do(func() error {
		if err := h.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("garuda indonesia: rate limit: %w", err)
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

func mapToUnifiedFlights(searchResponse garudaindonesiamodel.SearchResponse) ([]flight.Flight, error) {
	if searchResponse.Status != "success" {
		return nil, fmt.Errorf("garuda indonesia: unexpected status %q", searchResponse.Status)
	}
	flights := make([]flight.Flight, 0)
	for _, v := range searchResponse.Flights {
		f, err := mapFlight(v)
		if err != nil {
			slog.Warn("garuda indonesia: skip invalid flight", "flight_id", v.FlightID, "error", err)
			continue
		}
		flights = append(flights, f)
	}
	return flights, nil
}

func mapFlight(v garudaindonesiamodel.Flight) (flight.Flight, error) {
	depAirport, depTime := v.Departure.Airport, v.Departure.Time
	arrAirport, arrTime := v.Arrival.Airport, v.Arrival.Time
	stops := v.Stops
	if n := len(v.Segments); n > 0 {
		first, last := v.Segments[0], v.Segments[n-1]
		depAirport, depTime = first.Departure.Airport, first.Departure.Time
		arrAirport, arrTime = last.Arrival.Airport, last.Arrival.Time
		stops = n - 1
	}
	dep, err := time.Parse(time.RFC3339, depTime)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse departure time: %w", err)
	}
	arr, err := time.Parse(time.RFC3339, arrTime)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse arrival time: %w", err)
	}
	if !arr.After(dep) {
		return flight.Flight{}, fmt.Errorf("arrival %s not after departure %s", arr, dep)
	}
	cabinClass, err := aviation.NormalizeCabinClass(v.FareClass)
	if err != nil {
		return flight.Flight{}, err
	}
	// dari timestamp: termasuk layover, mengabaikan duration_minutes yang tidak konsisten
	totalMinutes := int(arr.Sub(dep).Minutes())
	currentFlight := flight.Flight{
		ID:       v.FlightID + "_" + aviation.GarudaIndonesiaProvider.String(),
		Provider: aviation.GarudaIndonesiaProvider.String(),
		Airline: flight.Airline{
			Name: v.Airline,
			Code: v.AirlineCode,
		},
		FlightNumber: v.FlightID,
		Departure: flight.Airport{
			Airport:   depAirport,
			City:      aviation.GetCityFromAirportCode(depAirport),
			DateTime:  dep.Format(time.RFC3339),
			Timestamp: dep.Unix(),
		},
		Arrival: flight.Airport{
			Airport:   arrAirport,
			City:      aviation.GetCityFromAirportCode(arrAirport),
			DateTime:  arr.Format(time.RFC3339),
			Timestamp: arr.Unix(),
		},
		Duration: flight.Duration{
			TotalMinutes: totalMinutes,
			Formatted:    timeutil.GetFormattedDuration(totalMinutes),
		},
		Stops: stops,
		Price: flight.Price{
			Amount:    v.Price.Amount,
			Currency:  v.Price.Currency,
			Formatted: helper.GetFormattedCurrency(v.Price.Currency, v.Price.Amount),
		},
		AvailableSeats: v.AvailableSeats,
		CabinClass:     cabinClass,
		Aircraft:       helper.StringPtr(v.Aircraft),
		Amenities:      aviation.NormalizeAmenities(v.Amenities),
		Baggage: flight.Baggage{
			CarryOn: formatPieces(v.Baggage.CarryOn),
			Checked: formatPieces(v.Baggage.Checked),
		},
	}
	return currentFlight, nil
}

func formatPieces(n int) string {
	if n == 1 {
		return "1 piece"
	}
	return fmt.Sprintf("%d pieces", n)
}
