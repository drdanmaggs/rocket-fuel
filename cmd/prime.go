package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/drdanmaggs/rocket-fuel/internal/hookutil"
	"github.com/drdanmaggs/rocket-fuel/internal/prime"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var primeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Output Integrator context (board + workers + repo)",
	Long: `Gathers board state, active workers, and repo context into a single
markdown document. Used to prime the Integrator's Claude Code session.

Output is suitable for piping to 'claude --prompt-file'.`,
	RunE: runPrime,
}

func init() {
	rootCmd.AddCommand(primeCmd)
}

func runPrime(cmd *cobra.Command, _ []string) error {
	return runPrimeWith(os.Stdin, cmd.OutOrStdout())
}

func runPrimeWith(input io.Reader, out io.Writer) error {
	// Detect role from input (usually Claude Code hook JSON).
	role := hookutil.DetectRole(input)

	repoDir, err := repoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	in := &prime.Input{
		RepoDir: repoDir,
		Branch:  primeCurrentBranch(),
	}

	// Workers get repo context only (no board, no workers).
	// Integrators get everything.
	if role == hookutil.RoleIntegrator {
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
	}

	_, _ = fmt.Fprint(out, prime.Build(in))
	return nil
}

func primeCurrentBranch() string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "git", "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
