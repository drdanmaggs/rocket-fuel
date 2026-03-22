package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/drdanmaggs/rocket-fuel/internal/hookutil"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var shouldContinueCmd = &cobra.Command{
	Use:    "should-continue",
	Hidden: true, // internal — called by Stop hook
	Short:  "Check if the Integrator should keep working",
	RunE:   runShouldContinue,
}

func init() {
	rootCmd.AddCommand(shouldContinueCmd)
}

func runShouldContinue(_ *cobra.Command, _ []string) error {
	return runShouldContinueWith(os.Stdin)
}

func runShouldContinueWith(input io.Reader) error {
	// If the caller is a Worker, allow stop immediately (don't check board or workers).
	role := hookutil.DetectRole(input)
	if role == hookutil.RoleWorker {
		return nil
	}

	repoDir, err := repoRoot()
	if err != nil {
		// Not in a repo — allow stop.
		return nil
	}

	// Check board for Ready items.
	cfg, err := loadProjectConfig()
	if err == nil {
		board, fetchErr := project.FetchBoard(ghRunner, cfg.Owner, cfg.ProjectNumber)
		if fetchErr == nil {
			next := project.NextReady(board)
			if next != nil {
				// Check capacity.
				tm := tmux.New()
				s, gatherErr := status.Gather(tm, session.DefaultSessionName, repoDir)
				if gatherErr == nil {
					activeWorkers := 0
					for _, w := range s.Workers {
						if w.WindowOpen {
							activeWorkers++
						}
					}
					maxWorkers := loadMaxWorkers(repoDir)
					if activeWorkers < maxWorkers {
						fmt.Fprintln(os.Stderr, "Ready items on the board with capacity available. Continue operational duties: dispatch, monitor workers, review PRs.")
						os.Exit(2)
					}
				}
			}
		}
	}

	// Check for active workers that might complete soon.
	tm := tmux.New()
	s, err := status.Gather(tm, session.DefaultSessionName, repoDir)
	if err == nil && len(s.Workers) > 0 {
		for _, w := range s.Workers {
			if w.WindowOpen {
				fmt.Fprintln(os.Stderr, "Workers are still active. Monitor their progress, check for completions, review PRs.")
				os.Exit(2)
			}
		}
	}

	// Nothing to do — allow stop.
	return nil
}
