# Drawings by default (T83)

**Measured: drawings come more often, and still in one task of four or five.** 2026-10-06. Forty tasks, each written by Claude's model in a new chat, on ten cells of topics and levels, with the instructions of T83 (R228): a guide that tells the model to draw whenever a task has something to see, 95 reference tasks that draw where 24 did, and a package that always shows one of them where its topic has one at that level.

- **8 of the 32 tasks of topics with frames drew, and none of the 8 controls.** The bar fixed before the run was 16 of 32, and it is not met. In all, 8 of 40 tasks drew. Before T83 the model drew in 9 of 118 tasks of T64a, all of them in one cell, enumeration at `5-6`, which this run did not hold, and in none of a whole lesson of T62 (07, finding 11). The cells of the two runs differ. In the three they share — enumeration, knights and liars and arithmetic at `3-4` — neither run drew.
- **Where it drew:**
  - weighing and pouring, 3 of 4 tasks;
  - clocks, 2 of 3, since the cell's first task went to the trial series;
  - gaps, 2 of 4;
  - games, 1 of 4.

  It did not draw at all in parity, knights and liars, ordering or enumeration.
- **Seven of the eight drawings were right.** Each showed what its question gives and nothing it asks, and none decorated. The eighth, `gaps-34-3`, drew whole the gaps whose count is the task's trap, against the rule that such a row is drawn only cut short.
- **Drawings cost attempts, and no request ran out of them.** 8 of the 21 refused attempts, each in a chat of its own, were refused over a drawing among their reasons. Twice, refused for labels its question did not name, the model handed the task in again without its drawing rather than name the places in the words.

The soft way moved the number, not far enough. The recommendations at the end are for the author to decide.

## What this is

The measurement T83 promised: whether the model now draws much more often, with the instructions and the content of T83 as they ship, and with nothing that forces a drawing.

- **Service.** Built from the working tree of 2026-10-06, with T83 in it and uncommitted; the instructions at version `d42d50d5a122`. Other sessions had uncommitted changes in the tree too, none of them under `internal/`, `content/` or `cmd/`. It ran on this machine as `just play-server` runs it, on port 8081, with the development sign-in, the profile in memory, and the day's ceilings raised to 200 accepted tasks and 30 failed requests.
- **Chat.** Claude Code 2.1.280 in `-p` mode, model `claude-sonnet-5`, as `just play` runs it: no tools of its own, the lesson's tools only. **Every task was asked for in a new chat**, by one message of the parent naming the topic, the level and the difficulty, as in 10-variety.
- **The answer.** The right letter, read from what the model handed in, recorded by `submit_answer` through MCP Inspector with no turn of the model, as the card records it.
- **Profile.** One for the whole run: the pseudonym Pip, grade 4, interests space, football and dinosaurs, lessons in English.
- **The cells,** four tasks each, held in four rounds of all ten, so that every topic met the profile as it grew:

| Cell | Level, difficulty | Reference tasks that draw at that level | What a package shows |
|---|---|---|---|
| Clocks | `1-2`, 3 | 21 of 25 | two or three that draw |
| Weighing and pouring | `3-4`, 2 | 11 of 25 | one or two that draw |
| Gaps and boundaries | `3-4`, 3 | 7 of 25 | one that draws, by turns or by the package's rule |
| Parity and alternation | `3-4`, 3 | 1 of 25 | one that draws |
| Knights and liars | `3-4`, 3 | 2 of 25 | one, of difficulty 4, by the package's rule |
| Ordering | `3-4`, 3 | 1 of 25 | one, of difficulty 2, by the package's rule |
| Games with a winning strategy | `5-6`, 3 | 4 of 9 | one that draws |
| Enumeration | `3-4`, 3 | none | none: frames only |
| Arithmetic with a trick (control) | `3-4`, 3 | none, no frames | none |
| Pigeonhole principle (control) | `3-4`, 3 | none, no frames | none |

- **The trial series took the first task.** In the first chat the model saved the topic to the profile as `lesson_topic` and passed no topic to `next_task`. A topic kept in the profile waits for the trial series of the first five tasks to end, so the rule chose the first task's topic: ordering at `1-2`. In every later chat the model passed the topic itself, and the service took it, in the trial series too. Tasks are counted below by the topic they came on.

## The bar, fixed before the run

| | Bar | Measured | Met |
|---|---|---|---|
| 1 | At least 16 of the 32 tasks of topics with frames draw, and every cell whose package shows a reference task that draws draws at least once | 8 of 32; parity, knights and ordering at `3-4`, whose packages show one, never drew | no |
| 2 | No drawing gives the answer away, takes the key step or decorates, each read by hand | none gives the answer away or decorates; `gaps-34-3` draws whole the gaps its trap is about | no |
| 3 | No drawing without mathematics in the controls | none in the controls | yes |
| 4 | No request runs out of attempts with a drawing's refusal among them | no request ran out of attempts | yes |

## Drawings by topic

`just report` over the run's log, its table "Drawings by topic":

| Topic | Accepted | With a drawing |
|---|---:|---:|
| algorithms.weighing_pouring | 4 | 3 |
| time.clocks | 3 | 2 |
| counting.gaps | 4 | 2 |
| games.strategy | 4 | 1 |
| parity.alternation | 4 | 0 |
| logic.knights_liars | 4 | 0 |
| logic.ordering | 5 | 0 |
| combinatorics.enumeration | 4 | 0 |
| arithmetic.tricks (control) | 4 | 0 |
| pigeonhole.basic (control) | 4 | 0 |
| **All** | **40** | **8** |

The model drew most where the package showed the picture its task needed. All three drawings of weighing and pouring are two empty jugs, the picture of their reference tasks, and the drawings of gaps are rows cut short, like theirs. In clocks it drew the timetable frame both times, never the clock face that 21 reference tasks of the level show. Where a task needed a picture the package did not show — a race to 21, a row of 45 toy models — it drew nothing. Three reference tasks that draw, in the package of `clocks-12-2`, did not make it draw the start of a football match.

## The eight drawings

Each is the drawing as the card showed it.

- **`weighing-34-1`**, two empty bottles of 4 and 3 litres. The answer, 4 steps, is not in the picture.

  ```
     A         B
   │     │   │     │
   │  0  │   │  0  │
   └─────┘   └─────┘
      4         3
  ```

- **`gaps-34-1`**, beacons every 5 metres along a 60-metre tunnel from S to M. Cut short, so the 13 beacons cannot be counted off.

  ```
  S     5          M
  ●─────●───...────●
  ```

- **`games-56-1`**, a corridor of 8 bays. The answer, bays 4 and 5, is not marked.

  ```
  1  2  3  4  5  6  7  8
  ●──●──●──●──●──●──●──●
  ```

- **`clocks-12-3`**, an event from 9:45 to 11:20. The length asked is not drawn.

  ```
  ┌───┬───────┬───────┐
  │ A │  9:45 │ 11:20 │
  └───┴───────┴───────┘
  ```

- **`weighing-34-3`**, two empty tanks of 5 and 3 litres.

  ```
    A       B
  │   │   │   │
  │ 0 │   │ 0 │
  └───┘   └───┘
    5       3
  ```

- **`gaps-34-3`**, footprints 1 to 4 whole, then 10. The stretch asked, from 4 to 10, is cut short. The stretch given, 18 metres from 1 to 4, is drawn whole, and its three gaps can be counted off the picture, though they are the trap of the task: 18 shared by 3 gaps, not by 4.

  ```
  1  2  3  4    10
  ●──●──●──●─...─●
  ```

- **`clocks-12-4`**, two spacewalks with their starts, and their ends marked as unknown.

  ```
  ┌───┬───────┬───────┐
  │ A │   9:45│  ??:??│
  ├───┼───────┼───────┤
  │ B │   9:50│  ??:??│
  └───┴───────┴───────┘
  ```

- **`weighing-34-4`**, two empty jugs of 3 and 5 litres.

  ```
    A       B
  │   │   │   │
  │ 0 │   │ 0 │
  └───┘   └───┘
    3       5
  ```

## Why attempts were refused

61 attempts were handed in for 40 tasks, and 21 were refused. No request used up its three attempts. 21 tasks were accepted at the first attempt, 1.5 attempts a task on average.

| Check | Counted by it | Failed it | Failed it over a drawing |
|---|---:|---:|---:|
| `bad_structure` | 15 | 15 | 3 |
| `drawing_mismatch` | 3 | 5 | 5 |
| `readability` | 3 | 7 | — |
| `solver_error` | 0 | 1 | — |

The first two columns are `just report`'s. The last was read from the transcripts: the eight attempts it counts are eight different attempts, in eight chats.

- **A value written as text.** Three times the structure gave an object's `value` as text — `"0"`, `"5"`, `"9:45"` — where the format takes a whole number. The guide says to give each object its value and does not say that it is a whole number, nor where a time goes.
- **Labels the question does not name.** Five times a drawing declared labels its question did not name. In `ordering-34-2` the places of a boarding queue were numbered 1 to 5, and the question said only "first" and "third". Twice, in `ordering-34-2` and `ordering-34-4`, the next attempt dropped the drawing instead of naming the places in the words. Those are two drawings lost to the check.
- **A brief that came back broken.** Eleven attempts, in nine topics, handed the brief back without `target_concept`, one of them also with a value written as text. One handed back a brief that was not JSON, and one a brief without its `setting`. With the two other values written as text, these are the 15 attempts `bad_structure` counted. The brief is the model copying it by hand, unrelated to drawings, and the most frequent refusal of the run.

The seconds from a request to its accepted task were 141 at the median, 220 at the 90th percentile and 382 at the longest. They are not compared with T64a's 69 at the median: T64a ran other cells under an older guide.

## Recommendations

The levers of T83 lift the share in the topics where the reference tasks and the frames show the kind of picture a task of the cell would have. They do not lift it elsewhere, and they do not make the model draw where it could. In order of cost:

1. **Say it where the model reads first.** The words of `get_package` and step 3 of the server's instructions name the guide and say nothing of a drawing; the rule to draw is in the guide's last third. One sentence in each — "draw whenever the task has something to see, as the guide says" — is a change of words alone.
2. **Spare the drawing its refusals.** Tell the guide that a value is a whole number, and that a time or anything else that is not one stays in the label and the words. Tell it that a drawing refused for its labels is mended by naming the labels in the question, not dropped.
3. **Require a drawing.** This is the step the author set for a run that misses its bar, and R228 left it for after this run. It has two forms. In the first, `brief.constraints` asks for a drawing in a topic with frames, the model may decline with a reason in `rationale`, and the log counts the reasons. In the second, a check refuses a task of such a topic that comes without a drawing, at the price of attempts and of pictures drawn where none helps.

Measuring 1 and 2 takes a run of the same size, about 41 chats of the author's subscription.

## How to repeat it

The driver lives outside the repository. What it does:

```sh
# the service, on a port of its own, with the day's ceilings raised
MATHTRAIL_PLAY_DIR=~/mathtrail-play-t83 PORT=8081 MATHTRAIL_DAILY_TASKS=200 MATHTRAIL_DAILY_FAILED=30 \
  setsid just play-server &

# the profile, once
printf '%s' "Hello! I'm the parent. Please set up maths lessons for my child: pseudonym Pip, grade 4, interests: space, football and dinosaurs. Lessons in English. Please create the profile now." \
  | MATHTRAIL_PLAY_DIR=~/mathtrail-play-t83 just LOCAL_MCP=http://localhost:8081/mcp play -p \
      --output-format stream-json --verbose --session-id "$(cat /proc/sys/kernel/random/uuid)" > profile.jsonl

# then, for each round and each cell, a new chat
printf '%s' "Hi, it's Pip's parent. Pip would like the next task. For a while we're practising one topic only: Clocks (time.clocks), at grade level 1-2, difficulty 3 out of 5. Please keep to exactly that topic, level and difficulty." \
  | MATHTRAIL_PLAY_DIR=~/mathtrail-play-t83 just LOCAL_MCP=http://localhost:8081/mcp play -p \
      --output-format stream-json --verbose --session-id "$(cat /proc/sys/kernel/random/uuid)" > task.jsonl

# the answer, as the card records it: the task's id and the right letter from task.jsonl
just LOCAL_MCP=http://localhost:8081/mcp inspect-cli --method tools/call --tool-name submit_answer \
  --tool-arg task_id=<the task's id> --tool-arg answer=<the right letter>
```

The task's id and its drawing are in the last result of the chat whose `screen` is `task`; the right letter is in the last `submit_task` the model called. A watchdog ends any chat older than 16 minutes; none stalled in this run. `go run ./cmd/report < <the server's log>` gives the counts of the tables above. The drawings, the refusals over a drawing and the verdicts of the bar were read from the chats' transcripts.
