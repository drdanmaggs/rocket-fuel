package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	fns := buildMissionControlFuncs(dryRun)

	if !loop {
		result, err := missioncontrol.RunCycle(fns)
		if err != nil {
			return err
		}
		printCycleResult(cmd, result)
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

	missioncontrol.Loop(ctx, interval, fns, logFn)

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
		for _, r := range results {
			if r.Reaped {
				reaped++
			}
		}
		return fmt.Sprintf("reaped %d worker(s)", reaped), nil
	}

	return missioncontrol.Funcs{
		Dispatch: dispatchFn,
		Reap:     reapFn,
	}
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
