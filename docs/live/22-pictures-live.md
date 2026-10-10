# The pictures, live (T70.6)

**Measured: the pictures the card draws cost less than the text drawings they replaced, though most tasks now draw their solution too, and every clock shows its time. The one new cause of refusals came from the rule that keeps a total under its picture, and the guide now states that rule.** 2026-10-10. The four lessons of 2026-10-08 ([21-crooked-drawings](21-crooked-drawings.md)) were held again, message for message, on the version with pictures drawn by the card (T70.1–T70.4) and the card of how an answer went (T70.5). A fifth lesson walked the card's own path.

- **Cheaper, on fewer tokens.** The 19 turns cost $7.47 against $7.78. The model wrote 6 % fewer output tokens (91,649 against 97,843) and thought 16 % less (27,100 thinking tokens against 32,094). Most tasks now also draw their solution, 22 of 37, so the pictures pay for the card of how an answer went and leave change.
- **Faster, mostly by the day.** A task reached the card 27.9 s after the message on average, against 45.5 s, and a turn took 50.7 s against 76.5 s. The API streamed 98 output tokens a second against 69 on 2026-10-08, though. At that speed the old run would have taken about 32 s to a task, so the change itself saved some 4 s. Tokens and money are what compare across days.
- **Every clock shows its time.** The card draws the hands from the time the model gives. All 9 clock faces of the tasks show the time their question gives, and all 3 of the solutions show the answer. On 2026-10-08, 5 of the 10 clock faces pointed a hand at the wrong number.
- **The drawing's refusals are gone, and the total's rule made new ones.**
  - 9 of the 46 hand-ins were refused, against 7 of 45. Labels and values written as text made 6 of the 7 on 2026-10-08, and now pictures made 2.
  - Five refusals were a total handed in with no picture of the solution, which T70.5's format refuses. The guide now says outright that a task with no picture of the solution has no total. A re-check of lesson 1, where three of them fell, gave none in 11 hand-ins.
  - The child waited for a second attempt at 3 of the 19 tasks asked for, against 6.
- **The picture of the solution.**
  - 22 of the 37 tasks handed in drew one, 17 of them with the total under it.
  - It costs about 90 output tokens, $0.002. The guide's section on it adds 468 tokens to every package, about $0.005 with its reads from the cache. The card's result puts some 730 tokens into the chat.
  - All of it together is a few cents a lesson.
- **The card's path works.** Asked by the card to go over the answer, the model called `show_result` both times: once with the card's line, and once with the message alone. After an answer typed in the chat it recorded the answer with `submit_answer` 16 times out of 16, and explained in words without `show_result`. Claude Code draws no cards, and the run's session tells the model so.
- **Colours and flags.** Asked for flags of coloured stripes, the model drew a flag in the three colours the question names, in its own words. As the picture of the solution it drew all six flags, grouped by the top stripe, with the total 2 + 2 + 2 = 6.

## What this is

T70.6 of RUN.md. The run checks the pictures and the card of how an answer went in a chat before the release that ships them, v0.5.0. It measures what they change in tokens, refusals and the time to a task, and what the pictures of the solution cost. The author's own lesson in Claude, on Android and on the web, comes after the rollout. ChatGPT is T63's.

## The run

- **Service.** The tree of T70.5 before its commit, on `95dc6b9`, built into a binary and served as `just play-server` serves it:
  - port 8093, the development sign-in and the profile in memory;
  - the day's ceilings raised to 200 accepted tasks and 30 failed requests, as on 2026-10-08;
  - instructions at version `4b2b08d344de`. The re-check used `5009108d9b49`, the same instructions with the guide's sentence on the total.
- **Chat.** Claude Code 2.1.280 in `-p` mode, through `just play`. The model was `claude-opus-5-5` at `--effort medium`, as on 2026-10-08.
  - `PLAY_TOOLS` gained `show_result`. Without it the model could not have called the new tool. A test now holds the list to the tools the model is shown.
- **Profile.** One for the whole run, as on 2026-10-08: the pseudonym Тест, grade 3, space and football, lessons in Russian.
- **Lessons 1 to 4.** The messages of 2026-10-08 word for word, and the same drivers, each lesson one chat:
  - every message asked for a task with a picture and named its topic, level and difficulty;
  - from the second message on, it opened with the child's answer, the right letter;
  - the last task of each lesson was answered by a call to `submit_answer`, as its card would answer it.
- **Lesson 5, the card's path,** which Claude Code cannot draw, three tasks in six messages:
  - **The first task** was asked for as flags of coloured stripes. The card answered it wrong, by a call to `submit_answer` as the card makes it. The next message carried the line the card hands the model, word for word as the widget writes it, ahead of the card's words «Разбери ответ».
  - **The second task** came by «Другая задача», the button on the card of how an answer went. The card answered it right, and the next message was «Разбери ответ» alone, as from a host that keeps no line.
  - **The third task** came the same way, and the adult wrote that the child did not know.
- **Re-check.** Lesson 1 again, on a server of its own with a new profile, once the guide said a task with no picture of the solution has no total.

| Lesson | Tasks asked for |
|---|---|
| 1 | `time.clocks`: `1-2` at difficulties 2 and 4, `3-4` at 2 and 4, `5-6` at 3 |
| 2 | `algorithms.weighing_pouring` `3-4`, `counting.gaps` `1-2`, `logic.ordering` `3-4`, `parity.alternation` `3-4`, `combinatorics.enumeration` `3-4` |
| 3 | At `5-6`: `geometry.grid`, `logic.sets`, `ratio.sharing`, `fractions.parts`, `games.strategy` |
| 4 | `time.calendar` `3-4`, `pigeonhole.basic` `3-4`, `logic.knights_liars` `3-4`, `number.divisibility` `5-6` |
| 5 | `combinatorics.enumeration` `3-4` at difficulty 2, flags of coloured stripes; then two tasks by «Другая задача» |

**Cost and time.** Claude Code counted $7.47 for lessons 1 to 4, $0.83 for lesson 5 and $1.91 for the re-check. Counting tokens with the tokenizer took about $0.2 more. In all that is some $10.4 of the author's subscription, against the $8 the plan estimated. The chats ran for 16.4, 1.9 and 3.5 minutes. No chat stalled, and the watchdog ended none.

## Against the run of 2026-10-08

Lessons 1 to 4, 19 turns each. A turn's tokens and cost are Claude Code's own count. The time to a task runs from the message to the moment the service accepted the task it asked for, by the service's log.

| | 2026-10-08 | 2026-10-10 | Change |
|---|---:|---:|---:|
| Cost, $ | 7.78 | 7.47 | −4 % |
| Output tokens | 97,843 | 91,649 | −6 % |
| of them thinking | 32,094 | 27,100 | −16 % |
| Written to the cache, a turn | 27,225 | 25,995 | −5 % |
| Read from the cache, a turn | 441,305 | 442,511 | 0 % |
| A turn, mean (median), s | 76.5 (67.5) | 50.7 (50.6) | −34 % |
| To the task on the card, mean (median), s | 45.5 (46.2) | 27.9 (27.9) | −39 % |
| Output tokens a second of the API | 68.8 | 98.1 | +43 % |
| Hand-ins | 45 | 46 | |
| Refused | 7 | 9 | |
| Tasks asked for that took a second attempt | 6 of 19 | 3 of 19 | |
| A hand-in, bytes (the task in it) | 2,640 (1,951) | 2,402 (1,758) | −9 % |

By lesson:

| Lesson | Cost, $ | Output tokens | Thinking tokens | A turn, s |
|---|---:|---:|---:|---:|
| 1, clocks | 2.02 → 1.88 | 24,034 → 22,071 | 8,071 → 5,005 | 97.0 → 40.7 |
| 2 | 2.01 → 2.00 | 23,325 → 23,814 | 7,467 → 6,774 | 51.5 → 53.5 |
| 3 | 2.08 → 2.06 | 26,154 → 25,646 | 8,565 → 8,927 | 53.9 → 59.8 |
| 4 | 1.67 → 1.53 | 24,330 → 20,118 | 7,991 → 6,394 | 110.4 → 48.3 |

- **The clock lesson thought least.** It thought 38 % less, with no slash to aim at a number. The re-check of the same lesson thought for 6,070 tokens, so one lesson differs from the next by a fifth. Only the four lessons together say much.
- **Speed.** The API's speed is the lesson's whole output over its time in the API, by Claude Code's count. It was faster by 43 % on the day of this run, so the times above are mostly the day's. The share of the change is an estimate: the old run's 45.5 s, at today's speed, would be about 32 s against the 27.9 s measured.
- **One run on each side.** Both days ran the same messages once, so a difference of a few per cent is within what two runs of one day differ by.

## Refusals

| Cause | 2026-10-08 | 2026-10-10 |
|---|---:|---:|
| A label the question and the drawing do not share (`drawing_mismatch`) | 4 | 1, in the picture of the solution |
| A picture that shows the answer (`drawing_mismatch`) | — | 1 |
| A value written as text (`bad_structure`) | 2 | — |
| A total with no picture of the solution (`bad_structure`) | — | 5 |
| A sentence too long (`readability`) | 1 | 2 |
| A solver that does not parse (`solver_error`) | — | 1 |
| Hand-ins refused | 7 of 45 | 9 of 46 |

One hand-in had two causes, a total and a sentence.

**The total.** The format reads a total only under a picture of the solution (R276), and a total alone is refused. The model wrote one alone in five hand-ins: three for tasks written ahead and two for tasks asked for, three of the five in the clock lesson. Each was an equality of the solution, such as `35 + 20 = 55`, `78 + 78 + 24 = 180` or `5 × 4 = 20`, for a solution it had judged to have nothing to draw. The guide said to "leave both out when the solution has nothing to see", but did not say that the total goes with the picture. It now does: "the total stands under its picture and nowhere else, so a task with no `solution_picture` has no `solution_total` either". The re-check of lesson 1 handed in 11 tasks with no total refused, and its only refusal was a long sentence. It drew 8 pictures of the solution, against 3 in the same lesson before, and wrote one total.

A total that stands alone on the card, with no picture above it, would end this refusal for good. It would also change the card the author chose in T70.5's planning, so it is left to the author, as О-72 (below).

## The picture of the solution

Lessons 1 to 4:

| | Handed in | With a picture | With a picture of the solution | With a total |
|---|---:|---:|---:|---:|
| Asked for | 19 | 17 | 11 | 8 |
| Written ahead | 18 | 14 | 11 | 9 |

- **By kind.** Row 4, bars 4, clock 3, grid 2, Venn 2, piles 2, table 2, containers 1, balance 1, calendar 1.
- **By topic.** Every task of weighing and pouring, gaps, fractions, strategy games, the grid, sets, divisibility, ratio and the calendar drew one, and one of the two on parity. None of enumeration, knights and liars, ordering and the pigeonhole principle did, and 3 of the 9 clock tasks did. In lesson 5, the enumeration of flags drew all six of them.
- **Asked for without a picture.** Two tasks asked for "with a picture" came without one: one on parity and one on socks in the pigeonhole principle. The guide draws only where the picture shows something (R228), and the model judged these two to show nothing. On 2026-10-08 it drew in all 19.
- **The totals.** 14 of the 17 end in the number of the answer, as the format holds them to. The other three are clock tasks whose answer is a time. `8 + 3 + 1 = 12` adds hours up to 12:00, and `45 + 15 = 60` and `15 + 30 + 15 = 60` add minutes for 5:00 and 4:30. The format holds only a number or a label to the end of the total, so a time passes. Whether a total should end in the time is a question for the guide (below).

**What it costs.** Counted on `claude-opus-5-5`'s tokenizer the way the study counted ([widget-drawings](../widget-drawings.md), "Tokens"), at that study's prices. A price is the run's own: output at $20.09 a million tokens, a write to the cache at $8.01 and a read from it at $0.20.

| | Tokens | Cost |
|---|---:|---:|
| A picture of the solution with its total, as the model writes it (the 22 of lessons 1 to 4 and the flags of lesson 5: 2,080) | 90 | $0.0018 |
| A task's picture, for comparison (the 31 of lessons 1 to 4 and the flag of lesson 5: 1,703) | 53 | $0.0011 |
| The guide's section on the picture of the solution, in every package | 468 | $0.0037 written, $0.0015 read back 16 times |
| The result of `show_result` the model reads, the flags task's | 731 | $0.0059 written |

A turn cost $0.39 on average, so the pictures of the solution, their guide and their card cost about 3 % of a turn where all three come. The study expected 51 tokens for a task's picture, and the run's were 53.

## The card of how an answer went

- **After an answer typed in the chat,** the model called `submit_answer` 16 times out of 16, before it said anything of the answer. It never called `show_result`, and explained the answer in words. It had told the adult in lesson 2 that this chat shows no cards: "В этом чате карточки не показываются". The session of `just play` says that the chat draws no cards, and `show_result` draws one. Whether Claude's own model calls it after an answer typed where cards are drawn, as the instructions say, is for the author's lesson.
- **After an answer on the card,** asked to go over it, the model called `show_result` both times:
  - wrong, with the card's line: 11.2 s for the turn, the trap and the solution step by step beside the card;
  - right, with «Разбери ответ» alone: 5.5 s, the model finding the task on the card by itself.
- **The task written ahead.** «Другая задача» handed out the task written ahead: it was on the card 2.6 s and 2.7 s after the message.
- **"We don't know"** in the chat was recorded as `?`, and the solution was gone through.

## Not measured

- **Claude itself, on the web and on Android.** The card, its ask to go over the answer as Claude sends it, and the line the card hands the model. The run put that line in the message as text, since Claude Code has no way for a card to hand it over. That is the author's lesson after the rollout.
- **Whether the model calls `show_result` after an answer typed in the chat** where cards are drawn.
- **ChatGPT** is T63's.
- **Pictures on a real phone.** The cards of the run were drawn server-side by the widget's own components for the gallery below, and looked at in Chromium on a computer.

## Open

Both are О-72 of PRODUCT-V1.

1. **A total alone.** Should the card show a total when the solution has no picture? It would end the refusal measured above, at the price of a card the author did not choose.
2. **A total for a time.** For a time, should the total end in the time, or may it add up minutes, as `45 + 15 = 60` for 5:00 does?

## The gallery

Every task the chats put on the card is in the gallery, 27 of them with the re-check: the card as it is handed out, beside the card of how the answer went, drawn server-side by the widget's own components at a phone's width. It is at https://claude.ai/artifact/GmgbNL2FgMpQLCX2hcsY8e (private to the author). Drawing it showed that the colour naming a group of flags had no outline, so a white one vanished on a white card. It is outlined now, as the key's colours are.

## How to repeat

The drivers are kept outside the repository, in `~/mathtrail-play/runs/t706/` on the author's machine: the drivers of 2026-10-08, with each line of a chat's stream stamped as it arrived, a fifth lesson that answers on the card, the scripts that count the numbers and draw the gallery, and the messages of the five lessons.

```sh
# the service, from the tree, into bin/, which git ignores, on a port of its own
go build -o bin/server ./cmd/server
PORT=8093 MATHTRAIL_LOG_FORMAT=json MATHTRAIL_DEV_AUTH=true MATHTRAIL_DAILY_TASKS=200 MATHTRAIL_DAILY_FAILED=30 \
  MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" setsid ./bin/server > <run>/server.jsonl &

# each message of a lesson, the first with --session-id and the rest with --resume of the same id
printf '%s' "Здравствуйте! Я родитель … Давайте задачу на часы (time.clocks), уровень 1–2, сложность 2 из 5, и обязательно с картинкой часов." \
  | MATHTRAIL_PLAY_DIR=<run>/play MATHTRAIL_PLAY_MODEL=claude-opus-5-5 just LOCAL_MCP=http://localhost:8093/mcp play -p \
      --output-format stream-json --verbose --effort medium --session-id <uuid> > l1-1.jsonl

# the card's answer: submit_answer as the card calls it
curl -sS http://localhost:8093/mcp -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"submit_answer","arguments":{"task_id":"<id>","answer":"A"}}}'
```

- **The numbers.** A turn's tokens, thinking and cost are in the `result` line of its stream. A resumed chat counts its cost as a running total of the session. A task's acceptance is `task_accepted` in the service's log.
- **The gallery.** A script loads `StaticTask` and `StaticResult` through Vite's server-side rendering, as the site's build does. Each payload is read with the widget's own `readHandedTask` and `readAnswer`, and drawn with `preact-render-to-string`. The page gets the widget's stylesheet and tokens.
