package scheduler

import (
	"context"
	"time"
)

type dispatcher struct {
	queue  TaskQueue
	pool   WorkerPool
	ctx    context.Context
	cancel context.CancelFunc
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
	go d.loop()
}

func (d *dispatcher) Stop() {
	d.cancel()
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
