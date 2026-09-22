#!/usr/bin/env bash
# Runs on every devcontainer start, from the repository root.
set -euo pipefail

# Git hooks live in .githooks/; enable them once the folder exists.
if [[ -d .githooks ]]; then
    git config core.hooksPath .githooks
fi

# The docker inside this container hands its networks an MTU of 1500. When this container's
# own interface is smaller — a tunnel on the host is the usual reason — everything that
# crosses both stalls: a packet that fits here is too large for the tunnel, and nothing
# reports it. Large downloads die a few kilobytes in; small ones are fine, which is what
# makes it look like anything but the network.
interface="$(ip -o route show default | awk '{ print $5; exit }')"
mtu="$(cat "/sys/class/net/${interface}/mtu" 2> /dev/null || echo 1500)"

if [[ "$mtu" -lt 1500 ]]; then
    # Both keys, for the same reason they are both needed on the host: the first one governs
    # the default bridge, the second one every network created later.
    sudo tee /etc/docker/daemon.json > /dev/null <<JSON
{
    "mtu": ${mtu},
    "default-network-opts": {
        "bridge": {
            "com.docker.network.driver.mtu": "${mtu}"
        }
    }
}
JSON

    # That file is read when the daemon starts, and it is already running. Bring the bridge
    # it has down to size now — that is what image builds go through — and let the file take
    # care of the rest from the next start of this container onwards. The bridge exists only
    # once the daemon has come up, which is not guaranteed by the time this runs.
    if ip link show docker0 > /dev/null 2>&1; then
        sudo ip link set dev docker0 mtu "$mtu"
    fi
    echo "docker: mtu ${mtu}; networks made by compose follow after a container restart"
fi
