# CLAUDE.md — working rules for this repository

MathTrail v1 is a free, open-source (MIT) MCP app for olympiad-style maths, grades 1–6, that runs inside Claude and ChatGPT. The chat's own model writes each task. This service:

- picks the topic and difficulty with a deterministic rule;
- hands the model a brief, reference examples and formats;
- checks the submitted task: structure, a Starlark solver that brute-forces the options, the model's self-check, readability, near-duplicates and text drawings;
- shows the task in MCP Apps widgets without the answer;
- records the answer, explains the mistake and keeps Elo/IRT ratings.

Technical shape:

- one stateless Go binary on Google Cloud Run, with no database;
- MCP 2026-07-28 over stateless Streamable HTTP;
- the service is its own OAuth 2.1 authorization server, with Google sign-in; its state lives in sealed tokens;
- the child's profile is a JSON file in the parent's Google Drive (`drive.file`);
- widgets are Preact, bundled into one HTML file embedded in the binary;
- UI strings are looked up by BCP 47 locale;
- the service makes no LLM API calls of its own.

## Documents

- [PRODUCT-V1.md](PRODUCT-V1.md) — product requirements, platform research and the product decision log (О-questions in section 12). The source of truth for *what* we build.
- [RUN.md](RUN.md) — the implementation plan: tasks T00–T65, run one at a time. Full executor rules are in its «Как пользоваться» section.
- `SPEC.md` — the technical spec (written in RUN.md phase 1). The source of truth for *how*, once it exists.
- `docs/decisions.md` — implementation decision log, numbered R01… (created in T05). Read it before proposing design changes.
- `docs/architecture/` — Mermaid diagrams (phase 1). `docs/live/` — reports of live runs in real chat hosts.
- [docs/prototype/](docs/prototype/README.md) — frozen copies of the prototype SPEC, its decision log D01–D45 and research. Prototype decisions are cited as "D43 of the prototype".
- `prototype/` — a full local copy of the prototype repository (code, data, schemas, prompts, tests, diagrams); see "The prototype copy" below.

Project context lives in these files, not in chat history.

## How we work

- The author runs tasks one at a time: «Выполни задачу Txx из RUN.md». Do only that task; do not touch files that belong to other tasks unless the task says so.
- Before starting, read the RUN.md task and the PRODUCT-V1 / SPEC sections it cites. If the task contradicts them or something is missing, stop and ask. Do not silently fill gaps; record them as open questions (PRODUCT-V1 12.2, next free О-number) or in the "Замечания" section of the document being written.
- When done, mark the task in the RUN.md summary table: `[ ]` → `[x]`.
- Finish with a short report in Russian: what was done, which files, how to check it (commands), what is still open.
- Do not commit: the author commits after review.
- Live runs in Claude or ChatGPT spend the author's subscription limits, and cloud actions may cost money. Say so and wait for confirmation before starting one.
- Do not re-propose alternatives rejected in PRODUCT-V1, `docs/decisions.md` or the prototype decision log without new evidence. If a decision has to change, ask the author and update the log.
- Do not write SDK, protocol or API code from memory: check the current docs of the pinned version first (go-sdk, MCP Apps spec, Drive API, starlark-go, Google OAuth).

## Hard rules

- **Devcontainer only.** All work happens inside the devcontainer (`.devcontainer/`, from T01): building, tests, linters, mocks, running the server, the widget build, Terraform, gcloud and live runs. Nothing but Docker and VS Code is installed or run on the host. If you are not inside the container (`MATHTRAIL_DEVCONTAINER` is not `1`), stop and ask the author to "Reopen in Container". T00 and T01 (which creates the container) are the only host-side tasks.
- **Exact versions only.** No `latest`, no floating tags or ranges. This covers:
  - Docker images and devcontainer features: exact tag plus digest;
  - Go and its modules: exact versions in `go.mod`; starlark-go has no tags, so it is pinned by pseudo-version;
  - npm dependencies: exact versions plus a lockfile;
  - tools and VS Code extensions: exact versions;
  - GitHub Actions: pinned by commit SHA.

  Each tool's exact version is pinned directly where it is installed — as an `ARG` in
  `.devcontainer/Dockerfile`, as a literal in `devcontainer.json`. No separate versions file,
  no checksum verification.
- **Language:**
  - English: code, every comment in code and config files (Go, TS, SQL, YAML, HCL, Dockerfile, justfile, `.env.example`, …), model instructions and content in `content/`, locale keys, `CLAUDE.md` and `README.md`.
  - Russian: project documents (PRODUCT-V1, RUN, SPEC, `docs/`) and reports to the author.
  - Tasks for the child are written by the chat's model in the chat language.
- **One contest is never named.** Our tasks are olympiad-style problems. The repository never names the international multiple-choice contest the project started from, in any language (prototype D36). Other olympiads may be named, e.g. as problem sources. Content wording is always our own.
- **Children's data.**
  - Only a pseudonym, never a real name, birth date or school.
  - No personal data, task text or answers in logs or metrics.
  - The pseudonym never goes into task text.
- **The answer stays hidden** until the child answers. It must not appear in widget payloads, in unsealed fields of the Drive file or in logs.
- **Secrets** live only in environment variables, a local `.env` (never committed) or Secret Manager. No keys, tokens or client secrets in the repository.
- **Licenses.** The code and the content are MIT. Only add dependencies with MIT-compatible licenses, and keep the third-party license list current. The name MathTrail is not licensed to forks (О-4).

## Go standards

These follow `mentor-api`, with its known gaps fixed.

- **Layout:**
  - `cmd/server/main.go`;
  - `internal/` with `app` (DI container, HTTP server), `config`, `logger`, `version`, and one package per domain area;
  - embedded content in `content/`, the widget sources in `web/`.
- **Wiring:**
  - a hand-written container, `app.NewContainer(ctx, cfg, logger)`; resources close in reverse order, including when construction fails halfway;
  - `signal.NotifyContext` plus `errgroup` in `main`;
  - graceful shutdown uses `context.WithoutCancel(ctx)` with a timeout.
- **Interfaces:**
  - defined in the provider package; `NewX(...)` returns the interface and the implementation stays unexported;
  - mocks are generated by mockery v3 (`.mockery.yml`, testify template with expecters) into a `mocks/` folder next to the interface and committed. CI checks that they are up to date.
- **Config:**
  - environment variables only;
  - defaults as named constants;
  - `Validate()` returns errors that name the variable;
  - the listen port comes from `PORT` (Cloud Run);
  - dev-only switches, such as the dev auth stub, refuse to start when `K_SERVICE` is set.
- **Logging:**
  - `go.uber.org/zap`, injected into constructors, never global;
  - JSON encoder with the Cloud Run keys `severity`, `message` and `time`;
  - lowercase messages and snake_case keys;
  - no PII, task text or answers — log aggregates only (О-16), and every task result carries the instructions version (О-21).
- **Errors:**
  - wrap with short context: `fmt.Errorf("drive: save profile: %w", err)`;
  - use sentinel errors where callers branch on them, checked with `errors.Is`/`errors.As`;
  - log an error once, at the boundary;
  - never send internal error text to the client; tools return clear, user-facing messages the model can relay.
- **Context:** `ctx context.Context` is the first parameter of anything that blocks or does I/O; slow calls get their own child timeout.
- **Docs:** every package has a package comment; every exported identifier has a doc comment starting with its name. Comments explain *why*.
- **Version:** `internal/version` variables are set through `-ldflags -X` in both the Dockerfile and CI builds.
- **Tests:**
  - standard `testing` with `t.Fatalf`/`t.Errorf` in "got …, want …" form;
  - testify only for generated mocks;
  - table-driven subtests (`t.Run`) for similar cases;
  - external `_test` packages for public behaviour, internal ones for unexported helpers;
  - `go test -race` always; fixtures under `testdata/`;
  - no network in unit tests: Google, Drive and the chat hosts are faked with `httptest`.
- **Tooling:**
  - `gofmt -s`, plus `golangci-lint` with a committed `.golangci.yml` and a pinned version;
  - a pre-commit hook in `.githooks/`, enabled by `post-start`;
  - `govulncheck` and `gitleaks` in CI;
  - the justfile is the entry point for everything; `ci-*` recipes are what CI calls.
- **Images:** multi-stage Dockerfile, minimal non-root runtime image pinned by digest, `CGO_ENABLED=0`, `-trimpath`.

## The prototype copy

`prototype/` is a local, gitignored copy of the prototype repository. It serves as reference material for porting: catalogs and examples in `data/`, `schemas/`, `prompts/`, `src/taskgen/`, `tests/` (including `tests/example_checks/`) and `docs/architecture/`.

- Its rules file is renamed to `prototype/CLAUDE.prototype.md`; its rules do not apply here.
- Search tools that respect `.gitignore` (ripgrep) skip it. Use explicit paths (`Read`, `ls`, `grep -r prototype/...`).
- Never edit it and never commit anything from it except through a RUN.md task that ports content into the product.
- T65 deletes it once everything needed has been ported.

## Environment

Open the repository in VS Code and choose "Reopen in Container" (`.devcontainer/`). Inside the container:

- **Pinned versions.** No separate versions file: each tool's exact version is an `ARG` default in `.devcontainer/Dockerfile`, repeated as a literal in `.devcontainer/devcontainer.json` where needed.
  - The devcontainer image installs Go, gopls, dlv, Node.js, just, golangci-lint, mockery, cloudflared, the Docker Compose plugin and Claude Code straight from the Dockerfile, no checksum verification beyond what `go install` already does for Go modules.
  - Docker (docker-in-docker, Moby engine and buildx) comes from a devcontainer feature pinned in `devcontainer.json` and `devcontainer-lock.json`.
  - To change a version: edit the literal everywhere it appears (Dockerfile, `devcontainer.json` for Docker or the Claude Code extension), rebuild the container.
- **Container marker.** `MATHTRAIL_DEVCONTAINER=1` is set only inside the container. A session that does not see it is on the host and must stop (see "Devcontainer only").
- **Claude Code.** Its config lives in the `mathtrail-standalone-claude` volume, so the login survives rebuilds. Auto-update is off; the version is pinned.
- **Just recipes.** `just --list` shows all recipes.
- **Build context.** The devcontainer image is built with the repository root as context; `.dockerignore` keeps `.git`, `.env`, `prototype/` and `node_modules` out of it.
