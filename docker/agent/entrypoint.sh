#!/usr/bin/env bash
set -euo pipefail

log() { echo "[entrypoint] $*" >&2; }

if [[ -n "${TAILSCALE_TOKEN_FILE:-}" && -f "${TAILSCALE_TOKEN_FILE}" ]]; then
    log "tailscale token found at ${TAILSCALE_TOKEN_FILE}"

    log "starting tailscaled (userspace-networking)"
    tailscaled --tun=userspace-networking --socks5-server=localhost:1055 &
    TAILSCALED_PID=$!
    log "tailscaled started (pid=${TAILSCALED_PID})"

    sleep 1

    id_token=$(cat "${TAILSCALE_TOKEN_FILE}")
    token_len=${#id_token}
    log "token read (${token_len} bytes)"

    # JWT payload is base64url without padding; the || true prevents set -e from
    # exiting on a padding warning — the client_id check below catches a real failure.
    ts_audience=$(echo "${id_token}" \
      | cut -d. -f2 \
      | tr '_-' '/+' \
      | base64 -d 2>/dev/null \
      | jq -r 'if (.aud | type) == "array" then .aud[0] else .aud end // empty') || true

    client_id="${ts_audience##*/}"
    log "decoded aud=${ts_audience} client_id=${client_id}"

    if [[ -z "${client_id}" ]]; then
        log "ERROR: could not extract client_id from token aud — skipping tailscale up"
    else
        hostname="agent-$(echo "${AGENT_NAME:-unknown}-${AGENT_THREAD:-0}" \
            | tr '[:upper:]' '[:lower:]' \
            | tr -cs 'a-z0-9-' '-' \
            | sed 's/-\+/-/g; s/^-//; s/-$//')"
        log "running: tailscale up --client-id=${client_id} --advertise-tags=${TAILSCALE_TAG:-tag:mantis-shrimp} --hostname=${hostname}"
        if tailscale up \
                --client-id="${client_id}" \
                --id-token="${id_token}" \
                --advertise-tags="${TAILSCALE_TAG:-tag:mantis-shrimp}" \
                --hostname="${hostname}" \
                --reset; then
            log "tailscale up succeeded — enrolled as ${hostname}"
            tailscale status --json 2>/dev/null \
                | jq -r '"[entrypoint] tailscale status: self=" + (.Self.DNSName // "unknown") + " ip=" + (.Self.TailscaleIPs[0] // "none")' >&2 || true
        else
            log "ERROR: tailscale up failed (exit $?) — pod will run without tailnet access"
        fi
    fi
elif [[ -n "${TAILSCALE_TOKEN_FILE:-}" ]]; then
    log "WARN: TAILSCALE_TOKEN_FILE=${TAILSCALE_TOKEN_FILE} set but file not found — skipping tailscale enrollment"
else
    log "TAILSCALE_TOKEN_FILE not set — skipping tailscale enrollment"
fi

log "starting claude"
exec claude "$@"
