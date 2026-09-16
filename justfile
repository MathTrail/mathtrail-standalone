# MathTrail standalone: task runner. Every recipe runs inside the devcontainer (see CLAUDE.md).

set shell := ["bash", "-euo", "pipefail", "-c"]

# List all recipes
default:
    @just --list

# -- T03 spike: protocol and widget ----------------------------------------

# Build the widget, run the spike server, and open an unauthenticated cloudflared
# tunnel to it. Ctrl-C stops both the server and the tunnel.
spike-a port="8081": (_spike-a-run port "0")

# Same as spike-a, but restricts the server to protocol version 2026-07-28 only
# (T03: "включать после первого замера" — run this after a first plain spike-a).
spike-a-strict port="8081": (_spike-a-run port "1")

# Shared implementation of spike-a / spike-a-strict. Not meant to be called
# directly: `just` has no named-argument syntax, so `just spike-a strict=1` binds
# the literal string "strict=1" to `port` instead of failing — hence two thin
# recipes above instead of one with a second parameter.
_spike-a-run port strict:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! [[ "{{ port }}" =~ ^[0-9]+$ ]]; then
        echo "spike-a: '{{ port }}' is not a port number." >&2
        echo "Usage: just spike-a [port]   or   just spike-a-strict [port]" >&2
        exit 1
    fi

    cd spike/protocol/widget
    npm install
    npm run build
    cp dist/index.html ../embed/widget.html
    cd ..
    go build -o /tmp/mathtrail-spike-protocol .

    PORT={{ port }} STRICT_PROTOCOL={{ strict }} /tmp/mathtrail-spike-protocol &
    server_pid=$!
    trap 'kill "$server_pid" "${tunnel_pid:-}" 2>/dev/null || true' EXIT

    sleep 1
    if ! kill -0 "$server_pid" 2>/dev/null; then
        echo "spike-a: the server exited immediately; see its log above." >&2
        exit 1
    fi

    : > /tmp/mathtrail-spike-tunnel.log
    cloudflared tunnel --url "http://localhost:{{ port }}" --http-host-header "localhost:{{ port }}" >> /tmp/mathtrail-spike-tunnel.log 2>&1 &
    tunnel_pid=$!

    echo "Waiting for the tunnel URL..."
    url=""
    for _ in $(seq 1 30); do
        url=$(grep -oE 'https://[a-zA-Z0-9-]+\.trycloudflare\.com' /tmp/mathtrail-spike-tunnel.log | head -1 || true)
        [ -n "$url" ] && break
        sleep 1
    done
    if [ -z "$url" ]; then
        echo "Tunnel URL not found within 30s; check /tmp/mathtrail-spike-tunnel.log" >&2
    else
        echo ""
        echo "MCP endpoint:   ${url}/mcp"
        echo "Health check:   ${url}/healthz"
        echo ""
    fi

    wait "$server_pid" "$tunnel_pid"

# -- T04 spike: OAuth stub ---------------------------------------------------

# Build and run the stub OAuth authorization server, and open an unauthenticated
# cloudflared tunnel to it. Unlike spike-a, the tunnel is NOT given a fake Host
# header: the server derives its own issuer URL from the real Host on each
# request (see spike/oauth/oauth.go), so it needs to see the real tunnel hostname.
spike-b port="8082":
    #!/usr/bin/env bash
    set -euo pipefail
    if ! [[ "{{ port }}" =~ ^[0-9]+$ ]]; then
        echo "spike-b: '{{ port }}' is not a port number." >&2
        echo "Usage: just spike-b [port]" >&2
        exit 1
    fi

    cd spike/oauth
    go build -o /tmp/mathtrail-spike-oauth .

    PORT={{ port }} /tmp/mathtrail-spike-oauth &
    server_pid=$!
    trap 'kill "$server_pid" "${tunnel_pid:-}" 2>/dev/null || true' EXIT

    sleep 1
    if ! kill -0 "$server_pid" 2>/dev/null; then
        echo "spike-b: the server exited immediately; see its log above." >&2
        exit 1
    fi

    : > /tmp/mathtrail-spike-oauth-tunnel.log
    cloudflared tunnel --url "http://localhost:{{ port }}" >> /tmp/mathtrail-spike-oauth-tunnel.log 2>&1 &
    tunnel_pid=$!

    echo "Waiting for the tunnel URL..."
    url=""
    for _ in $(seq 1 30); do
        url=$(grep -oE 'https://[a-zA-Z0-9-]+\.trycloudflare\.com' /tmp/mathtrail-spike-oauth-tunnel.log | head -1 || true)
        [ -n "$url" ] && break
        sleep 1
    done
    if [ -z "$url" ]; then
        echo "Tunnel URL not found within 30s; check /tmp/mathtrail-spike-oauth-tunnel.log" >&2
    else
        echo ""
        echo "MCP endpoint:   ${url}/mcp"
        echo "Health check:   ${url}/healthz"
        echo ""
    fi

    wait "$server_pid" "$tunnel_pid"
