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
	Long:  `Destroys both the integrator and mission control sessions. The rocket lands.`,
	RunE:  runLand,
}

func init() {
	rootCmd.AddCommand(landCmd)
}

func runLand(cmd *cobra.Command, _ []string) error {
	tm := tmux.New()

	killed, err := session.TeardownAll(tm, session.DefaultSessionName)
	if err != nil {
		return fmt.Errorf("landing failed: %w", err)
	}

	if killed {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Rocket landed. All sessions destroyed.")
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active sessions found.")
	}

	return nil
}
