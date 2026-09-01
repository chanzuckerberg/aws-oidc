#!/usr/bin/env bash
# PreToolUse hook for Claude Code. Reads the hook JSON on stdin, inspects any
# ssh or tailscale ssh invocation, and blocks if the login user differs from
# AGENT_SSH_USER. Exits 2 to block the tool call; 0 to allow it. Claude Code
# fires PreToolUse hooks even in bypassPermissions mode.

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

is_ssh=false
if echo "${command}" | grep -qE '(^|[;&|`(]|\s)(ssh|tailscale\s+ssh)\s'; then
    is_ssh=true
fi

if [[ "${is_ssh}" == "false" ]]; then
    exit 0
fi

expected="${AGENT_SSH_USER:-}"

if [[ -z "${expected}" ]]; then
    echo "ssh-guard: AGENT_SSH_USER is not set — blocking SSH to prevent unaudited access" >&2
    exit 2
fi

if [[ "${expected}" == "root" ]]; then
    echo "ssh-guard: root is not allowed as AGENT_SSH_USER — blocking" >&2
    exit 2
fi

# Extract the login user from -l user or user@host forms. Both are common in
# ssh invocations. We check either syntax; missing user in user@host means the
# command would use the current OS user, which is "agent" (uid 1000), not the
# expected SSH user, so we block that too.
login_user=""

# Check -l flag: ssh -l user host or ssh ... -l user ...
if echo "${command}" | grep -qE '\s-l\s+\S+'; then
    login_user=$(echo "${command}" | grep -oE '\-l\s+\S+' | head -1 | awk '{print $2}')
fi

# Check user@host form if -l was not found.
if [[ -z "${login_user}" ]]; then
    # Match ssh user@host — extract the user part before @.
    if echo "${command}" | grep -qE '(ssh|tailscale ssh)\s[^@]*\s(\S+@\S+)'; then
        login_user=$(echo "${command}" \
            | grep -oE '(ssh|tailscale ssh)\s[^@]*\s(\S+@\S+)' \
            | grep -oE '\S+@\S+' \
            | head -1 \
            | cut -d@ -f1)
    fi
fi

if [[ -z "${login_user}" ]]; then
    echo "ssh-guard: could not determine login user from command — blocking" >&2
    echo "ssh-guard: use 'ssh ${expected}@<host>' or 'ssh -l ${expected} <host>'" >&2
    exit 2
fi

if [[ "${login_user}" == "root" ]]; then
    echo "ssh-guard: login as root is not allowed — blocking" >&2
    exit 2
fi

if [[ "${login_user}" != "${expected}" ]]; then
    echo "ssh-guard: login user '${login_user}' does not match expected '${expected}' — blocking" >&2
    exit 2
fi

exit 0
