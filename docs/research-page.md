# The "Research" page

The site's page of the numbers behind the product (О-57): what the paper about MathTrail found, how the student model stands against its goals, and where the reference tasks came from. This document decides how the page is built and where every number on it comes from (T72.12, R204). T72.13 builds the page from it, and T72.15 brings the live numbers. T73.2–T73.4 move the making of its data into the research pipeline on merge to `main`, and keep its shape: since T73.2 a release's data is made by `research.yml` (R234), and since T73.4 the paper's PDF beside it and a pull request's site too (R248).

The reference is the author's draft, `research/draft-ui/MathTrail - Исследование.html`: a self-unpacking page from a design tool, in Russian, the second of that name, of 2026-10-07 (R250). Its markup is line 382 of the file, a JSON string (`sed -n 382p F | jq -r .`). The first draft, of 2026-10-03, is in the history up to `29205e2`. The draft is ported, not copied: the site rebuilds it from its own components, with every number from data and every word from `research.yaml`, but since R250 in the draft's own look, its sizes, rules and colours, rather than the rest of the site's.

## 1. What the author decided (2026-10-05)

| # | Decision |
|---|---|
| 1 | The paper's title is on the page from its first release, though the paper is under double-blind review. The AIED 2026 call accepted preprints that follow Springer's policies. The 2027 call, the one the paper goes to, was not out on 2026-10-02 (`research/paper-a/submission.md`), so this is read again when it is: a named preprint, once indexed, cannot be withdrawn. |
| 2 | The PDF is the named build. It was to wait until it held no placeholder; since R260 (2026-10-07) the page offers it with every release, whatever placeholder it still prints. Authorship (K05) was filled on 2026-10-07, and the PDF prints `[TBD: K11: the artifact's DOI]` until the artifact has its DOI (K11, T73.6). |
| 3 | Every language links the English PDF. No Russian PDF is built (the author, 2026-10-07; T73.3, R239). |
| 4 | Search engines may index the PDF from the start. |
| 5 | The model's goals are bounded against the current service, as the bench's criterion bounds them: the bar is for the next rule. |
| 6 | Where a goal's bound is computed from the service's own value of the same measure, the service's row says "baseline" rather than a mark. |
| 7 | CI computes the bench's numbers from the commit it builds, and they are never committed. The other inputs are small committed files: the monthly live snapshot (§11) and the curated books. The PDF and its facts were to be committed too, until T73.4 had the release's run build them (R248). |
| 8 | The live block comes later, in a task of its own (T72.15). Until then the page shows the block's frame and says the data is still to come. T72.15.2 brought it (§11). |
| 9 | The live block's styles go into the shared stylesheet, which every page loads, rather than one of the page's own. Should the heaviest page pass its budget, the author decides again. The Russian home page passed it in release v0.3.13, and the author raised the budget by half, to 450 KiB (2026-10-07, R240). |
| 10 | The live snapshot comes with nobody at a keyboard (2026-10-05). A scheduled query keeps it, and once a month a GitHub App of the repository's own brings it to `main` by a pull request that merges itself (2026-10-06, §11, R226). |
| 11 | The page takes the second draft whole, its look as well as its parts, and shows the service's numbers alone: the rule before R187 leaves the page and its data (2026-10-07, R250). |
| 12 | The words of a paragraph both pages have stay the site's, which were written for accuracy and in whole sentences; the draft's new labels are taken as they are. Under the student model stand the draft's one line of the run and one sentence on the paper's commit, and the line on the bootstrap's resamples and the one that no child took part go. Until a month is counted whole, the live block shows the draft's three columns empty, under "Coming", rather than its simulated numbers (2026-10-07, R250). |

## 2. Address and frame

- **Address.** `/<locale>/research/`, words in `site/content/<locale>/research.yaml` (SPEC 8.12.2). Anchors: `#student-model`, `#live`, `#sources`, `#paper`.
- **Menu and footer.** "Research" stands before "About" in the menu and first in the footer (SPEC 8.12.2, R202). R202 measured the fold with five entries. With the sixth, the menu stays in a row down to 1039 px in Russian and 976 px in English in Chromium, and down to 1024 px and 962 px in WebKit, so it folds below 1100 px and «Исследование» keeps its place (R209). `just site-fold` measures the width again whenever an entry or a language is added. Since R251 "Research" shares a capsule of the menu with "Service", the technical pages, the page being read dark in it, and the menu with its seven entries folds below 1180 px: it stays in a row down to 1121 px in Russian and 1057 px in English in Chromium, 1107 px and 1043 px in WebKit. Since R258 the coach's entry carries a tag, "Beta", and the menu folds below 1230 px: it stays in a row down to 1166 px in Russian and 1102 px in English in Chromium, 1152 px and 1089 px in WebKit.
- **Lists.** The addresses join `web/tools/sitecheck/published.ts`. The README links `/en/research/`.
- **No script, nothing from elsewhere.** The page reads in full without a script and runs none (R160). It loads nothing from another origin, so the draft's unpkg React and its fonts' `preconnect` go.
- **Fonts.** Onest only (R157): the draft's Source Serif 4 alone weighed 0.9 MB.
- **Weight.** Since R250 the page's own rules weigh some 24 KB in the shared stylesheet, against some 5 KB before. With everything it loads, the page weighs 178,727 bytes in English and 183,457 in Russian while the live numbers are still to come, under the 450 KiB a page may weigh (R240). Every page loads the shared stylesheet, so each grows by the same 19 KB: the Russian home page, the heaviest page of the site, came to 346,233 bytes on 2026-10-07. In release v0.3.13 it had come to 307,632 bytes, past the 300 KiB of the time, and the author raised the budget by half (R240). The PDF is a link, not a resource the page fetches, so it does not count.
- **Languages.** English and Russian now, the rest with T66b. Drawings are pinned left to right (`dir="ltr"`), as the topics' map is; tables and words follow the page's direction.

## 3. The page, section by section

| Section | In the draft | On the site |
|---|---|---|
| Hero | "The model writes, the service checks." in two lines, the second bold; a lead; "The whole paper · PDF, 16 pages" and "Code and data"; beside them a drawing of a task's way: the chat's model writes the fence's task and its solver, MathTrail runs ten checks, the solver twice, and sends a task that fails one back along a dashed line, and the child solves one that passes, its answer sealed; under a rule, 603 reference tasks · 10 checks · 17 topics · 20 traps · grades 1–6 · open code, MIT | The same, under a badge in the menu capsule's greys that says the page is written for technical readers (R258). The heading is the paper's title up to its colon, in the page's language, and a test holds it to the paper. The PDF button waits for a clean PDF (§10), and its page count comes from the data. The drawing's fence, its numbers and its five options are chosen by the text and marked so (§9); its options are thesis 02's, and a test holds the fence's posts to the right one. The number of checks and of their cells come from the data (§5). "Code and data" leads to the repository. |
| Four theses, 01–04 | The chat's model writes; the solver runs twice; the answer is sealed; a little harder, but within reach. Each a row between rules: its place, its words, its drawing | The same, with each number from the product (§5) and the worked task's values marked as chosen by the text (§9). The draft's counts in words, "five tasks", "by two", are numbers from the data. The code sharpens three claims: the service keeps nothing of the child between requests; the solver runs a second time only when its first run picked exactly one option; the rule asks for the task whose chance is nearest the middle of the corridor. **Thesis 03 stands as the draft has it.** Since R152 the card is drawn by `next_task`, which carries no task, and reads its task through `read_task`, which carries no answer, so before the child answers the answer is in nothing the service hands the card or the chat's model. The page adds what the seal does not reach, which is the chat itself: the model that wrote the task knows its answer, and the host's record of the tools called shows it to an adult who opens it (О-27). |
| Student model, `#student-model` | One line of the run; a key; the measures in groups, each a row of its name, unit and mark, a bar of the service's number with its interval, the goal's dashed line and the ceiling's dotted one, and the number; a legend of three marks | The board of §6, with one sentence on the paper's commit under the line of the run. |
| Live data, `#live` | Promised → came true, came true less promised by answer count, the share right on the rule's tasks, side by side — all marked as simulated | The same three measures of the latest month counted whole, side by side (§8): a point for each range of chance on a square, a bar from nothing for each range of the child's answers, and the share right, large, as a point over the corridor. Each drawing is hidden from a screen reader, which reads a table of its numbers the eye does not see. Before a month is counted whole, and when its children were too few, the three columns stand empty under a badge that says so. Every state names the rule a range is shown by (§11). |
| Where the reference tasks came from, `#sources` | Public-domain books by author, a row each, with the year each died; that they belong to everyone; 603 tasks as a bar of 200 · 250 · 153 | The same. The books and their authors are curated data in `site/data.json` (`research.sources`), as "Why" keeps its works. A book is named by its whole title, and its author with the year they died: a first edition's year is not on record for every book, Perelman's ran through many editions, and in most of the world copyright runs out a set number of years after the author's death. The counts by level come from the content. |
| The whole paper, `#paper` | What the paper holds, and its page count, between two rules | The same, with the PDF once there is a clean PDF (§10); the page count stands on the first screen's button. |

**What leaves the draft:**
- the simulated live charts and their counts, which were not rounded as public counts must be: the columns stand empty until a month is counted whole;
- the live block's range of answers 1–5, the trial series, whose answers are weighed against no chance;
- the draft's wording of a paragraph the site's page already had, which stays as the site wrote it (decision 12), and of a book's title, which stays whole;
- a digit in the words: "1-й запуск" is "the first run", "смен на 100 ответов" changes "a hundred answers" (§9);
- the draft's own menu and footer, and its fonts' `preconnect`.

**What the site adds:** a sentence on the paper's commit under the line of the run (decision 12); the mark "on the edge" in the legend whenever a row holds it, since the legend names the marks the rows hold; the oracle's "best possible" on the lag's row only, as the draft has it; and a row of the board laid out as two lines on a narrow screen, its name and number over its bar, rather than a board 680 px wide that scrolls sideways.

The words are written anew in English from the Russian and are the source, as for every page (R159).

## 4. The data file

**`research.json`** holds every number of the page. The page's code reads it through one reader, `web/src/site/research.ts`, which checks it with zod as `data.ts` checks the site's data.

The file is not committed. It is made at build time, in `site/research/research.json`, which `.gitignore` lists. For a release and for a pull request it is made by the research's run of the same commit, which hands it to the site's build with the clean PDF beside it (§10); when the site is published by hand, the site's build job makes it (§7). Locally, `just research-data` makes it. One kind of input sits committed beside it in `site/research/`, the live snapshot (§11). The clean PDF and its facts lie there once a run has made them, and git ignores them (§10).

```jsonc
{
  "schema": 1,
  "built_from": { "commit": "<40 hex>", "date": "<the commit's date, UTC>" },
  "bench": {
    "producer": "tools/learners page",
    "inputs": "<the key of the bench's build (§7)>",
    "seed": 20261001, "experiment": "E-A3", "children": 1000, "answers": 200,
    "interval": 0.95, "resamples": 2000,
    "error_after": 200,
    "screen_windows": { "early": { "first": 6, "last": 20 }, "late": { "first": 150, "last": 200 } },
    "rules": [ { "id": "shrinking/both", "role": "service" },
               { "id": "oracle/both", "role": "ceiling" } ],
    "rows": [ {
      "id": "lag", "kind": "goal", "criterion": "step",
      "generator": "G2-half", "metric": "r6_lag", "unit": "logit", "read_as": "size", "better": "lower",
      "bound": { "value": 0.2420, "own": true },  // the file holds every number in full; rounded here
      "values": {
        "service": { "value": 0.5377, "low": 0.5202, "high": 0.5556, "mark": "baseline" },
        "ceiling": { "of": "oracle", "value": 0, "low": 0, "high": 0 } } } ]
  },
  "product": {
    "topics": 17, "traps": 20, "checks": 9, "grades": { "first": 1, "last": 6 },
    "reference_tasks": { "total": 603, "by_level": { "1-2": 200, "3-4": 250, "5-6": 153 } },
    "options": 5, "relabel": 2, "relabelled": ["C", "D", "E", "A", "B"], "guess": 0.2,
    "corridor": { "low": 0.7, "high": 0.85, "middle": 0.775 }, "trial_answers": 5
  },
  "paper": {
    "commit": "52ce86908135",
    "files": [ { "lang": "en", "path": "/assets/paper-a.en.pdf", "pages": 15, "bytes": 350493, "sha256": "…" } ]
  },
  "live": {
    "state": "ready", "rule": { "learners": 10, "answers": 30, "rounded_to": 5 }, "month": "2026-11",
    "total": { "learners": 45, "answers": 1235, "promised_mean": 0.754, "correct_share": 0.734 },
    "chances": [ { "from": 0.7, "to": 0.77, "learners": 40, "answers": 610, "promised_mean": 0.741, "correct_share": 0.725 } ],
    "kept_up": [ { "first": 6, "last": 20, "learners": 45, "answers": 550, "came_true_less_promised": -0.016, "standard_error": 0.009 } ]
  }
}
```

**Rows.** The service's number of a goal carries its mark: `reached`, `on_the_edge`, `not_reached`, or `baseline` on the service's number against a bound of its own (§6). A row says whose its ceiling is, `oracle` or `perfect`. What a row does not have is `null` rather than left out, so every row has the same keys: a context row has no `criterion`, no `bound` and no marks, and a screen row has no ceiling. `error_after` and `screen_windows` give the answers the labels name, "after 200 answers" and "answers 150–200", since the words hold no digit (§9). `relabelled` is where each letter of the first run stands in the second, so the page draws thesis 02 from the solver's own relabelling.

**Live numbers.** `live` has the same keys in every state, and what a state has no value for is `null` or empty: `coming` has no month, `too_few` has a month with no total and no range, and `ready` has the month's total and the ranges its views show, from the lowest. A range is named by the numbers it spans, the chances `from` and `to` or the answers `first` and `last`, with no `last` for the range of every answer past the others: the views name a range in words and digits, "under 0.50" or "6-20", which the page's words cannot hold. `rule` is the rule the public views show a cell by, so the words name it through the data.

**Rules for the file:**

- **Byte for byte the same** for the same inputs, on amd64, the architecture CI runs on, built at `GOAMD64` v1, Go's default, and run on a processor with FMA, as every runner is. At `GOAMD64` v3 Go fuses a multiplication and an addition into one step that rounds once, as it does on arm64, and `math.Exp`, which the rating's logistic calls, takes another path on a processor without FMA. In each case the last digits part.
  - Numbers are written in full, in the shortest form that reads back to the same value, as Go's `encoding/json` writes them; it is the form JavaScript writes a number in, so the page can trace every `<data value>` to the file. They are rounded only on the page, so the reader can recompute every mark exactly. The file can hold no NaN: `encoding/json` refuses one.
  - There is no wall clock: the date is the commit's.
  - T73.2's check that two runs give the same bytes rests on this, on amd64.
- **Values as the criterion reads them.** A goal on a size, such as the lag, holds the size of the value and of its interval, as the bench's criterion reads it (`sizeOf`: an interval across zero reads from zero). `read_as` names the reading, and `better` applies to it.
- **Two producers.** The bench's command `page` writes `bench` and `product` (§6, §5) for one build of the bench, and its command `page-file` adds what the commit being built adds: `built_from`, `paper` from the facts of the PDF the paper's run wrote (§10), and `live` from the committed snapshot of the live numbers, or still to come with none (§11). `just research-data` runs both (§7). Since T73.2 a release's file, and since T73.4 a pull request's, is written by `research.yml` with the same recipe, so the producer stays the bench's `page` and the file is the same byte for byte.
- **The reader refuses what does not add up:**
  - a mark that does not follow from the value, its interval, the bound and the direction, and "baseline" anywhere but on the service's number against a bound of its own, or missing there;
  - product counts that disagree with the catalogs and reference tasks the site already reads, the tasks by level included;
  - a constant that does not hold as the page states it: as many options as letters, the second run's letters the first's moved on by the shift, a shift that moves every letter, from 1 to one less than the options, the guess one option of all, and `0 < low < middle < high < 1`;
  - rules without exactly the two roles, the service and the ceiling, a row named twice, an interval whose ends are the wrong way round, and a number or a goal under zero, where a row's drawing starts;
  - two files of the paper at one address, or files with no English one among them;
  - live numbers whose rule is weaker than the public views', which a test holds to their SQL; whose state their numbers do not give; with a cell under their rule, not rounded to it, or counting more than its month; with ranges out of order or overlapping; or with a promise outside its range of chance (§11).
- **The build refuses a PDF** that is not beside the data, or whose size or hash differs from its facts, as it copies the paper's files into the site (`paperFiles` in `web/scripts/prerender-site.ts`). The build reads the file from `--research`, `site/research/research.json` by default; with none there, or one a run cut short, it names `just research-data`, which makes it.
- **One fixture of the whole file**, `site/research/testdata/research.json`, holds both sides to one shape.
  - The bench's test makes it from a run of ten children a cell, the fixture of the PDF's facts in `tools/learners/testdata/paper.json` and a made-up snapshot of a month shown, `tools/learners/testdata/live.json`, with the commit and the key fixed by the test. The site's tests make the live block's other states from it.
  - The test makes the file again byte for byte and holds every mark in it to its numbers; a meant change, of the model or of the content, rewrites it with `-update`, and the diff is read in review.
  - The site's reader of the file, `web/src/site/research.ts`, parses the same file in its own test, and the site's build tests build from it.
  - A change of shape in any block fails one side or the other.
  - The bytes hold on amd64 with FMA alone, so the fixture's test runs there and skips elsewhere, as the bench's carried-over check skips off amd64. It refuses `-update` anywhere else.

## 5. Where each number comes from

| Number on the page | Source |
|---|---|
| Reference tasks, in all and by level; topics; traps | The content: `content/examples/*.json`, `content/catalogs/topics.json`, `traps.json`. The site already reads them (`siteData()`); `product` repeats them so the reader can hold the two to each other. |
| Checks | `internal/domain/checks`, through the list it exports, `checks.Codes()`, which a test holds to every code the package declares. |
| Grades 1–6 | The first and the last school year the service is for, `profile.MinGrade` and `profile.MaxGrade`. |
| Options per task, the shift between the solver's two runs, the guess floor, the corridor, the trial series | The product's constants: `solver.Count`; the shift as the solver's own relabelling gives it, where the first letter of the first run stands in the second (`solver.Relabelled`), so the solver keeps its constant to itself; `rating.Guess`; the corridor's bounds, `rating.CorridorLow` and `rating.CorridorHigh`, which the bench reads too rather than keeping copies, and its middle; `rating.TrialAnswers`. |
| The worked task of the first screen and of thesis 02 | The fence's length and gap and the five option values are chosen by the text and marked as such (§9); the fence's posts are its right option. The second run's letters are computed with the solver's own relabelling. |
| The student model's board | The bench's `page` run (§6). |
| The paper's title | The paper: `research/paper-a/main.tex` for English, the first heading of `research/paper-a/draft.ru.md` for Russian. The words carry it, and a test holds the English words to `\title` and the Russian to that heading. |
| The PDF's pages, size and hash | The PDF itself, read when the paper's run copies it beside the data, or `just site-paper` on a laptop (§10); the site's build holds the file to its size and hash. |
| Books, their authors and the years the authors died | Curated data, `site/data.json` → `research.sources`: each author's id, the year they died and their books' ids, in the page's order. The names and the titles are words of `research.yaml`, under those ids. |
| Live numbers | The public views of the counts kept for years, through the snapshot a scheduled query keeps of them every day in `impact_site.live`, committed once a month as `site/research/live.json` (§11). |

## 6. The student model's board

**Where the numbers come from.** The bench gets a new command, `page`, which runs seven cells, those of the board and nothing else. It is a command of its own, as the guard is, rather than a set of rules: the bench's own run puts every rule on all nineteen generators and reads comparisons these cells do not hold.

| Rule | Generators |
|---|---|
| The service, `shrinking/both` | G0, G0-topics1, G2-half, G3 |
| The ceiling, `oracle/both` | G0, G2-half, G3 |

Until R250 it ran four more, the rule before R187 on the service's four generators, for a column the page no longer has.

A cell's children and its intervals depend on nothing but its rule and its generator: the children are drawn from seeds named after the generator, and every interval from a stream named after its cell and metric. So these cells give exactly the numbers of the whole run, to the last digit.

**Which goals.** A new `pageCriterion` reads both criteria's goals, those of the step and those of mastery, against the service as the baseline. Since R187 the baseline of mastery's criterion, the floor under the step, runs under the service's own mastery, so its cells are the service's. Rows:

| Row | Generator, metric | Bound | Service's mark |
|---|---|---|---|
| The lag behind a child who learns, `lag` | G2-half, `r6_lag` | 45 % of the service's lag | baseline |
| Tasks in the corridor, a child who learns, `corridor_learning` | G2-half, `r3_inside` | the service's, plus 27 % of the way to the ceiling | baseline |
| Children not caught up after a jump, `jump_unsettled` | G3, `r6_jump_unsettled` | at most 25 % | a mark |
| Masteries declared falsely, `false_mastery_static` and `false_mastery_wide` | G0 and G0-topics1, `r4_false` | at most 20 % | a mark |
| Answers until a mastery is declared, `late_mastery_static` and `late_mastery_wide` | G0 and G0-topics1, `r5_late_answers` | 1.5 times the service's | baseline |
| What the child is shown late in a run: the card's move and the rank's changes in answers 150–200, `card_move_*` and `rank_changes_*` | G0 and G2-half, `r8_move_p95_150_200`, `r8_rank_150_200` | the service's in answers 6–20 | a mark |
| *Context, no goal:* the error after 200 answers, a child who stays put, `error_static` | G0, `r1_rms_200` | — | — |
| *Context, no goal:* tasks in the corridor, a child who stays put, `corridor_static` | G0, `r3_inside` | — | — |

Each goal is read by the bench's own `readGoal`, so the page bounds a goal exactly as the criterion does. A test holds the rows to the goals of both criteria, all of them once and none besides.

The screen's goals for answers 6–20 are left out: their bound is the service's own value in the same window, so they could only ever say "baseline".

**Rows.** Each measure is a row: its name, its unit and the service's mark; a bar from nothing of the service's number, its interval a darker band over it, the goal a dashed line and the oracle's ceiling a dotted one, each labelled with its number; and the number with its interval. The bar runs to all for a share, and otherwise to one, or past one to the least of one, two and five times a power of ten that holds the interval, the goal and the ceiling (`barScale`): one for logits, ten for answers, a hundred for points, five for changes a hundred answers, as the draft draws them. A label alone stands after its line, or before it past the bar's middle. Two labels turn away from each other when each has three tenths of the bar to stand in; otherwise each takes its side as a label alone does, and the ceiling's drops to a second row under the bar wherever the two could meet (`linePlaces`). The ceiling is the oracle's for the step's measures. For mastery's measures it is perfection, no false mastery and no wait, since the oracle keeps the service's mastery and is no ceiling for it (`docs/learners.md`). The oracle shows no screen. A ceiling at nothing draws no line, and perfection draws none: the lag's row says instead, in its unit, that the best possible is the oracle's nothing, as the draft has it. The board's key names the ceiling only while a bar draws one.

**Marks.** A goal is read on the value and its 95 % interval, by the bench's own rule:
- **reached** when the whole interval is on the right side of the bound;
- **not reached** when the whole interval is on the wrong side;
- **on the edge** otherwise.

Some bounds are computed from the service's own value of the same measure, the same generator and metric: `own` in the data. There the service's mark is decided in advance: a share of its own value is never reached, and a multiple of it always is. So its row says **baseline**, and the page says plainly that such a goal is the bar for the next rule.
- The lag, the corridor of a child who learns and the answers until a mastery are such goals.
- The screen's late window is not, since its bound is the service's value in another window, a measure of its own.

The legend under the board explains the marks the rows hold, in the draft's words: "baseline" is «Планка» in Russian.

**Provenance.** Over the board, in one line: the commit the numbers were computed on and its date, the children per cell and the answers per child, and the seed. The board's key names the interval's 95 %, and the lead says that the children are simulated. One more line says that the paper reports the rule as it stood at the commit its numbers were computed on, `52ce869`. The PDF's version names another, the commit the PDF was built from (R239), so the line names the commit by what it is rather than as the paper's own (R248). This page follows the service as it runs, so the two differ.

## 7. The pipeline

- **On every build.** Before the site is built, `just research-data` runs, in the research's job "The page's data" for a release and for a pull request, and in the site's build job (`pages.yml`) for a site published by hand:
  - it builds the bench and names its key;
  - it takes the numbers kept under that key, or runs `tools/learners page` to compute them and keeps them under it, in `tools/learners/results/page/`, which git does not keep;
  - it runs `tools/learners page-file`, which holds the numbers to the key, so that numbers of another build are never the page's, writes `research.json` from them, from the commit being built and from the committed snapshot of the live numbers, and prints the goals table with a line on the live numbers, which the job writes to its summary.
- **The cache's key** is a hash of two things, and nothing else moves the numbers:
  - **The bench's program**, built so that the same sources make the same bytes wherever they are built (`go build -trimpath -buildvcs=false`). The program holds everything the cells depend on: every package it links, the product's included; the content it embeds; the versions of its modules; the language version each module's `go` line sets; the compiler; the architecture and its level, `GOAMD64`. The run's own parameters are the command's defaults, which `just research-data` does not change: the paper's seed and run, a thousand children a cell, two hundred answers each. So they are in the program too. The program is not stamped with its commit, which would change the key on every commit.
  - **Whether the processor has FMA**, which `math.Exp` decides as it runs, so that the same program can give other last digits on a processor without it.

  A hash of the sources, which T72.12 drew up, would need a list of every file the cells depend on, and the toolchain image's tag would change the key whenever any tool in the image moved. The program is the whole list, and is what runs.
- **On a pull request** the numbers are computed and shown in the summary but not published. The cache is restored and saved, so a pull request that leaves the model alone computes nothing. Go's own caches are restored on a pull request too.
- **On a run that publishes** the numbers are computed again from its commit, with no cache restored: nothing an earlier run left behind reaches the site, Go's caches included. That is a release, which `release.yml` starts after green CI on `main` (R195), and a build started by hand. The run only asks whether numbers are kept under its key, and keeps its own when none are, for the pull requests after it: a key once kept is never written again.
- **Cost.** The seven cells took 10 seconds on the twenty processors of the machine R250 was built on, about two processor-minutes; the eleven of T72.13.1 had taken 18 seconds there, against the whole run's 59 processor-minutes over its 190 cells. On the four processors of a runner that is well under a minute, plus the bench's build.
- **Failure.** A failed computation fails the build, and the release publishes nothing: the site stays as it was. A release only follows a CI run in which the bench's tests and its guard passed.
- **The guard holds the model, not the page's numbers.** T72.12 asked for the guard to hold the page's numbers to its bands.
  - It holds the service on every pull request (`ci.yml`), and the page's numbers come from the same code. A release only follows a green CI, so it never shows the numbers of a model the guard refused.
  - The page's own numbers are not held to the bands themselves: the bands are drawn for the guard's 300 children, and the page reads 1,000.
  - The page never shows the guard's numbers.
- **Handover (T73.2, R234; T73.4, R248).** A release's file is made by `research.yml`, called by `release.yml` with the release's tag. The experiments of paper A are held to their results, the job "Paper A" builds the paper, and the job "The page's data" computes the bench's numbers from the commit with no cache restored, as above. It takes the paper's facts in on every run, whatever placeholder the paper still prints (§10, R260), and hands `research.json` on with the PDF as the artifact `research-data`. `pages.yml`, called with `research: true`, downloads it into `site/research/`, holds its `built_from.commit` to the commit it builds, and makes no data of its own. A pull request's run makes the file the same way, with the bench's numbers restored, and its job "Build the site" builds the site from the artifact with the steps a release's build takes (`.github/actions/site`), publishing nothing. When the site is published by hand, `pages.yml` makes the file as above, with no PDF. The file's shape stays.

## 8. Drawing

Every chart is drawn at build time by the page's own Preact components, in SVG or as HTML bars. No library, no canvas, no script. The site already draws this way: the topics' map (`TopicMap.tsx`), the ages of the techniques (`.s-ages-bar`), tables in `TableFrame`.

| Chart | Drawn as | Its table |
|---|---|---|
| The student model | A bar in each row of the board (§6): the fill, the interval's band and the lines hidden from screen readers, which read the labels of the goal and the ceiling and the number beside the bar | Its rows |
| The reference tasks by level | A bar of three parts, each as wide as its count | Its labelled parts |
| The first screen's way of a task, the theses' diagrams | HTML, as in the draft | — |
| Live: promised → came true | A square whose two axes share their scale, from four tenths, or from the lowest tenth drawn, to one: the dashed diagonal where a promise comes true in SVG, the corridor's band, and a point per range of chance, the larger the more answers it holds, joined by a line within each run of ranges that meet, never across a range the month leaves out | Under it, hidden from the eye |
| Live: came true less promised, by the child's answers | A row for each range: its answers, an HTML bar from nothing, the dashed line in the middle, with a whisker of one standard error either way, and its number | Under it, hidden from the eye |
| Live: the share right on the rule's tasks | The share, large, and a point on a bar from nothing to one, over the corridor's band and a dashed line at its middle | Under it, hidden from the eye |

Each live drawing is hidden from screen readers, and its table, which a screen reader reads, is hidden from the eye (`s-hidden`): the draft sets the three side by side, with no room for a table beside each.

**Numbers** are formatted with `Intl.NumberFormat(locale)`, by the measure's unit (`ResearchNumbers.tsx`): shares as percentages with up to one decimal, logits with two, answers and changes a hundred answers with up to one, points whole, a chance with one to three decimals, an end of a range of chance with two, and a difference of chances, came true less promised, with its sign; years and the seed with no grouping, since they are names more than amounts. The live numbers' month is named by `<Month>` as the page's language names a month, in a `<time datetime>` of the month. An interval is its two ends in the words' pattern `{low}–{high}`: `formatRange` writes an interval whose ends round alike as one approximate number, "≈0", where the table means "from 0 to 0". Tables use tabular figures.

**Rejected:**

| Candidate, current version | Why not |
|---|---|
| Observable Plot 0.6.17 (ISC) | Outside a browser it needs a virtual DOM passed as its `document` option, JSDOM or linkedom. It makes its own tick labels, digits the guard of §9 cannot tell from typed ones. |
| Vega 6.4.0 with Vega-Lite 6.4.3 (BSD-3-Clause) | It renders SVG headless (`view.toSVG()`, renderer `none`), but measures text only with node-canvas installed. It formats numbers by d3-format locale definitions rather than `Intl`, so its separators would part from the rest of the site's. |
| Chart.js 4.5.1, uPlot 1.6.32 (MIT) | They draw on a canvas and cannot be read without a script. |

The versions and the behaviour are as their documentation states on 2026-10-05: the npm registry; Vega's "Usage" and "Locale API" pages; Plot's `docs/features/plots.md` and `docs/features/formats.md`. A library reconsidered later is checked again on the version it would be pinned at:
- its rendering without a browser, and the licence of whatever DOM it needs;
- whether its output is the same byte for byte;
- its hook for formatting numbers, and how many of the widget's 22 languages it covers;
- right-to-left text, and what it writes for screen readers;
- its weight per chart;
- `just licenses`.

## 9. No number typed by hand

The rule is the paper's (`research/tools/handtyped`): a number reaches the text only from data, and a check finds any other.

1. **Components** (`web/src/site/ResearchNumbers.tsx`).
   - `<Num>` writes a number from data as `<data value="0.5377">0,538</data>`: the value as the file holds it, the text as the page's language writes it.
   - `<Given>` marks a number the text chooses rather than reports, as `<span data-given>`: the worked task's fence and option values, an axis's 0 and 1, a thesis's place. The fence's sum is written in a `<span>` rather than `<code>`, which the scan reads as a commit. The authors' years are data, not given.
   - `<Counted>` writes a count with its noun. The noun's plural wordings, in the site's dictionary, hold no `{count}`: the count picks the form and is written beside it as `<Num>`. A sentence of the words takes the two as slots of its own, `{count}` and `{noun}`, from `countSlots`, and places them as its language orders them; no verb of it agrees with the count, which the data may move to any number.
   - An identifier from data that holds digits but is no number is written in its own element: a commit's hash in `<code>` by `<Commit>`, a date in `<time datetime>` by `<CommitDate>`, the live numbers' month in `<time datetime>` by `<Month>`. The scan passes such an element only when it stands for a value of the file: a code that begins the commit the numbers were computed on or the paper's commit, a time whose `datetime` is the commit's date or the live numbers' month.
2. **The words hold no digit.** A test fails on any digit in either `research.yaml` outside an allowed name. Counts in a sentence come through slots, as on "Techniques".
3. **The rendered page is scanned in every language** (`handTyped`, in `web/src/site/testing/`). It reads the page's `<title>`, its description and the tags a shared link shows, and in `<main>` every text node and the `title`, `aria-label` and `alt` attributes; the frame around `<main>` is the site's, the same on every page. The scan skips scripts and styles, `<data>`, what is marked as given, and the names with digits of a list kept beside the page's test, today XChaCha20-Poly1305 alone. It fails on any digit in any script (`\p{Nd}`), since T66b brings languages with digits of their own.
4. **Every `<data value>` is traced** to a value of `research.json` or of the curated data.
5. **A number changed in the data changes the page.** The test renders the page from its fixture and again with every number moved, and every `<data>` element must change.

## 10. The paper on the page

- **Title.** In English, "The Model Writes, the Service Checks: Olympiad-Style Maths Problems for Primary-School Children inside Chat Assistants". In Russian, the Russian draft's title. The page's heading is the title up to its colon, in two lines as the draft sets it (R250); a test holds the English to `research/paper-a/main.tex` and the Russian to the Russian draft's heading.
- **The PDF.** The research's job "Paper A" makes it on every release and every pull request (R239, R248):
  - it builds the named PDF with `just research paper-a`, which runs the paper's own checks first, beside the anonymous one;
  - `just _paper-to-site` refuses the PDF when the research's tree held changes no commit has, before the build or while it ran, since the PDF prints the commit it is built from, which must make it (R239). It copies the PDF to `site/research/paper-a.en.pdf` and writes to `site/research/paper.json` its pages, as the build's log gives them, its size, held to the file's, its hash and the commit the paper's numbers were computed on;
  - `just _paper-placeholders` names the lines of the paper's sources that still print `[Author]`, `[Affiliation]` or a `\TBD`. It reads `main.tex` and the sections and figures, but not the preamble, which defines `\TBD` and holds the anonymous build's own; a line that is a comment does not count, and a source it cannot read fails it. Today it names one line, the artifact's DOI of K11, since K05 was answered on 2026-10-07;
  - the job hands the PDF and its facts on to the page's data on every run, whatever placeholder the paper still prints (§7, R260). The run's summary names the lines that still print one.

  Neither file is committed. `just site-paper` makes both for a site built on a laptop, and names the placeholders the paper still prints. The site serves the PDF at `/assets/paper-a.en.pdf`: a folder `/research/` at the site's root would read to the site's build as a language. The site's build copies it there from beside the data and refuses it when its size or hash is not the one its facts give (§4). With no file, the page shows the title and "Code and data" but no PDF button, and its section on the paper says the PDF comes with the next release.
- **The paper's commit** on the page is the PDF's when the site ships one, and otherwise the one `research/evidence/product-stats.txt` names, the commit the paper's numbers were computed on. A release's run writes the PDF's facts from that same file, so the two part only for a PDF made on a laptop and left beside the data, and then the summary of `just research-data` says that the PDF is due to be built again; the build does not fail on it. It is the commit the paper describes, which its text prints, and not the one its version names: that is the commit the PDF is built from, a later one (R239). The page names it as the commit the paper's numbers were computed on (R248).
- **Every language** links the English PDF: no Russian PDF is built (R239).
- **Indexing.** The PDF is open to search engines.
- **Self-archiving.** Springer's self-archiving rules may limit which version can stay on the site once the paper is accepted; that is read when the venue decides.
- **Forks.** A fork removes the paper's PDF and title, which are the authors' work: `docs/self-hosting.md` says so, as it does for the coach's page.

## 11. Live numbers (T72.15)

The live block shows the same measures as `just report` (R153), counted for good and made public:
- promised → came true, by range of chance;
- came true less promised, by the child's answer count;
- the share right on the rule's tasks.

The answers are those to tasks the rule chose, after the trial series and without the hint. The ranges of chance are the report's six, and the answer counts are the report's ranges from 6 on; the draft's "1–5" is the trial series, which is left out. Every host is taken together, and the children of the load tool and MCP Inspector are left out.

- **What it needs** (built in T72.15.1, R221). The `answer_recorded` lines in the `activity` bucket carry `chance`, `tutor_mode`, `trial`, `hint_used` and `answers_bucket`, all the report's filter reads, and T67.2's nightly query now reads them too. It weighs the answers the report weighs and counts them into two tables by month and range, every instructions version together, since the version changes with nearly every release while the model does not (R212):
  - `chance_monthly`, by the report's six ranges of chance and over every range: the children, the answers, the right ones and the chances promised, in hundredths;
  - `kept_up_monthly`, by the range of the child's answers: the same, and the sums from which the mean and its error by child are computed with no child named.
- **Privacy, by R192 and R221:**
  - closed months only, each month on its own and never a running total, which would give a small month away by subtraction;
  - a range, or a month over every range, is shown only behind at least ten children and thirty answers;
  - no range is folded into `(others)`: a child's answers fall in many ranges, and `(others)` would count a child more than once. A hidden range is left out, as a topic of fewer than ten children is, and its answers can be had roughly from the month's whole less the ranges shown, to within the rounding to five, but never its children;
  - counts are rounded to five, and shares, means and errors are given to three places;
  - a hidden range shows neither a count nor a share.

  The public views are `chances`, `chances_total` and `kept_up` in `impact_public`. T72.12's "ten answers" yields to R192's ten children.
- **States.** The block is in one of three:
  - *coming*, with no snapshot, or one taken before a month was counted whole;
  - *too few*, when the latest month counted whole shows neither its total nor a range;
  - *ready*, when it shows its total, with whichever ranges its views show, or none.

  The block shows the latest month counted whole even when an earlier one showed more: each month stands on its own (R192), and an older month is never passed off as the latest. R187 rolled out on 2026-10-04, and the counts began in October, which is therefore never counted whole. So the first month is November, and its numbers come in December 2026 at the earliest.
- **Into the build: a monthly snapshot, decided at the start of T72.15 and taken with nobody at a keyboard since T72.15.3 (R226).**
  - The SQL, `infra/analytics/site/live.sql`, reads `closed_months`, `chances_total`, `chances` and `kept_up` and nothing else. It gives the whole snapshot as one value of JSON (`TO_JSON_STRING`) in one column, `live`, with the ranges from the lowest, since bq writes every number of a row as a string. The tests of the counts run it against the emulator and hold its keys to those the bench reads; a test that needs no emulator holds the scheduled query, the recipe and the workflow to the one table they share.
  - A scheduled query, run by the counter every day at 02:30 UTC, an hour after the night, replaces with that one row the table `impact_site.live`. The identity `mathtrail-live` may read that dataset and nothing else, and may run no query. Only a job of the GitHub environment `live-numbers`, which only `main` may run in, may borrow it.
  - `just site-live` reads the table with `bq head`, as the deployment the Terraform configuration names, with the credentials of whoever runs it, writes `site/research/live.json` and holds it to the rule with the bench. It refuses a table the scheduled query has not written for two days: the query has stopped, and the row may be of a month before the latest one counted whole.
  - The workflow `live.yml` does the same every day from the month's second, when the night has counted the month before, to its fifth, which take up a failed run or a late night, and when run by hand. Its first job reads the snapshot as `mathtrail-live` and checks that the exact counts refuse that identity. Its second holds the snapshot to the rule and compares it with the one `main` holds. Its third, when they differ, has a GitHub App of the repository's own open a pull request with it, which merges itself once the required checks pass, and the release after the merge publishes the site. Each job holds one credential at most.
  - The snapshot holds the month, the month's total and the ranges shown, named as the views name them, with every key in every state. It names no deployment: the page shows none, and the recipe prints the one it read.
  - The site's build needs no credentials. Each snapshot is a commit of its own, and the summary of the run that delivered it shows how it differs from the one before.
  - Reading BigQuery from `pages.yml` was rejected: its build runs npm, which should not hold `id-token: write`.
- **Into the page's data.** The bench's `page-file -live` reads the snapshot strictly and refuses one that shows what the rule hides or what no month could give:
  - a cell under ten children or thirty answers, a count not rounded to five, or one above its month's;
  - a range of chance the report does not count in, one there twice, or a promise outside its range;
  - a range of answers not named as the service names one, or ranges that overlap;
  - a share, a margin or an error that no answers give;
  - numbers of no month, or ranges of a month with no total;
  - a key it does not know, and a key it leaves out, which would read as nothing.

  It names each range by what it spans (§4), and a test holds its rule to the literals of the views' SQL. It does not hold a range of answers to the trial series of the commit being built: a month counted under another trial series would be refused by a commit that changed it. The site's reader checks the cells again by the rule in the data, and refuses a rule weaker than the views', which a test holds to their SQL too (§4).
- **On the page.** Three drawings side by side, each with its table hidden from the eye (§8); the month by `<Month>`; the rule, the trial series and the corridor through the data, since the words hold no digit (§9). The styles are in the shared stylesheet (§1, decision 9).
- **Forks.** A fork's live numbers are its own deployment's: it removes `site/research/live.json` and takes its own, and it either sets up an App of its own for `live.yml` or turns that workflow off (`docs/self-hosting.md`).

## 12. What gets built, and what is left

- **T72.13** builds the page, in two parts under rule 12:
  - **T72.13.1, the data:** the bench's commands `page` and `page-file`, `pageCriterion`, the fixture, the product's small exports, `just research-data` and `just site-paper`, and the step in `pages.yml`;
  - **T72.13.2, the page:** its components, its words in English and Russian, the drawings with their tables, the guard, the menu and the footer with the new measurement, the copy of the PDF into the built site, `published.ts`, the README, and `docs/self-hosting.md`. Built as this document says: the reader `web/src/site/research.ts`, the page in `ResearchPage.tsx` with its theses, its goals table and its numbers in components of their own, the authors in `site/data.json`, the guard of §9, the fold of R209 and its measure, `just site-fold`. `just site` makes the data before it builds.
- **T72.15** brought the live numbers, in three parts (§11):
  - **T72.15.1, the counts:** the nightly query's weighing, the two tables, the public and the private views, and the privacy policy;
  - **T72.15.2, the page:** the snapshot's SQL and `just site-live`, `page-file -live` with the made-up snapshot of the fixture, the reader's three states, and the block with its drawings, tables and words;
  - **T72.15.3, the schedule:** the snapshot's table and the identity that reads it, `just site-live` reading that table, and `live.yml` with its App.
- **T73.2–T73.4** move the making of `research.json` and of the PDFs into the research pipeline on merge to `main`; the file's shape and the page stay.
  - **T73.2, the data (R234):** `research.yml` holds the experiments to their results and makes `research.json` and `numbers.tex` from the release's commit. `pages.yml` downloads the data for a release and makes it on a pull request, and builds the site from the data in place with `just _site-build _site-check`, while `just site` still makes the data first. The build reads it at its default `--research`, beside the PDF it names, and the bench's steps both workflows run are one action, `.github/actions/page-data`.
  - **T73.3, the paper (R239):** a release builds the named and the anonymous PDF from those numbers, with the checks of every pull request, the named one printing its version, and hands them on as `paper-a`. No Russian PDF is built.
  - **T73.4, the paper on the page (R248):** `research.yml` builds the paper before the page's data, which takes the clean PDF in, and a pull request builds its site from the same run without publishing it. The PDF and its facts are no longer committed, and the page names the commit the paper's numbers were computed on.
  - **T73.5, the paper's releases (R254):** every version of the paper comes out as a release of its own on GitHub, tagged `paper-a/…`, with its artifact named and anonymous, each shown first to reproduce the experiments from itself; it goes to Zenodo only when the author runs `zenodo.yml` by hand. "Code and data" leads to the repository until the record's DOI exists.
- **R250, the second draft (2026-10-07):** the page's components and its rules in the shared stylesheet are drawn anew after the draft, with the first screen's way of a task, the theses as rows, the board of §6 and the live block's three columns; the bench's `page` runs seven cells, and `research.json` loses the rule before R187 and the goals' parameters, which no word of the page names any more.
- **R260, the PDF with every release (2026-10-07):** the page offers the named PDF of every release at `/assets/paper-a.en.pdf`, whatever placeholder it still prints, and the run's summary names the lines that print one; today that is the archive's DOI, until T73.6.

**Rejected:**
- the bench's numbers committed and checked for freshness: a committed file cannot name the commit it was computed on, any change of the model's inputs would need the whole run and a ten-thousand-line diff, and an arm64 machine parts in the last digits;
- the draft as it is: it needs a script and fetches from elsewhere;
- Source Serif 4;
- a Looker Studio report in a frame;
- the four chart libraries of §8;
- the report's own numbers on the page, which are exact and unrounded;
- the page's own numbers held to the guard's bands, which are drawn for 300 children: the guard holds the model on every pull request instead;
- "the last 30 days" in the live block;
- the rule before R187 beside the service (R250): the author found the two columns too much, and the bench's own run keeps the comparison;
- the draft's simulated live numbers until a month is counted whole (R250);
- a board 680 px wide that scrolls sideways on a narrow screen, as the draft's does (R250);
- shares behind ten answers rather than ten children;
- a floor under the live block's month: October, which mixes two models, is never counted whole;
- the latest month that showed numbers rather than the latest counted whole, which would pass an older month off as the latest;
- the deployment named in the snapshot, which the page never shows;
- the snapshot taken by hand each month, authorized views on `impact` for its reader, and a branch of data outside `main` (R226);
- a check that the ranges shown add up to no more than their month, which honest numbers fail, since each range is rounded to five on its own.

**Open for the author:**

1. **The 2027 call.** When AIED publishes it, its rule on preprints is read again before the title and the PDF stay public.
2. **The archive's DOI** (K11), which the author's first run of `zenodo.yml` brings (T73.6). Until then the PDF the page offers prints its placeholder (R260). Authorship (K05) was filled on 2026-10-07.

**Closed:** the answer in the card's memory (K10). The research found it at the paper's commit (S34, C099), the author accepted it on 2026-10-01 as a known limitation beside О-27, and T72.12 left it open here. R152, of 2026-10-03, had already removed it: `submit_task` draws no card, the card the service draws reads its task through `read_task`, and that carries no answer. A host that still draws a card from `submit_task`, from a list of the tools it fetched before R152, hands it that call's arguments, as its record of the tools called shows them, which О-27 leaves outside what the service hides. The paper reports the product at its own commit, `52ce869` of 2026-09-30, and so still describes the earlier card; the page describes the card as it is.

**Closed:** two copies of the draft. `draft/ui/site/research.html`, among the site's drafts (R156), was the same page as `research/draft-ui/`, with other links and two script tags. This page is ported from the latter; the former went with the site's drafts on 2026-10-06, and CLAUDE.md and SPEC 8.12.2 no longer name them as the site's reference.
