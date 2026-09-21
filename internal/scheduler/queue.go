package scheduler

import (
	"container/heap"
	"sync"
	"time"
)

type queuedTask struct {
	ITask
	sequence uint64
}
type TaskHeap []*queuedTask

func (h TaskHeap) Len() int { return len(h) }

func (h TaskHeap) Less(i, j int) bool {
	if h[i].GetScheduledAt().Equal(h[j].GetScheduledAt()) {
		if h[i].GetPriority() == h[j].GetPriority() {
			return h[i].sequence < h[j].sequence
		}
		return h[i].GetPriority() > h[j].GetPriority()
	}
	return h[i].GetScheduledAt().Before(h[j].GetScheduledAt())
}

func (h TaskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *TaskHeap) Push(x interface{}) {
	item := x.(*queuedTask)
	*h = append(*h, item)
}

func (h *TaskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

type PriorityQueue struct {
	next uint64
	mu   sync.Mutex
	heap TaskHeap
}

func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		heap: make(TaskHeap, 0),
	}
	heap.Init(&pq.heap)
	return pq
}

func (pq *PriorityQueue) Push(task ITask) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if task == nil {
		return
	}
	pq.next++
	heap.Push(&pq.heap, &queuedTask{task, pq.next})
}

func (pq *PriorityQueue) Pop() ITask {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if len(pq.heap) == 0 {
		return nil
	}
	return heap.Pop(&pq.heap).(*queuedTask).ITask
}

func (pq *PriorityQueue) Peak() ITask {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if len(pq.heap) == 0 {
		return nil
	}
	return pq.heap[0]
}

func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.heap.Len()
}

func (pq *PriorityQueue) PopReady(now time.Time) ITask {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if len(pq.heap) == 0 || pq.heap[0].GetScheduledAt().After(now) {
		return nil
	}
	return heap.Pop(&pq.heap).(*queuedTask).ITask
}
