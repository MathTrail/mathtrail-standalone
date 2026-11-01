# Tasks solved from the picture (T105.1)

**Parts 1 and 2 of 4: sections 1–3 on 2026-10-27 (T105.1.1) and sections 4–5 on 2026-11-01 (T105.1.2), with the probe's findings and the decisions after each part.** Sections 6–13 and 15–17 come with T105.1.3–T105.1.4 (RUN.md, table 1a). Each part writes only for the genres kept after the part before it: by the author after part 1, and from part 2 on by the study's recommendation. From part 2 on, the parts run one after another without waiting for the author: where a part leaves a choice, the next follows the study's recommendation, marked so, until the author decides.

A task solved from the picture has its facts on the picture alone: the walls of a maze, the lines of a figure, the numbers on tiles. Without the picture it cannot be solved. Today MathTrail's facts live in the words. The picture only shows what the words say, and every check, the solver among them, works from the words (R69, R228, SPEC 4.4). This study asks which such tasks the service can draw and prove by program, what that changes and what it costs.

## What part 1 finds

- **Every genre can be proved by a program that reads the picture.** The probe wrote 51 honest tasks, at least three for each of the 17 genres, and every one was proved: both runs of its solver, handed the picture, agreed on the right option. The probe proved the die only on a straight path; the blind agent's die, with its side faces marked beside its first cell, turned twice and was proved too (section 2).
- **Four new kinds of picture are enough, one of them general.** A figure on squared paper, `lattice`, carries 12 genres. Identical tiles (`tiles`), rows of shapes with their totals (`shapes`) and a staircase (`stairs`) carry three more, and two genres need no new kind at all: a fence is today's `row`, and ice cream with toppings is today's `piles`.
- **Today's checks cannot tell a solver that reads the picture from one that copies it.** They accepted 54 of 57 lazy solvers derived from the honest ones; the 3 they refused were crowns, refused only because their picture shows the answer, as below.
- **One candidate check works: record what the solver reads.** Requiring every member of the picture that carries facts to be read catches all 51 lazy solvers that copy the picture's numbers, read it in part, or return a constant. It refuses 3 of the 54 honest tasks: two leave a member unread on purpose, and the third writes its tile's sides into the code, which deserves the refusal. It costs no extra run. Runs on a changed picture catch fewer lazy solvers, and banning the picture's numbers from the code refuses most honest tasks (section 3).
- **A solver that reads the picture but answers another question passes every check but the service's own solver.** That gap is the same for the word tasks of today.
- **The check that "the picture shows the answer" refuses whole genres.** It refused all three crowns, and a net or a die whenever the answer is a number drawn on it, because the question no longer gives that number (section 14; T105.1.2 takes it up).
- **Searches stay small, but naive ones do not.** The most an honest solver took was 2,172,392 steps, 9 % of the ceiling, in 157 ms. Two naive brute forces hit a limit: four shapes tried from 0 to 20 ran out of the 25 million steps, and ten sticks placed on four sides asked `product` for more than its million elements.
- **A model writing from the draft package alone wrote tasks that pass.** Blind agents, which saw only what a chat's model would, wrote one task for each of the 17 genres, each at its first try. All 17 were proved, and an independent check found every answer right. Today's review accepted 16, refusing the crown only because its picture shows the answer, and the read rule refused none. They cost far more than planned (section 14).

**What the author decided** (end of this file): all 17 genres go on to T105.1.2, with two new topics and the rest in today's topics.

## What part 2 finds

- **The check that the picture shows the answer can keep its rule and let candidates through.** A picture that draws the right value among at least three wrong ones draws candidates the child must choose among, as a crown's numbers or a net's faces do. That one exception, with a rule that `tiles` and `shapes` mark the asked place with `?`, refuses 1 of the 57 honest and edge tasks, where today's check refuses 6. It refuses none of the blind agents' tasks and changes nothing for the reference tasks. It catches all 12 pictures whose `?` was filled with the answer and 57 of 59 with the answer written in, against 12 and 58 today (section 4).
- **A drawn solution is caught, but not by its values.** A route drawn as dashed lines was refused by the read rule, since the solver never reads it. A mirror drawn finished was refused by the solver's own second run, which then found a different answer.
- **Since T115 the card of how an answer went keeps the task's picture** in a box at its head, open in a chat, where the lesson is, so the child sees it above the picture of the solution. Every probe task drew its solution in the task's own kind but one.
- **No program can hold "no facts in the words".** The numbers a question writes are a genre's rules (the 7 of a die), or names of what is drawn (face 3 of a net), as often as facts. The read rule and a new self-check issue type hold it instead (section 5).
- **Repeats have to be judged by the picture.** By the question alone, more than a third of the pairs of new tasks of one genre with different pictures count as repeats, and most repeats pass. The picture's own elements catch every shared picture and refuse 1 of 139 other pairs.
- **Readability and language need nothing new.** No task of the 79 was refused for either as it stands, and the blind agents' 17 passed at their first try.
- **No genre is held back by the checks.**

**Decided by the study's recommendation, until the author decides** (end of this file): the hidden answer's exception and its `?` rule, no program rule for facts in the words, and repeats judged by the picture, the same picture asked another question counted as one.

## 1. The starting point, checked

Every fact the plan rests on, read on `origin/main` at `99d8a65`:

| Fact (RUN.md T105 and T105.1) | Where | Holds |
|---|---|---|
| A picture is one of 13 kinds, from `clock` to `flags` | `internal/domain/picture/picture.go:26-38` | Yes |
| A label is at most 5 characters; only `flags` paints | `internal/domain/picture/limits.go:10`, `picture.go:65-70` (`noWords`) | Yes |
| The model describes the picture and the card draws it; the words carry every fact | `content/instructions/task_writing.md:80`; SPEC 4.4 "The picture adds nothing" | Yes |
| The solver sees only the five option texts | `internal/infra/starlark/contract.go:20` (`argumentsOf`), `:43` (exactly one argument); `starlark.go:184` | Yes |
| The checks hold the picture to its labels and to not showing the answer, and read nothing else of it | `internal/domain/checks/picture.go:46-70`, `showsanswer.go:28`; R269 | Yes |
| D05 of the prototype excludes spatial geometry, visual counting and symmetry | `docs/prototype/decisions.md:19` | Yes |
| SPEC 1.2 and R68 keep the D05 exclusions | `SPEC.md:87-89`; `docs/decisions.md:500` ("The prototype's D05 exclusions are a decision of their own, and they stand") | Yes |
| R269 chose a closed list because a free kind was named six ways in one run | `docs/decisions.md:508`, `:531` | Yes |
| The service learns on every call whether the client declared cards, and only the words of its replies use it; a client that declares nothing is taken to draw them | `internal/transport/mcp/tool.go:231`, `chat.go:43` (`appsOf`), `:61` (`chatOf`), `:83` (`chatFrom`); R305, R307 | Yes |
| A card of an earlier release reads an unknown kind as no picture | `web/src/widget/payload.ts:78`, `:158` (`.catch(undefined)`) | Yes |
| The near-duplicate and readability checks read the question alone | `internal/domain/checks/review.go:214`, `:224`; `duplicate.go:14` (0.7), `readability.go:16` | Yes |
| 450 reference tasks at `1-2` and `3-4`, five to a cell | `content/content_test.go:286` | Yes |
| A new topic changes `next_task`'s description, needs a name in every language and must be taught at `5-6` | `internal/transport/mcp/nexttask.go:121`, `content/content_test.go:168`; `web/src/i18n/dictionaries.test.ts` | Yes |
| A topic never given comes before the others; ideas carry no weights | `internal/domain/tutor/tutor.go:522` (`oldest`); `content/idea.go:42` | Yes |
| After an answer the card shows how it went in the task's place, so the task's picture is gone; every task draws its solution | R286 (`docs/decisions.md:1116`), `web/src/widget/payload.ts:230-237`; R289, `internal/domain/checks/structure.go:218` | Yes |
| A screen reader hears only the kind's name; R274 refused captions beside words that say the same | `web/src/design/picture/diagram.tsx:163`; `docs/decisions.md:1814` | Yes |
| "Always" in the server's instructions holds 2,047 of the 2,048 characters a host keeps | `content/content_test.go:478-484` | Yes |
| The widget weighs 1,937,396 bytes, 503,335 compressed; pictures added 43 KB, 15 KB compressed, in T70.3.1 | a local build of 2026-10-26, measured again by the study's exploration; `docs/plan/phase-11.md:155` | Yes, for 12 kinds and 22 languages then |

Found while planning this study, and true on `99d8a65` as well:

- **The time a hand-in may take has no room for more runs.** A run of the solver has 2 seconds (`internal/config/config.go:63`) and may wait 3 for its slot (`:65`). The 10 seconds RUN.md names are the solver's share of the slowest hand-in, two runs and two waits (SPEC 6.6, `SPEC.md:1167`). With five calls to Drive at their 10 seconds each, that is the platform's whole minute (R122, `docs/decisions.md:1329`). The margin of an access token counts the same two runs (R117, `docs/decisions.md:82`), and an instance has one slot for solvers (`docs/load.md:7`).
- **The solver runs before the lesson's language is known.** `Examine` runs the solver before the profile is read (R61, `docs/decisions.md:479`), and the picture's numbers are read with the lesson's decimal mark, which only `Judge` knows (`internal/domain/checks/picture.go:69`).
- **"The picture shows the answer" refuses a drawn number equal to the answer** unless the question gives that number itself (`internal/domain/checks/showsanswer.go:28`, `:132`). A task solved from the picture gives no numbers in its question.
- **The research programs that run the whole review over the reference tasks** are pinned to v0.1.53 (`research/go.mod:6`) and do not build against `main`, so the probe has a program of its own (section 14).

### Gemini's three ways

1. **The model writes the picture's code, free SVG for instance.** Rejected. A free drawing can be checked neither for its labels nor for the answer, and no program can read its facts to prove the answer, which is the condition this study found sufficient. R269's reasons stand: a closed list is what makes a picture checkable and drawable.
2. **A set of pictures with parameters: the model writes the parameters and the card draws.** This is how MathTrail already works (R269), and it is this study's way. The question was only which kinds are missing, and section 2 answers it: four.
3. **A separate illustrator agent.** If the service called a model for it, it would break the rule that the service makes no calls to a model of its own (RUN.md rule 5, CLAUDE.md). The chat's model already describes the picture in the same turn as it writes the task. Rejected.

## 2. The genres

### Where the genres' limits come from

From the sources of R309, 3 to 5 problems a genre, read only for their sizes and counts; their words are kept in a scratch folder and quoted nowhere:

- **problems.ru**, eleven subjects at difficulty up to 3 and grades up to 6: "Geometry on squared paper" (1259), cutting (200, 369, 373, 377), the plane cut by lines (372), nets (722, 1053), the cube (825), walks of graphs (246) and axes of symmetry (504). They gave 135 problems, 67 with a picture. The 60 problems of E. G. Kozlova's "Сказки и подсказки" the listing gave held 6 more with a picture.
- What they set: 12 and 24 matches laid as squares of two by two and three by three (Kozlova, problems 124 and 125); figures of up to 3 × 6 cut in two, and ten L-pieces of four cells laid as 5 × 8 (the Mathematical Festival, 2024 and 2003); three-cell corners making a square (the Moscow district olympiad, 2012); three figures laid one on another and pierced through, then laid out side by side (the Lomonosov tournament, 2025), which is the folding genre's punch; a cube rolled over its edges back to its place (problem 65435); the corners of a sheet folded in (64695); 20 sticks of four lengths to make a square (Kozlova, 207), which is more than the probe's search for sticks handles (below).
- **MathNet**, 16 problems with a picture and a short answer drawn at random from its 3D, combinatorial geometry, enumeration-with-symmetry and graph-theory sections. Nearly all are for older pupils, as T104.1 found: those that fit `5-6` once simplified are twelve points on squared paper joined into squares, an ant walking the edges of a cube, and a white rectangle cut from a grid of 9 × 9.
- Every problems.ru problem cited came from Kozlova's book, the Mathematical Festival, the Moscow district olympiad or the Lomonosov tournament, by its line of sources; the contest the prototype began from is not among them (PRODUCT 10).
- The fence, the staircase, the ice cream, the goat on a rope, the crown of tiles, the numbers under shapes, the nets and the maze turned up in none of the subjects read, the walks of graphs aside. Their limits are the study's judgement, and are said to be.

### The genres, one by one

"Steps" is the most the probe's honest solvers took on the genre, against the ceiling of 25,000,000; "kind" is the picture's kind in the probe. A topic marked *new* is the study's proposal, the author's to choose.

| Genre | Levels | Kind and members | Limits (probe) | Solver reads, steps | Answer | Topic | Risk to the hidden answer |
|---|---|---|---|---|---|---|---|
| Crown of tiles | 3-4, 5-6 | new `tiles`: sides, layout, joint, a label in each triangle | 2–12 tiles of 3–6 sides | the tiles; follows the joints round, 43,443 | a number on a tile | parity.alternation | **High**: the answer is one of the tile's numbers, always drawn |
| Numbers under shapes | 1-2, 3-4 | new `shapes`: rows of shapes, a total for each | 1–5 rows of 1–6 shapes | every shape tried from 0 to 20: 2,172,392 for three shapes, over the ceiling for four | a number | arithmetic.tricks | A total equal to the answer |
| Staircase | 1-2, 3-4 | new `stairs`: steps, marks on steps, jumps | 3–16 steps, 4 marks, 4 jumps | the steps and the jumps, 220 | a step or a count of moves | counting.gaps | A mark on the meeting step |
| Sticks of two lengths | 3-4 | `lattice`: sticks drawn as lines on squared paper | up to 8 sticks by `product`; 10 hit its cap | the lines' lengths, 1,950,031 | a length or "impossible" | combinatorics.enumeration | Low |
| Ice cream | 1-2, 3-4 | today's `piles`, round marks for flavours, square ones for toppings | up to 40 a pile | the piles, 125 | a count | combinatorics.enumeration | A pile as large as the answer |
| Fence of rails | 1-2, 3-4 | today's `row`: posts, a label under the gaps, `copies` for two sides | 2–12 posts | the items and the gaps, 167 | a length or a count | counting.gaps | Low, once the row is drawn whole |
| Maze | 3-4, 5-6 | `lattice`: walls as lines, S and E marked | 8 × 8 | walls and marks, breadth-first: 21,821 at 8 × 8 | moves, cells, or which exit | combinatorics.enumeration, or *new* | A path drawn; the guide forbids it |
| Path on cells | 5-6 | `lattice`: blocked cells, S and F | 8 × 8 | cells and marks, 1,383 | a count of routes | combinatorics.enumeration | Low |
| Goat on a rope | 5-6 | `lattice`: the shed's cells, the tie point P, the rope as a line, the treats as points | the tie at a corner of a rectangular shed | everything, 3,449; squares of lengths, no square roots | a count of points | geometry.grid | Low |
| Triangles or squares in a figure | 3-4, 5-6 | `lattice`: lines | lines meet only at corners of the lattice | the lines, 11,507 | a count | *new*, or geometry.grid | Low |
| Pieces making a figure | 5-6 | `lattice`: the shape's cells, the piece's cells marked | turned, never turned over | cells and marks, 7,455 | a count or "it cannot be done" | geometry.grid | A piece's label as the answer |
| Matches | 5-6 | `lattice`: each match a line of length 1 | 3 × 3 cells | the lines, 76,736 for the least to take away | a count | *new*, or geometry.grid | Low |
| Rolling die | 5-6 | `lattice`: the path's cells, the top face in the first cell, the side faces as marks in the cells beside it | a straight path in the probe; two turns in the blind agent's | cells and marks, 1,070; the blind agent's 46,207 | a face, or "cannot be told" | *new* | **High**: a face drawn is the answer after some rolls |
| Nets of a cube | 5-6 | `lattice`: six cells with a number in each | any net of six cells | cells and marks, folds the net: 2,231 | a number, a sum, or a count of faces | *new* | **High** when the answer is a number of the net |
| Folding paper | 5-6 | `lattice`: the sheet, the folds dashed with their order at one end, the punch marked O | up to 8 folds | everything, 1,694 | a count of holes | *new*, or geometry.grid | Low |
| Cutting paper | 5-6 | `lattice`: the sheet's cells, the cuts as lines | 8 × 8 | cells and lines, 3,527 | a count of pieces or cells | geometry.grid | Low |
| Axis of symmetry and mirror | 5-6 | `lattice`: the figure's cells, an axis dashed | axes through the middle of the figure's frame | the cells, 1,025 | a count of axes or cells | *new*, or geometry.grid | Low |

Notes on the table:

- **The die.** The probe's own die drew only its top face, so only a straight path could be proved: rolling straight on, the die shows its top again after four rolls and the face opposite after two, and the faces opposite add up to 7. The blind agent found a better way within `lattice`: it marked two side faces in the cells beside the first one, each face turned toward its mark, and its die turned twice on a path of five rolls and was proved. The rule that a mark beside the path names the face turned toward it is the task's own words, which a kind for a die would make the format's.
- **The goat** keeps to whole numbers because the probe's goat is tied at a corner of a rectangular shed: every length along a wall is whole, and a straight stretch is compared by its square. Elsewhere it would need a square root, which the sandbox lacks; the helpers would need one more. The blind agent's goat went another way: its rope was counted in steps from cell to cell, written in the peg's cell, which a breadth-first walk proves with no lengths at all.
- **The triangles** are counted from the corners the lines pass through. Two lines that cross between corners hide a corner of a triangle from such a program, so the reader would refuse such a picture: a rule of the format, as "a cell lies in its grid" is one now. A straight line is drawn as one line, so that a side made of two drawn lines is not missed.
- **Mirror images of clocks and of words** stay out: a clock seen in a mirror is a clock whose time the picture would have to give, which is its answer.

### One general kind or several narrow ones

R269 chose a closed list because a free description named one subject six ways. `lattice` is not free: its members are fixed, each a list of numbers checked against the lattice, and a picture that breaks a rule is refused member by member, as a grid is today. Against narrow kinds — a maze, a net, matches, a die, a fold — it compares so:

| | One `lattice` | Narrow kinds |
|---|---|---|
| Genres carried | 12 of the 17 | one or two each, so about 9 kinds |
| Members | 7: width, height, cells, lines, dashed, points, marks | 2 to 4 each, about 25 in all |
| Code to draw | one function | about nine |
| What the checks can read | corners, cells, lines and labels; not what they mean | what each means: a wall, a face, a fold |
| The service's own solver | needs a field naming the kind of question (section 3) | could know the question from the kind |

The study proposes the general kind. What a narrow kind adds, knowing what a line means, matters only for the service's own solver, which section 3 keeps as an option.

### Limits that keep a picture legible

Checked by arithmetic now; the drawings come in part 4:

- `lattice` of 8 cells across leaves 35.75 px a cell in the 286 px a picture has on the narrowest card (`web/src/design/picture/text.ts:21`): room for a label of up to two characters at 14 px in a cell, and labels at corners for points.
- `stairs` of 16 steps leaves 17.9 px a step across and down, enough for a label above a step; jumps drawn as arcs over the steps.
- `tiles`: twelve squares in a row leave 23.8 px a tile, too little for four triangles' labels at 11 px; in a row, six tiles is the most that keeps 47 px a tile. A ring of ten pentagons, the genre's own picture, round a circle of 200 px gives each pentagon about 60 px, by the circle's length shared out (2π × 100 ÷ 10), room for five labels. The limit of a row is to be lowered to 6 when the drawing is written.
- `shapes`: six shapes and a total in a row need about 40 px each, which fits.

### What no kind may carry

The task's picture goes to the card unsealed, before the answer. So no kind has a member that holds a solution: a path through the maze, the sheet unfolded, the die's faces after rolling. Marks numbering a path belong to the picture of the solution, sealed until the answer (R276), and the probe's tasks drew their paths and counts there.

### What a new topic costs

Read from the code, since the author chooses topics before part 3 measures the rule:

- `next_task`'s description lists every topic (`internal/transport/mcp/nexttask.go:121`), so a new topic changes it, and ChatGPT reviews a changed description again before it serves it.
- A topic has a name in each of the 64 languages of the card, and the site's data and pages.
- A topic must be taught at `5-6` (`content/content_test.go:168`), and its reference tasks, ideas, solver templates and picture examples follow the rules every topic does.
- The rule gives a topic the child has never been given before the others (`internal/domain/tutor/tutor.go:522`), so a new topic reaches every child within the next few tasks.

## 3. How a program proves the answer from the picture

### The solver's second input

The probe's copy of the service hands the solver the task's picture as a second argument. `solve(options)` stays as it is, and `solve(options, picture)` gets the picture; `solveOf` takes either (`contract.go:43` in the copy). The 705 solvers of the reference tasks run as before: the copy's whole `go test ./...` passes with the change, and so does the build of `tools/load`.

The path the picture takes is `Examine` → a new `solver.VerdictWith` → a `PictureRunner`, a `Runner` that can also hand a program the picture → the sandbox's `RunWith` → `solve`. Both runs of a hand-in get the same picture, and only the options are relabelled. In the copy the change is about 340 lines of code: 269 for the picture as Starlark sees it, the rest across the contract, the sandbox and `Examine`.

### How the picture reaches it

The solver runs in `Examine`, before the profile is read, so before the lesson's language is known (R61). Two ways:

1. **The description as it came, with no decimal mark read.** The probe's way. Numbers written as JSON numbers reach the solver as numbers, and labels as texts, so a label `12` is `int("12")` in the program. A decimal label, `2,5` in German, would have to be read by the program itself, and none of the probe's 17 genres writes one.
2. **The language in `Examine`.** The solver could read the picture as the checks do. It means reading the profile, or the request, before the solver runs, which changes R61's halves, or `submit_task` taking a language, which changes R100.

The study proposes the first: it changes no decision, and the genres need no decimals.

### What the solver reads it as

An object becomes a dictionary and a list of numbers or texts a tuple, so a cell `[2, 3]` reads as `(2, 3)` and can stand in a set. A dictionary of the picture reads like Starlark's own: `picture["cells"]`, `picture.get("lines", ())`, `keys`, `items`, `values`, `len`, `in`. It is frozen. A list of objects, such as a lattice's points, is a list of such dictionaries. The values know their length, because the sandbox's capped built-ins refuse one that does not (`internal/infra/starlark/vocabulary.go:145-155`).

No new helper was needed. The probe's solvers use what the sandbox has: `gcd` for the corners a line passes, `product` and `combinations` for searches, `while` and sets for breadth-first walks, since recursion is off. Integer squares stand in for square roots. Two templates would be new, for a bounded search over the shapes' numbers and for placing pieces, because the naive search runs out (section 2).

### What "two witnesses" means now

The model's reasoning and the program were two witnesses to the answer of the words (SPEC 5.7). With the picture handed to the program, the program becomes a witness to the picture: it works the answer out from what is drawn. A picture drawn unlike the answer the model meant is then caught by the check that exists, `solver_disagrees`. The probe's fence drawn with one post too few was refused so, while a solver writing its numbers out would have passed it. What two witnesses still cannot see is a program that reads the picture and answers another question, the gap the next table's last column shows.

### Telling a solver that reads the picture from one that does not

The probe derived lazy solvers from the honest ones by rules written and hashed before any candidate was coded (section 14):

- **copy**: the picture's members written into the program as literals, the argument never read;
- **part**: only the largest member written out, the others read;
- **constant**: `return match(options, <the right option>)`;
- **other**: the picture read fully, but the answer worked out by the mistake of a trap, and the task's right option set to that answer, as a model that believed the trap would write it.

Each candidate was run on its own, on 54 honest tasks (the 51, the member left unread, the sides in the code and the picture drawn unlike its answer) and on the lazy ones:

| Candidate | Honest refused (54) | Blind refused (17) | Copy caught (17) | Part caught (17) | Constant caught (17) | Other caught (6) |
|---|---:|---:|---:|---:|---:|---:|
| The solver read nothing of the picture | 0 | 0 | 17 | 5 | 17 | 0 |
| **A member that carries facts went unread** | **3** | **0** | **17** | **17** | **17** | 1 |
| Any member went unread, layout included | 34 | 7 | 17 | 17 | 17 | 2 |
| One changed picture moves nothing | 11 | 1 | 17 | 14 | 17 | 3 |
| Three changed pictures move nothing | 3 | 0 | 17 | 12 | 17 | 0 |
| Five changed pictures move nothing | 2 | 0 | 17 | 8 | 17 | 0 |
| A number the picture holds stands in the code | 47 | 13 | 17 | 17 | 7 | 4 |
| The same, numbers above 2 only | 13 | 4 | 14 | 14 | 7 | 0 |
| A number an option is stands in the code | 25 | 4 | 14 | 14 | 17 | 1 |
| The service's own solver disagrees (8 honest, 2 blind, 3 of each lazy kind) | 1 | 0 | 0 | 0 | 0 | 3 |

- **Recording what is read** is the one that works. The picture's dictionaries record every member the program reads, on each run, at no cost: the runs happen anyway. A member counts as carrying facts unless it only lays the picture out (`width`, `height`, `layout`, `copies`) or holds nothing but `?`, the mark of what is asked: the blind agent's fence marked its gaps `?` and asked for them, and its solver rightly never read them. Of the 3 honest tasks it refuses, two leave a member unread on purpose: a fence whose question counts trees and never needs the rails' length, and a square whose corners are labelled for the question's words only. The third is a crown solved with its four sides written into the code, which would go wrong on a tile of five, so its refusal is right. The guide can say that a picture draws only what the task uses; the labels a question names, as in the square ABCD, would count as read when the question names them. What it cannot see is a program that reads a member and ignores it. None of the derived solvers does, and runs on a changed picture would catch one.
- **Runs on a changed picture** move one fact at random, a number by one or an item out of a list, and ask whether the program's answer moves. A program that copies the picture never moves, so it is caught. One that reads a part moves when the part it reads changes: with five changes, 9 of the 17 escaped. Two honest tasks never moved, because their answer is "it cannot be told" or "it cannot be done" and stays so. Changes that are bound to move the answer would need the service to know the question, which is the own solver's work. The five extra runs cost a median of 0.54 ms and at most 84 ms on the probe's machine, but each is a run of up to 2 seconds that the minute of R122 and the margin of R117 have no room for: adopting them changes both decisions.
- **Numbers in the code** do not tell the two apart. Honest programs write 0, 1 and 2, directions, `range(4)` and the 7 of a die, and the picture holds the same numbers.
- **The service's own solver**, written for three kinds of question (counting a figure's triangles, the shortest way through a maze, a fence's length), caught all three lazy solvers that answer another question, and refused one honest task: the fence drawn wrong, which it should refuse. It needs the task to say which kind of question it is, a member the format does not have, and a program in Go for every kind of question. It would stand beside the model's solver, not in its place, for the kinds of question where a model's own is weakest; T105.1.4 weighs which those are.

**What the study proposes:** record what the program reads, refuse a solver of a task solved from the picture that leaves a member carrying facts unread, and tell the model which member, never what it holds. Leave runs on a changed picture and number bans out. Keep the service's own solver as an option for the genres where answering another question is likely, priced in part 4.

### The cost against the limits

- Recording reads adds nothing to the time or the steps of a run.
- A picture adds its description to the program's input: the probe's largest, a sheet with three folds, is 307 bytes, and the 8 KB limit is on the source alone.
- `just ci-load` hands in no task with a picture solved from it, so none of this reaches it until a scenario does; the hand-in it would add costs one solver's run, as now.

## 4. The hidden answer

*Part 2, measured on `0727341`, where the probe gives every result of part 1 again (section 14).*

### What today's check refuses, and why

Today a picture shows the answer when it draws the right option's value where the question does not give it (`internal/domain/checks/showsanswer.go:28`). A task solved from the picture gives no values in its question, so the check refuses every task whose answer is a value drawn on the picture. It refused 6 of the 57 honest and edge tasks, all for this alone: the three crowns and the crown with its sides in the code, the die whose top face comes back after four rolls, and the net asked which number is opposite 3. It refused the blind agent's crown, and the net asked again in other words.

In each of them the answer is one of several values drawn, and the skill is to find which: the number a crown's joints carry round, the face a net folds opposite, the face a die shows after its rolls. What R228 forbids is a picture from which the answer is simply read, such as a calendar page that shows the day asked for. Here reading the picture is the task, but reading alone does not give the answer.

### What each kind shows

What the check holds to the answer is what a kind's reading calls shown (`internal/domain/picture/*.go`, `Shown`):

- `lattice`: the labels of its points and of its cells. Its cells, lines and dashed lines are drawn, not written: a count of them is the child's to make.
- `tiles`: the label in each triangle. `shapes`: the totals. `stairs`: the labels on the steps, since the steps are not numbered and a jump is drawn, not written.
- Of today's kinds the probe uses, `row` shows its labels (of labels that count the items only the last), the labels under the items, the gaps and the span; `piles` shows each pile's label, its value and how many counters it has.

### Pictures that show the answer on purpose

From every honest, edge and blind task whose right option fits on a label, a script wrote pictures that show the answer, by rules written and hashed before any draft rule was coded (section 14):

- **The unknown answered**, 12: every `?` of the picture replaced by the right option, the crowns, the shapes, a die and a fence.
- **The answer written in**, 59: the right option added where the kind writes a value of its own, a mark in the first free cell of a lattice, a row's span, a pile's value, a mark on a free step.
- **The solution drawn**, 3, by hand: a maze with its shortest way drawn dashed, the robot's grid with one route drawn dashed, and the figure of the symmetry task with the cells it needed filled in.

### Draft rules

Each draft rule was applied to every task, together with the 104 reference tasks that draw today:

| Rule | Honest and edge refused (57) | Blind refused (17) | Reworded refused (5) | Reference tasks refused (104) | Unknown answered caught (12) | Answer written caught (59) | Solution drawn caught (3) |
|---|---:|---:|---:|---:|---:|---:|---:|
| Today's | 6 | 1 | 1 | 0 | 12 | 58 | 0 |
| The right value drawn, and no wrong one | 1 | 0 | 0 | 0 | 0 | 45 | 0 |
| A picture of `tiles` or `shapes` with no `?` | 0 | 0 | 0 | 0 | 10 | 0 | 0 |
| Either of the two above | 1 | 0 | 0 | 0 | 10 | 45 | 0 |
| Today's, unless two or more wrong values are drawn; or no `?` | 1 | 0 | 0 | 0 | 11 | 53 | 0 |
| **Today's, unless three or more wrong values are drawn; or no `?`** | **1** | **0** | **0** | **0** | **12** | **57** | **0** |
| Today's, unless four or more wrong values are drawn; or no `?` | 5 | 0 | 0 | 0 | 12 | 57 | 0 |

- **The rule the study proposes** keeps today's check and lets it pass a picture that draws the right value among at least three wrong ones, the five options' values standing as candidates the child must choose among. It refuses one honest task, the die whose answer is the face drawn in its first cell, which a child could guess by sight; the model writes another path. Its two misses add a label the honest picture already drew: a second `E` in a maze whose question names E, a second 6 on a net that has one. With two wrong values instead of three, five more pictures that show their answer slip through, those of two dice, two ice creams and a fold; with four, the probe's crowns, which draw three wrong values, are refused again.
- **What the exception lets through.** A picture that writes the answer in beside three wrong values passes it by construction, and the probe wrote no such picture, so nothing here says how often a model would. The guide forbids writing the answer on the picture, as it does today, and the check before release (part 4) counts such pictures among its refusals.
- **The `?` rule** catches 10 of the 12 pictures whose asked value was filled in: the crowns and the shapes, and not the die and the fence, whose kinds it does not cover. Their purposes already say "mark the triangle asked about with ?" and "mark the total asked about with ?". A `lattice`, `row` or `stairs` asks other things, so the rule stays with the two kinds.
- **No rule on values catches a drawn solution**, but other checks do. The maze and the grid drew their routes as dashed lines that their solvers never read, and the read rule of section 3 refused both. The symmetry task with its cells filled in was refused because its solver, which reads the picture, then found nothing to add and disagreed with the answer. A drawn solution is either a member the solver does not need or a picture that answers another question.
- **Word tasks keep today's check**, which is what R228's calendar rule needs. On the reference tasks the two rules agree, so this only says which tasks the relaxed rule applies to: those whose brief holds `facts_on_the_picture`.

### The picture of the solution, and the card after the answer

Every one of the 74 tasks of the probe and the blind sample drew its solution (R289), 73 in the task's own kind and one in a table, and the review accepted every such picture it reached. In a lattice the solution draws a path dashed or numbers it with marks, cells filled for a mirror or a cut, and the folded sheet's holes. A crown fills its `?`, the shapes write their total, and the stairs mark the meeting step. The solution's picture may show the answer, since it is sealed until the child answers (R276).

Since R315 (T115), the card of how an answer went keeps the task's question and picture at its head, in a box that folds them away, open in a chat (`web/src/widget/ResultCard.tsx:334`, `TaskWording.tsx:28`). In a chat, where the lesson is, the box is open and the child sees the task's picture above the solution's, so the loss that section 1 found on `99d8a65` (R286) no longer holds there; the site's own cards show the box folded.

### What a refusal says

New refusals name the member and what to do, never an option or its letter (R62):

- "task.picture marks no place with ?: mark the triangle or the total the question asks about with ?";
- the read rule's, from section 3, naming the member left unread;
- the repeat of a picture, from section 5.

## 5. The words on a short question

### Facts in the words

A task solved from the picture keeps every fact on the picture. A program could hold that only by the numbers the question writes, and those are not facts often enough:

| Question writes | Honest and edge (57) | Blind (17) | Reference tasks that draw (104) |
|---|---:|---:|---:|
| Any number | 13 | 2 | 103 |
| A number its picture shows | 5 | 0 | 54 |

The 13 honest questions write a die's 7 (opposite faces add up to it), the floor's 0, a cell's side of 1, and the labels a question points at: the faces 1, 2 and 3 of a net, the cells marked 1 of a piece. They are rules of the genre or names of what is drawn, not facts of the task. A ban on numbers shown would refuse 5 honest tasks and catch nothing the probe wrote. Applied to the word tasks it would refuse half of them, which is expected, since their facts are in their words.

**What the study proposes:** no program rule. Two checks already hold the facts to the picture: the read rule requires the solver to read every member that carries facts, and an honest solver that read a fact from the words would leave its member unread. The self-check gets a new issue type, `facts_in_words`: a fact the solution uses stands in the question, or the picture does not hold it. The guide's section on tasks solved from the picture names the rule.

### Near-duplicates

A task solved from the picture keeps its words and changes its picture, which turns the near-duplicate check around. Of the 149 pairs of tasks of one genre (honest, edge, blind and reworded, within each genre), 10 share one picture exactly. Five are the reworded tasks with the tasks they reword. The other five are the pairs a shared picture makes beside them: the crown with its sides written into the code and the crown it copies; the fence drawn unlike its answer with that fence and with its rewording; and the net asked two questions with the other question and with its rewording. The other 139 pairs differ in their pictures.

| Fingerprint | Same picture caught (10) | Different pictures refused (139) |
|---|---:|---:|
| The question alone, at 0.7 (today's) | 4 | 51 |
| The question with the picture's elements in one set, at 0.7 | 4 | 40 |
| The picture's elements alone, at 0.7 | 10 | 7 |
| The picture's elements alone, at 0.8 | 10 | 3 |
| **The picture's elements alone, at 0.9** | **10** | **1** |

An element is a member of the description, each item of a list one: a line, a cell, a mark. Today's fingerprint refuses more than a third of the new tasks as repeats and lets most repeats through. Joining the picture to the question's trigrams does not help, because a question has a hundred trigrams and a picture a dozen elements. The picture alone separates them: at 0.9 it catches every shared picture and refuses one other pair, two sheets drawn 0.92 alike. The question is still compared with the reference questions of its level, and none of the 79 tasks copied one.

**What the study proposes:** for a task solved from the picture, a repeat of a task the child was given is judged by its picture: elements 0.9 alike or more. The profile keeps a second sketch for such a task, of its picture, as it keeps one of the question now. Part 4 counts its bytes. Whether one picture asked another question is a new task is the author's question. The study counts it a repeat, since the child solves the same figure.

### Readability and language

None of the 79 tasks was refused for readability, and none for its language. In part 1 three of the probe's questions were reworded after a first run to pass readability, while the blind agents' 17 passed at their first try. Their questions are not short, since they state the genre's rules: 29 words on average at `1-2` (34 at most), 36 at `3-4` and 37 at `5-6` for the executor's set, and 38 to 59 for the blind agents'. Their Flesch-Kincaid grades stayed within the limits, at most 3.7 against 4 at `1-2`, 5.8 against 6 at `3-4` and 6.8 against 8 at `5-6`. The rules need no change. A short question that left the genre's rules to the picture would pass them more easily still.

## 14. The probe

*Parts 1 and 2. T105.1.3–T105.1.4 add to it.*

- **The copy.** A clone of the repository at `99d8a65`, with the draft as one commit on it: the four kinds (`lattice`, `tiles`, `shapes`, `stairs`) read as strictly as the thirteen, with their limits, schema and examples; the solver's second input and the recording of reads; a section of the guide, "A task solved from the picture", that tells the model when a brief's `constraints` hold `facts_on_the_picture`; and 17 ideas, one a genre, added to five topics' lists. The guide grew from 17,204 to 19,553 bytes. `go build ./...`, `go test ./...` and the build of `tools/load` pass in the copy.
- **The program**, `t1051`, a Go module that imports the copy as `tools/load` imports the service. It builds packages (`packages`), runs both runs of a solver with the picture and the whole review (`examine`), and each candidate on its own (`candidates`).
- **The set.** 114 tasks: 51 honest, at least three a genre; 6 edge cases (a member left unread on purpose, sides written into the code, a picture drawn unlike the answer, and three pictures at their limits); and 57 lazy solvers derived from the first honest task of each genre, 17 of each of three kinds and 6 that answer another question. The rules of the derivation and the blind agent's briefs were written first, and their SHA-256 recorded on 2026-10-27 before any candidate was coded: `b35c261a…218190` for the rules, `3aed7c7f…1034f7` for the briefs.
- **What the review said of the honest tasks.** The 51 honest tasks and the two edge cases that only read differently were all proved, and the review accepted 47 of the 53. The 6 it refused were refused by "the picture shows the answer" alone: the three crowns, the crown with its sides in the code, the die after four rolls and the net asked which number is opposite. T105.1.2 reads them, and the word checks, from the same records. After a first run three of the probe's questions were reworded to pass readability, and a picture of a solution got labels its words name; after that no readability or near-duplicate check refused an honest task.
- **The limits.** Of the three pictures at their limits, the maze of 8 × 8 was proved in 21,821 steps; the four shapes ran out of the 25 million steps, and the ten sticks asked `product` for 1,048,576 placements, past its cap of a million.
- **The blind agents.** Agents on the study's own model, `claude-opus-5-5`, were given the packages the copy's builder made for the 17 briefs, with the draft guide, the kinds' examples and the solver's contract inside, and nothing else. Each was told to read only its package, to write only to a folder of its own, to run no program and to use no connector, MathTrail's above all, whose `submit_task` would have reached production, and to write each task once.
  - **The first agent** took all 17 packages. It read them all before writing, so every turn carried a context of about 220 thousand tokens. After 77 turns it had written 4 tasks, the fence, the maze, the shapes and the staircase, for 176,791 tokens written and 10.1 million read again from its cache, and it was stopped. Its brief had not listed the packages' names, and it asked for them.
  - **Thirteen more agents** then wrote the other 13 tasks, one package each, at the author's order. Each got the package with its guide written out as a file of its own, so that a long line of JSON could not be cut on reading.
  - **The records of their calls** show every agent kept to the rules: the 13 read only their two files and wrote only their one, and the first read only the packages' folder and wrote only its own; none ran a program or touched a connector.
  - **What they wrote.** 17 tasks, one a genre, each at its first try. All 17 were proved on both runs. Every answer was right when checked apart from the agents: worked out by hand for each task, and by the service's own solver for the maze and the triangles. Today's review accepted 16; it refused the crown only because its picture shows the answer. Of the candidates, the read rule refused none of the 17, nor did three or five changes of the picture; one change left the staircase's answer where it was. Bans on numbers would have refused 4 to 13 of them.
  - **What they found that the probe had not.** The die's side faces drawn as marks beside its first cell, which let a die turn. A rope counted in steps from cell to cell, with no lengths at all. Pieces named by letters other than A to E, which keeps them apart from the options' letters. And a gap in the draft: the package never named `lattice`'s member for dashed lines. The agent that needed it guessed `dashed`, rightly; the example of a kind has to show every member the kind has.
  - **What they cost.** 1,989,181 tokens written in all: 176,791 by the first agent and 1,812,390 by the thirteen, whose tasks took from 38,349 tokens (the ice cream) to 492,223 (the sticks), 100,895 in the middle. The plan said 150 to 200 thousand for all 17, and, after the first agent, 600 thousand for the other 13. The genres the first agent wrote were the easy ones, and the estimate drawn from them was three times too low.
  - **What 17 tasks show.** That every genre can be written right at the first try by Claude from the draft package, and the kinds of slip a check must meet. They do not show how often a model fails over many tasks, nor how other models do; that is the check before release.
- **Part 2: the copy on a new base.** The copy was rebuilt on `0727341`, which holds T115's card of how an answer went, with part 1's draft laid on it unchanged. `go build ./...`, `go test ./...` and the build of `tools/load` pass, and the probe gives every result of part 1 again, task by task, on all 114 tasks.
- **Part 2: the set.** Its rules were written and their SHA-256 recorded on 2026-11-01, before any draft rule of part 2 was coded: `6eb67052…dac6c13`. A script derived 71 pictures that show the answer from the honest, edge and blind tasks. By hand came three pictures with the solution drawn and five tasks with their question said another way over the same picture.
- **Part 2: the program.** Three more commands: `derive2` writes the pictures that show the answer, `hidden` applies the draft rules of section 4 to every task and to the 104 reference tasks that draw, and `words` measures section 5. The draft rules live in the copy beside the checks they bend (`internal/domain/checks/draft_t1052.go` there), so that they read pictures exactly as the service does.
- **Kept** in `~/.claude/keep/t105.1/`, which outlives the container: the program, the sets and their builders, the outputs, and the copy's draft as a patch with its base commit.

## What the author decides after part 1

For each genre: does it go on to T105.1.2, and into which topic? The study's proposal:

| Genre | Proved | Proposal |
|---|---|---|
| Crown of tiles | Yes | Goes on, in parity.alternation; T105.1.2 must answer the hidden-answer check first |
| Numbers under shapes | Yes, with a bounded search | Goes on, in arithmetic.tricks |
| Staircase | Yes | Goes on, in counting.gaps |
| Sticks of two lengths | Yes, up to 8 sticks | Goes on, in combinatorics.enumeration |
| Ice cream | Yes | Goes on, in combinatorics.enumeration, with today's `piles` |
| Fence | Yes | Goes on, in counting.gaps, with today's `row` |
| Maze | Yes | Goes on, in combinatorics.enumeration, or in a new topic of routes with the path on cells |
| Path on cells | Yes | Goes on, in combinatorics.enumeration |
| Goat on a rope | Yes, tied at a corner | Goes on, in geometry.grid |
| Triangles or squares | Yes, lines meeting at corners | Goes on, in a new topic of counting figures, with matches |
| Pieces | Yes | Goes on, in geometry.grid |
| Matches | Yes | Goes on, with the triangles |
| Rolling die | Yes, with its side faces marked | Goes on; a kind of its own for a die would make the marks' meaning the format's |
| Nets | Yes | Goes on, in a new topic of cubes and paper, with folding |
| Folding paper | Yes | Goes on, with the nets |
| Cutting paper | Yes | Goes on, in geometry.grid |
| Symmetry | Yes, cells about an axis | Goes on, in geometry.grid or with the nets |

New topics in this proposal: two at most, counting figures and cubes and paper, each taught at every level, since a topic must be taught at `5-6` and these genres start at `1-2` in the problem books. The other genres join topics of today, where a task solved from the picture would be a new kind of task beside the word tasks.

**Decided by the author on 2026-10-27.** All 17 genres go on, the die in its narrow form: a straight path and the top face. The topics are as proposed. Two new topics, each taught at every level: counting figures, with the triangles and the matches; and cubes and paper, with the nets, the folding and the die. The other genres join today's topics as a new kind of task:

- the crown joins parity.alternation, and the numbers under shapes join arithmetic.tricks;
- the staircase and the fence join counting.gaps;
- the sticks, the ice cream, the maze and the path on cells join combinatorics.enumeration;
- the goat, the pieces, the cutting and the symmetry join geometry.grid.

The blind agent's other 13 tasks were written by an agent a package (section 14).

## What the author decides after part 2

The parts now run one after another, so what follows was decided by the study's recommendation on 2026-11-01, and stands until the author decides otherwise:

- **The hidden answer.** A task whose brief holds `facts_on_the_picture` keeps today's check, except that a picture drawing the right value among at least three wrong ones does not show the answer; a picture of `tiles` or `shapes` marks the asked place with `?`. Word tasks keep today's check as it is.
- **Facts in the words.** No program rule; the read rule and a new self-check issue type, `facts_in_words`, hold it.
- **Repeats.** A task solved from the picture repeats a task the child was given when its picture's elements are 0.9 alike or more, whatever its words; the profile keeps a sketch of its picture. The same picture asked another question counts as a repeat.
- **Genres.** All 17 go on to T105.1.3: the checks hold none back.
