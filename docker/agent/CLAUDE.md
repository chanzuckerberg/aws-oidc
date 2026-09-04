# GitHub access

Use the `gh` CLI for every interaction with GitHub. `gh` carries this agent's GitHub App
credentials and routes each organization to the correct App installation, so it succeeds
where a raw `git` remote call or a hand-built API request fails.

## Use `gh` for anything that talks to GitHub

- Clone a repository: `gh repo clone <owner>/<repo>`.
- Pull requests: `gh pr create`, `gh pr view`, `gh pr diff`, `gh pr checkout`, `gh pr review`, `gh pr merge`.
- Issues: `gh issue create`, `gh issue view`, `gh issue list`.
- CI and workflow status: `gh pr checks`, `gh run view`, `gh run list`.
- Any GitHub REST or GraphQL call, including reading a file without cloning: `gh api ...`.

## Use `git` only for local version control

`gh` has no equivalent for local history, so keep using `git` for work inside a checkout:
`git status`, `git add`, `git commit`, `git switch` and `git branch`, `git rebase`, `git log`,
`git diff`, and `git push` to publish a branch. Never commit or push to a repository's primary
branch. Open a pull request with `gh pr create` instead.

## Prefer git worktrees

Several sessions share this pod and its `/workspace` volume at the same time. Switching
branches in one checkout changes it for every session, so give each task its own working tree
rather than switching branches in place.

- List existing worktrees before you start: `git -C /workspace/<repo> worktree list`.
- Create a worktree for a task: `git -C /workspace/<repo> worktree add /workspace/<repo>-<branch> -b <branch>`.
- Work in that directory, commit there, and open a pull request with `gh pr create`.
- Remove it when the task is done: `git -C /workspace/<repo> worktree remove /workspace/<repo>-<branch>`.

## Never

- Never call `https://api.github.com` with `curl` or `wget`. Use `gh api`.
- Never add or change git remotes, tokens or credential helpers. Authentication is already set up.
