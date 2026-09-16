# mathtrail-standalone

Free, open-source standalone MathTrail app for LLM applications. It generates and validates adaptive maths olympiad-style tasks for grades 1–6, with interactive MCP widgets.

> **Status:** early development. Nothing is usable yet; see the plan in [RUN.md](RUN.md).

## How it works

MathTrail connects to Claude or ChatGPT as an MCP app. A parent (or tutor) adds it to their own chat, and the child solves tasks there:

1. The service picks a topic and difficulty for the child and gives the chat's model a brief, reference examples and formats.
2. The chat's model writes the task.
3. The service checks the task: structure, a solver program that brute-forces the answer options, readability, near-duplicates and text drawings.
4. The child sees the task in a widget, answers with a button, and gets an explanation of the mistake. Ratings are updated.

Design points:

- **No LLM calls of its own.** Only the chat's model writes text.
- **Stateless.** The service stores nothing between requests. The child's profile is a JSON file in the parent's own Google Drive, and the answer to the current task is encrypted there.
- **Self-contained.** It runs as a single Go binary on Google Cloud Run: no database, no queues, no other services.

## Development

All development happens in the devcontainer; nothing but Docker and VS Code is needed on the host.

1. Install Docker and VS Code with the Dev Containers extension.
2. Open the repository and choose "Reopen in Container". The first build downloads the pinned toolchain and takes a few minutes.
3. `just --list` shows the available recipes.

## Documents

Project documents are in Russian; code, comments and model instructions are in English.

- [PRODUCT-V1.md](PRODUCT-V1.md) — product requirements.
- [RUN.md](RUN.md) — implementation plan, one task at a time.
- [docs/prototype/](docs/prototype/README.md) — documents from the prototype this product grew out of.
- [CLAUDE.md](CLAUDE.md) — working rules for Claude Code in this repository.

## License and trademark

The code and the bundled content (catalogs, reference tasks, model instructions) are released under the [MIT License](LICENSE).

The MIT License does not grant any rights to the name "MathTrail" or its logos. If you fork this project and run it publicly, please give your version a different name.
