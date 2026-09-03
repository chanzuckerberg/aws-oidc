#!/usr/bin/env bash
set -euo pipefail

# git credential helper for the shared GitHub App. It routes each request to the App
# installation that can reach the requested repository's owner: an owner listed in
# GITHUB_APP_INSTALLATION_MAP uses that installation, every other owner uses the default
# GITHUB_APP_INSTALLATION_ID. A GitHub App installation token only reaches one
# organization, so this is how one agent reaches repositories in more than one org.
#
# Routing needs the owner, which git only sends when
# credential.https://github.com.useHttpPath is true (set in the image's git config).
#
# Exiting 0 without output makes git fall through to its other helpers, which is what
# should happen on an image running without a GitHub App configured.
[[ "${1:-}" == "get" ]] || exit 0
[[ -n "${GITHUB_APP_ID:-}" ]] || exit 0
[[ -n "${GITHUB_APP_PRIVATE_KEY_FILE:-}" ]] || exit 0

owner=""
while IFS='=' read -r key value; do
  [[ -z "${key}" ]] && break
  if [[ "${key}" == "path" ]]; then
    owner="${value%%/*}"
  fi
done

lower() { printf '%s' "$1" | tr '[:upper:]' '[:lower:]'; }

installation_id="${GITHUB_APP_INSTALLATION_ID:-}"
if [[ -n "${owner}" && -n "${GITHUB_APP_INSTALLATION_MAP:-}" ]]; then
  owner=$(lower "${owner}")
  for entry in ${GITHUB_APP_INSTALLATION_MAP//,/ }; do
    if [[ "$(lower "${entry%%=*}")" == "${owner}" ]]; then
      installation_id="${entry#*=}"
      break
    fi
  done
fi

[[ -n "${installation_id}" ]] || exit 0

# Assign first: a command substitution inside printf's arguments would swallow the failure
# and hand git an empty password instead of surfacing the API error.
token=$(github-app-token --installation-id "${installation_id}")
printf 'username=x-access-token\npassword=%s\n' "${token}"
