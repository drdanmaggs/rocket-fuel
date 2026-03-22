package hookutil

import (
	"encoding/json"
	"io"
)

type Role string

const (
	RoleIntegrator Role = "integrator"
	RoleWorker     Role = "worker"
)

func DetectRole(input io.Reader) Role {
	var data struct {
		AgentType string `json:"agent_type"`
	}

	if err := json.NewDecoder(input).Decode(&data); err != nil {
		return RoleIntegrator // Default fallback
	}

	if data.AgentType == "integrator" {
		return RoleIntegrator
	}

	return RoleIntegrator // Default fallback for now
}
