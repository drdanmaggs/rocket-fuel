package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var mergeSafetyCmd = &cobra.Command{
	Use:    "check-merge-safety",
	Hidden: true, // internal — called by PreToolUse hook
	Short:  "Check if a PR is safe to merge (CI green, not draft)",
	RunE:   runMergeSafety,
}

func init() {
	rootCmd.AddCommand(mergeSafetyCmd)
}

func runMergeSafety(cmd *cobra.Command, _ []string) error {
	// Read the tool input from stdin (Claude Code sends it as JSON).
	var input struct {
		ToolInput struct {
			Command string `json:"command"`
		} `json:"tool_input"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		// If we can't parse input, allow the merge (fail-open for non-hook usage).
		return nil
	}

	// Extract PR number from the command.
	prNum := extractPRNumber(input.ToolInput.Command)
	if prNum == "" {
		// Can't determine PR — allow (fail-open).
		return nil
	}

	// Check PR state.
	out, err := exec.CommandContext(context.Background(),
		"gh", "pr", "view", prNum,
		"--json", "isDraft,statusCheckRollup",
	).Output()
	if err != nil {
		// Can't check PR — allow (fail-open).
		return nil
	}

	var pr struct {
		IsDraft bool `json:"isDraft"`
		Checks  []struct {
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"statusCheckRollup"`
	}
	if err := json.Unmarshal(out, &pr); err != nil {
		return nil
	}

	// Rule 1: draft PRs cannot be merged.
	if pr.IsDraft {
		fmt.Fprintln(os.Stderr, "PR is still a draft — undraft before merging")
		os.Exit(2)
	}

	// Rule 2: all CI checks must be complete.
	for _, check := range pr.Checks {
		if check.Status != "COMPLETED" {
			fmt.Fprintln(os.Stderr, "CI is still running — wait for all checks to complete")
			os.Exit(2)
		}
	}

	// Rule 3: no CI failures.
	for _, check := range pr.Checks {
		if check.Conclusion == "FAILURE" {
			fmt.Fprintln(os.Stderr, "CI has failures — do not merge until all checks pass")
			os.Exit(2)
		}
	}

	// All checks passed — allow merge.
	return nil
}

func extractPRNumber(command string) string {
	// Handle: gh pr merge 1345, gh pr merge 1345 --squash, etc.
	parts := strings.Fields(command)
	for i, p := range parts {
		if p == "merge" && i+1 < len(parts) {
			num := parts[i+1]
			if len(num) > 0 && num[0] >= '0' && num[0] <= '9' {
				return num
			}
		}
	}
	return ""
}
