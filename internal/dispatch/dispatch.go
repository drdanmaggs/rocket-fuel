// Package dispatch handles automated worker spawning from the project board.
package dispatch

import (
	"fmt"
	"time"

	"github.com/drdanmaggs/rocket-fuel/internal/project"
)

// FailRetryTTL is how long a failed issue is skipped before retrying.
const FailRetryTTL = 15 * time.Minute

// FailedIssues tracks issues that failed to spawn, with timestamps for TTL expiry.
type FailedIssues map[int]time.Time

// Record marks an issue as recently failed.
func (f FailedIssues) Record(issueNumber int) {
	f[issueNumber] = time.Now()
}

// ShouldSkip returns true if the issue failed recently (within TTL).
func (f FailedIssues) ShouldSkip(issueNumber int) bool {
	failTime, ok := f[issueNumber]
	if !ok {
		return false
	}
	if time.Since(failTime) > FailRetryTTL {
		delete(f, issueNumber)
		return false
	}
	return true
}

// Config holds dispatch settings.
type Config struct {
	MaxWorkers int // maximum concurrent workers (default 3)
}

// SpawnFunc spawns a worker for the given issue number.
// The cmd layer provides a real implementation that calls worker.Spawn.
type SpawnFunc func(issueNumber int) error

// TransitionFunc moves a board item to a new status.
type TransitionFunc func(itemID, targetStatus string) error

// Deps holds pre-fetched dependencies for a dispatch cycle.
type Deps struct {
	Board          *project.BoardSummary
	ActiveWorkers  int
	SpawnFunc      SpawnFunc
	TransitionFunc TransitionFunc
	FailedIssues   FailedIssues // issues that failed to spawn — skip until TTL expires
}

// Result describes what happened during a dispatch cycle.
type Result struct {
	Dispatched  bool
	IssueNumber int
	WorkerName  string
	Reason      string
}

// Run executes one dispatch cycle: check for Ready items, check capacity, spawn.
func Run(cfg Config, deps Deps) (*Result, error) {
	// Find next ready item, skipping previously failed issues.
	next := nextDispatchable(deps.Board, deps.FailedIssues)
	if next == nil {
		return &Result{Reason: "nothing to dispatch"}, nil
	}

	// Check capacity.
	if deps.ActiveWorkers >= cfg.MaxWorkers {
		return &Result{
			Reason: fmt.Sprintf("at capacity: %d/%d workers", deps.ActiveWorkers, cfg.MaxWorkers),
		}, nil
	}

	// Spawn worker.
	if deps.SpawnFunc != nil {
		if err := deps.SpawnFunc(next.Number); err != nil {
			// Track the failure so we don't retry until TTL expires.
			if deps.FailedIssues != nil {
				deps.FailedIssues.Record(next.Number)
			}
			return nil, fmt.Errorf("spawn worker for #%d: %w", next.Number, err)
		}
	}

	// Move item from Scoped to In Progress.
	if deps.TransitionFunc != nil && next.ID != "" {
		if err := deps.TransitionFunc(next.ID, "In Progress"); err != nil {
			// Log but don't fail — the worker is already spawned.
			return &Result{
				Dispatched:  true,
				IssueNumber: next.Number,
				WorkerName:  fmt.Sprintf("worker-%d", next.Number),
				Reason:      fmt.Sprintf("dispatched #%d (board transition failed: %v)", next.Number, err),
			}, nil
		}
	}

	return &Result{
		Dispatched:  true,
		IssueNumber: next.Number,
		WorkerName:  fmt.Sprintf("worker-%d", next.Number),
		Reason:      fmt.Sprintf("dispatched #%d", next.Number),
	}, nil
}

// nextDispatchable returns the first Ready item that hasn't previously failed.
func nextDispatchable(board *project.BoardSummary, failed FailedIssues) *project.Item {
	for _, name := range project.ReadyColumnNames() {
		items := board.Columns[name]
		for i := range items {
			if failed != nil && failed.ShouldSkip(items[i].Number) {
				continue
			}
			return &items[i]
		}
	}
	return nil
}
