package airasiarepository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/reynerpantou/bookcabin/common/aviation"
	"github.com/reynerpantou/bookcabin/common/config"
	randomutil "github.com/reynerpantou/bookcabin/common/random"
	timeutil "github.com/reynerpantou/bookcabin/common/time"
	airasiamodel "github.com/reynerpantou/bookcabin/internal/model/airasia"
	"github.com/reynerpantou/bookcabin/internal/model/flight"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

type HTTP interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error)
}

type httpImpl struct {
	cfg                config.AirlineConfig
	mockSearchResponse airasiamodel.SearchResponse
}

func NewHTTP(ctx context.Context, cfg config.AirlineConfig) (HTTP, error) {
	bytes, err := os.ReadFile(cfg.Mock.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read airasia mock file: %w", err)
	}
	var mockSearchresponse airasiamodel.SearchResponse
	if err = json.Unmarshal(bytes, &mockSearchresponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal airasia mock response: %w", err)
	}
	return &httpImpl{
		cfg:                cfg,
		mockSearchResponse: mockSearchresponse,
	}, nil
}

func (h *httpImpl) Search(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error) {
	if params == nil {
		return nil, fmt.Errorf("params is nil")
	}
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
		return nil, fmt.Errorf("failed to get airasia search response")
	}
	return mapToUnifiedFlights(h.mockSearchResponse)
}

func mapToUnifiedFlights(searchResponse airasiamodel.SearchResponse) ([]flight.Flight, error) {
	if searchResponse.Status != "ok" {
		return nil, fmt.Errorf("airasia: unexpected status %q", searchResponse.Status)
	}
	flights := make([]flight.Flight, 0)
	for _, v := range searchResponse.Flights {
		f, err := mapFlight(v)
		if err != nil {
			slog.Warn("airasia: skip invalid flight", "flight_number", v.FlightCode, "error", err)
			continue
		}
		flights = append(flights, f)
	}
	return flights, nil
}

func mapFlight(v airasiamodel.Flight) (flight.Flight, error) {
	dep, err := time.Parse(time.RFC3339, v.DepartTime)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse depart_time: %w", err)
	}
	arr, err := time.Parse(time.RFC3339, v.ArriveTime)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse arrive_time: %w", err)
	}
	if !arr.After(dep) {
		return flight.Flight{}, fmt.Errorf("arrival %s not after departure %s", arr, dep)
	}
	cabinClass, err := aviation.NormalizeCabinClass(v.CabinClass)
	if err != nil {
		return flight.Flight{}, err
	}
	totalMinutes := int(arr.Sub(dep).Minutes())
	currentFlight := flight.Flight{
		ID:       v.FlightCode + "_" + aviation.AirAsiaProvider.String(),
		Provider: aviation.AirAsiaProvider.String(),
		Airline: flight.Airline{
			Name: aviation.AirAsiaProvider.String(),
			Code: aviation.GetAirlineIATACodeFromFlightCode(v.FlightCode),
		},
		FlightNumber: v.FlightCode,
		Departure: flight.Airport{
			Airport:   v.FromAirport,
			City:      aviation.GetCityFromAirportCode(v.FromAirport),
			DateTime:  dep.Format(time.RFC3339),
			Timestamp: dep.Unix(),
		},
		Arrival: flight.Airport{
			Airport:   v.ToAirport,
			City:      aviation.GetCityFromAirportCode(v.ToAirport),
			DateTime:  arr.Format(time.RFC3339),
			Timestamp: arr.Unix(),
		},
		Duration: flight.Duration{
			TotalMinutes: totalMinutes,
			Formatted:    timeutil.GetFormattedDuration(totalMinutes),
		},
		Price: flight.Price{
			Amount:   v.PriceIDR,
			Currency: "IDR",
		},
		AvailableSeats: v.Seats,
		CabinClass:     cabinClass,
		Aircraft:       nil,               // airasia doesn't have aircraft data
		Amenities:      make([]string, 0), // airasia doesn't have amenities data
		Baggage:        parseBaggage(v.BaggageNote),
	}
	stops := 0
	if !v.DirectFlight {
		stops = len(v.Stops)
	}
	currentFlight.Stops = stops
	return currentFlight, nil
}

func parseBaggage(note string) flight.Baggage {
	cabin, checked, found := strings.Cut(note, ",")
	if !found {
		return flight.Baggage{CarryOn: strings.TrimSpace(note)}
	}
	b := flight.Baggage{
		CarryOn: strings.TrimSpace(cabin),
		Checked: strings.TrimSpace(checked),
	}
	if strings.Contains(strings.ToLower(b.Checked), "additional fee") {
		b.Checked = "Additional fee"
	}
	return b
}
