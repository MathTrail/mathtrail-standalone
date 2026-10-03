# Models through their command-line clients: what is available

Task S07 of [RUN.md](../RUN.md): before any protocol is written, know which models the research can use for free, how far, and how well isolated.

**Status (2026-09-25): done in the scope the author cut it to (Q57).** The paper comes first. This file keeps what was prepared and found; S41 takes up the rest before its first run.
- No model has been called.
- The trial calls, the logins to Codex CLI and Gemini CLI, and the verdict on each subscription's terms are the author's (K01).
- What only a run can show is marked *to confirm by the trial*.
- The work here passed two rounds of review.

## What is installed

| Client | Version | Where | Signed in |
|---|---|---|---|
| Claude Code | 2.1.280 | the devcontainer, pinned in its Dockerfile | yes, the author's subscription |
| Codex CLI (`@openai/codex`) | 0.157.0, Apache-2.0 | [containers/cli](../containers/cli/Dockerfile): Node 24.21.0 pinned by digest, packages from a lockfile, no install steps run | no |
| Gemini CLI (`@google/gemini-cli`) | 0.61.0, Apache-2.0 | the same image | no |

The two other clients live in their own image rather than in the devcontainer, whose image the product owns (Q55). Their logins live in the named volumes `mathtrail-research-codex` and `mathtrail-research-gemini`, never in the image.

## Isolation

A trial call must answer from the model alone. It must not read the repository — `content/` holds the reference tasks and their solvers — and must not run code to check its own answer. Each client gets the same four protections:
- an empty working directory;
- no tools;
- no project instructions;
- no session kept.

| Client | How | Source, read 2026-09-25 |
|---|---|---|
| Claude Code | `claude -p --output-format json --model M --tools "" --strict-mcp-config --setting-sources project --disable-slash-commands --no-session-persistence --max-turns 1`, run in an empty temporary directory. `--tools ""` disables every tool, and settings are read only from that empty directory, so no user hook or permission applies. | `claude --help` of 2.1.280; [CLI reference](https://code.claude.com/docs/en/cli-reference) |
| Codex CLI | `codex exec --ephemeral --ignore-user-config --skip-git-repo-check --json --sandbox read-only -m M -c approval_policy=never -c web_search=disabled -c project_doc_max_bytes=0 --disable F… -`, prompt on standard input, in the image's empty `/work`. `--disable` turns off every feature that adds a tool: `shell_tool`, `view_image`, `goals`, `multi_agent`, `apps`, `plugins`, `remote_plugin`, `tool_suggest`, `skill_search`, `code_mode_host`, `browser_use`, `browser_use_external`, `computer_use`, `in_app_browser`, `image_generation`, `sleep_tool`, `workspace_dependencies`, `hooks`, `shell_snapshot`, `skill_mcp_dependency_install`. Web search is off, the sandbox is read-only, `AGENTS.md` is not read, and no session file is written. What still reaches the model is in the next table. | [non-interactive mode](https://learn.chatgpt.com/docs/non-interactive-mode), [configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference); `codex exec --help` and `codex features list` of 0.157.0 |
| Gemini CLI | `gemini --skip-trust --output-format stream-json -m M -p PROMPT`, in the image's empty `/work`, with its temporary directory in memory (`--tmpfs /home/node/.gemini/tmp:uid=1000,gid=1000`: owned by the user the CLI runs as, or Gemini CLI cannot write there). A system policy in the image denies every tool (`toolName = "*"`, `decision = "deny"`, `priority = 999`), and a tool denied that way is left out of what the model is offered. | [headless mode](https://github.com/google-gemini/gemini-cli/blob/v0.61.0/docs/cli/headless.md), [policy engine](https://github.com/google-gemini/gemini-cli/blob/v0.61.0/docs/reference/policy-engine.md) of v0.61.0 |

The third trial call tests the isolation directly: it asks the model to print a file that holds a token made for that run. It passes only if the token is not in the reply.

**Checked without calling a model.** [tools/isolation](../tools/isolation/main.go) points a client at a local endpoint that records every request and refuses it, then lists the tools each request offered the model. It needs no login and spends nothing. The endpoint listens on the loopback interface, so it runs beside the client — for Codex and Gemini inside the same container, started with `--network none`, a made-up key and an empty directory for the client's home instead of the login volume. Run on 2026-09-25:

| Client | What the model is offered | What else it keeps |
|---|---|---|
| Gemini CLI 0.61.0 | **nothing**, once the deny policy is valid. At priority 1000 Gemini CLI rejected the file ("priority must be <= 999"), and the model was offered eight tools, among them `read_file`, `list_directory`, `grep_search` and `google_web_search`. Fixed to 999. Headless runs also refuse an untrusted directory, so the call needs `--skip-trust`. | Each session's transcript goes to `~/.gemini/tmp/<project>/chats/`, which no setting turns off. Mounted as memory (`--tmpfs`), it goes with the container. |
| Codex CLI 0.157.0 | **Current models (`gpt-6-*`, `gpt-5.6-*`):** a JavaScript tool (`exec`, a V8 isolate with no files and no network), its `wait`, `request_user_input` and the sub-agent tools — with every documented feature off. The model catalog marks them `code_mode_only`, and no setting changes that. **Legacy `gpt-5.5`:** `request_user_input` and `apply_patch`, with the same features off. | With `--ephemeral` no session file. Its home still gets state databases, the bundled skills and a snapshot of the shell's environment, unless `shell_snapshot` is off. |
| Claude Code 2.1.280 | *not checked yet*: the check runs `claude` with a temporary configuration directory, outside an agent session, so the author runs it. | `--no-session-persistence` |

So a Codex call cannot read the repository or reach the network, but it can compute. Whether E-A2 accepts that is settled in S41 (Q56).

What the login volume holds reaches every call. Gemini CLI reads `~/.gemini/settings.json` and a global `~/.gemini/GEMINI.md` from it, and has no switch like Codex's `--ignore-user-config`. So the volume holds the login and nothing else; S41 checks that before its first run.

To confirm by the trial:
- that `--setting-sources project` loads nothing in an empty directory;
- that the backends accept what the offline check saw.

## Which model answered

A client can switch models silently — for example when a limit runs out. The model that actually answered is therefore recorded for every call.
- **Claude Code:** the JSON result of `-p` is not described in the reference read. *To confirm by the trial* which field names the model (the result is expected to break usage down by model).
- **Codex CLI:** `--json` emits `thread.started`, `turn.started`, `turn.completed`, `turn.failed`, `item.*` and `error` events. The documentation does not say whether any names the model. *To confirm by the trial*, and if none does, from the session file Codex writes.
- **Gemini CLI:** the `result` event of `stream-json` carries "per-model token usage breakdowns". On a free Google account, "model requests will be made across the Gemini model family as determined by Gemini CLI", so the model can change between calls.

## Terms of use

The author decides for each subscription whether scripted research runs are allowed (K01). What the terms say:

- **Claude** ([Consumer Terms](https://www.anthropic.com/legal/consumer-terms), effective 8 October 2025, read 2026-09-25):
  - Section 3 forbids accessing the Services "through automated or non-human means, whether through a bot, script, or otherwise", "except when you are accessing our Services via an Anthropic API Key or where we otherwise explicitly permit it".
  - Anthropic's help pages describe `claude -p` on a Pro or Max plan as a supported use that draws on the plan ([Claude Code with a Pro or Max plan](https://support.claude.com/en/articles/11145838-use-claude-code-with-your-pro-or-max-plan); [the Agent SDK with a Claude plan](https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan), dated 16 June 2026). That second page announces that such use draws on a monthly credit, then that the change is paused.
  - The same terms forbid using the Services to develop or train competing models, which the research does not do.
- **ChatGPT and Codex** ([Terms of Use](https://openai.com/policies/terms-of-use/), effective 1 January 2026; read through the Internet Archive's copy of 2 September 2026, because the page refuses automated readers):
  - they forbid to "automatically or programmatically extract data or Output" and to "use Output to develop models that compete with OpenAI";
  - the user owns the Output;
  - Codex's own documentation offers `codex exec` "from scripts and CI" and recommends an API key for automation, calling a ChatGPT login there the advanced option.
- **Gemini** ([terms by sign-in method](https://github.com/google-gemini/gemini-cli/blob/v0.61.0/docs/resources/tos-privacy.md), v0.61.0):
  - a Google account with Google AI Pro or Ultra falls under the Google Terms of Service and the Google One Additional Terms;
  - a free Google account falls under the Google Terms of Service and the privacy notice of Gemini Code Assist for individuals;
  - the page warns that reaching the services behind Gemini CLI "using third-party software, tools, or services" breaks the terms; the research uses the CLI itself.

Nothing here is legal advice.

## Limits

What the documentation says; what a subscription actually allows is measured by the trial and by the pilot (S42, S52).
- **Claude:** usage resets every five hours, and Max plans also have a weekly limit ([help](https://support.claude.com/en/articles/11145838-use-claude-code-with-your-pro-or-max-plan)). Whether `claude -p` draws on the plan's limits or on a separate credit is unsettled (see the terms above).
- **Gemini CLI** ([quotas](https://github.com/google-gemini/gemini-cli/blob/v0.61.0/docs/resources/quota-and-pricing.md), v0.61.0), in model requests per user per day:
  - free Google account: 1,000;
  - Google AI Pro: 1,500;
  - Google AI Ultra: 2,000;
  - free Gemini API key: 250, Flash models only.
- **Codex:** the limits of ChatGPT plans were not found in the pages read.

## Open models on this machine

What the devcontainer sees:
- AMD Ryzen AI 9 365, 20 threads, AVX-512;
- 93 GiB of memory, about 68 GiB free;
- an integrated Radeon 880M through `/dev/dri`, and no NVIDIA GPU;
- 580 GB of free disk.

That runs quantised models of up to about 30 billion parameters on the processor. The integrated GPU might speed them up through Vulkan. Both speeds are *to measure*.

| Model | Licence (Hugging Face, 2026-09-25) | Why a candidate |
|---|---|---|
| `openai/gpt-oss-20b` | Apache-2.0 | 21B parameters with 3.6B active; a strong solver for its size |
| `Qwen/Qwen3-30B-A3B`, `Qwen/Qwen3-14B` | Apache-2.0 | another family; a mixture of experts that runs on a processor |
| `mistralai/Mistral-Small-3.2-24B-Instruct-2506` | Apache-2.0 | a third family |
| `deepseek-ai/DeepSeek-R1-Distill-Qwen-14B` | MIT | a reasoning model |
| `google/gemma-3-27b-it` | the Gemma licence, gated | not MIT or Apache: the author decides (K01) |
| `meta-llama/Llama-3.3-70B-Instruct` | the Llama 3.3 licence, gated | not MIT or Apache, and too large for this memory at useful speed |

The runtime would be llama.cpp in its own image, pinned by digest, with the weights pinned by file hash. Nothing is downloaded or run until the author confirms which models the research may use.

## How the trial runs

```sh
just research cli-image                  # build the clients' image (no model is called)
just research cli-shell codex            # sign in once: codex login
just research cli-shell gemini           # sign in once: gemini, then choose a sign-in method
just research cli-trial claude <model>   # three isolated calls; spends the subscription's limits
just research cli-trial codex <model>
just research cli-trial gemini <model>
just research cli-trial codex <model> --dry-run   # print the three commands; calls nothing
```

[cli-trial.sh](cli-trial.sh) keeps every call's raw output and appends a line to `cli-trials/log.tsv`: when, which client and version, the model asked for, the seconds taken, of which how many went to starting a container, the exit status, and for the third call whether the isolation held — `yes` only if the call succeeded and the token is not in its output, `unknown` if it failed first. Any argument other than `--dry-run` stops the script before it calls anything. The raw outputs are read afterwards to find where each client names the model that answered; S41 builds the full harness on what they show.

## Remarks

- **Signing in inside a container.** The clients sign in through a browser, and how that works inside a container is not yet tried: Codex may need a device code, and Gemini a manual one. The first login settles it, and this file is updated then.
- **Licences of outputs.** The terms of each subscription govern what may be published from the models' outputs. The research publishes tasks the models wrote (S65), which the terms read here allow: the user owns the output. It never trains a model on them.
