package cmd

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the Rocket Fuel tmux session",
	Long: `Creates a tmux session with Integrator and Dashboard windows,
then attaches in control mode (-CC) so iTerm2 renders them as native tabs.

If a session already exists, attaches to it.`,
	RunE: runUp,
}

func init() {
	upCmd.Flags().Bool("dry-run", false, "Create session but don't attach (for testing)")
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, _ []string) error {
	tm := tmux.New()
	sessionName := session.DefaultSessionName

	created, err := session.Setup(tm, sessionName)
	if err != nil {
		return fmt.Errorf("session setup failed: %w", err)
	}

	if created {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created session %q with windows: integrator, dashboard\n", sessionName)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Attaching to existing session %q\n", sessionName)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Dry run — not attaching.")
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Attaching with tmux -CC (iTerm2 control mode)...")

	// AttachCC replaces this process via exec — does not return on success.
	if err := tm.AttachCC(sessionName); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}

	return nil
}
