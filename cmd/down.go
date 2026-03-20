package cmd

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Tear down the Rocket Fuel tmux session",
	Long:  `Destroys the Rocket Fuel tmux session and all its windows.`,
	RunE:  runDown,
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func runDown(cmd *cobra.Command, _ []string) error {
	tm := tmux.New()
	sessionName := session.DefaultSessionName

	killed, err := session.Teardown(tm, sessionName)
	if err != nil {
		return fmt.Errorf("teardown failed: %w", err)
	}

	if killed {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Session %q destroyed.\n", sessionName)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No active session %q found.\n", sessionName)
	}

	return nil
}
