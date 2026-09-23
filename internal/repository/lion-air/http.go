package lionairrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/reynerpantou/bookcabin/common/config"
	randomutil "github.com/reynerpantou/bookcabin/common/random"
	timeutil "github.com/reynerpantou/bookcabin/common/time"
	lionairmodel "github.com/reynerpantou/bookcabin/internal/model/lion-air"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

type HTTP interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) (lionairmodel.SearchResponse, error)
}

type httpImpl struct {
	cfg                config.AirlineConfig
	mockSearchResponse lionairmodel.SearchResponse
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
	}, nil
}

func (h *httpImpl) Search(ctx context.Context, params *requestparamsmodel.RequestParams) (lionairmodel.SearchResponse, error) {
	if params == nil {
		return lionairmodel.SearchResponse{}, fmt.Errorf("params is nil")
	}
	ctx, cancel := context.WithTimeout(
		ctx,
		h.cfg.Timeout.Duration(),
	)
	defer cancel()
	err := timeutil.SetRandomDelay(ctx, h.cfg.Mock.MinDelay.Duration(), h.cfg.Mock.MaxDelay.Duration())
	if err != nil {
		return lionairmodel.SearchResponse{}, err
	}
	if !randomutil.IsSuccess(h.cfg.Mock.SuccessRate) {
		return lionairmodel.SearchResponse{}, fmt.Errorf("failed to get lion air search response")
	}
	return h.mockSearchResponse, nil
}
