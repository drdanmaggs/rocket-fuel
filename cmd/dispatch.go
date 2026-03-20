package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/drdanmaggs/rocket-fuel/internal/dispatch"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
	"github.com/spf13/cobra"
)

const defaultMaxWorkers = 3

var dispatchCmd = &cobra.Command{
	Use:   "dispatch",
	Short: "Dispatch a worker from the Scoped column",
	Long: `Checks the project board for Scoped items. If capacity allows,
spawns a worker on the highest-priority Scoped issue and moves it
to In Progress.`,
	RunE: runDispatch,
}

func init() {
	dispatchCmd.Flags().Bool("dry-run", false, "Show what would be dispatched without acting")
	rootCmd.AddCommand(dispatchCmd)
}

func runDispatch(cmd *cobra.Command, _ []string) error {
	repoDir, err := statusRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	cfg, err := loadProjectConfig()
	if err != nil {
		return fmt.Errorf("no project linked: %w\nRun: rocket-fuel project link <project-url>", err)
	}

	board, err := project.FetchBoard(cfg.Owner, cfg.ProjectNumber)
	if err != nil {
		return fmt.Errorf("fetch board: %w", err)
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	s, err := status.Gather(tm, sessionName, repoDir)
	if err != nil {
		return fmt.Errorf("gather status: %w", err)
	}

	activeWorkers := 0
	for _, w := range s.Workers {
		if w.WindowOpen {
			activeWorkers++
		}
	}

	maxWorkers := loadMaxWorkers(repoDir)
	dryRun, _ := cmd.Flags().GetBool("dry-run")

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

	result, err := dispatch.Run(dispatch.Config{MaxWorkers: maxWorkers}, dispatch.Deps{
		Board:          board,
		ActiveWorkers:  activeWorkers,
		SpawnFunc:      spawnFn,
		TransitionFunc: transitionFn,
	})
	if err != nil {
		return err
	}

	if dryRun {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] %s\n", result.Reason)
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Reason)
	}

	return nil
}

// ghRunner executes gh CLI commands — the real implementation of project.GHRunner.
func ghRunner(args ...string) ([]byte, error) {
	return exec.CommandContext(context.Background(), "gh", args...).Output()
}

func loadMaxWorkers(repoDir string) int {
	data, err := os.ReadFile(filepath.Join(repoDir, ".rocket-fuel", "config.json"))
	if err != nil {
		return defaultMaxWorkers
	}

	var cfg struct {
		MaxWorkers int `json:"max_workers"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.MaxWorkers <= 0 {
		return defaultMaxWorkers
	}
	return cfg.MaxWorkers
}
