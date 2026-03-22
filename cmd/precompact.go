package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/drdanmaggs/rocket-fuel/internal/hookutil"
	"github.com/drdanmaggs/rocket-fuel/internal/prime"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
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

	// Workers don't need to do anything on precompact yet.
	if role == hookutil.RoleWorker {
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
