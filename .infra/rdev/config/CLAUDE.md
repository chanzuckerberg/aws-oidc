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

## Never

- Never call `https://api.github.com` with `curl` or `wget`. Use `gh api`.
- Never add or change git remotes, tokens or credential helpers. Authentication is already set up.
