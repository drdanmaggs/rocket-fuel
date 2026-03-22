package worker

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// ExtractIssueNumber extracts the issue number from a worker worktree directory path.
// The worktree directory should be named "worker-N" where N is the issue number.
// Returns the issue number and nil error on success, or 0 and an error if the path
// does not match the expected worktree naming convention.
func ExtractIssueNumber(cwd string) (int, error) {
	// Get the last component of the path (directory name)
	dirName := filepath.Base(cwd)

	// Check if it matches the pattern "worker-N"
	if !strings.HasPrefix(dirName, "worker-") {
		return 0, fmt.Errorf("not in a worker worktree: expected directory name matching 'worker-N', got %q", dirName)
	}

	// Extract the numeric part after "worker-"
	numberStr := strings.TrimPrefix(dirName, "worker-")

	// Parse as integer
	issueNum, err := strconv.Atoi(numberStr)
	if err != nil {
		return 0, fmt.Errorf("invalid worker directory name %q: %w", dirName, err)
	}

	return issueNum, nil
}

// CycleWorker builds a bash script to restart a worker in a tmux window.
// It extracts the issue number from the worktree cwd, determines the window name
// (using the directory basename), builds a restart command, and delegates to BuildCycleScript.
func CycleWorker(cwd, sessionName string) (string, error) {
	// Extract issue number from worktree path
	issueNum, err := ExtractIssueNumber(cwd)
	if err != nil {
		return "", err
	}

	// Use directory basename as window name (e.g., "worker-42")
	windowName := filepath.Base(cwd)

	// Build restart command
	restartCmd := fmt.Sprintf(
		"cd %s && claude --agent worker --dangerously-skip-permissions 'Continue working on issue #%d. Run gh issue view %d to read the full issue. Check git log for your progress so far.'",
		cwd, issueNum, issueNum,
	)

	// Build and return the cycle script
	script := BuildCycleScript(sessionName, windowName, restartCmd)
	return script, nil
}

// BuildCycleScript builds a bash script that kills and restarts Claude in a tmux window.
// It sleeps briefly to allow the hook to return, sends Ctrl-C to terminate the current process,
// and then sends the restart command.
func BuildCycleScript(sessionName, windowName, restartCmd string) string {
	target := fmt.Sprintf("%s:%s", sessionName, windowName)
	return fmt.Sprintf(
		"sleep 2 && tmux send-keys -t '%s' C-c && sleep 1 && tmux send-keys -t '%s' '%s' Enter",
		target, target, restartCmd,
	)
}
