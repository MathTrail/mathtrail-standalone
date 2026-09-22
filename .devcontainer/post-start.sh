#!/usr/bin/env bash
# Runs on every devcontainer start, from the repository root.
set -euo pipefail

# Git hooks live in .githooks/; enable them once the folder exists.
if [[ -d .githooks ]]; then
    git config core.hooksPath .githooks
fi
