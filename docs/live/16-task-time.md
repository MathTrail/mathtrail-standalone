# Where the minute goes (T85)

**Measured: in claude.ai a task reaches the card 36 s after the adult presses Enter, and half of that the model spends typing out the task it hands in. The service's own work is 4 s, nearly all of it Drive, three quarters of it two writes of the profile.** 2026-10-06. Families wait longer than that, 55 s on average. How long a task takes depends on the family's chat as much as on the service, above all on how hard the chat's model is set to reason.

- **The author's six tasks in claude.ai, timed on the screen against the log.** They ran on Opus 5.5 at medium effort, on the deployed version with drawings by default (T83). From Enter to the task on the card took 36.3 s (32.5–41.5 s). Of that:
  - the model typed its hand-in for 18.5 s (51 %) and thought before it for 6.5 s (18 %);
  - the service took 4.1 s (11 %), nearly all of it Drive;
  - the model took 2.6 s from Enter to its first call, and 2.0 s from one call to the next;
  - the card's poll took 1.9 s and its ticks 0.8 s;
  - the host asked no question of permission.
- **Families in production, the 43 tasks that reached a card.** The service saw 51.8 s on average (median 43.6 s, 90th percentile 72.8 s). With the stages before the first call and after the last poll, a family waits some 55 s. Of the 51.8 s:
  - the gap in which the model writes took 32.5 s, together with any question of permission the host asked inside it;
  - the card's poll took 7.6 s, longer than the author's because pages were hidden and cards slowed down;
  - the model's step between the calls took 5.0 s;
  - the service's calls took 4.3 s, nearly all of it Drive, and refusals 2.5 s.
- **The model's setting decides what its writing is made of.**
  - At medium effort in claude.ai, the model thinks for 6.5 s and types for 18.5 s.
  - In Claude Code at `xhigh`, the same model on the same version thinks for 86 s and types for 16 s.
  - On Sonnet at Claude Code's default, R205's list of ten ideas added some 45 s of thinking.
  - The hand-in is about 5 KB. Its brief alone, which the service already holds, is a seventh of it: some 2.6 s of the author's 18.5 s of typing, if the author's hand-ins were built like the chats'.
- **Drive.** Every write of the profile takes 1.4 s, and a task makes two on its path: three quarters of the service's time. The reads make up most of the rest. The checks and the solver's two runs take milliseconds.
- **Cold starts are paid when a chat connects, not while a child waits.** An instance starts about 18 times a day, and its first request waits 0.6 s more. In two weeks, a fresh instance met a call of a task's path at most twice: one `next_task` and one card's poll.
- **No lever on the service's side comes near 5 s.**
  - The levers no decision has ruled out save 2.0 s of the author's wait and 5.6 s of the families'. These are the card asking every 2 s while answers come, `next_task`'s write after its answer, and fewer refusals.
  - Three alternatives that R152 and R201 rejected, measured here as new evidence, would bring the savings to some 6 s and 13 s.
  - A shorter hand-in would save up to 8 s more at medium effort.
  - What remains is the model writing a task at the moment it is asked for.
  - Only a task written before it is asked for gets under 5 s, and only if the card can take it by itself (see "Writing ahead").

## What this is

Families and the author wait a minute or more for a task (07, finding 7). The author's goal is under 5 seconds. This report takes the wait apart, stage by stage. For each stage it says what it costs and what could shorten it. The report changes nothing; the levers at the end are for the tasks that follow.

The service sees only its own part of the wait: its calls, each with its time and what Drive took of it. The model's time between the calls shows in the log only as gaps. Before the first call and after the last, it does not show at all. So the wait was measured from six sources, each used where it sees best:

| Source | What it sees | Sample |
|---|---|---|
| The service's log in production, 2026-09-22 to 2026-10-06 | each call of each request, joined into tasks; Drive per call; the card's polls | 43 tasks that reached a card, 17 of them on the last version before T83 |
| The platform's request log and its lines on instances | what the platform adds to each request; cold starts | 4 170 requests to `/mcp`, 259 instance starts |
| Kept traces | the tree of one call, and, since Claude sends one trace for a whole turn of its model, the calls of one task in one view | 6 traces |
| `just play`: Claude Code's stream, each line stamped as it arrived | the model's own turn: the first token, thinking, and the hand-in written part by part | 20 new chats on the deployed version, on Opus 5.5 at `xhigh`; 275 older ones on Sonnet 5 |
| The author's lessons in claude.ai, in the log | the same stages as in production, on the deployed version | 13 tasks, two sittings |
| The author's screen recording, against the log | Enter, the model's first call, its thinking and its typing, the card drawn, the host's questions, the task appearing | 6 tasks, the second sitting |

**How the calls of one task were joined.** The service logs no link between the calls of one request. Each call is an HTTP request with its own id, and the task's request id is never logged. An account has one child, and so at most one open request. So each account's calls were walked in the order they began:
- `next_task` opens a task;
- `get_package`, the hand-ins and the card's polls join it;
- acceptance, a third refusal or fifteen minutes of quiet close it.

A call began at its line's time less its `duration_ms`. Each joined task was checked against the service's own count: `seconds_since_request` lies between the calls that bound it in every task, 67 in production, 13 of the author's and 20 in the chats. The count may run a second past those calls, because the request's opening is kept to the second.

**Which tasks.** In those two weeks, 93 tasks were accepted in Claude in production:
- 67 of them went through `get_package`, which the versions before R152 did not have;
- 43 of those 67 were taken away by a card's poll;
- the other 24 reached no card, 20 of them in 45 minutes of one night with no poll at all, most likely the reviewers' demo profile being filled (T71.3).

A family's wait is measured on those 43. The load tool's tasks (host `other`) are left out.

**The screen.** The author held two sittings of Opus 5.5 at medium effort, the setting they use, of 7 tasks and 6 tasks. The second was recorded. The recorder stamps its own start in UTC, and the frames were set against the log by it. Three checks confirm the alignment:
- the waiting card appears within a second after `next_task` ends;
- the task appears 0.5–1.0 s after the poll that took it ends;
- the label of the hand-in's call appears before the call reaches the service.

Moments were read at a frame a second, so each is good to half a second.

## In claude.ai: from Enter to the task on the card

The author's six recorded tasks, in seconds. Every stage is bounded by an event seen on the screen or written in the log. The stages follow one another, so their sum is the whole wait. Every second of all six tasks is named, to the half second the frames allow.

| Stage | Mean | Range | Share | Measured by |
|---|---:|---:|---:|---|
| The adult presses Enter, to the model's first call, `next_task` | 2.6 | 2.1–3.1 | 7 % | screen, log |
| `next_task` on the server | 2.1 | 1.6–3.8 | 6 % | log |
| the waiting card on the screen, after `next_task`'s answer, beside what follows | ≤ 1 | | — | screen |
| `next_task`'s answer, to `get_package`'s call | 2.0 | 1.8–2.1 | 6 % | log |
| `get_package` on the server | 0.2 | 0.1–0.2 | 1 % | log |
| The model thinks: the package's answer, to the hand-in's call showing in the chat | 6.5 | 4.2–8.3 | 18 % | log, screen |
| The model types the hand-in: its call showing, to its arrival | 18.5 | 16.5–20.6 | 51 % | screen, log |
| `submit_task` on the server | 1.8 | 1.6–2.4 | 5 % | log |
| The task saved, to the card's poll that took it | 1.9 | 0.6–3.6 | 5 % | log |
| The card ticks off its steps and shows the task | 0.8 | 0.5–1.0 | 2 % | screen |
| Questions of permission | 0 | | 0 % | screen |
| **From Enter to the task on the card** | **36.3** | **32.5–41.5** | | |

RUN.md asks that the named stages explain at least 90 % of the median wait in claude.ai. In each of the six tasks they explain all of it, to the frames' half second, so the median task's share is 100 %. Each share is the stage's part of the mean, rounded.

Before Enter, the card's "Another task" button only puts its words into the chat's input, under Claude's warning about running a prompt (07, finding 5). The author pressed Enter 1.0–2.0 s later.

While the model thinks, Claude shows a title of its thinking, such as "Working out a queue position counting puzzle". While it types, Claude shows the hand-in's call, "Hand in a written task". At medium effort the thinking was short and the typing long, in all six tasks.

**Both sittings in the log**, 13 tasks:
- the service saw 38.5 s on average (median 36.2 s): the model's writing 29.0 s, Drive 4.3 s, the card's poll 2.0 s;
- one refusal cost 21.6 s;
- the first sitting took 43.4 s on average over 7 tasks, the second 32.9 s over 6.

## Families in production

The 43 tasks that reached a card, in seconds. The stages follow one another without a gap, so their means add up to the mean of the whole span. Their medians and 90th percentiles do not add up.

| Stage | Mean | Median | 90th | Share |
|---|---:|---:|---:|---:|
| `next_task` on the server, all of it Drive | 2.2 | 1.9 | 3.7 | 4 % |
| `next_task`'s answer, to `get_package`'s call | 5.0 | 2.2 | 6.8 | 10 % |
| `get_package` on the server | 0.2 | 0.2 | 0.2 | 0 % |
| The package's answer, to the first hand-in: the model writes, and any question of permission | 32.5 | 30.9 | 45.5 | 63 % |
| Refusals and writing again | 2.5 | 0.0 | 0.0 | 5 % |
| The accepting hand-in on the server, all but 10 ms of it Drive | 1.9 | 1.7 | 2.2 | 4 % |
| The task saved, to the card's poll that took it | 7.6 | 3.4 | 10.2 | 15 % |
| **The span the service sees** | **51.8** | **43.6** | **72.8** | |

What is left unsplit is the 32.5 s gap. It holds the model's thinking, its typing, and any question of permission the host asked before the hand-in. The author's tools, always allowed, asked none, so how long a new family's questions take was not measured. Cold starts lie outside the gap (see "The platform").

Before that span, the adult's message reaches the model, which then makes its first call. After it, the card ticks off its steps. The author's screen puts these at 2.6 s and 0.8 s, so a family waits some 55 s. On the last version before T83 alone, 17 tasks, the span was 54.7 s (median 44.5 s, 90th percentile 130 s): the model's writing 35.3 s, the card's poll 7.1 s, Drive 4.4 s.

The families' writing, 32.5 s, sits between the author's 29 s at medium effort and the chats' 104 s at `xhigh`. Which model and setting a family uses, the service cannot see.

**By group**, the same 43 tasks: the whole span, mean in seconds, and where it helps, the first hand-in's writing and the writing with refusals.
- **Instructions:** the five versions before R205 (26 tasks) took 50.0, the first hand-in 30.6 and with refusals 33.6. R205's version (17 tasks) took 54.7, the first hand-in 35.3 and with refusals 37.1.
- **Attempts:** one (39 tasks) 49.1; two (3) 71.3; three (1) 98.6.
- **The first task of a sitting** (3 tasks): 58.6, against 51.3 for the rest.
- **Level and difficulty:** 40 of the 43 tasks were at `1-2`, so the levels cannot be compared here. Difficulty 1 took 49.3 (18 tasks) and difficulty 2 took 57.5 (20 tasks).
- **Drawings:** the log says whether a task drew only since T83, and none of these tasks is that recent.
- **Requests found open:** none of the 43. One task reached by a second `next_task` took 229 s, and no card took it.

```mermaid
sequenceDiagram
    actor Adult
    participant Claude as Claude (host)
    participant Model as Chat's model
    participant MT as MathTrail
    participant Drive
    participant Card
    Adult->>Claude: asks for a task ("Another task", then Enter)
    Claude->>Model: the message
    Note over Claude,Model: 2.6 s to the first call
    Model->>MT: next_task
    MT->>Drive: read the profile, write the request
    Note over MT,Drive: 2.1 s, all of it Drive
    MT-->>Claude: the card's resource
    Claude->>Card: draws the waiting card (under 1 s)
    Note over Model: 2.0 s
    Model->>MT: get_package
    Note over MT: 0.2 s
    Note over Model: thinks 6.5 s, types the hand-in 18.5 s
    Model->>MT: submit_task
    MT->>Drive: read, check, write
    Note over MT,Drive: 1.8 s; the checks take milliseconds
    loop every 4 s, every 15 s after two minutes without news
        Card->>MT: read_task (one Drive read, 0.2 s)
    end
    Note over Card: 1.9 s after the save, then 0.8 s of ticks
    Card-->>Adult: the task, 36.3 s after Enter
```

The durations are the means of the author's six recorded tasks. Families' are in the table above.

## The model writes the task

**In claude.ai** (the author's six tasks, Opus 5.5 at medium effort), the model thought for 6.5 s after the package and typed for 18.5 s.

**In Claude Code**, `just play` ran 20 chats on the deployed version (`d42d50d5a122`):
- the model was `claude-opus-5-5` at `--effort xhigh`, the nearest Claude Code has to claude.ai's Extra reasoning;
- each chat held one task;
- the five cells, four chats each, were taken in turn.

From the package's answer to the end of the hand-in:

| | Mean, s | Median, s | 90th, s |
|---|---:|---:|---:|
| Whole | 104.4 | 77.8 | 199.1 |
| Before the first token | 1.9 | 1.4 | 3.4 |
| Thinking | 86.3 | 61.2 | 172.5 |
| Typing the hand-in | 16.1 | 13.8 | 23.0 |

Thinking takes 71 % of a chat's output tokens. The hand-in is 5.2 KB on average. Here are its parts, with the time the model spent typing each (mean, seconds) and their size:

| Part | Typing, s | Bytes |
|---|---:|---:|
| `core_idea`, opening with the list of ten ideas | 3.5 | 1 065 |
| `solver` | 2.2 | 821 |
| the brief | 2.1 | 733 |
| `self_check` | 1.8 | 539 |
| `design_thought_process` | 1.7 | 503 |
| the drawing's structure (7 tasks) | 1.4 | 478 |
| `solution` | 1.3 | 416 |
| `distractors` | 1.2 | 378 |
| `question` | 0.7 | 283 |
| the drawing (7 tasks) | 0.4 | 194 |
| `options`, `hint`, `correct_answer` | 0.5 | 194 |

By cell, the whole writing, mean (median):

| Cell | Writing, s | Thinking, s |
|---|---:|---:|
| Ordering, `1-2`, difficulty 2 | 103 (65) | 85 (51) |
| Arithmetic with a trick, `3-4` | 72 (68) | 57 (55) |
| Weighing and pouring, `3-4`, drawn | 170 (176) | 148 (154) |
| Enumeration, `5-6` | 67 (60) | 52 (45) |
| Knights and liars, `5-6` | 110 (109) | 90 (90) |

In weighing and pouring, the package always shows a drawn reference task. That cell thinks nearly three times as long as arithmetic at the same level, and it did so in T83's run on Sonnet as well (168 s of thinking against 60 s). These runs cannot say whether the drawing or the topic costs it: no cell holds the topic without a drawing.

**The older chats** ran on Sonnet 5 at Claude Code's default effort. The same split comes from the block times Claude Code writes, without the piece-by-piece stamps:

| Run | Instructions | Chats | Writing, mean (median), s | Thinking, mean (median), s | Typing, mean, s |
|---|---|---:|---:|---:|---:|
| T64a | before R205 | 119 | 70 (57) | 60 (47) | 10 |
| T80, base | before R205 | 58 | 65 (49) | 55 (40) | 11 |
| T80, after | R205 | 58 | 114 (104) | 100 (92) | 14 |
| T83 | R205, drawings | 40 | 141 (133) | 126 (118) | 15 |

On Sonnet, R205's list of ten ideas cost some 45 s of thinking. In production, in chats with a card, the same change cost little. The model's first hand-in took 30.6 s on the five versions before it (26 tasks) and 35.3 s after it (17 tasks).

**What this says.** The model reads its context fast: under 2 s from the package's answer to its first token. After that, two things take its time:
- thinking, which the setting of the family's chat decides, and which the guide decides through how much it asks the model to settle: the ten ideas, the drawing, the traps, the self-check;
- typing, which the size of the hand-in decides. At medium effort, typing is most of the writing.

## The service

| Call | Mean, ms | Median, ms | 90th, ms | Drive's operations |
|---|---:|---:|---:|---|
| `next_task` | 2 053 | 1 856 | 2 843 | read, read, write |
| `get_package` | 203 | 196 | 241 | read |
| `submit_task` | 1 941 | 1 769 | 2 641 | read, read, write |
| `read_task` | 233 | 193 | 247 | read |

Nearly all of it is Drive. Without Drive, no call of a task's path takes more than 19 ms at the 95th percentile (`just report`). Drive by operation, over every call of the two weeks:

| Operation | n | Mean, ms | Median, ms | 90th, ms |
|---|---:|---:|---:|---:|
| `update` (a write) | 314 | 1 438 | 1 354 | 1 959 |
| `download` (a read) | 1 502 | 232 | 194 | 251 |
| `list` (finding the file) | 323 | 318 | 295 | 378 |
| `get` | 238 | 93 | 80 | 113 |

No operation was retried. A write of the profile is the one expensive thing the service does: 1.4 s, twice on a task's path and once more for each refusal. In a kept trace of `submit_task`:
- the solver's two runs: 1 ms;
- the checks: 11 ms;
- the reads: 163 and 285 ms;
- the write: 1 140 ms.

## The card

The card asks `read_task` 0–1.5 s after it is drawn, then 4 s after each answer. It slows to every 15–16.5 s once 120 s pass with no news or a call goes unanswered, and it does not ask at all while its page is hidden. Each ask is one Drive read, 0.2 s. Once the task comes, the card ticks off its steps and shows the task 0.8 s later.

In production, the time from the task's save to the poll that took it was 7.6 s on average, median 3.4 s, 90th percentile 10.2 s, longest 113 s. The pause before that poll was up to 6 s in 35 of 43 tasks, 6–18 s in 2 and longer in 5; one task was taken by the card's first poll. A card asks some 8–9 times a task. In the author's sittings, with the page in view, the same stage took 1.9 s.

In the recorded tasks, the model's last words, "a new task is on the card", reached the chat up to 3 s before the card showed the task. The card learns of the task only at its next poll.

## The platform

- **What it adds to a request.** The service's own line says whether it kept a request's trace. A request whose trace the service dropped waits 14 ms more on average (median 2 ms); one whose trace it kept waits 129 ms more (median 124 ms, longest 211 ms). The difference is the kept trace's spans, delivered before the request ends, with a deadline of 200 ms. The platform's timestamp of a request is its arrival.
- **One more request before each call.** Claude sends the endpoint 1.83 requests for each tool call. In a kept trace, a request of some 0.1 s comes before each of `next_task`, `get_package` and `submit_task`.
- **Cold starts.** The platform started 259 instances in the two weeks, about 18 a day, since none is kept warm; 258 of them began listening within the service's lines. Two views of them:
  - **The platform's.** 236 instances took requests to `/mcp`. For 122 of them, the first request could be joined to the service's line. Each of those 122 was the host's connection, before any tool call, and waited 0.6 s more than the service's own time (90th percentile 1.3 s, longest 3.2 s). After such a connection, the next `next_task` on the instance came a median of 35 s later.
  - **The service's.** Only 28 of the 258 instances served any tool call at all. Six served their first within 10 s of starting. Two of those six could lie on a task's path: a `next_task`, and a card's poll, `read_task`.

  So in two weeks, a cold start met a task at most twice. The author's first sitting met an instance started four minutes before, by a deployment.

## Traces

Claude sends one trace for a whole turn of its model. Of the two weeks' traces, 74 hold more than one call, 41 of them exactly `next_task`, `get_package` and `submit_task`. So a kept trace is a picture of one task as the service saw it, with the model's gaps between the calls. One of them, from the last version before T83:

| At, ms | Took, ms | Span |
|---:|---:|---|
| 0 | 99 | a request before the call |
| 225 | 2 453 | `next_task`: Drive read 227, read 189, write 1 880 |
| 6 030 | 95 | a request before the call |
| 6 262 | 194 | `get_package`: Drive read 191 |
| 46 968 | 105 | a request before the call |
| 47 229 | 1 742 | `submit_task`: the solver 1, Drive read 163, the checks 11, read 285, write 1 140 |

The model's writing is the empty stretch from 6.5 s to 47 s. A trace with a refusal shows the second writing just as plainly: the refused hand-in at 31.9 s, the accepted one 13 s later.

What a trace does not show:
- the turn before the first call and after the last: the message reaching the model, and the card;
- the card's polls, which carry traces of their own;
- the link between calls, on a host that does not send one trace a turn;
- a turn whose trace was dropped, nine in ten. Of the kept ones, one in eight lost its spans to the delivery's deadline (`just report`: 46 of 339).

The spans keep what SPEC 12.5 promises. The client's address, the peer's address, the user agent and the path read `redacted`, and the platform's own span names only the service's address.

**What the log could add**, as proposals, to make the next measurement simpler:
- the task's request id on the lines of `next_task`, `get_package`, `submit_task` and `read_task`, so that a task's calls join without walking an account's calls;
- `read_task`'s answer, writing or on the card, so that the poll that took the task is known rather than inferred;
- `seconds_since_request` in milliseconds, or the request's opening kept to the millisecond;
- the size of the hand-in, by part.

## Levers

Each lever's gain was counted task by task: the span the service saw, less what the lever saves there. The families' 43 tasks span 51.8 s. The author's six recorded tasks span 32.9 s in the log, the same tasks whose whole wait the screen puts at 36.3 s. The median of a gain is therefore the median of the shorter waits, not the median of the savings.

A lever can only save what lies on the card's path:
- The accepting hand-in's write must land in Drive before a poll can find the task, so of the two writes only `next_task`'s can leave the path.
- A card on a hidden page asks nothing, so a faster poll saves nothing where the page was hidden when the task came. That was 5 of the families' 43 tasks and none of the author's.

**Levers no decision has ruled out**

| Lever | Stage | Families: gain, mean (median), s | The author's, s | Cost | Risk to the tasks |
|---|---|---|---|---|---|
| The card asks every 2 s while answers come, and keeps the 15-s pace only after a question goes unanswered | the card's poll | −1.4 (−0.7) | −0.3 | constants in `waiting.ts`; twice the reads of Drive while a task is written; it changes R152's pace | 30 asks a minute from each waiting card. That is half of an account's pace of 60, and some seven cards waiting at once on one instance reach the instance's fuse of 200 a minute |
| `next_task` answers before its write to Drive | the service | −1.7 (−1.5) | −1.7 (−1.5) | a change to the store's contract. With CPU allocated only while a request is served, a write after the answer is throttled, unless the instance is billed whole | `get_package` and the card's first poll may read the profile before the request is in it; a write lost after the answer leaves a request nobody can find |
| Fewer refusals | refusals | up to −2.5 (−0.5) | 0: no refusal in the recorded sitting | the guide, the hand-in's form | none if the checks stay as they are; 3 of the chats' 4 refusals were the structure |
| A shorter hand-in: the brief not handed back, the list of ideas not written out, no `design_thought_process`, a lighter self-check | typing | not measured: the families' typing is not timed | up to −8 at medium effort, estimated | the hand-in's form, SPEC 4.2–4.5 | the service loses what it checks the model by; the list of ideas is how R205 moves the model along |
| The model thinks less: no list of ideas, a drawing only where it shows something | thinking | unknown without a run that changes one thing at a time | little at medium (6.5 s); most of the writing at `xhigh` | the guide | the variety of ideas (11.9, О-67), drawings (R228) |
| A warm instance | cold start | about 0: a cold start met a task's path at most twice in two weeks | 0 | about one instance's month | — |

The first three together save 5.6 s of the families' span (median 2.6 s) and 2.0 s of the author's (median 1.5 s).

The estimate for a shorter hand-in works like this. Each part of the hand-in takes its share of the bytes in the chats, priced at the author's typing time, 18.5 s for some 5 KB. Four parts would go:
- the brief, 733 bytes;
- the list of ten ideas, about two thirds of `core_idea`'s 1 065;
- `design_thought_process`, 503;
- half of the self-check, 270.

Together they make up some 42 % of the bytes: up to 8 s.

**Alternatives a decision rejected, priced here as new evidence**

| Alternative | Rejected by, and why | Families: gain, mean (median), s | The author's, s |
|---|---|---|---|
| The card learns of the task at once, by a question the service holds open until there is news | R152: a host may cut it, and an instance pays for the time it waits | −2.8 (−2.1) | −1.5 |
| The package in `next_task`'s answer | R152: the card receives the whole of that answer, and the package must not reach it | −5.2 (−2.4) | −2.1 |
| No ticks before the task | R201: steps a card draws and never takes read as a card that stopped halfway | −0.8 | −0.8 |

The measured seconds are the new evidence. The reasons of R152 and R201 stand, and reopening either is the author's decision.

**What is left**, on the same bases: the families' wait of some 55 s, the author's 36.3 s.
- With the three open levers: the families' some 50 s, the author's some 34 s.
- With the rejected alternatives as well, the held-open poll in place of the 2-s pace: the families' some 42 s, the author's some 30 s.
- A shorter hand-in, priced for the author alone: some 8 s less again. The author's wait would come to some 27 s with the open levers and some 22 s with everything.

The rest is the model reading the package, thinking and typing a task at the moment it is asked for. Neither wait comes near 5 s.

## Writing ahead, against 5 seconds

The author's idea (07, finding 7) is that the model writes the next task before it is asked for, and the service keeps it sealed in the parent's Drive. The log says how often the task would have been ready.

**How often it would be ready.** Suppose the model had begun the next task as the last one reached the card, and taken as long as it actually did. Then, of 34 pairs of tasks in one sitting with a card, the next task would have been written by the time it was asked for in 19 (56 %). A child spent a median of 49 s between a task appearing and asking for the next (90th percentile 224 s). The wait left over when it was not ready came to 10 s on average over the 34 pairs (median 0, 90th percentile 36 s). The first task of a sitting is never ready.

**What a ready task still costs.**
- If the model has to take it, the adult's message still goes through a turn of the model and a call. That is 2.6 s to the first call, 2.1 s of `next_task`, and the card's poll after it.
- If the card takes the task itself, with a call of its own, as it already calls `read_task` and `submit_answer`, a ready task costs one read of the profile, one write (about 1.6 s), and the card's drawing. This is the only way under 5 s this report found. It would also mend the button that only puts its words into the chat's input (07, finding 5), which cost the author 1.0–2.0 s before each Enter.

**What it asks for.**
- О-6 again: it cancelled a buffer of ready tasks for v1 and left it to the paid edition.
- The rule chooses the topic and the difficulty from the last answers, so a task written ahead is chosen before the answer that should have chosen it.
- The model's turn grows by a task it may never hand out. A task never asked for spends the family's messages for nothing.
- A task written ahead is sealed like the answer and checked when it is written. A profile changed in between, its language or its topic for the lessons, makes it stale.

## Other findings

- The service keeps a request's opening to the second, so `seconds_since_request` may run up to a second past the calls that bound it.
- The log's own audit (`just report`) found two things outside this task. The `http_request` lines of 2026-09-22 to 25 carry a `path` for 404s (602 lines, older builds). `telemetry_export_failed` (11 lines) is an event the report's table does not name. Neither holds a task, an answer or an address.
- The model streams the brief, the task and the self-check as strings of JSON, which the service reads for the JSON they hold (R137). Claude Code prints them already parsed.

## How to repeat it

**Whether the tool belongs in the repository:** not yet. The log's proposals above (the task's request id on its calls, `read_task`'s answer) would turn the join into a filter, and the chats' and the screen's parts were made for one study. If the levers are measured after each change, the first of those tasks can move `tasktime` into `tools/`, beside `tools/load`.

The tools live outside the repository, in `~/mathtrail-play/runs/task-time/`:
- `bin/tasktime` is a Go program with the standard library alone;
- `bin/run-chats.sh` drives the chats;
- `bin/read-prod.sh` reads the logs;
- `bin/video.sh` cuts frames in `jrottenberg/ffmpeg:7.1-alpine@sha256:8ec1ee1f6a0fcd37c97725827b6b7832795c9596e3439b8da56d7700d61ae778`.

```sh
# production: the service's lines as the report recipe flattens them; the platform's
# request log of /mcp through a projection that keeps no address; instance starts
ROUND=round1 FRESH=30d bin/read-prod.sh
bin/tasktime tasks < prod/round1/service.jsonl > out/tasks.jsonl
bin/tasktime stages out/tasks.jsonl
bin/tasktime drive < prod/round1/service.jsonl
bin/tasktime starts prod/round1/requests.json prod/round1/starts.json prod/round1/service.jsonl

# the chats: a clean copy of the deployed commit, a server on a port of its own,
# the profile by a call, one task a chat, each line of the stream stamped on arrival
git archive <commit> | tar -x -C src
MODEL=claude-opus-5-5 EFFORT=xhigh bin/run-chats.sh 1 20
bin/tasktime streams chats/*-[0-9].tsv > out/chats.jsonl
bin/tasktime writes out/chats.jsonl

# the screen: a recording whose start is stamped in UTC, frames a second
bin/video.sh probe <file>; bin/video.sh sheets <file>
```

A chat of the run is:

```sh
printf '%s' "Hi, it's Pip's parent. Pip would like the next task: <Topic> (<id>), at grade level <level>, difficulty <d> out of 5. Please ask for exactly that topic, level and difficulty." \
  | MATHTRAIL_PLAY_MODEL=claude-opus-5-5 just LOCAL_MCP=http://localhost:8082/mcp play -p \
      --output-format stream-json --verbose --include-partial-messages --effort xhigh --session-id <uuid> \
  | perl -MTime::HiRes=time -ne 'BEGIN { $| = 1 } printf "%.6f\t%s", time, $_'
```

## Cost

- Cloud: $0. Reading the logs and the traces is within the free quotas.
- The 20 chats: $8.81 by Claude Code's count, $0.44 a chat, from the author's subscription.
- The author's lessons: 13 tasks in claude.ai.
