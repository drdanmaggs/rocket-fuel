package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/session"
	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/drdanmaggs/rocket-fuel/internal/worker"
	"github.com/spf13/cobra"
)

var workCmd = &cobra.Command{
	Use:   "work <issue-number-or-url>",
	Short: "Spawn a worker on a GitHub issue",
	Long: `Creates a git worktree, opens a new tmux window, and launches
Claude Code to work on the specified GitHub issue.

The issue is read from GitHub and routed to the appropriate skill
based on its labels (workflow:tdd, workflow:bug-fix, etc.).`,
	Args: cobra.ExactArgs(1),
	RunE: runWork,
}

func init() {
	rootCmd.AddCommand(workCmd)
}

func runWork(cmd *cobra.Command, args []string) error {
	issueRef := args[0]

	// Parse issue number from URL or plain number.
	issueNumber, err := parseIssueRef(issueRef)
	if err != nil {
		return err
	}

	// Fetch issue from GitHub.
	issue, err := fetchIssue(issueNumber)
	if err != nil {
		return fmt.Errorf("fetch issue: %w", err)
	}

	// Get repo root.
	repoDir, err := repoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	// Ensure worktrees directory exists.
	worktreesDir := repoDir + "/.worktrees"
	if err := os.MkdirAll(worktreesDir, 0o755); err != nil {
		return fmt.Errorf("create .worktrees dir: %w", err)
	}

	tm := tmux.New()
	cfg := worker.SpawnConfig{
		RepoDir:     repoDir,
		SessionName: session.DefaultSessionName,
	}

	if err := worker.Spawn(tm, cfg, *issue); err != nil {
		return fmt.Errorf("spawn worker: %w", err)
	}

	skill := skillFromLabels(issue.Labels)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Spawned worker-%d on issue #%d: %s (skill: %s)\n",
		issue.Number, issue.Number, issue.Title, skill)

	return nil
}

func parseIssueRef(ref string) (int, error) {
	// Handle URLs like https://github.com/owner/repo/issues/42
	if strings.Contains(ref, "/issues/") {
		parts := strings.Split(ref, "/issues/")
		if len(parts) == 2 {
			ref = parts[1]
		}
	}

	// Strip leading #
	ref = strings.TrimPrefix(ref, "#")

	n, err := strconv.Atoi(ref)
	if err != nil {
		return 0, fmt.Errorf("invalid issue reference %q: must be a number or GitHub URL", ref)
	}
	return n, nil
}

type ghIssue struct {
	Number int       `json:"number"`
	Title  string    `json:"title"`
	Body   string    `json:"body"`
	Labels []ghLabel `json:"labels"`
}

type ghLabel struct {
	Name string `json:"name"`
}

func fetchIssue(number int) (*worker.Issue, error) {
	out, err := exec.CommandContext(context.Background(), "gh", "issue", "view", strconv.Itoa(number), "--json", "number,title,body,labels").Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue view: %w", err)
	}

	var gh ghIssue
	if err := json.Unmarshal(out, &gh); err != nil {
		return nil, fmt.Errorf("parse issue JSON: %w", err)
	}

	labels := make([]string, len(gh.Labels))
	for i, l := range gh.Labels {
		labels[i] = l.Name
	}

	return &worker.Issue{
		Number: gh.Number,
		Title:  gh.Title,
		Body:   gh.Body,
		Labels: labels,
	}, nil
}

func skillFromLabels(labels []string) string {
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
	return "/epc"
}
