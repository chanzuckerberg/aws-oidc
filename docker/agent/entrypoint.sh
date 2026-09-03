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

# ensure_agent_plugins installs the shared CZI ai-toolchain plugins so every
# agent has them by default. Claude Code's CLI does not auto-install plugins from
# managed settings, so the install runs here. State lives on the persistent
# workspace volume; a marker keeps this to a single install per volume.
ensure_agent_plugins() {
    command -v claude >/dev/null 2>&1 || return 0
    local marker=/workspace/.claude/.czi-ai-toolchain-installed
    [[ -f "${marker}" ]] && return 0
    log "installing default plugins from the czi-ai-toolchain marketplace"
    run_as_agent claude plugin marketplace add chanzuckerberg/ai-toolchain >/dev/null 2>&1 || true
    local plugin all_ok=true
    for plugin in czi-general czi-infra; do
        if run_as_agent claude plugin install "${plugin}@czi-ai-toolchain" >/dev/null 2>&1; then
            log "installed plugin ${plugin}@czi-ai-toolchain"
        else
            all_ok=false
            log "WARN: failed to install ${plugin}@czi-ai-toolchain"
        fi
    done
    if [[ "${all_ok}" == "true" ]]; then
        run_as_agent bash -c 'mkdir -p /workspace/.claude && touch /workspace/.claude/.czi-ai-toolchain-installed'
    else
        log "WARN: some czi-ai-toolchain plugins failed to install (needs network egress and GitHub App read access to chanzuckerberg/ai-toolchain)"
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
        hostname="agent-$(echo "${local_part:-unknown}-${AGENT_NAME:-unknown}-${AGENT_THREAD:-0}" \
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
        else
            log "ERROR: tailscale up failed — continuing without tailscale"
        fi
    fi
elif [[ -n "${TAILSCALE_TOKEN_FILE:-}" ]]; then
    log "WARN: TAILSCALE_TOKEN_FILE=${TAILSCALE_TOKEN_FILE} set but file not found — skipping"
fi

ensure_agent_plugins

exec "$@"
