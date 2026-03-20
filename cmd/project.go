package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/project"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage the linked GitHub Project",
	Long:  `Link a GitHub Project board and view its status. The Integrator uses this as its brain.`,
}

var projectLinkCmd = &cobra.Command{
	Use:   "link <project-url-or-number>",
	Short: "Link a GitHub Project to this session",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectLink,
}

var projectStatusCmd = &cobra.Command{
	Use:   "board",
	Short: "Show the project board status",
	RunE:  runProjectBoard,
}

func init() {
	projectCmd.AddCommand(projectLinkCmd)
	projectCmd.AddCommand(projectStatusCmd)
	rootCmd.AddCommand(projectCmd)
}

func runProjectLink(cmd *cobra.Command, args []string) error {
	ref := args[0]

	owner, number, err := parseProjectRef(ref)
	if err != nil {
		return err
	}

	cfg := project.Config{
		ProjectNumber: number,
		Owner:         owner,
	}

	if err := saveProjectConfig(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Linked project #%d (owner: %s)\n", number, owner)
	return nil
}

func runProjectBoard(cmd *cobra.Command, _ []string) error {
	cfg, err := loadProjectConfig()
	if err != nil {
		return fmt.Errorf("no project linked: %w\nRun: rocket-fuel project link <project-url>", err)
	}

	board, err := project.FetchBoard(cfg.Owner, cfg.ProjectNumber)
	if err != nil {
		return fmt.Errorf("fetch board: %w", err)
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), project.FormatBoard(board))
	return nil
}

func parseProjectRef(ref string) (string, int, error) {
	// Handle URLs like https://github.com/users/owner/projects/1
	// or https://github.com/orgs/owner/projects/1
	if strings.Contains(ref, "/projects/") {
		parts := strings.Split(ref, "/projects/")
		if len(parts) == 2 {
			numStr := strings.TrimRight(parts[1], "/")
			num, err := strconv.Atoi(numStr)
			if err != nil {
				return "", 0, fmt.Errorf("invalid project number in URL: %q", numStr)
			}

			// Extract owner from the path.
			pathParts := strings.Split(parts[0], "/")
			owner := pathParts[len(pathParts)-1]
			return owner, num, nil
		}
	}

	return "", 0, fmt.Errorf("invalid project reference %q: use a GitHub Project URL (e.g., https://github.com/users/owner/projects/1)", ref)
}

func configDir() (string, error) {
	repoDir, err := repoRoot()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(repoDir, ".rocket-fuel")
	return dir, os.MkdirAll(dir, 0o755)
}

func saveProjectConfig(cfg project.Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "project.json"), data, 0o644)
}

func loadProjectConfig() (*project.Config, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(dir, "project.json"))
	if err == nil {
		var cfg project.Config
		if parseErr := json.Unmarshal(data, &cfg); parseErr == nil {
			return &cfg, nil
		}
	}

	// No config file — try auto-discovery from GitHub.
	cfg, discoverErr := discoverAndSaveProject()
	if discoverErr != nil {
		return nil, fmt.Errorf("no project linked and auto-discovery failed: %w", discoverErr)
	}
	return cfg, nil
}

func discoverAndSaveProject() (*project.Config, error) {
	owner, repo, err := repoOwnerAndName()
	if err != nil {
		return nil, err
	}

	// Try discover first.
	cfg, err := project.Discover(ghRunner, owner, repo)
	if err != nil {
		// No existing project — create one.
		fmt.Fprintf(os.Stderr, "No GitHub Project found for %s/%s — creating one...\n", owner, repo)
		cfg, err = project.Create(ghRunner, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("auto-create project failed: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Created project #%d for %s/%s\n", cfg.ProjectNumber, owner, repo)
	} else {
		fmt.Fprintf(os.Stderr, "Auto-discovered project #%d for %s/%s\n", cfg.ProjectNumber, owner, repo)
	}

	// Save for next time.
	if saveErr := saveProjectConfig(*cfg); saveErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save project config: %v\n", saveErr)
	}

	return cfg, nil
}

func repoOwnerAndName() (string, string, error) {
	out, err := exec.CommandContext(context.Background(),
		"gh", "repo", "view", "--json", "owner,name",
	).Output()
	if err != nil {
		return "", "", fmt.Errorf("detect repo: %w", err)
	}

	var repo struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &repo); err != nil {
		return "", "", fmt.Errorf("parse repo info: %w", err)
	}

	return repo.Owner.Login, repo.Name, nil
}
