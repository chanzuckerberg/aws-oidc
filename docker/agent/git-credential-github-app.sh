#!/usr/bin/env bash
set -euo pipefail

# Exiting 0 without output makes git fall through to its other helpers, which is what should
# happen on an image running without a GitHub App configured.
[[ "${1:-}" == "get" ]] || exit 0
[[ -n "${GITHUB_APP_ID:-}" ]] || exit 0
[[ -n "${GITHUB_APP_INSTALLATION_ID:-}" ]] || exit 0
[[ -n "${GITHUB_APP_PRIVATE_KEY_FILE:-}" ]] || exit 0

# Assign first: a command substitution inside printf's arguments would swallow the failure and
# hand git an empty password instead of surfacing the API error.
token=$(github-app-token)
printf 'username=x-access-token\npassword=%s\n' "${token}"
