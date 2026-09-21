# Structure

## What lives where

The 20 tracked files form one Go module. The terminal entry point and UI live in `cmd/scheduler`; all scheduling logic and tests share `internal/scheduler`. Root documents describe the v1 contract and delivery history; the Makefile supplies local commands ([go.mod:1](../../go.mod#L1), [main.go:11](../../cmd/scheduler/main.go#L11), [Makefile:5](../../Makefile#L5)).

```text
cmd/scheduler/           Terminal demonstration and dashboard
internal/scheduler/      Engine, contracts, task model, queue, workers and tests
docs/project-guide/     This maintainer guide
README.md               Usage and public behavior
V1.md                   Delivery contract
HISTORY.md              Evidence-backed delivery history
go.mod                  Module, toolchain requirement and dependencies
Makefile                Build, run, test, lint and clean commands
.golangci.yml           Linter selection
```

## Repository root

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [go.mod](../../go.mod#L1) | Module path, Go directive and dependency versions. | Module metadata. | Go tooling. |
| [Makefile](../../Makefile#L1) | Local build/test/run/lint/clean workflow. | Make targets. | Developers. |
| [.golangci.yml](../../.golangci.yml#L1) | Select linters and timeout. | Linter configuration. | `make lint`. |
| [README.md](../../README.md#L1) | Usage, ownership and lifecycle contract. | Documentation. | Users and maintainers. |
| [V1.md](../../V1.md#L1) | Scope and acceptance criteria. | Documentation. | Delivery review. |
| [HISTORY.md](../../HISTORY.md#L1) | Timeline and reported verification evidence. | Documentation. | Maintainers. |

## Terminal command

Implementation folder: [cmd/scheduler/](../../cmd/scheduler).

| File | Responsibility | Key exports or entry points | Called by |
| --- | --- | --- | --- |
| [main.go](../../cmd/scheduler/main.go#L14) | Log adapter, four-worker scheduler, eight demo tasks, terminal program and normal shutdown. | `main`, `logWriter.Write` (package-local type). | Go executable runtime; standard logger. |
| [ui.go](../../cmd/scheduler/ui.go#L59) | Bubble Tea model, stats ticks, log viewport, styled dashboard and quit keys. | `initialModel`, model `Init`, `Update`, `View` (package-local model). | Main and Bubble Tea runtime. |

## Scheduler package

Implementation folder: [internal/scheduler/](../../internal/scheduler). Files share the same package; these are responsibility boundaries rather than separate deployable services.

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [interfaces.go](../../internal/scheduler/interfaces.go#L8) | Task/queue/pool/dispatcher contracts and snapshot types. | `ITask`, `TaskQueue`, `WorkerPool`, `Dispatcher`, `WorkerStats`, `SchedulerStats`. | Every scheduling component; UI consumes stats. |
| [engine.go](../../internal/scheduler/engine.go#L9) | Component wiring, owned task IDs, lifecycle and statistics facade. | `Scheduler`, `NewScheduler`, `Start`, `Stop`, `Schedule`, `GetStats`. | Main, UI and integration/lifecycle tests. |
| [task.go](../../internal/scheduler/task.go#L8) | Mutable task model, function execution, result callbacks and fluent configuration. | `TaskStatus`, status constants, `TaskFunc`, `BaseTask`, `NewBaseTask`, `WithDelay`, `WithPeriodic`, `WithRetry`. | Main/tests construct; pool executes; queue reads keys; lifecycle handler updates. |
| [queue.go](../../internal/scheduler/queue.go#L9) | Stable heap ordering, mutex-protected queue, atomic ready removal. | `TaskHeap`, `PriorityQueue`, `NewPriorityQueue`, `Push`, `Pop`, `Peak`, `Len`, `PopReady`. | Engine, dispatcher, strategies and queue tests. |
| [dispatcher.go](../../internal/scheduler/dispatcher.go#L9) | Idempotent polling-loop lifecycle and queue-to-pool handoff. | `NewDispatcher` returns `Dispatcher`. | Engine; dispatcher calls queue and pool. |
| [pool.go](../../internal/scheduler/pool.go#L11) | Worker goroutines, cancellation-aware submission, execution recovery, status copies and atomic counters. | `NewWorkerPool` returns `WorkerPool`. | Engine wires lifecycle callback; dispatcher submits tasks. |
| [strategies.go](../../internal/scheduler/strategies.go#L8) | Requeue successful periodic tasks or retry failed BaseTasks. | `HandleTaskLifecycle`. | Engine-installed pool completion callback. |
| [integration_test.go](../../internal/scheduler/integration_test.go#L11) | Demonstrates immediate/delayed/periodic execution and eventual retry success. | `TestSchedulerIntegration`, `TestRetryLogic`. | `go test`. |
| [queue_test.go](../../internal/scheduler/queue_test.go#L9) | Checks time-first and priority-second heap ordering. | `TestPriorityQueue`. | `go test`. |
| [lifecycle_test.go](../../internal/scheduler/lifecycle_test.go#L10) | Checks blocked dispatch cancellation, stop-before-start, typed-nil BaseTask, FIFO ties, worker survival after panic, concurrent lifecycle calls. | Five `Test...` functions. | `go test`. |

## Excluded

The dependency checksum lockfile [go.sum](../../go.sum) and routine ignore rules [.gitignore](../../.gitignore) do not add runtime design information. Generated binaries and local decision logs are not part of the significant source inventory. This guide adds documentation only; no tests or implementation files are generated by it.
