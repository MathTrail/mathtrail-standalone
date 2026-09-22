# Development container for MathTrail standalone.
# Exact versions only: the base image is pinned by digest, every tool below by an ARG default.
# Kept compatible with the classic builder as well as BuildKit: no heredocs, no RUN --mount,
# no TARGETARCH (arch is detected at RUN time via dpkg instead).
#
# Two stages. `toolchain` holds what it takes to build and check the code, and is the stage
# the automated checks run inside. `devcontainer` adds what a person needs on top of that and
# is the default stage, so building this file with no target gives the full environment.

# -- toolchain: builds and checks the code -----------------------------------

FROM mcr.microsoft.com/devcontainers/base:3.0.8-noble@sha256:d7c468679f45a52ad3673d06656b5bf16e488990b17216a1fa295d9e1f89d724 AS toolchain

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

ARG GO_VERSION=1.27.1
# Node.js LTS for the widget build.
ARG NODE_VERSION=24.21.0
ARG JUST_VERSION=1.58.0
ARG GOLANGCI_LINT_VERSION=2.13.2
ARG MOCKERY_VERSION=3.8.0

RUN arch="$(dpkg --print-architecture)"; \
    case "$arch" in \
        amd64) go_arch=amd64; node_arch=x64; gnu_arch=x86_64; mockery_arch=x86_64 ;; \
        arm64) go_arch=arm64; node_arch=arm64; gnu_arch=aarch64; mockery_arch=arm64 ;; \
        *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    workdir="$(mktemp -d)"; \
    \
    echo "go ${GO_VERSION}"; \
    curl -fsSL --retry 3 "https://go.dev/dl/go${GO_VERSION}.linux-${go_arch}.tar.gz" | tar -C /usr/local -xz; \
    \
    echo "node ${NODE_VERSION}"; \
    mkdir -p /usr/local/node; \
    curl -fsSL --retry 3 "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-${node_arch}.tar.gz" \
        | tar -C /usr/local/node --strip-components=1 -xz; \
    \
    echo "just ${JUST_VERSION}"; \
    curl -fsSL --retry 3 "https://github.com/casey/just/releases/download/${JUST_VERSION}/just-${JUST_VERSION}-${gnu_arch}-unknown-linux-musl.tar.gz" \
        | tar -C "$workdir" -xz just; \
    install -m 0755 "$workdir/just" /usr/local/bin/just; \
    \
    echo "golangci-lint ${GOLANGCI_LINT_VERSION}"; \
    curl -fsSL --retry 3 "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${go_arch}.tar.gz" \
        | tar -C "$workdir" --strip-components=1 -xz "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${go_arch}/golangci-lint"; \
    install -m 0755 "$workdir/golangci-lint" /usr/local/bin/golangci-lint; \
    \
    echo "mockery ${MOCKERY_VERSION}"; \
    curl -fsSL --retry 3 "https://github.com/vektra/mockery/releases/download/v${MOCKERY_VERSION}/mockery_${MOCKERY_VERSION}_Linux_${mockery_arch}.tar.gz" \
        | tar -C "$workdir" -xz mockery; \
    install -m 0755 "$workdir/mockery" /usr/local/bin/mockery; \
    \
    rm -rf "$workdir"

ENV PATH="/usr/local/go/bin:/usr/local/node/bin:/home/vscode/go/bin:/home/vscode/.local/bin:${PATH}" \
    GOTOOLCHAIN=local

# -- devcontainer: the toolchain, plus what a person works with --------------

FROM toolchain AS devcontainer

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

# Go editor tooling, installed with `go install`; the Go extension must not fetch its own
# copies (see devcontainer.json settings).
ARG GOPLS_VERSION=0.23.0
ARG DELVE_VERSION=1.27.2
# Tunnel for live runs against chat hosts.
ARG CLOUDFLARED_VERSION=2026.9.1
# Docker Compose plugin. Installed by us because the docker-in-docker feature cannot pin it;
# ours in /usr/local/lib/docker/cli-plugins takes precedence over the feature's copy.
ARG DOCKER_COMPOSE_VERSION=5.5.1
# Claude Code CLI. The VS Code extension in devcontainer.json must have the same version.
ARG CLAUDE_CODE_VERSION=2.1.273
# The cloud is described in Terraform and signed into with gcloud, both from here.
ARG TERRAFORM_VERSION=1.16.3
ARG GCLOUD_VERSION=585.0.0

RUN arch="$(dpkg --print-architecture)"; \
    case "$arch" in \
        amd64) go_arch=amd64; gnu_arch=x86_64 ;; \
        arm64) go_arch=arm64; gnu_arch=aarch64 ;; \
        *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    \
    echo "cloudflared ${CLOUDFLARED_VERSION}"; \
    curl -fsSL --retry 3 -o /usr/local/bin/cloudflared "https://github.com/cloudflare/cloudflared/releases/download/${CLOUDFLARED_VERSION}/cloudflared-linux-${go_arch}"; \
    chmod 0755 /usr/local/bin/cloudflared; \
    \
    echo "docker compose ${DOCKER_COMPOSE_VERSION}"; \
    install -d -m 0755 /usr/local/lib/docker/cli-plugins; \
    curl -fsSL --retry 3 -o /usr/local/lib/docker/cli-plugins/docker-compose "https://github.com/docker/compose/releases/download/v${DOCKER_COMPOSE_VERSION}/docker-compose-linux-${gnu_arch}"; \
    chmod 0755 /usr/local/lib/docker/cli-plugins/docker-compose

# Infrastructure tooling. The x86_64 archive of gcloud carries a Python interpreter of
# its own and runs straight out of the tarball; the arm one expects to find a system
# Python, which this image deliberately does not have, so there gcloud is left out and
# the same commands are run from Google's own container image instead.
RUN arch="$(dpkg --print-architecture)"; \
    case "$arch" in \
        amd64) tf_arch=amd64 ;; \
        arm64) tf_arch=arm64 ;; \
        *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    workdir="$(mktemp -d)"; \
    \
    echo "terraform ${TERRAFORM_VERSION}"; \
    curl -fsSL --retry 3 -o "$workdir/terraform.zip" "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${tf_arch}.zip"; \
    unzip -q -d "$workdir" "$workdir/terraform.zip"; \
    install -m 0755 "$workdir/terraform" /usr/local/bin/terraform; \
    \
    if [ "$arch" = "amd64" ]; then \
        echo "gcloud ${GCLOUD_VERSION}"; \
        curl -fsSL --retry 3 "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${GCLOUD_VERSION}-linux-x86_64.tar.gz" | tar -C /usr/local -xz; \
    else \
        echo "gcloud ${GCLOUD_VERSION}: skipped, no interpreter in the ${arch} archive"; \
    fi; \
    \
    rm -rf "$workdir"

ENV PATH="/usr/local/google-cloud-sdk/bin:${PATH}"

# gopls and dlv are built from source; the Go checksum database verifies the modules.
RUN echo "gopls ${GOPLS_VERSION}, dlv ${DELVE_VERSION}" \
    && GOBIN=/usr/local/bin go install "golang.org/x/tools/gopls@v${GOPLS_VERSION}" \
    && GOBIN=/usr/local/bin go install "github.com/go-delve/delve/cmd/dlv@v${DELVE_VERSION}"

# The config dirs are created here so the named volumes mounted over them inherit
# vscode ownership.
USER vscode
RUN mkdir -p /home/vscode/.claude /home/vscode/.config/gcloud \
    && curl -fsSL https://claude.ai/install.sh | bash -s "${CLAUDE_CODE_VERSION}"
USER root
