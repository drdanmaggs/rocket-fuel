package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
	"github.com/spf13/cobra"
)

var sessionEndedCmd = &cobra.Command{
	Use:    "session-ended",
	Hidden: true, // internal — called by SessionEnd hook
	Short:  "Handle Claude session ending (worker completion, integrator crash)",
	RunE:   runSessionEnded,
}

func init() {
	rootCmd.AddCommand(sessionEndedCmd)
}

func runSessionEnded(_ *cobra.Command, _ []string) error {
	repoDir, err := repoRoot()
	if err != nil {
		return nil // not in a repo, nothing to do
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	// Reap any completed workers.
	results, err := worker.Reap(tm, sessionName, repoDir, worker.ReapConfig{})
	if err != nil {
		return nil
	}

	// Nudge the Integrator about completed workers.
	for _, r := range results {
		if r.Reaped {
			msg := buildCompletionNudge(r)
			_ = tm.SendKeys(sessionName, session.WindowIntegrator, msg)
		}
	}

	return nil
}

func buildCompletionNudge(r worker.ReapResult) string {
	name := r.WindowName
	issueNum := strings.TrimPrefix(name, "worker-")

	prInfo := checkWorkerPR("rf/issue-" + issueNum)
	if prInfo != "" {
		return fmt.Sprintf("[Watchdog] Worker #%s completed. %s. Review and update the board.", issueNum, prInfo)
	}
	return fmt.Sprintf("[Watchdog] Worker #%s completed (no PR found). Check if work was finished.", issueNum)
}

func checkWorkerPR(branch string) string {
	out, err := exec.CommandContext(context.Background(),
		"gh", "pr", "list", "--head", branch, "--json", "number,title", "--limit", "1",
	).Output()
	if err != nil {
		return ""
	}

	var prs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}
	if err := json.Unmarshal(out, &prs); err != nil || len(prs) == 0 {
		return ""
	}

	return fmt.Sprintf("PR #%d: %s", prs[0].Number, prs[0].Title)
}
