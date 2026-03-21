# TDD Plan: Split-pane dashboard

## Context
The Integrator gets blocked by permission prompts when merging PRs, stopping all work. A split-pane dashboard in the integrator tab solves this: Claude runs on the left, a live dashboard runs on the right showing status, activity feed, and pending approvals. The Visionary approves merges when ready without interrupting the Integrator.

## Architecture
tmux split-pane within the integrator window. Left pane: Claude Code (70%). Right pane: `rf dashboard` command (30%). The dashboard polls status every 10 seconds, tails an activity log, and reads a file-based approval queue. Mission control writes activity events. The Integrator queues merge requests instead of merging directly.

## Session Constants
Test command: `go test -race ./...`
Test file pattern: colocated `*_test.go`
Test helpers: `internal/testutil/` (tmux.go, git.go)

## Slice 1: tmux SplitPane + session wiring
Type: unit + integration | Status: pending
Files: `internal/tmux/tmux.go`, `internal/session/session.go`, `internal/session/session_integration_test.go`, `cmd/up.go`

- [ ] SplitPane creates a horizontal split in a tmux window
- [ ] SplitPane runs a command in the new pane
- [ ] session.Setup splits integrator window with `rf dashboard` in right pane (30%)
- [ ] Integration test: split pane exists after Setup

## Slice 2: `rf dashboard` — live status display
Type: unit | Status: pending
Files: `internal/dashboard/dashboard.go`, `internal/dashboard/dashboard_test.go`, `cmd/dashboard.go`
Builds on: Slice 1

- [ ] Render() formats worker status for narrow pane (~35 chars wide)
- [ ] Render() shows session state (active/inactive)
- [ ] Render() shows worker list with PR status
- [ ] Dashboard refreshes on a 10-second ticker
- [ ] Graceful handling when no workers exist

## Slice 3: Activity feed from mission control
Type: unit | Status: pending
Files: `internal/dashboard/activity.go`, `internal/dashboard/activity_test.go`, `cmd/heartbeat.go`
Builds on: Slice 2

- [ ] Mission control writes timestamped events to .rocket-fuel/activity.log
- [ ] Dashboard reads last 10 lines from activity log
- [ ] Events include: dispatch, reap, nudge sent
- [ ] Activity section renders below status in dashboard

## Slice 4: Approval queue for pending merges
Type: unit | Status: pending
Files: `internal/dashboard/approvals.go`, `internal/dashboard/approvals_test.go`, `internal/prime/integrator.md`
Builds on: Slice 3

- [ ] AddPending writes merge request to .rocket-fuel/pending-merges.json
- [ ] ListPending returns current queue
- [ ] RemovePending removes approved item
- [ ] Dashboard shows pending approvals at top of display
- [ ] Integrator prompt updated: queue merges instead of merging directly
- [ ] Integrator prompt: respond to "merge #N" by executing from queue
