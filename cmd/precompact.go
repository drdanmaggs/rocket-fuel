package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/drdanmaggs/rocket-fuel/internal/hookutil"
	"github.com/drdanmaggs/rocket-fuel/internal/prime"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
	"github.com/spf13/cobra"
)

var precompactCmd = &cobra.Command{
	Use:    "precompact",
	Hidden: true,
	Short:  "Handle PreCompact hook — re-prime or cycle worker",
	RunE:   runPrecompact,
}

func init() {
	rootCmd.AddCommand(precompactCmd)
}

func runPrecompact(cmd *cobra.Command, _ []string) error {
	return runPrecompactWith(os.Stdin, cmd.OutOrStdout())
}

func runPrecompactWith(input io.Reader, out io.Writer) error {
	// Detect role from input (usually Claude Code hook JSON).
	role := hookutil.DetectRole(input)

	// Workers get session cycling — kill and restart with fresh context.
	if role == hookutil.RoleWorker {
		cwd, err := os.Getwd()
		if err != nil {
			return nil // can't determine worktree, skip cycling
		}

		script, err := worker.CycleWorker(cwd, session.DefaultSessionName)
		if err != nil {
			return nil // not in a worker worktree, skip
		}

		// Spawn background process — same pattern as spawnDashboardSplit.
		cmd := exec.CommandContext(context.Background(), "bash", "-c", script)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		_ = cmd.Start()

		return nil
	}

	// Integrators re-prime their context (same as prime command).
	repoDir, err := repoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	in := &prime.Input{
		RepoDir: repoDir,
		Branch:  primeCurrentBranch(),
	}

	// Load board state (optional — project may not be linked).
	if cfg, loadErr := loadProjectConfig(); loadErr == nil {
		board, fetchErr := project.FetchBoard(ghRunner, cfg.Owner, cfg.ProjectNumber)
		if fetchErr == nil {
			in.Board = board
		}
	}

	// Load worker status.
	tm := tmux.New()
	s, gatherErr := status.Gather(tm, session.DefaultSessionName, repoDir)
	if gatherErr == nil {
		in.Status = s
	}

	_, _ = fmt.Fprint(out, prime.Build(in))
	return nil
}
