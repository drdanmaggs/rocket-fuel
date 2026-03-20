package testutil

import "testing"

// FakeTmux records tmux commands for testing without a real tmux server.
type FakeTmux struct {
	Commands []TmuxCommand
}

// TmuxCommand records a tmux invocation.
type TmuxCommand struct {
	Action string
	Args   []string
}

// NewFakeTmux creates a new FakeTmux recorder.
func NewFakeTmux(_ *testing.T) *FakeTmux {
	return &FakeTmux{}
}

// Record stores a tmux command invocation.
func (f *FakeTmux) Record(action string, args ...string) {
	f.Commands = append(f.Commands, TmuxCommand{Action: action, Args: args})
}

// CommandCount returns the number of recorded commands.
func (f *FakeTmux) CommandCount() int {
	return len(f.Commands)
}

// LastCommand returns the most recently recorded command.
// Returns an empty TmuxCommand if none recorded.
func (f *FakeTmux) LastCommand() TmuxCommand {
	if len(f.Commands) == 0 {
		return TmuxCommand{}
	}
	return f.Commands[len(f.Commands)-1]
}

// HasCommand checks if any recorded command matches the given action.
func (f *FakeTmux) HasCommand(action string) bool {
	for _, cmd := range f.Commands {
		if cmd.Action == action {
			return true
		}
	}
	return false
}
