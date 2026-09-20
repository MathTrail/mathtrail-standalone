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
| 4 | The task: format and content | T12 |
| 5 | The checks a submitted task passes | T12 |
| 6 | The Starlark solver | T13 |
| 7 | The MCP tools | T14 |
| 8 | Widgets, screens and languages | T14 |
| 9 | Sign-in and tokens | T15 |
| 10 | Limits | T15 |
| 11 | Configuration | T15 |
| 12 | Logging and metrics | T15 |

Sections 4–12 are not written yet. Section 9 will be assembled from [docs/architecture/02-auth.md](docs/architecture/02-auth.md), and sections 7–8 from [03-flows.md](docs/architecture/03-flows.md), [04-profile.md](docs/architecture/04-profile.md) and [05-storage.md](docs/architecture/05-storage.md).

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

## Remarks on PRODUCT and RUN

Collected while writing this part; none of them changes a product decision.

1. **Nine reference tasks per topic at `5-6`, not twenty-five.** It is a deliberate asymmetry with grades 1–4 and the reason is the review cost of hand-written content. If T62 shows the model doing noticeably worse at difficulties 1 and 5 at that level, a fourth batch adds the missing anchors. **For:** T37–T39, T62.
2. **The seven new topics are `5-6` only.** `number.divisibility` and `logic.sets` would work at `3-4` too, but that would mean another 18 reference tasks and another review. **For:** a later edition.
3. **`geometry.grid` is the one topic that can fail its entry exam.** О-30 planned for that; T37 is where it is decided, and the replacement is chosen from the same list with no new decision round.
4. **The prototype's reference tasks have no `hint`.** The format the model must produce always does, and the new `5-6` tasks will. Backfilling hints into the 450 ported examples is optional work nobody is scheduled to do. **For:** T23, T36.
5. **The `excluded_skills` cap in 04-profile reads "15 (the catalog's size)".** The catalog is 25 now, so the cap is the catalog's size, not the literal 15. **For:** T26.
6. **The rank boundaries are defined here** because they are rating arithmetic, and T57 only draws them. If SPEC section 8 turns out to be a better home, move them there rather than duplicating them. **For:** T14, T57.
