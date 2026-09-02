#!/usr/bin/env bash
set -euo pipefail

log() { echo "[tailscale-init] $*" >&2; }

log "waiting for tailscaled socket..."
for i in $(seq 1 30); do
    if [[ -S /var/run/tailscale/tailscaled.sock ]]; then
        log "tailscaled socket ready (waited ${i}s)"
        break
    fi
    sleep 1
    if [[ $i -eq 30 ]]; then
        log "ERROR: tailscaled socket did not appear after 30s"
        exit 1
    fi
done

id_token=$(cat "${TAILSCALE_TOKEN_FILE}")
log "token read (${#id_token} bytes)"

# JWT payload is base64url without padding; || true prevents set -e from exiting on padding
# warnings — the empty-client_id check below catches a real decode failure.
ts_audience=$(echo "${id_token}" \
  | cut -d. -f2 \
  | tr '_-' '/+' \
  | base64 -d 2>/dev/null \
  | jq -r 'if (.aud | type) == "array" then .aud[0] else .aud end // empty') || true

client_id="${ts_audience##*/}"
log "decoded aud=${ts_audience} client_id=${client_id}"

if [[ -z "${client_id}" ]]; then
    log "ERROR: could not extract client_id from token aud"
    exit 1
fi

hostname="agent-$(echo "${AGENT_NAME:-unknown}-${AGENT_THREAD:-0}" \
    | tr '[:upper:]' '[:lower:]' \
    | tr -cs 'a-z0-9-' '-' \
    | sed 's/-\+/-/g; s/^-//; s/-$//')"

log "running: tailscale up --client-id=${client_id} --advertise-tags=${TAILSCALE_TAG:-tag:mantis-shrimp} --hostname=${hostname}"
tailscale up \
    --client-id="${client_id}" \
    --id-token="${id_token}" \
    --advertise-tags="${TAILSCALE_TAG:-tag:mantis-shrimp}" \
    --hostname="${hostname}" \
    --reset

log "enrolled as ${hostname}"
tailscale status --json 2>/dev/null \
    | jq -r '"[tailscale-init] status: self=" + (.Self.DNSName // "unknown") + " ip=" + (.Self.TailscaleIPs[0] // "none")' >&2 || true
