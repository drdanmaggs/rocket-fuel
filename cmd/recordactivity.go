package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var recordActivityCmd = &cobra.Command{
	Use:    "record-activity",
	Hidden: true, // internal — called by PostToolUse hook
	Short:  "Record worker activity timestamp for stuck detection",
	RunE:   runRecordActivity,
}

func init() {
	rootCmd.AddCommand(recordActivityCmd)
}

func runRecordActivity(_ *cobra.Command, _ []string) error {
	repoDir, err := repoRoot()
	if err != nil {
		return nil // not in repo, skip
	}

	// Determine which worker we are by checking the cwd.
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	// Extract worker number from worktree path (.worktrees/worker-N).
	workerDir := filepath.Base(cwd)
	if len(workerDir) < 7 {
		return nil // not in a worker worktree
	}

	// Write timestamp to .rocket-fuel/activity/<worker-name>
	activityDir := filepath.Join(repoDir, ".rocket-fuel", "activity")
	if err := os.MkdirAll(activityDir, 0o755); err != nil {
		return nil
	}

	activityFile := filepath.Join(activityDir, workerDir)
	timestamp := time.Now().Format(time.RFC3339)
	return os.WriteFile(activityFile, []byte(timestamp), 0o644)
}
