# The adult runs the lesson (T84)

**Measured: the model talks to the adult, keeps to MathTrail's bounds and keeps the answer, and it gave the child's gender away until the rule said "never he or she".** 7–8 October 2026. Six scenarios of `testdata/live/scenarios.json`, re-voiced as the adult, each in a chat of its own against a new child, in Claude Code as `just play` runs it, with the instructions of T84 (R261). Then a re-check of two scenarios once the rule on the child's gender was fixed, and one chat on the instructions before T84 as a control.

- **The voice holds.** In every chat the model spoke to the adult: it said "read it to Fox" or "read it out" before a task, asked what the child chose, and after an answer explained the step, mostly addressing nobody. The adult typed every message, and nothing the model said waited for the child to type.
- **The bounds hold.** Homework called no MathTrail tool; the model said in a sentence that MathTrail gives olympiad tasks for grades 1 to 6, then helped on its own, as the ChatGPT package's negative case has it. Grade 8 saved nothing, and the model named grades 1 to 6. Asked for the answer before the child had answered, the model kept it.
- **"We don't know" is an answer.** Three times out of three it was recorded as `?`, with `confused` set and no trap, before the solution was gone through. A question about the task got help without the answer, and nothing was recorded.
- **The child's gender: found and fixed.** Spoken of in the third person, the child became "she" or "he" after the pseudonym's grammatical gender, in 9 places across the Russian chats, where the model had once addressed the child as "you". The rule now says to speak of the child by the pseudonym, never as he or she, in the present tense. In the re-check the same steps gave 1 slip where they had given 6.
- **No task was written ahead, and T84 did not cause it.** Sonnet 5 in Claude Code never called `prepare_task`, and neither did the control chat on the instructions before T84. Claude Code shows the model a result's payload rather than its words, so "Now call prepare_task" never reaches it there. T89's Opus 5.5 at `xhigh` called it in every chat.

## What this is

The live check T84 planned before the merge: whether a model reading the new instructions and descriptions runs the lesson with the adult, keeps MathTrail to its own tasks and keeps the answer. The author's lesson in claude.ai, after the release, is the second half: the card, the result's words and the host's own model are seen there.

- **Service.** Built from the working tree of 7–8 October with T84 in it, beside other sessions' work on the card's polling (T91), the home page and the page "Research". It ran on this machine as `just play-server` runs it, on port 8083, with the development sign-in, the profile in memory, and the day's ceilings raised to 200 accepted tasks and 30 failed requests. Instructions at version `68541922096b` for the main run, `52c3e20158fa` for the re-check.
- **Chat.** Claude Code 2.1.280 in `-p` mode, model `claude-sonnet-5` at its default effort, as `just play` runs it: no tools of its own, the lesson's tools only, and no cards. Every message of a scenario went to one session in turn, `--resume` after the first.
- **The answer.** The driver read the right letter of the task on the card from what the model handed in for it, and the adult's message carried that letter, or another for a wrong answer; the model recorded it with `submit_answer`, as a lesson with no card does.
- **The scenarios.** The four of earlier runs, re-voiced: every message is the adult's, and each first message says the parent is typing. Two new ones: "boundaries", the three negative cases of `plugin/plugin.json`, and "questions", a question about the task and "we don't know".

## The run

| Scenario | Messages | Tasks accepted | What the watch list found |
|---|---|---|---|
| first-lesson (ru) | 11 | 4, each at the first attempt | Profile with no real name asked for; two wrong answers explained from their traps; the third task lower on the ladder than the first (`1-2` at difficulty 1, after `3-4` at 2); "Не знаем" recorded as `?`; the hint given when asked, and the answer after it recorded with `hint_used`; the progress at 4 of 5 with no rating. Four slips of gender |
| first-grader (ru) | 5 | 2 | The four skills left out passed as catalog ids; no multiplication, division or number above 20; the answer after 200 seconds recorded as `slow`. One slip of gender |
| own-choice (ru) | 9 | 3 | "A harder one" asked for at difficulty 4 as the model's own choice, `tutor_mode` `llm`. The answer sent again was not tried: the driver left `{letter}` unfilled, and the model asked which letter was meant; the scenario now gives that step its letter. "This one doesn't suit, another" brought a question back — what did not suit — rather than a new task, so nothing was skipped. One slip of gender |
| english (en) | 8 | 3 | Every task in English, every one passing readability at the first attempt; the wrong answer explained from its trap; "We don't know this one" recorded as `?`; the progress in English, in the trial series |
| boundaries (en) | 6 | 1 | Homework: no tool, the sentence on what MathTrail is, then help; grade 8: nothing saved, grades 1 to 6 named; the answer kept before the child's, the model asking for the child's pick or "don't know" — but not offering the hint, which the instructions' later section asks |
| questions (ru) | 4 | 1 | The question got a walk through the conditions without the answer, nothing recorded; "Не знаем" recorded as `?` and the solution gone through. Three slips of gender |

In all, 14 tasks were accepted, every one at its first attempt, one of them with a drawing. No task was written ahead (below).

## The child's gender

Before T84 the model addressed the child, and the rule kept that wording free of gender: praise the step, use the present tense. Talking to the adult, the model speaks of the child, and in Russian it took the gender of the pseudonym's noun: "Лиса" made the child "she" ("какой ответ она выберет", "не угадала", "когда будет готова"), "Сова" a past tense in the feminine ("какую букву выбрала Сова"), "Ёж" and "Ёжик" "he" ("прочитайте ему", "он ещё не ответил"). Nine slips in four Russian scenarios.

The rule in "Always", and the words that come with an answer, now say it outright: speak of the child by the pseudonym, never as he or she, in the present tense. A re-check of the first seven messages of "first-lesson" and of the whole of "questions" on the fixed rule gave one slip ("Лиса посчитала") where the same steps had given six, and the model spoke of the child as "Лиса" and "Ёж" in the present tense elsewhere. Once it put a question to the child by name inside an explanation ("Ёж, подскажи"), against "addressing nobody", which stands below "Always" and so is cut by Claude Code.

## No task written ahead

`prepare_task` was called in none of the 43 turns of the main run, so every next task was written when it was asked for. The model sees two prompts for it in Claude Code, both unchanged by T84: `prepare_task`'s own description ("Call it once a task is on the card, and as the last step of every turn of a lesson") and `next_task`'s. The third, "Now call prepare_task" in the words of an accepted task, never reaches it there: Claude Code hands the model a tool's payload in place of its words when the tool declares an output schema, which `submit_task` and `next_task` do. A control chat on the instructions before T84 (`5a3ed55dcbda`), the same two messages of "english", did not call it either. In T89 Opus 5.5 at `xhigh` called it in each of its ten chats. So it is the model and the host, not T84; whether Claude's own model in claude.ai writes ahead is for the author's lesson to show.

## What the lesson in claude.ai is to show

The author's lesson after the release, about ten tasks on the author's own profile, by the cases of `plugin/plugin.json` and the examples of `docs/listing.md`:

- the card's part, which no run here had: an option pressed, "Another task" sent by Enter, a topic chosen on the card, and the waiting card's words to the adult;
- whether the model still gives the child a gender, where it reads the reminder in an answer's words too;
- whether it writes the next task ahead, with the result's words and the whole instructions in front of it;
- whether, asked for the answer early, it offers the hint.

## How to repeat it

The driver is outside the repository, in `~/mathtrail-play/runs/t84/bin/run-scenarios.sh`: a server started anew for each scenario, one chat a scenario, a watchdog that ends a chat older than sixteen minutes.

```sh
# a copy of the tree, then every scenario, or the ones named
rsync -a --exclude .git --exclude web/node_modules ./ ~/mathtrail-play/runs/t84/src/
MODEL=claude-sonnet-5 ~/mathtrail-play/runs/t84/bin/run-scenarios.sh
# another folder, tree or set of scenarios
RUN=<dir> SRC=<tree> SCENARIOS=<file> ~/mathtrail-play/runs/t84/bin/run-scenarios.sh first-lesson questions
```

## Cost

- The main run: $2.89 by Claude Code's count, for 6 chats and 14 tasks. The re-check: $0.88 for 2 chats and 4 tasks. The control chat: $0.27. In all $4.04 of the author's subscription.
- Cloud: $0; the server ran on the development machine.
