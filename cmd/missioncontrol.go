package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/drdanmaggs/rocket-fuel/internal/dashboard"
	"github.com/drdanmaggs/rocket-fuel/internal/missioncontrol"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
	"github.com/spf13/cobra"
)

var missionControlCmd = &cobra.Command{
	Use:    "mission-control",
	Hidden: true, // internal command — launched by rf launch, not user-facing
	Short:  "Run dispatch + reap cycle",
	Long: `Executes one dispatch + reap cycle. With --loop, runs continuously
on a configurable interval. The background process that keeps the machine running.`,
	RunE: runMissionControl,
}

func init() {
	missionControlCmd.Flags().Bool("loop", false, "Run continuously")
	missionControlCmd.Flags().Duration("interval", 3*time.Minute, "Loop interval (requires --loop)")
	missionControlCmd.Flags().Bool("dry-run", false, "Show what would happen without acting")
	rootCmd.AddCommand(missionControlCmd)
}

func runMissionControl(cmd *cobra.Command, _ []string) error {
	loop, _ := cmd.Flags().GetBool("loop")
	interval, _ := cmd.Flags().GetDuration("interval")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	repoDir, err := repoRoot()
	if err != nil {
		return err
	}

	fns := buildMissionControlFuncs(dryRun)

	if !loop {
		result, err := missioncontrol.RunCycle(fns)
		if err != nil {
			return err
		}
		printCycleResult(cmd, result)
		recordCycleActivity(repoDir, result)
		return nil
	}

	// Loop mode: run until SIGINT/SIGTERM.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Mission Control active (interval: %s, dry-run: %v)\n", interval, dryRun)

	logFn := func(msg string) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), msg)
	}

	missioncontrol.LoopWithActivity(ctx, interval, fns, logFn, repoDir)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Mission Control offline.")
	return nil
}

func buildMissionControlFuncs(dryRun bool) missioncontrol.Funcs {
	dispatchFn := func() (string, error) {
		result, err := runDispatchCycle(dryRun)
		if err != nil {
			return "", err
		}
		return result.Reason, nil
	}

	reapFn := func() (string, error) {
		if dryRun {
			return "[dry-run] reap skipped", nil
		}

		repoDir, err := repoRoot()
		if err != nil {
			return "", err
		}

		tm := tmux.New()
		sessionName := session.DefaultSessionName

		results, err := worker.Reap(tm, sessionName, repoDir)
		if err != nil {
			return "", err
		}

		if len(results) == 0 {
			return "nothing to reap", nil
		}

		reaped := 0
		var nudges []string
		for _, r := range results {
			if r.Reaped {
				reaped++
				// Build nudge message for the integrator.
				nudges = append(nudges, buildReapNudge(r))
			}
		}

		// Nudge the integrator about completed workers.
		if len(nudges) > 0 {
			for _, msg := range nudges {
				_ = tm.SendKeys(sessionName, session.WindowIntegrator, msg)
			}
		}

		return fmt.Sprintf("reaped %d worker(s)", reaped), nil
	}

	return missioncontrol.Funcs{
		Dispatch: dispatchFn,
		Reap:     reapFn,
	}
}

// buildReapNudge creates a message to send to the Integrator about a reaped worker.
func buildReapNudge(r worker.ReapResult) string {
	// Extract issue number from window name (worker-42 or #42: title).
	name := r.WindowName
	issueNum := strings.TrimPrefix(name, "worker-")

	// Check if there's a PR for this worker's branch.
	branchName := "rf/issue-" + issueNum
	prInfo := checkPRForBranch(branchName)

	if prInfo != "" {
		return fmt.Sprintf("[Mission Control] Worker #%s completed. %s. Review and update the board.", issueNum, prInfo)
	}
	return fmt.Sprintf("[Mission Control] Worker #%s completed (no PR found). Check if work was finished.", issueNum)
}

// checkPRForBranch checks if a PR exists for the given branch.
func checkPRForBranch(branch string) string {
	out, err := exec.CommandContext(context.Background(),
		"gh", "pr", "list", "--head", branch, "--json", "number,title,url", "--limit", "1",
	).Output()
	if err != nil {
		return ""
	}

	var prs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(out, &prs); err != nil || len(prs) == 0 {
		return ""
	}

	return fmt.Sprintf("PR #%d: %s", prs[0].Number, prs[0].Title)
}

func printCycleResult(cmd *cobra.Command, result *missioncontrol.CycleResult) {
	if result.DispatchErr != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dispatch error: %v\n", result.DispatchErr)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dispatch: %s\n", result.DispatchResult)
	}

	if result.ReapErr != nil {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "reap error: %v\n", result.ReapErr)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "reap: %s\n", result.ReapResult)
	}
}

func recordCycleActivity(repoDir string, result *missioncontrol.CycleResult) {
	if result.DispatchErr != nil {
		_ = dashboard.WriteActivity(repoDir, fmt.Sprintf("dispatch error: %v", result.DispatchErr))
	} else if result.DispatchResult != "" {
		_ = dashboard.WriteActivity(repoDir, fmt.Sprintf("dispatch: %s", result.DispatchResult))
	}

	if result.ReapErr != nil {
		_ = dashboard.WriteActivity(repoDir, fmt.Sprintf("reap error: %v", result.ReapErr))
	} else if result.ReapResult != "" {
		_ = dashboard.WriteActivity(repoDir, fmt.Sprintf("reap: %s", result.ReapResult))
	}
}
