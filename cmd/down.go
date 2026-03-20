package cmd

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var landCmd = &cobra.Command{
	Use:   "land",
	Short: "Shut down the Rocket Fuel session",
	Long:  `Destroys the Rocket Fuel tmux session and all its windows. The rocket lands.`,
	RunE:  runLand,
}

func init() {
	rootCmd.AddCommand(landCmd)
}

func runLand(cmd *cobra.Command, _ []string) error {
	tm := tmux.New()
	sessionName := session.DefaultSessionName

	killed, err := session.Teardown(tm, sessionName)
	if err != nil {
		return fmt.Errorf("landing failed: %w", err)
	}

	if killed {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Rocket landed. Session %q destroyed.\n", sessionName)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No active session %q found.\n", sessionName)
	}

	return nil
}
