---
name: worktree-reset
description: "Reset this worktree after a merged PR — cleans branch, fetches latest, prepares for new work"
allowed-tools: Bash(git *), Bash(gh *)
---

# Post-merge worktree reset

Current state:
- Branch: !`git branch --show-current`
- Worktrees: !`git worktree list`
- Status: !`git status --porcelain`

## Instructions

1. Verify the current branch's PR has been merged using `gh pr view --json state`
2. If not merged, ask the user if they want to proceed anyway
3. Delete the remote branch: `git push origin --delete <branch>`
   - If this fails (already deleted), that's fine - continue
4. Fetch latest changes: `git fetch origin --prune`
5. Detach HEAD at origin/main: `git switch --detach origin/main`
6. Delete the local branch: `git branch -D <branch>`
7. Clean the working tree: `git clean -fd && git restore .`
8. Report what was cleaned up and confirm the worktree is ready for new work

## Safety checks

- NEVER reset if current branch is `main` or `master`
- If already in detached HEAD state, just clean and confirm ready
- Always verify PR state before proceeding
