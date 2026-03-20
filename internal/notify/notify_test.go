package notify

import (
	"errors"
	"testing"
)

type mockRunner struct {
	selectedWindow string
	selectErr      error
}

func (m *mockRunner) HasSession(_ string) bool    { return true }
func (m *mockRunner) NewSession(_ string) error   { return nil }
func (m *mockRunner) NewWindow(_, _ string) error { return nil }
func (m *mockRunner) KillSession(_ string) error  { return nil }
func (m *mockRunner) AttachCC(_ string) error              { return nil }
func (m *mockRunner) SendKeys(_, _, _ string) error        { return nil }

func (m *mockRunner) SelectWindow(_, window string) error {
	if m.selectErr != nil {
		return m.selectErr
	}
	m.selectedWindow = window
	return nil
}

func TestSurfaceSelectsWindow(t *testing.T) {
	t.Parallel()

	tm := &mockRunner{}

	err := Surface(tm, "rocket-fuel", "worker-42", "")
	if err != nil {
		t.Fatalf("Surface failed: %v", err)
	}

	if tm.selectedWindow != "worker-42" {
		t.Errorf("expected selected window 'worker-42', got %q", tm.selectedWindow)
	}
}

func TestSurfaceReturnsErrorOnSelectFailure(t *testing.T) {
	t.Parallel()

	tm := &mockRunner{selectErr: errors.New("no such window")}

	err := Surface(tm, "rocket-fuel", "nonexistent", "test")
	if err == nil {
		t.Fatal("expected Surface to fail when SelectWindow fails")
	}
}

func TestSurfaceWithEmptyMessageSkipsNotification(t *testing.T) {
	t.Parallel()

	tm := &mockRunner{}

	// No notification sent when message is empty — just window switch.
	err := Surface(tm, "rocket-fuel", "integrator", "")
	if err != nil {
		t.Fatalf("Surface failed: %v", err)
	}

	if tm.selectedWindow != "integrator" {
		t.Errorf("expected 'integrator', got %q", tm.selectedWindow)
	}
}
