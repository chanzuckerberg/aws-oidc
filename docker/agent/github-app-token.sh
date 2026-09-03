#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_APP_ID:?GITHUB_APP_ID is not set}"
: "${GITHUB_APP_PRIVATE_KEY_FILE:?GITHUB_APP_PRIVATE_KEY_FILE is not set}"

installation_id=""
while [[ $# -gt 0 ]]; do
  case "$1" in
  --installation-id)
    installation_id="${2:-}"
    shift 2
    ;;
  --installation-id=*)
    installation_id="${1#*=}"
    shift
    ;;
  *)
    shift
    ;;
  esac
done
: "${installation_id:=${GITHUB_APP_INSTALLATION_ID:?GITHUB_APP_INSTALLATION_ID is not set}}"

api_url="${GITHUB_API_URL:-https://api.github.com}"
cache_dir="${GITHUB_APP_TOKEN_CACHE_DIR:-${HOME:-/tmp}/.cache/github-app}"
cache_file="${cache_dir}/token-${installation_id}"
refresh_skew_seconds=300

now=$(date -u +%s)

if [[ -r "${cache_file}" ]]; then
  read -r cached_expiry cached_token <"${cache_file}" || true
  if [[ -n "${cached_token:-}" && "${cached_expiry:-0}" =~ ^[0-9]+$ ]] &&
    ((now < cached_expiry - refresh_skew_seconds)); then
    printf '%s\n' "${cached_token}"
    exit 0
  fi
fi

base64url() {
  openssl base64 -A | tr '+/' '-_' | tr -d '='
}

header=$(printf '{"alg":"RS256","typ":"JWT"}' | base64url)
payload=$(printf '{"iat":%d,"exp":%d,"iss":"%s"}' "$((now - 60))" "$((now + 540))" "${GITHUB_APP_ID}" | base64url)
signature=$(printf '%s.%s' "${header}" "${payload}" |
  openssl dgst -sha256 -sign "${GITHUB_APP_PRIVATE_KEY_FILE}" -binary | base64url)

response=$(curl -fsSL -X POST \
  -H "Authorization: Bearer ${header}.${payload}.${signature}" \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "${api_url}/app/installations/${installation_id}/access_tokens")

token=$(printf '%s' "${response}" | jq -re '.token')
expires_at=$(printf '%s' "${response}" | jq -re '.expires_at')

umask 077
mkdir -p "${cache_dir}"
tmp_file="${cache_file}.$$"
printf '%s %s\n' "$(date -u -d "${expires_at}" +%s)" "${token}" >"${tmp_file}"
mv -f "${tmp_file}" "${cache_file}"

printf '%s\n' "${token}"
