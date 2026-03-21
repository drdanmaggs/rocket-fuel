package cmd

import (
	"fmt"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
	"github.com/spf13/cobra"
)

const meetingSessionName = "rf-meeting"

var meetCmd = &cobra.Command{
	Use:   "meet",
	Short: "Open a meeting with the Integrator",
	Long: `Starts a Meeting Room session — a separate Claude Code instance
for brainstorming ideas, scoping issues, and making product decisions.

The output of meetings is GitHub issues. The Integrator picks them
up from the board automatically.`,
	RunE: runMeet,
}

func init() {
	rootCmd.AddCommand(meetCmd)
}

func runMeet(cmd *cobra.Command, _ []string) error {
	tm := tmux.New()
	out := cmd.OutOrStdout()

	// Check if meeting session already exists.
	if tm.HasSession(meetingSessionName) {
		_, _ = fmt.Fprintln(out, "Joining existing meeting...")
		return tm.AttachCC(meetingSessionName)
	}

	// Create meeting session.
	if err := tm.NewSession(meetingSessionName); err != nil {
		return fmt.Errorf("create meeting session: %w", err)
	}

	// Rename window 0 to "meeting".
	_ = tm.RenameWindow(meetingSessionName, "0", "meeting")

	// Launch Claude with meeting-specific context.
	meetingPrompt := `You are in a Rocket Fuel Meeting Room. The Visionary (human) called this meeting to brainstorm, scope ideas, or make product decisions.

Your job:
1. Listen to the Visionary's ideas
2. Help scope them into actionable GitHub issues (use /issue-scope if appropriate)
3. Create well-documented issues with gh issue create
4. Add appropriate labels (workflow:tdd, workflow:epc, etc.)

When the meeting ends, the Integrator will pick up the new issues from the board automatically.

Do NOT dispatch workers or manage the board — that's the Integrator's job. Your only output is GitHub issues.`

	launchCmd := fmt.Sprintf("claude --dangerously-skip-permissions %s", shellQuoteMeeting(meetingPrompt))
	if err := tm.SendKeys(meetingSessionName, "meeting", launchCmd); err != nil {
		return fmt.Errorf("launch meeting claude: %w", err)
	}

	_, _ = fmt.Fprintln(out, "Meeting Room open. Attaching...")

	return tm.AttachCC(meetingSessionName)
}

func shellQuoteMeeting(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
