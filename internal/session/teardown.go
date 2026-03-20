package session

import (
	"fmt"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// Teardown destroys the Rocket Fuel tmux session.
// Returns true if a session was killed, false if none existed.
func Teardown(tm tmux.Runner, sessionName string) (bool, error) {
	if !tm.HasSession(sessionName) {
		return false, nil
	}

	if err := tm.KillSession(sessionName); err != nil {
		return false, fmt.Errorf("kill session: %w", err)
	}

	return true, nil
}
