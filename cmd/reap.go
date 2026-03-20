package cmd

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
	"github.com/spf13/cobra"
)

var reapCmd = &cobra.Command{
	Use:   "reap",
	Short: "Clean up completed workers",
	Long: `Finds workers whose tmux windows have exited (Claude Code session ended)
and removes their git worktrees.`,
	RunE: runReap,
}

func init() {
	reapCmd.Flags().Bool("dry-run", false, "Show what would be cleaned without doing it")
	rootCmd.AddCommand(reapCmd)
}

func runReap(cmd *cobra.Command, _ []string) error {
	repoDir, err := gitRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	results, err := worker.Reap(tm, sessionName, repoDir)
	if err != nil {
		return fmt.Errorf("reap failed: %w", err)
	}

	if len(results) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No workers to reap.")
		return nil
	}

	for _, r := range results {
		if r.Reaped {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Reaped %s (%s)\n", r.WindowName, r.Reason)
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Skipped %s (%s)\n", r.WindowName, r.Reason)
		}
	}

	return nil
}
