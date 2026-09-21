# Decisions

Observed behavior and documented contracts are separated from inferred motivation. These observations describe the current implementation; they do not establish the original author's intent.

## One process owns the complete task lifecycle

- **What:** The scheduler owns an in-memory queue and task IDs, and acceptance transfers task ownership.
- **Evidence:** The ownership comment and validation live in [engine.go:74](../../internal/scheduler/engine.go#L74); the v1 scope explicitly excludes persistence and distributed scheduling ([V1.md:5](../../V1.md#L5)).
- **Why:** The exclusion is confirmed by the contract. Keeping all state local also avoids coordination between independent task owners; that benefit is inferred.
- **Tradeoff:** Simple local execution and synchronization, but process exit loses work, queue size is unbounded and accepted IDs are never released ([engine.go:87](../../internal/scheduler/engine.go#L87), [queue.go:58](../../internal/scheduler/queue.go#L58)).
- **Confidence:** Scope and ownership confirmed by documentation/code; rationale beyond the scope is inferred.

## Time first, priority second, FIFO third

- **What:** The heap prioritizes earliest scheduled time; numeric priority only breaks exact time ties, with enqueue sequence as the final tie-breaker.
- **Evidence:** [queue.go:17](../../internal/scheduler/queue.go#L17), [queue_test.go:9](../../internal/scheduler/queue_test.go#L9), [lifecycle_test.go:45](../../internal/scheduler/lifecycle_test.go#L45).
- **Why, inferred:** Preserve delayed-work eligibility and deterministic ordering for equal keys.
- **Tradeoff:** Predictable ties, but a high-priority task with a later due time does not outrank an earlier low-priority task. Independently created immediate tasks usually have different `time.Now()` timestamps ([task.go:32](../../internal/scheduler/task.go#L32)).
- **Confidence:** Ordering confirmed by implementation/tests; motivation inferred.

## Polling dispatch with unbuffered handoff

- **What:** One dispatcher polls every 100 ms and blocks when every worker is busy; the built-in queue checks eligibility and removes a task in one critical section.
- **Evidence:** [dispatcher.go:48](../../internal/scheduler/dispatcher.go#L48), [queue.go:92](../../internal/scheduler/queue.go#L92), [pool.go:33](../../internal/scheduler/pool.go#L33).
- **Why, inferred:** Keep timer management simple and avoid a second buffered task queue between dispatcher and workers.
- **Tradeoff:** Best-effort timing, no immediate wakeup on new submissions, and one already-popped task may wait outside queue statistics. The atomic operation requires a concrete queue type assertion because `TaskQueue` lacks it ([dispatcher.go:67](../../internal/scheduler/dispatcher.go#L67)).
- **Confidence:** Mechanism confirmed; motivation inferred. Best-effort timing is explicitly documented ([README.md:18](../../README.md#L18)).

## Completion-relative periods and linear retries

- **What:** Only BaseTask receives built-in rescheduling. Successful periodic tasks become due at completion plus interval; failed tasks consume a retry and wait `RetryCount * 100ms`.
- **Evidence:** [strategies.go:8](../../internal/scheduler/strategies.go#L8), with callback wiring in [engine.go:29](../../internal/scheduler/engine.go#L29).
- **Why, inferred:** Reuse one task object and one queue path for repeated execution; completion-driven requeue prevents normal periodic execution of the same object from overlapping itself.
- **Tradeoff:** The period includes runtime and dispatch delays; there is no fixed calendar cadence, jitter or exponential backoff. RetryCount persists after success, so periodic jobs share a lifetime retry budget. A periodic job stops after exhausting it ([task.go:58](../../internal/scheduler/task.go#L58), [strategies.go:22](../../internal/scheduler/strategies.go#L22)).
- **Confidence:** Behavior confirmed; implementation motivation inferred.

## Cancellation before joining workers

- **What:** Stop marks the instance terminal, cancels the task context, stops dispatch, then joins workers. Concurrent callers wait on a shared done channel.
- **Evidence:** [engine.go:58](../../internal/scheduler/engine.go#L58), [pool.go:82](../../internal/scheduler/pool.go#L82), [lifecycle_test.go:10](../../internal/scheduler/lifecycle_test.go#L10).
- **Why, inferred:** Release a blocked Submit and let active work finish without holding the scheduler mutex; the test explicitly covers a running task attempting to schedule after cancellation.
- **Tradeoff:** No restart or queue drain. Task cooperation is required, and a task cannot call Stop on its own scheduler without waiting for itself ([README.md:16](../../README.md#L16)).
- **Confidence:** Lifecycle contract confirmed; reasoning inferred from control flow and tests.

## Gotchas

- **Default retries:** `NewBaseTask` defaults to three retries, meaning up to four executions; `WithRetry(0)` disables retry. RetryLimit counts additional attempts ([task.go:32](../../internal/scheduler/task.go#L32), [strategies.go:23](../../internal/scheduler/strategies.go#L23)).
- **Panic coverage is narrow:** Only `Execute` is recovered. Getters, OnSuccess, OnFailure and the completion callback execute outside that guard. The README's “task panic” statement should be read with this boundary ([pool.go:113](../../internal/scheduler/pool.go#L113), [README.md:16](../../README.md#L16)).
- **Custom typed nils:** Schedule specially rejects a nil `*BaseTask`; a nil pointer of another `ITask` type reaches its getters and may panic. Custom tasks also do not get automatic periodic/retry behavior ([engine.go:78](../../internal/scheduler/engine.go#L78), [strategies.go:9](../../internal/scheduler/strategies.go#L9)).
- **Snapshots are not exact totals:** Completed/failed counters count attempts, including retry and periodic executions. Dashboard “Total Tasks” adds queue depth but omits active tasks and the dispatcher's blocked handoff; worker rendering shows at most four workers ([pool.go:124](../../internal/scheduler/pool.go#L124), [ui.go:149](../../cmd/scheduler/ui.go#L149), [ui.go:194](../../cmd/scheduler/ui.go#L194)).
- **Demo cancellation:** Several sample functions sleep instead of selecting on context cancellation, so normal quit can wait for those sleeps. A UI Run error calls `os.Exit(1)` before Stop ([main.go:34](../../cmd/scheduler/main.go#L34), [main.go:56](../../cmd/scheduler/main.go#L56)).
- **Public lower-level constructors:** Pool Stop alone closes its quit channel but does not cancel the context supplied by its caller. Scheduler Stop provides that cancellation. `SetOnComplete` is an unsynchronized assignment intended by current wiring to occur before Start ([pool.go:39](../../internal/scheduler/pool.go#L39), [pool.go:73](../../internal/scheduler/pool.go#L73), [engine.go:29](../../internal/scheduler/engine.go#L29)).
- **Mutation after acceptance:** BaseTask fields are exported and unsynchronized. Editing or reading them after ownership transfer can race and corrupt heap ordering; use GetStats for supported observation ([engine.go:74](../../internal/scheduler/engine.go#L74), [task.go:20](../../internal/scheduler/task.go#L20)).

## Conventions

- Components expose small interfaces in one file, while implementation structs for pool and dispatcher remain unexported ([interfaces.go:8](../../internal/scheduler/interfaces.go#L8), [pool.go:11](../../internal/scheduler/pool.go#L11), [dispatcher.go:9](../../internal/scheduler/dispatcher.go#L9)).
- Configuration uses chainable BaseTask methods before Schedule; follow that ownership pattern when adding examples ([task.go:67](../../internal/scheduler/task.go#L67), [main.go:42](../../cmd/scheduler/main.go#L42)).
- Tests reside in package `scheduler` and combine channels/wait groups with bounded timeouts for completion; queue-order tests compare explicit due times ([integration_test.go:11](../../internal/scheduler/integration_test.go#L11), [lifecycle_test.go:10](../../internal/scheduler/lifecycle_test.go#L10), [queue_test.go:12](../../internal/scheduler/queue_test.go#L12)). The current tests do not establish production latency or throughput ([README.md:25](../../README.md#L25)).
