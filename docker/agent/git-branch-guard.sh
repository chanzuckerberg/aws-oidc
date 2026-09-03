#!/usr/bin/env bash
# PreToolUse hook for Claude Code. Reads the hook JSON on stdin and blocks any
# Bash command that would commit to, or push directly to, a repository's primary
# branch (main, master, or the remote's default). Agents must branch and open a
# pull request. Exits 2 to block the tool call; 0 to allow it. Claude Code fires
# PreToolUse hooks even in bypassPermissions mode.

set -euo pipefail

input=$(cat)

tool_name=$(echo "${input}" | jq -r '.tool_name // empty')
if [[ "${tool_name}" != "Bash" ]]; then
    exit 0
fi

command=$(echo "${input}" | jq -r '.tool_input.command // empty')
if [[ -z "${command}" ]]; then
    exit 0
fi

if ! grep -qE '(^|[;&|`(]|[[:space:]])git([[:space:]]|$)' <<<"${command}"; then
    exit 0
fi

# Match commit and push as the git subcommand, tolerating global options
# (-C path, -c key=val, --flag[=val]) between "git" and the subcommand.
git_global='([[:space:]]+(-C[[:space:]]+[^[:space:]]+|-c[[:space:]]+[^[:space:]]+|--[A-Za-z-]+(=[^[:space:]]+)?))*'
is_commit=false
is_push=false
if grep -qE "(^|[;&|\`(]|[[:space:]])git${git_global}[[:space:]]+commit([[:space:]]|$)" <<<"${command}"; then
    is_commit=true
fi
if grep -qE "(^|[;&|\`(]|[[:space:]])git${git_global}[[:space:]]+push([[:space:]]|$)" <<<"${command}"; then
    is_push=true
fi

if [[ "${is_commit}" == "false" && "${is_push}" == "false" ]]; then
    exit 0
fi

# Resolve the repository this command runs against: prefer an explicit
# "git -C <path>", otherwise the tool's working directory.
cwd=$(echo "${input}" | jq -r '.cwd // empty')
repo_dir="${cwd:-$(pwd)}"
if grep -qE '(^|[[:space:]])git[[:space:]]+-C[[:space:]]+[^[:space:]]+' <<<"${command}"; then
    repo_dir=$(grep -oE '(^|[[:space:]])git[[:space:]]+-C[[:space:]]+[^[:space:]]+' <<<"${command}" | head -1 | awk '{print $NF}' || true)
fi

git_c=(git -C "${repo_dir}")

protected_names="main master"
default_branch=$("${git_c[@]}" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's#^origin/##' || true)
if [[ -n "${default_branch}" ]]; then
    protected_names="${protected_names} ${default_branch}"
fi

is_protected() {
    local name="${1:-}" candidate
    [[ -n "${name}" ]] || return 1
    for candidate in ${protected_names}; do
        [[ "${name}" == "${candidate}" ]] && return 0
    done
    return 1
}

current_branch=$("${git_c[@]}" symbolic-ref --quiet --short HEAD 2>/dev/null || true)

deny() {
    echo "git-branch-guard: $1" >&2
    echo "git-branch-guard: agents must not put work on a primary branch directly." >&2
    echo "git-branch-guard: branch first, then open a PR, e.g. 'git switch -c <branch>', push, and 'gh pr create'." >&2
    exit 2
}

if [[ "${is_commit}" == "true" ]] && is_protected "${current_branch}"; then
    deny "refusing to commit on primary branch '${current_branch}'"
fi

if [[ "${is_push}" == "true" ]]; then
    push_segment=$(grep -oE "git${git_global}[[:space:]]+push[^;&|]*" <<<"${command}" | head -1 || true)
    refspecs=${push_segment#*push}

    ref_count=0
    for tok in ${refspecs}; do
        [[ "${tok}" == -* ]] && continue
        ref_count=$((ref_count + 1))
        dst=${tok##*:}
        dst=${dst#+}
        if [[ "${dst}" == "HEAD" ]]; then
            dst="${current_branch}"
        fi
        if is_protected "${dst}"; then
            deny "refusing to push to primary branch '${dst}'"
        fi
    done

    # A bare push (no refspec, at most a remote) sends the current branch. Block
    # it when that branch is primary.
    if [[ ${ref_count} -le 1 ]] && is_protected "${current_branch}"; then
        deny "refusing to push primary branch '${current_branch}'"
    fi
fi

exit 0
