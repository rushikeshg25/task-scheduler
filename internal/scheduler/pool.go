package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

type workerPool struct {
	life             sync.Mutex
	started, stopped bool
	ctx              context.Context
	workerCount      int
	taskChan         chan ITask
	wg               sync.WaitGroup
	quit             chan struct{}
	onComplete       func(ITask)

	mu             sync.RWMutex
	workerStatuses []WorkerStats
	tasksCompleted int64
	tasksFailed    int64
}

func NewWorkerPool(workerCount int) WorkerPool {
	if workerCount < 1 {
		workerCount = 1
	}
	return &workerPool{
		workerCount:    workerCount,
		taskChan:       make(chan ITask),
		quit:           make(chan struct{}),
		workerStatuses: make([]WorkerStats, workerCount),
	}
}

func (p *workerPool) SetOnComplete(onComplete func(ITask)) {
	p.onComplete = onComplete
}

func (p *workerPool) GetStats() []WorkerStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	stats := make([]WorkerStats, len(p.workerStatuses))
	copy(stats, p.workerStatuses)
	return stats
}

func (p *workerPool) GetCounters() (int64, int64) {
	return atomic.LoadInt64(&p.tasksCompleted), atomic.LoadInt64(&p.tasksFailed)
}

func (p *workerPool) Start(ctx context.Context) {
	p.life.Lock()
	defer p.life.Unlock()
	if p.started || p.stopped {
		return
	}
	p.started = true
	p.ctx = ctx
	p.mu.Lock()
	defer p.mu.Unlock()
	log.Printf("Starting worker pool with %d workers", p.workerCount)
	for i := 0; i < p.workerCount; i++ {
		p.workerStatuses[i] = WorkerStats{ID: i, Status: "Idle"}
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

func (p *workerPool) Stop() {
	p.life.Lock()
	if !p.stopped {
		p.stopped = true
		close(p.quit)
	}
	p.life.Unlock()
	p.wg.Wait()
}
func (p *workerPool) Submit(task ITask) {
	if task == nil {
		return
	}
	p.life.Lock()
	ctx := p.ctx
	p.life.Unlock()
	if ctx == nil {
		return
	}
	select {
	case p.taskChan <- task:
	case <-p.quit:
	case <-ctx.Done():
	}
}

func (p *workerPool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.quit:
			return
		case <-ctx.Done():
			return
		case task := <-p.taskChan:
			p.executeTask(ctx, id, task)
		}
	}
}

func (p *workerPool) executeTask(ctx context.Context, workerID int, task ITask) {
	p.mu.Lock()
	p.workerStatuses[workerID].Status = "Executing"
	p.workerStatuses[workerID].CurrentTaskID = task.GetID()
	p.mu.Unlock()

	log.Printf("Worker %d: Executing task %s", workerID, task.GetID())
	err := executeSafely(ctx, task)
	if err != nil {
		log.Printf("Worker %d: Task %s failed: %v", workerID, task.GetID(), err)
		task.OnFailure(err)
		atomic.AddInt64(&p.tasksFailed, 1)
	} else {
		log.Printf("Worker %d: Task %s completed", workerID, task.GetID())
		task.OnSuccess()
		atomic.AddInt64(&p.tasksCompleted, 1)
	}

	p.mu.Lock()
	p.workerStatuses[workerID].Status = "Idle"
	p.workerStatuses[workerID].CurrentTaskID = ""
	p.mu.Unlock()

	if p.onComplete != nil {
		p.onComplete(task)
	}
}

func executeSafely(ctx context.Context, task ITask) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("task panic: %v", p)
		}
	}()
	return task.Execute(ctx)
}
