package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/dispatch"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
)

// repoRoot returns the root directory of the current git repository.
func repoRoot() (string, error) {
	out, err := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repo: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// dispatchFailedIssues tracks issues that failed to spawn.
// Persists across heartbeat cycles to avoid infinite retries.
var dispatchFailedIssues = make(map[int]bool)

// runDispatchCycle executes one dispatch cycle: fetch board, check capacity, spawn.
// Used by both the dispatch command and the mission control loop.
func runDispatchCycle(dryRun bool) (*dispatch.Result, error) {
	repoDir, err := repoRoot()
	if err != nil {
		return nil, err
	}

	cfg, err := loadProjectConfig()
	if err != nil {
		return nil, fmt.Errorf("no project linked: %w", err)
	}

	board, err := project.FetchBoard(ghRunner, cfg.Owner, cfg.ProjectNumber)
	if err != nil {
		return nil, err
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	s, err := status.Gather(tm, sessionName, repoDir)
	if err != nil {
		return nil, err
	}

	activeWorkers := 0
	for _, w := range s.Workers {
		if w.WindowOpen {
			activeWorkers++
		}
	}

	maxWorkers := loadMaxWorkers(repoDir)

	spawnFn := func(issueNumber int) error {
		if dryRun {
			return nil
		}
		issue, fetchErr := fetchIssue(issueNumber)
		if fetchErr != nil {
			return fetchErr
		}
		return worker.Spawn(tm, worker.SpawnConfig{
			RepoDir:     repoDir,
			SessionName: sessionName,
		}, *issue)
	}

	transitionFn := func(itemID, targetStatus string) error {
		if dryRun {
			return nil
		}
		return project.TransitionItem(ghRunner, cfg.Owner, cfg.ProjectNumber, itemID, targetStatus)
	}

	return dispatch.Run(dispatch.Config{MaxWorkers: maxWorkers}, dispatch.Deps{
		Board:          board,
		ActiveWorkers:  activeWorkers,
		SpawnFunc:      spawnFn,
		TransitionFunc: transitionFn,
		FailedIssues:   dispatchFailedIssues,
	})
}
