#!/usr/bin/env bash
# Runs on the host — the machine VS Code itself runs on — before the container starts.
#
# Docker hands its networks an MTU of 1500. When the host reaches the internet through a
# tunnel, which in practice means a VPN, the real limit is lower, and packets that fit inside
# the container are dropped on the way out with nothing to say so: small requests answer,
# large downloads stop after a few kilobytes, and the failure looks like anything but the
# network. Giving the container a network built with the host's own MTU settles it, and asks
# nothing of whoever opens this repository.

set -euo pipefail

network="mathtrail-standalone"

# The interface the default route leaves by. Absent on hosts without iproute2 — macOS, for
# one — and then the Docker default is as good a guess as any.
mtu=1500
interface="$(ip -o route show default 2> /dev/null | awk '{ print $5; exit }' || true)"
if [[ -n "$interface" && -r "/sys/class/net/${interface}/mtu" ]]; then
    mtu="$(cat "/sys/class/net/${interface}/mtu")"
fi

existing="$(docker network inspect "$network" \
    --format '{{ index .Options "com.docker.network.driver.mtu" }}' 2> /dev/null || true)"

if [[ -z "$existing" ]]; then
    docker network create --opt "com.docker.network.driver.mtu=${mtu}" "$network" > /dev/null
    echo "network ${network}: created with mtu ${mtu}"
elif [[ "$existing" != "$mtu" ]]; then
    # A network in use cannot be changed, and taking it down from here would take the
    # container with it.
    echo "network ${network} has mtu ${existing}, this host now has ${mtu}." >&2
    echo "To pick the new one up: close the container, then docker network rm ${network}" >&2
fi
