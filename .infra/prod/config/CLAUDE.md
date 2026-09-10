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

## Writing a PR consistent with the Orion guidelines

Orion reads a PR on four dimensions and runs deterministic checks. Use these guidelines while writing or revising a pull request.

### What a strong PR looks like

Orion gives a light-touch read on four dimensions (🟢 great / 🟡 good / 🟠 could improve):

- **Intent** -- Is the PR's purpose easy to understand from its title and description (what changed, and why)?
- **Scope** -- Is the change focused and cohesive, or does it bundle unrelated concerns or drive-by edits that would be easier to review separately?
- **Hygiene** -- Does it fit repo conventions and stay reviewable -- no unexplained giant dumps, silent dependency/data changes, or obvious style slips?
- **Verification** -- Is there some sign the change was tested or otherwise checked (tests, a described manual check, sample output)?

Aim for clear intent, focused scope, clean hygiene, and visible verification. These describe an ideal, not a bar every PR must clear -- some changes are unavoidably large or complex -- so treat them as guidance, not targets.

### Mechanical checks to avoid

- `banned_future_import` -- Adds `from __future__ import ...`, which is banned in this py3.9+ repo.
- `sys_path_insert` -- Adds sys.path.insert/append/+= hacks instead of proper imports.
- `bare_except` -- Adds a bare `except:` clause.
- `inline_import_in_function` -- Many indented (in-function) imports, typical of bulk dumps skipping top-level imports. A few are legit (conditional/optional), so this counts.
- `relative_import` -- Adds a relative import (`from . ...`); this repo requires absolute imports.
- `legacy_typing_import` -- Imports typing.List/Dict/Optional/Union/... instead of py3.9 builtins or `|`.
- `legacy_type_directive` -- Adds a `# type: ignore` / `# mypy:` / `# pyright:` directive; this repo uses `ty` (`# ty: ignore`).
- `checkpoint_backcompat_shim` -- Adds/extends `maybe_update_checkpoint_for_backwards_compatibility`, a checkpoint back-compat hook this repo bans (it silently no-ops for sharded checkpoints, so it gives a false sense of safety).
- `title_equals_branch` -- PR title is just the branch name (or its last segment), separators aside.
- `non_descriptive_title` -- Title has fewer than a few real words after stripping prefixes.
- `empty_body` -- PR description is empty.
- `template_only_body` -- Description is just the unmodified PR template -- headers, checklist, and the HTML-comment placeholder, with no real prose added.
- `thin_body_on_large_diff` -- Big diff with a near-empty / template-only description.
- `silent_lockfile_or_dep_change` -- A lockfile or project manifest changed (dependencies, env vars, or build/tool config) with no mention in the description.
- `many_top_level_dirs` -- Touches many top-level directories (possible bundled concerns).
- `docs_readme_drive_by` -- Edits a top-level README alongside unrelated code (scope-creep drive-by).
- `oversized_file_change` -- A single source file changes by >= LARGE_FILE_LOC lines (hard to review). Catches bulk dumps even when GitHub omits the patch for the huge file.
- `bulk_data_or_generated` -- Adds large data / generated files (bulk not accounted for).

### Suggested prompt

> Review my PR's title, description, and diff against the Orion guide above. For each dimension (intent, scope, hygiene, verification), tell me what would strengthen it, and flag any of the mechanical checks I'm tripping. Then help me revise.
