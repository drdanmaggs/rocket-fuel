package dispatch

import (
	"fmt"
	"strings"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/project"
)

func TestRun_nothingToDispatchWhenNoScopedItems(t *testing.T) {
	t.Parallel()

	board := &project.BoardSummary{
		Columns: map[string][]project.Item{
			"Backlog": {{Number: 1, Title: "Backlog item"}},
		},
	}

	result, err := Run(Config{MaxWorkers: 3}, Deps{
		Board:         board,
		ActiveWorkers: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Dispatched {
		t.Error("expected no dispatch when no Scoped items")
	}
	if result.Reason != "nothing to dispatch" {
		t.Errorf("expected reason 'nothing to dispatch', got %q", result.Reason)
	}
}

func TestRun_skipsWhenAtCapacity(t *testing.T) {
	t.Parallel()

	board := &project.BoardSummary{
		Columns: map[string][]project.Item{
			"Scoped": {{Number: 42, Title: "Ready issue"}},
		},
	}

	result, err := Run(Config{MaxWorkers: 3}, Deps{
		Board:         board,
		ActiveWorkers: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Dispatched {
		t.Error("expected no dispatch at capacity")
	}
	if result.Reason != "at capacity: 3/3 workers" {
		t.Errorf("expected capacity message, got %q", result.Reason)
	}
}

func TestRun_returnsErrorOnSpawnFailure(t *testing.T) {
	t.Parallel()

	board := &project.BoardSummary{
		Columns: map[string][]project.Item{
			"Scoped": {{Number: 42, Title: "Ready issue"}},
		},
	}

	spawnFn := func(_ int) error {
		return fmt.Errorf("worktree already exists")
	}

	_, err := Run(Config{MaxWorkers: 3}, Deps{
		Board:         board,
		ActiveWorkers: 0,
		SpawnFunc:     spawnFn,
	})
	if err == nil {
		t.Fatal("expected error on spawn failure")
	}
	if !strings.Contains(err.Error(), "#42") {
		t.Errorf("expected error to mention issue number, got: %v", err)
	}
}

func TestRun_dispatchesWhenCapacityAvailable(t *testing.T) {
	t.Parallel()

	board := &project.BoardSummary{
		Columns: map[string][]project.Item{
			"Scoped": {{Number: 42, Title: "Ready issue", ID: "PVTI_42", Labels: []string{"workflow:tdd"}}},
		},
	}

	var spawnedIssue int
	spawnFn := func(issueNumber int) error {
		spawnedIssue = issueNumber
		return nil
	}

	result, err := Run(Config{MaxWorkers: 3}, Deps{
		Board:         board,
		ActiveWorkers: 1,
		SpawnFunc:     spawnFn,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Dispatched {
		t.Errorf("expected dispatch, got reason: %q", result.Reason)
	}
	if result.IssueNumber != 42 {
		t.Errorf("expected issue 42, got %d", result.IssueNumber)
	}
	if spawnedIssue != 42 {
		t.Errorf("expected spawn called with 42, got %d", spawnedIssue)
	}
}

func TestRun_transitionsItemAfterSpawn(t *testing.T) {
	t.Parallel()

	board := &project.BoardSummary{
		Columns: map[string][]project.Item{
			"Scoped": {{Number: 42, Title: "Ready", ID: "PVTI_42"}},
		},
	}

	var transitionedID, transitionedStatus string
	transitionFn := func(itemID, status string) error {
		transitionedID = itemID
		transitionedStatus = status
		return nil
	}

	_, err := Run(Config{MaxWorkers: 3}, Deps{
		Board:          board,
		ActiveWorkers:  0,
		SpawnFunc:      func(_ int) error { return nil },
		TransitionFunc: transitionFn,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transitionedID != "PVTI_42" {
		t.Errorf("expected transition for PVTI_42, got %q", transitionedID)
	}
	if transitionedStatus != "In Progress" {
		t.Errorf("expected status 'In Progress', got %q", transitionedStatus)
	}
}

func TestRun_doesNotTransitionOnSpawnFailure(t *testing.T) {
	t.Parallel()

	board := &project.BoardSummary{
		Columns: map[string][]project.Item{
			"Scoped": {{Number: 42, Title: "Ready", ID: "PVTI_42"}},
		},
	}

	transitionCalled := false
	transitionFn := func(_, _ string) error {
		transitionCalled = true
		return nil
	}

	_, _ = Run(Config{MaxWorkers: 3}, Deps{
		Board:          board,
		ActiveWorkers:  0,
		SpawnFunc:      func(_ int) error { return fmt.Errorf("fail") },
		TransitionFunc: transitionFn,
	})

	if transitionCalled {
		t.Error("expected transition NOT to be called on spawn failure")
	}
}
