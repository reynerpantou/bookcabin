package flightsearchusecase

import (
	"context"
	"strings"
	"sync"
	"time"

	requestparamsmodel "github.com/reynerpantou/bookcabin/internal/model/request-params"
)

type cacheInfo struct {
	value     loadResponse
	expiresAt time.Time
}

type cacheImpl struct {
	mutex          sync.RWMutex
	keyToCacheInfo map[string]cacheInfo
	ttl            time.Duration
}

func newLocalCache(ctx context.Context, ttl, cleanupInterval time.Duration) *cacheImpl {
	cache := &cacheImpl{
		keyToCacheInfo: make(map[string]cacheInfo),
		ttl:            ttl,
	}
	go cache.free(ctx, cleanupInterval)
	return cache
}

func (c *cacheImpl) get(key string) (loadResponse, bool) {
	c.mutex.RLock()
	entry, found := c.keyToCacheInfo[key]
	c.mutex.RUnlock()
	if !found || time.Now().After(entry.expiresAt) {
		return loadResponse{}, false
	}
	return entry.value, true
}

func (c *cacheImpl) set(key string, value loadResponse) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.keyToCacheInfo[key] = cacheInfo{value: value, expiresAt: time.Now().Add(c.ttl)}
}

func (c *cacheImpl) free(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			c.mutex.Lock()
			for key, entry := range c.keyToCacheInfo {
				if now.After(entry.expiresAt) {
					delete(c.keyToCacheInfo, key)
				}
			}
			c.mutex.Unlock()
		}
	}
}

func getCacheKey(params *requestparamsmodel.RequestParams) string {
	return strings.Join([]string{params.Origin, params.Destination, params.DepartureDate, params.CabinClass}, "|")
}
