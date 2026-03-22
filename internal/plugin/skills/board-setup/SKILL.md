---
name: board-setup
description: "Create or verify the dedicated Rocket Fuel project board with GitHub Project automation"
---

# Board Setup Skill

This skill guides the Integrator through creating a GitHub Project board for Rocket Fuel with the required workflow columns and automation.

## Overview

The Rocket Fuel orchestrator uses a GitHub Project board to track work items across multiple workers. This skill ensures the board exists with the correct columns and automation rules configured.

## Required Columns

The board must have the following columns in this exact order:

1. **Backlog** — New issues awaiting scope review
2. **Ready** — Scoped issues ready for worker assignment
3. **Scoped** — Currently assigned to a worker
4. **In Progress** — Worker is actively working on the issue
5. **In Review** — Work completed, waiting for review/merge
6. **Done** — Completed and closed

## Setup Steps

### Step 1: Create the Project Board

Use the GitHub CLI to create a new project in the repository:

```bash
gh project create --owner drdanmaggs --title "Rocket Fuel" --template table
```

Or visit https://github.com/orgs/drdanmaggs/projects and click "New project".

### Step 2: Create Columns

Using the GitHub CLI, create each column:

```bash
gh project field-create --project "Rocket Fuel" --name "Backlog" --field-type single_select
gh project field-create --project "Rocket Fuel" --name "Ready" --field-type single_select
gh project field-create --project "Rocket Fuel" --name "Scoped" --field-type single_select
gh project field-create --project "Rocket Fuel" --name "In Progress" --field-type single_select
gh project field-create --project "Rocket Fuel" --name "In Review" --field-type single_select
gh project field-create --project "Rocket Fuel" --name "Done" --field-type single_select
```

Or manually create columns via the GitHub UI by clicking the "+" button on the right side of the project board.

### Step 3: Configure Automation (Optional)

Link the board to the repository to enable automatic status updates:

- Issues opened go to **Backlog**
- Issues closed go to **Done**

This is configured in the project settings under "Automation".

## Verification

To verify the board is correctly set up:

1. Navigate to the Rocket Fuel project board
2. Confirm all six columns are present: Backlog, Ready, Scoped, In Progress, In Review, Done
3. Verify you can move issues between columns

Once configured, the Integrator can begin assigning work by moving issues to the **Scoped** column with a worker assignment.
