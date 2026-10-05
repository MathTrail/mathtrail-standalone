# New ideas, not only new words (T80)

**Measured: the tasks moved off the reference tasks and a series held more ideas, but criterion 11.9 is still not met.** 2026-10-05. The same protocol ran twice side by side, before and after R205, on the three series of T64a poorest in ideas, 58 tasks each.

What worked:
- The model followed the package's idea number in 57 tasks of 58.
- It dressed a reference task its own package showed in 1 to 4 tasks of 58, by the two labellers, against 37 of 58 before.
- The series held 7 ideas in enumeration (4 to 6 before), 11 to 16 in knights and liars (7 to 8 before) and 11 in arithmetic with a trick (2 to 3 before).
- The same task with the same numbers came back 3 times, where it had come back 22 times.

What did not:
- Arithmetic alone cleared criterion 11.9 as the author set it in О-59. In enumeration the model's favourite came three times: three-digit codes from the digits 1 to 5, ending in an even digit, answer 24. In knights and liars the same chain of accusations came twice.
- The idea number moves the model along a list, but the list is the model's own, written afresh in every chat. Its favourite stands at a different place each time, so a number meant for a new idea lands on the favourite again.
- The median time to write a task doubled, 129 seconds against 66.

The difficulty held to the bar, with more tasks rated easier. What to change next is О-67.

## What this is

T64a found that the tasks of one topic run out of ideas long before they run out of words ([10-variety](10-variety.md)). A series of twenty held four to seven ideas, and the same task with the same numbers came back unseen by the near-duplicate check. Three tasks of four dressed a reference task their package showed, as the guide asked, and the model's own ideas repeated as well. R205 changes two things at once:
- the guide no longer sends the model to the reference tasks' structure;
- the package names an idea that moves with every task of the topic, out of a list the model writes before it chooses.

This run measures what that does to the ideas of a series, and what it costs in refusals, time and difficulty. The protocol and the bar below were written before the runs began.

- **Two runs, side by side.** The base runs the guide as it was before R205, at the instructions version `7ed3c64f063d`; the change runs it after, at `da0a46357653`. Both are built from commit `1349b4b` with nothing else between them, and each runs as a service of its own, on a port of its own, with its own profile in memory. The tasks go in turn: the base's task k, then the change's task k. The two runs therefore meet the model at the same hours, and a slow hour of the API weighs on both alike.
- **The pilot changed the guide once.** The first task of each run was a pilot. In the first draft of R205 the list of ideas could hold the reference tasks' own ideas, as the best known of the topic. The model listed four ideas instead of ten, with the three reference tasks' ideas first, and wrote the fourth though the package named the first: the task may be none of the reference tasks. The guide now asks for a numbered list of exactly ten ideas, none of them a reference task's, and for the task at the number and no other. The change's run was then started again from a new profile, and its pilot task followed the number. The first draft's pilot is kept apart and counts in no table below.
- **After the runs the worked example changed.** The review of this report found the guide's example at odds with the rules it shows: it built a child's first task of a topic on idea 4, where a package names idea 1, and its task was of difficulty 1, not 2. The example is now a task on idea 1 of its list, a runner overtaking the one in second place, and a test holds it to the package. The rules the runs measured are unchanged; the version moves with the text.
- **The cells.** Three series of T64a, the poorest in ideas of each topic, all at difficulty 3:

  | Series | Topic | Level | Reference tasks of the cell | Ideas in T64a |
  |---|---|---|---|---|
  | s2 | Enumeration | `5-6` | 3 | 4 |
  | s3 | Knights and liars | `3-4` | 5, three shown, turning | 5 |
  | s6 | Arithmetic with a trick | `5-6` | 3 | 4 |

  Each series runs to twenty tasks, whatever is refused on the way.
- **The protocol** is T64a's ("How to repeat it" there), with one change: the parent's message. Since T78 the guide for the chat tells the model to save a topic for the lessons when it is asked to keep to one. T64a's message asked exactly that. A topic saved that way would wait for the trial series to end and send its tasks elsewhere, and a series of another topic would then be refused. So the message now asks for one task, naming its topic, level and difficulty:

  > Hi, it's Pip's parent. Pip would like the next task: Enumeration (combinatorics.enumeration), at grade level 5-6, difficulty 3 out of 5. Please ask for exactly that topic, level and difficulty.

  Every chat is checked: `next_task` asked for that topic, level and difficulty as the model's own choice, and `save_profile` was not called after the chat that made the profile. A chat that does otherwise stops the run. None did.
- **The chat.** Claude Code 2.1.280 in `-p` mode, with the model `claude-sonnet-5` and the lesson's tools alone, as `just play` runs it. Every task is asked for in a new chat. The answer is the right letter, recorded by a call of `submit_answer` with no turn of the model, as the card records it. A chat that runs past 15 minutes is ended. Its request is left open for the next chat, and the task is counted as lost to the chat, not to the checks.
- **The profile.** One for each run, made by the same first chat: the pseudonym Pip, grade 4, interests space, football and dinosaurs, lessons in English, as in T64a. The day's ceilings are raised to 200 accepted tasks and 30 failed requests.

## The bar, fixed before the runs

1. **Criterion 11.9,** as the author set it in О-59. It is held for each series of twenty, in each run:
   - no task is lost to the checks;
   - at least 8 ideas;
   - no idea more than 5 times;
   - no task repeats an earlier one with the same numbers and the same answer.

   Attempts refused as near-duplicates are counted and reported, but not held to.
2. **The difficulty is kept.** In the change's run the median rated difficulty is 3. The share of its tasks rated 2 to 4 is no more than 10 points below the base's.
3. **Time and attempts.** In the change's run the median seconds to write a task is no more than 15 % above the base's, and the attempts per task no more than 0.2 above. The task asks for "no more than the base". On sixty tasks a run has noise, and the list of ideas costs a few seconds of writing by itself, so the author agreed this margin with the plan.

Criterion 11.9 is met if every series of the change's run clears point 1. The bar is the same for the base, which shows where the guide started from. If the change's run misses it, the result is reported as it is, and the guide is not tuned again on the same cells within this task.

## How the ideas are counted

- **Blind.** The tasks of both runs are shuffled together. Each labeller sees only the question, the options, the right answer and the solution: no run, no order, no `core_idea` and no `design_thought_process`, since the list of ideas would give the run away. Each task is shown beside the reference tasks of its level.
- **Two labellers,** working independently: two agents, one of them a different model, each given one cell at a time and nothing else to read. Each gives every task an idea and a difficulty. The bar holds only if it holds by each labelling. If the two disagree on whether a series clears it, the question goes to the author. They did not disagree.
- **One idea.** Two tasks share an idea when a child who solved one would solve the other the same way, whatever the story, the names and the numbers. T64a's labels set how fine the ideas are cut: "count the pairs among n", "think of a number, and the steps cancel it".
- **The same task.** Every pair of tasks of a series with the same idea and the same answer is read in full, and counted when its numbers are the same too. In knights and liars the numbers are the puzzle: the same statements under other names.
- **The difficulty** is rated from 1 to 5 within the level by the four things of the guide's "The difficulty" (R196): the steps, the search, the conditions and how well the trap is hidden. The labeller is shown the reference tasks of the level at every difficulty it has, not only the three of the cell. A retold reference task of difficulty 3 would otherwise rate 3 just for looking like one.
- **Following the number.** Read from `core_idea` in the change's run alone, after the labels are in:
  - whether the list was written;
  - whether the task is the idea at its number;
  - how much the lists of one series agree.

## The runs

By each run's own server log, as `go run ./cmd/report` adds it up:

| | Base | Change |
|---|---|---|
| Chats, tasks asked for | 60 | 60 |
| Tasks handed out | 58 | 58 |
| Lost to the checks | 0 | 0 |
| Lost to a chat that stalled | 2 | 2 |
| Attempts handed in, refused | 84, 26 | 91, 33 |
| Accepted at the first attempt | 34 | 27 |
| Attempts per task | 1.45 | 1.57 |
| Seconds to write: median / 90th percentile / longest | 66 / 194 / 883 | 129 / 181 / 631 |
| Refused for the structure, of which a brief without its `target_concept` | 14, 14 | 27, 24 |
| Refused as near-duplicates | 8 | 2 |
| Refused for readability | 6 | 3 |
| Refused by the solver | 0 | 2 |

- **No request ran out of its attempts.** Every task that was lost was lost to a chat whose model went quiet.
  - The base lost two, the 15th of enumeration and the 12th of knights and liars, each ended at its 15 minutes.
  - The change lost the 2nd of enumeration the same way, and the 19th of enumeration to a chat that ran for eight hours. The signal of its time limit never reached it, so the run stood still until the chat was ended by hand.
  - From then on, a watchdog ended any chat of the run older than 16 minutes. It had nothing more to end.
- **The time.**
  - **By series.** The change took a median of 124 seconds in enumeration (100 before), 141 in knights and liars (77) and 121 in arithmetic (47). The list of ten ideas the model writes first is most of that.
  - **In model time.** The change's chats took the model 147 minutes against 89.

## The ideas

By the two labellers, A / B:

| Series | Run | Tasks | Ideas | The idea used most | First task repeating an idea | Dressing a reference task its package showed | The same task again |
|---|---|---|---|---|---|---|---|
| s2 | base | 19 | 4 / 6 | digits arranged under a condition: 8 / 7 | 4 / 5 | 5 / 5 | 6 |
| s2 | change | 18 | 7 / 7 | digits arranged under a condition: 6 / 6 | 4 / 4 | 0 / 0 | 2 |
| s3 | base | 19 | 7 / 8 | a chain of accusations: 8 / 7 | 2 / 2 | 13 / 12 | 11 |
| s3 | change | 20 | 11 / 16 | 4 / 3 | 9 / 9 | 3 / 0 | 1 |
| s6 | base | 20 | 3 / 2 | a number plus a part of it: 17 / 17 | 2 / 2 | 19 / 20 | 5 |
| s6 | change | 20 | 11 / 11 | 4 / 4 | 6 / 6 | 1 / 1 | 0 |

- **The base repeated as T64a found.**
  - Arithmetic gave one reference task, "a number plus half of it plus 5 is 50", 17 times of 20, five times as a repeat with the same numbers.
  - Knights and liars gave two puzzles seven and six times, unchanged but for the names: "A: B lies. B: C lies. C: A and B both lie", and "A: B and C are both knights. B: C lies".
  - Enumeration gave grid paths around a blocked square five times and digits arranged under a condition eight times, four of them the same task.
- **The change kept off the reference tasks and repeated less.** Its repeats with the same numbers were three:
  - enumeration's 5th, 7th and 13th tasks: three-digit codes from the digits 1 to 5, no digit twice, even, 24 — the very task T64a found four times in the same cell;
  - knights and liars' 4th and 9th: the chain of accusations above.
- **The labellers cut the knights differently and agreed on the rest.** B split the puzzles of knights and liars more finely than A, 16 ideas against 11. On every verdict below they agree.

## The bar

| Series of the change | No task lost to the checks | At least 8 ideas | No idea more than 5 times | No task again with the same numbers | Met |
|---|---|---|---|---|---|
| s2, enumeration | yes | no: 7 / 7 | no: 6 / 6 | no: 2 | no |
| s3, knights and liars | yes | yes: 11 / 16 | yes: 4 / 3 | no: 1 | no |
| s6, arithmetic with a trick | yes | yes: 11 / 11 | yes: 4 / 4 | yes | yes |

- **Criterion 11.9 is not met.** One series of three clears it. The base cleared none: its best series held 7 to 8 ideas, the most used of them 7 to 8 times.
- **The difficulty is kept, at the edge.**
  - The change's median rated difficulty is 3 by both labellers.
  - Its share of tasks rated 2 to 4 is 90 % by A against the base's 100 %, exactly the 10 points the bar allows, and 93 % against 95 % by B.
  - The tasks lean easier: a mean of 2.5 against 2.9 by A, and 2.6 against 2.9 by B. In enumeration and arithmetic the change's median is 2. Six of its arithmetic tasks rated 1 are tricks of calculation, such as 98 × 102, or 23 × 48 + 23 × 52 with one multiplication, which the list of ideas reaches for once the reference tasks' ideas are off it.
- **Time is not kept.**
  - The median time to write a task is 95 % above the base's, against the 15 % the bar allows.
  - The attempts per task are 0.12 above the base's, within the 0.2 allowed.

## Following the number

- **The model followed the number.**
  - In 57 tasks of 58 it opened `core_idea` with a numbered list of ten problems and built the task on the one at the package's number. In the round the package named, it wrote the usual form or a variant.
  - Once, in the 3rd task of arithmetic, it took the idea above the number.
- **The lists do not agree from chat to chat.** In enumeration an idea of numbers or codes made from given digits stood on 18 lists of 18, at places from 2 to 10, and on some lists twice. A child's tasks 5, 7 and 13 were therefore the same code, under three different numbers.
- **The list still holds reference tasks' ideas sometimes.** In the 5th task of enumeration the model opened its list with the ideas of the cell's three reference tasks, against the guide; the number, 4, led past them to the codes.

## By words

The highest similarity of each accepted task to any task accepted before it on the profile, by the exact measure of the near-duplicate check, as in T64a:

| Series | Run | Highest | Mean of each task's highest | Tasks at 0.6 or more | First attempt refused as a near-duplicate |
|---|---|---|---|---|---|
| s2 | base | 0.67 | 0.41 | 2 | task 5 |
| s2 | change | 0.51 | 0.31 | 0 | task 14 |
| s3 | base | 0.74 | 0.54 | 6 | task 8 |
| s3 | change | 0.66 | 0.51 | 6 | task 17 |
| s6 | base | 0.70 | 0.49 | 4 | none |
| s6 | change | 0.50 | 0.26 | 0 | none |

The change's three repeats with the same numbers passed the check: the three codes, two of them in the same setting, at 0.37 to 0.47 from each other, and the chain of accusations at 0.43. What T64a found about the measure stands. It sees words, and a repeat in a new setting slips under it.

## Other findings

1. **A brief handed back without its topic grew more common with the change.** It happened 24 times against 14, as the model writes more before it hands the brief back. Each costs an attempt and about a minute. The author kept `target_concept` required in О-59, and this is new evidence for the alternative T64a raised: taking the topic from the request.
2. **The brief's traps never change within a series when the child answers right.** The rule takes the child's own traps first, and a child who makes no mistakes has none. So every brief of a series named the same two traps, those the reference tasks of the cell use most. Whether that pulls the tasks towards one idea is not measured here. The run did not change the rule.
3. **A stalled chat can outlive its time limit.** One chat of the change ran eight hours. A run of this kind needs a watchdog of its own beside the limit of each chat.

## What it means

- **The reference tasks are no longer what a series repeats.** Without the instruction to take their structure, the model left them: in 4 tasks of 58 at most, against 37.
- **The repeats now come from the model's own list.** The number does move the model along the list it is given. But each new chat writes the list afresh, and the model's favourites, codes from the digits 1 to 5 above all, stand at a different place of it every time. A number therefore cannot promise a new idea, and the same task came back three times in one series.
- **The promise "the tasks do not run out" holds better than before, and still not as criterion 11.9 asks.** Over the three series the ideas doubled, and the same task with the same numbers came back 3 times instead of 22. One series of three met the criterion.
- **It costs time.** A child waits about two minutes for a task instead of one, and the list is why. The product is in development, and the author merged R205 before this report. The question of whether the list stays is the first of О-67.

## What to change next

О-67, for the author. The options, the most promising first:

1. **A list of the content's own.** Ten ideas for each topic and level, written once and reviewed, in the catalog. The package then names the idea itself rather than a place on a list the model writes. It is the same idea in every chat, no favourite can come round under another number, and the model no longer spends a minute writing the list. The cost is content work: a line for each idea of each cell the catalog teaches. It needs no solvers and no reference tasks.
2. **The model's list, kept.** The first chat of a topic hands its list in, the profile keeps it, and later packages show it back. This is a new field of `submit_task` and of the profile, with a schema and a rollout, and the list keeps the model's favourites.
3. **A check against the same task with the same numbers.** The numbers of a question and its answer are kept for the child's past tasks of a topic, and an exact match is refused. It is cheap, and catches the identical tasks this run and T64a found. It does nothing for an idea repeated with other numbers.
4. **No list.** The rest of R205 stays and the list goes. The time comes back and the reference tasks stay off, and the ideas are left to the model's favourites, as in the base of this run.

The recommendation is 1. This run shows the number does its job when the list under it holds still. The list is the part that does not hold still, and the part that costs the time.

## Cost

- **Estimated before the start:** about 125 calls of `claude -p`, about $20 in API terms, about four hours, and $0 of cloud.
- **Actual:**
  - 124 chats: two profiles, a first pilot of the change in two chats, and sixty in each run;
  - $24.38 in API terms by Claude Code's count: $10.29 for the base, $13.85 for the change and $0.24 for the first pilot;
  - 238 minutes of the model;
  - 13 hours of the clock, from 03:12 to 16:10 UTC, eight of them the stalled chat;
  - $0 of cloud.
- **The labelling:** six runs of an agent, each reading one cell's tasks once, on the same subscription.

## How to repeat it

Two copies of one commit side by side, one of them with the change laid over it, each serving its own run:

```sh
# the base and the change, each from its own copy of the commit
MATHTRAIL_PLAY_DIR=~/mathtrail-play/runs/variety-ideas/base PORT=8080 \
  MATHTRAIL_DAILY_TASKS=200 MATHTRAIL_DAILY_FAILED=30 just play-server      # in the base's copy
MATHTRAIL_PLAY_DIR=~/mathtrail-play/runs/variety-ideas/after PORT=8081 \
  MATHTRAIL_DAILY_TASKS=200 MATHTRAIL_DAILY_FAILED=30 just play-server      # in the change's copy

# a chat of one run: the profile first, then one chat a task
printf '%s' "Hi, it's Pip's parent. Pip would like the next task: Knights and liars (logic.knights_liars), at grade level 3-4, difficulty 3 out of 5. Please ask for exactly that topic, level and difficulty." \
  | MATHTRAIL_PLAY_DIR=~/mathtrail-play/runs/variety-ideas/after just LOCAL_MCP=http://localhost:8081/mcp \
      play -p --output-format stream-json --verbose --session-id "$(cat /proc/sys/kernel/random/uuid)" > task.jsonl

# the answer, as the card records it
just LOCAL_MCP=http://localhost:8081/mcp inspect-cli --method tools/call --tool-name submit_answer \
  --tool-arg task_id=<the card's task id> --tool-arg answer=<the right letter>
```

- **The profile** is made by the first chat of each run, with T64a's message.
- **The task's id** is in the result of the accepted `submit_task`, and the right letter in what the model handed in.
- **The driver** was T64a's `series-continue.sh`, extended in four ways:
  - two runs in turn;
  - a check of every chat against the protocol;
  - each series run on to twenty whatever is refused;
  - a watchdog for chats past their time.
- **The comparison** is T64a's program, run over each run's attempts.
- **The labels of ideas** came from the blind labelling above, by a key from each anonymous task back to its run and its place.
- **A run's totals:** `go run ./cmd/report < <a run's server log>` adds a run up as `just report` adds up a deployment.
