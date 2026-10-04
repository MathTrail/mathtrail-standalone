# Do the tasks run out? (T64a)

**Measured: the words do not run out, the ideas do.** 2026-10-04. Six series of tasks, each on one topic at one level, every task written by Claude's model in a new chat, on one profile. The near-duplicate check let a whole series of twenty through in one of the six; in the other five it refused its first attempt at the 4th to the 8th task. No task was lost to it: run on to twenty, the series handed out 118 tasks of 120, the two missing ones lost to chats that stalled, and no request used up its three attempts. Read for their mathematics, the tasks repeat far sooner. A series of twenty held four to seven ideas, the most used of them six to ten times, and the first idea came back by the 2nd to the 4th task. The same task — the same numbers, the same answer — came back again and again at a similarity of 0.26 to 0.68, below the threshold.

The six series held 32 ideas among their 118 tasks. Three tasks of four dress one of the reference tasks their package showed in the child's interest, which is what the guide asks for. The check refuses a near-copy of a reference task, and a repeat of the child's own task only when it comes back in a setting that idea already had: it refused 10 such attempts, while 61 accepted tasks repeated an idea in a setting it already had. A lower threshold would catch more repeats in some topics only by refusing new tasks in others. Criterion 11.9 is not met as written. What to change first is the reference tasks' part in the package: the threshold cannot do it on its own, and the settings are not the cause (О-59).

## What this is

Criterion 11.9 of PRODUCT: the promise of its section 1 that "the tasks do not run out", measured. Tasks are asked for one after another on one topic at one level, on one profile, until the near-duplicate check refuses one or there are twenty, on three topics of different kinds at two levels each.

- **Service.** Built from commit `55dd676`, the instructions at version `e676a02096bd`. T72.10 has changed the guide since (R196), so a later run that measures a change compares with a run of the version that change starts from, not with these numbers alone. It ran on this machine with the development sign-in and the profile in memory, as `just play-server` runs it. It was built from a clean copy of the commit rather than from the working tree, where other tasks had changes under way. For the run, the day's ceilings were raised to 200 accepted tasks and 30 failed requests (`MATHTRAIL_DAILY_TASKS`, `MATHTRAIL_DAILY_FAILED`): a series of twenty is a whole day's allowance by itself, and T64.3 measured the ceilings.
- **Chat.** Claude Code 2.1.280 in `-p` mode, model `claude-sonnet-5`, as `just play` runs it: no tools of its own, the lesson's tools only. **Every task was asked for in a new chat**, by one message of the parent naming the topic, the level and the difficulty. The model passed all three to `next_task` with a reason, so every request was the model's choice (`tutor_mode` `llm`). A parent who asks for one topic gets the same today, and so will a topic chosen on the card (T78). A new chat is the strict case: the model sees none of the tasks it wrote before, and only the service remembers them. That is how a child meets a topic practised over weeks, in a chat a day.
- **The answer.** The right letter, read from what the model handed in, recorded by a call of `submit_answer` with no turn of the model, as the card records it; the first task's was given in the chat. Answering matters: the reference tasks a package shows and its setting turn with the number of answers, so a run of skipped tasks would have shown the model the same package twenty times.
- **Profile.** One for the whole run: the pseudonym Pip, grade 4, interests space, football and dinosaurs, lessons in English. The series ran in turn on it, so each task was compared with every task before it.
- **Why English.** Both halves of the check apply only there: against the reference tasks, which are in English, and against the child's past tasks. The same task with new numbers measures 0.85 in English and 0.74 in Russian (R56), so in English a repeat is caught rather than slipping just under 0.7.
- **The cells.** Difficulty 3 at `3-4` and at `5-6`, for each topic. A cell of `3-4` holds five reference tasks, and a package shows three of them, turning with each answer. A cell of `5-6` holds three, and every package shows the same three in a turning order.
- **The rule of a series.** As the plan has it, a series stops at its first refused attempt. Every series that stopped was then run on to twenty, within the 120 tasks approved for the run: the stop gives the number criterion 11.9 asks for, and the run-on shows whether the stream dries up.
- **Measured beside the service.** For each task: how alike it is to every task before it, by the exact measure the check approximates for past tasks (`checks.Similarity`); the check's own verdict on the profile's sketches (`checks.NearDuplicate`), which agreed with the service on all 164 attempts; and its mathematical idea, labelled by hand. The questions were read from the chats' transcripts.
- **Roles.** The author chose the protocol: local, a new chat per task, English, three topics at two levels. Claude ran it, read the log and the transcripts, labelled the ideas and wrote this report.

## The six series, to the first refusal

| Series | Topic | Level, difficulty | Reference tasks of the cell | Tasks before the first near-duplicate | What the refused attempt repeated |
|---|---|---|---|---|---|
| s1 | Enumeration | `3-4`, 3 | 5, three shown, turning | none in 20 | — |
| s2 | Enumeration | `5-6`, 3 | 3, the same three | 6 | the child's first task of the series, the same idea in the same setting: 0.66 exact, 0.7 or more by the sketch |
| s3 | Knights and liars | `3-4`, 3 | 5, three shown, turning | 5 | a reference task, nearly word for word: 0.71 |
| s4 | Knights and liars | `5-6`, 3 | 3, the same three | 3 | a reference task, nearly word for word: 0.71 |
| s5 | Arithmetic with a trick | `3-4`, 3 | 5, three shown, turning | 7 | a reference task, nearly word for word: 0.71 |
| s6 | Arithmetic with a trick | `5-6`, 3 | 3, the same three | 6 | the child's first task of the series, the same story and steps: 0.73 |

In s2, s4, s5 and s6 the model's next attempt was accepted, each time with a different idea. In s3 the chat stopped after the refusal: the model's stream slowed to a trickle and then went quiet, and the run's limit of 15 minutes a chat ended it. The request was left open, and the next one replaced it once its wait had run out, as designed. That loss is the chat's, not the check's.

## Run on to twenty

| Series | Tasks handed out | Attempts refused as near-duplicates | Other refusals | Seconds to write: median / mean / longest | Drawings |
|---|---|---|---|---|---|
| s1 | 20 | 0 | readability 4, structure 3 | 58 / 61 / 96 | 0 |
| s2 | 19 of 20, one chat stalled | 2 | structure 4 | 87 / 160 / 578 | 9 |
| s3 | 19 of 20, one chat stalled | 6 | readability 5, structure 1, solver 1 | 71 / 112 / 248 | 0 |
| s4 | 20 | 6 | structure 7 | 136 / 127 / 213 | 0 |
| s5 | 20 | 2 | structure 2, readability 1 | 38 / 47 / 88 | 0 |
| s6 | 20 | 4 | structure 3 | 63 / 70 / 234 | 0 |
| **All** | **118 of 120** | **20** | **structure 20, readability 10, solver 1** | **69 / 95 / 578** | **9** |

- **Nothing ran dry.** 119 requests were opened, 164 attempts handed in and 46 refused, and no request used up its three attempts. 76 tasks were accepted at the first attempt, 1.4 attempts a task on average. The second stalled chat, in s2, never handed anything in: its request was taken over by the next chat, whose task therefore counts the eight and a half minutes the stalled chat had held it besides its own minute: 578 seconds, the longest time of the run.
- **What the 20 refused attempts were.** 13 resembled only a past task of the child's, 6 only a reference task, and one both. Read by hand, 15 repeated an idea the check had in view: a past task of the same idea, or the reference task it copied. The other 5, all in knights and liars, were different puzzles, refused for the rules paragraph, the names and the setting they shared with an earlier task, or, once, with a reference task.
- **A past task is met in its own setting.** Every one of the 14 refusals against a past task met a task of the same setting: 10 repeated its idea, and 4, in knights and liars, were a different puzzle. With three interests turning one per answer, the setting comes back every third task, and a repeated idea in the same dress is what the check can see. Most of those it did not see: 61 accepted tasks repeated an idea of their series in a setting that idea already had.
- **The other refusals.** Structure: 20 attempts, 17 of them a brief handed back without its `target_concept` (finding 2). Readability: 10 attempts, all at `3-4`, nine by a Flesch–Kincaid grade of 6.1 to 7.8 against 6 and one by a sentence of 33 words against 25, each fixed at a later attempt. One solver failed to run, and the next attempt's did.
- **The time to write** runs from the request to its acceptance, the attempts after the first included. The median is 69 seconds, the prototype's own number; 81 tasks of 118 came within 90 seconds. At `3-4` the median is 59 seconds and at `5-6` 81. Knights and liars at `5-6` took the longest, a median of 136 seconds.

## How alike the tasks came out

### By words: what the check measures

The highest similarity of each accepted task to any task accepted before it on the profile, by the exact measure:

| Series | Highest | Mean of each task's highest | Tasks at 0.6 or more |
|---|---|---|---|
| s1 | 0.66 | 0.36 | 2 of 20 |
| s2 | 0.68 | 0.50 | 4 of 19 |
| s3 | 0.68 | 0.53 | 6 of 19 |
| s4 | 0.66 | 0.54 | 3 of 20 |
| s5 | 0.61 | 0.36 | 1 of 20 |
| s6 | 0.68 | 0.50 | 5 of 20 |

Against the reference tasks of its level, no accepted task came above 0.68. By words, the tasks the check passed sat at 0.3 to 0.6 from each other, with no wall in sight.

### By ideas: what a child notices

Each accepted task was labelled by hand with its mathematical idea, such as "count the pairs among n", "pick two, not both of one kind" or "think of a number, and the steps cancel it".

| Series | Tasks | Ideas | The idea used most | The first task repeating an idea | Reference tasks of the cell never used |
|---|---|---|---|---|---|
| s1 | 20 | 6 | count the pairs among n, with or without an extra game: 8 | 4 | 0 of 5 |
| s2 | 19 | 4 | paths on a grid around a blocked square: 9 | 2 | 1 of 3 |
| s3 | 19 | 5 | "A: B lies. B: C lies. C: A and B both lie": 8 | 3 | 1 of 5 |
| s4 | 20 | 7 | "if you asked B, she would say that C is a liar": 6 | 3 | 0 of 3 |
| s5 | 20 | 6 | think of a number, undo two steps: 10 | 4 | 0 of 5 |
| s6 | 20 | 4 | think of a number, and the steps cancel it: 8 | 2 | 1 of 3 |

86 of the 118 tasks (73 %) dress a reference task their own package showed. The guide of the run asked for this: "take their structure and how hard they are, never their story or their numbers". The model takes the structure, that is, the idea, and changes the story and the numbers. 13 more take the idea of a reference task their package did not show, classics the model reaches for by itself: the pairs among n came three times in s1 with no reference task of pairs in sight, and the sum of a run four times in s6. The other 19 have ideas no reference task of the cell has, such as "how many even three-digit numbers from these digits".

The same task kept coming back, and the check saw none of it:

| Tasks | The task | Similarity |
|---|---|---|
| s1, 4th and 19th | five astronauts, each pair takes a photo or makes a call: 10 | 0.40 |
| s2, 7th, 8th, 11th and 18th | three-digit codes from the digits 1 to 5, no digit twice, even: 24; the 8th and the 18th both an airlock's code | 0.26 to 0.64 |
| s5, 9th and 12th | a number doubled, then 9 added, gives 29 | 0.43 |
| s5, 14th and 20th | a number tripled, then 9 added, gives 30; dinosaurs both times | 0.30 |
| s6, 6th and 9th | a number, a third of it and 8 more make 48; football both times | 0.68 |
| s6, 4th and 11th | add 8, triple, take 6, divide by 3, take the number away | 0.49 |
| s4, 8th and 14th | the same puzzle of three speakers on Dino Island, with other names | 0.64 |

In knights and liars a repeated idea is the same puzzle: the same statements under other names, with the same answer.

**No threshold on this measure separates a repeat from a new task cleanly.** Over the pairs of tasks of one series:

| Series | Same idea: pairs, least / median / most | Different ideas: pairs, least / median / most | Same-idea pairs less alike than the most alike pair of different ideas |
|---|---|---|---|
| s1 | 42: 0.14 / 0.33 / 0.66 | 148: 0.12 / 0.19 / 0.34 | 23 of 42 |
| s2 | 58: 0.17 / 0.48 / 0.68 | 113: 0.14 / 0.20 / 0.35 | 14 of 58 |
| s3 | 46: 0.33 / 0.44 / 0.59 | 125: 0.27 / 0.40 / 0.68 | 46 of 46 |
| s4 | 32: 0.35 / 0.46 / 0.66 | 158: 0.18 / 0.33 / 0.60 | 30 of 32 |
| s5 | 51: 0.16 / 0.32 / 0.61 | 139: 0.07 / 0.15 / 0.44 | 38 of 51 |
| s6 | 55: 0.20 / 0.42 / 0.68 | 135: 0.08 / 0.23 / 0.46 | 35 of 55 |

In every series a share of the repeats sits below some pair of different tasks: a quarter in s2, more than half in the others. A threshold low enough to catch those refuses new tasks with them. Knights and liars is the extreme case: every repeat of a puzzle in s3 is less alike than some pair of different puzzles. The rules paragraph every question opens with is most of their trigrams, as R56 foresaw, so the measure tells the puzzles apart in neither direction.

### The settings

The brief's setting turned through the child's three interests, space, football and dinosaurs, one per answer. The model kept it in all 118 tasks, once writing "space station" for "space". Within one interest the story changed from task to task. s1's space alone held rover codes, a photo of each pair of astronauts, the order of launches, a team for a satellite, fuel canisters and radio calls. The idea did not change with it. R56 accepted that an idea repeated in a new setting would pass, having measured such a pair at 0.61 to 0.63; this run passed more than that. Of the 86 tasks that repeated an idea of their series, 25 did so in a new setting and 61 in a setting the idea already had, with new names and numbers. A new setting hides a repeat from the check entirely: every repeat of a past task it caught came back in a setting its idea already had.

## What it means

- **Criterion 11.9 is not met as written.** Twenty tasks in a row with no attempt refused as a near-duplicate came in one series of six; in the other five the first such refusal came at the 4th to the 8th task.
- **No child would have been left without a task by the check.** A refusal cost one attempt, and on the strength of the medians above that is a minute or two of waiting.
- **The promise holds for the words and not for the ideas.** A child practising one topic at one level gets a new text every time, and the same few problems inside it, the same numbers among them.
- **This was the hard case.** In an ordinary lesson the rule moves a child from topic to topic, and the level with the answers, so a cell comes back rarely and its few ideas last longer. Practising one topic is what this run measured: a parent can ask for it today, and the card will offer it (T78).

## What to fix

The plan named three candidates.

- **The threshold (SPEC 5.6, R56): not on its own.** One threshold serves every topic, and to let this run's new tasks through it has to stay above the most alike pair of different ideas in a series: 0.68 in knights and liars, 0.46 in the other two topics. Below that it refuses new tasks, as R56 found across the reference tasks at 0.6. Above it, the same tasks with the same numbers measured at 0.26 to 0.43 still pass. Raising the threshold lets through the near-copies it caught here. R56's own answer for the topics that open with a paragraph of rules, a measure that subtracts each topic's recurring trigrams, is aimed at the five different puzzles refused here; it was not measured on these tasks.
- **The settings (the rule, SPEC 3.2): no.** They turn as designed and the model keeps them. More of them would carry a repeat's words further from its first, and the check would see it less often.
- **The reference tasks in the package (SPEC 4.1.1): yes, because that is where the ideas come from.** The rotation is not at fault: over twenty answers at `3-4` it showed the model every task of the cell, and in s1 and s5 the model used all five. What decides is that the guide tells the model to take a reference task's structure. The ideas of a cell are then its three or five reference tasks, a model that sees none of its past tasks reaches for the same one or two of them, and seven of the twenty refusals were against a reference task. Ways to change it, the cheapest first:
  1. **The guide.** Take the level and the kind of reasoning from the reference tasks, and write a problem that is none of them, not one of them in a new setting. This is one instruction, measured by running this protocol again on the same six cells. The run will also show how far the model's own repertoire reaches: the classics it took unasked repeated too.
  2. **A memory of ideas, without the texts.** The model names in `submit_task` the reference task its task builds on, if any. The profile keeps the last few of these for each topic, and the package shows the others first.
  3. **More reference tasks at `5-6`**, where a cell holds three, and in each of the three cells one of them gave six to nine of the twenty tasks: `enum-56-d3-3`'s grid paths 9 in s2, `kl-56-d3-1`'s "if you asked B" 6 in s4, `tri-56-d3-1`'s cancelling steps 8 in s6.

  Which of these, and whether criterion 11.9 should count refused attempts, tasks lost or ideas, is for the author to decide: О-59.

## Other findings

1. **`just play` could not write a task since T62.8.** Its list of the tools the model may call without asking lacked `get_package`, which T62.8 added, and in `-p` mode a tool off that list is refused. Fixed in this task, in the `justfile`.
2. **The model drops a field of the brief it hands back.** 17 attempts were refused because the brief came back without its `target_concept`, which the request fixes and the model may not change; in s4, six tasks of twenty lost an attempt to it. The service could take the topic from the request rather than ask for it back. Not changed here, because it changes what `submit_task` takes.
3. **The sketch decided one refusal.** In s2 the exact similarity was 0.66 and the sketch's estimate crossed 0.7, its error of about 0.03 as R54 describes it. Both readings are of the same idea in the same setting.
4. **Two chats stalled**, one after a refusal and one before its first attempt. In both, the model's stream slowed to a trickle of thinking for ten minutes and more while the subscription's five-hour window stood at 4 % and 16 %, so no limit was near. A child would have seen the card waiting.

## Cost

Estimated before the start: about 242 calls of `claude -p`, $30 to $40 in API terms by Claude Code's count, two to three hours, and $0 of cloud.

Actual: 122 calls, because the answers took no turn of the model; $20.36 in API terms, counted over the 120 calls that finished; 199 minutes of the model and four hours of the clock, from 17:35 to 21:31 UTC; $0 of cloud. The subscription's five-hour window stood at 36 % and the weekly one at 29 % when the run ended, other sessions' use included.

## How to repeat it

```sh
# the service, with the day's ceilings raised
MATHTRAIL_DAILY_TASKS=200 MATHTRAIL_DAILY_FAILED=30 just play-server

# the profile, once
printf '%s' "Hello! I'm the parent. Please set up maths lessons for my child: pseudonym Pip, grade 4, interests: space, football and dinosaurs. Lessons in English. Please create the profile now." \
  | just play -p --output-format stream-json --verbose --session-id "$(cat /proc/sys/kernel/random/uuid)"

# each task: a new chat, then the answer as the card records it
printf '%s' "Hi, it's Pip's parent. Pip would like the next task. For a while we're practising one topic only: Knights and liars (logic.knights_liars), at grade level 3-4, difficulty 3 out of 5. Please keep to exactly that topic, level and difficulty." \
  | just play -p --output-format stream-json --verbose --session-id "$(cat /proc/sys/kernel/random/uuid)" > task.jsonl
just inspect-cli --method tools/call --tool-name submit_answer --tool-arg task_id=<the card's task id> --tool-arg answer=<the right letter>
```

The task's id is in the result of the accepted `submit_task` in `task.jsonl`, and the right letter in what the model handed in. A series stops at the first `task_submitted` line of the server's log whose `failed` holds `near_duplicate`, or at twenty. `go run ./cmd/report < <the server's log>` adds the run up as `just report` adds up a deployment. Running this protocol again after a change to the guide or the package gives the numbers to compare.
