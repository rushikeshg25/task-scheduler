package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCancelRunningAndBlockedDispatch(t *testing.T) {
	s := NewScheduler(1)
	started := make(chan struct{})
	s.Schedule(NewBaseTask("running", 1, func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		s.Schedule(NewBaseTask("late", 1, func(context.Context) error { return nil }))
		return ctx.Err()
	}))
	s.Schedule(NewBaseTask("queued", 1, func(context.Context) error { return nil }))
	s.Start()
	<-started
	done := make(chan struct{})
	go func() { s.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stop deadlocked")
	}
	s.Stop()
	s.Start()
	if s.Schedule(NewBaseTask("after", 1, func(context.Context) error { return nil })) {
		t.Fatal("accepted after shutdown")
	}
}
func TestStopBeforeStartAndNil(t *testing.T) {
	s := NewScheduler(0)
	var task *BaseTask
	if s.Schedule(task) {
		t.Fatal("typed nil")
	}
	s.Stop()
	s.Start()
	s.Stop()
}
func TestStableQueue(t *testing.T) {
	q := NewPriorityQueue()
	now := time.Now()
	for _, id := range []string{"a", "b", "c"} {
		task := NewBaseTask(id, 1, nil)
		task.ScheduledAt = now
		q.Push(task)
	}
	for _, id := range []string{"a", "b", "c"} {
		if q.PopReady(now).GetID() != id {
			t.Fatal("unstable queue")
		}
	}
}
func TestPanicDoesNotLoseWorker(t *testing.T) {
	s := NewScheduler(1)
	defer s.Stop()
	done := make(chan struct{})
	s.Schedule(NewBaseTask("panic", 1, func(context.Context) error { panic("boom") }).WithRetry(0))
	s.Schedule(NewBaseTask("next", 1, func(context.Context) error { close(done); return nil }))
	s.Start()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("lost worker")
	}
}
func TestConcurrentStartStop(t *testing.T) {
	s := NewScheduler(1)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); s.Start(); s.Stop() }()
	}
	wg.Wait()
}
