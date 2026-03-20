package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/drdanmaggs/rocket-fuel/internal/launch"
	"github.com/drdanmaggs/rocket-fuel/internal/prime"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/selfupdate"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "launch",
	Short: "Start the Rocket Fuel tmux session",
	Long: `Creates a tmux session with the Integrator, launches Claude Code,
and attaches in control mode (-CC) so iTerm2 renders a native tab.

Mission control runs in a separate background session.
If a session already exists, attaches to it without relaunching.`,
	RunE: runUp,
}

func init() {
	upCmd.Flags().Bool("dry-run", false, "Create session but don't attach (for testing)")
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Pre-flight: must be in a git repo.
	if _, err := repoRoot(); err != nil {
		return fmt.Errorf("not in a git repository — cd into your project first")
	}

	// Self-update: check if binary is stale, rebuild if needed.
	selfUpdate(out)

	tm := tmux.New()
	sessionName := session.DefaultSessionName

	// Show launch banner with project info.
	printLaunchBanner(out)

	created, err := session.Setup(tm, sessionName)
	if err != nil {
		return fmt.Errorf("session setup failed: %w", err)
	}

	if created {
		// Launch Claude Code in the integrator window with prime context.
		if launchErr := launchIntegrator(tm, sessionName); launchErr != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not launch integrator: %v\n", launchErr)
		}

		// Start mission control in a separate detached session.
		if mcCreated, mcErr := session.SetupMissionControl(tm); mcErr != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not start mission control: %v\n", mcErr)
		} else if mcCreated {
			_ = tm.SendKeys(session.MissionControlSession, session.WindowMissionCtrl, "rf mission-control --loop")
		}

		_, _ = fmt.Fprintln(out)
	} else {
		_, _ = fmt.Fprintf(out, "  Reattaching to existing session\n\n")
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		_, _ = fmt.Fprintln(out, "Dry run — not attaching.")
		return nil
	}

	// AttachCC replaces this process via exec — does not return on success.
	if err := tm.AttachCC(sessionName); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}

	return nil
}

func printLaunchBanner(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Rocket Fuel")
	_, _ = fmt.Fprintln(w)

	if repoDir, err := repoRoot(); err == nil {
		_, _ = fmt.Fprintf(w, "  Repo:    %s\n", filepath.Base(repoDir))
	}

	if cfg, err := loadProjectConfig(); err == nil {
		if board, fetchErr := project.FetchBoard(cfg.Owner, cfg.ProjectNumber); fetchErr == nil && board.ProjectTitle != "" {
			_, _ = fmt.Fprintf(w, "  Project: %s (#%d)\n", board.ProjectTitle, cfg.ProjectNumber)
		} else {
			_, _ = fmt.Fprintf(w, "  Project: #%d (owner: %s)\n", cfg.ProjectNumber, cfg.Owner)
		}
	} else {
		_, _ = fmt.Fprintln(w, "  Project: none (will auto-discover)")
	}

	_, _ = fmt.Fprintln(w)
}

func launchIntegrator(tm tmux.Runner, sessionName string) error {
	repoDir, err := repoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	in := &prime.Input{
		RepoDir: repoDir,
		Branch:  primeCurrentBranch(),
	}

	promptPath := filepath.Join(repoDir, "prompts", "integrator.md")
	if data, readErr := os.ReadFile(promptPath); readErr == nil {
		in.IntegratorPrompt = string(data)
	}

	if cfg, loadErr := loadProjectConfig(); loadErr == nil {
		board, fetchErr := project.FetchBoard(cfg.Owner, cfg.ProjectNumber)
		if fetchErr == nil {
			in.Board = board
		}
	}

	s, gatherErr := status.Gather(tm, sessionName, repoDir)
	if gatherErr == nil {
		in.Status = s
	}

	contextPath, err := launch.WritePrimeContext(repoDir, in)
	if err != nil {
		return err
	}

	launchCmd := launch.IntegratorCommand(contextPath)
	return tm.SendKeys(sessionName, "integrator", launchCmd)
}

func selfUpdate(w io.Writer) {
	binaryPath, err := os.Executable()
	if err != nil {
		return
	}

	result, err := selfupdate.Check(SourceDir, Version, binaryPath)
	if err != nil || result == nil {
		return
	}

	if result.Updated {
		_, _ = fmt.Fprintf(w, "  Updated rf: %s -> %s\n", result.OldVersion, result.NewVersion)
		_ = syscall.Exec(binaryPath, os.Args, os.Environ())
	}
}
