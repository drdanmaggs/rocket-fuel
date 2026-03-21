// Package dispatch handles automated worker spawning from the project board.
package dispatch

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/project"
)

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
	FailedIssues   map[int]bool // issues that failed to spawn — skip on retry
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
			// Track the failure so we don't retry next cycle.
			if deps.FailedIssues != nil {
				deps.FailedIssues[next.Number] = true
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
func nextDispatchable(board *project.BoardSummary, failed map[int]bool) *project.Item {
	for _, name := range project.ReadyColumnNames() {
		items := board.Columns[name]
		for i := range items {
			if failed != nil && failed[items[i].Number] {
				continue
			}
			return &items[i]
		}
	}
	return nil
}
