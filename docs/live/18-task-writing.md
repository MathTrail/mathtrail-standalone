# A shorter hand-in (T89)

**Measured: with the brief left with the request, the idea named by the service and no `design_thought_process`, the model hands in half as much and writes a task 40 % faster: 62 s from the package to the hand-in, against 104 s in T85, on the same model at the same setting.** 2026-10-07. Ten chats in Claude Code, Opus 5.5 at `xhigh`, the five cells of T85 twice each, against a local server built from T89's tree (instructions `b66be584423b`).

- **The hand-in is half as long, and typing it takes half the time.** The model typed 2.8 KB on average against 5.2 KB, in 8.4 s against 16.1 s (median 7.6 s against 13.8 s).
  - `core_idea` went from 1,065 bytes, the list of ten ideas with the one chosen, to 214, a sentence or two.
  - The brief, 733 bytes, and `design_thought_process`, 503, are gone.
  - The other parts are much as they were.
- **The model thinks less, too.** It thought for 52 s against 86 s (median 38 s against 61 s), with no list of ideas to write and choose from. Thinking is still 70 % of a chat's output tokens, against 71 %.
- **Every cell is faster,** by 28 % to 58 % of its mean writing.
- **Every task was built on the idea the package named.** In each cell's second chat the package named the next idea, and the task followed it. Every task written ahead was named the idea after the one of the task on the card, as R243 has it, and followed it too. Every `core_idea` was one or two sentences.
- **No refusal came from the new form.** 7 of 27 attempts were refused, by the checks that refused before: the solver, a near-copy of a reference task, a drawing's value, readability. No hand-in was mended, and none brought a member the form no longer reads.
- **Each chat also wrote the next task ahead,** as T88 has it: 10 requests ahead, 10 tasks kept, 8 of them at the first attempt.
- **Not measured here:** claude.ai at medium effort, where the author's lesson after the release will time the wait from Enter to the card; the families' time, which the log of the release will show; criterion 11.9, which needs a series of its own; what a drawing costs, since 2 of these 10 tasks drew against 7 of T85's 20; and whether the model writes `core_idea` and the self-check in English in a lesson of another language, since every chat here was in English.

## What this is

T85 ([16-task-time](16-task-time.md)) found that the model spends half the wait typing its hand-in, and that part of it the service does not need. T89 took that part away (R241, R242, R244):
- the hand-in no longer carries the brief, which the request keeps;
- the service names the idea of the topic in words from a list of its own, so the model no longer writes a list of ten and picks one;
- `design_thought_process`, which nobody read, is gone;
- slips of form with one reading are mended rather than refused.

This run asks what that did to the model's writing, on the same model, the same setting and the same cells as T85's chats.

## The run

`just play` ran ten chats one after another, each asking for one task of a cell in the parent's words, as in T85:
- the model was `claude-opus-5-5` at `--effort xhigh`;
- the five cells of T85, two chats each, were taken in turn;
- each chat asked for one task, and after it the model wrote the next one ahead with `prepare_task`, as every lesson does since T88.

The server ran from a copy of T89's working tree, before any commit, with its own seal key, in memory. Only the task asked for in each chat is compared with T85: it is written from `get_package`'s package, as T85's were. The task written ahead is counted with the service's log below.

## The model writes the task

From the package's answer to the end of the first hand-in:

| | T89, mean | T89, median | T85, mean | T85, median |
|---|---:|---:|---:|---:|
| Whole, s | 62.3 | 49.7 | 104.4 | 77.8 |
| Before the first token, s | 1.5 | 1.2 | 1.9 | 1.4 |
| Thinking, s | 52.3 | 38.4 | 86.3 | 61.2 |
| Typing the hand-in, s | 8.4 | 7.6 | 16.1 | 13.8 |
| The hand-in, bytes | 2,756 | 2,380 | 5,200 | |

A median of an even count is its lower middle value, as `tasktime` and `just report` take it.

The hand-in's parts, as the model typed them: the time spent typing each, mean, and its size.

| Part | T89, s | T89, bytes | T85, s | T85, bytes |
|---|---:|---:|---:|---:|
| `core_idea` | 0.7 | 214 | 3.5 | 1,065 |
| `solver` | 1.9 | 684 | 2.2 | 821 |
| the brief | — | — | 2.1 | 733 |
| `self_check` | 1.7 | 494 | 1.8 | 539 |
| `design_thought_process` | — | — | 1.7 | 503 |
| the drawing's structure (T89: 2 tasks; T85: 7) | 1.0 | 398 | 1.4 | 478 |
| `solution` | 1.1 | 346 | 1.3 | 416 |
| `distractors` | 1.2 | 389 | 1.2 | 378 |
| `question` | 0.6 | 244 | 0.7 | 283 |
| the drawing (T89: 2 tasks; T85: 7) | 0.3 | 99 | 0.4 | 194 |
| `options`, `hint`, `correct_answer` | 0.6 | 217 | 0.5 | 194 |

By cell, the whole writing in seconds: T89's mean and its two chats, T85's mean (median) of four:

| Cell | T89 | T85 |
|---|---:|---:|
| Ordering, `1-2`, difficulty 2 | 47.2 (28.1, 66.4) | 103 (65) |
| Arithmetic with a trick, `3-4` | 30.5 (32.2, 28.8) | 72 (68) |
| Weighing and pouring, `3-4`, drawn | 121.0 (182.5, 59.5) | 170 (176) |
| Enumeration, `5-6` | 48.3 (43.0, 53.7) | 67 (60) |
| Knights and liars, `5-6` | 64.2 (78.8, 49.7) | 110 (109) |

Two chats a cell are few. One chat of weighing and pouring thought for 169.5 s, longer than that cell's mean of 148 s in T85, and the other for 50.6 s. The whole run shows the change better than any cell.

## The service's log

`go run ./cmd/report` over the server's log of the run:
- 20 requests: 10 asked for and 10 written ahead;
- 27 hand-ins, 7 refused and none out of attempts. The refusals were counted by `solver_error` 3 times, `near_duplicate` twice, `bad_structure` once (a drawing's value that was not a whole number) and `readability` once. The attempt counted by `bad_structure` also failed `drawing_mismatch` and `solver_error`;
- the 10 tasks asked for were accepted, 7 at the first attempt, a median of 51 s from the request to the acceptance, 80 s at the 90th percentile and 203 s at the longest;
- the 10 tasks written ahead were kept, 8 at the first attempt, a median of 48 s and 105 s at the longest. Each next chat asked for another cell, so 9 of them were let go as asked, and none was handed out;
- the hand-ins, as the service measures them in the JSON it builds again, came to a median of 2,107 bytes and 3,638 at the 90th percentile. The task was 1,328 bytes, the self-check 414, the solver 402 and the core idea 211;
- nothing was mended, every hand-in was of the new form, and every line kept to the fields decided for its event.

The service's sizes are smaller than what the model typed: the model's stream holds the request id, the JSON's spaces and the escapes of the solver's lines, and the service measures the solver as the program it is.

## The ideas

| Chat | The idea the package named | The task's core idea |
|---|---|---|
| Ordering, 1 | 1, round 1: a runner who overtakes the one in a place takes that place | passing the runner in third place, then being passed by one |
| Ordering, 2 | 2: two orders of the same children, by height and by age | the height facts chain one way, the age facts another |
| Arithmetic, 1 | 1: multiplying in a clever order, 25 with 4, 5 with 2 | pairing 25 with 4 and 5 with 2 |
| Arithmetic, 2 | 2: a number thought of and multiplied, the multiplication undone | undoing the steps backwards |
| Weighing, 1 | 1: two jugs and a tap | two 3-litre pails into a 5-litre one, the litre moved across |
| Weighing, 2 | 2: a balance and two weights, the loads one weighing can weigh | 2 kg and 5 kg on either pan weigh four amounts |
| Enumeration, 1 | 1: menus with one combination not allowed | 36 lunches less the 3 with the banned pair |
| Enumeration, 2 | 2: dominoes, a double counted once | even totals from even and odd halves, doubles once |
| Knights and liars, 1 | 1: islanders each saying how many knights there are | at most one knight, so exactly one |
| Knights and liars, 2 | 2: what another islander would answer | a statement about a statement makes Ben a liar |

The first task of ordering is a variant of its idea rather than the guide's own example of it: the model passed the third runner and was then passed, where the guide's example passes the second.

The tasks written ahead, each written while the task asked for was on the card, were named the next idea: the second of the topic's list in each cell's first chat, and the third in its second. Each was built on it: two orders by height and by age, a number undone, two weights on a balance, a domino set's dots, what another islander would say; then a tie in a chain of heights, pairs that make a hundred, three coins and a lighter fake, two dice, a circle of knights and liars. Before R243 each of them would have been named the idea of the task on the card. The tasks written ahead were let go when the next chat asked for another cell, and let go tasks are not counted, so each cell's second chat was named the idea its first chat's task written ahead had had.

The first task of weighing and pouring was refused once as a near-copy of a reference task: the idea named, two jugs and a tap, is also a reference task's, and the model's first try told that task again. The near-duplicate check caught it, and the second try passed it.

## What is still to measure

- **The wait in claude.ai.** The author's lesson after the release, at medium effort with a screen recording, times Enter to the task on the card, the share of tasks a card takes ready (T88) and the refusals. At medium effort T85 found typing to be most of the writing, 18.5 s of 25, so the saving there should be the larger share.
- **The families.** `just report` shows the sizes of every hand-in by version and host, and the tasks written ahead; the time from request to acceptance comes from the same lines.
- **Variety, criterion 11.9.** These chats held two tasks a cell; a series of twenty on one topic, as in T80, would show whether the lists keep the ideas apart in practice.
- **Drawings.** Only 2 of the 10 tasks asked for drew. What a drawing costs is still unknown, and R228 is to be judged on numbers that separate it from the topic.

## How to repeat it

The tools are T85's, outside the repository, in `~/mathtrail-play/runs/`: the copy of the driver for this run is `t89/bin/run-chats.sh`, and `tasktime` is T85's. The driver takes the letter of the answer from the last hand-in before `prepare_task`, the task on the card. This run's server was built from T89's working tree before its commit; a repeat builds it from the commit that carries T89, which gives the same version of the instructions while none of the files the version is made of has changed: the instructions, the solver's templates, the drawing frames and the ideas.

```sh
# a clean copy of the commit, a server on a port of its own, the profile by a call,
# one task asked for a chat and one written ahead
git archive <commit> | tar -x -C t89/src
MODEL=claude-opus-5-5 EFFORT=xhigh t89/bin/run-chats.sh 1 10
t89/bin/tasktime-bin streams t89/chats/*-[0-9].tsv > t89/out/chats.jsonl
t89/bin/tasktime-bin writes t89/out/chats.jsonl
(cd t89/src && go run ./cmd/report) < t89/play/server-<time>.jsonl
```

## Cost

- The ten chats: $5.86 by Claude Code's count, $0.59 a chat, from the author's subscription. A chat writes two tasks now, where T85's wrote one for $0.44.
- Cloud: $0; the server ran on the development machine.
