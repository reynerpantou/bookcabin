package batikairrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/reynerpantou/bookcabin/common/config"
	timeutil "github.com/reynerpantou/bookcabin/common/time"
	batikairmodel "github.com/reynerpantou/bookcabin/internal/model/batik-air"
	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

type HTTP interface {
	Search(ctx context.Context, params *requestparamsmodel.RequestParams) (batikairmodel.SearchResponse, error)
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

func (h httpImpl) Search(ctx context.Context, params *requestparamsmodel.RequestParams) (batikairmodel.SearchResponse, error) {
	if params == nil {
		return batikairmodel.SearchResponse{}, fmt.Errorf("params is nil")
	}
	ctx, cancel := context.WithTimeout(
		ctx,
		h.cfg.Timeout.Duration(),
	)
	defer cancel()
	err := timeutil.SetRandomDelay(ctx, h.cfg.Mock.MinDelay.Duration(), h.cfg.Mock.MaxDelay.Duration())
	if err != nil {
		return batikairmodel.SearchResponse{}, err
	}
	return h.mockSearchResponse, nil
}
