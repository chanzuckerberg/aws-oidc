#!/usr/bin/env bash
set -euo pipefail

# gh authenticates with one token, and a GitHub App installation token only reaches one
# organization. When no token is already set, mint one for the installation that can reach
# the target repository's owner. Resolve the owner from an explicit --repo/-R, from a
# repository or API path named on the command line, from GH_REPO, or from the current
# checkout's origin remote, look it up in GITHUB_APP_INSTALLATION_MAP, and fall back to the
# default installation when it is not listed or cannot be determined.
if [[ -z "${GH_TOKEN:-}" && -z "${GITHUB_TOKEN:-}" && -n "${GITHUB_APP_ID:-}" ]]; then
  owner=""

  args=("$@")
  for ((i = 0; i < ${#args[@]}; i++)); do
    case "${args[i]}" in
    -R | --repo)
      owner="${args[$((i + 1))]:-}"
      owner="${owner%%/*}"
      break
      ;;
    --repo=*)
      owner="${args[i]#--repo=}"
      owner="${owner%%/*}"
      break
      ;;
    esac
  done

  # `gh repo clone owner/repo` names the repository as a positional argument, which is how
  # the entrypoint clones. Only the repo subcommands are scanned, because elsewhere a flag
  # value such as `--head user/branch` has the same shape as owner/repo.
  if [[ -z "${owner}" && "${args[0]:-}" == "repo" ]]; then
    case "${args[1]:-}" in
    archive | clone | delete | edit | fork | rename | set-default | sync | view)
      for ((i = 2; i < ${#args[@]}; i++)); do
        if [[ "${args[i]}" == "--" ]]; then
          break
        fi
        if [[ "${args[i]}" =~ ^([A-Za-z0-9_.-]+)/[A-Za-z0-9_.-]+$ ]]; then
          owner="${BASH_REMATCH[1]}"
          break
        fi
      done
      ;;
    esac
  fi

  if [[ -z "${owner}" && "${args[0]:-}" == "api" ]]; then
    for ((i = 1; i < ${#args[@]}; i++)); do
      path="${args[i]#/}"
      case "${path}" in
      repos/*/* | orgs/*)
        path="${path#*/}"
        owner="${path%%/*}"
        break
        ;;
      esac
    done
  fi

  if [[ -z "${owner}" && -n "${GH_REPO:-}" ]]; then
    owner="${GH_REPO%%/*}"
  fi

  if [[ -z "${owner}" ]]; then
    remote_url=$(git config --get remote.origin.url 2>/dev/null || true)
    case "${remote_url}" in
    *github.com[:/]*)
      owner="${remote_url#*github.com}"
      owner="${owner#[:/]}"
      owner="${owner%%/*}"
      ;;
    esac
  fi

  installation_id="${GITHUB_APP_INSTALLATION_ID:-}"
  if [[ -n "${owner}" && -n "${GITHUB_APP_INSTALLATION_MAP:-}" ]]; then
    owner=$(printf '%s' "${owner}" | tr '[:upper:]' '[:lower:]')
    for entry in ${GITHUB_APP_INSTALLATION_MAP//,/ }; do
      key=$(printf '%s' "${entry%%=*}" | tr '[:upper:]' '[:lower:]')
      if [[ "${key}" == "${owner}" ]]; then
        installation_id="${entry#*=}"
        break
      fi
    done
  fi

  if [[ -n "${installation_id}" ]]; then
    GH_TOKEN=$(github-app-token --installation-id "${installation_id}")
    export GH_TOKEN
  fi
fi

exec /usr/local/lib/gh/bin/gh "$@"
