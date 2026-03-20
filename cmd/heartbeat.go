package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drdanmaggs/rocket-fuel/internal/dispatch"
	"github.com/drdanmaggs/rocket-fuel/internal/heartbeat"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Run dispatch + reap cycle",
	Long: `Executes one dispatch + reap cycle. With --loop, runs continuously
on a configurable interval. The dumb, reliable background process.`,
	RunE: runHeartbeat,
}

func init() {
	heartbeatCmd.Flags().Bool("loop", false, "Run continuously")
	heartbeatCmd.Flags().Duration("interval", 3*time.Minute, "Loop interval (requires --loop)")
	heartbeatCmd.Flags().Bool("dry-run", false, "Show what would happen without acting")
	rootCmd.AddCommand(heartbeatCmd)
}

func runHeartbeat(cmd *cobra.Command, _ []string) error {
	loop, _ := cmd.Flags().GetBool("loop")
	interval, _ := cmd.Flags().GetDuration("interval")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	fns := buildHeartbeatFuncs(dryRun)

	if !loop {
		result, err := heartbeat.RunCycle(fns)
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

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Heartbeat starting (interval: %s, dry-run: %v)\n", interval, dryRun)

	logFn := func(msg string) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), msg)
	}

	heartbeat.Loop(ctx, interval, fns, logFn)

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Heartbeat stopped.")
	return nil
}

func buildHeartbeatFuncs(dryRun bool) heartbeat.Funcs {
	dispatchFn := func() (string, error) {
		repoDir, err := statusRepoRoot()
		if err != nil {
			return "", err
		}

		cfg, err := loadProjectConfig()
		if err != nil {
			return "no project linked", nil
		}

		board, err := project.FetchBoard(cfg.Owner, cfg.ProjectNumber)
		if err != nil {
			return "", err
		}

		tm := tmux.New()
		sessionName := session.DefaultSessionName

		s, err := status.Gather(tm, sessionName, repoDir)
		if err != nil {
			return "", err
		}

		activeWorkers := 0
		for _, w := range s.Workers {
			if w.WindowOpen {
				activeWorkers++
			}
		}

		maxWorkers := loadMaxWorkers(repoDir)

		spawnFn := func(issueNumber int) error {
			if dryRun {
				return nil
			}
			issue, fetchErr := fetchIssue(issueNumber)
			if fetchErr != nil {
				return fetchErr
			}
			return worker.Spawn(tm, worker.SpawnConfig{
				RepoDir:     repoDir,
				SessionName: sessionName,
			}, *issue)
		}

		result, err := dispatch.Run(dispatch.Config{MaxWorkers: maxWorkers}, dispatch.Deps{
			Board:         board,
			ActiveWorkers: activeWorkers,
			SpawnFunc:     spawnFn,
		})
		if err != nil {
			return "", err
		}

		return result.Reason, nil
	}

	reapFn := func() (string, error) {
		if dryRun {
			return "[dry-run] reap skipped", nil
		}

		repoDir, err := statusRepoRoot()
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

	return heartbeat.Funcs{
		Dispatch: dispatchFn,
		Reap:     reapFn,
	}
}

func printCycleResult(cmd *cobra.Command, result *heartbeat.CycleResult) {
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
