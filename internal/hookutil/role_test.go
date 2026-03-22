package hookutil_test

import (
	"strings"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/hookutil"
)

func TestDetectRole_returnsIntegratorWhenAgentTypeIsIntegrator(t *testing.T) {
	t.Parallel()

	input := `{"agent_type": "integrator", "cwd": "/some/repo"}`
	reader := strings.NewReader(input)

	role := hookutil.DetectRole(reader)

	if role != hookutil.RoleIntegrator {
		t.Errorf("expected %q, got %q", hookutil.RoleIntegrator, role)
	}
}

func TestDetectRole_returnsWorkerWhenAgentTypeIsWorker(t *testing.T) {
	t.Parallel()

	input := `{"agent_type": "worker", "cwd": "/some/repo"}`
	reader := strings.NewReader(input)

	role := hookutil.DetectRole(reader)

	if role != hookutil.RoleWorker {
		t.Errorf("expected %q, got %q", hookutil.RoleWorker, role)
	}
}
