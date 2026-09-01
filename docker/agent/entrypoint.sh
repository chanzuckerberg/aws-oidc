#!/usr/bin/env bash
set -euo pipefail

if [[ -n "${TAILSCALE_TOKEN_FILE:-}" && -f "${TAILSCALE_TOKEN_FILE}" ]]; then
    tailscaled --tun=userspace-networking --socks5-server=localhost:1055 &
    TAILSCALED_PID=$!

    sleep 1

    id_token=$(cat "${TAILSCALE_TOKEN_FILE}")

    # JWT payload is base64url without padding; the || true prevents set -e from
    # exiting on a padding warning — the client_id check below catches a real failure.
    ts_audience=$(echo "${id_token}" \
      | cut -d. -f2 \
      | tr '_-' '/+' \
      | base64 -d 2>/dev/null \
      | jq -r 'if (.aud | type) == "array" then .aud[0] else .aud end // empty') || true

    client_id="${ts_audience##*/}"

    if [[ -z "${client_id}" ]]; then
        echo "entrypoint: could not extract client_id from tailscale token aud — skipping tailscale up" >&2
    else
        hostname="agent-$(echo "${AGENT_NAME:-unknown}-${AGENT_THREAD:-0}" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9-' '-' | sed 's/-\+/-/g; s/^-//; s/-$//')"
        tailscale up \
            --client-id="${client_id}" \
            --id-token="${id_token}" \
            --advertise-tags="${TAILSCALE_TAG:-tag:mantis-shrimp}" \
            --hostname="${hostname}" \
            --reset
        echo "entrypoint: tailscale enrolled as ${hostname}" >&2
    fi
fi

exec claude "$@"
