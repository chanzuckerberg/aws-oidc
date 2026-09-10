# Managed Agent Policy

## GitHub authentication

Use the `gh` CLI for every interaction with GitHub. The installed `gh` wrapper carries this
agent's GitHub App credentials and routes each organization to the correct App installation.

- Clone repositories with `gh repo clone <owner>/<repo>`.
- Use `gh pr`, `gh issue`, `gh run` and `gh api` for their respective GitHub operations.
- Never call `api.github.com` with `curl` or `wget`.
- Never add or change git remotes, tokens or credential helpers.

Use `git` for local version control. This includes status, staging, commits, branches, rebases,
logs and diffs. Use `git push` to publish a branch because `gh` has no push equivalent.

Never commit or push to a repository's primary branch. Open a pull request instead.

## Repository workflow

`/workspace` is persistent and shared across this agent's sessions. Configured repositories
are already cloned into `/workspace/<repo>` and authenticated. Use those checkouts instead of
cloning another copy.

Concurrent sessions can share a checkout. Create a worktree for each task instead of
switching the checkout's branch:

```bash
git -C /workspace/<repo> worktree list
git -C /workspace/<repo> worktree add /workspace/<repo>-<branch> -b <branch>
```

Remove the worktree after the task finishes.
