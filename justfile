# MathTrail: task runner.
# The ci-* recipes are the full checks, in the names the platform's other services
# use, so that the same command means the same thing in every repository.

set shell := ["bash", "-euo", "pipefail", "-c"]

BINARY := "bin/server"

# The build identity, stamped into the binary at link time. Computed once per
# run of just, so that a binary and the image built beside it carry the same
# words.
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo dev`
COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo unknown`
DATE := `date -u +%Y-%m-%dT%H:%M:%SZ`
SYMBOLS := "github.com/MathTrail/mathtrail-standalone/internal/version"
LDFLAGS := "-X " + SYMBOLS + ".Version=" + VERSION + " -X " + SYMBOLS + ".Commit=" + COMMIT + " -X " + SYMBOLS + ".Date=" + DATE

# List all recipes
default:
    @just --list

# -- Go ---------------------------------------------------------------------

# Format every Go file in place
fmt:
    gofmt -s -w .

# Fail if any Go file needs formatting
fmt-check:
    #!/usr/bin/env bash
    set -euo pipefail
    unformatted=$(gofmt -s -l .)
    if [ -n "$unformatted" ]; then
        echo "These files need gofmt -s; run just fmt:" >&2
        echo "$unformatted" >&2
        exit 1
    fi

# Lint
lint:
    golangci-lint run ./...

# Run the tests
test:
    go test ./... -count=1

# Build the server binary into bin/
build:
    go build -trimpath -ldflags "{{ LDFLAGS }}" -o {{ BINARY }} ./cmd/server

# Run the server from source, with logs a person can read
run:
    MATHTRAIL_LOG_FORMAT=console MATHTRAIL_LOG_LEVEL=debug go run ./cmd/server

# Generate the mocks (mockery, config in .mockery.yaml)
mocks:
    #!/usr/bin/env bash
    set -euo pipefail
    if grep -qE '^packages: *\{\}' .mockery.yaml; then
        echo "mocks: no interfaces are configured yet; nothing to generate."
        exit 0
    fi
    mockery

# -- The full checks --------------------------------------------------------

# Formatting and lint
ci-lint: fmt-check
    golangci-lint run ./...

# Tests with the race detector and a coverage report
ci-test:
    go test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic
    go tool cover -func=coverage.out | tail -1

# Fail if the committed mocks are not what mockery generates
ci-mocks-check:
    #!/usr/bin/env bash
    set -euo pipefail
    if grep -qE '^packages: *\{\}' .mockery.yaml; then
        echo "ci-mocks-check: no interfaces are configured yet; nothing to compare."
        exit 0
    fi
    mockery
    # --porcelain rather than diff: a mock generated for the first time is not
    # tracked yet, and a diff cannot see a file git has never heard of.
    changed=$(git status --porcelain -- "*mocks*")
    if [ -n "$changed" ]; then
        echo "Mocks are outdated: run just mocks and commit the result." >&2
        echo "$changed" >&2
        exit 1
    fi

# -- Container --------------------------------------------------------------

# Build the runtime image
docker-build tag="mathtrail:dev":
    docker build \
        --build-arg VERSION="{{ VERSION }}" \
        --build-arg COMMIT="{{ COMMIT }}" \
        --build-arg DATE="{{ DATE }}" \
        -t {{ tag }} .

# Build the image and run it on port 8080
docker-run tag="mathtrail:dev": (docker-build tag)
    docker run --rm -e PORT=8080 -p 8080:8080 {{ tag }}


# -- Golden vectors from the prototype --------------------------------------

# Export the prototype's golden vectors into testdata/golden/ (see its export/README.md).
# Python comes from the uv image and PostgreSQL from the prototype's compose file, both
# pinned by tag and digest, so neither has to be installed anywhere.
golden:
    #!/usr/bin/env bash
    set -euo pipefail
    uv_image="ghcr.io/astral-sh/uv:0.12.13-python3.12-trixie-slim@sha256:87bc72093c0aa93cc962bd7c0498ddf416dad3ce9e1434724e936e02b72afe5d"
    repo="$(pwd)"

    if [ ! -d prototype/src/taskgen ]; then
        echo "golden: prototype/ is missing; the export runs the prototype's own code." >&2
        exit 1
    fi

    echo "==> PostgreSQL"
    (cd prototype && docker compose up -d --wait)

    # Whatever happens next, the database stops: with set -e a failing export
    # would otherwise leave it running until somebody notices.
    trap 'echo "==> stopping PostgreSQL"; (cd "$repo/prototype" && docker compose down)' EXIT

    echo "==> dependencies, schema, seed profiles, export"
    docker run --rm --network host \
        -v "$repo:/repo" -w /repo/prototype --user "$(id -u):$(id -g)" \
        -e HOME=/tmp -e UV_PROJECT_ENVIRONMENT=/tmp/venv -e UV_CACHE_DIR=/tmp/uvcache \
        -e DATABASE_URL=postgresql://taskgen:taskgen@127.0.0.1:5432/taskgen \
        "$uv_image" \
        bash -lc "uv sync --frozen \
            && uv run python -m taskgen.apply_schema --force \
            && uv run python -m taskgen.seed \
            && uv run python /repo/testdata/golden/export/export_golden.py --out /repo/testdata/golden"

    echo ""
    echo "Golden vectors are in testdata/golden/. A second run must produce byte-identical files."
