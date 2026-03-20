package prime

import (
	"strings"
	"testing"

	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
)

func TestBuild_includesIntegratorPrompt(t *testing.T) {
	t.Parallel()

	prompt := "You are the Integrator."
	input := &Input{
		IntegratorPrompt: prompt,
	}

	got := Build(input)

	if !strings.Contains(got, prompt) {
		t.Errorf("expected output to contain integrator prompt, got:\n%s", got)
	}
}

func TestBuild_includesBoardState(t *testing.T) {
	t.Parallel()

	input := &Input{
		Board: &project.BoardSummary{
			Columns: map[string][]project.Item{
				"Scoped": {
					{Number: 42, Title: "Add login page", Labels: []string{"workflow:tdd"}},
				},
				"In Progress": {
					{Number: 43, Title: "Fix auth bug"},
				},
			},
		},
	}

	got := Build(input)

	if !strings.Contains(got, "## Board") {
		t.Error("expected Board section header")
	}
	if !strings.Contains(got, "#42") {
		t.Error("expected issue #42 in output")
	}
	if !strings.Contains(got, "Add login page") {
		t.Error("expected issue title in output")
	}
	if !strings.Contains(got, "Scoped") {
		t.Error("expected column name in output")
	}
}

func TestBuild_includesWorkerStatus(t *testing.T) {
	t.Parallel()

	input := &Input{
		Status: &status.Summary{
			SessionActive: true,
			Workers: []status.WorkerStatus{
				{Name: "worker-42", WindowOpen: true, Branch: "rf/issue-42", HasPR: false},
				{Name: "worker-43", WindowOpen: false, Branch: "rf/issue-43", HasPR: true},
			},
		},
	}

	got := Build(input)

	if !strings.Contains(got, "## Workers") {
		t.Error("expected Workers section header")
	}
	if !strings.Contains(got, "worker-42") {
		t.Error("expected worker-42 in output")
	}
	if !strings.Contains(got, "active") {
		t.Error("expected active status for worker-42")
	}
	if !strings.Contains(got, "PR open") {
		t.Error("expected PR indicator for worker-43")
	}
}

func TestBuild_includesRepoContext(t *testing.T) {
	t.Parallel()

	input := &Input{
		RepoDir: "/home/user/my-project",
		Branch:  "main",
	}

	got := Build(input)

	if !strings.Contains(got, "## Repo") {
		t.Error("expected Repo section header")
	}
	if !strings.Contains(got, "/home/user/my-project") {
		t.Error("expected repo dir in output")
	}
	if !strings.Contains(got, "main") {
		t.Error("expected branch name in output")
	}
}

func TestBuild_missingBoardOmitsSection(t *testing.T) {
	t.Parallel()

	input := &Input{
		IntegratorPrompt: "You are the Integrator.",
		// Board is nil — no project linked
	}

	got := Build(input)

	if strings.Contains(got, "## Board") {
		t.Error("expected no Board section when board is nil")
	}
	if !strings.Contains(got, "You are the Integrator.") {
		t.Error("expected integrator prompt even without board")
	}
}

func TestBuild_missingWorkersShowsNone(t *testing.T) {
	t.Parallel()

	input := &Input{
		Status: &status.Summary{
			SessionActive: true,
			Workers:       nil,
		},
	}

	got := Build(input)

	if !strings.Contains(got, "## Workers") {
		t.Error("expected Workers section even with no workers")
	}
	if !strings.Contains(got, "none") {
		t.Error("expected 'none' when no workers exist")
	}
}

func TestBuild_fullAssemblyOrdersCorrectly(t *testing.T) {
	t.Parallel()

	input := &Input{
		IntegratorPrompt: "PROMPT_SECTION",
		Board: &project.BoardSummary{
			Columns: map[string][]project.Item{
				"Scoped": {{Number: 1, Title: "Issue one"}},
			},
		},
		Status: &status.Summary{
			SessionActive: true,
			Workers: []status.WorkerStatus{
				{Name: "worker-1", WindowOpen: true, Branch: "rf/issue-1"},
			},
		},
		RepoDir: "/repo",
		Branch:  "main",
	}

	got := Build(input)

	// Verify section ordering: prompt → board → workers → repo.
	promptIdx := strings.Index(got, "PROMPT_SECTION")
	boardIdx := strings.Index(got, "## Board")
	workersIdx := strings.Index(got, "## Workers")
	repoIdx := strings.Index(got, "## Repo")

	if promptIdx < 0 || boardIdx < 0 || workersIdx < 0 || repoIdx < 0 {
		t.Fatalf("missing section(s) in output:\n%s", got)
	}
	if promptIdx >= boardIdx || boardIdx >= workersIdx || workersIdx >= repoIdx {
		t.Errorf("sections out of order: prompt=%d board=%d workers=%d repo=%d",
			promptIdx, boardIdx, workersIdx, repoIdx)
	}
}
