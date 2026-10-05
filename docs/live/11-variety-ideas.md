# New ideas, not only new words (T80)

**A draft, written before the runs.** 2026-10-05. The protocol and the bar are fixed here first, so that the results are read against them rather than the other way round; the results follow once the runs are done.

## What this is

T64a found that the tasks of one topic run out of ideas long before they run out of words ([10-variety](10-variety.md)). A series of twenty held four to seven ideas, and the same task with the same numbers came back unseen by the near-duplicate check. Three tasks of four dressed a reference task their package showed, as the guide asked, and the model's own ideas repeated as well. R205 changes two things at once:
- the guide no longer sends the model to the reference tasks' structure;
- the package names an idea that moves with every task of the topic, out of a list the model writes before it chooses.

This run measures what that does to the ideas of a series, and what it costs in refusals, time and difficulty.

- **Two runs, side by side.** The base runs the guide as it was before R205, at the instructions version `7ed3c64f063d`; the change runs it after, at `da0a46357653`. Both are built from commit `1349b4b` with nothing else between them, and each runs as a service of its own, on a port of its own, with its own profile in memory. The tasks go in turn: the base's task k, then the change's task k. The two runs therefore meet the model at the same hours, and a slow hour of the API weighs on both alike.
- **The pilot changed the guide once.** The first task of each run was a pilot. In the first draft of R205 the list of ideas could hold the reference tasks' own ideas, as the best known of the topic. The model listed four ideas instead of ten, with the three reference tasks' ideas first, and wrote the fourth though the package named the first: the task may be none of the reference tasks. The guide now asks for a numbered list of exactly ten ideas, none of them a reference task's, and for the task at the number and no other. The change's run was then started again from a new profile, and its pilot task followed the number. The first draft's pilot is kept apart and counts in no table below.
- **The cells.** Three series of T64a, the poorest in ideas of each topic, all at difficulty 3:

  | Series | Topic | Level | Reference tasks of the cell | Ideas in T64a |
  |---|---|---|---|---|
  | s2 | Enumeration | `5-6` | 3 | 4 |
  | s3 | Knights and liars | `3-4` | 5, three shown, turning | 5 |
  | s6 | Arithmetic with a trick | `5-6` | 3 | 4 |

  Each series runs to twenty tasks, whatever is refused on the way.
- **The protocol** is T64a's ("How to repeat it" there), with one change: the parent's message. Since T78 the guide for the chat tells the model to save a topic for the lessons when it is asked to keep to one. T64a's message asked exactly that. A topic saved that way would wait for the trial series to end and send its tasks elsewhere, and a series of another topic would then be refused. So the message now asks for one task, naming its topic, level and difficulty:

  > Hi, it's Pip's parent. Pip would like the next task: Enumeration (combinatorics.enumeration), at grade level 5-6, difficulty 3 out of 5. Please ask for exactly that topic, level and difficulty.

  Every chat is checked: `next_task` asked for that topic, level and difficulty as the model's own choice, and `save_profile` was not called after the chat that made the profile. A chat that does otherwise stops the run.
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

- **Blind.** The tasks of both runs are shuffled together. Each labeller sees only the question, the options, the right answer and the solution: no run, no order, no `core_idea` and no `design_thought_process`, since the list of ideas would give the run away. Each task is shown beside the reference tasks of its cell.
- **Two labellers,** working independently. Each gives every task an idea and a difficulty. The bar holds only if it holds by each labelling. If the two disagree on whether a series clears it, the question goes to the author.
- **One idea.** Two tasks share an idea when a child who solved one would solve the other the same way, whatever the story, the names and the numbers. T64a's labels set how fine the ideas are cut: "count the pairs among n", "think of a number, and the steps cancel it".
- **The same task.** Every pair of tasks of a series with the same idea and the same answer is read in full, and counted when its numbers are the same too.
- **The difficulty** is rated from 1 to 5 within the level by the four things of the guide's "The difficulty" (R196): the steps, the search, the conditions and how well the trap is hidden. The labeller is shown the reference tasks of the level at every difficulty it has, not only the three of the cell. A retold reference task of difficulty 3 would otherwise rate 3 just for looking like one.
- **Following the number.** Read from `core_idea` in the change's run alone, after the labels are in:
  - whether the list was written;
  - whether the task is the idea at its number;
  - how much the lists of one series agree.

## Results

To follow the runs.

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

The profile is made by the first chat of each run, with T64a's message. The task's id is in the result of the accepted `submit_task`, and the right letter in what the model handed in. The script that drove the runs is T64a's `series-continue.sh` with two runs in turn, a check of every chat against the protocol, and a series run on to twenty whatever is refused. The comparison is T64a's program, run over each run's attempts. Its labels of ideas came from the blind labelling above, by a key from each anonymous task back to its run and its place. `go run ./cmd/report < <a run's server log>` adds a run up as `just report` adds up a deployment.
