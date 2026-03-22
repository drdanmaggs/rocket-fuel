package hookutil

import (
	"encoding/json"
	"io"
	"strings"
)

// Role identifies the type of agent session (Integrator or Worker).
type Role string

// Role constants for agent detection.
const (
	RoleIntegrator Role = "integrator"
	RoleWorker     Role = "worker"
)

// DetectRole parses Claude Code hook input JSON to determine the agent role.
// Primary: checks agent_type field. Fallback: checks if cwd contains .worktrees/.
// Defaults to Integrator when neither signal is present.
func DetectRole(input io.Reader) Role {
	var data struct {
		AgentType string `json:"agent_type"`
		CWD       string `json:"cwd"`
	}

	if err := json.NewDecoder(input).Decode(&data); err != nil {
		return RoleIntegrator // Default fallback
	}

	// Primary: check agent_type from hook input.
	if data.AgentType == string(RoleWorker) {
		return RoleWorker
	}

	// Fallback: detect Worker from worktree working directory.
	if data.AgentType == "" && strings.Contains(data.CWD, ".worktrees/") {
		return RoleWorker
	}

	return RoleIntegrator // Default: treat unknown as Integrator
}
