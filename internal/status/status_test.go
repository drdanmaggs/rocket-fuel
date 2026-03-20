package status

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockRunner struct {
	sessions map[string]bool
	windows  map[string]map[string]bool
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		sessions: make(map[string]bool),
		windows:  make(map[string]map[string]bool),
	}
}

func (m *mockRunner) HasSession(name string) bool { return m.sessions[name] }
func (m *mockRunner) NewSession(_ string) error   { return nil }
func (m *mockRunner) NewWindow(_, _ string) error { return nil }
func (m *mockRunner) KillSession(_ string) error  { return nil }
func (m *mockRunner) AttachCC(_ string) error     { return nil }

func (m *mockRunner) SelectWindow(session, window string) error {
	if m.windows[session] != nil && m.windows[session][window] {
		return nil
	}
	return &mockErr{}
}

type mockErr struct{}

func (e *mockErr) Error() string { return "not found" }

func TestGatherWithNoWorkers(t *testing.T) {
	t.Parallel()

	tm := newMockRunner()
	tm.sessions["rocket-fuel"] = true
	repoDir := t.TempDir()

	s, err := Gather(tm, "rocket-fuel", repoDir)
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	if !s.SessionActive {
		t.Error("expected session to be active")
	}

	if len(s.Workers) != 0 {
		t.Errorf("expected 0 workers, got %d", len(s.Workers))
	}
}

func TestGatherWithWorkers(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	worktreesDir := filepath.Join(repoDir, ".worktrees")

	// Create two worker directories.
	for _, name := range []string{"worker-1", "worker-2"} {
		if err := os.MkdirAll(filepath.Join(worktreesDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	tm := newMockRunner()
	tm.sessions["rocket-fuel"] = true
	tm.windows["rocket-fuel"] = map[string]bool{"worker-1": true}

	s, err := Gather(tm, "rocket-fuel", repoDir)
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	if len(s.Workers) != 2 {
		t.Fatalf("expected 2 workers, got %d", len(s.Workers))
	}

	// worker-1 should be active (window exists).
	if !s.Workers[0].WindowOpen {
		t.Error("expected worker-1 window to be open")
	}

	// worker-2 should be done (no window).
	if s.Workers[1].WindowOpen {
		t.Error("expected worker-2 window to be closed")
	}
}

func TestGatherWithInactiveSession(t *testing.T) {
	t.Parallel()

	tm := newMockRunner()
	repoDir := t.TempDir()

	s, err := Gather(tm, "rocket-fuel", repoDir)
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	if s.SessionActive {
		t.Error("expected session to be inactive")
	}
}

func TestFormatOutput(t *testing.T) {
	t.Parallel()

	s := &Summary{
		SessionActive: true,
		Workers: []WorkerStatus{
			{Name: "worker-1", WindowOpen: true, Branch: "rf/issue-1"},
			{Name: "worker-2", WindowOpen: false, Branch: "rf/issue-2", HasPR: true},
		},
	}

	out := Format(s)

	checks := []string{
		"Session: ACTIVE",
		"Workers: 2",
		"worker-1",
		"(active)",
		"worker-2",
		"(done)",
		"[PR open]",
	}

	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("expected output to contain %q\n\nGot:\n%s", check, out)
		}
	}
}
