package project

import (
	"encoding/json"
	"fmt"
)

// Create makes a new GitHub Project board for the repository.
func Create(run GHRunner, owner, repo string) (*Config, error) {
	out, err := run("project", "create", "--owner", owner, "--title", repo, "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	var resp struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parse create response: %w", err)
	}

	return &Config{
		Owner:         owner,
		ProjectNumber: resp.Number,
	}, nil
}

// Discover finds a GitHub Project linked to a repository via the GitHub API.
// Returns the first project found, or an error if none exist.
func Discover(run GHRunner, owner, repo string) (*Config, error) {
	query := fmt.Sprintf(
		`query { repository(owner: "%s", name: "%s") { projectsV2(first: 5) { nodes { id title number url } } } }`,
		owner, repo,
	)

	out, err := run("api", "graphql", "-f", "query="+query)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}

	var resp struct {
		Data struct {
			Repository struct {
				ProjectsV2 struct {
					Nodes []struct {
						ID     string `json:"id"`
						Title  string `json:"title"`
						Number int    `json:"number"`
						URL    string `json:"url"`
					} `json:"nodes"`
				} `json:"projectsV2"`
			} `json:"repository"`
		} `json:"data"`
	}

	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parse projects response: %w", err)
	}

	nodes := resp.Data.Repository.ProjectsV2.Nodes
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no GitHub Project found for %s/%s", owner, repo)
	}

	return &Config{
		Owner:         owner,
		ProjectNumber: nodes[0].Number,
	}, nil
}
