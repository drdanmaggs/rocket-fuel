package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

const stuckThreshold = 20 * time.Minute

var checkStuckCmd = &cobra.Command{
	Use:    "check-stuck",
	Hidden: true, // internal — called by watchdog loop
	Short:  "Detect workers with no activity for 20+ minutes",
	RunE:   runCheckStuck,
}

func init() {
	rootCmd.AddCommand(checkStuckCmd)
}

func runCheckStuck(cmd *cobra.Command, _ []string) error {
	repoDir, err := repoRoot()
	if err != nil {
		return nil
	}

	activityDir := filepath.Join(repoDir, ".rocket-fuel", "activity")
	entries, err := os.ReadDir(activityDir)
	if err != nil {
		return nil // no activity dir = no workers tracked
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName
	now := time.Now()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name() // e.g., "worker-42"

		// Check if worker window still exists.
		if !tm.HasSession(sessionName) {
			continue
		}

		// Read last activity timestamp.
		data, err := os.ReadFile(filepath.Join(activityDir, name))
		if err != nil {
			continue
		}

		lastActivity, err := time.Parse(time.RFC3339, string(data))
		if err != nil {
			continue
		}

		idle := now.Sub(lastActivity)
		if idle > stuckThreshold {
			msg := fmt.Sprintf("[Watchdog] Worker %s appears stuck — no activity for %s. Investigate or kill.", name, idle.Round(time.Minute))
			_ = tm.SendKeys(sessionName, session.WindowIntegrator, msg)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "STUCK: %s (idle %s)\n", name, idle.Round(time.Minute))
		}
	}

	return nil
}
