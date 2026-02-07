package scheduler

import (
	"context"
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusRetrying  TaskStatus = "retrying"
)

type TaskFunc func(ctx context.Context) error

type BaseTask struct {
	ID          string
	Priority    int
	Fn          TaskFunc
	Status      TaskStatus
	ScheduledAt time.Time
	Interval    time.Duration
	RetryLimit  int
	RetryCount  int
	LastError   error
}

func NewBaseTask(id string, priority int, fn TaskFunc) *BaseTask {
	return &BaseTask{
		ID:          id,
		Priority:    priority,
		Fn:          fn,
		Status:      StatusPending,
		ScheduledAt: time.Now(),
		RetryLimit:  3,
	}
}

func (t *BaseTask) GetID() string              { return t.ID }
func (t *BaseTask) GetPriority() int           { return t.Priority }
func (t *BaseTask) GetScheduledAt() time.Time  { return t.ScheduledAt }
func (t *BaseTask) GetInterval() time.Duration { return t.Interval }

func (t *BaseTask) Execute(ctx context.Context) error {
	t.Status = StatusRunning
	err := t.Fn(ctx)
	if err != nil {
		t.LastError = err
		return err
	}
	return nil
}

func (t *BaseTask) OnSuccess() {
	t.Status = StatusCompleted
}

func (t *BaseTask) OnFailure(err error) {
	t.Status = StatusFailed
	t.LastError = err
}

func (t *BaseTask) WithDelay(delay time.Duration) *BaseTask {
	t.ScheduledAt = time.Now().Add(delay)
	return t
}

func (t *BaseTask) WithPeriodic(interval time.Duration) *BaseTask {
	t.Interval = interval
	return t
}

func (t *BaseTask) WithRetry(limit int) *BaseTask {
	t.RetryLimit = limit
	return t
}
