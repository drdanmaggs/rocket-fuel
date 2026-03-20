package project

import (
	"fmt"
	"strings"
	"testing"
)

func TestFetchMeta_parsesFieldsAndOptions(t *testing.T) {
	t.Parallel()

	// Simulate gh project field-list JSON output.
	fieldJSON := `{"fields":[{"id":"PVTSSF_abc","name":"Status","type":"ProjectV2SingleSelectField","options":[{"id":"opt_1","name":"Backlog"},{"id":"opt_2","name":"Scoped"},{"id":"opt_3","name":"In Progress"},{"id":"opt_4","name":"Done"}]},{"id":"PVTF_title","name":"Title","type":"ProjectV2Field"}]}`

	// Simulate gh project view JSON output.
	projectJSON := `{"id":"PVT_123","title":"My Project"}`

	callCount := 0
	runner := func(_ ...string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(projectJSON), nil
		}
		return []byte(fieldJSON), nil
	}

	meta, err := FetchMeta(runner, "drdanmaggs", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.ProjectID != "PVT_123" {
		t.Errorf("expected project ID PVT_123, got %q", meta.ProjectID)
	}
	if meta.StatusFieldID != "PVTSSF_abc" {
		t.Errorf("expected status field ID PVTSSF_abc, got %q", meta.StatusFieldID)
	}
	if meta.StatusOptions["In Progress"] != "opt_3" {
		t.Errorf("expected In Progress option opt_3, got %q", meta.StatusOptions["In Progress"])
	}
	if meta.StatusOptions["Done"] != "opt_4" {
		t.Errorf("expected Done option opt_4, got %q", meta.StatusOptions["Done"])
	}
}

func TestTransitionItem_movesToTargetStatus(t *testing.T) {
	t.Parallel()

	projectJSON := `{"id":"PVT_proj"}`
	fieldJSON := `{"fields":[{"id":"PVTSSF_status","name":"Status","type":"ProjectV2SingleSelectField","options":[{"id":"opt_scoped","name":"Scoped"},{"id":"opt_ip","name":"In Progress"}]}]}`

	var editArgs []string
	callCount := 0
	runner := func(args ...string) ([]byte, error) {
		callCount++
		switch {
		case len(args) > 0 && args[0] == "project" && args[1] == "view":
			return []byte(projectJSON), nil
		case len(args) > 0 && args[0] == "project" && args[1] == "field-list":
			return []byte(fieldJSON), nil
		case len(args) > 0 && args[0] == "project" && args[1] == "item-edit":
			editArgs = args
			return []byte("{}"), nil
		}
		return nil, fmt.Errorf("unexpected call: %v", args)
	}

	err := TransitionItem(runner, "drdanmaggs", 1, "PVTI_item", "In Progress")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify item-edit was called with correct option ID.
	found := false
	for i, arg := range editArgs {
		if arg == "--single-select-option-id" && i+1 < len(editArgs) && editArgs[i+1] == "opt_ip" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected item-edit with option opt_ip, got: %v", editArgs)
	}
}

func TestTransitionItem_errorForInvalidStatus(t *testing.T) {
	t.Parallel()

	projectJSON := `{"id":"PVT_proj"}`
	fieldJSON := `{"fields":[{"id":"PVTSSF_status","name":"Status","type":"ProjectV2SingleSelectField","options":[{"id":"opt_scoped","name":"Scoped"}]}]}`

	callCount := 0
	runner := func(args ...string) ([]byte, error) {
		callCount++
		switch {
		case len(args) > 0 && args[1] == "view":
			return []byte(projectJSON), nil
		case len(args) > 0 && args[1] == "field-list":
			return []byte(fieldJSON), nil
		}
		return nil, fmt.Errorf("unexpected call")
	}

	err := TransitionItem(runner, "drdanmaggs", 1, "PVTI_item", "Nonexistent")
	if err == nil {
		t.Fatal("expected error for invalid status name")
	}
	if !strings.Contains(err.Error(), "Nonexistent") {
		t.Errorf("expected error to mention status name, got: %v", err)
	}
}

func TestFetchMeta_errorWhenNoStatusField(t *testing.T) {
	t.Parallel()

	fieldJSON := `{"fields":[{"id":"PVTF_title","name":"Title","type":"ProjectV2Field"}]}`
	projectJSON := `{"id":"PVT_123"}`

	callCount := 0
	runner := func(_ ...string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			return []byte(projectJSON), nil
		}
		return []byte(fieldJSON), nil
	}

	_, err := FetchMeta(runner, "owner", 1)
	if err == nil {
		t.Fatal("expected error when Status field missing")
	}
	if !strings.Contains(err.Error(), "status field not found") {
		t.Errorf("expected 'status field not found' error, got: %v", err)
	}
}

func TestMoveItem_returnsErrorOnFailure(t *testing.T) {
	t.Parallel()

	runner := func(_ ...string) ([]byte, error) {
		return nil, fmt.Errorf("gh failed: exit 1")
	}

	err := MoveItem(runner, MoveRequest{
		ProjectID: "PVT_abc",
		ItemID:    "PVTI_123",
		FieldID:   "PVTSSF_status",
		OptionID:  "opt_in_progress",
	})
	if err == nil {
		t.Fatal("expected error when gh fails")
	}
	if !strings.Contains(err.Error(), "move item") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestMoveItem_buildsCorrectCommand(t *testing.T) {
	t.Parallel()

	var executed []string
	runner := func(args ...string) ([]byte, error) {
		executed = args
		return []byte("{}"), nil
	}

	err := MoveItem(runner, MoveRequest{
		ProjectID: "PVT_abc",
		ItemID:    "PVTI_123",
		FieldID:   "PVTSSF_status",
		OptionID:  "opt_in_progress",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the gh command was built correctly.
	want := []string{
		"project", "item-edit",
		"--project-id", "PVT_abc",
		"--id", "PVTI_123",
		"--field-id", "PVTSSF_status",
		"--single-select-option-id", "opt_in_progress",
	}
	if len(executed) != len(want) {
		t.Fatalf("expected %d args, got %d: %v", len(want), len(executed), executed)
	}
	for i := range want {
		if executed[i] != want[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, want[i], executed[i])
		}
	}
}
