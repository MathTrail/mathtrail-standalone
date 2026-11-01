# Tasks solved from the picture (T105.1)

**All four parts: sections 1–3 on 2026-10-27 (T105.1.1), and on 2026-11-01 sections 4–5 (T105.1.2), 6–8 and 12 (T105.1.3), and 9–11, 13 and 15–17 (T105.1.4), with the probe's findings and the decisions after each part.** Each part writes only for the genres kept after the part before it: by the author after part 1, and from part 2 on by the study's recommendation. From part 2 on, the parts run one after another without waiting for the author: where a part leaves a choice, the next follows the study's recommendation, marked so, until the author decides.

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

## What part 3 finds

- **Nothing that chooses, writes, checks or hands out a task knows whether the chat shows the card.** Only the words of a reply do (section 6). Today's rule would give the new topics' tasks to chats without cards as to any, 16.8 % of their tasks in the simulation. Told the chat's kind, it gives none there but a task written ahead in a chat with cards, up to 0.9 % of that chat's tasks, which the task written ahead and the request still open have to record to let go.
- **A client that declares nothing would have to count as a chat of words for such tasks**, the reverse of today's default for word tasks.
- **The card has to ship alone, before the rest**, at least an hour apart, since ChatGPT may serve the page it keeps for up to an hour, and a card that does not know a kind shows the task with no picture. Everything else ships in one later release, which changes the instructions' version once.
- **The new topics set the frequency.** With them, about one task in five is solved from the picture in a chat with cards, against 2 % from the ideas left in today's topics. Every child meets both new topics within four days, and they take 8 to 9 % of all tasks each.
- **Reference tasks: 48 rather than 130.** One for each genre and level outside the count, and one at each difficulty and level for the new topics. The tests now ask 65 for each new topic. A package of a new topic, with three, stays under 50 KB of its 64.
- **About 900 more tokens a package**, only in a package for a task solved from the picture, if the guide's section rides there alone. A task's picture costs about twice today's, some 100 tokens, by a ratio of bytes to tokens read from T70's own counts.
- **The share of hosts that show the card is not measured.** The file of the author's command was not there when this part was written.

**Decided by the study's recommendation, until the author decides** (end of this file): such tasks only where the chat is known to show the card; the card released alone and first; the new topics with the ideas of part 1, no counter; one reference task per genre and level outside the count, and 15 for each new topic; the guide's section only in their packages, shipped in one release.

## What part 4 finds

- **The four kinds draw on a phone.** Each fits the 286 px of the narrowest card, left to right in every language, its labels 12 px or more (section 9). A row of tiles has to hold six at most, and a tile's label two characters.
- **They add 17,470 bytes to a card of 1,947,230**, and 3,294 to its 505,461 compressed: drawings, readers and names in 64 languages (section 10).
- **Five genres of 17 can be solved by ear.** A switch for the parent, through a skill of the catalog left out, serves a child who uses a screen reader at almost no cost. Descriptions read aloud would take about 68 KB in the card for five genres (section 11).
- **The mock-up** shows every genre on the real card, in both themes and in Arabic (section 13).
- **Sixteen files of the repository say what these tasks would make untrue**, five of them in the site's 10 languages (section 15).
- **The recommendation is to make all 17 genres**, in twelve tasks, T105.2 to T105.13 (section 16). The check in a live chat before release costs about $13 of the subscription.
- **Four questions of the product** are left to the author, О-76 to О-79 (section 17).

**Decided by the study's recommendation, until the author decides** (end of this file): all 17 genres, the two limits lowered, the switch for the screen reader, and the tasks T105.2 to T105.13.

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

## 6. Where the card is seen

*Part 3, read on `0727341`.*

### Which chats count as showing it

`chatOf` takes a call for one from a chat that draws cards when it comes from ChatGPT, or from a client that declares MCP Apps, or from one that declares nothing the server can read, as a client on an earlier version of the protocol does (`internal/transport/mcp/chat.go:43-46`, `:61-66`). Only a client that declares its capabilities without MCP Apps is taken for a chat of words. The kind of chat changes only the words of a reply (`chat.go:72-75`; `internal/transport/mcp/payload.go:303-308`). The payload is the same in both (`tool.go:235-238`), and no function of the rule, the checks or the package takes the kind.

For a word task, R305 chose the cheap mistake: a chat taken wrongly for one with cards costs the adult a message. For a task solved from the picture the same mistake gives a child a task that cannot be solved. **The study proposes** that such a task goes only where the chat is known to draw: ChatGPT, or a client that declared MCP Apps with the call. A client that declared nothing gets word tasks.

### How a task solved from the picture could reach a chat without the card

| Path | What the code does today | What keeping it out would take |
|---|---|---|
| A task written ahead | It is kept in the profile (`ReadyTask`, `internal/domain/profile/profile.go:65`). Before it is handed out, `letGo` holds it to the lesson's language and topic, the skills left out, the place asked for and the instructions' version (`internal/transport/mcp/ahead.go:63-82`, `profile/ready.go:252-286`), not to the chat | The task records that it is solved from the picture, and `letGo` drops it in a chat of words, a misfit of its own |
| A request still open | It is kept in the profile (`OpenRequest`, `profile/task.go:80-110`) and fulfilled by any call that brings its id, within 15 minutes and three attempts (`internal/transport/mcp/submittask.go:280`, `profile/request.go:77-80`). No chat is recorded | The request records it too, and a chat of words that finds such a request open asks for another task |
| A family on two hosts | One profile file per Google account (`internal/store/drive/drivestore.go:50-51`); nothing in it is kept per host | Covered by the two rows above, since both live in that file |
| A card that did not draw | A card that waits for its task asks the server, on its own, whether the task is ready (`read_task`, `web/src/widget/ComingCard.tsx:141-174`), so a request whose card asked was drawn. A task written ahead goes straight onto the card, which calls the server only when the child acts (`web/src/widget/hostedService.ts:43-108`). Nothing handles the card's start | For a request, the card's own asking: a request no card asked about may have no card, and its task waits for one. For the rest, the adult's word. The instructions already tell the model to read the task out when the adult says no card shows it (`content/instructions/mcp_instructions.md:38`); a task solved from the picture cannot be read out, so the model asks for another |
| A card of an earlier release | The card's address is fixed (`ui://mathtrail/app.html`, `widget.go:11`), ChatGPT may serve a page it keeps for up to an hour (GPT-49, `docs/catalogs.md:308`), and a card that does not know a kind shows the task with no picture (`web/src/widget/payload.ts:78`) | The order of releases below |
| The topic chosen, the trial series, a probe | Each is a brief of the rule, whose inputs are the profile, the catalog and the model's choice (`internal/domain/tutor/tutor.go:237-292`). A brief's constraints are always empty (`tutor.go:278`). The trial is the first five answers (`profile/progress.go:39`). While a topic is on pause, a probe comes on every sixth task the child has been given, counted over every topic (`tutor/pause.go:15`, `:98-101`, `tutor.go:698-706`). The rule has no checks of a topic of its own | The rule takes the chat's kind as an input. In a chat of words no brief asks for a picture, and the draw of an idea skips the ideas solved from the picture, as `NextIdeaSkipping` does in the copy |

The simulation of section 7 measures two of the paths. With the rule told the chat's kind and half the days in a chat of cards, a task written ahead at the end of such a day and solved from the picture was handed out the next morning in a chat of words: 249, 933 or 1,631 times in 360,000 tasks by the lever, 0.14 % to 0.91 % of the tasks of chats without cards. With the rule as it is today, which does not know the chat, the new topics went to chats without cards as often as to any: 16.8 % of their tasks.

### The order of releases

1. The card that draws the new kinds ships first, alone. The service does not know the kinds yet: a kind the service reads must have its example, and the examples are part of the instructions' version (section 12), which a release of the card alone leaves as it is.
2. The kinds' format and checks, their examples, the guide's section, the rule and the ideas ship together in a later release, at least an hour after the first, so that no page ChatGPT keeps is older than the kinds. The card's address stays as it is; whether to version it is О-66.

### The words the model reads

- **"Always"** in the server's instructions holds 2,047 of the 2,048 characters a host that cuts keeps (`content/content_test.go:478-492`). There is no room in it, and nothing about tasks solved from the picture needs to be there: the rules ride in the guide's section of the package and in the words of a reply.
- **A reply in a chat of words** asks the model to read the task out, with its options (`internal/transport/mcp/payload.go:337-341`), and the picture is left out of those words (`submittask.go:574-587`). With the paths above closed, no task solved from the picture reaches that reply.
- **The card's description** already names "a picture where it has one" (`content/instructions/widget_description.md:1`), and **the tools' descriptions** name the card and never a picture (`nexttask.go:105-121`, `submittask.go:223-230`). Neither changes for these tasks. The new topics change `next_task`'s list of topics (section 2), which ChatGPT reviews again.

### How many hosts show the card

The author's command counts 30 days of calls by the host and by what it declared. Its file was not there when this part was written, so the share is not measured here. The command is in RUN.md, T105.1, and part 4 reads the file if it comes.

## 7. The rule and how often

### What sets the frequency

The idea a task is built on is drawn from its topic's list, among the ideas not yet given in the round. The draw is a hash, so every idea of a round comes once and none is weighted (`internal/domain/tutor/idea.go:30-48`). It is made where a request is opened (`internal/transport/mcp/getpackage.go:163`, `:170`). So the share of ideas solved from the picture in a list is the share of such tasks in its topic. Three levers were measured against a baseline, L0, with no such idea, by rules written and hashed before the simulation was coded (section 14):

- **L1**, today's topics with the ideas of part 1 whose genres the author left there: 12 ideas in five topics;
- **L2**, L1 with a counter: in a chat with cards, every third task of a topic that has such ideas is built on one;
- **L3**, L1 with the two new topics the author chose, taught at every level, every idea of theirs solved from the picture.

The simulation runs the product's own rule on the copy (`tutor.Next`, `tutor.Ahead`, the task written ahead, the answer recorded), as T104.4's did. 200 children a grade, grades 1 to 6, ten tasks a day for 30 days, make 360,000 tasks a run. A child answers right two times in three, and spends a whole day in one kind of chat. The rule is told the chat's kind, as section 6 proposes: in a chat of words it does not see the new topics, and the draw of an idea skips those solved from the picture. A last run keeps the rule as it is today.

| Lever | All days with cards | Half the days with cards: of all tasks | of the tasks with cards | Handed out without cards | Children given one by day 1, 7, 30 (all days with cards) |
|---|---:|---:|---:|---:|---|
| L0 | 0 | 0 | 0 | 0 | none |
| L1 | 2.1 % | 1.5 % | 2.9 % | 249 | 14 %, 82 %, 100 % |
| L2 | 10.1 % | 5.2 % | 9.9 % | 933 | 14 %, 100 %, 100 % |
| L3 | 18.7 % | 12.0 % | 23.1 % | 1,631 | 29 %, 100 %, 100 % |
| L3, the rule as today | 18.0 % | 17.4 % | 18.5 % | 29,278 | 27 %, 100 %, 100 % |

- **Told the chat's kind, the rule gives none without cards**, by any lever, but for the task written ahead the day before (section 6). Not told, it gives the new topics' tasks to a chat without cards as to any other: with no days with cards at all, 16.8 % of the tasks were solved from the picture.
- **Skipped ideas and topics come back.** With half the days in a chat of words, what is skipped there comes on the days with cards: L1's share there rises from 2.1 % to 2.9 %, and L3's from 18.7 % to 23.1 %.
- **The new topics come early and take a large share.** Every child met both within four days, two in the middle, because the rule gives a topic never given before the others (`tutor.go:522`). They took 8.3 % and 8.6 % of all tasks, more than the 5.9 % an average topic of the 17 takes. They took it mostly from the pigeonhole principle, the calendar, gaps, parity and clocks, about two points each.
- **Every move of the rule can bring one**: its own choice, the trial series, a probe, the topic the adult chose, the model's own choice and "one more like it". Each is a brief, and each draws an idea the same way.
- **A first run counted wrong**, and is kept beside the second for the record. It left all 17 ideas of part 1 in today's topics, five of them twice over once the new topics came, and counted a new topic's task in a chat without cards as a word task. The review of this part found both.

**The rating.** The rating reads a task by its level and difficulty alone (`rating.Point`), so a task solved from the picture has to be as hard as a word task of the same level and difficulty. Whether the model writes them so is for the live numbers (T72.14) to show. Probes and the trial series, where the rating learns most from an answer, are where a miss would cost most.

**What moves elsewhere.** The learners' bench reads the catalog's topics and the reference tasks' traps and calls the rule (`tools/learners/main.go:209`, `:235`; `loop.go:167`), but draws no idea. The ideas alone move none of its numbers. The new topics and new reference tasks move them, and its guard is recorded again with the change that brings them.

**What the study proposes:** the new topics the author chose (L3), with the ideas of part 1 in today's topics, about one task in five in a chat with cards. The counter of L2 needs code in the rule and is kept back, to be added if the live numbers show too few such tasks in today's topics.

## 8. Reference tasks

### What the tests ask of a topic

- At `1-2` and `3-4`, a topic taught there has exactly five reference tasks at each difficulty, 25 a level (`content/content_test.go:286`). At `5-6` it has exactly three at each, 15 (`:321`). A topic taught at every level needs 65.
- Every package of a topic at a level shows a reference task that draws its solution, and one that draws where the level has one (`content/package_test.go:368`).
- A topic's reference tasks give its traps, the most frequent first (`ExampleTraps`). The rule hands them to a new child, and two tests fix them for `5-6` and for the calendar of `1-2` (`content/rule_test.go:104`, `:147`). The learners' bench reads the same traps.
- R309 gave `1-2` and `3-4` no new reference tasks, and R310 replaced eight of the calendar's, which kept 450. Debt Д21 waits to make the comparison with those levels' reference tasks more tolerant, should its refusals grow.

### Four ways

| Way | What it adds | What it touches |
|---|---|---|
| **A set of its own, outside the count** | One reference task solved from the picture for each genre at each level it is taught, shown only in a package whose brief holds `facts_on_the_picture` | The two count tests leave it out. R309 gets an exception. `ExampleTraps` leaves it out too, or the fixed traps and the bench move |
| Replacing tasks of the prototype, as R310 did | The same count, with 18 word tasks of today's topics giving way | The traps of the tasks replaced have to be kept for the fixed traps to hold. The check against the prototype's numbers reads its tasks by number and is unaffected. The site's pages of the topics read the reference tasks |
| New topics only | The two new topics at every level: 65 each under today's tests, 130 | Five tasks a cell from two or three genres would repeat one another, and the genres of today's topics would have none |
| The grid lowered to `1-2` and `3-4` | 50 reference tasks, word tasks as well | Every package, the bench and the site's pages of the grid |

### How many

The blind agents wrote every genre right at the first try with no reference task of its genre in their packages, only the kind's example and the guide's section (section 14). A reference task of the genre is therefore not what makes a model right. What it shows is the form: a question that says only what to find, a solver that reads the picture, a picture of the solution.

**The study proposes** the first way: one reference task per genre and level outside the count. That is 18 for the genres in today's topics: 4 at `1-2`, 7 at `3-4` and 7 at `5-6`. For the two new topics, the count tests would hold a topic all of whose tasks are solved from the picture to one reference task at each difficulty of each level, 15 a topic and 30 for both, rather than 65 each. That makes 48 in all. The genres of the new topics are written at `5-6` in section 2. At `1-2` and `3-4` they need simpler forms, such as a net with the faces numbered and one asked for, or the squares of four matches, which the tasks that write them settle.

### The package

In the copy, the budget test's 875 packages are 31,224 bytes on average and 40,291 at most (the case of four-byte letters), against 65,536 (`content/package_test.go:553`). A package shows a reference task without its solver (`content/package.go:177-191`), and a probe task so shown is 1,335 bytes on average and 2,951 at most, over 68 tasks. A package of a new topic shows three reference tasks (`content/content_test.go:283`), all solved from the picture; even counted on top of the largest package, three of the largest add 8,853 bytes and keep it under 50 KB. The 17 packages the blind agents read were 31 to 35 KB written out with indentation.

## 9. How the kinds are drawn, and how they look on a phone

*Part 4. The drawings are drafts in the copy's `web/` (`web/src/design/picture/lattice.tsx`, `tiles.tsx`, `totals.tsx`, `stairs.tsx` there), read by the card's own reader and drawn by its own diagram. They are seen in the mock-up (section 13).*

| Kind | How it is drawn | At its example | At its limits | Smallest label |
|---|---|---|---|---|
| `lattice` | The paper's grid faint, filled cells in the shading tone, lines strong, dashed lines dashed, a dot at each point with its label above it, a mark in its cell | 128 × 128 | 8 × 8 cells: 282 × 282, a cell of 31.75 px with 14 px round the paper for the labels at corners | 13 px |
| `tiles` | Each tile a regular polygon cut into triangles from its middle, a label in each, its first triangle facing the tile before it | 282 × 94 | A ring of 10 pentagons: 228 × 240. A row of 6 squares: 282 × 47. A row of 12 squares: 282 × 23.5 | 12 px with labels of 2 characters |
| `shapes` | Rows of outlined shapes, the same shape alike wherever it stands, and "= total" after each row | 106 × 106 | 5 rows of 6: 278 × 182 | 14 px |
| `stairs` | A staircase climbing right from the floor, steps unnumbered, a label above a step, each jump an arc with an arrowhead | 282 × 178 | 16 steps: 282 × 299 | 14 px |

- **Every kind fits the 286 px** a picture has on the narrowest card, and is laid out left to right in every language. The copy's tests draw each at its example and at its limits and hold it to that room (`drafts.test.tsx` there).
- **Two limits have to come down.** A row of twelve tiles is 23.5 px tall in the drawing's 282 px of room, too low for four labels a tile, so a row holds six tiles at most, as section 2 found by arithmetic. A tile's label stays at or above the card's floor of 11 px only up to two characters, drawn at 12 px in a ring of ten, so a label of `tiles` is at most two characters.
- **The card's height.** The tallest picture is a staircase of 16 steps, 299 px, about as tall as today's grid of eight rows. A task's card with a picture is 750 to 900 px tall today (О-74), and the new kinds do not make it taller.
- **What the drafts do not meet yet.** The card's own tests refuse them in 6 of their 4,691 cases, and the 4,685 others pass. The drafts set fill and opacity as attributes where a drawing speaks only through the card's classes, and draw an empty path where a lattice has no dashed line. The reader's limits are not yet listed under the service's names, and the new names have no notes for whoever translates them. T105.4 meets all four.
- **Scenes for `just web-layout`**, one a kind at its limits: the lattice of 8 × 8 with labels of two characters, the ring of ten pentagons, the row of six squares, five rows of six shapes and the staircase of 16 steps. In the task and in how its answer went, they are 10 scenes in every language the check measures.

## 10. Size

*Built in the copy with its own Vite configuration (`vite.config.widget.ts`), the page `internal/widget/widget.html` measured as it is and with `gzip -9`.*

| Build | Bytes | Compressed | Added |
|---|---:|---:|---:|
| Today's card, `0727341` | 1,947,230 | 505,461 | |
| With the readers of the four kinds | 1,950,353 | 506,313 | 3,123 (852) |
| And their names in every dictionary, the draft's English | 1,960,017 | 507,303 | 9,664 (990) |
| And their drawings | 1,964,700 | 508,755 | 4,683 (1,452) |
| **All of it** | | | **17,470 (3,294)** |

- **Today's card** has grown since section 1 measured it, from 1,937,396 bytes to 1,947,230, with T115's card of how an answer went.
- **A kind's three parts.** Minified alone, the drawings are 1,453 bytes (`lattice`), 1,645 (`tiles`), 1,269 (`shapes`) and 1,082 (`stairs`), against 1,699 for today's grid. The readers are about 800 bytes a kind. The names are 71 bytes in English, and in the 64 languages, at the ratio today's 13 names have to their English (1.77), about 8,000 bytes of words, with the dictionaries' keys on top of that.
- **Descriptions in a task** are 160 bytes on average for a task's picture and 227 for the picture of its solution, at most 557 and 777, over the probe's 68 tasks. They ride in the card's data and in the profile's current and kept tasks as today's do.
- **The profile** gains a sketch of the picture for each task solved from the picture, 128 characters as the question's is (section 5). The profile keeps up to 200 sketches of questions today; kept for the last 50 such tasks, sketches of pictures add about 7 KB to a file whose target is 64 KB.

## 11. The screen reader

Today a screen reader hears the kind's name and nothing else (`picture.<kind>`, `web/src/design/picture/diagram.tsx:163`), and the task's words carry its facts. A task solved from the picture has no facts in its words, so a child who cannot see the picture cannot solve it by ear.

| Genres | By ear, from the description | Why |
|---|---|---|
| Numbers under shapes, the staircase, the fence, the ice cream, the sticks | **Yes**, read as a list: "Row 1: triangle, triangle; sum 6", "A on the floor, B on step 10; a jump from the floor to step 2", the lengths of the sticks | The picture is a list of things and numbers |
| The crown | Hardly | Which triangle touches which is the whole task |
| The maze, the path on cells, the goat, the triangles, the pieces, the matches, the die, the nets, the folding, the cutting, the symmetry | **No** | The facts are a figure, and a figure read as a list of lines and cells is a puzzle of its own |

- **Why R269 and R274 do not settle it.** R269 refused a text form of the picture for chats without cards and for old cards, because the words carried every fact anyway; R274 refused captions for a screen reader beside words that say the same. Neither holds here: the words do not say it.
- **Descriptions read aloud** would serve five genres of 17. They need a sentence to each kind's reading in 64 languages, about 600 bytes of English, about 68 KB more in the card at the ratio of today's names.
- **The other way is a switch for the parent.** The profile already lets a parent leave out skills of the catalog (`Student.ExcludedSkills`). A skill "reading a picture", which every task solved from the picture needs, would let a parent of a child who uses a screen reader leave all of them out, as a chat without cards does. It needs no field of its own and no words in 64 languages but the skill's name.

**The study proposes** the switch, and the kind's name as the picture's label, as today. It is the author's choice (section 17).

## 12. The model and its tokens

### What changes for the model

- **The guide.** Part 1's section "A task solved from the picture" is 2,214 bytes. With 135 bytes of edits elsewhere in the guide, which every package carries, it took the guide from 17,204 to 19,553 bytes. Part 2 adds about 600 bytes: `?` in `tiles` and `shapes`, no answer written on the picture, a picture that draws only what the task uses, `facts_in_words`, and the repeat judged by the picture. If the section rides only in a package whose brief holds `facts_on_the_picture`, a word task pays only for the 135 bytes outside it, some 45 tokens.
- **A kind's example** rides in a package of its topic. The four new ones are 718 (`lattice`), 568 (`tiles`), 391 (`shapes`) and 434 bytes (`stairs`), against 478 on average for the 17 kinds.
- **The server's instructions, the card's description and the tools' descriptions** do not change (section 6).

### Tokens by bytes

T70 counted tokens exactly, and named no ratio of bytes to tokens. Two can be read from its numbers and the files of its commits:

- **Prose: 3.1 bytes a token.** The guide's section on the picture of the solution is 1,442 bytes at `70ea9ed`, and run 22 counted it at 468 tokens.
- **A description in compact JSON: 1.6 bytes a token.** The 95 reference pictures at `efeed8e` average 74.1 bytes, and T70 counted a reference task's picture at 47 tokens.

| Per task | Bytes | Tokens, by the ratio | Today |
|---|---:|---:|---:|
| The guide's section, parts 1 and 2 | about 2,800 | about 900 | 0 |
| A task's picture, as the model writes it (the probe's 68) | 160 | about 100 | 53, run 22 |
| The picture of the solution (the probe's 68, with no total) | 227 | about 140 | 90 with its total, run 22 |

A solver that reads the picture is 1,055 bytes on average, longer than a word task's by the reading. Code has no ratio of its own here; at the prose ratio it is some 340 tokens. All of these are estimates within the two ratios. The count with the tokenizer, about $0.2, is the author's to allow.

### The instructions' version

The version is a hash of every instruction file, the solver templates, the kinds' examples and the lists of ideas (`content/instructions.go:86-95`, `template.go:171-182`, `picture.go:89`, `idea.go:79`). The reference tasks, the catalogs and the schemas are not in it. A task written ahead to an older version is dropped when it would be handed out (`profile/ready.go:256-258`, `ahead.go:65-71`), and the model has written it for nothing.

**The study proposes** that the guide's section, the new kinds' examples, the templates and the ideas of the new topics ship in one release, which changes the version once. The release of the card alone (section 6) changes none of them, and the reference tasks can come in any release, since they are outside the version.

## 13. The mock-up

- **The page.** https://claude.ai/artifact/BWopr8H1CSUj6Pt9rhR6Zv, private, as T70's was.
- **What it shows.** The probe task of each of the 17 genres, and the crown, the maze and the shapes at their limits, on the card and in how its answer went: at 320 px in the light and the dark theme, at 640 px for the pictures at their limits, and the triangles and the maze on a card in Arabic, whose picture still runs left to right. Under each are what a screen reader hears today, the kind's name, and, for the five genres solvable by ear, what the card could read from the description.
- **How it is drawn.** As the site draws its cards with no browser: the copy's `StaticTask` and `StaticResult`, rendered with `preact-render-to-string` through Vite's server-side loading, each task read first by the card's own `readHandedTask`. 22 cards, each with its task and how its answer went. The card of how an answer went folds the task at its head, as the site's do; in a chat it is open (section 4).
- **What it does not show.** The draft's names of the kinds are English in every language, Arabic too, and the page shows no picture drawn on a real phone.

## 14. The probe

*Parts 1 to 4.*

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
- **Part 3: the rule.** Its rules were written and their SHA-256 recorded on 2026-11-01, before the simulation was coded: `e104ae6e…c4591`, and with an addendum after the review of this part found the first run's two faults (section 7), before the second: `7699cf56…73c773a899`. The program's command `rule` runs the copy's rule with a catalog that can add the two new topics, a view of it without them for a chat of words, and a draft draw of an idea that can skip some of a list's places (`tutor.NextIdeaSkipping` in the copy). The second run's 15 runs, five levers and three shares of days with cards, are 360,000 tasks each and took about ten minutes. A child answers right two times in three. A day is spent in one kind of chat, which makes the path of the task written ahead as common as the days that change from one kind to the other; how often families change is not known.
- **Part 3: the code read.** Every path of section 6 was read on `0727341` and checked against the lines cited.
- **Part 4: the drawings and the mock-up.** The copy's `web/` holds the four kinds' types, readers and drawings, their names in every dictionary in the draft's English, and two tests that draw each kind at its example and its limits (`drafts.test.tsx`, `drafts.sizes.test.tsx`); `tsc` and both tests pass. The card was built four times with the copy's Vite configuration for section 10. A script (`web/scripts/t1054-mockup.ts` there) renders the cards of section 13.
- **Kept** in `~/.claude/keep/t105.1/`, which outlives the container: the program, the sets and their builders, the outputs, and the copy's draft as a patch with its base commit.

## 15. Public texts

*Read on `0727341`. The study changes none of these; T105.2 and T105.12 do (section 16).*

| Text | Where | What it says today | What changes |
|---|---|---|---|
| SPEC | 1.2 (`SPEC.md:87-89`) | D05's exclusions stand | Lifted for the 17 genres, with the two new topics in its list of topics |
| | 1.6 (`:185`, `:193-197`) | Seventeen topics at `5-6`, and the batches of reference tasks | The new topics and the 48 reference tasks |
| | 3.7 (`:569-574`) | Why a task written ahead is let go | Also a task solved from the picture in a chat of words |
| | 4.4 (`:667`, `:671`, `:680`, `:684`, `:687`, `:699-713`) | Thirteen kinds; "the picture adds nothing", the wording carries every fact | Seventeen kinds; a task whose facts stand on the picture; the hidden answer's exception and `?` |
| | 4.5 (`:762`), 5.4 (`:833`, `:873-890`), 5.6 (`:924-941`), 5.7 (`:945-950`), 6.2 (`:1057-1059`) | The self-check, the picture's checks, the repeat by the question, a solver of the options alone | The new issue type, the exception, the repeat by the picture, the read rule, `solve(options, picture)` |
| | 7.2, 7.3 (`:1344`, `:1388`), 8.5 (`:1545`), 8.11 (`:1605`) | Without cards the task is read out, and "that is the whole of what it loses" | A chat without cards gets no such task, and a client that declares nothing counts as one for them |
| PRODUCT-V1 | `:20`, 2 (`:44`, `:58`), 4.2 (`:123`), 4.3 (`:142-144`), 4.6 (`:177-178`), the risk at `:367`, 11.2 (`:384`); О-10, О-11, О-30, О-70 | Only the textual part of the niche; thirteen kinds; the meaning of a picture cannot be programmed; no loss without cards | The new genres and kinds; the solver reads the picture; the loss without cards stated (О-76) |
| The site, in 10 languages | `help.md:49`, `terms.md:32`, `index.yaml:79`, `service.yaml:90`, `:129`, `:191`, `topics.yaml` | Without cards the task comes in words; no answer is written on a picture; the topics | Such tasks only where cards show; the exception; two new topics, each a page of its own |
| Directory listings | `plugin/plugin.json:20` and its Russian copy at `:103`; `docs/listing.md:40` | "has no answer written on it" | The exception, through a listing edit the reviewers approve |
| The model's instructions | `content/instructions/task_writing.md:12`, `:68`, `:80`, `:82`, `:104`; `mcp_instructions.md:38` | Everything needed is in the text; without cards the task is read out | The section for such tasks; a task that cannot be read out is not read, and another is asked for |
| Project documents | `docs/architecture/03-flows.md:183`, `:192`; R68, R228, R269, R305 in `docs/decisions.md` | The wording carries every fact; text mode loses nothing | Notes "since" on each decision, as the log keeps them |
| For grants | `docs/grants/README.md:18`, `docs/grants/logic-model.md:39`, `docs/usage-evidence-for-grants.md:193` | 705 reference tasks in 17 topics, 33 templates | The new counts |

- **Unchanged:** README, `docs/parents.md`, `docs/self-hosting.md`, the privacy policy, the site's pages "Why", "Techniques", "Research", "About" and "Coach", and the card's description. The tools' descriptions change only where `next_task` lists the topics (section 6). `docs/load.md` and `docs/research-page.md` are read again once the new reference tasks exist: the first holds the steps of the reference solvers, the second numbers the page computes. The paper is pinned to v0.1.53 and does not describe the picture's rules (`research/paper-a/sections/verification.tex:16-17`).
- **In numbers:** 16 files of the repository, two more to read again. The site's five files are in 10 languages each, 50 edits, and the two new topics' pages are 20 new files.
- The claims of this study about the product hold at `0727341`, and change with the commit it pins next.

## 16. Recommendation, and the tasks it makes

**Recommendation: make all 17 genres**, in the order below. Every genre is proved by a program that reads the picture (sections 2 and 3). The checks hold none back once the hidden answer has its exception (sections 4 and 5). Four kinds draw them on a phone, two of them general enough for twelve genres (section 9). The new topics give about one task in five in a chat with cards (section 7).

**The service's own solver** (section 3) is kept back. It is the one check that catches a solver that reads the picture and answers another question, and it costs a member of the format naming the kind of question and a program in Go for each kind, about three of the probe's genres' worth to begin with. The check before release counts how often a model's solver answers another question; if it does, a task of its own adds the service's solver for those genres.

What it costs, in one place:

- **The service**: the solver's second input and the read rule, about 340 lines in the probe; four kinds of picture with their checks; the chat's kind in the rule; the sketch of a picture in the profile.
- **The card**: 17,470 bytes, 3,294 compressed, with names in 64 languages.
- **The content**: two new topics, 12 ideas in today's topics and 48 reference tasks.
- **The model**: about 900 tokens in a package of a task solved from the picture, and about 50 more for its picture than a word task's.
- **The chats without cards** get none of these tasks (О-76).

**The tasks.** Each is one review (RUN.md rule 12). T105.2 makes the decisions and edits SPEC, as T70.1 did after T70's study. T105.5 to T105.9 merge together, in one branch with a commit each, so that the instructions' version changes once (section 12).

| Task | What | Who | Plan | Depends on |
|---|---|---|---|---|
| T105.2 | Decisions and SPEC: D05 lifted for the 17 genres, the solver reads the picture, the read rule, the hidden answer's exception and `?`, repeats by the picture, the chat's kind for these tasks, the new topics, reference tasks outside the count, the switch for the parent; every section of SPEC section 15 names; PRODUCT-V1 12.2 answered | the author and Claude | yes | T105.1 |
| T105.3 | The solver reads the picture: the second input of `solve`, the picture as Starlark reads it with its reads recorded, the read rule as a check, two templates (a bounded search, placing pieces) | Claude | yes | T105.2 |
| T105.4 | The card draws the four kinds: reader, drawings, names in 64 languages, scenes for `just web-layout`, tiles in a row at most 6 and their labels at most 2 characters. Released alone, before the service knows the kinds | Claude, the author proofreads | yes | T105.2 |
| T105.5 | The service reads the four kinds: format, limits, schema and examples | Claude | — | T105.3, T105.4 released an hour before |
| T105.6 | The checks of a task solved from the picture: the hidden answer's exception and `?`, the repeat by the picture with its sketch in the profile, `facts_in_words` in the self-check | Claude | yes | T105.5 |
| T105.7 | The rule knows the chat: the chat's kind into the rule and the draw of an idea; a task written ahead and a request still open record such a task and are let go in a chat of words; a client that declares nothing counts as one; the skill "reading a picture" for the parent's switch | Claude | yes | T105.5 |
| T105.8 | The two new topics in the catalog, at every level: names in 64 languages, ideas, the count tests for a topic all of whose tasks are solved from the picture, the site's data; `next_task`'s description, which ChatGPT reviews again | the author and Claude | yes | T105.5 |
| T105.9 | The guide's section, only in the packages of such tasks, and the guide's sentences section 15 names; the server's instructions on a task that cannot be read out; the kinds' examples, the templates, the 12 ideas in today's topics. The release of T105.5 to T105.9 | Claude | — | T105.6, T105.7, T105.8 |
| T105.10 | Reference tasks, first batch: 18, one for each genre and level in today's topics, outside the count | Claude, the author proofreads | — | T105.9 |
| T105.11 | Reference tasks, second batch: 30 for the new topics, one at each difficulty of each level, with the genres' simpler forms at `1-2` and `3-4` | Claude, the author proofreads | yes | T105.10 |
| T105.12 | Public texts and the site (section 15) | the author and Claude | yes | T105.11 |
| T105.13 | The check in a live chat, and the release | the author and Claude | yes | T105.12 |

**The check before release** (T105.13) measures what no probe can: how often a model's hand-in of such a task is refused, how many seconds its writing adds, and how many of the pictures accepted write the answer beside three wrong values, which the hidden answer's exception lets through (section 4). Two tasks of each of the 17 genres in Claude, 34 hand-ins, cost about $13 of the author's subscription at run 22's $0.39 a turn, and more if the model thinks longer over a picture. A lesson of each new topic on a phone in ChatGPT is the author's own.

## 17. Open questions

Questions of the product, written in PRODUCT-V1 12.2 from О-76 on. The file is in `.gitignore`, so they are written there locally.

- **О-76. Chats without cards.** PRODUCT-V1 11.2 and О-10 promise that a lesson without cards loses no function. A task solved from the picture goes only where the card is seen (section 6), so such a chat gets word tasks only. The study proposes that 11.2 say so.
- **О-77. A client that declares nothing.** Today it is taken to draw cards (R305). For a task solved from the picture the study proposes the reverse: such a client gets word tasks.
- **О-78. The screen reader.** A switch for the parent, through a skill of the catalog left out, or descriptions read aloud for the five genres solvable by ear (section 11). The study proposes the switch.
- **О-79. One picture, another question.** Is a task that keeps a picture the child was given and asks another question a new task? The study counts it a repeat (section 5).

Questions of the study, for the author in RUN.md's table 3:

- The share of hosts that show the card, from the author's command (section 6).
- The count with the tokenizer, about $0.2, of what section 12 estimates by bytes.

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

## What the author decides after part 3

Decided by the study's recommendation on 2026-11-01, until the author decides otherwise:

- **Where the card is seen.** A task solved from the picture goes only to ChatGPT, or to a client that declared MCP Apps with the call. The rule takes the chat's kind, and a task written ahead or a request still open that holds such a task is let go in a chat of words.
- **The order of releases.** The card that draws the new kinds ships alone and first. The kinds' format and checks, their examples, the guide's section, the rule and the ideas ship together at least an hour and one release later.
- **How often.** The two new topics, with the ideas of part 1 in today's topics; no counter in the rule for now.
- **Reference tasks.** One for each genre at each level it is taught, outside the count of today's topics, and one at each difficulty of each level for a new topic: 48 in all.
- **The model.** The guide's section rides only in a package whose brief holds `facts_on_the_picture`. The section, the new kinds' examples, the templates and the new topics' ideas ship in one release.

## What the author decides after part 4

Decided by the study's recommendation on 2026-11-01, until the author decides otherwise:

- **The genres.** All 17 are made.
- **The limits.** A row of tiles holds six at most, and a tile's label two characters.
- **The screen reader.** A switch for the parent, a skill of the catalog left out; the picture's label stays the kind's name.
- **The tasks.** T105.2 to T105.13 stand in RUN.md's table 1a, and T105.1 is marked done with them. T105.5 to T105.9 merge together, so that the instructions' version changes once.
- **The service's own solver.** Kept back, until the check before release shows models' solvers answering another question.
