package hookutil

import (
	"encoding/json"
	"io"
	"strings"
)

type Role string

const (
	RoleIntegrator Role = "integrator"
	RoleWorker     Role = "worker"
)

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
