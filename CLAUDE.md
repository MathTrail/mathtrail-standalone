# CLAUDE.md — working rules for this repository

## Your role

You are a world-class Go engineer and software architect: you design before you type, and you write code meant to be read and maintained for years.

That bar is concrete, not a slogan. Before you call anything done:

- the change reads top to bottom — no jumping between files to follow one thought;
- names state intent, so comments are rarely needed;
- each package has one reason to change;
- a failing test names exactly what broke.

Clean Architecture, SOLID and KISS are the defaults, and they serve readability — not the other way around. When a principle would add a layer nobody needs, simplicity wins and you say why. When a task pushes you toward "good enough for now", say so out loud instead of quietly shipping it.

## The project

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
- `prototype/` — a full local copy of the prototype repository (code, data, schemas, prompts, tests, diagrams); see "Reference copies" below.
- `research/` — the research program for two papers: its plan [research/RUN.md](research/RUN.md) (tasks S00–S76, in Russian), evidence, literature, experiments and paper sources, and a Go module of its own. It never changes the product's code or behaviour; it touches the repository's tooling only where its plan says so.

Project context lives in these files, not in chat history.

## How we work

- The author runs tasks one at a time: «Выполни задачу Txx из RUN.md». Do only that task; do not touch files that belong to other tasks unless the task says so.
- Before starting, read the RUN.md task and the PRODUCT-V1 / SPEC sections it cites. If the task contradicts them or something is missing, stop and ask. Do not silently fill gaps; record them as open questions (PRODUCT-V1 12.2, next free О-number) or in the "Замечания" section of the document being written.
- Research tasks (S00–S76) are run, checked and marked in `research/RUN.md`, under that file's rules; the bullets here that name RUN.md mean the product's plan. They have an autonomous mode, which the author turns on and off in that file. While it is on, they run one after another without waiting for review, and their checks and reviews follow that file. A gap or contradiction there does not stop the work: the executor resolves it, records the decision with its reason, and leaves for the author only the questions that file lists as critical.
- Before calling any change done — code, config, docs or content — run `/code-review high` on it. Skip it only when the author asks, for that change alone, and say in the report that it did not run.
  - Run it once the change is written and its checks — `just ci-lint`, `just ci-test`, whatever else the task names — are green.
  - It reads the working tree plus the branch's commits — those not yet pushed, or all since `main` without an upstream — but not untracked files. Mark new files with `git add -N` for the review and unmark them with `git reset -- <file>` afterwards: the mark makes `git stash` fail. Pass no path — a path is reviewed instead of the diff, not beside it.
  - Verify each finding against the repository before acting on it; do not use `--fix`. Fix a real defect your uncommitted work introduced, wherever it shows up, with a test that fails without the fix when the defect is in code. A defect your work did not introduce — in an earlier commit, in code your work does not touch, or in another session's uncommitted work in the same tree — goes to the report as open.
  - Drop a wrong finding with a one-line reason, and one that re-proposes a rejected alternative with no new evidence by naming the decision. A finding that questions PRODUCT-V1, SPEC or a decision for a reason that holds goes to the author as a question.
  - If anything was fixed, run the checks and the review once more. That second round is the last: its findings are handled the same way, and its fixes get the checks but no third review. The report says what was fixed, what was dropped and why, and what is left open.
  - `/code-review ultra` is billed and started only by the author; this rule does not cover it.
- When done, mark the task in the RUN.md summary table: `[ ]` → `[x]`.
- Finish with a short report in Russian: what was done, which files, how to check it (commands), what the review found, what is still open.
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
- **Language: English everywhere.** Code, every comment in code and config files (Go, TS, SQL, YAML, HCL, Dockerfile, justfile, `.env.example`, …), model instructions and content in `content/`, locale keys, and every project document — PRODUCT-V1, SPEC, everything under `docs/`, `CLAUDE.md` and `README.md`.
  - RUN.md and research/RUN.md are in Russian: they are the author's working plans of the tasks.
  - Reports to the author in chat are in Russian.
  - The frozen prototype copies under `docs/prototype/` stay in Russian as they were written: they are a historical record, cited but never rewritten.
  - Tasks for the child are written by the chat's model in the chat language.
- **Children's data.**
  - Only a pseudonym, never a real name, birth date or school.
  - No personal data, task text or answers in logs or metrics.
  - The pseudonym never goes into task text.
- **The answer stays hidden** until the child answers. It must not appear in widget payloads, in unsealed fields of the Drive file or in logs.
- **Secrets** live only in environment variables, a local `.env` (never committed) or Secret Manager. No keys, tokens or client secrets in the repository.
- **Licenses.** The code and the content are MIT. Only add dependencies with MIT-compatible licenses, and keep the third-party license list current. The name MathTrail is not licensed to forks (О-4).

## Architecture and code standards

Checked against `reference/mentor-api/`, the platform's reference service. Where we differ from it on purpose, the reason is spelled out — those are the "known gaps" mentioned below.

### Layout

```
cmd/server/main.go
cmd/sitegen/ sitecheck/   the public site: renders it, and refuses a broken one
internal/
  app/            DI container, HTTP server
  config/ logger/ version/
  domain/         pure logic — no I/O, no SDKs, no transport
    rating/ tutor/ checks/ solver/ profile/
  infra/          everything that talks to the outside world
    drive/ starlark/ seal/
  site/           render/ turns the site's sources into files, check/ judges them
  store/          storage interface plus its in-memory and Drive implementations
  transport/
    mcp/          tools, ui:// resources
    http/         /health, OAuth endpoints
content/          catalogs, reference tasks and model instructions, embedded
site/             the public site: texts per locale, templates, assets
web/              widget sources
infra/terraform/  the Google Cloud project as code — not to be confused with internal/infra/
research/         the research program: a separate Go module that imports the product, never the reverse
```

`mentor-api` groups the same way (`domain/`, `infra/`, `transport/`) and serves HTTP with gin, which we do too — but it puts `repository.go` and `handler.go` inside each domain package, so its domain imports gin and pgx. We keep the grouping and the framework, and drop that part: our domains are computation — ratings, checks, the solver contract — and they stay free of I/O. That is what makes them testable against the golden vectors from T16 with no mocks at all.

### The dependency rule

Dependencies point inward: `app` → `transport` → `domain`, and `infra` implements what `domain` declares. Nothing in `internal/domain/` may import `net/http`, the MCP SDK, the Drive API or `starlark-go`.

A domain package reaching for infrastructure is not a style problem — it means the code is in the wrong package. Move the I/O out to `infra/` or `transport/` and leave the decision-making behind:

```go
// wrong — rating now needs a network to be tested
func (s *Service) Update(ctx context.Context, userID string, correct bool) error {
    p, err := s.drive.LoadProfile(ctx, userID)  // domain doing I/O
    ...
}

// right — a pure function; the caller in transport/mcp owns loading and saving
func Update(p Profile, topic string, outcome Outcome) Profile
```

### Interfaces

As in `mentor-api`: the interface lives in the package that provides it, `NewX(...)` returns the interface, and the implementation stays unexported.

```go
type Storage interface {                          // what the package offers
    Load(ctx context.Context, userID string) (*profile.Profile, Revision, error)
    Save(ctx context.Context, p *profile.Profile, expected Revision) error
}
type driveStore struct{ client *drive.Client }    // unexported
func NewDriveStore(c *drive.Client) Storage       // constructor returns the interface
```

This is deliberately *not* the usual Go advice of "accept interfaces, return structs": we keep the platform's convention so that every service in MathTrail reads the same way and mockery always has an interface to generate from. Do not "fix" it back.

Keep them narrow — every interface in `mentor-api` has one to three methods. Past five, split it: an interface that large usually means the type does several jobs.

Extend by adding an implementation, not another branch. A second `switch` over kinds of storage or kinds of host is the signal to reach for an interface instead.

Implementations of the same interface must be interchangeable, and that is worth proving in code: `store` ships one contract test set that both the in-memory and the Drive implementation run (T40).

### Simplicity beats purity

- Introduce an abstraction when there is a second consumer, not in anticipation of one. Three similar lines beat a premature interface.
- No layer that only forwards calls. If a type's whole body is `return s.next.Do(x)`, delete it.
- If explaining the design takes longer than reading the code, the design is wrong.
- When simplicity and a principle collide, simplicity wins — and the trade-off goes into `docs/decisions.md` so nobody relitigates it later.

### Naming and size

- Packages are named after the domain (`rating`, `profile`, `solver`). Never `utils`, `helpers`, `common`, `misc` — those names attract anything and explain nothing.
- A function does one thing. If you have to scroll back to remember what it was doing, split it.
- A file that cannot be taken in at a glance usually means the package mixes topics.
- Errors read `fmt.Errorf("drive: save profile: %w", err)`: package, action, wrapped cause.

### Wiring

- A hand-written container, `app.NewContainer(ctx, cfg, logger)` — the same shape as `mentor-api`. No DI framework.
- Resources close in reverse order, including when construction fails halfway.
- `signal.NotifyContext` plus `errgroup` in `main`; graceful shutdown uses `context.WithoutCancel(ctx)` with a timeout.

### Config

- Environment variables only, read in `internal/config`, nowhere else (PRODUCT 6). Viper does the reading, as in `mentor-api`: one declaration per variable carries its name, its default and its type, and no file or remote source is registered.
- Defaults as named constants; `Validate()` returns errors that name the offending variable.
- The listen port comes from `PORT`.
- Dev-only switches, such as the dev auth stub, refuse to start when `K_SERVICE` is set.

### Logging

- `go.uber.org/zap`, injected into constructors, never global.
- JSON encoder with the Cloud Run keys `severity`, `message` and `time`. (`mentor-api` logs `ts` in ISO8601 — different runtime, different keys.)
- Lowercase messages, snake_case keys.
- No PII, task text or answers — aggregates only (О-16), and every task result carries the instructions version (О-21).

### Errors

- Wrap with short context; use sentinel errors where callers branch, checked with `errors.Is`/`errors.As`.
- Log an error once, at the boundary.
- Never send internal error text to the client; tools return clear, user-facing messages the model can relay.

### Context and docs

- `ctx context.Context` is the first parameter of anything that blocks or does I/O; slow calls get their own child timeout.
- Every package has a package comment; every exported identifier has a doc comment starting with its name. Comments explain *why*.
- `internal/version` variables are set through `-ldflags -X` in both the Dockerfile and CI builds.

### Tests

- Standard `testing` with `t.Fatalf`/`t.Errorf` in "got …, want …" form — the house style in `mentor-api`, where testify appears only as `mock.Anything` in generated mocks.
- Table-driven subtests (`t.Run`) for similar cases, with `t.Parallel()` where a case shares no state; external `_test` packages for public behaviour, internal ones for unexported helpers.
- Test helpers take `*testing.T` and start with `t.Helper()`, so a failure points at the test that broke, not at the helper.
- Anything that parses what the chat's model produced — task JSON, readability, near-duplicates, the text drawing — gets a fuzz test (`FuzzXxx(f *testing.F)`). That input is untrusted, a panic there takes a request handler down with it, and fuzzing is how those inputs get found before a live host finds them.
- Computation that holds an invariant — ratings, `seal`, trigram similarity, readability, the solver contract, the copies the content hands out — gets a property test with `gopter`: the test names the invariant and the library looks for the input that breaks it. A property replaces neither the worked examples of the golden vectors nor a fuzz test; it is what catches the case nobody thought of.
- A new check counts as covered only when its test fails with the check switched off. Break it, watch the test go red, put it back — that is how a test that passes by accident is found, and it is cheaper than any tool that does the same thing by rewriting the code.
- `go test -race` always; fixtures under `testdata/`.
- Golden files in `testdata/golden/` are the prototype's numbers, produced by its own export script (T16). They are the reference we are checked against, so no Go test ever rewrites them — an `-update` flag on those would make the test green by definition. A snapshot the service itself owns may have `-update`, and then the regenerated diff is part of the review.
- No network in unit tests: Google, Drive and the chat hosts are faked with `httptest`.
- Behind an interface stands the real implementation, whenever that implementation does no I/O, and a small hand-written fake where a test has to force it to behave a particular way. That is what proves a property rather than the fact of a call, and it keeps what a test assumes in one readable place.
- A mock earns its place when the real implementation needs the network, a clock, or is simply slow — and when what is being checked is that a call happened, with these arguments. It is then generated by mockery v3 into a `mocks/` folder next to the interface and committed; CI checks it is up to date. Config lives in `.mockery.yaml`, with expecters on, mirroring `mentor-api`, and its list of packages is empty until an interface of that kind exists.

### Tooling and images

- `gofmt -s`, plus `golangci-lint` with a committed `.golangci.yml` and a pinned version. (`mentor-api` has no lint config checked in — this is one of the gaps we close.)
- A pre-commit hook in `.githooks/`, enabled by `post-start`; `govulncheck` and `gitleaks` in CI.
- The justfile is the entry point for everything; `ci-*` recipes are what CI calls — same recipe names as `mentor-api` (`fmt`, `fmt-check`, `build`, `test`, `mocks`, `ci-lint`, `ci-test`, `ci-mocks-check`).
- Multi-stage Dockerfile, minimal non-root runtime image pinned by digest, `CGO_ENABLED=0`, `-trimpath`.

## Reference copies

`prototype/` is a local, gitignored copy of the prototype repository. It serves as reference material for porting: catalogs and examples in `data/`, `schemas/`, `prompts/`, `src/taskgen/`, `tests/` (including `tests/example_checks/`) and `docs/architecture/`.

- Its rules file is renamed to `prototype/CLAUDE.prototype.md`; its rules do not apply here.
- Search tools that respect `.gitignore` (ripgrep) skip it. Use explicit paths (`Read`, `ls`, `grep -r prototype/...`).
- Never edit it and never commit anything from it except through a RUN.md task that ports content into the product.
- T65 deletes it once everything needed has been ported.

`reference/mentor-api/` is a local, gitignored copy of the platform's reference Go service, kept for one reason: to check our conventions against a working implementation instead of writing them from memory. Read-only — never edit it, never copy code from it into the product, only conventions. The same ripgrep caveat applies: use explicit paths.

## Environment

Open the repository in VS Code and choose "Reopen in Container" (`.devcontainer/`).

The host needs Docker and VS Code and nothing else: no builder plugin, no daemon settings, nothing to edit as root. The image is written for any builder, and the network the container runs on is created before it starts by `.devcontainer/host-network.sh`, with the MTU of the host's own default route — a tunnel on the host would otherwise cut every large download inside the container short with an error that never mentions the network. `post-start` mirrors that same value onto the docker running inside the container.

Inside the container:

- **Pinned versions.** No separate versions file: each tool's exact version is an `ARG` default in `.devcontainer/Dockerfile`, repeated as a literal in `.devcontainer/devcontainer.json` where needed.
  - The Dockerfile has two stages. `toolchain` installs Go, Node.js, just, golangci-lint and mockery; `devcontainer` adds gopls, dlv, cloudflared, gh, the Docker Compose plugin, Terraform, gcloud and Claude Code on top, and is the stage the container is built from. gcloud is installed on x86_64 only, because its archive for arm carries no Python interpreter; there the same commands come from Google's own container image. The full checks run in the `toolchain` stage, published to GHCR and pinned by digest. Everything is installed straight from the Dockerfile, and downloads refuse a protocol downgrade. What is verified is verified by the source it comes from: the Go checksum database for `go install`, the registry hash for the npm package; the tarballs fetched with `curl` carry no checksum of their own. Of the npm packages only the Claude Code CLI may run its own install step, named explicitly, because that step is what puts its native binary in place.
  - Docker (docker-in-docker, Moby engine and buildx) comes from a devcontainer feature pinned in `devcontainer.json` and `devcontainer-lock.json`.
  - To change a version: edit the literal everywhere it appears (Dockerfile, `devcontainer.json` for Docker or the Claude Code extension), rebuild the container. The Claude Code CLI and its editor extension are one version in two files, so `just claude-update` raises both at once — with a version, or to the newest published one — and prints what it changed.
- **Container marker.** `MATHTRAIL_DEVCONTAINER=1` is set only inside the container. A session that does not see it is on the host and must stop (see "Devcontainer only").
- **Claude Code.** Its config lives in the `mathtrail-standalone-claude` volume, so the login survives rebuilds. Auto-update is off; the version is pinned.
- **`gh`.** What a change looks like on github.com is otherwise invisible from in here, and two parts of it are asked for by name: the alerts of the code scan, which a task closes before the next one starts, and whether the delivery that carried the change ran. `gh auth login` is answered once — the credentials live in the `mathtrail-standalone-gh` volume, like the other two logins.
- **Just recipes.** `just --list` shows all of them. The everyday ones are `just fmt`, `just test`, `just lint`, `just build` and `just run`; the two that must be green before a task is done are **`just ci-lint`** and **`just ci-test`** (race detector and coverage), and they are what CI runs.
- **Git hooks.** `.githooks/pre-commit` is enabled by `post-start` through `core.hooksPath`. It runs formatting, a build and the tests — enough to catch what is embarrassing, while the linter, the race detector, `govulncheck` and `gitleaks` wait for CI. Where no Go toolchain is reachable — a Git client outside the container, for instance — it says so and skips those checks instead of blocking the commit: CI runs them again and is what gates a merge.
- **The runtime image.** `just docker-build` builds it and `just docker-run` starts it on port 8080. Both base images are pinned by tag and digest in the `Dockerfile`: a Go builder and a distroless static runtime that has no shell and runs as a non-root user.
- **Anything that is not Go runs in a container too.** `just golden` exports the prototype's vectors using the pinned `uv` image and the prototype's own PostgreSQL compose file, so no Python and no database is ever installed into the devcontainer (`testdata/golden/export/README.md`).
- **Build context.** The devcontainer image is built with the repository root as context; `.dockerignore` keeps `.git`, `.env`, `prototype/`, `reference/`, `docs/`, `testdata/`, `research/` and `node_modules` out of it and out of the runtime image.
