# The deployed service under load (T64.3), the message limits of free Claude (T64.4), and their fixes (T64.5)

**T64.3: closed.** 2026-10-04. The `paces` and `ceiling` scenarios ran clean against `https://mcp.mathtrail.app`. The greedy account and the greedy address were refused by their paces in words anyone can follow, and the accounts beside them were never refused. The day's ceiling refused the sixth failed request and nothing else. The platform started three instances. The run cost 0.094 % of a month's free Cloud Run processor time, which is $0. Without Drive, the tools answer in a few milliseconds. Claude's model writes a task in under 90 seconds in 29 cases out of 30. The log holds nothing that looks like personal data.

The run found that the share of traces was decided for a run of requests rather than for each request: in the hour of the run, 0.7 % of the traces were kept instead of 10 %. Fixed within the task (R173), and confirmed once deployed: 0.104 in the hour of T64.4's lesson. The numbers of the limits and of the telemetry stay as they were, now measured (R174). The check from a phone was not done, and one finding is open — below.

**T64.4: closed.** 2026-10-04. One five-hour session of free Claude, Sonnet 5.5 at medium effort, held 14 tasks before its limit, where 3–6 were expected: a long lesson by SPEC 10.2, written in 26 minutes. A task costs one message; its wait and its answer on the card cost none. At the limit the open task can still be answered on its card, but no new task can be asked for, and the progress could not be opened. Four findings go to T64.5.

**T64.5: fixed in code, the live check to come.** 2026-10-04. The words that come with an accepted task said "record the answer with submit_answer", and the model asked the parent for the letter: they now say the card records the answer. The waiting card now speaks the language the task was asked in from its first word (R190). Cloud Monitoring's refusals were not two instances writing one series: all eight came from builds before T62.1, and none has come since. What Claude does at the limit with "Another task", and with the cards of a chat opened again, is for the live check below.

> The repository is public, so the report carries behaviour and numbers rather than identifiers: accounts are named by their role, and instances, traces and revisions are not written out beyond their numbers.

## What this is

Criteria 11.5–11.8 of PRODUCT on the deployed service, and the response-time goals of PRODUCT 6.

- **Service.** `https://mcp.mathtrail.app`, version `v0.2.3`, commit `52d0a1c`: that is what `/health` answers.
- **Accounts.** The author signed two in with `just load-signin`:
  - `parent` — the author's personal account, which only reads its profile in the run;
  - `load` — the company account, whose profile became the load's: the pseudonym Otter, grade 2, the load's tasks.
- **Several instances.** For the run the service's concurrency was lowered from 80 to 2: with it, three instances start from a single source of requests.
- **Roles.** The author signed in. The run, putting the service back, reading the log and the metrics, and this report are Claude's. The commands that change the service were allowed by the author with a narrow rule in `.claude/settings.local.json` for the length of the run.
- **Cost.** Estimated before the start: $0, up to 3 % of a month's free Cloud Run. Actual: 0.094 % (below).

## How the run went

Times are UTC, 2026-10-04.

1. **03:03–03:06. Two accounts signed in.**
   - VS Code did not forward from the container the port the tool picks, and the browser got "connection refused". The author forwarded port 8976 in the Ports panel, and the sign-in went through with `-callback-port 8976`.
   - The service refused the company account's first sign-in: `access_denied`, and in the log an `auth_callback` with the reason `no_drive`. The Drive box on Google's consent screen was left unticked, and the refusal is the behaviour as designed. The second sign-in, with the box ticked, went through.
2. **03:09. The state before the run.** Concurrency 80, up to 3 instances, 1 vCPU, 1 GiB, revision 00075.
3. **Before 03:18. Concurrency 2.** `gcloud run services update mathtrail … --concurrency=2`, revision 00076.
4. **03:18:05.** `just load paces -url https://mcp.mathtrail.app -accounts parent,load -rate 6`, one minute.
5. **03:21:18.** `just load ceiling -url https://mcp.mathtrail.app -accounts load,parent`, 1 min 22 s.
6. **03:23:38. Concurrency 80 again.** Revision 00077. Checked:
   - concurrency 80, up to 3 instances, 1 vCPU, 1 GiB, and `/health` answers `v0.2.3`, `52d0a1c`;
   - revision 00077 matches 00075 in full: image, resources, scaling, variables and secrets. The only differences are the revision's creator (the company account in place of the deployment's service account), the operation id, the random `nonce`, and the route label on the revision that now serves the traffic.

   That is the reconciliation with Terraform. The service is again what the pipeline deployed from `infra/terraform`, and the next delivery sees what it saw before the run. `terraform plan` was not run locally: it needs the billing account, which lives only in GitHub's secrets.
7. **After 03:26.** `just report 1h`, `just report 7d`, `just usage 1h`, `just usage 7d`, and narrow reads of the log for the traces.

## 11.5: the greedy one is refused, the others keep working

All five groups got the tool's verdict "clean — nothing the service should never do": no status of 400 or above, no protocol error, no answer to another's call.

**`paces`**, one minute:

| Who | Doing what | Calls | Let through | Refused by the pace | p50 / p95 of those let through |
|---|---|---:|---:|---:|---|
| `parent`, the greedy account | reads its profile 6 times a second | 359 | 178 | 181 | 600 ms / 1.05 s |
| `load`, a child beside it | a lesson at a live pace, 2 tasks | 15 | 15 | 0 | 955 ms / 3.3 s |
| the tool's address, greedy | the sign-in's document 6 times a second | 359 | 72 | 287, as `429` | 65 ms / 464 ms |

A refusal by the pace comes back in 64 ms (p50). It is one sentence for the model, marked as an error, with no number in it: "MathTrail received too many calls in a short time, so this call was not made. Wait a moment, then make the same call again."

The greedy account was let through 178 times in a minute, at a pace of 60 a minute with a burst of 20. That is the sum of three instances, each counting in its own memory, as SPEC 10.1 has it: "with N instances the effective rate is up to N times looser". The greedy address was let through 72 times at a pace of 20 with a burst of 6: three instances as well.

**`ceiling`**, 1 min 22 s:

| Who | Calls | Outcome |
|---|---:|---|
| `load`, failing its generations | 28 | 12 went through; 10 attempts refused as `solver_disagrees`; 5 requests ran out of their three attempts; the sixth request of the day got `limit_reached` |
| `parent`, reading beside it | 14 | all 14 went through |

The day's ceiling of five failed generations held at exactly the sixth request. The `load` account gets no new tasks until the end of 2026-10-04, UTC.

## Cloud Run's consumption against the free tier

As Cloud Monitoring counts it (`just usage`):

| Window | vCPU-seconds | GiB-seconds | Requests | Instance-seconds billed | Most instances in a minute |
|---|---:|---:|---:|---:|---:|
| The hour of the run, 02:27–03:27 | 169.0 (0.094 %) | 322.9 (0.090 %) | 830 (0.042 %) | 323.2 | 3 |
| The week, 09-27 03:47 – 10-04 03:47 | 622.3 (0.346 %) | 815.6 (0.227 %) | 1828 (0.091 %) | 848.9 | 3 |

The percentages are of a month's free tier: 180,000 vCPU-seconds, 360,000 GiB-seconds and 2 million requests. The vCPU-seconds run out first.

**How many tasks a month the free tier holds.** On the constrained container (`docs/load.md`) a task cost 0.36 vCPU-seconds, and the free tier held 500,000 tasks a month. There was no Drive there.

On the deployed service:
- the load's lesson — 10.15 vCPU-seconds per accepted task, about 17,700 tasks a month;
- the week of real traffic, the hour of the load left out — 453 vCPU-seconds for 30 accepted tasks, about 15 per task with all the cards' questions, the sign-ins and everybody else's requests. That is about 12,000 tasks a month.

The difference is the time spent waiting on Drive. With `cpu_idle` the platform bills an instance for as long as it holds at least one request, a request waiting on Drive included. Once there are more requests, they overlap on one instance, and a task costs less. Today's figure is the upper bound of a quiet service.

Logging, Trace and Monitoring were not measured on their own. The log of the hour of the run is 2085 lines, against 50 GiB a month free.

## Response time

### Tools without Drive: the goal is under a second (PRODUCT 6)

The service's time per call, in milliseconds; "without Drive" is the same time less the calls to Drive made in the same request (`just report`).

Real lessons in Claude over the week:

| Tool | Calls | With Drive, median / p95 | Without Drive, median / p95 |
|---|---:|---:|---:|
| `get_package` | 4 | 208 / 217 | 1 / 1 |
| `get_profile` | 4 | 999 / 1664 | 1 / 4 |
| `next_task` | 32 | 2174 / 3408 | 3 / 7 |
| `read_progress` | 26 | 638 / 823 | 2 / 4 |
| `read_task` | 37 | 200 / 236 | 1 / 2 |
| `save_profile` | 4 | 1798 / 3472 | 3 / 5 |
| `submit_answer` | 19 | 2227 / 4589 | 3 / 7 |
| `submit_task` | 34 | 1911 / 4107 | 13 / 33 |

Under load (the hour of the run, three instances, concurrency 2), without Drive: medians of 0–8 ms, and a p95 no higher than 13 ms. On the container without Drive (`docs/load.md`): a median of 4 ms, and 10 ms for the longest call. The goal is met with two orders of magnitude to spare.

All the rest of a call's time is Drive: medians from 0.2 to 2.2 s, and a p95 of up to 4.6 s. That is within the host's timeouts.

### Writing a task: the goal is 90 seconds

The seconds from the request `next_task` opened to the task's acceptance: the time in which the chat's model wrote the task and handed it in. The author's lessons in Claude over the week, with whatever model and level of reasoning the author had (in T62, Opus 5.5 with extended reasoning):

| Instructions version | Accepted | Median | 90th percentile | Longest |
|---|---:|---:|---:|---:|
| `7b5d171123e5` | 6 | 19 | 29 | 29 |
| `88b63e22129c` | 14 | 31 | 59 | 74 |
| `9c6aa546400d` | 6 | 23 | 69 | 69 |
| `bf40900be9b5` | 3 | 60 | 227 | 227 |
| `1e1cd63988d6` | 1 | 43 | 43 | 43 |

29 tasks of 30 came within 90 seconds, and one took 227 seconds. The free tier and its model are T64.4's to measure. The author's requirement from T62 (finding 7) is under 5 seconds, and that is a separate matter.

## 11.8: no personal data in the logs

`just report` holds every line of the service's log to three rules: an event from the table of decided events, only the fields decided for it, and nothing that looks like an email address.

- **The hour of the run:** 2085 lines, two sign-ins, a lesson and the refusals of the paces and of the day's ceiling among them. No breach.
- **The week:** 5065 lines since 2026-09-27 11:07. No breach.
- **Old lines.** The same report over the log's whole retention found:
  - 602 lines of `http_request` with a `path` field, all from 22 to 25 September, revisions 00002–00029, before R76 replaced the path with the route. None of them carries a query string: they are the paths of scanners, such as `/.env` and `/.git/HEAD`, and the service's own routes;
  - 11 lines of events not in the table, all older than 27 September.

  The report does not name such events, by design. Reading the lines themselves is a bulk read of the production log, which this session was not allowed. The lines leave with the log's retention.

No email address anywhere, in the new lines or the old.

## Traces: our share decides, but it decided by runs

**What was checked.** R49 and SPEC 12.5 say that a trace is kept by our share, 0.1, whatever the request arrives saying. The task was to confirm that with statistics and to choose the number.

**What the hour of the run showed.**
- Of 828 requests with a trace, 6 were kept: 0.007, against the 0.1 configured.
- Cloud Run itself sampled 41 requests of 830 (`traceSampled` in its request log). One of them was among our six. So the platform's flag does not decide, and our share does, as R49 meant.
- But the share came out 14 times smaller than configured. The cause is the trace identifiers Cloud Run's front end makes: it fills the right half once for a run of requests, and only the left half anew for each. The hour's 830 requests fell into 64 runs by the first 12 hex digits of the right half. 13 runs held 20–75 requests each, nearly all the load's requests among them; 51 held 1–6. How the front end makes its runs cannot be seen: they are not connections, since the greedy account had one connection for its 360 requests, and several runs.
- The standard `TraceIDRatioBased` reads exactly the right half. So a run was kept or dropped whole: 7 requests, in 4 runs.
- Read by the left half, the same requests would have kept 89, which is 0.107.
- The rest of the week without the hour of the run — the author's lessons and sign-ins — kept 109 traces of 941, which is 0.116. There the share landed near the configured one by chance, but there too the decision was made for a whole run at once.

**Fixed (R173).** The sampler now reads the whole identifier, stirred with MurmurHash3's finalizer: the right half is stirred and folded into the left, and the whole is stirred again. The rest is as it was: the request's flag does not count, and our spans follow their parent. A property test holds the share kept to the share configured, whichever half changes from one trace to the next. With the standard sampler it fails at the very first case.

**The number — 0.1 (R174).** A kept request carries from two to about ten spans: the server, the tool call, the calls to Drive, the task's checks. At a share of 0.1, the spans stay within Trace's free 2.5 million a month for as long as the requests stay within Cloud Run's free 2 million. At 0.2 they would not. Each kept request also waits for its own delivery.

**What to check after the rollout.** `just report 1d` on a build with R173: the share kept should come out near 0.1 over any window of a few hundred requests.

## The 200 ms flush deadline against real latencies

`telemetry_flush_failed` over the week — 36 lines:

| What | Lines | When |
|---|---:|---|
| Spans that missed 200 ms | 14 | 6 in the first 5 s of an instance's life (0.2–4.6 s), 8 on warm ones, from 18 s to 25 min |
| Measurements that missed 200 ms | 14 | on cold and warm instances alike |
| Measurements Cloud Monitoring refused, `400` | 8 | on warm instances |

3 more lines of `telemetry_failed`: the measurement reader's own timeout.

In the hour of the run, not one delivery of the six kept traces failed.

Over the week, 14 deliveries of spans failed beside 115 kept traces — about one in eight. The cold instances, which pay for the connection, had 6 of the 14. R49 expected the misses to come from them, but the warm ones had even more — 8.

**Decision (R174): 200 ms stays.** The deadline bounds a child's wait. Every request's time is in the log, whatever becomes of its trace. A sample that loses one trace in eight is still a sample. A longer deadline cannot be chosen blind: the log does not say how long a successful delivery took, so how many of the losses, say, 1 s would buy back is unknown.

## The decisions on the numbers (R174)

| Number | Was | Now | Why |
|---|---|---|---|
| An account's pace | 60 a minute, burst 20 | unchanged | The busiest minute of real lessons held 15 calls of one account. That is what a card asking about its task every 4 s gives. The greedy account hit its pace in the very first minute; the accounts beside it saw no refusal |
| An address's pace before sign-in | 20 a minute | unchanged | A sign-in is under ten requests; the greedy address got `429` |
| The instance's fuse | 200 a minute at each of its two doors | unchanged | The busiest minute of real lessons held 19 MCP messages on one instance. Under load an instance took 247 messages in a minute, most of them cut off by the account's pace before the instance counted them, and the tools' own time stayed within 13 ms. Two accounts cannot reach 200; on an ordinary day the fuse decides nothing |
| Failed generations a day | 5 | unchanged | The refusal came at exactly the sixth request, and the reading beside it was not touched |
| The share of traces | 0.1 | 0.1, read off the whole identifier (R173) | Above |
| The flush deadline | 200 ms | unchanged | Above |
| Alerts on metrics | an open question of R49 | none | A metric in an alerting policy costs $0.35 a month (R49) — a third of the $1 budget for one alert. The Billing budget (`infra/terraform/budget.tf`: 50 % and 100 % of the actual spend, 100 % of the forecast) alerts on the spend for free; `just report` and `just usage` show the behaviour, and at a few lessons a day that is enough |

At the other door, the sign-in's, the busiest minute of real traffic held 143 requests on one instance. Most likely these were scanners: on 22–23 September they asked for paths such as `/.env`, as the old lines with a path show. The report does not name the minute itself, and reading the lines for it is a bulk read of the log. Paths the service never declared are not counted by the paces (SPEC 10.1).

## Checklist 11.6: the task checks

| A bad task | The check | Covered by |
|---|---|---|
| A wrong answer | `solver_disagrees`: the solver found another option | `internal/domain/checks/review_test.go`: "a solver that finds another answer", "another answer is worked out"; "a self-check with another answer" |
| Two right options | `solver_disagrees`: "found more than one option that fits" | `TestEachWayASolverDisagreesIsToldApart`, "two options fit" |
| Unreadable text | `readability` | "a sentence too long for the level"; `readability_golden_test.go` against the prototype's vectors |
| A repeat | `near_duplicate` | "a copy of a reference task", "a repeat of a task the child has had"; `similarity_test.go` and `corpus_test.go` against the prototype's vectors |
| A drawing with a broken format | `drawing_format`, `drawing_mismatch` | "a drawing with a space at the end of a line", "a drawing without a label it declares"; `drawing_test.go`, `drawingmatch_test.go`, `drawing_property_test.go` |
| A good task is accepted | — | `TestAGoodTaskIsAccepted`; in `content/`, every reference task passes the checks and proves its answer with its solver, and its drawing passes the drawing checks |

The test set is carried over from the prototype: `testdata/golden/` is exported by `just golden` from `MathTrail/llm-taskgen-prototype` at commit `02638353482e`, and no Go test ever rewrites those files.

Live, over the week: real tasks from Claude were refused as `bad_structure` (2) and `drawing_mismatch` (1). The load's tasks were refused as `solver_disagrees` (15), and all 15 as `near_duplicate` as well.

`just ci-test` is green, this task's change included.

## Checklist 11.7: one image, no secrets, a licence, a copy of your own

| Item | How it was checked |
|---|---|
| One image | `Dockerfile`: the widget's build, the Go build, a distroless image with no shell, not root, everything by digest. The deployed service is this image |
| No external services but Google's sign-in and Drive | Besides them the service reaches Google Cloud Observability, only on a deployment and only when a project is named (`MATHTRAIL_TELEMETRY=auto`), and works without it; and the metadata documents of clients, which a client names itself at sign-in — part of OAuth |
| No secrets in the repository | `just ci-secrets` (gitleaks over the whole history): 84 commits, no leaks. The secrets live in Secret Manager and in environment variables |
| A licence | `LICENSE` — MIT; `THIRD_PARTY_LICENSES` is current: `just ci-licenses` is green |
| Instructions for a copy of your own | `docs/self-hosting.md` |

## What was not done, and what is open

- **The check from a phone was not done.** The author was to open the sign-in's document from a phone, on a mobile network, while the greedy address was being refused. Every request of the run came from one address, so the run cannot show that another address is served meanwhile. On a container, the `limits` scenario checks it, making its addresses up itself.
- **Cloud Monitoring refuses measurements (8 in a week).** The answer is `400 FAILED_PRECONDITION`: "One or more points were written more frequently than the maximum sampling period configured for the metric". In the case that was read, two points of one series were 5.5 s apart, while one instance sends its measurements at most once a minute. Most likely the series does not tell instances apart, and two of them write into it. The resource carries `faas.instance` but no `service.instance.id`. Before a fix, how OTLP metrics are laid out into series has to be checked against Google's documentation. The fix is T64.5's. **Answered in T64.5:** the series does tell instances apart, and all eight refusals came from builds before T62.1 (below).
- **The share of traces after the rollout — confirmed.** On `v0.2.4`, the first build with R173, the hour of T64.4's lesson kept 28 traces of 270 requests: 0.104, against the 0.1 configured. Two of the 28 deliveries ran out of the deadline, as R174 expects.
- **The company account's profile** was the load's until T64.4 deleted it and its lesson made a fresh one. To run `paces` with the account again on a clean profile, delete the profile file in Drive and then from the bin (`docs/load.md`).
- **`.claude/settings.local.json`**, with the permissions for this run, is for the author to delete: the run is over.

## The message limits of free Claude (T64.4)

Risk 9.4 of PRODUCT, the message limits of the free tiers eaten by the writing of tasks, measured in Claude: how many messages and turns a task takes, when the limit comes, and what the parent sees then.

- **Host and plan.** Claude on the web, the company account on the free plan, whose one custom connector is MathTrail. Claude's own interface is in English, and the lesson was in Russian.
- **Model.** Sonnet 5.5 at medium effort, as the free plan offers it. It thinks before its calls: "Thinking 11s" on a screenshot.
- **Service.** `v0.2.4`, commit `f09996c`.
- **Profile.** A fresh one. The load's profile was deleted from Drive first, and the lesson made a new one, grade 2, so its first five tasks were the trial series.
- **Roles.** The author held the lesson as the parent and the child, answering on the card, and took the screenshots. Claude read the service's log and wrote this section. The sheet of messages per task was not kept: the messages that asked for a task are counted from the log, one call of `next_task` for each.
- **Cost.** One session of the company account's free plan; no money.

### Expected, and found

Expected before the start: the limit after 3–6 tasks, 6–10 messages with the profile. A task adds about 25 KB to the conversation — 21 KB of them the package of `get_package`, 10 KB of that the guide, the same in every package — and every turn reads the whole conversation again. The decision log's own guess was three to five generations in a row (R14).

Found: 14 tasks. Times are UTC, 2026-10-04:

| Time | What happened |
|---|---|
| about 14:20 | The first message of a new chat; the session's five hours count from here |
| 14:24:20 – 14:25:27 | `get_profile`, then `save_profile`: the profile made in the chat |
| 14:25:40 | The first task asked for |
| 14:50 | After the 13th task, Claude's banner "You've used 90% of your session limit", which blocks nothing |
| 14:51:18 – 14:51:43 | The 14th task, asked for and accepted; after that turn, the limit |
| 14:52:01 | The 14th task's answer on the card, recorded |
| about 14:52 | "Another task" — the send fails |

The limit resets at 19:20 UTC, five hours after the session's first message.

### What a task costs

- **Messages:** one a task. Each of the 14 tasks was asked for by one message — the first in words, the rest by "Another task" and Enter — and each made one call of `next_task`. The profile and the questions asked in the chat came on top, and were not counted.
- **The model's turn:** `next_task`, `get_package` and `submit_task`, which was called 16 times for 14 tasks. 13 tasks were accepted at the first attempt; one was refused twice — for its drawing's format, then for a drawing that did not match its task — and accepted at the third.
- **The card, without the model:** `read_task` 132 times, about nine a task while its task was being written; `submit_answer` 14 times, every answer given on the card; `read_progress` 3 times.
- **The time to write a task:** a median of 31 s, a 90th percentile of 51 s, and 93 s for the longest: 13 tasks of 14 within the goal of 90 s.
- **The tools without Drive:** no more than 13 ms at the 95th percentile, as in T64.3.

### What the parent sees at the limit

1. At 90 %, a banner above the input: "You've used 90% of your session limit", with "Upgrade". It blocks nothing.
2. At the limit, once the turn under way has finished: a dialog, "Upgrade to keep chatting — You hit your 5-hour message limit. It resets at 2:20 PM, or you can upgrade for higher limits", and under the input, "You are out of free messages until 2:20 PM".
3. **What still works:** the answer on the open card. It was recorded at 14:52:01, after the limit: the card calls the service itself, without the model.
4. **What does not:** a new task. "Another task" put its words into the input under the warning "Use caution before running this prompt…", and the send failed: "Failed to send · Retry", and a toast, "You've hit your limit for Claude messages. Limits will reset at 2:20 PM". After that the chat showed no card, and the progress could not be opened.

### For risk 9.4

One session of free Claude holds one long lesson: 14 tasks, where SPEC 10.2 counts ten to fifteen to a long session, and below the service's own ceiling of twenty accepted tasks a day — so on the free plan it is Claude that ends a lesson, not the service. The session ran out 26 minutes after the first task, and the next one starts five hours after the first message of the last. The risk is real, but smaller than feared: a family gets a long lesson every five hours, not three tasks. The number holds for Sonnet 5.5 at medium effort, since a higher effort spends a session faster; and finding 1 below would have a parent spend two messages on every task. No question for PRODUCT 12.2 is needed.

### Findings

1. **The model asks for the answer's letter in the chat.** Under the first task's card, the model wrote: "Новое задание уже на карточке. Когда Бип выберет ответ, напишите мне букву (A–E), и я её запишу" — the task is on the card; when the child picks an answer, write me the letter and I will record it. The card records an answer by itself, and the instructions say so twice, the first time in "Always", which also says that once a card shows a task, the model says nothing about it. A parent who did as asked would spend a message and a turn on every task — the very limit this run measures. Whether the model wrote it under every card was not counted.
2. **The waiting card changes its language mid-wait.** Before `next_task` answers, the card speaks the host's language — English, the language of the author's Claude — and with the answer it switches to the profile's, Russian, and gains its top line and grade. That is how `ChoosingCard` is written ("it speaks the host's language"), but a child sees the wait begin in one language and go on in another. The host can hand the card the call's arguments before its result (`ui/notifications/tool-input`, `ontoolinput` in ext-apps 2.0.3), and `next_task` is called with `language`.
3. **After the limit, "Another task" cannot reach the chat, and the card has no word for it.** The send fails in the chat (above). What the card showed then was not seen: it reports an ask the host refused as not sent (`chat.not_sent`), and one the host took as taken, and which of the two Claude answers at the limit is not known.
4. **After the limit, no card and no progress.** Once the send failed, the chat showed the failed message alone, and there was no card to open the progress from — although the answer on the open card had been recorded half a minute before. Whether the cards come back when the chat is opened again, and whether their calls go through then, was not checked.

The fixes are T64.5 in RUN.md.

## The fixes of T64.5

### Finding 1: where the letter in the chat came from

The words of an accepted task, which the model reads the moment its task is on the card, ended "Never say which option is right before the child has answered; record the answer with submit_answer" — right after they read out the options A to E for a host with no card. The description of `submit_answer` opened "Records the child's answer to the task on the card". Both told the model that recording the answer was its job, and it asked the parent for the letter so as to do it. "Always" says "An answer on a card is recorded without you", but it is read once, at the start, and the words of the task came last. `next_task` says nothing of the answer.

Now (R103):
- the words of an accepted task, and of a task handed in again while the card shows it, say that where the card shows the task the child answers there and the card records the answer itself, so the model asks for no answer in the chat, which would cost the parent a message; without a card the model reads the task out and records the answer the child gives in the chat. The words of a task handed in again said nothing of a host without cards before;
- the descriptions of `submit_task` and `submit_answer` say the same of a host that shows cards;
- step 5 of "A task" in the instructions says the child reads and answers the task on the card, so the instructions version changes with it.

"Always" stays as it is: it ends at 2,038 characters of the 2,048 a host may keep. A test holds the words and the descriptions to it, and fails without the change.

### Finding 2: the waiting card speaks the ask's language

`ext-apps` 2.0.3, the version the widget pins, holds a host to telling a card the whole arguments of its call once, after the handshake and before the result: "The host MUST send this notification after the View completes initialization … and complete tool arguments become available. This notification is sent exactly once and is required before sendToolResult." `next_task` is always called with `language`, the chat's.

The card now keeps that argument, and nothing else of the arguments, and speaks it from its first word. Until the host has told it, the card shows nothing; a call cancelled before then is said in the host's language, since nothing comes after it. A refusal of the ask, and the first sign-in, whose payloads name no language, go on in the ask's language too (R190).

One switch remains, by R144: when the parent chose a language for the lessons other than the chat's, the wait begins in the chat's and turns to the parent's as the result comes, since the ask cannot know the profile. In T64.4 the chat and the lesson were both in Russian, so its wait would not have switched.

Whether Claude tells a card the arguments is for the live check. If it does not, the card stays blank until the result, which comes in about two seconds.

### Findings 3 and 4: at the limit

Both are for the live check. What the card does now after "Another task":
- the host took the message — "Once the ask reaches the chat, the new task will come below, in a new card.";
- the host refused it — "Not sent — try again".

In T64.4 Claude put the words into its input under a warning, and the send failed there, in Claude's own interface. If Claude answered the card before the send, the card shows the first note, and the refusal is Claude's alone to show. If it answered after, with a refusal, the card shows the second, and "try again" is then the wrong advice.

### The refused measurements of T64.3

**How Google lays the series out.** The Telemetry API files every OTLP measurement as a Prometheus series under the monitored resource `prometheus_target`, and its `instance` label is taken from the resource's `instance`, then `service.instance.id`, then `faas.instance`; a point with none is refused. The service's resource carries `faas.instance`, the instance id the platform's metadata names, so every instance writes series of its own. Each refusal names its resource with it.

**What the refusals were.** A narrow read of the log, allowed by the author, found the eight lines and nothing else of the kind in nine days:

| When, UTC | Revision | Series | Points apart |
|---|---|---|---:|
| 2026-09-29 15:38 | 00058 | `http_request/delta` | 2.9 s |
| 2026-09-30 01:30 | 00063 | `http_request/delta` | 4.8 s |
| 2026-09-30 02:01 | 00064 | `target_info/gauge` | 2.0 s |
| 2026-09-30 02:02 | 00064 | `http_request/delta` | 2.8 s |
| 2026-10-02 01:00 | 00070 | `target_info/gauge` | 2.0 s |
| 2026-10-02 01:04 | 00070 | `http_request/delta` | 5.5 s |
| 2026-10-02 01:30 | 00070 | `target_info/gauge` | 1.0 s |
| 2026-10-02 02:10 | 00070 | `http_request/delta` | 5.5 s |

Every one is from a build before T62.1, which reached the service on 2026-10-03. Those builds delivered measurements on the library's own clock as well as with requests, and the two fell due seconds apart after a quiet minute: the same instance wrote its own series twice. Since T62.1 none has come — through the load of T64.3, three instances among it, and the lesson of T64.4 — and no stop's delivery has failed either. The resource stays as it is.

A gap of 5.5 s was refused, so the least gap these series take is more than the 5 s Cloud Monitoring's quotas name. Deliveries a minute apart are far from it either way.

**What the measurements cost.** OTLP metrics are billed as Prometheus samples, $0.06 a million, with no free allotment. The 150 MiB a month R49 counted on is the allotment of metrics billed by bytes, which these are not. A point of a counter is one sample, and a point of a histogram is two plus one for each bucket that is not empty; a delivery carries up to about twenty. At a few lessons a day that is under a cent a month, and with three instances answering the whole month, under $0.20. A series of a process of its own, with `service.instance.id`, would add no sample: a sample is a point written, whichever series it is in. Whether the measurements are worth a cent against the promise of $0 is О-58.

### The live check

It needs the build with these changes deployed, which happens after the merge, and one five-hour session of the company account's free plan. Its cost is $0.

1. A new chat in Claude on the web, signed in to the company account. Claude's interface is in another language than the lesson's, as in T64.4.
2. The lesson goes as in T64.4. From the first ask, watch the waiting card: it should speak the lesson's language from its first word, or show nothing until `next_task` has answered, about two seconds. If it speaks the interface's language, the change has not reached the service.
3. Under every task's card, the model writes nothing about the answer, and in particular asks for no letter. Each answer is given on the card.
4. At Claude's limit, answer the open task on its card first, then press "Another task". Photograph the card's note twice: once its words stand in Claude's input, before they are sent, and again after the send has failed.
5. Reload the chat's page, and open the chat again from the list. Are the cards there, and does the last one show its task? Open the progress from it. Photograph what each shows.
6. Within the hour, `just report 1h`: `read_task` and `read_progress` after the limit say whether the cards' calls went through.

What follows from step 4: if the note changes to "Not sent — try again" once the send fails, the refusal can be seen, and the card should say instead that the chat's messages have run out and the task on the card stays. If it keeps "Once the ask reaches the chat…", Claude answers the card before the send, and the refusal is a property of the host. What follows from step 5: if the cards do not come back, or their calls do not go through, that goes into PRODUCT 9.4 as a property of the platform.

## How to repeat it

```sh
just load-signin parent -callback-port 8976
just load-signin load -callback-port 8976
gcloud run services update mathtrail --region=us-central1 --project=mathtrail-prod --concurrency=2 --quiet
just load paces -url https://mcp.mathtrail.app -accounts parent,load -rate 6
just load ceiling -url https://mcp.mathtrail.app -accounts load,parent
gcloud run services update mathtrail --region=us-central1 --project=mathtrail-prod --concurrency=80 --quiet
just report 1h
just usage 1h
```

Port 8976 must be forwarded in VS Code (the Ports panel). An account whose profile is real goes first in `paces`, and not first in `ceiling` (`docs/load.md`).

T64.4 is a lesson in Claude on the free plan, held as its section describes, followed by `just report 1h` while the lesson's hour is still the last one.
