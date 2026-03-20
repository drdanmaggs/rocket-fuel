package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultMaxWorkers = 3

var dispatchCmd = &cobra.Command{
	Use:   "dispatch",
	Short: "Dispatch a worker from the Scoped column",
	Long: `Checks the project board for Scoped items. If capacity allows,
spawns a worker on the highest-priority Scoped issue and moves it
to In Progress.`,
	RunE: runDispatch,
}

func init() {
	dispatchCmd.Flags().Bool("dry-run", false, "Show what would be dispatched without acting")
	rootCmd.AddCommand(dispatchCmd)
}

func runDispatch(cmd *cobra.Command, _ []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	result, err := runDispatchCycle(dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] %s\n", result.Reason)
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Reason)
	}

	return nil
}

// ghRunner executes gh CLI commands — the real implementation of project.GHRunner.
func ghRunner(args ...string) ([]byte, error) {
	return exec.CommandContext(context.Background(), "gh", args...).Output()
}

func loadMaxWorkers(repoDir string) int {
	data, err := os.ReadFile(filepath.Join(repoDir, ".rocket-fuel", "config.json"))
	if err != nil {
		return defaultMaxWorkers
	}

	var cfg struct {
		MaxWorkers int `json:"max_workers"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.MaxWorkers <= 0 {
		return defaultMaxWorkers
	}
	return cfg.MaxWorkers
}
