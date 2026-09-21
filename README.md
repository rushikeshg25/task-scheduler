# In-memory task scheduler v1

A priority queue, bounded worker pool and Bubble Tea dashboard for delayed, recurring and retrying tasks.

```go
s := scheduler.NewScheduler(4)
s.Start()
defer s.Stop()
accepted := s.Schedule(scheduler.NewBaseTask("job", 10, func(ctx context.Context) error {
    return doWork(ctx)
}).WithDelay(time.Second).WithRetry(3))
```

Tasks sort by scheduled time, then descending priority, then FIFO for ties. At least one worker is used. Queue inspection and popping a due task happen atomically. `Schedule` rejects nil/invalid tasks, duplicate IDs and post-shutdown submissions. An accepted task transfers ownership: configure it before scheduling and do not access its mutable fields afterward. IDs remain reserved for the lifetime of the scheduler. Built-in periodic/retry behavior applies to BaseTask; custom ITask implementations provide their own callbacks.

Start/Stop are idempotent, including stop-before-start. Stop cancels the worker context, halts dispatch and waits for active work. Tasks must cooperate with context cancellation; queued work is abandoned and the instance cannot restart. Tasks may submit other tasks, but must not call Stop themselves because Stop waits for active work. Use GetStats for synchronized worker/counter snapshots. A task panic becomes a failure eligible for its retry policy.

Retries back off by `retry_count * 100ms`. Periodic jobs reschedule from completion time, so one periodic task does not overlap itself. Dispatch polls every 100ms: scheduling is best-effort, not real-time. State is in memory only; persistence, distributed ownership and crash recovery are outside v1.

```sh
go run ./cmd/scheduler
go test -race ./...
```

Tests cover delayed/periodic execution, retry behavior, stable ordering, panic recovery and concurrent shutdown. No production throughput claims are made.
