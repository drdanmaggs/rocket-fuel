package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drdanmaggs/rocket-fuel/internal/dashboard"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:    "dashboard",
	Hidden: true, // internal — launched by rf launch in the split pane
	Short:  "Live status dashboard for the split pane",
	RunE:   runDashboard,
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Render immediately.
	renderDashboard(cmd)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			renderDashboard(cmd)
		}
	}
}

func renderDashboard(cmd *cobra.Command) {
	repoDir, err := repoRoot()
	if err != nil {
		return
	}

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	s, err := status.Gather(tm, sessionName, repoDir)
	if err != nil {
		return
	}

	// Read recent activity log.
	activityLines := dashboard.ReadActivity(repoDir, 5)

	// Read pending merges.
	pendingMerges := dashboard.ListPending(repoDir)

	// Clear screen and render.
	_, _ = fmt.Fprint(cmd.OutOrStdout(), "\033[2J\033[H")
	data := &dashboard.Data{
		Summary:          s,
		ActivityLog:      activityLines,
		PendingApprovals: pendingMerges,
	}
	_, _ = fmt.Fprint(cmd.OutOrStdout(), dashboard.Render(data))
}
