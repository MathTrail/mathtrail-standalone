# MathTrail: task runner.
# The ci-* recipes are the full checks, in the names the platform's other services
# use, so that the same command means the same thing in every repository.

set shell := ["bash", "-euo", "pipefail", "-c"]

BINARY := "bin/server"
MODULE := "github.com/MathTrail/mathtrail-standalone"

# The image the full checks run inside. Lowercase throughout: a registry rejects a
# repository name that is not.
TOOLCHAIN_IMAGE := "ghcr.io/mathtrail/mathtrail-standalone/toolchain"

# Where the months of DB-IP's IP to Country Lite the runtime image carries are
# kept, each an image of its own holding the file as DB-IP published it. DB-IP
# keeps a month to download for about three months; a copy of our own keeps
# every month an image was ever built with.
COUNTRIES_IMAGE := "ghcr.io/mathtrail/mathtrail-standalone/dbip-country-lite"

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
# The lesson's tools as a chat's model calls them. The ones only a card calls
# are not among them.
PLAY_TOOLS := "mcp__mathtrail__get_profile mcp__mathtrail__save_profile mcp__mathtrail__get_progress mcp__mathtrail__next_task mcp__mathtrail__get_package mcp__mathtrail__submit_task mcp__mathtrail__submit_answer"
# The Inspector shares this environment's network, so that it reaches the local
# server at its own address, and it listens on the loopback alone rather than on
# every interface its image asks for. Secrets it would keep — the tokens of a
# sign-in, once there is one — stay in its memory, and it opens no browser,
# since there is none in here.
INSPECTOR_RUN := "docker run --rm --init --network host -e HOST=127.0.0.1 -e MCP_INSPECTOR_SECRET_STORE=memory -e MCP_AUTO_OPEN_ENABLED=false " + INSPECTOR_IMAGE

# The image the runtime image is scanned with, pinned by tag and digest like every
# other image this repository runs.
TRIVY_IMAGE := "aquasec/trivy:0.74.0@sha256:62b1e65e8869bc4b4c6aa4fa2b21595256c7c2f6018a9d9ad61caf87187c1969"

# The image the widget's layout is measured in: Chromium and WebKit as the pinned
# Playwright release builds them, with the libraries and the fonts of every
# script they need, pinned by tag and digest. Its release is the one
# web/package.json pins playwright-core to, since a release drives only the
# browsers of its own build.
PLAYWRIGHT_IMAGE := "mcr.microsoft.com/playwright:v1.63.0-noble@sha256:eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27"

# The BigQuery emulator the SQL of the counts kept for years is tested against,
# pinned by tag and digest. It is no BigQuery: SQLite runs the queries under it.
BIGQUERY_EMULATOR_IMAGE := "ghcr.io/goccy/bigquery-emulator:0.8.1@sha256:f4e428d265a93dc5ce36c294e1c584c7c9b384117d47ab8ddbb63d8d50b7f393"

# The prototype the golden vectors are the numbers of: its public repository, at the
# commit the copy they were first exported from matched byte for byte.
PROTOTYPE_REPOSITORY := "https://github.com/MathTrail/llm-taskgen-prototype"
PROTOTYPE_COMMIT := "02638353482e25d6213120ba10caea3467a0437a"

# What a dependency's license may be: permissive, and compatible with releasing
# the result under MIT.
ALLOWED_LICENSES := "MIT,BSD-2-Clause,BSD-3-Clause,Apache-2.0,ISC"
# The licenses a font may carry besides: the site ships a font's files, and the
# Open Font License lets them travel with software under any license, so long
# as its notice and text travel with them. The npm check alone takes them,
# since the fonts come from npm, and it takes any package under one of them,
# font or not.
FONT_LICENSES := "OFL-1.1"
# The npm packages allowed a license outside that list, as name=license. Each is
# a tool that builds the widget and ends up in nothing that is shipped, and it is
# allowed only while it stays a development dependency. lightningcss, under
# MPL-2.0, is the CSS minifier the widget's bundler cannot be installed without.
NPM_LICENSE_EXCEPTIONS := "lightningcss=MPL-2.0"

# The Go template that prints the module a package of a tool's build comes
# from, and its version, when that module is neither the tool nor the service it
# is built against. It sits in a variable so that its braces never meet the
# interpolation of this file.
TOOL_MODULE_OF := '{{with .Module}}{{if not .Main}}{{if not .Replace}}{{.Path}} {{.Version}}{{end}}{{end}}{{end}}'

# The Go template that prints how the Docker daemon keeps its containers'
# cgroups, and on what kernel: they decide whether the load tool can measure
# an instance at all. It sits in a variable for the same reason.
DOCKER_CGROUPS := '{{.CgroupVersion}} {{.CgroupDriver}}, kernel {{.KernelVersion}}'

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

# A version asked for is the one published. Otherwise it is the next patch of
# the line the releases are numbered in, MAJOR.MINOR — the line's first when
# none of it is out yet. A line older than the newest release would publish a
# version below one already out, and is refused; so is a version already
# released, or one that is not vMAJOR.MINOR.PATCH. The arguments reach the
# script as arguments, never as its text, since a version asked for is typed by
# a person.
# Print the version the next release is published as, such as `just release-version 0.2`
[positional-arguments]
release-version line requested="":
    #!/usr/bin/env bash
    set -euo pipefail
    line="$1"
    requested="${2:-}"
    if [[ ! "$line" =~ ^[0-9]+\.[0-9]+$ ]]; then
        echo "release: the line $line is not MAJOR.MINOR" >&2
        exit 1
    fi
    newest="$(git tag --list 'v[0-9]*' --sort=-v:refname | head -1)"
    if [ -n "$requested" ]; then
        version="$requested"
    else
        last="$(git tag --list "v$line.[0-9]*" --sort=-v:refname | head -1)"
        if [ -z "$last" ]; then
            version="v$line.0"
        else
            version="v$line.$((${last##*.} + 1))"
        fi
        if [ -n "$newest" ] &&
            [ "$(printf '%s\n' "${newest#v}" "${version#v}" | sort -V | tail -1)" != "${version#v}" ]; then
            echo "release: the line $line is older than the newest release, $newest" >&2
            exit 1
        fi
    fi
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "release: $version is not vMAJOR.MINOR.PATCH" >&2
        exit 1
    fi
    if git rev-parse -q --verify "refs/tags/$version" > /dev/null; then
        echo "release: $version is already released" >&2
        exit 1
    fi
    echo "$version"

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
# every request to the MCP endpoint in — as the account a bearer name names, or
# as one development account when it names none — so that a client on this
# machine needs no Google account, and keeps each profile in the memory of the
# process, since it has no Drive to keep it in. The widget is built first, so
# that a card drawn from this server is the widget of these sources.
# Run the server from source, with logs a person can read and the development sign-in
run: web-build
    MATHTRAIL_LOG_FORMAT=console MATHTRAIL_LOG_LEVEL=debug MATHTRAIL_DEV_AUTH=true \
        MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" \
        go run ./cmd/server

# The same server with the real sign-in: a client is sent through the consent
# screen and Google, and the MCP endpoint lets in only the tokens it issued.
# The Google client is the one MATHTRAIL_GOOGLE_CLIENT_ID and
# MATHTRAIL_GOOGLE_CLIENT_SECRET name, and Google has to know
# http://localhost:8080/oauth/callback as one of its redirect URIs. Tokens die
# with the process, as the key they are sealed with does. The profile does not:
# as on a deployment, it is a file in the Drive of whoever signs in.
# Run the server from source with the Google sign-in, for a client that signs in for real
run-signin: web-build
    : "${MATHTRAIL_GOOGLE_CLIENT_ID:?set it and MATHTRAIL_GOOGLE_CLIENT_SECRET to a Google client that knows http://localhost:8080/oauth/callback}"
    MATHTRAIL_LOG_FORMAT=console MATHTRAIL_LOG_LEVEL=debug \
        MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" \
        go run ./cmd/server

# The web client opens on port 6274 and shows the whole traffic.
# Explore the MCP endpoint of a server started with `just run` in MCP Inspector
inspect *args:
    {{ INSPECTOR_RUN }} --server-url {{ LOCAL_MCP }} --transport http {{ args }}

# One method at a time, such as `just inspect-cli --method tools/list`, or
# `--method tools/call --tool-name <name> --tool-arg key=value`. Left alone the
# client picks the protocol version itself; `--protocol-era modern` makes it
# speak the newest one and `--protocol-era legacy` an older one. Against a
# server with the real sign-in, the client signs in first: it prints the
# address to open in a browser, and waits for the browser to come back to
# 127.0.0.1:6276, which that browser has to reach. It keeps the tokens in its
# memory, so every call signs in anew.
# Ask the MCP endpoint of a server started with `just run` or `just run-signin` from the terminal
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

# The drawings the limits of a text drawing are chosen by, each on the widget's
# own task card: a server with no sign-in behind a quick tunnel, whose address
# the tunnel prints. Add it to a chat host as a connector with no sign-in and
# ask for the drawings by number; the five options of a card rate its drawing,
# and every rating is a line of the log, JSON lines named by the time the
# server started. The tunnel hands the server its requests as addressed to the
# server itself: the protocol library refuses a request that reaches a server
# on this machine addressed to any other host.
# Show a chat host the calibration set of text drawings behind a quick tunnel, and log the ratings
drawings port="8090": web-build
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p "{{ PLAY_DIR }}"
    go build -o "{{ PLAY_DIR }}/drawings" ./cmd/drawings
    log="{{ PLAY_DIR }}/drawings-$(date -u +%Y%m%dT%H%M%SZ).jsonl"
    "{{ PLAY_DIR }}/drawings" -addr "127.0.0.1:{{ port }}" > "$log" &
    server=$!
    trap 'kill "$server" 2> /dev/null' EXIT
    # A tunnel to a server that never started would print an address that
    # answers nothing.
    until curl -fsS "http://127.0.0.1:{{ port }}/healthz" > /dev/null 2>&1; do
        if ! kill -0 "$server" 2> /dev/null; then
            echo "drawings: the server did not start; the end of its log:" >&2
            tail -n 5 "$log" >&2
            exit 1
        fi
        sleep 0.5
    done
    cloudflared tunnel --no-autoupdate --url "http://127.0.0.1:{{ port }}" \
        --http-host-header "127.0.0.1:{{ port }}"

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

    The Go modules below are linked into the programs this repository builds,
    and so are the standard library and the runtime of the Go release that
    builds them, listed as std. Each line is a license, what it covers, and the
    text of that license at the exact version that is built.

    HEADER
    # The release go.mod names is the one the releases are built with, and it
    # names its patch too, as every version here does: every Go release since
    # 1.21 is tagged with its patch, so a link without one would lead nowhere.
    release=$(awk '$1 == "go" { print $2; exit }' go.mod)
    if [[ ! "$release" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "go.mod names Go ${release}; the license list needs the exact release" >&2
        exit 1
    fi
    # The report goes to stdout and the progress of the scan to stderr, which is
    # noise here. The columns are module, license text, license. No module graph
    # lists the standard library and the runtime, which every binary carries,
    # so they are put at the head of the report as one more module.
    {
        echo "std,https://github.com/golang/go/blob/go${release}/LICENSE,BSD-3-Clause"
        go run {{ GO_LICENSES }} report ./... --ignore {{ MODULE }} 2>/dev/null
    } | awk -F, '{ printf "%-13s %-46s %s\n", $3, $1, $2 }'
    cat <<'WEB'

    The npm packages below are what the widget the server embeds and the site
    are built from: every package their lockfile needs outside development, and
    every package one of those names as an optional companion, used or not.
    Each line is a license, the package at the exact version in
    web/package-lock.json, and the page of that version in the npm registry.

    WEB
    node web/scripts/licenses.ts list
    cat <<'DATA'

    The data below ships in the image the server runs from, unchanged. Each
    line is a license, the data at the exact version the image carries, and
    where its maker publishes it. IP geolocation by DB-IP (https://db-ip.com),
    under the Creative Commons Attribution 4.0 International License.

    DATA
    # The month of the database is the tag of the image the runtime image
    # copies it from.
    month=$(sed -n 's|^FROM {{ COUNTRIES_IMAGE }}:\([0-9-]*\)@.* AS countries$|\1|p' Dockerfile)
    if [[ ! "$month" =~ ^[0-9]{4}-[0-9]{2}$ ]]; then
        echo "the Dockerfile copies the database of countries from no month of {{ COUNTRIES_IMAGE }}" >&2
        exit 1
    fi
    printf '%-13s %-46s %s\n' "CC-BY-4.0" "DB-IP IP to Country Lite ${month}" "https://db-ip.com/db/download/ip-to-country-lite"

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

# The page plays a chat host: it frames the widget's own page, as it is being
# worked on, and drives it through every state a lesson puts a card in, at a
# phone's width and a web page's, in the light and the dark theme. The cards
# name the version a build of this tree would.
# Serve the widget in every state of a lesson, to look at beside the design
[working-directory('web')]
web-preview: web-install
    VITE_VERSION="{{ VERSION }}" npm run --silent preview

# Every state of a lesson in the preview, in every language the widget speaks
# and the pseudo-language, at 320, 360 and 640 px, in Chromium and WebKit: it
# fails on a page that scrolls sideways, a part that sticks out of its card,
# and text that runs out of its box or is cut short. The browsers run in their
# image, as this user, over the repository as it is; what they find, and a
# picture of every card written right to left, are left in web/layout/.
# Arguments narrow the run: --engine chromium --language ar --width 320; and
# --shard 2/6 measures the second of six parts of it, for machines that share
# one run between them.
# Measure the widget's layout in real browsers, in every language, at every width a card must fit
[working-directory('web')]
web-layout *args: _playwright-pinned
    #!/usr/bin/env bash
    set -euo pipefail
    docker run --rm --init --shm-size=1g --user "$(id -u):$(id -g)" -e HOME=/tmp \
        -v "{{ justfile_directory() }}:{{ justfile_directory() }}" -w "{{ justfile_directory() }}/web" \
        {{ PLAYWRIGHT_IMAGE }} node scripts/layout.ts {{ args }}

# The README's pictures of a task, a wrong answer and the progress, each in the
# light and the dark theme: the preview's own scenes, photographed in Chromium
# from the image the layout is measured in, as a phone 428 px wide shows them
# at twice its density. They are written over docs/screens/, for the change to
# be looked at before it is kept. The cards name the newest release this tree
# follows, rather than the changes on top of it nobody has released; a clone
# that holds no release is refused, rather than photographed as "dev".
# Photograph the widget for the README
[working-directory('web')]
web-screens: _playwright-pinned
    #!/usr/bin/env bash
    set -euo pipefail
    release=$(git describe --tags --abbrev=0)
    docker run --rm --init --shm-size=1g --user "$(id -u):$(id -g)" -e HOME=/tmp \
        -e VITE_VERSION="$release" \
        -v "{{ justfile_directory() }}:{{ justfile_directory() }}" -w "{{ justfile_directory() }}/web" \
        {{ PLAYWRIGHT_IMAGE }} node scripts/screens.ts

# playwright-core drives only the browsers of its own release: a package moved
# without the image, or the image without the package, would find no browser to
# launch, and say so in words that name neither pin.
# Fail unless the widget's playwright-core and the browsers' image are one release
[working-directory('web')]
_playwright-pinned: web-install
    #!/usr/bin/env bash
    set -euo pipefail
    release=$(node -p 'require("./package.json").devDependencies["playwright-core"]')
    if [[ "{{ PLAYWRIGHT_IMAGE }}" != *":v${release}-"* ]]; then
        echo "playwright: web/package.json pins playwright-core ${release}, and the image is {{ PLAYWRIGHT_IMAGE }}; move them together" >&2
        exit 1
    fi

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
    node web/scripts/licenses.ts check --allowed {{ ALLOWED_LICENSES }},{{ FONT_LICENSES }} --except {{ NPM_LICENSE_EXCEPTIONS }}
    # The list is built first and compared second: a report that failed to run
    # would otherwise look exactly like a list somebody forgot to update.
    list=$(just _license-list)
    if ! printf '%s\n' "$list" | diff -u THIRD_PARTY_LICENSES - ; then
        echo "THIRD_PARTY_LICENSES is out of date: run just licenses and commit the result." >&2
        exit 1
    fi

# -- Site -------------------------------------------------------------------

# The site is drawn by the widget's own toolchain, from the texts in
# site/content/, and judged by a checker that knows nothing of how it was drawn.
# Build the site into site/dist/
[working-directory('web')]
site: web-install
    npm run --silent build:site -- --base {{ SITE_BASE }} --out ../{{ SITE_DIR }}

# Build the site and serve it, so a page can be read the way a visitor reads it
[working-directory('web')]
site-serve port="8081": site
    npm run --silent preview:site -- --outDir ../{{ SITE_DIR }} --host 0.0.0.0 --port {{ port }} --strictPort

# Build the site and refuse it if anything about it is wrong
[working-directory('web')]
ci-site: site
    npm run --silent check:site -- --base {{ SITE_BASE }} --dir ../{{ SITE_DIR }}

# The pictures a shared link to the site shows, one for each language: its
# home page as the site is built, laid out as a picture 1200 by 630 pixels and
# photographed in Chromium, from the image the widget's layout is measured
# in. They are written over site/assets/, for the change to be looked at
# before it is kept; a change to the home page's first screen deserves new
# ones.
# Photograph the site's sharing pictures
[working-directory('web')]
site-og: site _playwright-pinned
    docker run --rm --init --shm-size=1g --user "$(id -u):$(id -g)" -e HOME=/tmp \
        -v "{{ justfile_directory() }}:{{ justfile_directory() }}" -w "{{ justfile_directory() }}/web" \
        {{ PLAYWRIGHT_IMAGE }} node scripts/og.ts

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

# The SQL of the counts kept for years — the nightly script, the tables it
# fills and the views two reports read — run against a BigQuery emulator
# started for the run, on a port of its own, and removed after it. The tests
# are the files Terraform hands BigQuery, filled in the same way; they stand
# behind the build tag analytics, since a plain test run starts no container.
# Test the SQL of the counts kept for years against a BigQuery emulator
analytics-test:
    #!/usr/bin/env bash
    set -euo pipefail
    name="mathtrail-bigquery-$$"
    docker run -d --rm --name "$name" -p 127.0.0.1::9050 {{ BIGQUERY_EMULATOR_IMAGE }} --project=mathtrail > /dev/null
    trap 'docker rm -f "$name" > /dev/null 2>&1 || true' EXIT
    address="http://$(docker port "$name" 9050/tcp | head -1)"
    for _ in $(seq 1 60); do
        if curl -fs -o /dev/null "$address/bigquery/v2/projects/mathtrail/datasets"; then
            break
        fi
        sleep 1
    done
    MATHTRAIL_BIGQUERY_EMULATOR="$address" go test -tags analytics -race -count=1 ./infra/analytics/...

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
    # The runtime's right to read the key the children are counted under comes
    # with them, so that it has spread through the platform by the time the
    # service is told to read it.
    tf apply -auto-approve -input=false \
        -target=google_secret_manager_secret.seal_key \
        -target=google_secret_manager_secret.learner_key \
        -target='google_secret_manager_secret_iam_member.runtime["learner_key"]' \
        -target=google_secret_manager_secret.google_client_secret

    # The project is read from the file that names it rather than from an
    # output: an apply that was given targets refreshes only the outputs those
    # targets feed, and this one is fed by a variable.
    project=$(just _project)

    seal_key=$(tf output -raw secret_seal_key)
    learner_key=$(tf output -raw secret_learner_key)
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

    # The key the children are counted under is made the same way, and once:
    # it is never rotated with the sealing key.
    if has_version "$learner_key"; then
        echo "learner key: a version exists, leaving it alone"
    else
        echo "learner key: generating the first version"
        head -c 32 /dev/urandom | base64 | tr -d '\n' \
            | gcloud secrets versions add "$learner_key" --project="$project" --data-file=- > /dev/null
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

# The project the deployment lives in, as the Terraform configuration names it
_project:
    #!/usr/bin/env bash
    set -euo pipefail
    named="{{ TF_DIR }}/prod.auto.tfvars"
    if [ ! -f "$named" ]; then
        echo "$named is missing, and it is what names the project." >&2
        exit 1
    fi
    project=$(sed -n 's/^[[:space:]]*project_id[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$named" | head -1)
    if [ -z "$project" ]; then
        echo "$named names no project_id." >&2
        exit 1
    fi
    echo "$project"

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

# The month is DB-IP's: its file of that month, downloaded from DB-IP and kept
# unchanged in an image holding nothing else, under the month as its tag. A tag
# once published is never replaced, so the same month is the same bytes for as
# long as anything builds from it, and the Dockerfile is pinned to it by tag and
# digest like every other image. The image runs nothing, so one built for this
# machine serves a build for any platform. Without a month, this month's.
# Publish a month of the database of countries as an image, and pin the Dockerfile to it
countries-publish month="":
    #!/usr/bin/env bash
    set -euo pipefail

    month="{{ month }}"
    if [ -z "$month" ]; then
        month=$(date -u +%Y-%m)
    fi
    if [[ ! "$month" =~ ^[0-9]{4}-(0[1-9]|1[0-2])$ ]]; then
        echo "countries-publish: $month is not a month written as YYYY-MM" >&2
        exit 1
    fi
    reference="{{ COUNTRIES_IMAGE }}:${month}"

    # docker is signed in to the registry with the token gh holds, in a
    # configuration of this run's own, kept apart from the build's context:
    # the token goes nowhere it would outlive the run, and a sign-in of the
    # user's own is left as it was. The check below needs it as much as the
    # push does, since a month not yet made public is hidden from a client
    # that is not signed in.
    temp=$(mktemp -d)
    trap 'rm -rf "$temp"' EXIT
    export DOCKER_CONFIG="$temp/docker"
    work="$temp/context"
    mkdir "$work"
    user=$(gh api user --jq .login)
    if ! signin=$(gh auth token | docker login "${reference%%/*}" --username "$user" --password-stdin 2>&1); then
        echo "$signin" >&2
        exit 1
    fi

    if docker buildx imagetools inspect "$reference" > /dev/null 2>&1; then
        echo "$reference: published already, leaving it alone"
    else
        echo "dbip-country-lite ${month}"
        curl -fsSL --proto "=https" --proto-redir "=https" --retry 3 --retry-all-errors \
            -o "$work/countries.mmdb.gz" "https://download.db-ip.com/free/dbip-country-lite-${month}.mmdb.gz"
        gunzip -c "$work/countries.mmdb.gz" > "$work/dbip-country-lite.mmdb"
        rm "$work/countries.mmdb.gz"
        printf 'FROM scratch\nCOPY dbip-country-lite.mmdb /dbip-country-lite.mmdb\n' \
        | docker build \
            --label "org.opencontainers.image.source=https://github.com/MathTrail/mathtrail-standalone" \
            --label "org.opencontainers.image.licenses=CC-BY-4.0" \
            --label "org.opencontainers.image.version=${month}" \
            --label "org.opencontainers.image.description=IP to Country Lite by DB-IP (https://db-ip.com), the file of ${month}, unchanged, under CC BY 4.0" \
            --tag "$reference" --file - "$work"
        docker push "$reference"
    fi

    # The digest, not the tag: a tag can be moved, and a version here is exact.
    digest=$(docker buildx imagetools inspect "$reference" | awk '/^Digest:/ { print $2 }')
    sed -i "s|^FROM {{ COUNTRIES_IMAGE }}:.* AS countries$|FROM ${reference}@${digest} AS countries|" Dockerfile
    git --no-pager diff -- Dockerfile
    echo
    echo "Then: just licenses, for the month the list names."

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

    # A run of just of its own would work the version out again, so it is
    # handed this one, which may have been set rather than worked out.
    just --set VERSION "{{ VERSION }}" ci-image-widget "$tag" >&2
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

# The report reads what the deployed service logged within the window given —
# 1d, 7d, 30d, as far back as Cloud Logging keeps it — in the project the
# Terraform configuration names: every line the service wrote, since the report
# holds each of them to the rules of the log as well as adding them up: what it
# writes on its standard output and its standard error, and none of the
# platform's own records, of requests or of the instances.
# The lines are read whole into a file of their own before anything is added
# up, so that a read that fails — a sign-in that ran out, a right the account
# lacks — stops here rather than passing for a log with nothing in it. Cloud
# Logging keeps a line's own fields as the payload of an entry and moves its
# severity, its time and its trace out beside it, so those are put back before
# the report reads the line, and the instance that wrote it, which the entry's
# labels name, is put beside them. An entry says a trace was kept only when it
# was: one that names a trace and says nothing of it was dropped. The entries
# are taken out of the file one at a time, so that a long window costs the
# disk rather than the memory; a long window over a busy service is many pages
# of the log, and is best read a day at a time. A log of a local run, such as
# `just play-server` writes, is read as it is: go run ./cmd/report < that file.
# Add up the deployed service's log: tasks, refusals, time to write one, limits, tool calls with and without Drive, traces, the busiest minute, the rules of the log
report since="1d" service="mathtrail":
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v gcloud > /dev/null; then
        echo "report: this environment has gcloud on x86_64 alone. Elsewhere, read the entries with Google's" >&2
        echo "google-cloud-cli image and hand them to the report as this recipe does." >&2
        exit 1
    fi
    project=$(just _project)
    filter='resource.type="cloud_run_revision" AND resource.labels.service_name="{{ service }}"
        AND (log_id("run.googleapis.com/stdout") OR log_id("run.googleapis.com/stderr"))'
    entries=$(mktemp)
    trap 'rm -f "$entries"' EXIT
    gcloud logging read "$filter" --project="$project" --freshness="{{ since }}" --format=json > "$entries"
    jq -cn --stream 'fromstream(1 | truncate_stream(inputs))
        | (.jsonPayload // {}) + ({
            severity,
            time: .timestamp,
            "logging.googleapis.com/trace": .trace,
            "logging.googleapis.com/spanId": .spanId,
            "logging.googleapis.com/trace_sampled": (if .trace then (.traceSampled // false) else null end),
            instance: .labels.instanceId
        } | with_entries(select(.value != null)))' "$entries" \
        | go run ./cmd/report

# What the platform counted of the deployed service within the window given —
# 30m, 2h, 1d, 30d — as Cloud Monitoring keeps it: the processor and the
# memory the instances were allocated, which is what the free tier is counted
# in, the requests that reached them, the time they were billed for, and the
# most instances that served in one minute. Each series is summed a minute at
# a time; the instances that served are taken on average over each minute and
# added up across revisions, so that counts taken at different moments of a
# minute never add up to more than ran at once. The window is set against the
# free tier of a month: 180,000 vCPU-seconds, 360,000 GiB-seconds and two
# million requests. The platform shows a minute up to three minutes after it
# ends, so a window that ends now may leave out its last minutes. The token
# rides in a header read from a file descriptor, never on the command line.
# What the deployed service used of Cloud Run within a window, against the free tier of a month
usage since="1h" service="mathtrail":
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v gcloud > /dev/null; then
        echo "usage: this environment has gcloud on x86_64 alone." >&2
        exit 1
    fi
    since="{{ since }}"
    if [[ ! "$since" =~ ^[1-9][0-9]{0,4}[mhd]$ ]]; then
        echo "usage: a window is whole minutes, hours or days, such as 30m, 2h or 1d, not $since" >&2
        exit 1
    fi
    case "$since" in
    *m) seconds=$(( ${since%m} * 60 )) ;;
    *h) seconds=$(( ${since%h} * 3600 )) ;;
    *d) seconds=$(( ${since%d} * 86400 )) ;;
    esac
    project=$(just _project)
    token=$(gcloud auth print-access-token)
    now=$(date -u +%s)
    end=$(date -u -d "@$now" +%Y-%m-%dT%H:%M:%SZ)
    start=$(date -u -d "@$(( now - seconds ))" +%Y-%m-%dT%H:%M:%SZ)
    series() {
        curl -fsS --max-time 30 -G "https://monitoring.googleapis.com/v3/projects/$project/timeSeries" \
            -H @<(printf 'Authorization: Bearer %s\n' "$token") \
            --data-urlencode "filter=metric.type=\"run.googleapis.com/$1\" AND resource.type=\"cloud_run_revision\" AND resource.labels.service_name=\"{{ service }}\"${3:+ AND $3}" \
            --data-urlencode "interval.startTime=$start" \
            --data-urlencode "interval.endTime=$end" \
            --data-urlencode "aggregation.alignmentPeriod=60s" \
            --data-urlencode "aggregation.perSeriesAligner=$2" \
            --data-urlencode "aggregation.crossSeriesReducer=REDUCE_SUM" \
            --data-urlencode "view=FULL"
    }
    summed='[.timeSeries[]?.points[]?.value | (.doubleValue // (.int64Value | tonumber))] | add // 0'
    most='[.timeSeries[]?.points[]?.value | (.doubleValue // (.int64Value | tonumber))] | max // 0'
    vcpu=$(series container/cpu/allocation_time ALIGN_SUM | jq "$summed")
    gib=$(series container/memory/allocation_time ALIGN_SUM | jq "$summed")
    requests=$(series request_count ALIGN_SUM | jq "$summed")
    billed=$(series container/billable_instance_time ALIGN_SUM | jq "$summed")
    instances=$(series container/instance_count ALIGN_MEAN 'metric.labels.state="active"' | jq "$most")
    share() { awk -v used="$1" -v free="$2" 'BEGIN { printf "%.3f %%", 100 * used / free }'; }
    echo "# What the platform counted"
    echo
    echo "Cloud Run, the service {{ service }}, from $start to $end, as Cloud Monitoring counts it."
    echo
    echo "| | Used | Of a month's free tier |"
    echo "|---|---:|---:|"
    echo "| vCPU-seconds | $(printf '%.1f' "$vcpu") | $(share "$vcpu" 180000) |"
    echo "| GiB-seconds | $(printf '%.1f' "$gib") | $(share "$gib" 360000) |"
    echo "| Requests | $(printf '%.0f' "$requests") | $(share "$requests" 2000000) |"
    echo "| Instance-seconds billed | $(printf '%.1f' "$billed") | |"
    echo "| Most instances that served in one minute | $(printf '%.1f' "$instances") | |"


# -- Golden vectors from the prototype --------------------------------------

# The export runs the prototype's own code, from a clone of PROTOTYPE_COMMIT made for
# the run and removed after it. Python comes from the uv image and PostgreSQL from the
# prototype's compose file, both pinned by tag and digest, so neither has to be
# installed anywhere; the database starts empty and is removed with its volume.
# Export the prototype's golden vectors into testdata/golden/ (see its export/README.md)
golden:
    #!/usr/bin/env bash
    set -euo pipefail
    uv_image="ghcr.io/astral-sh/uv:0.12.13-python3.12-trixie-slim@sha256:87bc72093c0aa93cc962bd7c0498ddf416dad3ce9e1434724e936e02b72afe5d"
    repo="$(pwd)"
    prototype="$(mktemp -d)"
    compose=(docker compose --project-name mathtrail-golden --project-directory "$prototype")

    # Whatever happens next, the database goes once it was asked for, and so does
    # the clone: with set -e a failing export would otherwise leave either behind
    # until somebody noticed. The clone goes even when the database will not, and
    # that is said.
    database=no
    cleanup() {
        if [ "$database" = yes ]; then
            echo "==> removing PostgreSQL"
            "${compose[@]}" down --volumes || echo "golden: the database of project mathtrail-golden was not removed" >&2
        fi
        rm -rf "$prototype"
    }
    trap cleanup EXIT

    # The one commit, and none of the history around it.
    echo "==> the prototype at {{ PROTOTYPE_COMMIT }}"
    git -C "$prototype" init --quiet
    git -C "$prototype" fetch --quiet --depth 1 "{{ PROTOTYPE_REPOSITORY }}" "{{ PROTOTYPE_COMMIT }}"
    git -C "$prototype" checkout --quiet --detach FETCH_HEAD

    echo "==> PostgreSQL"
    database=yes
    "${compose[@]}" up -d --wait

    echo "==> dependencies, schema, seed profiles, export"
    docker run --rm --network host \
        -v "$repo:/repo" -v "$prototype:/prototype" -w /prototype --user "$(id -u):$(id -g)" \
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

# -- Load --------------------------------------------------------------------

# The load tool is a Go module of its own under tools/load, so that what it
# builds with stays out of the service's dependencies and out of the license
# file the service ships. It imports the service, so its go.mod follows the
# service's: a change to the service's go.mod is followed by `just load-tidy`.

# A run starts the service itself, from the image of what is in the tree,
# built first, in a container of an instance's size — unless it is given a
# service somebody else started with -url, such as one `just run` started, or
# an image of its own with -image. Arguments go to the tool as they are. The
# tool is built and then run, rather than run through the go command, which
# answers every failing exit with 1: its own exit is 0 for a clean run, 1 for
# one that found something, and 2 for one that could not run.
# A deployed service is run against with the accounts a parent signed in on
# it, since it has no development sign-in: `just load paces -url
# https://mcp.mathtrail.app -accounts parent,load` (docs/load.md).
# Run a scenario of the load tool, such as `just load lesson`, or `just load lesson -url http://localhost:8080`
[positional-arguments]
[working-directory('tools/load')]
load scenario *args:
    #!/usr/bin/env bash
    set -euo pipefail
    scenario="$1"
    shift
    go build -o bin/load .
    for arg in "$@"; do
        case "$arg" in
        -url | -url=* | --url | --url=* | -image | -image=* | --image | --image=*)
            exec bin/load -scenario "$scenario" "$@"
            ;;
        esac
    done
    just docker-build
    exec bin/load -scenario "$scenario" -image mathtrail:dev "$@"

# The parent opens the address it prints and signs in with Google; the browser
# comes back to this computer, or the parent pastes the address it ended up at.
# The tokens are kept in the user's own configuration, never in the repository.
# The deployed service is the one signed in on unless the arguments name
# another with -url, which, given later, is the one the tool takes.
# Sign an account in on the deployed service for the load's runs against it, such as `just load-signin parent`
[positional-arguments]
[working-directory('tools/load')]
load-signin name *args:
    #!/usr/bin/env bash
    set -euo pipefail
    go build -o bin/load .
    name="$1"
    shift
    exec bin/load -signin "$name" -url https://mcp.mathtrail.app "$@"

# Every scenario of a container, against the image of what is in the tree,
# built once, in a container of an instance's size: one vCPU and 1 GiB, told
# what a deployment of that size is told; the two of the deployed service are
# run by hand against it. Each run is held to the most memory its instance may
# hold — its peak as measured, with room to spare; for the solvers that keep
# hundreds of MiB, whose peak is wherever the collector chose to work, just
# under the instance — and one that finds something does not stop the rest:
# every run reports, each under the command that ran it, and the worst exit is
# the recipe's. Saturation runs twice: once through the pace of the instance,
# and once with eighty hand-ins in flight past it, which is the sandbox's queue
# alone. Arguments go to every run, so `just ci-load -env
# MATHTRAIL_SOLVER_CONCURRENCY=32` runs them all against a sandbox of
# thirty-two slots on one processor. They are for what the service is told:
# the size of the instance and the ceilings were measured together, and a
# ceiling held against another -memory or -cpus means nothing.
# Run every load scenario against the image, holding each run to its memory
[positional-arguments]
[working-directory('tools/load')]
ci-load *args: docker-build
    #!/usr/bin/env bash
    set -uo pipefail
    go build -o bin/load . || exit 2
    echo "load: docker runs its containers' cgroups $(docker info --format '{{ DOCKER_CGROUPS }}')" >&2
    extra=("$@")
    worst=0
    one() {
        local command=(-image mathtrail:dev -cpus 1 -memory 1g "$@" "${extra[@]}")
        printf '> `load %s`\n\n' "${command[*]}"
        code=0
        bin/load "${command[@]}" || code=$?
        if ((code > worst)); then
            worst=$code
        fi
        echo
    }
    one -scenario lesson -memory-ceiling 64m
    one -scenario saturation -memory-ceiling 64m
    one -scenario saturation -children 80 -rate 27 -env MATHTRAIL_RATE_INSTANCE_PER_MIN=1000000 -memory-ceiling 96m
    one -scenario limits -memory-ceiling 64m
    one -scenario cold -memory-ceiling 64m
    one -scenario adversarial -variants appends -memory-ceiling 256m
    one -scenario adversarial -variants helper -memory-ceiling 128m
    one -scenario adversarial -variants tuples -memory-ceiling 960m
    one -scenario adversarial -variants pairs -memory-ceiling 960m
    one -scenario adversarial -variants sets -memory-ceiling 960m
    one -scenario adversarial -variants product -memory-ceiling 288m
    exit "$worst"

# The load tool's tests, with the race detector
[working-directory('tools/load')]
load-test:
    go test ./... -race -count=1

# A tool that is a module of its own under tools/ is held to what the service
# is: tidy, formatted, linted, linking only what the service's licenses allow,
# and free of known vulnerabilities. A module the tool builds with and the
# service also uses is held to the version the service's own build selects: the
# tool would otherwise run the service against code the service does not ship
# with. Only the modules that give the tool's build a package are compared —
# the service's module graph names older versions of modules neither build
# takes a package from. The tool is named by its directory, as its recipes are.
[positional-arguments]
_tool-lint tool:
    #!/usr/bin/env bash
    set -euo pipefail
    cd "tools/$1"
    if ! go mod tidy -diff > /dev/null; then
        echo "$1-lint: tools/$1/go.mod or go.sum is not what the code asks for: run just $1-tidy" >&2
        exit 1
    fi
    unformatted=$(gofmt -s -l .)
    if [ -n "$unformatted" ]; then
        echo "These files need gofmt -s; run just fmt:" >&2
        echo "$unformatted" >&2
        exit 1
    fi
    golangci-lint run ./...
    # The tool's own code and the service's are this repository's, under its
    # own license; what is checked is everything else the tool links.
    go run {{ GO_LICENSES }} check ./... --allowed_licenses={{ ALLOWED_LICENSES }} --ignore {{ MODULE }}
    go run {{ GOVULNCHECK }} ./...
    tool=$(go list -deps -test -f '{{ TOOL_MODULE_OF }}' ./... | LC_ALL=C sort -u)
    service=$(cd ../.. && go list -m all | LC_ALL=C sort -u)
    apart=$(LC_ALL=C join <(printf '%s\n' "$tool") <(printf '%s\n' "$service") | awk '$2 != $3')
    if [ -n "$apart" ]; then
        echo "$1-lint: the tool builds with other versions than the service (module, the tool's, the service's):" >&2
        echo "$apart" >&2
        exit 1
    fi

# The load tool's formatting, lint, licenses, known vulnerabilities, and the versions it shares with the service
load-lint: (_tool-lint "load")

# Bring the load tool's go.mod and go.sum to what its code asks for, as a change to the service's go.mod calls for
[working-directory('tools/load')]
load-tidy:
    go mod tidy

# -- Learners ----------------------------------------------------------------

# The learners' bench is a Go module of its own under tools/learners, as the
# load tool is: simulated children answer there through the service's own
# rule, profile and rating, so that a change to how the service places a child
# is measured before it ships, and nothing the bench needs reaches the service.
# It imports the service, so its go.mod follows the service's: a change to the
# service's go.mod is followed by `just learners-tidy`.

# A run writes its tables, its summary and what it was into
# tools/learners/results, over the run kept there, and prints the summary. It
# runs the bench's own set of rules unless -rules names another, whose results
# go into a directory of their own within results; it draws the children of
# the paper's run unless it is given -seed or -experiment, or a set that draws
# children of its own. Arguments go to the bench as they are; a quick look,
# such as `just learners -children 100 -out /tmp/learners`, is written
# elsewhere, so that the run kept is a whole one. The service's version is
# stamped in as a build of the service carries it, so that run.txt says what
# code the numbers came from.
# Run a set of the learners' bench's rules on every generator, and print the summary
[positional-arguments]
[working-directory('tools/learners')]
learners *args:
    go run -ldflags "-X {{ SYMBOLS }}.Version={{ VERSION }}" . -out results "$@"

# The learners' bench's tests, with the race detector
[working-directory('tools/learners')]
learners-test:
    go test ./... -race -count=1

# The learners' bench's formatting, lint, licenses, known vulnerabilities, and the versions it shares with the service
learners-lint: (_tool-lint "learners")

# Bring the learners' bench's go.mod and go.sum to what its code asks for, as a change to the service's go.mod calls for
[working-directory('tools/learners')]
learners-tidy:
    go mod tidy

# -- Research ----------------------------------------------------------------

# The recipes live in research/justfile. The arguments reach it exactly as
# given, without passing through the shell a second time.
[doc('A recipe of the research module: `just research` lists them')]
[positional-arguments]
research *ARGS:
    @{{ quote(just_executable()) }} --justfile research/justfile --working-directory research "$@"
