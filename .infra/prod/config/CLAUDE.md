# Personal Rules

## Response Style

Be concise and to the point. Answer the core question directly without adding extra context, caveats, or related details unless I specifically ask for them.

Do not append extra tidbits, caveats, or follow-up offers at the end of a response. Stop once the question is answered. If I want more detail or a follow-up, I will ask.

## Don't Reply to GitHub Comments When Addressing Them

When I ask you to address a comment (from a PR review, an issue, or anywhere on GitHub), fix the underlying code or content. Do not reply to the comment and do not post a new comment on GitHub.

- Do not use `gh pr comment`, `gh pr review`, `gh api` to post comments, or any other means of writing a comment.
- Do not resolve, react to, or otherwise respond to the comment on GitHub.
- Just make the change. I will handle any responses on GitHub myself.

## Use the `gh` CLI for GitHub

Always use the GitHub CLI (`gh`) for anything that touches GitHub: pull requests, issues, reviews, releases, checks, repo and workflow operations, and ad-hoc API calls via `gh api`.

Do not reach for raw `git` against GitHub remotes, `curl` against `api.github.com`, or the web UI when a `gh` command does the job. `gh` is already authenticated, handles HTTPS auth without SSH host-key prompts, and keeps output scriptable.

Examples:
- Open a PR with `gh pr create`, not by printing a "create a PR" URL.
- Read PR state with `gh pr view` / `gh pr checks`, not by scraping the web page.
- Hit unsupported endpoints with `gh api`, not `curl -H "Authorization: ..."`.

## Create PRs as Drafts

Always open pull requests in draft mode. Marking a PR ready for review notifies reviewers, and an agent-created PR is not ready until a human has checked it. The author will undraft it when it is actually ready.

- Pass `--draft` to `gh pr create`.
- Never mark a PR ready for review (`gh pr ready`) unless explicitly asked.
- This applies to every PR, including stacked PRs.

```bash
gh pr create --draft --base main --title "..." --body "..."
```

## Conventional Commits for PR Titles

When creating pull requests, use conventional commit format for the title:

```
<type>(<scope>): <short summary>
```

Common types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `ci`, `perf`, `build`

Examples:
- `feat(oidc): store refresh token expiry in serialized token`
- `refactor(storage): move compression into shared package`
- `fix(cache): handle corrupted token gracefully`
- `chore(deps): bump golang.org/x/oauth2 to v0.25`

## No Test Plan in PR Descriptions

Do not include a "Test plan" section (or any checklist of testing TODOs) when creating or editing pull request descriptions. Test plans vary by project and will be added manually by the author.

PR bodies should only contain a concise summary of what changed and why.

## PR Descriptions: Impact Over File Inventories

Do not include a "What changed" section (or any per-file breakdown) in PR descriptions. A bullet-per-file list of each path and the mechanical edit made to it is noise — the diff already shows that. Drop the section entirely.

Focus on the impact and the idea: what problem the change solves, the approach taken, and any consequences a reviewer needs to know (behavior changes, migrations, deployment ordering, follow-ups). Describe the change at the level of concepts and outcomes, not files.

Reference a specific file only when it is load-bearing for understanding the change (for example, "the precedence logic lives in `ServiceAccountTokenProvider`"), not to enumerate everything that was touched.

Bad: a "What changed" section like
- `foo.ts` — removed the `Bar` class and added `Baz`
- `config.yaml` — wired the new key
- `docker-compose.yml` — pass the env var through

Good: "Replaces the KMS-signed token exchange with the pod's projected ServiceAccount token, and adds an inline-token escape hatch so the backend still authenticates during local development where no projected token exists."

## Prose Style

Write in plain English. No arrows, symbols, or shorthand in place of words.

Bad: "github-script v8→v9, add-and-commit v9→v10"
Good: "bumped github-script to v9 and add-and-commit to v10"

Bad: "foo -> bar, baz -> qux"
Good: "renamed foo to bar and baz to qux"

Keep sentences short. Prefer active voice. Avoid filler phrases like "in order to", "please note that", or "it is worth mentioning".

No em dashes. Do not use dashes to join clauses. Use a period and start a new sentence.

No colons for emphasis or dramatic effect. Colons introduce lists or code, not restatements.

Bad: "Detection was manual: a person noticed broken ingress."
Good: "A person noticed broken ingress manually."

No scare quotes. Quotes are for direct quotations only. If a word is the right word, use it plain.

Bad: "showed the NLBs as 'orphaned' with no detail"
Good: "showed the NLBs as orphaned with no detail"

Do not pack two independent facts into one sentence joined by a dash or parenthetical. Break them into separate sentences.

On first use of a term, spell it out with the abbreviation in parentheses. Use the abbreviation from then on.

Good: "network load balancers (NLBs)" on first mention, then "NLBs" afterward.

Break up run-on sentences and long comma-separated lists. Use a bulleted or numbered list instead when there are more than two or three items.

Bad: "Bumps github-script to v9, add-and-commit to v10, upload-artifact to v7, download-artifact to v8, release-please-action to v5, and create-pull-request to v8."
Good:
- bumped github-script to v9
- bumped add-and-commit to v10
- bumped upload-artifact to v7

Lead with impact, not mechanics. Say what changed and why it matters before explaining how.

Bad: "Renames the app-id input to client-id in six workflow files."
Good: "Silences the deprecation warning from create-github-app-token by renaming the app-id input to client-id."

## Shell Tilde Expansion in Non-Interactive Shells

`~` is **not** reliably expanded in non-interactive shells (e.g., CI, agent tool calls, subshells). Always use `$HOME` instead.

This applies especially to colon-separated values like `PATH`, `KUBECONFIG`, `LD_LIBRARY_PATH`, etc., where `~` after `:` is never expanded even in zsh.

```bash
# BAD — ~ may not expand, creating literal `~` directories
export KUBECONFIG=~/.kube/config:~/.kube/other.yaml

# GOOD — $HOME always expands
export KUBECONFIG=$HOME/.kube/config:$HOME/.kube/other.yaml
```

## Error Handling Style (Go)

Always check errors with a separate `if err != nil` block. Never inline the error check with the function call.

```go
// BAD — inline error check
if err := doSomething(); err != nil {
	return fmt.Errorf("doing something: %w", err)
}

// GOOD — assign then check
err := doSomething()
if err != nil {
	return fmt.Errorf("doing something: %w", err)
}
```

## No Underscore Prefix on Variable Names (Go)

Do not use a leading underscore (`_`) on local or package-level variable names in Go. Use plain camelCase for unexported identifiers.

```go
// BAD
_flagName := "verbose"
var _configFile string

// GOOD
flagName := "verbose"
var configFile string
```

## Uber Go Style Guide (Go)

Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

Key points:

- **Error wrapping**: Use `fmt.Errorf("short context: %w", err)`. Avoid "failed to" prefixes.
- **Handle errors once**: Either return the error (and let the caller log/exit) or log and degrade — never both.
- **Interfaces**: Add compile-time checks: `var _ MyInterface = (*MyType)(nil)`.
- **Unnecessary else**: Use default + conditional override instead of if/else.
- **Reduce nesting**: Handle errors/special cases first, return early.
- **Struct init**: Use field names; omit zero-value fields; `var x T` for zero-value structs; `&T{}` not `new(T)`.
- **Maps/slices**: Specify capacity when known. Use `nil` not `[]T{}` for empty returns.
- **Exit only in main**: `os.Exit`/`log.Fatal` only in `main()`. All other code returns errors.
- **Don't panic**: Return errors instead. Use `t.Fatal` in tests.
- **Defer for cleanup**: Use defer for closing files, releasing locks, etc.
- **Functional options**: Use for constructors with 3+ optional params.
- **Channel size**: One or unbuffered.
- **Goroutines**: Never fire-and-forget. Every goroutine must have a way to stop and be waited on.

## Cobra: Local Flag Access (Go)

Retrieve flag values inside the `RunE` / `Run` function via `cmd.Flags().GetXxx()` instead of binding them to package-level variables with `StringVar`, `BoolVar`, etc.

```go
// BAD — package-level variable bound in init()
var nodeLocalCache string

func init() {
	rootCmd.PersistentFlags().StringVar(&nodeLocalCache, "node-local-cache", "", "path to cache dir")
}

// GOOD — read the flag where it's used
RunE: func(cmd *cobra.Command, args []string) error {
	nodeLocalCache, err := cmd.Flags().GetString("node-local-cache")
	if err != nil {
		return fmt.Errorf("missing node-local-cache flag: %w", err)
	}
	// ...
}
```

Flag *registration* (`cmd.Flags().String(...)` or `cmd.PersistentFlags().String(...)`) still belongs in `init()` or a setup function — only the *read* moves into the command body.

## Don't Extract Single-Use Helpers

Inline code is better than a tiny helper that's only called once. Extracting a function adds a jump for the reader — they have to scroll to the helper's definition, read it, then mentally substitute it back at the call site. For a single use, that's pure overhead.

A few extra lines at the call site is fine. Prefer verbosity at the call site to indirection through a small helper. Wait to extract until there's a second caller (or the operation is genuinely reusable beyond the current file).

This rule does not apply to:
- Public/exported helpers in shared packages — those exist to *be* called by other code.
- Operations whose name carries real semantic value (e.g. wrapping a complex condition behind `isAdmin(user)` to document intent).

Bad:
```go
func envBasenames(dirs []string) []string {
	out := make([]string, len(dirs))
	for i, d := range dirs {
		out[i] = filepath.Base(d)
	}
	return out
}

// ...later, single call site:
names := envBasenames(toLint)
```

Good:
```go
names := make([]string, len(toLint))
for i, d := range toLint {
	names[i] = filepath.Base(d)
}
```

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

## Your workspace already has repositories

`/workspace` is a persistent volume shared across this agent's sessions. The repositories the
agent is configured with are already cloned there, one per directory at `/workspace/<repo>`,
and authenticated. Look there first for the source you need before cloning anything yourself.
Treat each `/workspace/<repo>` as the primary checkout and make a worktree for your task
rather than working on its default branch.

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
