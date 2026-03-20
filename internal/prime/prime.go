// Package prime provides context injection for the Integrator agent.
// Build assembles board state, worker status, and repo context into
// a single markdown document — the Integrator's "eyes."
package prime

import (
	"fmt"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/drdanmaggs/rocket-fuel/internal/status"
)

// Input holds the pre-fetched data needed to build the Integrator context.
type Input struct {
	IntegratorPrompt string
	Board            *project.BoardSummary
	Status           *status.Summary
	RepoDir          string
	Branch           string
}

// Build assembles the Integrator context document from pre-fetched data.
func Build(in *Input) string {
	var b strings.Builder

	if in.IntegratorPrompt != "" {
		_, _ = fmt.Fprintln(&b, in.IntegratorPrompt)
	}

	if in.Board != nil {
		_, _ = fmt.Fprintln(&b)
		_, _ = fmt.Fprintln(&b, "## Board")
		_, _ = fmt.Fprintln(&b)
		_, _ = fmt.Fprint(&b, project.FormatBoard(in.Board))
	}

	if in.Status != nil {
		_, _ = fmt.Fprintln(&b)
		_, _ = fmt.Fprintln(&b, "## Workers")
		_, _ = fmt.Fprintln(&b)
		_, _ = fmt.Fprint(&b, status.Format(in.Status))
	}

	if in.RepoDir != "" || in.Branch != "" {
		_, _ = fmt.Fprintln(&b)
		_, _ = fmt.Fprintln(&b, "## Repo")
		_, _ = fmt.Fprintln(&b)
		if in.RepoDir != "" {
			_, _ = fmt.Fprintf(&b, "Directory: %s\n", in.RepoDir)
		}
		if in.Branch != "" {
			_, _ = fmt.Fprintf(&b, "Branch: %s\n", in.Branch)
		}
	}

	return b.String()
}
