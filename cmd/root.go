// Package cmd implements the rocket-fuel CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rocket-fuel",
	Short: "Visionary/Integrator multi-agent orchestrator",
	Long:  `Rocket Fuel is a multi-agent orchestrator that composes tmux-CC, GitHub Projects, Claude Code skills, and git worktrees into a Visionary/Integrator workflow.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
