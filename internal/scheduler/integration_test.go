package scheduler

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSchedulerIntegration(t *testing.T) {
	s := NewScheduler(2)
	s.Start()
	defer s.Stop()

	var wg sync.WaitGroup
	wg.Add(3)

	executed := make(map[string]bool)
	var mu sync.Mutex

	s.Schedule(NewBaseTask("immediate", 1, func(ctx context.Context) error {
		mu.Lock()
		executed["immediate"] = true
		mu.Unlock()
		wg.Done()
		return nil
	}))

	s.Schedule(NewBaseTask("delayed", 1, func(ctx context.Context) error {
		mu.Lock()
		executed["delayed"] = true
		mu.Unlock()
		wg.Done()
		return nil
	}).WithDelay(100 * time.Millisecond))

	s.Schedule(NewBaseTask("periodic", 1, func(ctx context.Context) error {
		mu.Lock()
		if !executed["periodic"] {
			executed["periodic"] = true
			wg.Done()
		}
		mu.Unlock()
		return nil
	}).WithPeriodic(50 * time.Millisecond))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for tasks to complete")
	}

	mu.Lock()
	defer mu.Unlock()
	if !executed["immediate"] || !executed["delayed"] || !executed["periodic"] {
		t.Errorf("Not all tasks were executed: %v", executed)
	}
}

func TestRetryLogic(t *testing.T) {
	s := NewScheduler(1)
	s.Start()
	defer s.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	retryCount := 0
	var mu sync.Mutex

	s.Schedule(NewBaseTask("retry-task", 1, func(ctx context.Context) error {
		mu.Lock()
		retryCount++
		if retryCount == 3 {
			mu.Unlock()
			wg.Done()
			return nil
		}
		mu.Unlock()
		return fmt.Errorf("temporary failure")
	}).WithRetry(3))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for retry task to complete")
	}

	mu.Lock()
	defer mu.Unlock()
	if retryCount != 3 {
		t.Errorf("Expected 3 executions, got %d", retryCount)
	}
}
