# MathTrail: task runner.
# The ci-* recipes are the full checks, in the names the platform's other services
# use, so that the same command means the same thing in every repository.

set shell := ["bash", "-euo", "pipefail", "-c"]

BINARY := "bin/server"
MODULE := "github.com/MathTrail/mathtrail-standalone"

# The image the full checks run inside. Lowercase throughout: a registry rejects a
# repository name that is not.
TOOLCHAIN_IMAGE := "ghcr.io/mathtrail/mathtrail-standalone/toolchain"

# Exact versions of the tools that only the full checks need. Each is a Go
# program run straight from its module, so none of them is installed anywhere.
GOVULNCHECK := "golang.org/x/vuln/cmd/govulncheck@v1.8.0"
GITLEAKS := "github.com/zricethezav/gitleaks/v8@v8.30.1"
GO_LICENSES := "github.com/google/go-licenses/v2@v2.0.1"

# What a dependency's license may be: permissive, and compatible with releasing
# the result under MIT.
ALLOWED_LICENSES := "MIT,BSD-2-Clause,BSD-3-Clause,Apache-2.0,ISC"

# The origin the site is published on. Every absolute address on the site, and
# the CNAME that claims the domain, are built from this one value.
SITE_BASE := "https://mathtrail.app"
SITE_DIR := "site/dist"

# Where the cloud this service runs in is described.
TF_DIR := "infra/terraform"

# The build identity, stamped into the binary at link time. Computed once per
# run of just, so that a binary and the image built beside it carry the same
# words.
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo dev`
COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo unknown`
DATE := `date -u +%Y-%m-%dT%H:%M:%SZ`
SYMBOLS := MODULE + "/internal/version"
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

# Rewrite THIRD_PARTY_LICENSES from the dependency graph
licenses:
    #!/usr/bin/env bash
    set -euo pipefail
    list=$(just _license-list)
    # The file is replaced only once the whole list is in hand: a report that
    # died halfway would otherwise leave the repository claiming fewer
    # dependencies than it ships.
    printf '%s\n' "$list" > THIRD_PARTY_LICENSES
    echo "THIRD_PARTY_LICENSES: $(grep -c 'https://' THIRD_PARTY_LICENSES) modules."

# The license list as the dependency graph reports it right now
_license-list:
    #!/usr/bin/env bash
    set -euo pipefail
    cat <<'HEADER'
    Third-party licenses
    ====================

    The Go modules below are linked into the programs this repository builds.
    Each line is a license, the module it covers, and the text of that license
    at the exact version in go.mod. MathTrail itself is MIT; see LICENSE.

    Rewrite this file with: just licenses

    HEADER
    # The report goes to stdout and the progress of the scan to stderr, which is
    # noise here. The columns are license, module, license text.
    go run {{ GO_LICENSES }} report ./... --ignore {{ MODULE }} 2>/dev/null \
        | awk -F, '{ printf "%-13s %-46s %s\n", $3, $1, $2 }'

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

# Fail on a known vulnerability in code the service actually reaches
ci-vuln:
    go run {{ GOVULNCHECK }} ./...

# Fail if anything in the history of the repository looks like a secret
ci-secrets:
    #!/usr/bin/env bash
    set -euo pipefail
    # A history git declines to read scans as zero commits and reports itself clean,
    # which is the one outcome worse than a failure. Refuse to start instead.
    git rev-parse --verify HEAD > /dev/null
    go run {{ GITLEAKS }} git --no-banner --redact .

# Fail if a dependency carries a license we may not ship, or the list is stale
ci-licenses:
    #!/usr/bin/env bash
    set -euo pipefail
    go run {{ GO_LICENSES }} check ./... --allowed_licenses={{ ALLOWED_LICENSES }}
    # The list is built first and compared second: a report that failed to run
    # would otherwise look exactly like a list somebody forgot to update.
    list=$(just _license-list)
    if ! printf '%s\n' "$list" | diff -u THIRD_PARTY_LICENSES - ; then
        echo "THIRD_PARTY_LICENSES is out of date: run just licenses and commit the result." >&2
        exit 1
    fi

# -- Site -------------------------------------------------------------------

# Render the site into site/dist/
site:
    go run ./cmd/sitegen -base {{ SITE_BASE }} -out {{ SITE_DIR }}

# Render the site and serve it, so a page can be read the way a visitor reads it
site-serve port="8081":
    go run ./cmd/sitegen -base {{ SITE_BASE }} -out {{ SITE_DIR }} -serve :{{ port }}

# Render the site and refuse it if anything about it is wrong
ci-site: site
    go run ./cmd/sitecheck -base {{ SITE_BASE }} -dir {{ SITE_DIR }}

# -- Infrastructure ---------------------------------------------------------

# Format the Terraform sources in place
tf-fmt:
    terraform -chdir={{ TF_DIR }} fmt -recursive

# Refuse Terraform that is misformatted or does not describe a valid configuration
tf-check:
    #!/usr/bin/env bash
    set -euo pipefail
    terraform -chdir={{ TF_DIR }} fmt -check -recursive
    # Without a backend and without credentials: this reads the configuration,
    # it does not go near a deployment.
    terraform -chdir={{ TF_DIR }} init -backend=false -input=false
    terraform -chdir={{ TF_DIR }} validate

# Create the project, the bucket of its state and the identity the pipeline uses
bootstrap:
    bash infra/bootstrap.sh

# Show what an apply would change, without changing anything
ci-tf-plan:
    #!/usr/bin/env bash
    set -euo pipefail
    terraform -chdir={{ TF_DIR }} init -backend-config=backend.hcl -input=false
    terraform -chdir={{ TF_DIR }} plan -input=false -no-color

# Bring the cloud up to what this repository says it should be
ci-tf-apply:
    #!/usr/bin/env bash
    set -euo pipefail

    tf() { terraform -chdir={{ TF_DIR }} "$@"; }

    tf init -backend-config=backend.hcl -input=false

    # The secrets come first, on their own. A revision cannot start until they
    # hold something, and what they hold is not described in this repository.
    tf apply -auto-approve -input=false \
        -target=google_secret_manager_secret.seal_key \
        -target=google_secret_manager_secret.google_client_secret

    # The project is read from the file that names it rather than from an
    # output: an apply that was given targets refreshes only the outputs those
    # targets feed, and this one is fed by a variable.
    project=$(sed -n 's/^[[:space:]]*project_id[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' {{ TF_DIR }}/prod.auto.tfvars | head -1)
    if [ -z "$project" ]; then
        echo "ci-tf-apply: {{ TF_DIR }}/prod.auto.tfvars names no project_id." >&2
        exit 1
    fi

    seal_key=$(tf output -raw secret_seal_key)
    client_secret=$(tf output -raw secret_google_client)

    has_version() {
        [ -n "$(gcloud secrets versions list "$1" --project="$project" --limit=1 --format='value(name)' 2>/dev/null)" ]
    }

    # The sealing key is made here and read by nobody: not a person, not the
    # state, not a log. A key that never existed outside Secret Manager is one
    # nobody has to be trusted with.
    if has_version "$seal_key"; then
        echo "seal key: a version exists, leaving it alone"
    else
        echo "seal key: generating the first version"
        head -c 32 /dev/urandom | base64 | tr -d '\n' \
            | gcloud secrets versions add "$seal_key" --project="$project" --data-file=- > /dev/null
    fi

    # This one cannot be generated: it exists only in the Google console, and it
    # is passed in through the environment for exactly this step.
    if has_version "$client_secret"; then
        echo "client secret: a version exists, leaving it alone"
    elif [ -n "${GOOGLE_OAUTH_CLIENT_SECRET:-}" ]; then
        echo "client secret: adding the first version"
        printf '%s' "$GOOGLE_OAUTH_CLIENT_SECRET" \
            | gcloud secrets versions add "$client_secret" --project="$project" --data-file=- > /dev/null
    else
        echo "ci-tf-apply: $client_secret holds no version and GOOGLE_OAUTH_CLIENT_SECRET is not set." >&2
        echo "The service cannot start without it: take it from the Google OAuth client." >&2
        exit 1
    fi

    tf apply -auto-approve -input=false
    tf output

# -- Container --------------------------------------------------------------

# Publish the image the full checks run inside, and print its exact reference
ci-toolchain-image:
    #!/usr/bin/env bash
    set -euo pipefail

    # The tag is a function of what the image is built from: one definition of
    # the environment means one image, and any edit to it means a new one.
    reference="{{ TOOLCHAIN_IMAGE }}:$(sha256sum .devcontainer/Dockerfile | cut -c1-16)"

    if ! docker buildx imagetools inspect "$reference" > /dev/null 2>&1; then
        echo "This environment is not published yet; building it." >&2
        docker build --target toolchain --file .devcontainer/Dockerfile --tag "$reference" . >&2
        docker push "$reference" >&2
    fi

    # The digest, not the tag: a tag can be moved, and a version here is exact.
    # Read with awk rather than a Go template, whose braces would collide with
    # the interpolation syntax of this file.
    digest=$(docker buildx imagetools inspect "$reference" | awk '/^Digest:/ { print $2 }')

    # The one thing on stdout, so that a caller can read it with a substitution.
    echo "{{ TOOLCHAIN_IMAGE }}@${digest}"

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

# -- Deployment -------------------------------------------------------------

# Build the runtime image, push it, and print the exact reference it got
ci-image-push image:
    #!/usr/bin/env bash
    set -euo pipefail

    image="{{ image }}"

    # The tag says which commit an image was built from, so that the registry
    # can be read by a person. What is deployed is the digest below: a tag can
    # be moved afterwards and a digest cannot.
    tag="${image}:{{ COMMIT }}"

    # A push needs a credential helper for that registry, and the registry is
    # the first component of the image path.
    gcloud auth configure-docker "${image%%/*}" --quiet >&2

    just docker-build "$tag" >&2
    docker push "$tag" >&2

    # Read with awk rather than a Go template, whose braces would collide with
    # the interpolation syntax of this file.
    digest=$(docker buildx imagetools inspect "$tag" | awk '/^Digest:/ { print $2 }')

    # The one thing on stdout, so that a caller can read it with a substitution.
    echo "${image}@${digest}"

# Roll a new revision of the service, and wait until it is the one serving
ci-deploy service region image:
    #!/usr/bin/env bash
    set -euo pipefail

    image="{{ image }}"
    if [[ "$image" != *@sha256:* ]]; then
        echo "ci-deploy: an image is deployed by digest, never by tag: $image" >&2
        exit 1
    fi

    # The service is created once, with everything around it; a deployment only
    # ever changes which image it serves.
    gcloud run services update "{{ service }}" \
        --region="{{ region }}" \
        --image="$image" \
        --quiet

# Refuse a deployment that does not answer, or answers as another build
ci-smoke url:
    #!/usr/bin/env bash
    set -euo pipefail

    body=""
    for attempt in $(seq 1 10); do
        body=$(curl -fsS --max-time 10 "{{ url }}/health") && break
        echo "health: no answer yet (attempt ${attempt})" >&2
        sleep 3
    done

    if [ -z "$body" ]; then
        echo "smoke: {{ url }}/health never answered" >&2
        exit 1
    fi

    # Answering is half of it. The commit in the answer is what says the
    # revision now serving is the one just built, rather than the one before it.
    if [[ "$body" != *'"commit":"{{ COMMIT }}"'* ]]; then
        echo "smoke: {{ url }} is serving another build: $body" >&2
        exit 1
    fi

    echo "smoke: {{ url }} answers as {{ COMMIT }}"


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

# Print the deployment target, one fact per line, for a caller to read
ci-tf-outputs:
    #!/usr/bin/env bash
    set -euo pipefail
    tf() { terraform -chdir={{ TF_DIR }} "$@"; }
    echo "deploy_service_account=$(tf output -raw deploy_service_account)"
    echo "image_repository=$(tf output -raw image_repository)"
    echo "service=$(tf output -raw service_name)"
    echo "region=$(tf output -raw region)"
    echo "public_url=$(tf output -raw public_url)"
