package cmd

import (
	"fmt"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/notify"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var surfaceCmd = &cobra.Command{
	Use:   "surface <window-name> [message]",
	Short: "Bring a tmux window to the foreground",
	Long: `Switches the active iTerm2 tab (via tmux-CC) to the named window
and optionally sends a macOS notification.

This is the Integrator's primary communication channel to the Visionary.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSurface,
}

func init() {
	rootCmd.AddCommand(surfaceCmd)
}

func runSurface(cmd *cobra.Command, args []string) error {
	windowName := args[0]

	var message string
	if len(args) > 1 {
		message = strings.Join(args[1:], " ")
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	if err := notify.Surface(tm, sessionName, windowName, message); err != nil {
		return fmt.Errorf("surface failed: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Surfaced %s\n", windowName)
	return nil
}
