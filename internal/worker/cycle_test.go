package worker

import (
	"strings"
	"testing"
)

func TestExtractIssueNumberReturnsIssueNumberFromWorktreeCwd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cwd       string
		wantNum   int
		wantError bool
	}{
		{
			name:      "extracts issue number from worktree path",
			cwd:       "/home/user/repo/.worktrees/worker-42",
			wantNum:   42,
			wantError: false,
		},
		{
			name:      "returns error when not in a worktree",
			cwd:       "/home/user/repo",
			wantNum:   0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ExtractIssueNumber(tt.cwd)

			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error for cwd %q, got nil", tt.cwd)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for cwd %q: %v", tt.cwd, err)
			}

			if got != tt.wantNum {
				t.Errorf("ExtractIssueNumber(%q) = %d, want %d", tt.cwd, got, tt.wantNum)
			}
		})
	}
}

func TestCycleWorkerReturnsScriptFromWorktreeCwdAndSessionName(t *testing.T) {
	t.Parallel()

	// Arrange: a valid worktree cwd and session name.
	cwd := "/home/user/repo/.worktrees/worker-42"
	sessionName := "rf-integrator"

	// Act: build the cycle script from the worktree cwd.
	script, err := CycleWorker(cwd, sessionName)
	// Assert: no error.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert: script targets the correct session and contains the issue number in window name.
	if !strings.Contains(script, sessionName) {
		t.Errorf("script should contain session name %q, got:\n%s", sessionName, script)
	}

	// Assert: script references issue 42 (extracted from worktree path).
	if !strings.Contains(script, "42") {
		t.Errorf("script should reference issue number 42, got:\n%s", script)
	}

	// Assert: script contains kill (C-c) and restart elements.
	if !strings.Contains(script, "C-c") {
		t.Errorf("script should contain 'C-c' to kill current process, got:\n%s", script)
	}

	if !strings.Contains(script, "send-keys") {
		t.Errorf("script should contain 'send-keys' tmux command, got:\n%s", script)
	}
}

func TestBuildCycleScriptReturnsBashScriptThatKillsAndRestartsClaude(t *testing.T) {
	t.Parallel()

	session := "rf-integrator"
	window := "#42: fix auth"
	restartCmd := "cd /tmp && claude --agent worker"

	script := BuildCycleScript(session, window, restartCmd)

	// Should sleep 2 seconds to let the hook return
	if !strings.Contains(script, "sleep 2") {
		t.Errorf("script should contain 'sleep 2', got:\n%s", script)
	}

	// Should send Ctrl-C to kill Claude
	if !strings.Contains(script, "send-keys") || !strings.Contains(script, "C-c") {
		t.Errorf("script should contain 'send-keys' and 'C-c' to kill Claude, got:\n%s", script)
	}

	// Should target the correct session:window
	expectedTarget := session + ":" + window
	if !strings.Contains(script, expectedTarget) {
		t.Errorf("script should contain target %q, got:\n%s", expectedTarget, script)
	}

	// Should contain the restart command
	if !strings.Contains(script, restartCmd) {
		t.Errorf("script should contain restart command %q, got:\n%s", restartCmd, script)
	}
}
