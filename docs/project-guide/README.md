# Task Scheduler Project Guide

> Generated: 2026-09-21 from commit `11d6895`.

## What this is

This Go project runs delayed, periodic and retrying functions through an in-memory priority queue and a fixed worker pool ([engine.go:21](../../internal/scheduler/engine.go#L21)). Its executable is a terminal demonstration for developers, with sample jobs and a Bubble Tea dashboard ([main.go:25](../../cmd/scheduler/main.go#L25)). The scheduler lives in an `internal` package, so the exported API is available inside the module's permitted Go import tree rather than as a generally importable external library ([module](../../go.mod#L1), [package](../../internal/scheduler/engine.go#L1)).

## Run it

From the repository root, with Go 1.24.6 or a compatible newer toolchain ([go.mod:3](../../go.mod#L3)) and an interactive terminal:

```bash
go run ./cmd/scheduler
go test -race ./...
make build
```

The build produces `bin/scheduler`; `q` or Ctrl+C exits the dashboard ([Makefile:5](../../Makefile#L5), [ui.go:95](../../cmd/scheduler/ui.go#L95)). There are no application environment variables or external services to configure in the entry point; worker count and sample tasks are hardcoded ([main.go:27](../../cmd/scheduler/main.go#L27)).

## The five-file tour

| # | File | Why this one | Then look at |
| --- | --- | --- | --- |
| 1 | [cmd/scheduler/main.go](../../cmd/scheduler/main.go#L25) | Starts the scheduler and terminal program, submits sample tasks. | [Startup](02-flow.md#startup) |
| 2 | [internal/scheduler/engine.go](../../internal/scheduler/engine.go#L21) | Wires the components, owns acceptance and shutdown. | [Components](01-architecture.md#components) |
| 3 | [internal/scheduler/queue.go](../../internal/scheduler/queue.go#L17) | Defines time, priority and FIFO ordering plus atomic due removal. | [Scheduling and execution](02-flow.md#scheduling-and-execution) |
| 4 | [internal/scheduler/pool.go](../../internal/scheduler/pool.go#L113) | Runs functions, updates counters, invokes lifecycle callbacks. | [Scheduler package](03-structure.md#scheduler-package) |
| 5 | [internal/scheduler/strategies.go](../../internal/scheduler/strategies.go#L8) | Closes the loop with retries and periodic rescheduling. | [Decisions](05-decisions.md#completion-relative-periods-and-linear-retries) |

## Reading order for this guide

1. [Architecture](01-architecture.md) — components, contracts and process-local state.
2. [Flow](02-flow.md) — startup, execution, recurring jobs, UI and shutdown.
3. [Structure](03-structure.md) — every significant source file and its callers.
4. [Tech stack](04-tech-stack.md) — version pins and tooling.
5. [Decisions](05-decisions.md) — evidence, inferred rationale and gotchas.

## Open questions

- Should periodic jobs receive a fresh retry budget after success? The current implementation never resets `RetryCount`, and an exhausted failure ends periodic scheduling ([strategies.go:14](../../internal/scheduler/strategies.go#L14), [task.go:58](../../internal/scheduler/task.go#L58)).
- Is arbitrary custom `ITask` hardening intended? The README broadly promises invalid-input rejection and panic conversion, while typed-nil rejection specifically handles `*BaseTask`, and recovery surrounds `Execute` only ([README.md:14](../../README.md#L14), [engine.go:78](../../internal/scheduler/engine.go#L78), [pool.go:120](../../internal/scheduler/pool.go#L120)).
- What operational limits are acceptable for queue size, retained IDs and cancellation latency? There is no capacity limit or ID eviction, and shutdown waits on task cooperation ([engine.go:87](../../internal/scheduler/engine.go#L87), [queue.go:58](../../internal/scheduler/queue.go#L58), [pool.go:73](../../internal/scheduler/pool.go#L73)). The delivery history explicitly leaves production operation unvalidated ([HISTORY.md](../../HISTORY.md#open-questions)).
