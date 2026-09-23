package timeutil

import (
	"context"
	"math/rand/v2"
	"time"
)

func SetDelay(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func SetRandomDelay(ctx context.Context, min time.Duration, max time.Duration) error {
	return SetDelay(ctx, randomDuration(min, max))
}

func randomDuration(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	return min + time.Duration(rand.Int64N(int64(max-min)+1))
}
