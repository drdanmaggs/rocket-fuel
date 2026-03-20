package worker

import (
	"os"
	"path/filepath"
	"testing"
)

// mockTmuxRunner for reap tests.
type mockTmuxRunner struct {
	sessions map[string]bool
	windows  map[string]map[string]bool // session -> window -> exists
}

func newMockTmuxRunner() *mockTmuxRunner {
	return &mockTmuxRunner{
		sessions: make(map[string]bool),
		windows:  make(map[string]map[string]bool),
	}
}

func (m *mockTmuxRunner) HasSession(name string) bool {
	return m.sessions[name]
}

func (m *mockTmuxRunner) NewSession(name string) error {
	m.sessions[name] = true
	if m.windows[name] == nil {
		m.windows[name] = make(map[string]bool)
	}
	return nil
}

func (m *mockTmuxRunner) NewWindow(session, name string) error {
	if m.windows[session] == nil {
		m.windows[session] = make(map[string]bool)
	}
	m.windows[session][name] = true
	return nil
}

func (m *mockTmuxRunner) SelectWindow(session, window string) error {
	if m.windows[session] != nil && m.windows[session][window] {
		return nil
	}
	return &mockSelectError{}
}

func (m *mockTmuxRunner) KillSession(name string) error {
	delete(m.sessions, name)
	delete(m.windows, name)
	return nil
}

func (m *mockTmuxRunner) AttachCC(_ string) error              { return nil }
func (m *mockTmuxRunner) SendKeys(_, _, _ string) error        { return nil }

type mockSelectError struct{}

func (e *mockSelectError) Error() string { return "window not found" }

func TestReapCleansUpCompletedWorkers(t *testing.T) {
	t.Parallel()

	// Set up a fake repo with a worktrees directory.
	repoDir := t.TempDir()
	worktreesDir := filepath.Join(repoDir, ".worktrees")

	if err := os.MkdirAll(filepath.Join(worktreesDir, "worker-42"), 0o755); err != nil {
		t.Fatal(err)
	}

	// tmux has no window for worker-42 (session ended).
	tm := newMockTmuxRunner()
	tm.sessions["rocket-fuel"] = true

	results, err := Reap(tm, "rocket-fuel", repoDir)
	if err != nil {
		t.Fatalf("Reap failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// The worktree dir was just a plain dir (not a real worktree),
	// so git worktree remove will fail, but we still get a result.
	r := results[0]
	if r.WindowName != "worker-42" {
		t.Errorf("expected window name 'worker-42', got %q", r.WindowName)
	}
}

func TestReapSkipsActiveWorkers(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	worktreesDir := filepath.Join(repoDir, ".worktrees")

	if err := os.MkdirAll(filepath.Join(worktreesDir, "worker-99"), 0o755); err != nil {
		t.Fatal(err)
	}

	// tmux HAS the window for worker-99 (still active).
	tm := newMockTmuxRunner()
	tm.sessions["rocket-fuel"] = true
	_ = tm.NewWindow("rocket-fuel", "worker-99")

	results, err := Reap(tm, "rocket-fuel", repoDir)
	if err != nil {
		t.Fatalf("Reap failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Reaped {
		t.Error("expected active worker to NOT be reaped")
	}
	if r.Reason != "window still active" {
		t.Errorf("expected reason 'window still active', got %q", r.Reason)
	}
}

func TestReapHandlesNoWorktreesDir(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	tm := newMockTmuxRunner()

	results, err := Reap(tm, "rocket-fuel", repoDir)
	if err != nil {
		t.Fatalf("Reap failed: %v", err)
	}

	if results != nil {
		t.Errorf("expected nil results when no worktrees dir, got %v", results)
	}
}
