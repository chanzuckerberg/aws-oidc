#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${GH_TOKEN:-}" && -z "${GITHUB_TOKEN:-}" && -n "${GITHUB_APP_ID:-}" ]]; then
  GH_TOKEN=$(github-app-token)
  export GH_TOKEN
fi

exec /usr/local/lib/gh/bin/gh "$@"
