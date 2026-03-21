package dashboard

import (
	"strings"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/status"
)

func TestRender_noWorkers(t *testing.T) {
	t.Parallel()

	data := &DashboardData{
		Summary: &status.Summary{SessionActive: true},
	}
	out := Render(data)

	if !strings.Contains(out, "DASHBOARD") {
		t.Error("expected DASHBOARD header")
	}
	if !strings.Contains(out, "ACTIVE") {
		t.Error("expected ACTIVE session")
	}
	if !strings.Contains(out, "none") {
		t.Error("expected 'none' for no workers")
	}
}

func TestRender_withWorkers(t *testing.T) {
	t.Parallel()

	data := &DashboardData{
		Summary: &status.Summary{
			SessionActive: true,
			Workers: []status.WorkerStatus{
				{Name: "worker-42", WindowOpen: true, HasPR: false},
				{Name: "worker-43", WindowOpen: false, HasPR: true},
			},
		},
	}
	out := Render(data)

	if !strings.Contains(out, "Workers: 2") {
		t.Error("expected worker count")
	}
	if !strings.Contains(out, "worker-42") {
		t.Error("expected worker-42")
	}
	if !strings.Contains(out, "[PR]") {
		t.Error("expected PR indicator")
	}
}
