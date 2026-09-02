#!/usr/bin/env bash
set -euo pipefail

log() { echo "[entrypoint] $*" >&2; }

if [[ -n "${TAILSCALE_TOKEN_FILE:-}" && -f "${TAILSCALE_TOKEN_FILE}" ]]; then
    log "starting tailscaled (kernel networking)"
    tailscaled \
        --state=/workspace/.tailscale/state \
        &
    TAILSCALED_PID=$!

    log "waiting for tailscaled socket..."
    for i in $(seq 1 30); do
        if [[ -S /var/run/tailscale/tailscaled.sock ]]; then
            log "socket ready (waited ${i}s)"
            break
        fi
        sleep 1
        if [[ $i -eq 30 ]]; then
            log "ERROR: socket did not appear after 30s — continuing without tailscale"
            exec "$@"
        fi
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
                --ssh \
                --reset; then
            log "enrolled — $(tailscale ip 2>/dev/null || echo 'ip unknown')"
        else
            log "ERROR: tailscale up failed — continuing without tailscale"
        fi
    fi
elif [[ -n "${TAILSCALE_TOKEN_FILE:-}" ]]; then
    log "WARN: TAILSCALE_TOKEN_FILE=${TAILSCALE_TOKEN_FILE} set but file not found — skipping"
fi

exec "$@"
