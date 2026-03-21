package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose Rocket Fuel installation and configuration",
	Long:  `Checks hooks, tools, session state, and project configuration.`,
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Rocket Fuel Doctor")
	_, _ = fmt.Fprintln(out)

	allGood := true

	// 1. Check tools
	_, _ = fmt.Fprintln(out, "Tools:")
	for _, tool := range []string{"tmux", "gh", "git", "claude"} {
		if _, err := exec.LookPath(tool); err != nil {
			_, _ = fmt.Fprintf(out, "  ✗ %s — not found\n", tool)
			allGood = false
		} else {
			_, _ = fmt.Fprintf(out, "  ✓ %s\n", tool)
		}
	}
	_, _ = fmt.Fprintln(out)

	// 2. Check repo
	repoDir, err := repoRoot()
	if err != nil {
		_, _ = fmt.Fprintln(out, "Repo: ✗ not in a git repository")
		return nil
	}
	_, _ = fmt.Fprintf(out, "Repo: ✓ %s\n", filepath.Base(repoDir))
	_, _ = fmt.Fprintln(out)

	// 3. Check project config
	cfg, err := loadProjectConfig()
	if err != nil {
		_, _ = fmt.Fprintln(out, "Project: ✗ no project linked")
		allGood = false
	} else {
		_, _ = fmt.Fprintf(out, "Project: ✓ #%d (owner: %s)\n", cfg.ProjectNumber, cfg.Owner)
	}
	_, _ = fmt.Fprintln(out)

	// 4. Check hooks
	_, _ = fmt.Fprintln(out, "Hooks (.claude/settings.json):")
	settingsPath := filepath.Join(repoDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		_, _ = fmt.Fprintln(out, "  ✗ settings.json not found — run rf launch to install hooks")
		allGood = false
	} else {
		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			_, _ = fmt.Fprintln(out, "  ✗ settings.json is invalid JSON")
			allGood = false
		} else {
			hooks, _ := settings["hooks"].(map[string]interface{})
			expectedHooks := []string{
				"SessionStart",
				"PreCompact",
				"PostToolUse",
				"SessionEnd",
				"Stop",
				"StopFailure",
				"PreToolUse",
			}
			for _, hook := range expectedHooks {
				if _, ok := hooks[hook]; ok {
					_, _ = fmt.Fprintf(out, "  ✓ %s\n", hook)
				} else {
					_, _ = fmt.Fprintf(out, "  ✗ %s — missing\n", hook)
					allGood = false
				}
			}
		}
	}
	_, _ = fmt.Fprintln(out)

	// 5. Check hook commands exist
	_, _ = fmt.Fprintln(out, "Hook commands:")
	hookCmds := []string{
		"rf prime",
		"rf should-continue",
		"rf session-ended",
		"rf record-activity",
		"rf handle-stop-failure",
		"rf check-merge-safety",
	}
	rfPath, _ := exec.LookPath("rf")
	if rfPath == "" {
		rfPath = filepath.Join(os.Getenv("HOME"), "go", "bin", "rf")
	}
	for _, hc := range hookCmds {
		if _, err := os.Stat(rfPath); err == nil {
			_, _ = fmt.Fprintf(out, "  ✓ %s\n", hc)
		} else {
			_, _ = fmt.Fprintf(out, "  ✗ %s — rf binary not found\n", hc)
			allGood = false
		}
	}
	_, _ = fmt.Fprintln(out)

	// 6. Check session state
	_, _ = fmt.Fprintln(out, "Session:")
	tm := tmux.New()
	if tm.HasSession(session.DefaultSessionName) {
		_, _ = fmt.Fprintf(out, "  ✓ %s active\n", session.DefaultSessionName)

		// Check windows
		if windows, err := tm.ListWindowNames(session.DefaultSessionName); err == nil {
			for _, w := range windows {
				_, _ = fmt.Fprintf(out, "    - %s\n", w)
			}
		}
	} else {
		_, _ = fmt.Fprintf(out, "  - %s not running\n", session.DefaultSessionName)
	}
	_, _ = fmt.Fprintln(out)

	// 7. Summary
	if allGood {
		_, _ = fmt.Fprintln(out, "All checks passed.")
	} else {
		_, _ = fmt.Fprintln(out, "Some checks failed. Run rf launch to fix hook installation.")
	}

	return nil
}
