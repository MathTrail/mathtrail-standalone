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

# The client the MCP endpoint is explored with while it runs locally: MCP
# Inspector, from the image its project publishes, pinned by tag and digest so
# that every package inside it is the one that was checked, and nothing of it is
# installed here.
INSPECTOR_IMAGE := "ghcr.io/modelcontextprotocol/inspector:2.7.0@sha256:ae22e300c9fc6f4088b490d78fa8079f139642de632cd9ad40f7d9a3d50b8fd8"
# Where a local server started with `just run` serves the endpoint.
LOCAL_MCP := "http://localhost:8080/mcp"
# Where a live run of the lesson is held: a folder outside the repository, so
# that the chat's model finds none of its files — no project instructions, no
# code, no reference task with its answer. The server's log of the run is kept
# there too.
PLAY_DIR := env("MATHTRAIL_PLAY_DIR", home_directory() / "mathtrail-play")
# The model a live run writes its tasks with, by its id rather than by an alias
# that moves on to a newer model, so that runs compare; MATHTRAIL_PLAY_MODEL
# names another.
PLAY_MODEL := env("MATHTRAIL_PLAY_MODEL", "claude-sonnet-5")
# The lesson's tools as a chat's model calls them. The one only a card calls is
# not among them.
PLAY_TOOLS := "mcp__mathtrail__get_profile mcp__mathtrail__save_profile mcp__mathtrail__get_progress mcp__mathtrail__next_task mcp__mathtrail__submit_task mcp__mathtrail__submit_answer"
# The Inspector shares this environment's network, so that it reaches the local
# server at its own address, and it listens on the loopback alone rather than on
# every interface its image asks for. Secrets it would keep — the tokens of a
# sign-in, once there is one — stay in its memory, and it opens no browser,
# since there is none in here.
INSPECTOR_RUN := "docker run --rm --init --network host -e HOST=127.0.0.1 -e MCP_INSPECTOR_SECRET_STORE=memory -e MCP_AUTO_OPEN_ENABLED=false " + INSPECTOR_IMAGE

# The image the runtime image is scanned with, pinned by tag and digest like every
# other image this repository runs.
TRIVY_IMAGE := "aquasec/trivy:0.74.0@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969"

# What a dependency's license may be: permissive, and compatible with releasing
# the result under MIT.
ALLOWED_LICENSES := "MIT,BSD-2-Clause,BSD-3-Clause,Apache-2.0,ISC"
# The npm packages allowed a license outside that list, as name=license. Each is
# a tool that builds the widget and ends up in nothing that is shipped, and it is
# allowed only while it stays a development dependency. lightningcss, under
# MPL-2.0, is the CSS minifier the widget's bundler cannot be installed without.
NPM_LICENSE_EXCEPTIONS := "lightningcss=MPL-2.0"

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

# Build the server binary into bin/, with the widget built into it
build: web-build
    go build -trimpath -ldflags "{{ LDFLAGS }}" -o {{ BINARY }} ./cmd/server

# The version is a parameter, because a release knows the number it is about to
# publish before any tag carries it; without one, the same version the rest of
# the build uses. The widget is built first, so that every binary carries it.
# Build the release artifacts into dist/: the server for every architecture, and the sums of what was built
release-artifacts version=VERSION: (web-build version)
    #!/usr/bin/env bash
    set -euo pipefail
    ldflags="-s -w \
        -X {{ SYMBOLS }}.Version={{ version }} \
        -X {{ SYMBOLS }}.Commit={{ COMMIT }} \
        -X {{ SYMBOLS }}.Date={{ DATE }}"
    rm -rf dist
    mkdir -p dist
    for arch in amd64 arm64; do
        name="mathtrail-server_{{ version }}_linux_${arch}"
        CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build \
            -trimpath -ldflags "$ldflags" -o "dist/$name" ./cmd/server
    done
    cd dist
    sha256sum * > SHA256SUMS
    echo "dist: $(ls | wc -l) files"

# The sealing key is made fresh for the run and kept nowhere: locally there is
# nothing sealed that has to outlive the process. The development sign-in lets
# every request to the MCP endpoint in as one account, which is what a client
# on this machine needs before any real sign-in exists. The widget is built
# first, so that a card drawn from this server is the widget of these sources.
# Run the server from source, with logs a person can read and the development sign-in
run: web-build
    MATHTRAIL_LOG_FORMAT=console MATHTRAIL_LOG_LEVEL=debug MATHTRAIL_DEV_AUTH=true \
        MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" \
        go run ./cmd/server

# The web client opens on port 6274 and shows the whole traffic.
# Explore the MCP endpoint of a server started with `just run` in MCP Inspector
inspect *args:
    {{ INSPECTOR_RUN }} --server-url {{ LOCAL_MCP }} --transport http {{ args }}

# One method at a time, such as `just inspect-cli --method tools/list`, or
# `--method tools/call --tool-name <name> --tool-arg key=value`. Left alone the
# client picks the protocol version itself; `--protocol-era modern` makes it
# speak the newest one and `--protocol-era legacy` an older one.
# Ask the MCP endpoint of a server started with `just run` from the terminal
inspect-cli *args:
    {{ INSPECTOR_RUN }} --cli --server-url {{ LOCAL_MCP }} --transport http {{ args }}

# The server keeps its profiles in memory, so every start is a first sign-in:
# a live run starts it anew for each scenario. Each start writes a log of its
# own, JSON lines named by the time it started, so that the journal of one
# child is read from one file.
# Run the server for a live run of the lesson: the development sign-in, a new child, JSON lines into a log of its own
play-server:
    mkdir -p "{{ PLAY_DIR }}"
    MATHTRAIL_LOG_FORMAT=json MATHTRAIL_DEV_AUTH=true \
        MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" \
        go run ./cmd/server > "{{ PLAY_DIR }}/server-$(date -u +%Y%m%dT%H%M%SZ).jsonl"

# The chat is Claude Code in the run's folder, connected through its .mcp.json
# to the server of `just play-server` and to nothing else. It has none of its
# own tools — no files, no shell, no web — and the lesson's tools need no
# permission; testdata/live/session.md tells it this is a lesson, not a coding
# session. Arguments go to claude as they are: `just play` opens the chat, and
# `just play -p --resume <session>` sends it the message on standard input.
# Hold a lesson with the server of `just play-server`, as a chat in Claude Code that draws no cards
[positional-arguments]
play *args:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p "{{ PLAY_DIR }}"
    printf '{"mcpServers": {"mathtrail": {"type": "http", "url": "%s"}}}\n' "{{ LOCAL_MCP }}" > "{{ PLAY_DIR }}/.mcp.json"
    session="$(cat "{{ justfile_directory() }}/testdata/live/session.md")"
    cd "{{ PLAY_DIR }}"
    exec claude --model "{{ PLAY_MODEL }}" --strict-mcp-config --mcp-config .mcp.json --tools "" \
        --allowedTools "{{ PLAY_TOOLS }}" --append-system-prompt "$session" "$@"

# Show one reference task beside the solver that proves its answer
solver id:
    #!/usr/bin/env bash
    set -euo pipefail
    task=$(jq -r --arg id '{{ id }}' '
        .[] | select(.id == $id)
        | "\(.id)   answer: \(.correct_answer)   \(.topic), grade \(.grade_level), difficulty \(.difficulty)\n",
          .question, "",
          (.options | to_entries | sort_by(.key) | map("  \(.key)  \(.value)") | join("\n"))
    ' content/examples/*.json)
    if [ -z "$task" ]; then
        echo "no reference task with id {{ id }}" >&2
        exit 1
    fi
    printf '%s\n\n' "$task"
    program="content/examples/solvers/{{ id }}.star"
    if [ -f "$program" ]; then cat "$program"; else echo "(no solver yet)"; fi

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
    echo "THIRD_PARTY_LICENSES: $(grep -c 'https://' THIRD_PARTY_LICENSES) entries."

# The license list as the dependency graphs report it right now
_license-list:
    #!/usr/bin/env bash
    set -euo pipefail
    cat <<'HEADER'
    Third-party licenses
    ====================

    MathTrail itself is MIT; see LICENSE. Rewrite this file with: just licenses

    The Go modules below are linked into the programs this repository builds.
    Each line is a license, the module it covers, and the text of that license
    at the exact version in go.mod.

    HEADER
    # The report goes to stdout and the progress of the scan to stderr, which is
    # noise here. The columns are license, module, license text.
    go run {{ GO_LICENSES }} report ./... --ignore {{ MODULE }} 2>/dev/null \
        | awk -F, '{ printf "%-13s %-46s %s\n", $3, $1, $2 }'
    cat <<'WIDGET'

    The npm packages below are what the widget the server embeds is built from:
    every package its lockfile needs outside development. Each line is a
    license, the package at the exact version in web/package-lock.json, and the
    page of that version in the npm registry.

    WIDGET
    node web/scripts/licenses.ts list

# -- Widget -----------------------------------------------------------------

# Nothing is installed while the installed packages are newer than the lockfile,
# so that a build does not pay for an install it does not need.
# Install the widget's packages as its lockfile pins them, when they are missing or stale
[working-directory('web')]
web-install:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! node_modules/.package-lock.json -nt package-lock.json ]; then
        npm ci --no-audit --no-fund
    fi

# The widget tells a host its name and the version of the build it came with.
# The page it leaves is checked at once, and an old one is removed first, so
# that a build which wrote nothing cannot pass on the last one: the page must
# name the handshake a host starts a widget with, which the placeholder beside
# it does not.
# Build the widget into the page the server embeds
[working-directory('web')]
web-build version=VERSION: web-install
    #!/usr/bin/env bash
    set -euo pipefail
    page=../internal/widget/widget.html
    rm -f "$page"
    VITE_VERSION="{{ version }}" npm run --silent build:widget
    if ! grep -q "ui/initialize" "$page"; then
        echo "web-build: $page is not a widget a host can start" >&2
        exit 1
    fi

# Format the widget's sources in place
[working-directory('web')]
web-fmt: web-install
    npm run --silent fmt

# Fail if a widget source needs formatting or breaks a lint rule
[working-directory('web')]
web-lint: web-install
    npm run --silent check

# Run the widget's tests
[working-directory('web')]
web-test: web-install
    npm run --silent test

# -- The full checks --------------------------------------------------------

# Formatting and lint
ci-lint: fmt-check
    golangci-lint run ./...

# Tests with the race detector and a coverage report
ci-test:
    go test ./... -race -count=1 -coverprofile=coverage.out -covermode=atomic
    go tool cover -func=coverage.out | tail -1

# The widget's formatting and lint, its types, its tests with their coverage, and a build
[working-directory('web')]
ci-web: web-build
    npm run --silent check
    npm run --silent typecheck
    npm run --silent test -- --coverage

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

# govulncheck reports only what the code can reach. npm has no such analysis, so
# every package the widget is built from is held to the advisories, at high and
# above; the tools that only build it are not, since none of them is shipped.
# Fail on a known vulnerability in code the service reaches or a package the widget is built from
ci-vuln:
    go run {{ GOVULNCHECK }} ./...
    npm audit --prefix web --omit=dev --audit-level=high

# Fail on a vulnerability in the image the service is shipped in. Trivy runs from
# its own image, so nothing about it is installed here, and the same command gives
# the same verdict on a laptop and in the checks.
ci-image-scan tag="mathtrail:dev": (docker-build tag)
    docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        {{ TRIVY_IMAGE }} image \
        --severity HIGH,CRITICAL \
        --ignore-unfixed \
        --exit-code 1 \
        --no-progress \
        {{ tag }}

# The image is started the way a developer's machine starts it — with the
# development sign-in, which a deployment refuses, and a key made for the run —
# and asked for the widget the way a host asks. The page arrives inside a JSON
# answer that escapes its markup, so it is known by bare words: the placeholder
# names itself, and a built widget names the handshake it opens.
# Fail if the image serves the placeholder, or anything else, in place of the widget
ci-image-widget tag="mathtrail:dev": (docker-build tag)
    #!/usr/bin/env bash
    set -euo pipefail
    # Removed when the check ends rather than when the server does, so that a
    # server which stopped at once still has its logs to show.
    container=$(docker run -d -p 127.0.0.1::8080 -e PORT=8080 \
        -e MATHTRAIL_DEV_AUTH=true \
        -e MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" \
        {{ tag }})
    trap 'docker rm -f "$container" > /dev/null' EXIT
    address=""
    for attempt in $(seq 1 20); do
        # A container that has stopped publishes no port.
        address=$(docker port "$container" 8080/tcp 2> /dev/null | head -1) || true
        if [ -n "$address" ] && curl -fsS --max-time 2 "http://${address}/health" > /dev/null 2>&1; then
            break
        fi
        address=""
        echo "health: no answer yet (attempt ${attempt})" >&2
        sleep 0.5
    done
    if [ -z "$address" ]; then
        echo "ci-image-widget: {{ tag }} never answered its health check" >&2
        docker logs "$container" >&2
        exit 1
    fi
    # The container knows itself as localhost:8080, the one host it serves.
    if ! answer=$(curl -fsS --max-time 10 -X POST "http://${address}/mcp" \
        -H 'Host: localhost:8080' \
        -H 'Content-Type: application/json' \
        -H 'Accept: application/json, text/event-stream' \
        -d '{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"ui://mathtrail/app.html"}}'); then
        echo "ci-image-widget: {{ tag }} did not answer a read of the widget" >&2
        docker logs "$container" >&2
        exit 1
    fi
    if [[ "$answer" == *widget-placeholder* ]]; then
        echo "ci-image-widget: {{ tag }} serves the placeholder: no widget was built into it" >&2
        exit 1
    fi
    if [[ "$answer" != *ui/initialize* ]]; then
        echo "ci-image-widget: {{ tag }} serves a page that is not the widget" >&2
        exit 1
    fi
    echo "ci-image-widget: {{ tag }} serves the widget"

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
    node web/scripts/licenses.ts check --allowed {{ ALLOWED_LICENSES }} --except {{ NPM_LICENSE_EXCEPTIONS }}
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

# Raise the pinned version of the Claude Code CLI and its editor extension. The
# version stays exact — this only removes the part where a person edits the same
# number in two files and misses one. Without an argument, whatever npm calls
# latest today; the change is printed and committed like any other.
claude-update version="":
    #!/usr/bin/env bash
    set -euo pipefail

    dockerfile=".devcontainer/Dockerfile"
    devcontainer=".devcontainer/devcontainer.json"

    wanted="{{ version }}"
    if [ -z "$wanted" ]; then
        wanted=$(curl -fsSL --proto "=https" --proto-redir "=https" \
            https://registry.npmjs.org/@anthropic-ai/claude-code/latest \
        | sed -n 's/.*"version":"\([^"]*\)".*/\1/p')
    fi
    if [[ ! "$wanted" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "claude-update: $wanted is not a version" >&2
        exit 1
    fi

    current=$(sed -n 's/^ARG CLAUDE_CODE_VERSION=//p' "$dockerfile")
    if [ "$current" = "$wanted" ]; then
        echo "claude code $current: already pinned"
        exit 0
    fi

    # The CLI and the extension are one version in two files, and a rebuild with
    # them apart is the failure this recipe exists to prevent.
    sed -i "s/^ARG CLAUDE_CODE_VERSION=.*/ARG CLAUDE_CODE_VERSION=${wanted}/" "$dockerfile"
    sed -i "s/\"anthropic.claude-code@[^\"]*\"/\"anthropic.claude-code@${wanted}\"/" "$devcontainer"

    echo "claude code $current -> $wanted"
    git --no-pager diff -- "$dockerfile" "$devcontainer"
    echo
    echo "Rebuild the container for this to take effect: Dev Containers: Rebuild Container."

# Build the runtime image
docker-build tag="mathtrail:dev":
    docker build \
        --build-arg VERSION="{{ VERSION }}" \
        --build-arg COMMIT="{{ COMMIT }}" \
        --build-arg DATE="{{ DATE }}" \
        -t {{ tag }} .

# Build the image and run it on port 8080, with a sealing key made for this run
docker-run tag="mathtrail:dev": (docker-build tag)
    docker run --rm -e PORT=8080 \
        -e MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" \
        -p 8080:8080 {{ tag }}

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

    just ci-image-widget "$tag" >&2
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
        # A redirect may not downgrade the connection. The address itself is the
        # caller's to choose: this also probes a container on localhost.
        body=$(curl -fsS --max-time 10 --proto-redir "=https" "{{ url }}/health") && break
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

    # The MCP endpoint is there, and it lets nobody in without signing in. Its
    # refusal names the document a client begins a sign-in from.
    headers=$(curl -sS --max-time 10 --proto-redir "=https" -o /dev/null -D - \
        -X POST -H 'Content-Type: application/json' -d '{}' "{{ url }}/mcp")
    status=$(printf '%s\n' "$headers" | awk 'NR == 1 { print $2 }')
    if [ "$status" != "401" ]; then
        echo "smoke: {{ url }}/mcp answered $status to a request nobody signed in, want 401" >&2
        exit 1
    fi
    metadata="{{ url }}/.well-known/oauth-protected-resource/mcp"
    if ! printf '%s\n' "$headers" | grep -qiF "resource_metadata=\"$metadata\""; then
        echo "smoke: the refusal of {{ url }}/mcp does not name $metadata" >&2
        exit 1
    fi

    # That document is served, and it is the endpoint's.
    document=$(curl -fsS --max-time 10 --proto-redir "=https" "$metadata")
    if [[ "$document" != *"\"resource\":\"{{ url }}/mcp\""* ]]; then
        echo "smoke: $metadata does not describe {{ url }}/mcp: $document" >&2
        exit 1
    fi

    echo "smoke: {{ url }} answers as {{ COMMIT }}, and its MCP endpoint wants a sign-in and says where it begins"


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

# -- Research ----------------------------------------------------------------

# The recipes live in research/justfile. The arguments reach it exactly as
# given, without passing through the shell a second time.
[doc('A recipe of the research module: `just research` lists them')]
[positional-arguments]
research *ARGS:
    @{{ quote(just_executable()) }} --justfile research/justfile --working-directory research "$@"
