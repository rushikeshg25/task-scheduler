package scheduler

import (
	"context"
	"testing"
	"time"
)

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()

	now := time.Now()

	t1 := NewBaseTask("t1", 10, func(ctx context.Context) error { return nil })
	t1.ScheduledAt = now.Add(100 * time.Millisecond)

	t2 := NewBaseTask("t2", 1, func(ctx context.Context) error { return nil })
	t2.ScheduledAt = now.Add(10 * time.Millisecond)

	t3 := NewBaseTask("t3", 20, func(ctx context.Context) error { return nil })
	t3.ScheduledAt = now.Add(100 * time.Millisecond)

	pq.Push(t1)
	pq.Push(t2)
	pq.Push(t3)

	res := pq.Pop()
	if res.GetID() != "t2" {
		t.Errorf("Expected t2, got %s", res.GetID())
	}

	res = pq.Pop()
	if res.GetID() != "t3" {
		t.Errorf("Expected t3, got %s", res.GetID())
	}

	res = pq.Pop()
	if res.GetID() != "t1" {
		t.Errorf("Expected t1, got %s", res.GetID())
	}
}
