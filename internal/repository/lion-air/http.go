package lionairrepository

import (
	"context"

	"github.com/reynerpantou/bookcabin/common/config"
)

type HTTP interface {
	Search(context.Context)
}

type httpImpl struct {
	cfg config.AirlineConfig
}

func NewHTTP(ctx context.Context, cfg config.AirlineConfig) (HTTP, error) {
	return &httpImpl{
		cfg: cfg,
	}, nil
}

func (h httpImpl) Search(ctx context.Context) {
	if !h.cfg.Enabled {
		return
	}
}
