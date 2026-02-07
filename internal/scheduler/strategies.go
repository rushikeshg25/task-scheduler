package scheduler

import (
	"log"
	"time"
)

func HandleTaskLifecycle(q TaskQueue, task ITask) {
	bt, ok := task.(*BaseTask)
	if !ok {
		return
	}

	switch bt.Status {
	case StatusCompleted:
		if bt.Interval > 0 {
			bt.ScheduledAt = time.Now().Add(bt.Interval)
			bt.Status = StatusPending
			q.Push(bt)
			log.Printf("Rescheduled periodic task %s for %v", bt.ID, bt.ScheduledAt)
		}
	case StatusFailed:
		if bt.RetryCount < bt.RetryLimit {
			bt.RetryCount++
			backoff := time.Duration(bt.RetryCount) * 100 * time.Millisecond
			bt.ScheduledAt = time.Now().Add(backoff)
			bt.Status = StatusRetrying
			q.Push(bt)
			log.Printf("Retrying task %s (%d/%d) in %v", bt.ID, bt.RetryCount, bt.RetryLimit, backoff)
		} else {
			log.Printf("Task %s failed after %d retries. Last error: %v", bt.ID, bt.RetryLimit, bt.LastError)
		}
	}
}
