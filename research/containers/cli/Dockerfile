# Codex CLI and Gemini CLI at exact versions, for trial runs of models through
# their subscriptions. They live apart from the devcontainer, whose image the
# product owns; their logins live in named volumes mounted at runtime, never in
# the image.
#
# The Node base is the devcontainer's version, pinned by tag and digest. The
# packages come from the lockfile beside this file, and no package runs an
# install step of its own. The two optional native helpers that need one go
# unbuilt: a terminal emulator, which a trial run does not use, and access to
# the system keychain, which a container does not have — without it Gemini CLI
# keeps its login in an encrypted file in its home, which is the login volume.
FROM node:24.21.0-bookworm-slim@sha256:0e0ff40c39bc087845bfb27465a0df4ea419520094bc35842ff83dd8cbe6f9b6

WORKDIR /opt/cli
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts --no-audit --no-fund && npm cache clean --force

ENV PATH=/opt/cli/node_modules/.bin:$PATH

# Gemini CLI reads system policies from this directory; the one here denies
# every tool.
COPY deny-all-tools.toml /etc/gemini-cli/policies/deny-all-tools.toml

# The logins' volumes mount over these directories. An empty volume takes the
# ownership of the directory it covers, so the user who runs the CLIs can write
# to it.
RUN mkdir -p /home/node/.codex /home/node/.gemini && chown node:node /home/node/.codex /home/node/.gemini

# A run starts in an empty working directory that holds nothing of the project.
USER node
WORKDIR /work
