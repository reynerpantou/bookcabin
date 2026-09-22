package safewaitgroup

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

type SafeWaitGroup struct {
	wg sync.WaitGroup
}

func NewSafeWaitGroup() *SafeWaitGroup {
	return &SafeWaitGroup{}
}

func (s *SafeWaitGroup) Go(processName string, f func()) {
	s.wg.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[%s] goroutine panic: %v\n%s", processName, r, debug.Stack())
			}
		}()
		f()
	})
}

func (s *SafeWaitGroup) Wait() {
	s.wg.Wait()
}

func (s *SafeWaitGroup) WaitTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return nil
	case <-timer.C:
		return context.DeadlineExceeded
	}
}
