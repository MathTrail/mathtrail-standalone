# The "Research" page

The site's page of the numbers behind the product (О-57): what the paper about MathTrail found, how the student model stands against its goals, and where the reference tasks came from. This document decides how the page is built and where every number on it comes from (T72.12, R204). T72.13 builds the page from it. T72.15 later brings the live numbers. T73.2–T73.4 move the making of its data into the research pipeline on merge to `main`, and keep its shape.

The reference is the author's draft, `research/draft-ui/MathTrail - Исследование.html`: a self-unpacking page from a design tool, in Russian. Its markup is line 387 of the file, a JSON string (`sed -n 387p F | jq -r .`). The draft is ported, not copied: the site rebuilds it from its own components, in its own look, with every number from data.

## 1. What the author decided (2026-10-05)

| # | Decision |
|---|---|
| 1 | The paper's title is on the page from its first release, though the paper is under double-blind review. The AIED 2026 call accepted preprints that follow Springer's policies. The 2027 call, the one the paper goes to, was not out on 2026-10-02 (`research/paper-a/submission.md`), so this is read again when it is: a named preprint, once indexed, cannot be withdrawn. |
| 2 | The PDF is the named build, and only once it holds no placeholder. Today it prints `[Author]`, `[Affiliation]`, three `[TBD: K05 …]` blocks and `[TBD: K11: the artifact's DOI]`, so the page has no PDF button until authorship (K05) is filled and the artifact has its DOI (K11). The build refuses a PDF that still holds a placeholder. |
| 3 | Every language links the English PDF until T73.3 builds the Russian one on merge. |
| 4 | Search engines may index the PDF from the start. |
| 5 | The model's goals are bounded against the current service, as the bench's criterion bounds them: the bar is for the next rule. |
| 6 | Where a goal's bound is computed from the service's own value of the same measure, the service's row says "baseline" rather than a mark. |
| 7 | CI computes the bench's numbers from the commit it builds, and they are never committed. The other inputs are small committed files: the PDF's facts, later the monthly live snapshot, and the curated books. |
| 8 | The live block comes later, in a task of its own (T72.15). Until then the page shows the block's frame and says the data is still to come. |

## 2. Address and frame

- **Address.** `/<locale>/research/`, words in `site/content/<locale>/research.yaml` (SPEC 8.12.2). Anchors: `#student-model`, `#live`, `#sources`, `#paper`.
- **Menu and footer.** "Research" stands before "About" in the menu and first in the footer (SPEC 8.12.2, R202). R202 measured the fold with five entries. A sixth moves it, so T72.13 measures it again in Chromium and WebKit, in both languages, and records the new width. If the Russian «Исследование» moves the fold too far, the author chooses a shorter word.
- **Lists.** The addresses join `web/tools/sitecheck/published.ts`. The README links `/en/research/`.
- **No script, nothing from elsewhere.** The page reads in full without a script and runs none (R160). It loads nothing from another origin, so the draft's unpkg React and its fonts' `preconnect` go.
- **Fonts.** Onest only (R157): the draft's Source Serif 4 alone weighed 0.9 MB.
- **Weight.** The shared stylesheets and fonts weigh about 96 KB; the page's HTML, with its tables and inline drawings, should weigh 40–80 KB. That is 140–180 KB in all, under the 300 KiB a page may weigh. The PDF is a link, not a resource the page fetches, so it does not count.
- **Languages.** English and Russian now, the rest with T66b. Drawings are pinned left to right (`dir="ltr"`), as the topics' map is; tables and words follow the page's direction.

## 3. The page, section by section

| Section | In the draft | On the site |
|---|---|---|
| Hero | The paper's title as the heading; a lead; "The whole paper · PDF, 16 pages" and "Code and data"; chips: 603 reference tasks · 9 checks · 17 topics · 20 traps · grades 1–6 · MIT | The same. The title is the paper's, in the page's language. The PDF button waits for a clean PDF (§10), and its page count comes from the file. Every chip's number comes from data (§5). "Code and data" leads to the repository. |
| Four theses, 01–04 | The chat's model writes; the solver runs twice; the answer is sealed; a little harder, but within reach | The same, with each number from the product (§5) and the worked example's values marked as chosen by the text (§9). **Thesis 03 is corrected.** The draft says the answer is in neither the card nor the logs. In fact, `submit_task` draws the card, and MCP Apps hands the card that call's arguments, the key and the solution among them (K10). The card never shows them, but until the child answers they sit in its memory. The page says so. What the product does about it is open (§12). |
| Student model, `#student-model` | A table: metric, now, goal, ceiling, new rule, mark. Its "now" was the rule before R187, and its "new rule" was empty | The goals table of §6: before R187, now, goal, ceiling and mark, with the run's provenance under it. |
| Live data, `#live` | Promised → came true, came true less promised by answer count, the share right on the rule's tasks — all marked as simulated | The block's frame and a sentence that the numbers come once there are enough of them (§11). T72.15 fills it. |
| Where the reference tasks came from, `#sources` | Public-domain books by author, with years; a flow from a book to a checked task; 603 tasks as a bar of 200 · 250 · 153 | The same. The books, their authors and years are curated data in `site/data.json` (`research.sources`), as "Why" keeps its works. The counts by level come from the content. |
| The whole paper, `#paper` | What the paper holds, and its page count | The same, with the PDF and its page count once there is a clean PDF (§10). |

**What leaves the draft:**
- the "expected" band beside the estimate's error, which no data gives;
- the simulated live charts and their counts, which were not rounded as public counts must be;
- "the last 30 days", which R192's closed months replace;
- the empty "new rule" column: the rule it waited for shipped in R187, so the table compares before and now;
- the draft's own menu and footer, and Source Serif 4.

The theses and the sources keep the draft's order and substance. Their words are written anew in English from the Russian and are the source, as for every page (R159).

## 4. The data file

**`research.json`** holds every number of the page. The page's code reads it through one reader, `web/src/site/research.ts`, which checks it with zod as `data.ts` checks the site's data.

The file is not committed. It is made at build time, in `site/research/research.json`, which `.gitignore` lists. In CI the file is made in the site's build job (§7). Locally, `just research-data` makes it. Two kinds of input sit committed beside it in `site/research/`: the clean PDF with its facts (§10), and later the live snapshot (§11).

```jsonc
{
  "schema": 1,
  "built_from": { "commit": "<40 hex>", "date": "<the commit's date>" },
  "bench": {
    "producer": "tools/learners page",
    "inputs": "<sha256 of the inputs (§7)>",
    "seed": 20261001, "experiment": "E-A3", "children": 1000, "answers": 200,
    "interval": 0.95, "resamples": 2000,
    "goal_parameters": { "lag_share": 0.45, "corridor_share": 0.27, "unsettled_most": 0.25, "false_most": 0.20, "late_times": 1.5 },
    "rules": [ { "id": "shrinking/both", "role": "service" },
               { "id": "earlier/both", "role": "earlier" },
               { "id": "oracle/both", "role": "ceiling" } ],
    "rows": [ {
      "id": "lag", "kind": "goal", "criterion": "step",
      "generator": "G2-half", "metric": "r6_lag", "unit": "logit", "read_as": "size", "better": "lower",
      "bound": { "value": 0.2420, "own": true },  // the file holds every number in full; rounded here
      "values": {
        "service": { "value": 0.5377, "low": 0.5202, "high": 0.5556 },
        "earlier": { "value": 0.6092, "low": 0.5896, "high": 0.6296 },
        "ceiling": { "value": 0.0000, "low": 0.0000, "high": 0.0000 } },
      "mark": "baseline" } ]
  },
  "product": {
    "topics": 17, "traps": 20, "checks": 9, "grades": { "first": 1, "last": 6 },
    "reference_tasks": { "total": 603, "by_level": { "1-2": 200, "3-4": 250, "5-6": 153 } },
    "options": 5, "relabel": 2, "guess": 0.2,
    "corridor": { "low": 0.70, "high": 0.85, "middle": 0.775 }, "trial_answers": 5
  },
  "paper": {
    "commit": "52ce86908135",
    "files": [ { "lang": "en", "path": "/research/paper-a.en.pdf", "pages": 15, "bytes": 350493, "sha256": "…" } ]
  },
  "live": { "state": "coming" }
}
```

**Rules for the file:**

- **Byte for byte the same** for the same inputs, on amd64, the architecture CI runs on. On arm64 Go fuses a multiplication and an addition into one step, and the last digits part.
  - Numbers are written in full, in the shortest form that reads back to the same value (Go's `strconv.FormatFloat(v, 'g', -1, 64)`), and rounded only on the page, so the reader can recompute every mark exactly.
  - There is no wall clock: the date is the commit's.
  - T73.2's check that two runs give the same bytes rests on this, on amd64.
- **Values as the criterion reads them.** A goal on a size, such as the lag, holds the size of the value and of its interval, as the bench's criterion reads it (`sizeOf`: an interval across zero reads from zero). `read_as` names the reading, and `better` applies to it.
- **One producer per block.** The bench's new command `page` writes `bench` and `product` (§6, §5). `just research-data` adds `paper` from the committed facts of the PDF and `live` from the snapshot, or `{"state": "coming"}` when there is none. T73.2 later writes the same file from `research.yml`: only the producer's name changes.
- **The reader refuses what does not add up:**
  - a mark that does not follow from the value, its interval, the bound and the direction;
  - product counts that disagree with the catalogs and reference tasks the site already reads;
  - a PDF the site does not ship, or one whose size or hash differs from its facts;
  - later, a live cell shown below the privacy thresholds, or a count not rounded as §11 says.
- **One fixture of the whole file**, `site/research/testdata/research.json`, holds both sides to one shape.
  - The assembler makes it from a run of ten children, a fixture of the PDF's facts and no live snapshot, with the commit fixed by the fixture.
  - Its test makes the file again byte for byte; a meant change rewrites it with `-update`, and the diff is read in review.
  - The reader's test parses the same file.
  - A change of shape in any block fails one side or the other.
  - The bytes hold on amd64 alone, so the fixture's test runs there and skips elsewhere, as the bench's carried-over check does. It refuses `-update` off amd64.

## 5. Where each number comes from

| Number on the page | Source |
|---|---|
| Reference tasks, in all and by level; topics; traps | The content: `content/examples/*.json`, `content/catalogs/topics.json`, `traps.json`. The site already reads them (`siteData()`); `product` repeats them so the reader can hold the two to each other. |
| Checks | `internal/domain/checks`, through a list it exports, `checks.Codes()`, held by a test to its constants. Today the nine codes are constants alone, which neither the bench nor TypeScript can count. |
| Grades 1–6 | The catalog's grade levels. |
| Options per task, the shift between the solver's two runs, the guess floor, the corridor, the trial series | The product's constants: `solver.Count`, the solver's relabelling (unexported today as `relabel`, so it gets an accessor), `rating.Guess`, the corridor's bounds (unexported but for its middle, so they get an accessor), `rating.TrialAnswers`. |
| The worked example of thesis 02 | Its five option values are chosen by the text and marked as such (§9). The second run's letters are computed with the solver's own relabelling. |
| The goals table | The bench's `page` run (§6). |
| The paper's title | The paper: `research/paper-a/main.tex` for English, the first heading of `research/paper-a/draft.ru.md` for Russian. The words carry it, and a test holds the English words to `\title`. |
| The PDF's pages, size and hash | The PDF itself, read when it is copied into the site (§10). |
| Books, authors and years | Curated data, `site/data.json` → `research.sources`. |
| Live numbers | T72.15's public aggregates (§11). |

## 6. The goals table

**Where the numbers come from.** The bench gets a new rule set, `page`: eleven cells, those of the table and nothing else.

| Rule | Generators |
|---|---|
| The service, `shrinking/both` | G0, G0-topics1, G2-half, G3 |
| The rule before R187, `earlier/both` | G0, G0-topics1, G2-half, G3 |
| The ceiling, `oracle/both` | G0, G2-half, G3 |

A cell's children and its intervals depend on nothing but its rule and its generator: the children are drawn from seeds named after the generator, and every interval from a stream named after its cell and metric. So these cells give exactly the numbers of the whole run, to the last digit.

**Which goals.** A new `pageCriterion` reads both criteria's goals, those of the step and those of mastery, against the service as the baseline. Since R187 the baseline of mastery's criterion, the floor under the step, runs under the service's own mastery, so its cells are the service's. Rows:

| Row | Generator, metric | Bound | Service's row |
|---|---|---|---|
| The lag behind a child who learns | G2-half, `r6_lag` | 45 % of the service's lag | baseline |
| Tasks in the corridor, a child who learns | G2-half, `r3_inside` | the service's, plus 27 % of the way to the ceiling | baseline |
| Children not caught up after a jump | G3, `r6_jump_unsettled` | at most 25 % | a mark |
| Masteries declared falsely | G0 and G0-topics1, `r4_false` | at most 20 % | a mark |
| Answers until a mastery is declared | G0 and G0-topics1, `r5_late_answers` | 1.5 times the service's | baseline |
| What the child is shown late in a run: the card's move and the rank's changes in answers 150–200 | G0 and G2-half, `r8_move_p95_150_200`, `r8_rank_150_200` | the service's in answers 6–20 | a mark |
| *Context, no goal:* the error after 200 answers, a child who stays put | G0, `r1_rms_200` | — | — |
| *Context, no goal:* tasks in the corridor, a child who stays put | G0, `r3_inside` | — | — |

The screen's goals for answers 6–20 are left out: their bound is the service's own value in the same window, so they could only ever say "baseline".

**Columns.** Before R187 (the earlier rule), now (the service), the goal, the ceiling and the mark. The ceiling is the oracle's for the step's measures. For mastery's measures it is perfection, no false mastery and no wait, since the oracle keeps the service's mastery and is no ceiling for it (`docs/learners.md`). The oracle shows no screen.

**Marks.** A goal is read on the value and its 95 % interval, by the bench's own rule:
- **reached** when the whole interval is on the right side of the bound;
- **not reached** when the whole interval is on the wrong side;
- **on the edge** otherwise.

Some bounds are computed from the service's own value of the same measure, the same generator and metric: `own` in the data. There the service's mark is decided in advance: a share of its own value is never reached, and a multiple of it always is. So its row says **baseline**, and the page says plainly that such a goal is the bar for the next rule.
- The lag, the corridor of a child who learns and the answers until a mastery are such goals.
- The screen's late window is not, since its bound is the service's value in another window, a measure of its own.
- The rule before R187 gets marks everywhere: no bound is its own.

**Provenance.** Under the table: the commit the numbers were computed on, the seed, the children per cell and the answers per child, and that no child and no language model took part. One more line says that the paper reports the rule as it stood at its own commit, `52ce869`. This page follows the service as it runs, so the two differ.

## 7. The pipeline

- **On every build.** The site's build job (`pages.yml`), before it builds the site:
  - restores the bench's cache;
  - runs `tools/learners page` when the cache misses;
  - assembles `research.json` with `just research-data`;
  - writes the goals table to the job's summary.
- **The cache's key** is a hash of everything the cells depend on:
  - the files of every package the bench builds from (`go list -deps`), the product's included;
  - the embedded content;
  - the bench's own sources, but not its `results/`;
  - both `go.mod` and both `go.sum` files, since a module's `go` line sets the language and its defaults;
  - the run's own parameters as the command line gives them: the rule set, the children, the answers, the seed and the experiment;
  - the toolchain image's tag;
  - the processor's architecture, since arm64 parts in the last digits.

  Nothing else moves the numbers, and anything that does changes the key.
- **On a pull request** the numbers are computed and shown in the summary but not published. The cache is restored and saved, so a pull request that leaves the model alone computes nothing.
- **On a run that publishes** the numbers are computed again from its commit, with no cache restored: nothing an earlier run left behind reaches the site. That is a release, which `release.yml` starts after green CI on `main` (R195), and a build started by hand. The result is saved to the cache for the pull requests after it.
- **Cost.** The eleven cells take about 3.4 processor-minutes: the whole run's 59 over its 190 cells, times eleven. On the four processors of a runner that is a minute or two, plus the bench's build. T72.13 measures it and writes the figure here.
- **Failure.** A failed computation fails the build, and the release publishes nothing: the site stays as it was. A release only follows a CI run in which the bench's tests and its guard passed.
- **The guard holds the model, not the page's numbers.** T72.12 asked for the guard to hold the page's numbers to its bands.
  - It holds the service on every pull request (`ci.yml`), and the page's numbers come from the same code. A release only follows a green CI, so it never shows the numbers of a model the guard refused.
  - The page's own numbers are not held to the bands themselves: the bands are drawn for the guard's 300 children, and the page reads 1,000.
  - The page never shows the guard's numbers.
- **Handover.** T73.2 makes the research pipeline the one producer of this file on merge; `pages.yml` then downloads it rather than making it. The file's shape stays.

## 8. Drawing

Every chart is drawn at build time by the page's own Preact components, in SVG or as HTML bars. No library, no canvas, no script. The site already draws this way: the topics' map (`TopicMap.tsx`), the ages of the techniques (`.s-ages-bar`), tables in `TableFrame`.

| Chart | Drawn as | Its table |
|---|---|---|
| The goals | The table is the chart: a bar for each row, a whisker for the interval and a tick for the bound, hidden from screen readers, which read the cells | Itself |
| The reference tasks by level | A bar of three parts, each as wide as its count | Its labelled parts |
| The theses' diagrams | HTML, as in the draft | — |
| Later (T72.15): promised → came true | SVG: the diagonal, the corridor's band, a point per range of chance | Under it |
| Later (T72.15): came true less promised, by answer count | HTML bars around zero | Under it |
| Later (T72.15): the share right on the rule's tasks | One bar, with the corridor and the promised share | A row |

**Numbers** are formatted with `Intl.NumberFormat(locale)`: shares as percentages, intervals with `formatRange` where the runtime has it, and tables in tabular figures.

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

1. **Two components.**
   - `<Num>` writes a number from data as `<data value="0.5377">0,538</data>`, or as a marked `<tspan>` inside SVG.
   - `<Given>` marks a number the text chooses rather than reports: the worked example's option values, an axis's 0 and 1. Book years are data, not given.
   - An identifier from data that holds digits but is no number is written in its own element: a commit's hash in `<code>`, a date in `<time datetime>`. The scan passes such an element only when its text is a value of the file: the commit the numbers were computed on, the paper's commit, the commit's date.
2. **The words hold no digit.** A test fails on any digit in `research.yaml` outside an allowed name. Counts in a sentence come through slots with plural wordings, as on "Techniques".
3. **The rendered page is scanned in every language.** Every text node is read, and so are the `title`, `aria-label` and `alt` attributes and the description. The scan skips scripts and styles, the two components' elements, and the names with digits of a list kept beside the page: XChaCha20-Poly1305, AIED 2027, Glicko-2. It fails on any run of digits in any script (`\p{Nd}`), since T66b brings languages with digits of their own.
4. **Every `<data value>` is traced** to a value of `research.json` or of the curated data.
5. **A number changed in the data changes the page.** The test renders the page from its fixture and again with every number moved, and every `<data>` element must change.

## 10. The paper on the page

- **Title.** In English, "The Model Writes, the Service Checks: Olympiad-Style Maths Problems for Primary-School Children inside Chat Assistants". In Russian, the Russian draft's title. Both are words of `research.yaml`, the English held to `research/paper-a/main.tex` by a test.
- **The PDF.** A recipe of T72.13, `just site-paper`:
  - builds the named PDF with `just research paper-a`;
  - refuses a paper whose sources still print `[Author]`, `[Affiliation]` or a `\TBD` — today the three of K05 and the artifact's DOI of K11;
  - copies the PDF to `site/research/paper-a.en.pdf` and writes its pages, size, hash and the paper's commit to `site/research/paper.json`.

  Both are committed: CI builds no paper until T73.3 brings TeX Live into it. With no file, the page shows the title and "Code and data" but no PDF button.
- **Every language** links the English PDF until T73.3 builds the Russian one.
- **Indexing.** The PDF is open to search engines.
- **Self-archiving.** Springer's self-archiving rules may limit which version can stay on the site once the paper is accepted; that is read when the venue decides.
- **Forks.** A fork removes the paper's PDF and title, which are the authors' work: `docs/self-hosting.md` says so, as it does for the coach's page.

## 11. Live numbers, later (T72.15)

The live block shows the same measures as `just report` (R153), counted for good and made public:
- promised → came true, by range of chance;
- came true less promised, by the child's answer count;
- the share right on the rule's tasks.

The answers are those to tasks the rule chose, after the trial series and without the hint. The ranges of chance are the report's six, and the answer counts are the report's ranges from 6 on; the draft's "1–5" is the trial series, which is left out. Every host is taken together, and the children of the load tool and MCP Inspector are left out.

- **What it needs.** The `answer_recorded` lines in the `activity` bucket already carry `chance`, `tutor_mode`, `trial`, `hint_used` and `answers_bucket`, all the report's filter reads; only the `impact` tables lack them. So T67.2's nightly query gains two tables, by month and instructions version:
  - `chance_monthly`: children, answers, right answers and the chances promised, summed;
  - `kept_up_monthly`: children, answers, and the sums from which the mean and its error by child are computed with no child named.
- **Privacy, by R192:**
  - closed months only, each month on its own and never a running total, which would give a small month away by subtraction;
  - a cell is shown only behind at least ten children and thirty answers;
  - the groups of a month that fall short — a version, a range — fold into `(others)`, topped up from the smallest until the fold itself clears both thresholds, as R192 folds the public views. So no hidden group can be had by subtracting the shown ones from a month's whole;
  - counts are rounded to five;
  - a hidden cell shows neither a count nor a share.

  T72.12's "ten answers" yields to R192's ten children.
- **States:** coming, too few, ready. R187 rolled out on 2026-10-04, so October mixes two models. The first whole month is November, and its numbers come in December 2026 at the earliest.
- **Into the build: a monthly snapshot, decided at the start of T72.15.** A recipe like `just impact`, with the author's own credentials, writes `site/research/live.json` with the deployment, the months and the views read. It is committed by a pull request each month.
  - The snapshot needs no credentials in CI and gives the numbers a second look in review.
  - The alternative, reading BigQuery from `pages.yml` through Workload Identity, needs its own task. `impact_public` is not an authorized view today, so its reader could read the exact `impact` tables. The identity pool binds by repository alone, and the build job runs npm, which should not hold `id-token: write`.

## 12. What gets built, and what is left

- **T72.13** builds the page. Under rule 12 it likely splits in two:
  - the data: the bench's `page` set and command, `pageCriterion`, the golden file, the product's small exports, `just research-data` and `just site-paper`, and the step in `pages.yml`;
  - the page: its components, its words in English and Russian, the drawings with their tables, the guard, the menu and the footer with the new measurement, `published.ts`, the README, and `docs/self-hosting.md`.
- **T72.15** brings the live numbers: §11.
- **T73.2–T73.4** move the making of `research.json` and of the PDFs into the research pipeline on merge to `main`; the file's shape and the page stay.

**Rejected:**
- the bench's numbers committed and checked for freshness: a committed file cannot name the commit it was computed on, any change of the model's inputs would need the whole run and a ten-thousand-line diff, and an arm64 machine parts in the last digits;
- the draft as it is: it needs a script and fetches from elsewhere;
- Source Serif 4;
- a Looker Studio report in a frame;
- the four chart libraries of §8;
- the report's own numbers on the page, which are exact and unrounded;
- the page's own numbers held to the guard's bands, which are drawn for 300 children: the guard holds the model on every pull request instead;
- "the last 30 days" in the live block;
- shares behind ten answers rather than ten children.

**Open for the author:**

1. **The answer in the card's memory (K10).** The product's rule that the answer stays hidden does not hold for the card: the host hands it `submit_task`'s arguments, the key and the solution among them. Neither PRODUCT, SPEC nor the decisions record it. Thesis 03 says it as it is. Whether the product draws the card from another call, or records an accepted limitation, is a decision of its own.
2. **The 2027 call.** When AIED publishes it, its rule on preprints is read again before the title and the PDF stay public.
3. **The PDF waits** for authorship (K05) and for the artifact's DOI (K11).

**Not this task's, noted:** two copies of the draft.
- `draft/ui/site/research.html`, among the site's drafts (R156), is the same page as `research/draft-ui/`, with other links and two script tags. This page is ported from the latter, and the former goes when the site's drafts do (T65).
- CLAUDE.md and the opening paragraph of SPEC 8.12.2 name `draft/ui/site/` as the site's reference.
