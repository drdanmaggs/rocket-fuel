package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// testSocket is the isolated tmux socket name for this test process.
// Set by SetupTmuxSocket in TestMain, used by NewRealTmux.
var testSocket string

// SetupTmuxSocket creates an isolated tmux socket for the test process.
// Call this in TestMain for packages that need real tmux integration tests.
// Returns a cleanup function that kills the tmux server on that socket.
//
// Pattern (from gastown):
//
//	func TestMain(m *testing.M) {
//	    cleanup := testutil.SetupTmuxSocket()
//	    code := m.Run()
//	    cleanup()
//	    os.Exit(code)
//	}
func SetupTmuxSocket() func() {
	testSocket = fmt.Sprintf("rf-test-%d", os.Getpid())
	return func() {
		// Kill the tmux server on our isolated socket.
		_ = exec.CommandContext(context.Background(), "tmux", "-L", testSocket, "kill-server").Run() //nolint:errcheck // best-effort cleanup
		testSocket = ""
	}
}

// SkipIfNoTmux skips the test if tmux is not installed.
func SkipIfNoTmux(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed, skipping")
	}
}

// RealTmux wraps real tmux commands against an isolated socket.
// Use for integration tests that need to verify tmux actually works.
type RealTmux struct {
	socket string
}

// NewRealTmux creates a RealTmux using the test process's isolated socket.
// Requires SetupTmuxSocket to have been called in TestMain.
func NewRealTmux(t *testing.T) *RealTmux {
	t.Helper()
	SkipIfNoTmux(t)

	if testSocket == "" {
		t.Fatal("testutil.SetupTmuxSocket() must be called in TestMain before using NewRealTmux")
	}

	return &RealTmux{socket: testSocket}
}

// Run executes a tmux command against the isolated socket.
func (rt *RealTmux) Run(t *testing.T, args ...string) string {
	t.Helper()

	fullArgs := append([]string{"-L", rt.socket}, args...)
	cmd := exec.CommandContext(context.Background(), "tmux", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tmux %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}

	return strings.TrimSpace(string(out))
}

// RunExpectError executes a tmux command expecting failure.
func (rt *RealTmux) RunExpectError(t *testing.T, args ...string) string {
	t.Helper()

	fullArgs := append([]string{"-L", rt.socket}, args...)
	cmd := exec.CommandContext(context.Background(), "tmux", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected tmux %s to fail, but it succeeded:\n%s", strings.Join(args, " "), out)
	}

	return strings.TrimSpace(string(out))
}

// NewSession creates a tmux session and registers cleanup.
func (rt *RealTmux) NewSession(t *testing.T, name string) {
	t.Helper()

	rt.Run(t, "new-session", "-d", "-s", name)
	t.Cleanup(func() {
		// Best-effort kill — session may already be dead.
		_ = exec.CommandContext(context.Background(), "tmux", "-L", rt.socket, "kill-session", "-t", name).Run() //nolint:errcheck
	})
}

// HasSession checks if a session with the given name exists.
func (rt *RealTmux) HasSession(t *testing.T, name string) bool {
	t.Helper()

	fullArgs := []string{"-L", rt.socket, "has-session", "-t", name}
	cmd := exec.CommandContext(context.Background(), "tmux", fullArgs...)
	return cmd.Run() == nil
}

// ListWindows returns the window names for a session.
func (rt *RealTmux) ListWindows(t *testing.T, session string) []string {
	t.Helper()

	out := rt.Run(t, "list-windows", "-t", session, "-F", "#{window_name}")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// Socket returns the isolated socket name (for passing to subprocess env).
func (rt *RealTmux) Socket() string {
	return rt.socket
}

// RecordingTmux records tmux commands for unit tests where real tmux isn't needed.
// Use for testing orchestration logic (what commands WOULD be issued) without tmux.
type RecordingTmux struct {
	Commands []TmuxCommand
}

// TmuxCommand records a tmux invocation.
type TmuxCommand struct {
	Action string
	Args   []string
}

// NewRecordingTmux creates a new command recorder for unit tests.
func NewRecordingTmux() *RecordingTmux {
	return &RecordingTmux{}
}

// Record stores a tmux command invocation.
func (r *RecordingTmux) Record(action string, args ...string) {
	r.Commands = append(r.Commands, TmuxCommand{Action: action, Args: args})
}

// CommandCount returns the number of recorded commands.
func (r *RecordingTmux) CommandCount() int {
	return len(r.Commands)
}

// LastCommand returns the most recently recorded command.
// Returns an empty TmuxCommand if none recorded.
func (r *RecordingTmux) LastCommand() TmuxCommand {
	if len(r.Commands) == 0 {
		return TmuxCommand{}
	}
	return r.Commands[len(r.Commands)-1]
}

// HasCommand checks if any recorded command matches the given action.
func (r *RecordingTmux) HasCommand(action string) bool {
	for _, cmd := range r.Commands {
		if cmd.Action == action {
			return true
		}
	}
	return false
}
