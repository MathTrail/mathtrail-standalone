# SPEC — MathTrail v1

> **What this is.** The technical specification: *how* MathTrail v1 is built. *What* it must do is [PRODUCT-V1.md](PRODUCT-V1.md), and the product decisions numbered О-… live in its section 12.1. Implementation decisions numbered R… are in [docs/decisions.md](docs/decisions.md), the architecture diagrams in [docs/architecture/](docs/architecture/), and the frozen prototype materials — cited here as "the prototype's D17" or "the prototype's SPEC 5.6" — in [docs/prototype/](docs/prototype/README.md).
>
> **How it is written.** One section per part of the system, filled in by the tasks of [RUN.md](RUN.md) phase 1. Sections still to be written say so. Until T15 renumbers everything, other documents cite these sections **by name**, not by number.
>
> **Who it is for.** Whoever implements a part of the service and needs the numbers, the formats and the edge cases in one place, without re-deriving them from the product document.

## Contents

| # | Section | Written in |
|---|---|---|
| 1 | [Grade levels, catalogs and reference tasks](#1-grade-levels-catalogs-and-reference-tasks) | T11 |
| 2 | [Ratings and the difficulty corridor](#2-ratings-and-the-difficulty-corridor) | T11 |
| 3 | [Choosing the next task: the rule](#3-choosing-the-next-task-the-rule) | T11 |
| 4 | [The task: what the model gets and what it hands in](#4-the-task-what-the-model-gets-and-what-it-hands-in) | T12 |
| 5 | [The checks a submitted task passes](#5-the-checks-a-submitted-task-passes) | T12 |
| 6 | The Starlark solver | T13 |
| 7 | The MCP tools | T14 |
| 8 | Widgets, screens and languages | T14 |
| 9 | Sign-in and tokens | T15 |
| 10 | Limits | T15 |
| 11 | Configuration | T15 |
| 12 | Logging and metrics | T15 |

Sections 6–12 are not written yet. Section 9 will be assembled from [docs/architecture/02-auth.md](docs/architecture/02-auth.md), and sections 7–8 from [03-flows.md](docs/architecture/03-flows.md), [04-profile.md](docs/architecture/04-profile.md) and [05-storage.md](docs/architecture/05-storage.md).

---

# 1. Grade levels, catalogs and reference tasks

## 1.1 The three grade levels

A child has a **grade** from 1 to 6; everything else works with the **level** the grade falls into.

| Level | Grades | Longest sentence | Flesch–Kincaid (English only) | Reference tasks |
|---|---|---|---|---|
| `1-2` | 1, 2 | 20 words | grade + 3 | from the prototype |
| `3-4` | 3, 4 | 25 words | grade + 3 | from the prototype |
| `5-6` | 5, 6 | **30 words** | grade + 3 | written in T37–T39 |

The sentence limits for `1-2` and `3-4` are the prototype's D38, measured on its 450 reference tasks; `5-6` continues the series, as О-30 requires. The Flesch–Kincaid margin of +3 is D38's and applies to English tasks only (the prototype's D42). The checks themselves are section 5.

The level governs four things: which topics exist, what "difficulty 3" means, the readability limits above, and which reference tasks the model is shown. **Difficulty is a level from 1 to 5 inside the level** — a difficulty-3 task for grades 1–2 and one for grades 5–6 are different tasks (the prototype's D36).

## 1.2 The topic catalog

A topic is chosen **from the catalog**; the model never invents one, or histories from different children stop being comparable. Two rules decide what may be in it:

1. **Words only.** The task must be statable in words, with no picture and no spatial imagination.
2. **Brute-forceable.** A short program must be able to enumerate the possibilities and confirm that exactly one option is correct. This is what the solver check rests on (section 6), and it is not negotiable: a topic that cannot be checked by enumeration does not enter the catalog.

### The ten topics of grades 1–4, extended to 5–6

All ten carry over unchanged and all are now available at `5-6` as well. What changes with the level is the size of the search, the number of conditions and the arithmetic allowed, not the idea.

| id | Topic | Levels | What "harder" means at `5-6` |
|---|---|---|---|
| `logic.ordering` | Ordering | 1-2, 3-4, 5-6 | More people and more comparisons, including indirect ones and a contradiction to rule out |
| `logic.knights_liars` | Knights and liars | 3-4, 5-6 | Three or four speakers, statements about statements |
| `combinatorics.enumeration` | Enumeration | 1-2, 3-4, 5-6 | Two restrictions at once, counting by cases, small products |
| `counting.gaps` | Gaps and boundaries | 1-2, 3-4, 5-6 | Two-dimensional arrangements, cuts in several directions, closed rings |
| `time.clocks` | Clocks | 1-2, 3-4, 5-6 | Several time zones or a clock that runs fast, intervals crossing midnight |
| `time.calendar` | Calendar and age | 1-2, 3-4, 5-6 | Leap years, "the same weekday next year", ages in the past and the future |
| `pigeonhole.basic` | Pigeonhole principle | 1-2, 3-4, 5-6 | Worst case over two attributes at once, "how many to be sure of three of a kind" |
| `parity.alternation` | Parity and alternation | 1-2, 3-4, 5-6 | Parity as an invariant: which final states are reachable at all |
| `arithmetic.tricks` | Arithmetic with a trick | 1-2, 3-4, 5-6 | Longer chains, "I thought of a number" run backwards, digit puzzles |
| `algorithms.weighing_pouring` | Weighing and pouring | 3-4, 5-6 | Three jugs or three weighings, the shortest sequence rather than any sequence |

### The seven topics added for grades 5–6

О-12 asked for percentages, fractions, geometry in words "and others"; О-30 allows geometry only as a provably enumerable subset. Each row states how a program checks it, because that is the entry requirement.

| id | Topic | Description | How a program checks it |
|---|---|---|---|
| `fractions.parts` | Parts and shares | A part of a whole, what is left, comparing simple fractions | Exact integer arithmetic over a small whole; the solver enumerates the candidate wholes or parts and compares fractions as pairs of integers |
| `percent.basic` | Percentages | Percentages of a quantity, discounts and increases, "what percentage is it" | The same, with the percentage as a fraction of hundredths; no floating point anywhere |
| `ratio.sharing` | Ratios and sharing | Sharing a quantity in a ratio, "how many more", recipes scaled up | Enumerate the integer splits of a bounded total and keep those that satisfy the ratio |
| `geometry.grid` | Figures on a square grid | Perimeter and area counted in cells, cutting along cell lines, figures made of cells on coordinates | The grid is small and finite: the solver enumerates cells, cuts or placements. This is exactly the enumerable subset О-30 allows, and nothing outside it is permitted — no free-hand geometry in words |
| `number.divisibility` | Divisibility and remainders | Divisibility rules, remainders, the smallest number with a property | Iterate over a bounded range and test the property |
| `logic.sets` | Overlapping groups | How many do both, how many do neither, the largest or smallest possible overlap | Enumerate the integer counts that satisfy every stated bound |
| `games.strategy` | Games with a winning strategy | Two players take turns, who wins with correct play, what the first move must be | A full game-tree search over a small state space: the position is a few small integers |

**If `geometry.grid` fails in practice** — that is, if T37 cannot write reference tasks whose solvers genuinely enumerate — it is dropped and replaced by another enumerable topic from the same list (averages, digit puzzles, or ratios of times). О-30 settled this in advance so that no new round of decisions is needed.

### What stays out

The prototype's D05 exclusions stand, and the new topics do not reopen them: spatial geometry (nets of cubes, rolling a cube, folding and cutting paper), visual counting ("how many triangles"), symmetry and reflections, geometry described in words beyond the grid subset above, heavy arithmetic for its own sake, school motion-and-work problems, and patterns in sequences — where several continuations are always plausible and no program can prove the rule is unique.

## 1.3 The trap catalog

A trap is the mistake behind a wrong option. The fourteen of the prototype carry over unchanged; six are added for the new topics. Every wrong option in every task names one of these ids (section 4).

| Added id | Description | Mostly used in |
|---|---|---|
| `percent_wrong_base` | Took the percentage of the wrong quantity — of the new price instead of the old one | `percent.basic`, `ratio.sharing` |
| `part_whole_swap` | Swapped the part and the whole: answered with the part when asked for the whole, or the other way round | `fractions.parts`, `percent.basic` |
| `ratio_total_confusion` | Treated a share of a ratio as a share of the total, or added the ratio parts wrongly | `ratio.sharing` |
| `remainder_vs_quotient` | Gave the quotient where the remainder was asked, or the other way round | `number.divisibility` |
| `area_perimeter_swap` | Counted the perimeter where the area was asked, or the other way round | `geometry.grid` |
| `first_move_assumed` | Assumed the player who moves first always wins, without checking the position | `games.strategy` |

Twenty traps in all. The catalog stays closed for the same reason the topic catalog does: `trap_hit` is what makes a wrong answer a diagnosis, and free-form trap names cannot be counted.

## 1.4 The catalog of constraint skills

`excluded_skills` names what the child has not met at school yet, and a task may use none of it — neither in the wording nor in a trap. The prototype's fifteen skills carry over; ten are added, all of them things a grade 5–6 task might reach for and a grade 3 child will not have seen.

| Added id | Description |
|---|---|
| `fractions_arithmetic` | Adding, subtracting and comparing fractions with different denominators |
| `decimals` | Decimal notation and arithmetic with it |
| `percentages` | Percentages as hundredths, and percentage of a quantity |
| `ratios` | Ratios and proportion |
| `negative_numbers` | Numbers below zero |
| `powers_squares` | Squares, cubes and powers |
| `area_perimeter` | Area and perimeter as quantities |
| `coordinates` | Points and coordinates on a grid |
| `averages` | The arithmetic mean |
| `divisibility_rules` | Divisibility rules and remainders |

Twenty-five skills in all. The cap on `excluded_skills` in a profile is the size of the catalog (04-profile).

## 1.5 The format of a reference task

Reference tasks are few-shot examples: they show the model the shape, the style and — most of all — how a trap is attached to every wrong option. They live in the binary as `content/examples/<topic>.json`, one file per topic, each an array of tasks.

```json
{
  "id": "wp-56-d3-2",
  "topic": "algorithms.weighing_pouring",
  "grade_level": "5-6",
  "difficulty": 3,
  "question": "…",
  "drawing": "…",
  "drawing_structure": { },
  "options": { "A": "…", "B": "…", "C": "…", "D": "…", "E": "…" },
  "correct_answer": "D",
  "hint": "…",
  "solution": "…",
  "distractors": {
    "A": { "trap": "stopped_early", "text": "…" },
    "B": { "trap": "off_by_one", "text": "…" },
    "C": { "trap": "missed_case", "text": "…" },
    "E": { "trap": "ignored_condition", "text": "…" }
  },
  "solver": "…"
}
```

| Field | Required | Notes |
|---|---|---|
| `id` | yes | `<topic abbreviation>-<level>-d<difficulty>-<number>`, unique across the content |
| `topic`, `grade_level`, `difficulty` | yes | Ids from the catalogs; difficulty 1–5 inside the level |
| `question` | yes | English. The model writes in the chat's language; the examples set the idea and the structure, not the language |
| `drawing`, `drawing_structure` | no | A text drawing and its structural description, where the topic needs one — mostly `geometry.grid`. The format is section 4, the frames are T36b |
| `options` | yes | Exactly five, all different |
| `correct_answer` | yes | Exactly one letter |
| `hint` | new tasks only | A nudge that does not give the answer away. The 450 ported tasks have none; the format the model must produce always does |
| `solution` | yes | Short, step by step |
| `distractors` | yes | One entry per wrong option: a trap id from the catalog and the text the child sees after answering |
| `solver` | yes, after T29–T31 | The Starlark program that brute-forces this very task. It makes every example re-checkable on the bench, and it is not what goes into the package — the model is shown the generalised templates of `content/solvers/` instead (R08) |

**Nobody else's wording, anywhere.** The ideas come from open sources; the text is ours (the prototype's D23 and research/12). Ideas are not protected by copyright, wordings are.

## 1.6 How many reference tasks, and in which batches

**Grades 1–4 — 450 tasks, unchanged.** Five per (topic × level × difficulty) across the eighteen topic-and-level pairs of the prototype. They are ported as they are (T23); T29–T31 add a Starlark solver to each.

**Grades 5–6 — 153 new tasks.** Seventeen topics exist at `5-6` (the ten extended plus the seven new), and each gets **nine tasks: three each at difficulties 2, 3 and 4**.

Why nine rather than the prototype's twenty-five per pair. The package hands the model three examples of the brief's topic, preferring the same difficulty (section 3). Three at each of the three middle difficulties satisfies that exactly where it matters, and difficulties 1 and 5 — the extremes a rule aiming at the middle of the corridor rarely asks for — fall back to the nearest anchor, labelled with its own difficulty so the model can see which way to lean. Twenty-five per pair would be 425 new tasks, each needing a wording, five options, five trap labels, a solution, a hint and a solver, all proofread by hand; nine per pair buys the coverage that the selection rule can actually use.

| Batch | Task | What is in it | Count |
|---|---|---|---|
| 1 | T37 | The seven new topics: `fractions.parts`, `percent.basic`, `ratio.sharing`, `geometry.grid`, `number.divisibility`, `logic.sets`, `games.strategy` | 63 |
| 2 | T38 | Five extended topics: `logic.ordering`, `logic.knights_liars`, `combinatorics.enumeration`, `counting.gaps`, `time.clocks` | 45 |
| 3 | T39 | Five extended topics: `time.calendar`, `pigeonhole.basic`, `parity.alternation`, `arithmetic.tricks`, `algorithms.weighing_pouring` | 45 |

Three batches, so no T39a is needed. The first batch is the largest on purpose: it is the one that settles whether the new topics survive contact with a solver, `geometry.grid` above all.

**Which three examples go into the package.** For the brief's topic and level: three tasks of the requested difficulty; if there are fewer, top up from the nearest difficulty, then from the next nearest; if the topic has nothing at that level, from the level below. Which three, when there are more than three, rotates by the child's answer count, so a child asking for the same topic twice does not see the same examples (the prototype's D43).

---

# 2. Ratings and the difficulty corridor

## 2.1 The model

The child's level and the task's difficulty live on one scale — a Rasch model with a guessing floor, updated Elo-style after every answer (the prototype's D17). Ratings are computed by code, never by the model.

**P = 0.2 + 0.8 · σ(θ + δ_topic − β)**, where σ(x) = 1 / (1 + e^(−x))

| Symbol | Meaning | Where it lives |
|---|---|---|
| θ | The child's overall level | `ratings.theta` (04-profile) |
| δ_topic | The correction for one topic; 0 for a topic never met, so a new topic starts from the overall level | `topics[t].delta` |
| β | The difficulty of the task: **β = difficulty − 3**, so 1 → −2 and 5 → +2 | Derived, not stored |
| 0.2 | The guessing floor: five options and no penalty for a wrong answer | Constant |

A fresh profile, θ = 0 and δ = 0, therefore expects:

| difficulty | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| P | 0.9046 | 0.7848 | 0.6000 | 0.4152 | 0.2954 |

## 2.2 What happens after an answer

With S = 1 for a correct answer and S = 0 for a wrong one:

- θ += K_θ · (S − P)
- δ_topic += K_δ · (S − P)

Each K decays with the number of answers behind it: **K = K₀ / (1 + a · n)**.

| Constant | Value | n counts |
|---|---|---|
| K₀ for θ | 0.2 | the child's answers, `ratings.answers` |
| K₀ for δ | 0.4 | the answers in that topic, `topics[t].answers` |
| a | 0.05 | — |

The two K₀ differ for a reason the prototype measured (its D37): with one K₀ = 0.4 the topic level θ + δ moved by almost 0.8 · (S − P) on the first answers, and two failures in a row shifted a child by nearly a whole difficulty level.

**β is not updated in v1.** The prototype kept a bank and re-used tasks, so a task's difficulty was worth refining. Here every task is written for one child and never handed out again (PRODUCT 4.4), so β is computed from the task's difficulty and discarded. The paid edition, calibrating across many children, is where the third update belongs (PRODUCT 4.5).

**Only correctness enters the formula** (О-33). The hint, the pace, "I don't understand" and the run of failures are recorded (04-profile) and change what is chosen next (section 3) — never the rating, which stays an honest estimate of what the child knows. The pace tag itself is computed by the server from the time between handing the task out and the answer: `fast` under 60 seconds, `slow` over 180, `normal` in between, as in the prototype; T15 confirms the numbers in the configuration.

## 2.3 The corridor and the recommended difficulty

The **corridor** is the range of difficulties where the probability of a correct answer is P ∈ [0.70, 0.85]. Math Garden keeps success near 0.75 and the "85 % rule" gives the upper bound (the prototype's research/06).

On the β scale the corridor is the interval **[θ + δ − 1.4663, θ + δ − 0.5108]**, which is 0.9555 wide — narrower than the gap of 1.0 between difficulty levels. So **the corridor holds at most one level, and sometimes none**, and after every wrong answer it slides toward easier tasks, whether or not the recommended level moves with it.

The **recommended difficulty** is therefore defined without needing the corridor to be non-empty: the level from 1 to 5 whose P is closest to the middle of the corridor, **0.775**, carrying a marker of where it landed — `inside`, `harder than the corridor` (P < 0.70) or `easier than the corridor` (P > 0.85).

## 2.4 The chess scale and the ranks

What the child is shown is not θ but a chess-style number (PRODUCT 4.5):

**R = 1500 + (400 / ln 10) · rating**, with 400 / ln 10 = 173.717793

Applied to θ it gives the overall rating; applied to θ + δ_topic, the rating for a topic. It is displayed rounded to a whole number.

Above the number comes a rank: five steps, named by dictionary keys and translated in T59 (О-48, R12). The step width is one corridor — 0.9555 on the θ scale, **166 rating points** — centred on the starting 1500, so that moving up a rank means roughly "tasks a whole corridor harder than before are now in reach":

| Rank | Overall rating R |
|---|---|
| 1 | below 1334 |
| 2 | 1334 – 1499 |
| 3 | 1500 – 1665 |
| 4 | 1666 – 1831 |
| 5 | 1832 and above |

The rank is computed when it is drawn and is not stored (R12).

## 2.5 When a topic counts as mastered

О-32 asked for an automatic, deterministic criterion. It is:

> A topic becomes **mastered** when the child has answered it at least **5** times and has just finished a run of **3** answers that were all correct, each on a task whose P at the moment it was handed out was **≤ 0.775**, and none of which used a hint.

Each part earns its place: P ≤ 0.775 means "the harder half of the corridor or harder", so easy tasks cannot add up to mastery; the hint would make the run measure the hint instead of the child; five answers keep a lucky start from mastering a topic on the third task; three in a row is short enough to be reachable and long enough to rule out guessing, which is worth 0.2 per attempt.

The run is `topics[t].top_streak` in the profile: it grows on an answer meeting all the conditions, resets to zero on a wrong answer, and stays put on a correct but easy or hinted answer.

**Mastery is lost** after **two consecutive wrong answers** in that topic: `mastered_since` is cleared and the topic returns to the rotation. One wrong answer is a slip and does not undo anything.

A mastered topic is skipped when the rule looks for something new, until every topic of the level is mastered — after which all of them are back in the rotation (section 3).

## 2.6 Worked example 1: a new child, ten answers

Two topics, A = `combinatorics.enumeration` and B = `parity.alternation`, starting from an empty profile. θ and δ are the values **before** the answer, θ′ and δ′ the values after; the corridor column is the β interval after the update, and "next d" is what the rule would ask for next in that topic.

| # | topic | d | β | θ | δ | P | S | θ′ | δ′ | corridor in β | next d | run |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | A | 3 | +0 | 0.0000 | 0.0000 | 0.6000 | 1 | 0.0800 | 0.1600 | −1.23 … −0.27 | 2 (P 0.82, inside) | 0 |
| 2 | A | 3 | +0 | 0.0800 | 0.1600 | 0.6478 | 1 | 0.1471 | 0.2942 | −1.03 … −0.07 | 2 (P 0.85, inside) | 0 |
| 3 | A | 4 | +1 | 0.1471 | 0.2942 | 0.4911 | 0 | 0.0578 | 0.1156 | −1.29 … −0.34 | 2 (P 0.81, inside) | 0 |
| 4 | A | 3 | +0 | 0.0578 | 0.1156 | 0.6346 | 1 | 0.1214 | 0.2427 | −1.10 … −0.15 | 2 (P 0.84, inside) | 0 |
| 5 | B | 3 | +0 | 0.1214 | 0.0000 | 0.6242 | 0 | 0.0173 | −0.2497 | −1.70 … −0.74 | 2 (P 0.75, inside) | 0 |
| 6 | B | 2 | −1 | 0.0173 | −0.2497 | 0.7464 | 1 | 0.0579 | −0.1531 | −1.56 … −0.61 | 2 (P 0.77, inside) | 0 |
| 7 | A | 4 | +1 | 0.0579 | 0.2427 | 0.4656 | 1 | 0.1401 | 0.4209 | −0.91 … 0.05 | 3 (P 0.71, inside) | 1 |
| 8 | A | 4 | +1 | 0.1401 | 0.4209 | 0.5136 | 1 | 0.2122 | 0.5765 | −0.68 … 0.28 | 3 (P 0.75, inside) | 2 |
| 9 | A | 4 | +1 | 0.2122 | 0.5765 | 0.5579 | 1 | 0.2753 | 0.7125 | −0.48 … 0.48 | 3 (P 0.78, inside) | **3 → mastered** |
| 10 | B | 3 | +0 | 0.2753 | −0.1531 | 0.6244 | 1 | 0.3271 | −0.0165 | −1.16 … −0.20 | 2 (P 0.83, inside) | 0 |

After the tenth answer: θ = 0.3271, an overall rating of **1557**; topic A sits at θ + δ = 1.0397, a rating of **1681** over 7 answers, mastered at step 9; topic B at 0.3106, a rating of **1554** over 3 answers.

**Step 1 by hand.** θ = 0, δ = 0, difficulty 3 so β = 0. Then σ(0 + 0 − 0) = 0.5 and P = 0.2 + 0.8 · 0.5 = **0.6**. The answer is correct, S = 1, so S − P = 0.4. No answers have been given yet, so n = 0 for both: K_θ = 0.2 / (1 + 0.05 · 0) = 0.2 and K_δ = 0.4. Hence θ = 0.2 · 0.4 = **0.08** and δ = 0.4 · 0.4 = **0.16**, and the topic rating becomes 1500 + 173.7178 · 0.24 ≈ **1542**.

**Mastery at step 9.** By then topic A has seven answers, and steps 7, 8 and 9 were three correct answers in a row at difficulty 4 with P of 0.4656, 0.5136 and 0.5579 — all under 0.775 — with no hint. Step 7 is where the run starts rather than step 4, because at step 4 the topic still had fewer than five answers.

## 2.7 Worked example 2: three failures in a row

A settled child — θ = 0.9 over 20 answers, topic A at δ = 0.3 over 8 answers — gets difficulty 3 three times and fails all three. This is the path the rule takes after a failure (section 3), and it shows the corridor sliding.

| # | d | β | θ | δ | P | S | θ′ | δ′ | corridor in β | next d |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 3 | +0 | 0.9000 | 0.3000 | 0.8148 | 0 | 0.8185 | 0.0672 | −0.58 … 0.37 | 3 (P 0.77, inside) |
| 2 | 3 | +0 | 0.8185 | 0.0672 | 0.7664 | 0 | 0.7437 | −0.1442 | −0.87 … 0.09 | 3 (P 0.72, inside) |
| 3 | 3 | +0 | 0.7437 | −0.1442 | 0.7164 | 0 | 0.6755 | −0.3353 | −1.13 … −0.17 | **2** (P 0.83, inside) |
| 4 | 2 | −1 | 0.6755 | −0.3353 | 0.8340 | 1 | 0.6910 | −0.2924 | −1.07 … −0.11 | 2 (P 0.84, inside) |

The corridor moves toward easier tasks after **every** failure — its β interval slides from [−0.58, 0.37] to [−1.13, −0.17] — while the recommended level, which can only be a whole number, holds at 3 for two failures and drops to 2 on the third. That is the intended behaviour and worth knowing before anyone reports it as a bug: one slip does not change what the child is offered, a run of them does.

## 2.8 What T25 must reproduce

Both tables are the test vectors for the ratings package. The implementation computes in `float64` and must match the four decimals printed here to within 1e-4, and the ratings rounded to whole numbers exactly. The inputs are fully specified: the constants of 2.1 and 2.2, β = difficulty − 3, and the step sequences above. The golden files carried over from the prototype (T16) are a separate, larger set; these two examples exist so that a failure can be read by eye.

---

# 3. Choosing the next task: the rule

## 3.1 What the rule is for

The rule builds the **brief** — the task's terms of reference — from the profile alone, with no model involved and no randomness. It runs inside `next_task` (03-flows) and its output goes to the chat's model as a recommendation the model may adjust (3.4).

It is deterministic by design (the prototype's D40): the same profile always produces the same brief. That is what makes it a baseline the model can be compared against, and what makes it testable without mocks.

| Brief field | Where it comes from |
|---|---|
| `pedagogical_goal` | `reinforce` or `new_topic` — 3.2 |
| `target_concept` | A topic id from the catalog — 3.2 |
| `difficulty` | 1–5, the recommended difficulty for that topic (2.3) |
| `setting` | One of the child's interests, rotated — 3.2 |
| `traps_to_use` | Two trap ids — 3.2 |
| `excluded_skills` | Copied from the profile, all of them |
| `constraints` | Wording constraints, if any: length, vocabulary |
| `rationale` | One sentence naming what the choice was based on, and the model's reason when the model overrode the choice |

`motivate`, the prototype's third goal, is gone (О-34): the corridor already keeps tasks within reach, and a goal the rule never issued was dead weight in every test. The prototype's `profile_fields_used` is gone too — it existed to measure which profile fields mattered, and with a deterministic rule and aggregate-only logs (О-16) it has no reader.

## 3.2 The algorithm

**Goal and topic.**

1. If `ratings.consecutive_failures` > 0 → the goal is `reinforce` and the topic is **the topic of the last entry in `recent`** — the one just failed. (The window is never empty when the counter is above zero; 04-profile.)
2. Otherwise the goal is `new_topic` and the topic is chosen among the topics of the child's **level** that are **not mastered**: those never handed out come first, in catalog order; then the one whose `topics[t].last_issued` is the oldest.
3. If every topic of the level is mastered, the same choice runs over all of them.

**Difficulty.** The recommended difficulty for that topic (2.3), with its marker. A topic never met has δ = 0, so the child's overall level decides — a cold start that is right far more often than a fixed "start at 3".

**Setting.** The interests are walked in a circle: `interests[ratings.answers mod len(interests)]`. The prototype rotated by the number of history rows; with a bounded window (О-40) that number would start repeating, and the rule would quietly stop being deterministic, so the total answer count takes its place. No interests, no setting: the model picks one.

**Traps.** Two ids, in this order:

1. The child's most frequent traps in this topic, from `topics[t].traps`, most frequent first and ties broken by catalog order.
2. If fewer than two, top up with the most frequent traps of the reference tasks of that topic and level — the 450 examples and their successors carry trap labels already, so no separate list has to be kept in step (the prototype's D40).

**Constraints and prohibitions.** `excluded_skills` is copied whole from the profile. The child's free-form notes (О-31) do not enter the brief: they travel in the package as context for the model's tone and level, and the rule never reads them.

## 3.3 What changed from the prototype

| Prototype | v1 | Why |
|---|---|---|
| Searches the task bank first, and the brief is only needed when nothing fits | No bank at all: every task is written on the spot | О-6 — a buffer is a paid-edition feature |
| Reads the whole history | Reads the window and the per-topic summary | О-40 and the file size budget (04-profile) |
| Three goals, including `motivate` | Two | О-34 |
| Two grade levels | Three | О-12 |
| Setting rotates by the number of history rows | By the total answer count | The window makes the row count non-monotonic |
| `profile_fields_used` in the brief | Dropped | No reader left |

## 3.4 When the model chooses instead

The model may ask for a different topic or difficulty, and must give a reason (PRODUCT 4.1, the prototype's D43). It passes `topic`, `difficulty` and `reason` to `next_task`; the examples, the traps and the corridor all follow **its** choice, `rationale` keeps both its reason and the rule's suggestion, and `open_request.tutor_mode` becomes `llm` instead of `rule` (04-profile).

**The goal stays the rule's** (the prototype's D44). A brief may therefore read `reinforce` on a topic the child has never practised: that records "a failure just happened and the child asked for something else", which is a fact about the child, not a defect. The goal describes the child's situation; `tutor_mode` records who chose the topic.

When the model sets a difficulty of its own, that difficulty is used as given — the corridor is a recommendation, not a fence, and a deliberate step outside it is exactly what a tutor sometimes does.

## 3.5 The rule on the example profile

Running the rule on the profile printed in 04-profile: `consecutive_failures` is 1, so the goal is `reinforce` and the topic is the one from the last entry of the window, `combinatorics.enumeration` — the task just failed.

That topic stands at θ + δ = 0.42 + 0.31 = 0.73, which puts the corridor at β ∈ [−0.74, 0.22] and the five levels at P = 0.9510, 0.8795, 0.7398, 0.5463 and 0.3754. The nearest to 0.775 is **difficulty 3**, at P = 0.7398, inside the corridor; difficulty 2 at 0.8795 is easier than the corridor. So the child is offered the same difficulty they have just failed — the corridor has already slid toward easier tasks, but not yet far enough to cross a level, exactly as 2.7 describes. A second failure would move it.

The setting is `interests[57 mod 3]` = `interests[0]` = `space`. The traps are the child's most frequent in that topic: `missed_case`, seen three times, then `double_count`, seen once. The prohibitions are `division_with_remainder`. The rationale names the failure and the topic.

Every number in that paragraph comes from the file in 04-profile and from 2.1–2.3, which is the point: the rule reads nothing else.

---

# 4. The task: what the model gets and what it hands in

The chat's model writes the task; the service makes no LLM calls of its own (PRODUCT 4.3). This section is the contract between the two: what goes out in the package, and what has to come back in `submit_task`. The JSON schemas themselves are written in T23 and live in the binary; the field lists here are what those schemas encode.

## 4.1 The package for the model

Returned by `next_task`, never rendered as a card (03-flows), and assembled fresh for each request:

| Part | What it is | Size |
|---|---|---|
| The brief | 4.2, built by the rule or adjusted by the model (section 3) | ~0.6 KB |
| The corridor | The recommended difficulty, its marker, and the corridor as a β interval (2.3) | small |
| The topic | Its id, name and description from the catalog | small |
| The traps | The whole catalog of 20, with descriptions — the two in the brief are a recommendation, and the model needs the others to choose an alternative that fits its plot | ~2 KB |
| The prohibitions | The description of every skill in `excluded_skills` | ~0.5 KB |
| The child's context | Grade, interests, and the free-form notes, capped at 500 characters (О-31). The pseudonym is **not** repeated here — it has no business in a task, and the package is the one place it is easy to leave out | ≤ 0.8 KB |
| Three reference tasks | 4.1.1 | ~3.6 KB |
| Solver templates | One or two for this topic from `content/solvers/`, marked as samples one may depart from (О-43, R08) | ~2 KB |
| Drawing frames | The frames for this topic from `content/drawings/`, if it has any (О-44, R09) | ~1 KB |
| The formats | The fields the model must return, 4.3–4.5, as prose plus one worked example | ~2 KB |
| The limits | The readability limits for the level (1.1) and the drawing limits (5.4) | small |
| The self-check checklist | 4.5 | ~0.7 KB |
| The instructions version | The hash that every log line about this task will carry (О-21) | small |

**The budget is 16 KB.** A typical package is around 13 KB, which is a few thousand tokens of the chat's own context — a cost the family pays out of their message limit, so it is kept deliberately small. If a package would exceed the budget, parts are dropped in this order: the third reference task, the second solver template, the drawing frames beyond the first. The brief, the catalogs, the formats and the checklist are never dropped.

On a repeat attempt the package is not sent again: `submit_task` answers with the refusal codes, and the model already has everything else in its context (03-flows).

### 4.1.1 Which three reference tasks

Three examples of the brief's topic and level: of the requested difficulty first, then the nearest difficulty, then the next nearest, then the level below (1.6). Where more than three are available — every grade 1–4 cell has five — the three rotate by `ratings.answers`, so a child asking for the same topic twice does not get the same examples twice (the prototype's D43). At `5-6` a cell holds exactly three and the rotation has nothing to do.

Reference tasks are in English whatever the chat's language: they carry the idea, the structure and the trap labelling, not the wording (1.5).

## 4.2 The brief

Built by the rule (section 3) and handed back by the model with `submit_task`, which is where it may have adjusted the plot and the traps (the prototype's D43).

| Field | Type | Notes |
|---|---|---|
| `pedagogical_goal` | `reinforce` \| `new_topic` | The rule's, never the model's (the prototype's D44, О-34) |
| `target_concept` | topic id | From the catalog |
| `difficulty` | 1–5 | Inside the level |
| `setting` | string | The plot: one of the child's interests, or the model's own if there are none |
| `traps_to_use` | array of trap ids | Two from the rule; the model may swap them for others from the catalog |
| `excluded_skills` | array of skill ids | Copied from the profile and never reduced |
| `constraints` | array of strings | Wording constraints, may be empty |
| `rationale` | string | One sentence: what the choice rests on, and the model's reason if it overrode the rule |

## 4.3 The task

What `submit_task` carries, together with the request id, the brief, the solver program and the self-check. The language is a separate argument, not a field (the prototype's 5.2).

| Field | Type | Why it exists |
|---|---|---|
| `core_idea` | string | The mathematical core before the plot: the idea and why the answer is what it is. Writing from the idea outwards works better than writing a story and hoping (the prototype's research/03) |
| `design_thought_process` | string | How the plot and the traps were chosen. Not shown to anybody; it makes the model state its reasoning before committing to it |
| `question` | string | The wording the child reads, in the chat's language |
| `drawing`, `drawing_structure` | string, object | Optional — 4.4 |
| `options` | object `A`–`E` | Exactly five, all different |
| `correct_answer` | `A`–`E` | Exactly one |
| `hint` | string | One leading question or the first step. It must not give the answer away, and the child sees it on request before answering |
| `solution` | string | Step by step, in the language a tutor would use with a child of that grade. Returned after the answer |
| `distractors` | object | One entry per wrong option: `trap` (a catalog id) and `text` (what the child reads after choosing it) |

`distractors` is the heart of the product: `trap` is what turns a wrong answer into a diagnosis (`trap_hit` in the profile), and `text` is what the child actually gets told. What it must satisfy is 5.3.

Everything that gives the answer away — `correct_answer`, the `distractors`, the `solution` and the solver — goes into the sealed block of the profile and reaches nobody until the child has answered (04-profile, 03-flows).

## 4.4 The text drawing

The model draws, following the rules in the instructions (О-11); the service checks the format, always and programmatically (О-11а). Both fields are optional and come together: a drawing without its structure cannot be checked against the wording.

**`drawing`** — a block of monospaced text, lines separated by `\n`. The rules the model is given:

- draw only when the wording genuinely needs it, and never to decorate;
- a number line, a grid, a balance, a pouring diagram, a clock face — the recurring subjects have ready frames in the package (О-44), and a frame is a starting point, not an obligation;
- keep it inside the limits of 5.4: they are what a phone can show;
- label the objects the question names, with the same labels, in the same alphabet;
- no colour, no trailing spaces, no decoration characters outside the allowed set.

**`drawing_structure`** — the same picture as data, so the service can compare it with the wording without understanding either:

```json
{
  "kind": "number_line",
  "objects": [
    { "id": "A", "label": "A", "value": 3 },
    { "id": "B", "label": "B", "value": 7 }
  ],
  "relations": [
    { "type": "left_of", "from": "A", "to": "B" }
  ]
}
```

`kind` is free text naming the subject (`number_line`, `grid`, `balance`, `pouring`, `clock`); `objects` carry an id, the label as it appears in the drawing, and an optional value; `relations` are triples. The service reads only the labels (5.4); `kind` and `relations` are there for the model's own discipline and for the frames, and no check depends on them beyond their presence being well-formed.

## 4.5 The self-check

Before handing in, the model checks its own task against an explicit checklist and submits the result. Without being asked directly, models notice an ambiguity but rarely mention it (the prototype's research/05).

| Field | Type | Notes |
|---|---|---|
| `issues` | array of `{type, severity, comment}` | `type`: `ambiguous`, `missing_data`, `multiple_correct`, `no_correct`, `too_hard_for_grade`, `needs_picture`, `factual_error`. `severity`: `blocking` or `minor` |
| `option_check` | object `A`–`E` | Why each option is right or wrong — the model's own pass over all five |
| `final_answer` | `A`–`E` or `UNSOLVABLE` | The answer the model arrives at when it solves its own task afresh |

The checklist, sent in the package, asks the model to look for: missing data; a wording that can be read two ways; negations; vague words and ranges ("several", "about"); conditions that contradict each other; and a correct option that answers a different question than the one asked.

The self-check is weaker than an independent reviewer would be — the model is marking its own homework — so the service never relies on it alone (section 5). It is one of five independent reasons a task can be rejected, and the only one that can see an ambiguity.

---

# 5. The checks a submitted task passes

## 5.1 The order, and what the model is told

The checks run in the order below. It is chosen so that nothing expensive runs before the cheap thing that would have rejected the task anyway, and — the rule of 05-storage — so that everything not needing the profile happens before the profile is read:

| # | Check | Code | Needs |
|---|---|---|---|
| 1 | Structure | `bad_structure` | the submission alone |
| 2 | The explanations behind the wrong options | `distractor_explanations` | the submission alone |
| 3 | The drawing's format | `drawing_format` | the submission alone |
| 4 | The drawing against the wording | `drawing_mismatch` | the submission alone |
| 5 | Readability | `readability` | the submission and the grade |
| 6 | The solver runs | `solver_error` | the Starlark sandbox |
| 7 | The solver and the self-check agree with the answer | `solver_disagrees` | the sandbox's result |
| 8 | The self-check has no blocking issue | `self_check_blocking` | the submission alone |
| 9 | Not a near-duplicate | `near_duplicate` | **the profile** |

**Every failed check is reported at once**, so the model can fix everything in one more attempt; the **first failure in this order is the primary code**, the one that goes into the log and the one the state machine counts (03-flows). This is the prototype's rule and its reason holds: with a median generation of 69 seconds, a second attempt that fixes one thing and trips over the next is a minute of a child's patience spent for nothing.

A check whose input is missing does not run and is reported as blocked: with no `options` there is nothing for the solver to decide, and saying "the solver failed" would send the model hunting in the wrong place.

**No refusal ever quotes an answer letter or an option's text.** `submit_task` draws a card (03-flows), and everything in its result reaches the widget; the model holds its own draft and needs no quoting to fix it.

## 5.2 Structure

Mechanical, and all of it deterministic:

- the submission matches the schema: every required field present, of the right type, non-empty where a string is expected;
- exactly five options, `A`–`E`, and all five texts different after trimming;
- `correct_answer` is one of them;
- `distractors` has exactly the four other letters, no more and no fewer;
- every `trap` is an id from the catalog; `target_concept` is a topic id; every entry of `excluded_skills` is a skill id;
- the brief's `target_concept` and `difficulty` match what the open request recorded, unless the model declared an override with a reason (section 3.4);
- `excluded_skills` is the profile's list entire — the model may add to it, never remove;
- `hint` and `solution` are present and non-empty;
- the self-check is present with all three of its fields.

There is **no check that the child's pseudonym stayed out of the task** (О-36). The generation package leaves it out (4.1), but the model knows it anyway — it is in the profile tool's result and the model greets the child by it — so keeping it out of the task stays a rule of the instructions, as in the prototype. A text search would cost an attempt every time a pseudonym is an ordinary word, which for a child choosing their own is most of the time.

## 5.3 The explanations behind the wrong options

Four conditions, all deterministic, no judgement of meaning (О-46, R10). For each of the four `distractors[*].text`:

1. the four texts are pairwise different;
2. a text is neither equal to nor a prefix of the `solution` or the `hint`;
3. a text is not the catalog's description of its own trap, repeated verbatim;
4. a text is at least **6 words** long, or **12 characters** for a writing system counted in characters (5.5).

Anything cleverer — "is this explanation meaningful?" — needs a judgement the service cannot make without an LLM call it is not allowed to make, and a wrong guess costs the child another attempt (О-46). The refusal names the option and the condition.

## 5.4 The drawing: format and match

Applies only when `drawing` is present. **Every limit here is conservative until T58 calibrates it on real widgets** (О-11а); they are configuration, not constants in the code.

**Format** — `drawing_format`:

| Limit | Value until T58 | Why |
|---|---|---|
| Width | **30 screen cells** | T03 rendered 30 cells legibly on a phone card 353 px wide, with no horizontal scrolling. A cell is one East Asian *narrow* character; a wide one counts as two |
| Height | **12 lines** | Fits a card without pushing the buttons off a phone screen |
| Allowed characters | ASCII U+0020–U+007E, box drawing U+2500–U+257F, block elements U+2580–U+259F, geometric shapes U+25A0–U+25FF, arrows U+2190–U+21FF, plus `\n` | Everything a frame needs and nothing that renders differently from font to font |
| Forbidden outright | control characters other than `\n`, tab, bidi controls (U+200E, U+200F, U+202A–U+202E, U+2066–U+2069), zero-width characters (U+200B–U+200D, U+FEFF), non-breaking and typographic spaces (U+00A0, U+2000–U+200A, U+3000), combining marks | They make the drawing look different to the checker and to the child, which is the whole attack surface of a text picture |
| Runs of spaces | at most 20 in a row | A drawing held together by a long run of spaces falls apart in any proportional fallback font |
| Trailing whitespace | none | It survives no round trip and breaks alignment |

**Match** — `drawing_mismatch`: every label the wording names must be an object in `drawing_structure`, and every label in `drawing_structure` must appear in the `drawing` text. A label is a token the wording marks as one — a single Latin capital, a digit or a short quoted name. What the check cannot do is decide whether the picture *means* what the wording says; that is the self-check's job, and О-37 deliberately gave it no code of its own.

## 5.5 Readability, across writing systems

Measured on the `question` only. Two checks, and which of them applies depends on the language, not on the child.

**The longest sentence**, in every language. Sentences are split on a sentence-ending mark followed by optional closing quotes or brackets and then whitespace or the end of the text; for scripts whose punctuation is not followed by a space, the mark alone ends the sentence. The marks: `.` `!` `?` `…` `。` `！` `？` `؟` `۔` `।` `॥` `።` `፨`.

The unit depends on the script of the text: **words** where words are separated by spaces, **characters** for the scripts written without them — Han, Hiragana, Katakana, Thai, Lao, Khmer, Myanmar, Tibetan. The script is decided by which of them most of the letters belong to.

| Level | Longest sentence, words | Longest sentence, characters |
|---|---|---|
| `1-2` | 20 | 40 |
| `3-4` | 25 | 50 |
| `5-6` | 30 | 60 |

The word limits are the prototype's, measured on 450 reference tasks (its D38), and extended to `5-6` in 1.1. The character limits are set at twice the word limit by analogy, because nobody has measured them: if an acceptance run in Chinese or Japanese (T62, T63) shows them biting, they move.

**Flesch–Kincaid**, English only: the grade index of the `question` must be at most the child's grade + 3. It applies when the task's language tag has the primary subtag `en`, and to nothing else — the formula counts syllables in English (the prototype's D38 and D42). The margin of +3 is measured: at +1 only 48 % of the grade 1–2 reference tasks passed for a first-grader, and the index is noisy on texts this short.

## 5.6 Near-duplicates

The promise that "the tasks do not run out" (PRODUCT 1) is checked from the other side: a new task must not be a near-repeat of one this child has already seen, or of a reference task the model was shown.

**The measure** is the one the prototype used, in a form that works without a database: the `question` is lowercased, every non-alphanumeric character becomes a space, each word is padded with two leading spaces and one trailing space, and the set of its three-character substrings is taken — the normalisation of Postgres's `pg_trgm`. Similarity is the Jaccard index of two such sets, and **0.6 or above is a near-duplicate**, as in the prototype.

For the scripts written without spaces (5.5) the same construction on **character bigrams** replaces it, because word trigrams of a text with no word boundaries measure nothing. Bigram sets are smaller and overlap more, so the threshold there is **0.7**. Both numbers are calibrated in T32 against the reference corpus, which is 450 tasks of known pairwise similarity.

**What it is compared against:**

- **The reference tasks** of the same level — their texts are in the binary, so the comparison is exact.
- **The child's past tasks**, through the fingerprints in the profile (04-profile), which are sketches rather than texts: the profile stores no task text, by design (О-40).

**The fingerprint** is therefore a MinHash sketch of exactly the trigram set above: **64 hash values, one byte each**, so the estimated Jaccard index is the share of the 64 positions that agree. One byte per position adds a collision bias of about (1 − J)/256 — under half a percentage point at the threshold, which a 0.6 cut-off does not notice — and the sampling error of 64 positions is about 0.06, which is why the threshold is not placed near a cliff. Sixty-four bytes per task is 88 characters of base64, and two hundred of them are about 19 KB of the profile.

## 5.7 What the solver has to show

The model submits a Starlark program that brute-forces its own task and prints the list of correct options. The contract, the sandbox and its limits are section 6 (T13). Two codes come out of it:

- `solver_error` — the program did not run to a usable answer: it crashed, exceeded its step or time limit, or printed something that is not a list of option letters. The model is told which of those happened and what the limit was, in one short line; the interpreter's own message is not passed through.
- `solver_disagrees` — the program ran and disagreed: it found no correct option, or more than one, or exactly one that is not `correct_answer`. **The same code covers the self-check** disagreeing: `final_answer` different from `correct_answer`, or `UNSOLVABLE`. The model is told which of the three verdicts disagreed, without the letters being quoted (5.1).

Two independent witnesses to the same answer is the strongest guarantee the service has, and it is the reason a topic that cannot be brute-forced does not enter the catalog (1.2).

## 5.8 The self-check's verdict

`self_check_blocking` — the self-check contains an issue with `severity: blocking`. The issue's own `comment` goes back to the model: it wrote it, and handing it back is what makes a self-check worth asking for.

A `minor` issue does not reject the task. The prototype kept minor issues for manual review; v1 stores no task text (О-38, О-40), so a minor issue is counted in the aggregate log by its `type` and its text is discarded.

## 5.9 The refusal codes, in one table

Every code the prototype had, plus what v1 added. Nothing was dropped.

| Code | From | Costs an attempt |
|---|---|---|
| `bad_structure` | the prototype | yes |
| `distractor_explanations` | new — О-46, R10 | yes |
| `drawing_format` | new — О-37 | yes |
| `drawing_mismatch` | new — О-37 | yes |
| `readability` | the prototype | yes |
| `solver_error` | the prototype | yes |
| `solver_disagrees` | the prototype, now also covering the self-check's answer | yes |
| `self_check_blocking` | the prototype | yes |
| `near_duplicate` | the prototype | yes |
| `stale_request` | new — the flow, not the task (03-flows) | no |
| `attempts_exhausted` | new — the flow (03-flows) | closes the request |
| `limit_reached` | new — the flow (03-flows) | no |

One prototype behaviour is deliberately **not** carried over: a separate path for "the model got its own answer wrong". It was already folded into `solver_disagrees` there, and it stays folded here.

And one code that was considered and does not exist: anything about the pseudonym reaching the task text (О-36, 5.2).

## 5.10 Attempts

Three attempts per request (PRODUCT 4.3). After the third refusal the request closes, the child is handed nothing, the model may ask for a new task, and the daily limit is untouched — a refused attempt is not a generation (О-35, 03-flows).

Three is the prototype's number and its reasoning holds: after two pointed corrections a model usually starts cycling through the same broken variants, and a refusal after three is itself a signal — which topics and which traps the models of the world stumble on is exactly what the aggregate log is for (О-16).

---

## Remarks on PRODUCT and RUN

Collected while writing this part; none of them changes a product decision.

1. **Nine reference tasks per topic at `5-6`, not twenty-five.** It is a deliberate asymmetry with grades 1–4 and the reason is the review cost of hand-written content. If T62 shows the model doing noticeably worse at difficulties 1 and 5 at that level, a fourth batch adds the missing anchors. **For:** T37–T39, T62.
2. **The seven new topics are `5-6` only.** `number.divisibility` and `logic.sets` would work at `3-4` too, but that would mean another 18 reference tasks and another review. **For:** a later edition.
3. **`geometry.grid` is the one topic that can fail its entry exam.** О-30 planned for that; T37 is where it is decided, and the replacement is chosen from the same list with no new decision round.
4. **The prototype's reference tasks have no `hint`.** The format the model must produce always does, and the new `5-6` tasks will. Backfilling hints into the 450 ported examples is optional work nobody is scheduled to do. **For:** T23, T36.
5. **The `excluded_skills` cap in 04-profile reads "15 (the catalog's size)".** The catalog is 25 now, so the cap is the catalog's size, not the literal 15. **For:** T26.
6. **The rank boundaries are defined here** because they are rating arithmetic, and T57 only draws them. If SPEC section 8 turns out to be a better home, move them there rather than duplicating them. **For:** T14, T57.

Added while writing sections 4 and 5:

7. **The profile's fingerprint is now decided, and it is bigger than 04-profile assumed.** A MinHash sketch of 64 one-byte values (5.6) is 64 bytes rather than the 32 that document's size table guessed, so its fingerprint line reads about 19 KB instead of 11 and a typical file about 38 KB instead of 30 — still well inside the 64 KB target, and its note 2 anticipated exactly this. **For:** T26, and two numbers to correct in [04-profile](docs/architecture/04-profile.md).
8. **The character limits for scripts without spaces are set by analogy, not measured.** Twice the word limit is a guess that nobody has tested on a real Chinese or Japanese task. The acceptance runs are where it will show. **For:** T32, T62, T63.
9. **The near-duplicate thresholds are two numbers, not one.** 0.6 for word trigrams is the prototype's, measured; 0.7 for character bigrams is reasoned from the smaller sets. T32 has 450 reference tasks to calibrate both against before anything ships. **For:** T32.
10. **Every drawing limit is provisional until T58** and lives in the configuration, not in the code. If the calibration moves the width, the drawing frames of T36b are re-checked against the new number (R09). **For:** T34, T36b, T58.
11. **`design_thought_process` is written and never read.** It exists to make the model state its plan before committing to it, and the service throws it away — it is not stored, not logged and not shown. If T36 finds the package tight, this is one field whose cost is entirely in the model's output, not ours. **For:** T36.
