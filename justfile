# MathTrail standalone: task runner. Every recipe runs inside the devcontainer (see CLAUDE.md).

set shell := ["bash", "-euo", "pipefail", "-c"]

# List all recipes
default:
    @just --list
