# Tech Stack

## Languages and runtimes

| Language or runtime | Version | Pinned at |
| --- | --- | --- |
| Go | `go 1.24.6` directive; no separate toolchain directive. | [go.mod:3](../../go.mod#L3) |
| Module | `github.com/rushikeshg25/task-scheduler`. | [go.mod:1](../../go.mod#L1) |

## Frameworks and major libraries

| Library | Version | Used for | Pin and use |
| --- | --- | --- | --- |
| Bubble Tea | v1.3.10 | Terminal event loop, model messages, ticks and alternate screen. | [go.mod:8](../../go.mod#L8), [main.go:30](../../cmd/scheduler/main.go#L30), [ui.go:82](../../cmd/scheduler/ui.go#L82) |
| Bubbles | v0.21.1 | Spinner and log viewport widgets. | [go.mod:7](../../go.mod#L7), [ui.go:8](../../cmd/scheduler/ui.go#L8) |
| Lip Gloss | v1.1.0 | Colors, borders and dashboard layout. | [go.mod:10](../../go.mod#L10), [ui.go:15](../../cmd/scheduler/ui.go#L15) |
| Go `container/heap` | Bundled with Go toolchain. | Priority queue heap operations. | [go.mod:3](../../go.mod#L3), [queue.go:4](../../internal/scheduler/queue.go#L4) |
| Go `context`, `sync`, `sync/atomic`, `time` | Bundled with Go toolchain. | Cancellation, locks, wait groups, counters and polling. | [go.mod:3](../../go.mod#L3), [pool.go:3](../../internal/scheduler/pool.go#L3), [dispatcher.go:3](../../internal/scheduler/dispatcher.go#L3) |

All third-party entries are marked `// indirect` in the manifest, including the three libraries directly imported by the command; the import sites above establish actual usage. The rest of the manifest contains terminal/color/text/platform support dependencies ([go.mod:5](../../go.mod#L5)).

## Data and infrastructure

| Service or mechanism | Role | Configured at |
| --- | --- | --- |
| Process-local Go heap | Task heap, reserved IDs, task objects and statistics; no datastore. | [engine.go:21](../../internal/scheduler/engine.go#L21), [queue.go:44](../../internal/scheduler/queue.go#L44) |
| Goroutines and unbuffered channel | Fixed worker concurrency and blocking task handoff. | [pool.go:27](../../internal/scheduler/pool.go#L27), [pool.go:66](../../internal/scheduler/pool.go#L66) |
| Interactive terminal | Input events and rendered dashboard. | [main.go:30](../../cmd/scheduler/main.go#L30), [ui.go:91](../../cmd/scheduler/ui.go#L91) |

There are no datastore or service version pins because the startup path constructs only local components and sample tasks. Persistence and distributed scheduling are explicitly excluded by the v1 contract ([V1.md:5](../../V1.md#L5)).

## Tooling

| Tool | Role | Configured at |
| --- | --- | --- |
| Go build/run | Compile `bin/scheduler` or run the terminal command. | [Makefile:5](../../Makefile#L5) |
| Go testing and race detector | Run all packages with verbose output and `-race`. | [Makefile:8](../../Makefile#L8), [integration_test.go:11](../../internal/scheduler/integration_test.go#L11) |
| golangci-lint | Local lint command; installed executable version is not pinned. | [Makefile:14](../../Makefile#L14), [.golangci.yml:1](../../.golangci.yml#L1) |
| Make | Convenience targets; version is not pinned. | [Makefile:1](../../Makefile#L1) |

## Notes

- Linter configuration enables errcheck, gosimple, govet, staticcheck, unused, revive, gofmt and misspell with a five-minute timeout; compatibility depends on the installed golangci-lint version ([.golangci.yml:1](../../.golangci.yml#L1)).
- No tracked CI or container/deployment configuration appears in the [source inventory](03-structure.md). Local commands are the documented verification entry point, and the history reports an earlier passing race-detector run ([HISTORY.md](../../HISTORY.md#delivery-verification)); this documentation pass did not rerun it.
