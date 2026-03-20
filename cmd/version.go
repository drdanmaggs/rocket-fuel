package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// SourceDir is set at build time via ldflags — path to the rocket-fuel source repo.
var SourceDir = ""

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Rocket Fuel",
	Run: func(cmd *cobra.Command, _ []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "rf %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
