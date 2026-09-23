package batikairrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/reynerpantou/bookcabin/common/aviation"
	"github.com/reynerpantou/bookcabin/common/config"
	"github.com/reynerpantou/bookcabin/common/helper"
	randomutil "github.com/reynerpantou/bookcabin/common/random"
	timeutil "github.com/reynerpantou/bookcabin/common/time"
	batikairmodel "github.com/reynerpantou/bookcabin/internal/model/batik-air"
	"github.com/reynerpantou/bookcabin/internal/model/flight"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

const batikTimeLayout = "2006-01-02T15:04:05-0700"

var baggageRe = regexp.MustCompile(`(?i)(\d+)\s*kg\s+cabin\s*,\s*(\d+)\s*kg\s+checked`)

type HTTP interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) ([]flight.Flight, error)
}

type httpImpl struct {
	cfg                config.AirlineConfig
	mockSearchResponse batikairmodel.SearchResponse
}

func NewHTTP(ctx context.Context, cfg config.AirlineConfig) (HTTP, error) {
	bytes, err := os.ReadFile(cfg.Mock.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read batik air mock file: %w", err)
	}
	var mockSearchresponse batikairmodel.SearchResponse
	if err = json.Unmarshal(bytes, &mockSearchresponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batik air mock response: %w", err)
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
		return nil, fmt.Errorf("failed to get batik air search response")
	}
	return mapToUnifiedFlights(h.mockSearchResponse)
}

func mapToUnifiedFlights(searchResponse batikairmodel.SearchResponse) ([]flight.Flight, error) {
	if searchResponse.Code != http.StatusOK {
		return nil, fmt.Errorf("batik air: unexpected status %d", searchResponse.Code)
	}
	flights := make([]flight.Flight, 0)
	for _, v := range searchResponse.Results {
		f, err := mapFlight(v)
		if err != nil {
			slog.Warn("batik air: skip invalid flight", "flight_number", v.FlightNumber, "error", err)
			continue
		}
		flights = append(flights, f)
	}
	return flights, nil
}

func mapFlight(v batikairmodel.Flight) (flight.Flight, error) {
	dep, err := time.Parse(batikTimeLayout, v.DepartureDateTime)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse depart_time: %w", err)
	}
	arr, err := time.Parse(batikTimeLayout, v.ArrivalDateTime)
	if err != nil {
		return flight.Flight{}, fmt.Errorf("parse arrive_time: %w", err)
	}
	if !arr.After(dep) {
		return flight.Flight{}, fmt.Errorf("arrival %s not after departure %s", arr, dep)
	}
	cabinClass, err := aviation.NormalizeCabinClass(v.Fare.Class)
	if err != nil {
		return flight.Flight{}, err
	}
	totalMinutes := int(arr.Sub(dep).Minutes())
	currentFlight := flight.Flight{
		ID:       v.FlightNumber + "_" + aviation.BatikAirProvider.String(),
		Provider: aviation.BatikAirProvider.String(),
		Airline: flight.Airline{
			Name: v.AirlineName,
			Code: v.AirlineIATA,
		},
		FlightNumber: v.FlightNumber,
		Departure: flight.Airport{
			Airport:   v.Origin,
			City:      aviation.GetCityFromAirportCode(v.Origin),
			DateTime:  dep.Format(time.RFC3339),
			Timestamp: dep.Unix(),
		},
		Arrival: flight.Airport{
			Airport:   v.Destination,
			City:      aviation.GetCityFromAirportCode(v.Destination),
			DateTime:  arr.Format(time.RFC3339),
			Timestamp: arr.Unix(),
		},
		Duration: flight.Duration{
			TotalMinutes: totalMinutes,
			Formatted:    timeutil.GetFormattedDuration(totalMinutes),
		},
		Stops: v.NumberOfStops,
		Price: flight.Price{
			Amount:   v.Fare.TotalPrice,
			Currency: v.Fare.CurrencyCode,
		},
		AvailableSeats: v.SeatsAvailable,
		CabinClass:     cabinClass,
		Aircraft:       helper.StringPtr(v.AircraftModel),
		Amenities:      aviation.NormalizeAmenities(v.OnboardServices),
		Baggage:        parseBaggage(v.BaggageInfo),
	}
	return currentFlight, nil
}

// 7kg cabin, 20kg checked -> carry_on: 7 kg, checked: 20 kg
func parseBaggage(info string) flight.Baggage {
	m := baggageRe.FindStringSubmatch(info)
	if m == nil {
		return flight.Baggage{CarryOn: strings.TrimSpace(info)}
	}
	return flight.Baggage{CarryOn: m[1] + " kg", Checked: m[2] + " kg"}
}
