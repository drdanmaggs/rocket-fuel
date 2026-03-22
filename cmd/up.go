package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/drdanmaggs/rocket-fuel/internal/launch"
	rfplugin "github.com/drdanmaggs/rocket-fuel/internal/plugin"
	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/projects"
	"github.com/drdanmaggs/rocket-fuel/internal/selfupdate"
	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "launch",
	Short: "Start the Rocket Fuel tmux session",
	Long: `Creates a tmux session with the Integrator, launches Claude Code,
and attaches in control mode (-CC) so iTerm2 renders a native tab.

Claude gets its context via a SessionStart hook that runs rf prime.
Mission control runs in a separate background session.`,
	RunE: runUp,
}

func init() {
	upCmd.Flags().Bool("dry-run", false, "Create session but don't attach (for testing)")
	rootCmd.AddCommand(upCmd)
}

// checkTmux verifies that tmux is installed and available on PATH.
func checkTmux() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux is required but not installed. Install with: brew install tmux (macOS) or apt install tmux (Linux)")
	}
	return nil
}

func runUp(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	// Pre-flight: tmux must be installed.
	if err := checkTmux(); err != nil {
		return err
	}

	// Pre-flight: must be in a git repo.
	repoDir, err := repoRoot()
	if err != nil {
		// Try to load from project registry.
		reg, regErr := projects.LoadRegistry()
		if regErr != nil || len(reg) == 0 {
			// Registry empty — discover repos automatically.
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				return fmt.Errorf("not in a git repository — cd into your project first")
			}
			reg = projects.DiscoverProjects(homeDir)
			if len(reg) == 0 {
				return fmt.Errorf("no git repositories found — cd into your project and run rf launch")
			}
		}

		// Show available projects and prompt user to pick one.
		selected, pickErr := pickProject(out, reg)
		if pickErr != nil {
			return fmt.Errorf("project selection failed: %w", pickErr)
		}

		// Change to the selected project directory.
		if err := os.Chdir(selected.Path); err != nil {
			return fmt.Errorf("could not cd into %s: %w", selected.Path, err)
		}

		// Now verify we're in a git repo.
		repoDir, err = repoRoot()
		if err != nil {
			return fmt.Errorf("selected project is not a git repository: %w", err)
		}
	}

	// Extract Claude Code plugin (agents, skills) to ~/.claude/plugins/rocket-fuel/.
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		pluginDir := filepath.Join(homeDir, ".claude", "plugins", "rocket-fuel")
		if err := rfplugin.ExtractPlugin(pluginDir); err != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not extract plugin: %v\n", err)
		}
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

	// Always ensure hooks are up to date (even on reattach).
	if err := launch.EnsureClaudeSettings(repoDir); err != nil {
		_, _ = fmt.Fprintf(out, "  Warning: could not set up Claude hooks: %v\n", err)
	}

	if created {
		// Launch Claude Code in the integrator window.
		launchCmd := launch.IntegratorCommand()
		if err := tm.SendKeys(sessionName, session.WindowIntegrator, launchCmd); err != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not launch integrator: %v\n", err)
		}

		// Split the integrator window AFTER tmux -CC attaches.
		spawnDashboardSplit(sessionName)

		// Launch watchdog in its window.
		if err := tm.SendKeys(sessionName, session.WindowWatchdog, "rf watchdog --loop"); err != nil {
			_, _ = fmt.Fprintf(out, "  Warning: could not launch watchdog: %v\n", err)
		}

		_, _ = fmt.Fprintln(out)
	} else {
		_, _ = fmt.Fprintf(out, "  Reattaching to existing session\n\n")
	}

	// Remember this project for future launches.
	repoName := filepath.Base(repoDir)
	if err := projects.SaveProject(repoDir, repoName); err != nil {
		_, _ = fmt.Fprintf(out, "  Warning: could not save project to registry: %v\n", err)
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

// pickProject prompts the user to select a project from the registry.
func pickProject(w io.Writer, registry []projects.Project) (*projects.Project, error) {
	_, _ = fmt.Fprintln(w, "Rocket Fuel")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "  Recent projects:")
	_, _ = fmt.Fprintln(w)

	for i, p := range registry {
		_, _ = fmt.Fprintf(w, "  %d) %s (%s)\n", i+1, p.Name, p.Path)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprint(w, "  Select a project (number): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(registry) {
		return nil, fmt.Errorf("invalid selection")
	}

	return &registry[choice-1], nil
}

func printLaunchBanner(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Rocket Fuel")
	_, _ = fmt.Fprintln(w)

	if repoDir, err := repoRoot(); err == nil {
		_, _ = fmt.Fprintf(w, "  Repo:    %s\n", filepath.Base(repoDir))
	}

	if cfg, err := loadProjectConfig(); err == nil {
		if board, fetchErr := project.FetchBoard(ghRunner, cfg.Owner, cfg.ProjectNumber); fetchErr == nil && board.ProjectTitle != "" {
			_, _ = fmt.Fprintf(w, "  Project: %s (#%d)\n", board.ProjectTitle, cfg.ProjectNumber)
		} else {
			_, _ = fmt.Fprintf(w, "  Project: #%d (owner: %s)\n", cfg.ProjectNumber, cfg.Owner)
		}
	} else {
		_, _ = fmt.Fprintln(w, "  Project: none (will auto-discover)")
	}

	_, _ = fmt.Fprintln(w)
}

func selfUpdate(w io.Writer) {
	binaryPath, err := os.Executable()
	if err != nil {
		_, _ = fmt.Fprintf(w, "  Self-update check failed: %v\n", err)
		return
	}

	result, err := selfupdate.Check(SourceDir, Version, binaryPath)
	if err != nil {
		_, _ = fmt.Fprintf(w, "  Self-update check failed: %v\n", err)
		return
	}

	if result == nil {
		return
	}

	if result.Skipped != "" {
		return
	}

	if result.Updated {
		_, _ = fmt.Fprintf(w, "  Updated rf: %s -> %s\n", result.OldVersion, result.NewVersion)
		_ = syscall.Exec(binaryPath, os.Args, os.Environ())
	}
}

// spawnDashboardSplit starts a background process that waits for -CC to attach,
// then splits the integrator window with the dashboard.
func spawnDashboardSplit(sessionName string) {
	script := fmt.Sprintf(
		`sleep 3 && tmux split-window -t %s:%s -h -p 30 'export PATH="$HOME/go/bin:$PATH" && rf dashboard' && tmux select-pane -t %s:%s.0`,
		sessionName, session.WindowIntegrator,
		sessionName, session.WindowIntegrator,
	)

	cmd := exec.CommandContext(context.Background(), "bash", "-c", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	_ = cmd.Start()
}
