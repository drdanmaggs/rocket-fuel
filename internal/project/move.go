package project

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// GHRunner executes gh CLI commands and returns the output.
// This allows testing without calling the real gh binary.
type GHRunner func(args ...string) ([]byte, error)

// MoveRequest holds the IDs needed to move a project item to a new status.
type MoveRequest struct {
	ProjectID string // project node ID (PVT_...)
	ItemID    string // item node ID (PVTI_...)
	FieldID   string // status field node ID (PVTSSF_...)
	OptionID  string // status option node ID
}

// ProjectMeta holds cached project metadata needed for board operations.
type ProjectMeta struct {
	ProjectID     string            // project node ID
	StatusFieldID string            // Status field node ID
	StatusOptions map[string]string // status name → option ID
}

// FetchProjectMeta retrieves project ID, status field ID, and option IDs.
func FetchProjectMeta(run GHRunner, owner string, number int) (*ProjectMeta, error) {
	// Get project ID.
	projOut, err := run("project", "view", strconv.Itoa(number),
		"--owner", owner, "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("fetch project: %w", err)
	}

	var proj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(projOut, &proj); err != nil {
		return nil, fmt.Errorf("parse project: %w", err)
	}

	// Get fields.
	fieldOut, err := run("project", "field-list", strconv.Itoa(number),
		"--owner", owner, "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("fetch fields: %w", err)
	}

	var fields struct {
		Fields []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Type    string `json:"type"`
			Options []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"options"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(fieldOut, &fields); err != nil {
		return nil, fmt.Errorf("parse fields: %w", err)
	}

	meta := &ProjectMeta{
		ProjectID:     proj.ID,
		StatusOptions: make(map[string]string),
	}

	for _, f := range fields.Fields {
		if f.Name == "Status" && f.Type == "ProjectV2SingleSelectField" {
			meta.StatusFieldID = f.ID
			for _, opt := range f.Options {
				meta.StatusOptions[opt.Name] = opt.ID
			}
			break
		}
	}

	if meta.StatusFieldID == "" {
		return nil, fmt.Errorf("Status field not found in project")
	}

	return meta, nil
}

// TransitionItem is a convenience function that fetches project metadata
// and moves an item to the target status in one call.
func TransitionItem(run GHRunner, owner string, number int, itemID, targetStatus string) error {
	meta, err := FetchProjectMeta(run, owner, number)
	if err != nil {
		return err
	}

	optionID, ok := meta.StatusOptions[targetStatus]
	if !ok {
		return fmt.Errorf("unknown status %q: valid options are %v", targetStatus, statusNames(meta.StatusOptions))
	}

	return MoveItem(run, MoveRequest{
		ProjectID: meta.ProjectID,
		ItemID:    itemID,
		FieldID:   meta.StatusFieldID,
		OptionID:  optionID,
	})
}

func statusNames(opts map[string]string) []string {
	names := make([]string, 0, len(opts))
	for k := range opts {
		names = append(names, k)
	}
	return names
}

// MoveItem changes a project item's status field to a new value.
func MoveItem(run GHRunner, req MoveRequest) error {
	_, err := run(
		"project", "item-edit",
		"--project-id", req.ProjectID,
		"--id", req.ItemID,
		"--field-id", req.FieldID,
		"--single-select-option-id", req.OptionID,
	)
	if err != nil {
		return fmt.Errorf("move item: %w", err)
	}
	return nil
}
