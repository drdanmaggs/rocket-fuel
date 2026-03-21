package cmd

import (
	"encoding/json"
	"os"

	"github.com/drdanmaggs/rocket-fuel/internal/dashboard"
	"github.com/spf13/cobra"
)

var stopFailureCmd = &cobra.Command{
	Use:    "handle-stop-failure",
	Hidden: true, // internal — called by StopFailure hook
	Short:  "Log API errors from Claude sessions",
	RunE:   runStopFailure,
}

func init() {
	rootCmd.AddCommand(stopFailureCmd)
}

func runStopFailure(_ *cobra.Command, _ []string) error {
	repoDir, err := repoRoot()
	if err != nil {
		return nil
	}

	// Read error info from stdin.
	var input struct {
		ErrorType    string `json:"error_type"`
		ErrorMessage string `json:"error_message"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		// Can't parse — log generic error.
		_ = dashboard.WriteActivity(repoDir, "StopFailure: unknown error")
		return nil
	}

	// Log to activity feed.
	msg := "StopFailure: " + input.ErrorType
	if input.ErrorMessage != "" {
		msg += " — " + input.ErrorMessage
	}
	_ = dashboard.WriteActivity(repoDir, msg)

	return nil
}
