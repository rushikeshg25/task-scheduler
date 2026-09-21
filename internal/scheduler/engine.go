package scheduler

import (
	"context"
	"log"
	"sync"
)

type Scheduler struct {
	done             chan struct{}
	mu               sync.Mutex
	started, stopped bool
	ids              map[string]bool
	queue            TaskQueue
	pool             WorkerPool
	dispatcher       Dispatcher
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewScheduler(workerCount int) *Scheduler {
	if workerCount < 1 {
		workerCount = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	q := NewPriorityQueue()
	p := NewWorkerPool(workerCount)

	p.SetOnComplete(func(t ITask) {
		HandleTaskLifecycle(q, t)
	})

	d := NewDispatcher(q, p)

	return &Scheduler{
		queue:      q,
		ids:        make(map[string]bool),
		done:       make(chan struct{}),
		pool:       p,
		dispatcher: d,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started || s.stopped {
		return
	}
	s.started = true
	s.pool.Start(s.ctx)
	s.dispatcher.Start()
	log.Println("Scheduler started (Modular)")
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	if s.stopped {
		done := s.done
		s.mu.Unlock()
		<-done
		return
	}
	s.stopped = true
	s.cancel()
	s.mu.Unlock()
	s.dispatcher.Stop()
	s.pool.Stop()
	close(s.done)
}

// Schedule transfers ownership of task to the scheduler. IDs must be unique.
func (s *Scheduler) Schedule(task ITask) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if bt, ok := task.(*BaseTask); ok && bt == nil {
		return false
	}
	if s.stopped || task == nil || task.GetID() == "" || s.ids[task.GetID()] {
		return false
	}
	if bt, ok := task.(*BaseTask); ok && (bt == nil || bt.Fn == nil || bt.Interval < 0 || bt.RetryLimit < 0) {
		return false
	}
	s.ids[task.GetID()] = true
	s.queue.Push(task)
	return true
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
