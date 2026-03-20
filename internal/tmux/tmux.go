// Package tmux provides an interface and implementation for tmux operations.
package tmux

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Runner defines the interface for tmux operations.
// Real implementation calls tmux CLI; test doubles record commands.
type Runner interface {
	HasSession(name string) bool
	NewSession(name string) error
	NewWindow(session, name string) error
	SelectWindow(session, window string) error
	KillSession(name string) error
	AttachCC(session string) error
}

// CLI implements Runner by calling the real tmux binary.
type CLI struct {
	socket string // optional: use -L socket for isolation
}

// New creates a CLI runner using the default tmux socket.
func New() *CLI {
	return &CLI{}
}

// NewWithSocket creates a CLI runner using a specific tmux socket.
// Used for testing (isolated socket) or multi-instance scenarios.
func NewWithSocket(socket string) *CLI {
	return &CLI{socket: socket}
}

// HasSession checks if a tmux session with the given name exists.
func (c *CLI) HasSession(name string) bool {
	return c.run("has-session", "-t", name) == nil
}

// NewSession creates a new detached tmux session.
func (c *CLI) NewSession(name string) error {
	return c.run("new-session", "-d", "-s", name)
}

// NewWindow creates a new window in the given session.
func (c *CLI) NewWindow(session, name string) error {
	return c.run("new-window", "-t", session, "-n", name)
}

// SelectWindow switches to the named window in the session.
func (c *CLI) SelectWindow(session, window string) error {
	return c.run("select-window", "-t", session+":"+window)
}

// KillSession destroys the named session.
func (c *CLI) KillSession(name string) error {
	return c.run("kill-session", "-t", name)
}

// AttachCC replaces the current process with tmux -CC attach.
// This hands control to tmux in control mode — iTerm2 will render
// tmux windows as native tabs.
//
// This function does not return on success (it execs).
func (c *CLI) AttachCC(session string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	args := []string{"tmux"}
	if c.socket != "" {
		args = append(args, "-L", c.socket)
	}
	args = append(args, "-CC", "attach", "-t", session)

	// Replace this process with tmux — does not return on success.
	return syscall.Exec(tmuxPath, args, os.Environ())
}

// SendKeys sends keystrokes to a specific window in a session.
func (c *CLI) SendKeys(session, window, keys string) error {
	return c.run("send-keys", "-t", session+":"+window, keys, "Enter")
}

func (c *CLI) run(args ...string) error {
	fullArgs := args
	if c.socket != "" {
		fullArgs = append([]string{"-L", c.socket}, args...)
	}

	cmd := exec.CommandContext(context.Background(), "tmux", fullArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}
