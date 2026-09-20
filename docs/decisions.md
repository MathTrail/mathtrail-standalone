# Implementation decision log

> Why the service is built the way it is. Every entry says what was decided, why, what was rejected, and where it lives.
>
> **The rule:** do not re-propose something rejected here without new evidence. If a decision has to change, ask the author and update this file.
>
> Product decisions live in [PRODUCT-V1.md, section 12.1](../PRODUCT-V1.md#121-settled) under the numbers О-…; this file holds implementation decisions, R01… The first entries were collected in T05 from the phase 0 spikes: [T03 — protocol and widget](live/01-spike-protocol.md), [T04 — the OAuth stub](live/02-spike-oauth.md). The prototype's decisions D01–D45 are in [docs/prototype/decisions.md](prototype/decisions.md) and are cited as "D43 of the prototype".

## Protocol and SDK

**R01. Go SDK — v1.8.0.** PRODUCT 7 named v1.7.0; as of 2026-09-16 the current release is v1.8.0 (14.09.2026), a patch release with no protocol changes: the same set of supported versions, with 2026-07-28 still the newest. It adds limits on SSE event size, body size and JSON depth, the `ServerOptions.SupportedProtocolVersions` setting, and an explicit `-32022` code in place of a bare 400. Spike T03 was built and verified on it. — *Why:* strictly newer and fully compatible; there is no reason to stay on the previous patch. — *Where it lives:* `go.mod` (T17), PRODUCT 7.

**R02. The server does not narrow the set of protocol versions.** We work only over 2026-07-28 (О-20, О-28), but `SupportedProtocolVersions` is left at its default: the SDK already prefers the newest version when negotiating. — *Why:* in strict mode Claude fails to add the connector. Before the main session the host sends two short legacy `initialize` requests with an empty version, and rejecting them by version breaks the "Connect" step (status 400); in the default mode the same runs go entirely over 2026-07-28 — not one real request negotiated an older version. — *Rejected:* a strict "2026-07-28 only" transport; a stateless 2025-11-25 fallback path (О-28). — *Where it lives:* T41; to be re-checked in ChatGPT in T63.

## Sign-in and permissions

**R03. Client registration is CIMD, with DCR kept as a fallback.** The authorization server metadata declares `client_id_metadata_document_supported: true`, and the server fetches the document itself over an HTTPS `client_id`; `/register` (RFC 7591) is still supported for backwards compatibility. — *Why:* as of 2026-07-28 DCR is formally deprecated, and Claude showed up live with CIMD (`client_id: https://claude.ai/oauth/mcp-oauth-client-metadata`) without ever falling back to DCR; the ChatGPT documentation also recommends CIMD, though that has not been verified live. — *Where it lives:* T07, T47–T49.

**R04. Insufficient permissions on a tool return HTTP 403 with a `WWW-Authenticate` header.** The `_meta["mcp/www_authenticate"]` field in the call result is sent in addition, for ChatGPT's sake, but it cannot be the only mechanism. — *Why:* the live run in T04: on a 403 Claude went to `/authorize` for the wider scope and retried the failed call by itself, with no human involved; a tool that answered only through `_meta` looked to it like an ordinary error. — *Consequence:* the model never sees the step-up — in its context there is a single successful call — so its account of the authorization state is unreliable and only the server log can be trusted. The same rule applies to every acceptance run. — *Where it lives:* T47–T49.

## Tools and environment

**R05. There is no separate tool-versions file.** Each exact version is pinned where the tool is installed: an `ARG` in `.devcontainer/Dockerfile`, a literal in `devcontainer.json` and `devcontainer-lock.json`, a module version in `go.mod`, an exact version plus a lockfile for npm, a commit SHA for GitHub Actions. — *Why:* one source of truth per tool; a separate versions file would have to be kept in sync with those same places by hand, and the build reads the versions from them anyway. — *Rejected:* a shared versions file; checksum verification beyond what `go install` already does. — *Where it lives:* CLAUDE.md ("Exact versions only"), T01.

## Widget build

**R06. The `@modelcontextprotocol/ext-apps` package, version 2.0.0; the library is the reference, not the text of the specification.** Where the MCP Apps 2026-01-26 specification disagrees with the library — in `ui/message` and `ui/update-model-context` its `content` field is always an array of blocks, not a single object — we follow the library. — *Why:* its code is the actual runtime the hosts run; the discrepancy was found in T03 while building the widget. — *Where it lives:* T42, `web/package.json`.

## Instructions for the model

**R07. Text written for the model never relies on an exact tool name.** Claude prefixes tool names with the connector name (`MathTrail:show_widget`). The instructions describe tools by purpose and by their unprefixed name, without requiring an exact match; text shown to the child and the adult mentions no tool names at all. — *Why:* the prefix is added by the host and differs from host to host, so an instruction tied to an exact name breaks when the host changes. Found during the live run in T03. — *Where it lives:* T36; to be re-checked in ChatGPT in T63.

## Content for the model

**R08. Solver templates live in `content/solvers/`, one or two per topic.** A template is a short program following the solver contract, with marked slots for the numbers of the task, generalized from a solver already written for a reference task; the id of that reference task is in a comment. The package builder includes only the templates of the chosen topic and stays within the size budget, and the instructions call a template a sample one may depart from. — *Why:* the solver is the only programmatic guarantee that there is exactly one correct option (PRODUCT 4.3), and an error in it costs the child a whole attempt; a worked-through example of the typical enumeration removes the most common cause of `solver_error` for a weaker model. — *Rejected:* a mandatory skeleton, which would have made tasks bend towards whatever the template can do; and having no templates at all (О-43). The risk of sameness is countered by the rule rotating topics (T27) and by the duplicate check (T32). — *Where it lives:* T36a, T36.

**R09. Text drawing frames live in `content/drawings/`.** A frame is generic, carries no numbers from any particular task, and comes with a structural description of its objects and their relations; the model fills in the numbers and the labels. Every frame passes the format check at the default limits itself — otherwise it is not a frame. — *Why:* the drawing stops depending on what the model happens to remember about box-drawing characters, and the format check (О-11а) passes on the first attempt more often. — *Rejected:* drawing rules in the instructions with no ready-made frames (О-44); free-hand drawings remain allowed and are checked against the same limits. — *Where it lives:* T36b, T36; once the limits are calibrated, the frames are re-checked in T58.

## Task checks

**R10. The explanation behind a wrong option is checked deterministically, with no judgement of meaning.** Four conditions: the texts of the four distractors are pairwise distinct; a text neither equals nor is a prefix of the solution or the hint; a text is not the verbatim description of the trap from the catalog; and a text is not shorter than the minimum for its writing system (words, or characters for CJK and Thai). A rejection gets its own code and names the option that needs fixing. — *Why:* "a wrong answer is a diagnosis" is one of the three pillars of the product (PRODUCT 1), and in text mode the explanation is the main thing the child gets. — *Rejected:* instructions and reference tasks alone with no programmatic check; and any heuristic for "meaningfulness" — with a median generation time of 69 seconds, extra strictness costs the child another attempt (О-46). — *Where it lives:* T32a; the rejection code is defined in T12.

## Widgets and showing progress

**R11. There is no "advice for the adult" block in v1.** The task card stays a child's card: no hidden block of leading questions, no widget-only tool for it, no questions in the sealed block. — *Why:* the author's decision (О-45) — the adult guides the child themselves and gets leading questions from the model in the chat. — *Consequence:* task T55a is deleted and the wording of PRODUCT section 1 is rewritten; if the idea comes back, it goes through an open question in 12.2 again. — *Where it lives:* PRODUCT 1, 2, 4.2.

**R12. A rank above the rating number.** Five steps, with boundaries computed from the rating corridor and names taken from the locale dictionaries. The rank is not stored in the profile: it is a matter of presentation, not data. — *Why:* "1500" means nothing to a seven-year-old, and the scale itself does not change (PRODUCT 4.5). — *Where it lives:* T57; the keys in T54, the translations in T59 (О-48).

**R13. The rhythm of practice is not shown.** No counter of tasks solved today in the profile, no counting them from the tail of the history, no streaks, no penalties for skipping and no reminders. — *Why:* the author's decision (О-49) — progress stays about knowledge, not about volume; pushing a 7–10-year-old towards regularity is out of place (PRODUCT 10). — *Where it lives:* T09 and T57 stay unchanged.

## The boundaries of v1

**R14. There are no printable task sheets in v1.** There will be no tool that hands the adult 3–5 checked tasks as a single text. — *Why:* the profile stores fingerprints of past tasks, not their texts (О-40), so a sheet could only be assembled from tasks generated on the spot: 3–5 generations in a row, which is the message limit of a free tier, a daily-limit calculation of its own, and in effect a return to the cancelled task buffer (О-6). — *Where it lives:* PRODUCT 2 ("not included"); no separate task is added in phase 7 (О-47).
