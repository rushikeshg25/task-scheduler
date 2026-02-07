package scheduler

import (
	"context"
	"time"
)

type ITask interface {
	GetID() string
	GetPriority() int
	GetScheduledAt() time.Time
	GetInterval() time.Duration
	Execute(ctx context.Context) error
	OnFailure(err error)
	OnSuccess()
}

type TaskQueue interface {
	Push(task ITask)
	Pop() ITask
	Peak() ITask
	Len() int
}

type WorkerStats struct {
	ID            int
	Status        string
	CurrentTaskID string
}

type SchedulerStats struct {
	QueueSize      int
	WorkerCount    int
	Workers        []WorkerStats
	TasksCompleted int64
	TasksFailed    int64
}

type WorkerPool interface {
	Start(ctx context.Context)
	Stop()
	Submit(task ITask)
	SetOnComplete(onComplete func(ITask))
	GetStats() []WorkerStats
	GetCounters() (completed, failed int64)
}

type Dispatcher interface {
	Start()
	Stop()
}
