package garudaindonesiarepository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/reynerpantou/bookcabin/common/config"
	randomutil "github.com/reynerpantou/bookcabin/common/random"
	timeutil "github.com/reynerpantou/bookcabin/common/time"
	garudaindonesiamodel "github.com/reynerpantou/bookcabin/internal/model/garuda-indonesia"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

type HTTP interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) (garudaindonesiamodel.SearchResponse, error)
}

type httpImpl struct {
	cfg                config.AirlineConfig
	mockSearchResponse garudaindonesiamodel.SearchResponse
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
	}, nil
}

func (h *httpImpl) Search(ctx context.Context, params *requestparamsmodel.RequestParams) (garudaindonesiamodel.SearchResponse, error) {
	if params == nil {
		return garudaindonesiamodel.SearchResponse{}, fmt.Errorf("params is nil")
	}
	ctx, cancel := context.WithTimeout(
		ctx,
		h.cfg.Timeout.Duration(),
	)
	defer cancel()
	err := timeutil.SetRandomDelay(ctx, h.cfg.Mock.MinDelay.Duration(), h.cfg.Mock.MaxDelay.Duration())
	if err != nil {
		return garudaindonesiamodel.SearchResponse{}, err
	}
	if !randomutil.IsSuccess(h.cfg.Mock.SuccessRate) {
		return garudaindonesiamodel.SearchResponse{}, fmt.Errorf("failed to get garuda indonesia search response")
	}
	return h.mockSearchResponse, nil
}
