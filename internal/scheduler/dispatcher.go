package scheduler

import (
	"context"
	"sync"
	"time"
)

type dispatcher struct {
	mu               sync.Mutex
	started, stopped bool
	wg               sync.WaitGroup
	queue            TaskQueue
	pool             WorkerPool
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewDispatcher(queue TaskQueue, pool WorkerPool) Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &dispatcher{
		queue:  queue,
		pool:   pool,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (d *dispatcher) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.started || d.stopped {
		return
	}
	d.started = true
	d.wg.Add(1)
	go func() { defer d.wg.Done(); d.loop() }()
}

func (d *dispatcher) Stop() {
	d.mu.Lock()
	d.stopped = true
	d.cancel()
	d.mu.Unlock()
	d.wg.Wait()
}

func (d *dispatcher) loop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.dispatch()
		}
	}
}

func (d *dispatcher) dispatch() {
	for {
		if d.ctx.Err() != nil {
			return
		}
		if q, ok := d.queue.(*PriorityQueue); ok {
			task := q.PopReady(time.Now())
			if task == nil {
				return
			}
			d.pool.Submit(task)
			continue
		}
		task := d.queue.Peak()
		if task == nil {
			return
		}

		if time.Now().After(task.GetScheduledAt()) {
			task = d.queue.Pop()
			d.pool.Submit(task)
		} else {
			return
		}
	}
}
