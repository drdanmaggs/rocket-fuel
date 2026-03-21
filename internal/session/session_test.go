package session

import (
	"testing"
)

// mockRunner is a test double for tmux.Runner that records calls.
type mockRunner struct {
	sessions map[string]bool
	windows  map[string][]string
	calls    []string
	failOn   string // if set, return error when this method is called
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		sessions: make(map[string]bool),
		windows:  make(map[string][]string),
	}
}

func (m *mockRunner) HasSession(name string) bool {
	m.calls = append(m.calls, "HasSession:"+name)
	return m.sessions[name]
}

func (m *mockRunner) HasWindow(session, window string) bool {
	m.calls = append(m.calls, "HasWindow:"+session+":"+window)
	for _, w := range m.windows[session] {
		if w == window {
			return true
		}
	}
	return false
}

func (m *mockRunner) NewSession(name string) error {
	m.calls = append(m.calls, "NewSession:"+name)
	if m.failOn == "NewSession" {
		return errMock
	}
	m.sessions[name] = true
	return nil
}

func (m *mockRunner) NewWindow(session, name string) error {
	m.calls = append(m.calls, "NewWindow:"+session+":"+name)
	if m.failOn == "NewWindow" {
		return errMock
	}
	m.windows[session] = append(m.windows[session], name)
	return nil
}

func (m *mockRunner) SelectWindow(session, window string) error {
	m.calls = append(m.calls, "SelectWindow:"+session+":"+window)
	if m.failOn == "SelectWindow" {
		return errMock
	}
	return nil
}

func (m *mockRunner) KillSession(name string) error {
	m.calls = append(m.calls, "KillSession:"+name)
	delete(m.sessions, name)
	return nil
}

func (m *mockRunner) AttachCC(_ string) error {
	// Not called by Setup — AttachCC is handled by the cmd layer.
	return nil
}

func (m *mockRunner) SendKeys(session, window, keys string) error {
	m.calls = append(m.calls, "SendKeys:"+session+":"+window+":"+keys)
	return nil
}

func (m *mockRunner) ListWindowNames(session string) ([]string, error) {
	m.calls = append(m.calls, "ListWindowNames:"+session)
	return m.windows[session], nil
}

func (m *mockRunner) RenameWindow(session, target, newName string) error {
	m.calls = append(m.calls, "RenameWindow:"+session+":"+target+":"+newName)
	return nil
}

func (m *mockRunner) SplitPane(session, window, direction string, _ int, _ string) error {
	m.calls = append(m.calls, "SplitPane:"+session+":"+window+":"+direction)
	return nil
}

func (m *mockRunner) Run(_ ...string) error { return nil }

var errMock = &mockError{}

type mockError struct{}

func (e *mockError) Error() string { return "mock error" }

func TestSetupCreatesSession(t *testing.T) {
	t.Parallel()

	tm := newMockRunner()

	created, err := Setup(tm, "test-session")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if !created {
		t.Error("expected Setup to report session was created")
	}

	if !tm.sessions["test-session"] {
		t.Error("expected session to exist")
	}

	// Setup creates mission-control as an additional window.
	windows := tm.windows["test-session"]
	if len(windows) != 1 {
		t.Errorf("expected 1 additional window (mission-control), got %d: %v", len(windows), windows)
	}
	if len(windows) > 0 && windows[0] != WindowMissionCtrl {
		t.Errorf("expected window %q, got %q", WindowMissionCtrl, windows[0])
	}
}

func TestSetupIsIdempotent(t *testing.T) {
	t.Parallel()

	tm := newMockRunner()
	tm.sessions["existing"] = true

	created, err := Setup(tm, "existing")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if created {
		t.Error("expected Setup to report session already existed")
	}

	// Should not have created any windows.
	if len(tm.windows["existing"]) != 0 {
		t.Errorf("expected no windows created, got %v", tm.windows["existing"])
	}
}

func TestSetupCleansUpOnSessionFailure(t *testing.T) {
	t.Parallel()

	tm := newMockRunner()
	tm.failOn = "NewSession"

	_, err := Setup(tm, "fail-session")
	if err == nil {
		t.Fatal("expected Setup to fail when NewSession fails")
	}
}
