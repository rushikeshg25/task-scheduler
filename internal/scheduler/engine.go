package scheduler

import (
	"context"
	"log"
)

type Scheduler struct {
	queue      TaskQueue
	pool       WorkerPool
	dispatcher Dispatcher
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewScheduler(workerCount int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	q := NewPriorityQueue()
	p := NewWorkerPool(workerCount)

	p.SetOnComplete(func(t ITask) {
		HandleTaskLifecycle(q, t)
	})

	d := NewDispatcher(q, p)

	return &Scheduler{
		queue:      q,
		pool:       p,
		dispatcher: d,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Scheduler) Start() {
	s.pool.Start(s.ctx)
	s.dispatcher.Start()
	log.Println("Scheduler started (Modular)")
}

func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	s.dispatcher.Stop()
	s.pool.Stop()
	s.cancel()
	log.Println("Scheduler stopped.")
}

func (s *Scheduler) Schedule(task ITask) {
	s.queue.Push(task)
}

func (s *Scheduler) GetStats() SchedulerStats {
	completed, failed := s.pool.GetCounters()
	return SchedulerStats{
		QueueSize:      s.queue.Len(),
		WorkerCount:    len(s.pool.GetStats()),
		Workers:        s.pool.GetStats(),
		TasksCompleted: completed,
		TasksFailed:    failed,
	}
}
