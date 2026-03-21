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
	HasWindow(session, window string) bool
	NewSession(name string) error
	NewWindow(session, name string) error
	SelectWindow(session, window string) error
	KillSession(name string) error
	AttachCC(session string) error
	SendKeys(session, window, keys string) error
	ListWindowNames(session string) ([]string, error)
	RenameWindow(session, target, newName string) error
	SplitPane(session, window, direction string, percent int, command string) error
	Run(args ...string) error
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

// HasWindow checks if a window with the given name exists in the session.
// Unlike SelectWindow, this does NOT change the active window — it's read-only.
func (c *CLI) HasWindow(session, window string) bool {
	out, err := c.output("list-windows", "-t", session, "-F", "#{window_name}")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if line == window {
			return true
		}
	}
	return false
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

// Run executes an arbitrary tmux command against the socket.
func (c *CLI) Run(args ...string) error {
	return c.run(args...)
}

// SplitPane splits a window horizontally (side by side) or vertically (top/bottom).
// The new pane runs the given command. Percent is the size of the new pane.
func (c *CLI) SplitPane(session, window, direction string, percent int, command string) error {
	target := session + ":" + window
	dirFlag := "-h" // horizontal = side by side
	if direction == "v" || direction == "vertical" {
		dirFlag = "-v"
	}

	args := []string{
		"split-window", dirFlag,
		"-t", target,
		"-p", fmt.Sprintf("%d", percent),
	}
	if command != "" {
		args = append(args, command)
	}

	return c.run(args...)
}

// ListSessions returns the names of all tmux sessions.
func (c *CLI) ListSessions() []string {
	out, err := c.output("list-sessions", "-F", "#{session_name}")
	if err != nil {
		return nil
	}
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// ListWindowNames returns the names of all windows in a session.
func (c *CLI) ListWindowNames(session string) ([]string, error) {
	out, err := c.output("list-windows", "-t", session, "-F", "#{window_name}")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// RenameWindow changes the name of a window in a session.
// Target can be a window name or index (e.g., "0" for the first window).
func (c *CLI) RenameWindow(session, target, newName string) error {
	windowTarget := session + ":" + target
	if target == "" {
		windowTarget = session
	}
	return c.run("rename-window", "-t", windowTarget, newName)
}

// SendKeys sends keystrokes to a specific window in a session.
func (c *CLI) SendKeys(session, window, keys string) error {
	return c.run("send-keys", "-t", session+":"+window, keys, "Enter")
}

func (c *CLI) output(args ...string) (string, error) {
	fullArgs := args
	if c.socket != "" {
		fullArgs = append([]string{"-L", c.socket}, args...)
	}

	cmd := exec.CommandContext(context.Background(), "tmux", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out)), nil
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
