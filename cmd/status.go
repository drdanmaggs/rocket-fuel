package cmd

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current Rocket Fuel status",
	Long:  `Displays session state, active workers, branches, and PR status.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, _ []string) error {
	repoDir, err := repoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	s, err := status.Gather(tm, sessionName, repoDir)
	if err != nil {
		return fmt.Errorf("gather status: %w", err)
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), status.Format(s))
	return nil
}
