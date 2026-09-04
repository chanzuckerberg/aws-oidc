#!/usr/bin/env bash
set -euo pipefail

log() { echo "[entrypoint] $*" >&2; }

# run_as_agent runs a command as the agent user (uid 1000) with its workspace
# home, whether the entrypoint itself is running as root (tailscale pods) or as
# agent (everything else).
run_as_agent() {
    if [[ "$(id -u)" -eq 0 ]]; then
        runuser -u agent -- env HOME=/workspace "$@"
    else
        env HOME=/workspace "$@"
    fi
}

# wait_for_dns blocks until a public name resolves or a short timeout elapses. tailscale up
# points resolv.conf at MagicDNS, whose forwarder needs a moment before it answers; without
# this the repository clone and plugin install below, and a session that connects at once, can
# hit a window where every lookup fails.
wait_for_dns() {
    local host="${DNS_READY_HOST:-github.com}" i
    for i in $(seq 1 20); do
        if getent hosts "${host}" >/dev/null 2>&1; then
            log "DNS ready (${host} resolved after $((i - 1))s)"
            return 0
        fi
        sleep 1
    done
    log "WARN: DNS not ready after 20s (${host} did not resolve); continuing"
}

# ensure_agent_repositories clones the repositories in AGENT_REPOSITORIES into /workspace so
# sessions find the source already checked out. It is idempotent: a repository whose checkout
# already exists is left untouched, which is what happens after the first session on a
# persistent workspace. A clone that fails logs a warning and never stops the agent booting.
ensure_agent_repositories() {
    [[ -n "${AGENT_REPOSITORIES:-}" ]] || return 0
    command -v gh >/dev/null 2>&1 || { log "WARN: gh not found; skipping repository clones"; return 0; }
    local spec repo dest
    for spec in ${AGENT_REPOSITORIES}; do
        if [[ ! "${spec}" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
            log "WARN: ignoring malformed repository '${spec}' (want owner/repo)"
            continue
        fi
        repo="${spec##*/}"
        dest="/workspace/${repo}"
        if [[ -e "${dest}/.git" ]]; then
            log "repository ${spec} already present at ${dest}"
            continue
        fi
        log "cloning ${spec} into ${dest}"
        if run_as_agent gh repo clone "${spec}" "${dest}" >/dev/null 2>&1; then
            log "cloned ${spec}"
        else
            log "WARN: failed to clone ${spec} (needs network egress and GitHub App access)"
        fi
    done
}

# ensure_agent_plugins installs the shared CZI Claude Code plugins so every agent has them by
# default. The czi-ai-toolchain marketplace splits its content into czi-general and czi-infra;
# the earlier single ai-toolchain plugin no longer exists. Claude Code's CLI does not
# auto-install plugins from managed settings, so the install runs here. State lives on the
# persistent workspace volume; a marker keeps this to a single successful install per volume,
# and is written only when every plugin installed so a partial failure is retried next boot.
ensure_agent_plugins() {
    command -v claude >/dev/null 2>&1 || return 0
    local marker=/workspace/.claude/.czi-plugins-installed
    [[ -f "${marker}" ]] && return 0
    log "adding marketplace czi-ai-toolchain"
    run_as_agent claude plugin marketplace add chanzuckerberg/ai-toolchain >/dev/null 2>&1 || true
    local plugin all_ok=1
    for plugin in czi-general czi-infra; do
        log "installing plugin ${plugin}@czi-ai-toolchain"
        if run_as_agent claude plugin install "${plugin}@czi-ai-toolchain" >/dev/null 2>&1; then
            log "installed ${plugin}"
        else
            all_ok=0
            log "WARN: failed to install ${plugin}@czi-ai-toolchain (needs network egress and GitHub App read access to chanzuckerberg/ai-toolchain)"
        fi
    done
    if [[ "${all_ok}" -eq 1 ]]; then
        run_as_agent bash -c 'mkdir -p /workspace/.claude && touch /workspace/.claude/.czi-plugins-installed'
        log "czi plugins installed"
    fi
}

if [[ "$(id -u)" -eq 0 ]]; then
    : > /etc/environment
    : > /etc/profile.d/agent-env.sh
    while IFS= read -r -d '' pair; do
        key=${pair%%=*}
        val=${pair#*=}
        case "${key}" in
            HOME|AWS_*|ANTHROPIC_*|GITHUB_*|GIT_*|AGENT_*)
                printf '%s=%s\n' "${key}" "${val}" >> /etc/environment
                printf "export %s='%s'\n" "${key}" "${val//\'/\'\\\'\'}" >> /etc/profile.d/agent-env.sh
                ;;
        esac
    done < /proc/self/environ

    if [[ ! -e /workspace/.bashrc ]] || ! grep -q '# agent-env' /workspace/.bashrc 2>/dev/null; then
        tmp=$(mktemp)
        printf '# agent-env\n[ -r /etc/profile.d/agent-env.sh ] && . /etc/profile.d/agent-env.sh\n' > "${tmp}"
        [[ -e /workspace/.bashrc ]] && cat /workspace/.bashrc >> "${tmp}"
        cat "${tmp}" > /workspace/.bashrc
        rm -f "${tmp}"
        chown 1000:1000 /workspace/.bashrc
    fi
fi

if [[ -n "${TAILSCALE_TOKEN_FILE:-}" && -f "${TAILSCALE_TOKEN_FILE}" ]]; then
    tun_args=()
    tailscale_mode="kernel TUN"
    if [[ -c /dev/net/tun || -c /dev/tun ]]; then
        log "found TUN device"
    else
        tailscale_mode="userspace networking"
        tun_args=(--tun=userspace-networking)
    fi

    mkdir -p /workspace/.tailscale

    while true; do
        log "starting tailscaled (${tailscale_mode})"
        rm -f /var/run/tailscale/tailscaled.sock
        tailscaled \
            --state=/workspace/.tailscale/state \
            --statedir=/workspace/.tailscale \
            "${tun_args[@]}" \
            &
        tailscaled_pid=$!

        for i in $(seq 1 30); do
            if [[ -S /var/run/tailscale/tailscaled.sock ]]; then
                sleep 1
                if kill -0 "${tailscaled_pid}" 2>/dev/null; then
                    log "daemon ready (waited ${i}s)"
                    break 2
                fi
            fi
            if ! kill -0 "${tailscaled_pid}" 2>/dev/null; then
                break
            fi
            sleep 1
        done

        if [[ "${tailscale_mode}" == "kernel TUN" ]]; then
            log "kernel TUN startup failed; falling back to userspace networking"
            wait "${tailscaled_pid}" 2>/dev/null || true
            tailscale_mode="userspace networking"
            tun_args=(--tun=userspace-networking)
            continue
        fi

        log "ERROR: tailscaled did not become ready; continuing without tailscale"
        exec "$@"
    done

    id_token=$(cat "${TAILSCALE_TOKEN_FILE}")
    log "token read (${#id_token} bytes)"

    ts_audience=$(echo "${id_token}" \
        | cut -d. -f2 | tr '_-' '/+' \
        | base64 -d 2>/dev/null \
        | jq -r 'if (.aud | type) == "array" then .aud[0] else .aud end // empty') || true
    client_id="${ts_audience##*/}"
    log "decoded aud=${ts_audience} client_id=${client_id}"

    if [[ -z "${client_id}" ]]; then
        log "ERROR: could not extract client_id from token aud — continuing without tailscale"
    else
        local_part="${AGENT_OWNER_EMAIL%%@*}"
        hostname="agent-$(echo "${local_part:-unknown}-${AGENT_NAME:-unknown}-${AGENT_WORKSPACE:-0}" \
            | tr '[:upper:]' '[:lower:]' \
            | tr -cs 'a-z0-9-' '-' \
            | sed 's/-\+/-/g; s/^-//; s/-$//')"
        log "enrolling as ${hostname} with tags ${TAILSCALE_TAG:-tag:mantis-shrimp}"
        if tailscale up \
                --client-id="${client_id}" \
                --id-token="${id_token}" \
                --advertise-tags="${TAILSCALE_TAG:-tag:mantis-shrimp}" \
                --hostname="${hostname}" \
                --reset; then
            log "enrolled — $(tailscale ip 2>/dev/null || echo 'ip unknown')"
            tailscale set --ssh
            log "SSH enabled"
            wait_for_dns
        else
            log "ERROR: tailscale up failed — continuing without tailscale"
        fi
    fi
elif [[ -n "${TAILSCALE_TOKEN_FILE:-}" ]]; then
    log "WARN: TAILSCALE_TOKEN_FILE=${TAILSCALE_TOKEN_FILE} set but file not found — skipping"
fi

ensure_agent_repositories
ensure_agent_plugins

exec "$@"
