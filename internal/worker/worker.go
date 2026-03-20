// Package worker manages spawning and tracking Claude Code workers in git worktrees.
package worker

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// Issue holds the GitHub issue context needed to spawn a worker.
type Issue struct {
	Number int
	Title  string
	Body   string
	Labels []string
}

// SpawnConfig holds configuration for spawning a worker.
type SpawnConfig struct {
	RepoDir     string // root of the git repo
	SessionName string // tmux session to create window in
}

// Spawn creates a git worktree, tmux window, and launches Claude Code for an issue.
func Spawn(tm tmux.Runner, cfg SpawnConfig, issue Issue) error {
	branchName := fmt.Sprintf("rf/issue-%d", issue.Number)
	worktreeDir := filepath.Join(cfg.RepoDir, ".worktrees", fmt.Sprintf("worker-%d", issue.Number))
	windowName := fmt.Sprintf("worker-%d", issue.Number)

	// Create git worktree.
	if err := createWorktree(cfg.RepoDir, worktreeDir, branchName); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}

	// Create tmux window.
	if err := tm.NewWindow(cfg.SessionName, windowName); err != nil {
		return fmt.Errorf("create window: %w", err)
	}

	// Send commands to the new window: cd into worktree and launch claude.
	skill := routeSkill(issue.Labels)
	prompt := buildPrompt(issue, skill)

	sendKeys := fmt.Sprintf("cd %s && claude --prompt %s", worktreeDir, shellQuote(prompt))
	if err := tmuxSendKeys(tm, cfg.SessionName, windowName, sendKeys); err != nil {
		return fmt.Errorf("send keys: %w", err)
	}

	return nil
}

// routeSkill determines which skill to use based on issue labels.
func routeSkill(labels []string) string {
	for _, label := range labels {
		switch label {
		case "workflow:tdd":
			return "/tdd"
		case "workflow:bug-fix":
			return "/bug-fix"
		case "workflow:epc":
			return "/epc"
		case "workflow:issue-scope":
			return "/issue-scope"
		}
	}
	return "/epc" // default skill
}

// buildPrompt creates the prompt string for Claude Code.
func buildPrompt(issue Issue, skill string) string {
	var b strings.Builder

	_, _ = fmt.Fprintf(&b, "You are a Rocket Fuel worker. Your task is issue #%d: %s\n\n", issue.Number, issue.Title)

	if issue.Body != "" {
		_, _ = fmt.Fprintf(&b, "Issue description:\n%s\n\n", issue.Body)
	}

	_, _ = fmt.Fprintf(&b, "Execute this using the %s skill. When done, create a PR with `gh pr create`.\n", skill)
	_, _ = fmt.Fprintf(&b, "Stay focused on this single issue. Don't scope-creep.")

	return b.String()
}

func createWorktree(repoDir, worktreeDir, branchName string) error {
	cmd := exec.CommandContext(
		context.Background(),
		"git", "worktree", "add", "-b", branchName, worktreeDir,
	)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}

	// Configure git hooks in the new worktree.
	setup := exec.CommandContext(context.Background(), "make", "setup")
	setup.Dir = worktreeDir
	_ = setup.Run() // best-effort — don't fail spawn if make setup fails

	return nil
}

func tmuxSendKeys(tm tmux.Runner, session, window, keys string) error {
	return tm.SendKeys(session, window, keys)
}

func shellQuote(s string) string {
	// Single-quote the string, escaping any existing single quotes.
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
