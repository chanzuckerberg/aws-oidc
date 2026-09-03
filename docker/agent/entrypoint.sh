#!/usr/bin/env bash
set -euo pipefail

log() { echo "[entrypoint] $*" >&2; }

if [[ -n "${TAILSCALE_TOKEN_FILE:-}" && -f "${TAILSCALE_TOKEN_FILE}" ]]; then
    tun_args=()
    tailscale_mode="kernel TUN"
    if [[ -c /dev/net/tun || -c /dev/tun ]]; then
        log "found TUN device"
    else
        tailscale_mode="userspace networking"
        tun_args=(--tun=userspace-networking)
    fi

    while true; do
        log "starting tailscaled (${tailscale_mode})"
        rm -f /var/run/tailscale/tailscaled.sock
        tailscaled \
            --state=/workspace/.tailscale/state \
            --statedir=/var/lib/tailscale \
            "${tun_args[@]}" \
            &
        tailscaled_pid=$!

        for i in $(seq 1 30); do
            if [[ -S /var/run/tailscale/tailscaled.sock ]] && \
                tailscale status --peers=false >/dev/null 2>&1; then
                log "daemon ready (waited ${i}s)"
                break 2
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

exec "$@"
