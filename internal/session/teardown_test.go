package session

import (
	"testing"
)

func TestTeardownKillsExistingSession(t *testing.T) {
	t.Parallel()

	tm := newMockRunner()
	tm.sessions["test-session"] = true

	killed, err := Teardown(tm, "test-session")
	if err != nil {
		t.Fatalf("Teardown failed: %v", err)
	}

	if !killed {
		t.Error("expected Teardown to report session was killed")
	}

	if tm.sessions["test-session"] {
		t.Error("expected session to be gone after Teardown")
	}
}

func TestTeardownNoopIfNoSession(t *testing.T) {
	t.Parallel()

	tm := newMockRunner()

	killed, err := Teardown(tm, "nonexistent")
	if err != nil {
		t.Fatalf("Teardown failed: %v", err)
	}

	if killed {
		t.Error("expected Teardown to report no session was killed")
	}
}
