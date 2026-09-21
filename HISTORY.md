# In-memory task scheduler history

> Scope: repository baseline through the v1 delivery branch.
> Last updated: 2026-09-21. Dates below retain Git author timezone offsets.

## At a glance

The v1 scope and acceptance contract is recorded in [V1.md](V1.md). Current behavior and limitations are documented in [README.md](README.md).

## Timeline

### 2026-02-07T09:20:43+05:30: Initial commit

- **What happened:** The repository records `Initial commit`.
- **Evidence:** [commit b850c1e141](https://github.com/rushikeshg25/task-scheduler/commit/b850c1e141d39ff08f0e76b14b78b0ceafc12f09).
- **Confidence:** Confirmed by Git history; the title alone does not establish runtime correctness.

### 2026-09-21T13:21:37+05:30: docs: define task-scheduler v1 contract

- **What happened:** The repository records `docs: define task-scheduler v1 contract`.
- **Evidence:** [commit 7c8c831d20](https://github.com/rushikeshg25/task-scheduler/commit/7c8c831d20e476b1472b4aace9caea2a93574ab0).
- **Confidence:** Confirmed by Git history; the title alone does not establish runtime correctness.

### 2026-09-21T13:35:07+05:30: fix: make shutdown cancellation safe and dispatch due tasks atomically

- **What happened:** The repository records `fix: make shutdown cancellation safe and dispatch due tasks atomically`.
- **Evidence:** [commit c04d20a517](https://github.com/rushikeshg25/task-scheduler/commit/c04d20a5179479d86ac0a0f51b3eaf86ddd4ab52).
- **Confidence:** Confirmed by Git history; the title alone does not establish runtime correctness.

### 2026-09-21T13:36:02+05:30: test: close lifecycle races and verify stable scheduling and panic recovery

- **What happened:** The repository records `test: close lifecycle races and verify stable scheduling and panic recovery`.
- **Evidence:** [commit 4ee9ab731a](https://github.com/rushikeshg25/task-scheduler/commit/4ee9ab731a20680cb9f6ab76bd18dd92818f3b42).
- **Confidence:** Confirmed by Git history; the title alone does not establish runtime correctness.

### 2026-09-21T13:38:57+05:30: docs: document task-scheduler v1 usage and limitations

- **What happened:** The repository records `docs: document task-scheduler v1 usage and limitations`.
- **Evidence:** [commit 06b49df4e8](https://github.com/rushikeshg25/task-scheduler/commit/06b49df4e8651c9d68cbb5ee8101f9b308088e9d).
- **Confidence:** Confirmed by Git history; the title alone does not establish runtime correctness.

## Delivery verification

Passed `go test -race ./...` for all packages, including UI compilation and [lifecycle tests](internal/scheduler/lifecycle_test.go).

## Turning points

The v1 contract made failure behavior, lifecycle semantics and executable verification part of the delivery. The new tests and README describe the resulting boundaries.

## Open questions

No production deployment or long-running operational validation was performed as part of this delivery.
