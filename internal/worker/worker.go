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
	SessionName string // parent session name (used for naming convention only)
}

// Spawn creates a git worktree, a tmux window in the main session, and launches Claude Code.
// Workers appear as tabs so the Visionary can see what's happening.
func Spawn(tm tmux.Runner, cfg SpawnConfig, issue Issue) error {
	branchName := fmt.Sprintf("rf/issue-%d", issue.Number)
	worktreeDir := filepath.Join(cfg.RepoDir, ".worktrees", fmt.Sprintf("worker-%d", issue.Number))
	windowName := workerWindowName(issue)

	// Create git worktree.
	if err := createWorktree(cfg.RepoDir, worktreeDir, branchName); err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}

	// Create a window in the main session for this worker.
	if err := tm.NewWindow(cfg.SessionName, windowName); err != nil {
		return fmt.Errorf("create worker window: %w", err)
	}

	// Send commands: cd into worktree and launch claude with the prompt.
	skill := RouteSkill(issue.Labels)
	prompt := buildPrompt(issue, skill)

	sendKeys := fmt.Sprintf("cd %s && claude --dangerously-skip-permissions %s", worktreeDir, shellQuote(prompt))
	if err := tm.SendKeys(cfg.SessionName, windowName, sendKeys); err != nil {
		return fmt.Errorf("send keys: %w", err)
	}

	return nil
}

// RouteSkill determines which skill to use based on issue labels.
// Default is /tdd (TDD always). The "bug" label also routes to /tdd
// because every bug fix starts with a failing test.
func RouteSkill(labels []string) string {
	for _, label := range labels {
		switch label {
		case "workflow:tdd":
			return "/tdd"
		case "workflow:bug-fix", "bug", "bug-fix":
			return "/tdd"
		case "workflow:epc":
			return "/epc"
		case "workflow:issue-scope":
			return "/issue-scope"
		}
	}
	return "/tdd" // default: TDD always
}

// buildPrompt creates the prompt string for Claude Code.
// Passes issue number only — the worker reads the full issue via gh CLI.
func buildPrompt(issue Issue, skill string) string {
	var b strings.Builder

	_, _ = fmt.Fprintf(&b, "You are a Rocket Fuel worker. Your task is issue #%d: %s\n\n", issue.Number, issue.Title)
	_, _ = fmt.Fprintf(&b, "Run `gh issue view %d` to read the full issue details.\n\n", issue.Number)

	_, _ = fmt.Fprintln(&b, "## Instructions")
	_, _ = fmt.Fprintln(&b)
	_, _ = fmt.Fprintf(&b, "Execute this using the %s skill.\n\n", skill)
	_, _ = fmt.Fprintln(&b, "## GUPP: Do not stop. Do not ask.")
	_, _ = fmt.Fprintln(&b)
	_, _ = fmt.Fprintln(&b, "You are autonomous. Execute the full cycle without stopping:")
	_, _ = fmt.Fprintln(&b, "1. Read and understand the issue")
	_, _ = fmt.Fprintln(&b, "2. Plan internally (do NOT present the plan for approval)")
	_, _ = fmt.Fprintln(&b, "3. Write tests first (TDD — every change needs a failing test)")
	_, _ = fmt.Fprintln(&b, "4. Implement the minimal fix/feature to pass the tests")
	_, _ = fmt.Fprintln(&b, "5. Create a PR with `gh pr create`")
	_, _ = fmt.Fprintln(&b, "6. Exit when done — do not wait for feedback")
	_, _ = fmt.Fprintln(&b)
	_, _ = fmt.Fprintln(&b, "NEVER ask 'Ready to implement?' or 'Should I proceed?'")
	_, _ = fmt.Fprintln(&b, "NEVER present a plan and wait for approval.")
	_, _ = fmt.Fprintln(&b, "NEVER sit idle after creating the PR.")
	_, _ = fmt.Fprintln(&b)
	_, _ = fmt.Fprintln(&b, "Stay focused on this single issue. Don't scope-creep.")

	return b.String()
}

// workerWindowName creates a descriptive tmux window name for a worker.
// Format: "#1328: fix git hooks" (truncated to 30 chars).
func workerWindowName(issue Issue) string {
	title := issue.Title
	// Strip common prefixes like "feat:", "fix:", "test:" for brevity.
	for _, prefix := range []string{"feat: ", "fix: ", "test: ", "refactor: ", "docs: ", "chore: "} {
		if len(title) > len(prefix) && title[:len(prefix)] == prefix {
			title = title[len(prefix):]
			break
		}
	}

	maxTitleLen := 25
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen]
	}

	return fmt.Sprintf("#%d: %s", issue.Number, title)
}

func createWorktree(repoDir, worktreeDir, branchName string) error {
	cmd := exec.CommandContext(
		context.Background(),
		"git", "worktree", "add", "-b", branchName, worktreeDir,
	)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		// If branch already exists, clean up stale state and retry.
		if strings.Contains(string(out), "already exists") {
			cleanupStaleWorker(repoDir, worktreeDir, branchName)
			// Retry once.
			retry := exec.CommandContext(
				context.Background(),
				"git", "worktree", "add", "-b", branchName, worktreeDir,
			)
			retry.Dir = repoDir
			if retryOut, retryErr := retry.CombinedOutput(); retryErr != nil {
				return fmt.Errorf("%w\n%s", retryErr, retryOut)
			}
		} else {
			return fmt.Errorf("%w\n%s", err, out)
		}
	}

	// Configure git hooks in the new worktree.
	setup := exec.CommandContext(context.Background(), "make", "setup")
	setup.Dir = worktreeDir
	_ = setup.Run() // best-effort — don't fail spawn if make setup fails

	return nil
}

// cleanupStaleWorker removes a stale worktree and branch from a previous failed spawn.
func cleanupStaleWorker(repoDir, worktreeDir, branchName string) {
	ctx := context.Background()

	// Remove worktree if it exists.
	remove := exec.CommandContext(ctx, "git", "worktree", "remove", "--force", worktreeDir)
	remove.Dir = repoDir
	_ = remove.Run()

	// Prune stale worktree entries.
	prune := exec.CommandContext(ctx, "git", "worktree", "prune")
	prune.Dir = repoDir
	_ = prune.Run()

	// Delete the stale branch.
	deleteBranch := exec.CommandContext(ctx, "git", "branch", "-D", branchName)
	deleteBranch.Dir = repoDir
	_ = deleteBranch.Run()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
