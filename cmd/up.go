package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drdanmaggs/rocket-fuel/internal/launch"
	"github.com/drdanmaggs/rocket-fuel/internal/prime"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the Rocket Fuel tmux session",
	Long: `Creates a tmux session with Integrator and Dashboard windows,
launches Claude Code in the Integrator tab with full project context,
then attaches in control mode (-CC) so iTerm2 renders them as native tabs.

If a session already exists, attaches to it without relaunching.`,
	RunE: runUp,
}

func init() {
	upCmd.Flags().Bool("dry-run", false, "Create session but don't attach (for testing)")
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, _ []string) error {
	tm := tmux.New()
	sessionName := session.DefaultSessionName

	created, err := session.Setup(tm, sessionName)
	if err != nil {
		return fmt.Errorf("session setup failed: %w", err)
	}

	if created {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created session %q with windows: integrator, heartbeat, dashboard\n", sessionName)

		// Launch Claude Code in the integrator window with prime context.
		if launchErr := launchIntegrator(tm, sessionName); launchErr != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: could not launch integrator: %v\n", launchErr)
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Launched Claude Code in integrator tab.")
		}

		// Launch heartbeat loop in the heartbeat window.
		if err := tm.SendKeys(sessionName, "heartbeat", "rf heartbeat --loop"); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: could not launch heartbeat: %v\n", err)
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Launched heartbeat in background tab.")
		}

		// Launch status in the dashboard window.
		if err := tm.SendKeys(sessionName, "dashboard", "rf status"); err != nil {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: could not launch dashboard: %v\n", err)
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Launched status in dashboard tab.")
		}
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Attaching to existing session %q\n", sessionName)
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Dry run — not attaching.")
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Attaching with tmux -CC (iTerm2 control mode)...")

	// AttachCC replaces this process via exec — does not return on success.
	if err := tm.AttachCC(sessionName); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}

	return nil
}

func launchIntegrator(tm tmux.Runner, sessionName string) error {
	repoDir, err := statusRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	in := &prime.Input{
		RepoDir: repoDir,
		Branch:  primeCurrentBranch(),
	}

	// Load integrator prompt.
	promptPath := filepath.Join(repoDir, "prompts", "integrator.md")
	if data, readErr := os.ReadFile(promptPath); readErr == nil {
		in.IntegratorPrompt = string(data)
	}

	// Load board state (optional).
	if cfg, loadErr := loadProjectConfig(); loadErr == nil {
		board, fetchErr := project.FetchBoard(cfg.Owner, cfg.ProjectNumber)
		if fetchErr == nil {
			in.Board = board
		}
	}

	// Load worker status (optional).
	s, gatherErr := status.Gather(tm, sessionName, repoDir)
	if gatherErr == nil {
		in.Status = s
	}

	// Write context file and launch Claude.
	contextPath, err := launch.WritePrimeContext(repoDir, in)
	if err != nil {
		return err
	}

	launchCmd := launch.IntegratorCommand(contextPath)
	return tm.SendKeys(sessionName, "integrator", launchCmd)
}
