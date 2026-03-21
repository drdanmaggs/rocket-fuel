//go:build integration

package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

var testSocket string

func TestMain(m *testing.M) {
	testSocket = fmt.Sprintf("rf-session-test-%d", os.Getpid())
	code := m.Run()
	// Best-effort cleanup of the isolated tmux server.
	_ = exec.CommandContext(context.Background(), "tmux", "-L", testSocket, "kill-server").Run()
	os.Exit(code)
}

// TestSetupCreatesAllWindows_Integration verifies that Setup creates a real
// tmux session with the correct windows (integrator, heartbeat, dashboard).
func TestSetupCreatesAllWindows_Integration(t *testing.T) {
	cli := tmux.NewWithSocket(testSocket)
	sessionName := "rf-e2e-setup-" + fmt.Sprintf("%d", os.Getpid())

	created, err := Setup(cli, sessionName)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(sessionName) })

	if !created {
		t.Error("expected session to be created")
	}

	// Verify session exists.
	if !cli.HasSession(sessionName) {
		t.Fatal("expected session to exist after Setup")
	}

	// Setup creates integrator (window 0 renamed) + mission-control.
	windows := listWindows(t, sessionName)

	if !contains(windows, "integrator") {
		t.Errorf("expected 'integrator' window, got: %v", windows)
	}
	if !contains(windows, "watchdog") {
		t.Errorf("expected 'mission-control' window, got: %v", windows)
	}
	if len(windows) != 2 {
		t.Errorf("expected exactly 2 windows, got %d: %v", len(windows), windows)
	}
}

// TestSetupIsIdempotent_Integration verifies that calling Setup on an
// existing session doesn't create duplicate windows.
func TestSetupIsIdempotent_Integration(t *testing.T) {
	cli := tmux.NewWithSocket(testSocket)
	sessionName := "rf-e2e-idem-" + fmt.Sprintf("%d", os.Getpid())

	_, err := Setup(cli, sessionName)
	if err != nil {
		t.Fatalf("first Setup failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(sessionName) })

	windowsBefore := listWindows(t, sessionName)

	// Second call should not create anything new.
	created, err := Setup(cli, sessionName)
	if err != nil {
		t.Fatalf("second Setup failed: %v", err)
	}
	if created {
		t.Error("expected second Setup to report session already existed")
	}

	windowsAfter := listWindows(t, sessionName)
	if len(windowsAfter) != len(windowsBefore) {
		t.Errorf("window count changed: before=%d after=%d", len(windowsBefore), len(windowsAfter))
	}
}

// TestSendKeysToWindow_Integration verifies that SendKeys actually delivers
// keystrokes to a real tmux window.
func TestSendKeysToWindow_Integration(t *testing.T) {
	cli := tmux.NewWithSocket(testSocket)
	sessionName := "rf-e2e-keys-" + fmt.Sprintf("%d", os.Getpid())

	_, err := Setup(cli, sessionName)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(sessionName) })

	// Send a simple echo command to the integrator window.
	err = cli.SendKeys(sessionName, "integrator", "echo ROCKET_FUEL_TEST_MARKER")
	if err != nil {
		t.Fatalf("SendKeys failed: %v", err)
	}

	// Retry pane capture — CI runners are slower, shell needs time to process.
	var out string
	for range 20 {
		time.Sleep(100 * time.Millisecond)
		out = tmuxRun(t, "capture-pane", "-t", sessionName+":integrator", "-p")
		if strings.Contains(out, "ROCKET_FUEL_TEST_MARKER") {
			return
		}
	}
	t.Errorf("expected test marker in pane output after 2s, got:\n%s", out)
}

// TestIntegratorWindowSelectedByDefault_Integration verifies that the
// integrator window is the active window after Setup.
func TestIntegratorWindowSelectedByDefault_Integration(t *testing.T) {
	cli := tmux.NewWithSocket(testSocket)
	sessionName := "rf-e2e-select-" + fmt.Sprintf("%d", os.Getpid())

	_, err := Setup(cli, sessionName)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(sessionName) })

	// Check which window is currently active.
	active := tmuxRun(t, "display-message", "-t", sessionName, "-p", "#{window_name}")
	if active != "integrator" {
		t.Errorf("expected integrator to be active window, got %q", active)
	}
}

func listWindows(t *testing.T, session string) []string {
	t.Helper()
	out := tmuxRun(t, "list-windows", "-t", session, "-F", "#{window_name}")
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

func tmuxRun(t *testing.T, args ...string) string {
	t.Helper()
	fullArgs := append([]string{"-L", testSocket}, args...)
	cmd := exec.CommandContext(context.Background(), "tmux", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tmux %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestSetupCreatesSplitPane verifies that the integrator window can be split
// into multiple panes after Setup completes.
func TestSetupCreatesSplitPane_Integration(t *testing.T) {
	cli := tmux.NewWithSocket(testSocket)
	sessionName := "rf-e2e-split-" + fmt.Sprintf("%d", os.Getpid())

	_, err := Setup(cli, sessionName)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(sessionName) })

	// Split the integrator window using -l (lines) instead of -p (percent)
	// since -p has compatibility issues with some tmux versions.
	if err := cli.Run("split-window", "-h", "-t", sessionName+":"+WindowIntegrator, "-l", "10", "sleep 30"); err != nil {
		t.Fatalf("split-window failed: %v", err)
	}

	// Verify 2 panes exist via list-panes.
	panes := tmuxRun(t, "list-panes", "-t", sessionName+":"+WindowIntegrator)
	if panes == "" {
		t.Fatal("expected pane list, got empty output")
	}

	paneCount := 0
	for _, line := range strings.Split(panes, "\n") {
		if line != "" {
			paneCount++
		}
	}

	if paneCount != 2 {
		t.Errorf("expected 2 panes after split, got %d", paneCount)
	}
}

// TestWorkerWindowNaming verifies that Spawn creates a window with the expected
// "#N: title" naming format.
func TestWorkerWindowNaming_Integration(t *testing.T) {
	cli := tmux.NewWithSocket(testSocket)
	sessionName := "rf-e2e-worker-" + fmt.Sprintf("%d", os.Getpid())

	_, err := Setup(cli, sessionName)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	t.Cleanup(func() { _ = cli.KillSession(sessionName) })

	// Simulate what Spawn does: create a window with "#<number>: <title>" format.
	windowName := "#128: fix database migration"
	if err := cli.NewWindow(sessionName, windowName); err != nil {
		t.Fatalf("NewWindow failed: %v", err)
	}

	// Verify the window exists and has the correct name.
	windows := listWindows(t, sessionName)
	if !contains(windows, windowName) {
		t.Errorf("expected window %q in list, got: %v", windowName, windows)
	}

	// Verify ListWorkerWindows would find this window.
	workerWindows := ListWorkerWindows(cli, sessionName)
	found := false
	for _, w := range workerWindows {
		if w == windowName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ListWorkerWindows to find %q, got: %v", windowName, workerWindows)
	}
}
