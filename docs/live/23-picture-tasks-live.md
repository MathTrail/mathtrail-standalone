# Tasks solved from the picture, live (T105.13)

**Measured: every one of the 34 tasks solved from the picture reached the child, each with its right answer, a solver that answers the question asked and no answer drawn beyond what the rules allow. A third of their hand-ins were refused, almost all as near-copies of the one reference task of their genre the package shows, and each refusal cost the child about 14 s more.** 2026-11-08. Nineteen chats in Claude Code, two tasks each, on the release of T105.12 (`42b4408`, instructions `04f8b87cefa6`), against a local server in a chat taken for one that draws the cards.

- **No wrong answer reached the card.** Two readers solved every accepted task from its picture alone, coordinate by coordinate:
  - every `correct_answer` was right;
  - every solver read the facts from `picture` and computed the quantity the question asks;
  - none hard-coded a fact the picture holds.
- **The answer is drawn twice, both times as the rules allow.** The two crowns of tiles draw the right number on a tile, each beside the values of three wrong options. No other picture shows the answer as a value.
- **Refusals: 16 of 50 hand-ins, and 15 of them failed `near_duplicate`.**
  - Each of those repeats a reference task of its level, at 0.71 to 1.00: in today's topics the reference task of the idea itself, which the package shows last, and in the two new topics the nearest one of the level.
  - One question copied its reference task word for word.
  - Every second attempt rewrote the question, to 0.19 to 0.65 alike, and was accepted. No request ran out of attempts.
  - A sentence of the guide that asks for the genre's rules in the model's own words (R327) took them away: a re-check of the five chats with the most of them gave none in 8 tasks.
  - Without the near-copies, 2 of 50 hand-ins were refused: one came with no picture at all, and one drew its answer, which the hidden answer's check caught.
- **A task solved from the picture takes longer to write.**
  - The first hand-in came 31 s after the package on average, median 26 s.
  - The four tasks in words of the same chats came after 19 s, median 17 s.
  - A second attempt took 14 s more.
  - From the request to the task on the card: median 43 s, 90th percentile 74 s, longest 80 s.
- **What it cost.** $9.44 for the 19 chats, $0.50 a chat, against the $13 the plan estimated. Each chat wrote two tasks and handed in about three times.
- **Not measured here:**
  - whether Claude's own surfaces declare MCP Apps with each call (SPEC remark 73): Claude Code declares none, so the run had to declare them for it, and the answer is in the deployed service's lines, which the author reads;
  - ChatGPT, where a lesson of each new topic on a phone is the author's own;
  - children: nobody answered these tasks.

## What this is

T105.13 of [RUN.md](../../RUN.md), the check before release that the study [picture-tasks](../picture-tasks.md) asks for in its section 16. T105.3 to T105.12 shipped the tasks solved from the picture in the 0.8 line: the solver that reads the picture, four new kinds, the checks, the rule that knows the chat, two new topics, the guide's section and 48 reference tasks. No probe can say how often a model's hand-in of such a task is refused, how long it takes to write, how often an accepted picture draws the answer beside three wrong values, or whether a model's solver reads the picture and answers another question (Д31, R319). This run asks a model in a chat.

## The run

- **Service.** The tree of `main` at `42b4408`, built into a binary and served as `just play-server` serves it, on a port of its own: the development sign-in and the profiles in memory, instructions `04f8b87cefa6`. The run's copy made one change to the code, outside the checks: the store in memory began with the profiles below.
- **The chat's kind.** Claude Code 2.1.280 declares `roots` and `elicitation` and no MCP Apps, so the service takes it for a chat of words and gives it no task solved from the picture (R322). A small proxy between Claude Code and the server added MCP Apps to what each request declares, as a host that draws the cards does, and signed each chat in as an account of its own. The run's session text told the model that the adult and the child see each task on its card.
- **Profiles.** One per chat:
  - the trial series over;
  - the lessons kept to one topic;
  - the round of the topic's ideas at the chat's level given but for the ideas the chat is for.

  Before the run, a script with no model called `next_task` and `get_package` for every profile: each opened a request on a task solved from the picture, on one of the ideas left.
- **Chat.** `just play` in `-p` mode, the model `claude-opus-5-5` at `--effort medium`, as in [22-pictures-live](22-pictures-live.md). Each chat sent one message, "Hello! Comet and I are ready for today's lesson. A task, please.". The model asked for a task, wrote it and then wrote the next one ahead, as every lesson does.
- **Genres.** Two ideas left in a round give two tasks solved from the picture in one chat: the one asked for, and the one written ahead. Where a level holds one such idea, the task written ahead was one in words. The 19 chats cover the 17 genres twice each:

| Chats | Topic and level | Genres |
|---|---|---|
| 2 | `parity.alternation` `3-4`, `5-6` | crown of tiles |
| 2 | `arithmetic.tricks` `1-2`, `3-4` | numbers under shapes |
| 2 | `counting.gaps` `1-2`, `3-4` | staircase, fence of rails |
| 2 | `combinatorics.enumeration` `3-4` | sticks of two lengths, ice cream |
| 2 | `combinatorics.enumeration` `5-6` | maze, path on cells |
| 2 | `geometry.grid` `5-6` | goat on a rope, pieces making a figure |
| 2 | `geometry.grid` `5-6` | cutting paper, axis of symmetry |
| 2 | `counting.figures` `3-4`, `5-6` | triangles or squares, matches |
| 3 | `geometry.cubes_paper` `1-2`, `3-4`, `5-6` | nets of a cube, rolling die, folding paper |

- **Watchdog.** A loop ended any chat older than 16 minutes. It ended none: the chats ran 50 to 126 s.
- **Reading the tasks.** Two readers, each given 17 of the 34 accepted tasks, the formats of the kinds and the guide's section, solved each task from its picture and judged its solver against its question.

## Hand-ins

| | Tasks | Hand-ins | Refused | Out of attempts |
|---|---:|---:|---:|---:|
| Solved from the picture | 34 | 50 | 16 | 0 |
| In words | 4 | 5 | 1 | 0 |
| All | 38 | 55 | 17 | 0 |

Why attempts were refused, by the first check each names:

| Check | Solved from the picture | In words |
|---|---:|---:|
| `near_duplicate` | 14, and one more beside `bad_structure` | — |
| `bad_structure`: no picture | 1 | — |
| `drawing_mismatch`: the picture shows the answer | 1 | — |
| `drawing_format`: a label of the solution's picture | — | 1 |

In [22-pictures-live](22-pictures-live.md) 9 of 46 hand-ins were refused, and in [18-task-writing](18-task-writing.md) 7 of 27.

## The near-copies

The near-duplicate check compares a task solved from the picture with the reference tasks of its level by its question alone, at 0.7 (R54, R56, R321). The study measured it on 79 tasks written with no reference task of their genre, and none copied one. Since R318 the package shows one, and the model copied its question:

| Where | First attempts refused as near-copies | Alike, the reference task they repeat |
|---|---:|---|
| Today's topics, 12 genres | 12 of 24 | 0.71–1.00; in all 12 the reference task of the idea itself |
| The two new topics, 5 genres | 3 of 10 | 0.82–0.85; the nearest reference task of the level |

- A question of such a task carries the genre's rules and nothing of its facts, which stand on the picture. Two tasks of one genre say much the same words, and the one reference task the package shows is the model's only example of them.
- No refused picture was a reference task's picture. In 12 of the 14 refusals with a picture, the second attempt kept the picture and only rewrote the rules in other words.
- The refusal is right by its own rule, since the question did copy a reference task, but the child gets nothing for it: the second task is the first in other words, 14 s later.

The author chose a sentence in the guide's section, which asks for the genre's rules in the model's own words (R327), over comparing the picture with the reference tasks' pictures.

## The re-check

The five chats with the most near-copies were held again on the guide with that sentence, the same profiles, model and message, against a server of its own. The instructions changed with the guide.

| | Tasks solved from the picture | First attempts refused as near-copies | Nearest reference task |
|---|---:|---:|---|
| The run | 8 | 5 | 0.85–1.00 among those refused |
| The re-check | 8 | 0 | 0.41–0.67 |

- One hand-in of the re-check was refused, the crown at `5-6`, by `solver_error`, and its second attempt was accepted.
- The five tasks asked for reached the card 28 to 69 s after their request, median 34 s.
- The re-check cost $2.28.

## The time to write

From the package to the hand-in, by the proxy's clock:

| | Hand-ins | Mean, s | Median, s | Longest, s |
|---|---:|---:|---:|---:|
| Solved from the picture, first attempt | 34 | 31.1 | 26.2 | 70.5 |
| asked for | 19 | 34.2 | 30.7 | 58.9 |
| written ahead | 15 | 27.2 | 23.4 | 70.5 |
| Solved from the picture, a second attempt | 16 | 14.3 | 13.8 | 21.5 |
| In words, first attempt | 4 | 19.1 | 16.8 | 27.0 |

- The longest were the crowns of tiles, the goat and the pieces, and the matches at `5-6`: 57 to 71 s.
- The four tasks in words are few, and the API's speed changes by the day, as [22-pictures-live](22-pictures-live.md) found. Tokens compare better: the chats wrote 155,869 output tokens, about 2,800 a hand-in, against about 2,000 a hand-in on 2026-10-10.

## What the readers found

- **Wrong answers: none.** All 34 `correct_answer`s agree with the readers' own solving. The readers checked three counts of triangles and the least set of matches with a program of their own.
- **Solvers that answer another question (Д31): none.**
  - Every solver reads the facts from `picture`. The one fixed number is a cell the question names.
  - One solver takes a shortcut: the crown at `5-6` assumes how each tile is turned, rather than searching every turn. The shortcut is right here because the six numbers come out distinct, and the solver does not check that. On a crown with a number repeated, it could prove a wrong answer.
- **The answer on the picture.** The two crowns draw the right number on a tile, beside three wrong options' values, as R320 allows. One sheet writes the order of its folds as 1 and 2, the way the lattice's format writes it, and its answer is 2. Naming folds by letters would keep such a number off the picture.
- **Small things a parent may notice:**
  - three questions name a count the picture shows: "five squares", "six squares" and "two lines";
  - a route on cells is called a maze though no wall is drawn;
  - a solution names "tile 2 in the picture", which only the solution's picture draws;
  - two cutting tasks say "every line drawn", which a child may read as the paper's own lines.

## What it means

- **The tasks are safe to give.** A solver that reads the picture, the read rule and the hidden answer's check kept every one of the 34 right, in all 17 genres. The check of the hidden answer caught the one picture that drew its answer.
- **The near-copies are the one cost the run found,** and it falls on the wait rather than on the task. A sentence of the guide took them away in the re-check.
- **Д31 stays a debt.** The run found no solver that answers another question, and one whose proof holds only for its own picture.

## Cost and time

Claude Code counted $9.44 for the 19 chats, 7 turns each on average, about $1 more than its first chat promised and $3.6 less than the plan's estimate. The chats ran for about 26 minutes one after another. What Claude Code declares was read from its connection to a stand-in server, before any message was sent.
