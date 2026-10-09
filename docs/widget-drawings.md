# Pictures drawn by the widget (T70)

**Recommendation: draw every task picture as graphics, from a short description the model writes, with no library.** 2026-10-08.

**Decided on 2026-10-08 (T70.1, R269).** The author went further than this recommendation: text pictures are thrown out entirely, so text mode keeps no text form either, and the wording carries every fact. The twelve kinds stay, the calendar among them. The checks read a picture's form, its labels and whether it shows the right answer; no rule asks for every number of a picture to be in the wording. The format is SPEC 4.4 and 5.4.

The run of 2026-10-08 (`docs/live/21-crooked-drawings.md`) found the text drawings crooked for two reasons the model cannot fix:

- Android's monospaced font has none of the drawing characters, so the columns part;
- a character grid cannot hold a clock face, so the hands point past their numbers.

This study asks whether the card can draw the pictures itself, and at what cost.

- **Coverage.** All 151 drawings of the study fit eleven kinds of picture: the 95 reference drawings, the 13 filled frames and the run's 43. Each kind is a description of a few fields.
- **No library.** Twelve small Preact functions draw them as SVG, eleven for those kinds and a twelfth for a calendar month. They add 4.6 KB to the widget compressed, 2 %. No library does better: the closest are Rough.js, for a hand-drawn look, and Konva, which draws on a canvas a screen reader cannot read.
- **Rendering.** The 151 pictures were measured on the real card in 16 combinations of browser, platform fonts, card width, theme and language, 2,416 measurements in all. None shifts, none scrolls sideways and no label runs out of its picture. The smallest label is 11 px.
- **Tokens.** Counted exactly on Opus 5.5's tokenizer:
  - the drawing's share of a task's package falls from 1,939 to 629 tokens;
  - what the model writes for a drawing falls from 129 to 51 tokens;
  - with the run's own prices, that is about $0.017 a task, 8 % of the $0.205 a task cost.
- **Not measured.** How long the model thinks over a text drawing, and how refusals change, need a live run.

The page of the study, every picture as text, as text on Android's fonts and as graphics, side by side: https://claude.ai/artifact/Ae1JjfT9JUQ9CWUZmPeu6h (private to the author).

## The starting point, checked

T70 began from four claims of an answer by Gemini. Each, against the pinned `@modelcontextprotocol/ext-apps` 2.0.3 and the hosts' documents (read 2026-10-08):

1. **"Konva cannot run inside ChatGPT or Claude."** Wrong for us. Our widget is an MCP Apps resource, our own HTML page with its own scripts. Whatever we bundle into it runs, under the host's content-security policy. That policy has no `'unsafe-eval'` (`dist/src/app.d.ts:46-52`: views run "under a strict CSP without `unsafe-eval`") and no network for an empty `ui.csp`. A bundled Konva runs; a library that compiles code at run time does not.
2. **"Render a PNG on the server and send a link."** Not needed, and the link would not load. With empty `resourceDomains` a host builds `img-src 'self' data:` (MCP Apps specification, `specification/2026-01-26/apps.mdx`, section "Content Security Policy Enforcement"). A `data:` PNG would load, but a raster has no theme, no text a screen reader reads, and is heavy where it sits in the result the model sees (О-39).
3. **"Claude can draw SVG in Artifacts."** True, and beside the point. The widget draws inline SVG itself. Nothing in the specification or Claude's MCP Apps guides restricts `<svg>` or `<canvas>`, and the widget already ships inline SVG icons.
4. **"The model gives a JSON description, the front end draws it with react-konva."** Right about the description, wrong about the tool. react-konva needs React's reconciler, and a description drawn by our own SVG functions needs nothing.

## What to draw from

**Today's structure cannot draw a picture.** `drawing_structure` (SPEC 4.4) has a free-text `kind`, objects whose `value` is optional, and relations nobody reads. Go reads only the labels (`checks/drawingmatch.go`). In none of the 34 reference clock faces does the structure hold the time: `clk-12-d1-1` is `{"kind":"clock","objects":[{"id":"hour","label":"4"}]}`. In the run, the model named its kinds six ways: `bar`, `bars`, `timeline`, `box`, `rows`, `groups`.

**The description.** A picture is one JSON object, `picture`, with a `kind` from a closed list and the fields that kind needs:

| Kind | Fields | Drawings of the study |
|---|---|---:|
| `clock` | `time` (`"9:40"`), optional labels at the hands' tips | 46 |
| `row` | `items` (`label`, `below`, `mark`: dot, ring or square, `skip` for a middle cut short), `gaps` (the label of every gap drawn), `span` (the label of the whole length), `line`, `copies` | 31 |
| `containers` | `items` (`capacity`, `amount`) | 15 |
| `grid` | `rows`, `cols` (their names), `filled` (cells such as `B2`), `marks` (a label in a cell) | 13 |
| `bars` | `bars` (`label`, `parts`, `shaded`, `length`, `segments`, `braces` under parts, `span`, `value`), `notes` (lines of equalities) | 13 |
| `piles` | `piles` (`label`, `count`, `shown`, `group`, `color`, `shape`, `boxed`), `box`, `across` | 11 |
| `venn` | `total`, `sets` (`label`, `count`), `both`, `neither` | 9 |
| `balance` | `left`, `right` (what stands on each pan) | 4 |
| `number_line` | `from`, `to`, `step`, `marks` (`at`, `label`) | 4 |
| `table` | `header`, `rows` of cells (timetables, numbered rows) | 3 |
| `ring` | `count`, `start` (`at`, `label`) | 2 |
| `calendar` | `weekdays`, `first`, `days`, `marks` | 0, for the site's pages |

**Coverage.** Every drawing of the study has a description, 151 of 151.

- **By parser.** 88 of the 95 reference drawings, by a parser for each kind that reads the drawing and its structure; 11 filled frames take their task's description.
- **By hand.** 52 drawings: the run's 43, seven reference drawings a parser would misread (braces, the ring, tables, two clocks with no time to parse) and two filled frames that differ from their task.

**Examples.** The frames become one example a kind:

```json
{"kind": "clock", "time": "4:30"}
{"kind": "grid", "rows": ["A", "B", "C"], "cols": ["1", "2", "3"], "filled": ["A1", "B1"], "marks": {"C3": "?"}}
{"kind": "bars", "bars": [{"label": "A", "parts": 3, "shaded": 3}, {"label": "B", "parts": 4, "shaded": 4}], "notes": ["A + B = 56", "A = ?"]}
{"kind": "venn", "total": "25", "sets": [{"label": "F", "count": "15"}, {"label": "C", "count": "12"}], "both": "5", "neither": "?"}
```

## What to draw with

Facts from the npm registry, bundlephobia (`https://bundlephobia.com/api/size?package=<name>@<version>`) and each project's pages, read 2026-10-08. The allowed licences are R21's: MIT, BSD-2, BSD-3, Apache-2.0 and ISC.

| Option | Licence | min / gzip | Preact without React | Output | Verdict |
|---|---|---|---|---|---|
| Our own Preact functions | ours | +11.7 / +4.6 KB, measured | yes | SVG | **Take it.** Exact geometry, theme colours by `currentColor` and CSS variables, `role="img"` with a label |
| Rough.js 4.6.6 | MIT | 27.1 / 8.8 KB | yes (`generator().toPaths()`) | SVG, canvas | Only for a hand-drawn look. Last release 2023-11-20 |
| d3-shape 3.2.0 + d3-path 3.1.0 | ISC | 32.8 / 5.7 KB | yes | path strings | Not needed: arcs and lines are a line of code each |
| svg.js 3.2.8, Two.js 0.8.24 | MIT | 92.6 / 30.0, 206.5 / 49.0 KB | yes, imperative | SVG | Redo what JSX does |
| Konva 10.7.1 | MIT | 186.7 / 55.6 KB | yes, imperative | canvas | "Canvas is a single element. Assistive technology sees no objects" (konvajs.org/docs/sandbox/Canvas_Editor.html); a theme change needs a redraw |
| react-konva 19.3.0 | MIT | 143.2 / 44.5 KB + React | no | canvas | Needs React's reconciler (konvajs/react-konva issue 55) |
| Mafs 0.21.0 | MIT | 321.9 / 94.9 KB | no | SVG | React only; its stylesheet imports a web font the policy blocks |
| JSXGraph 1.14.0 | MIT or LGPL-3.0 | 992.9 / 259.1 KB | yes | SVG | Its bundle holds `eval(` and a `new Function` test |
| Mermaid 12.1.0, Penrose 3.3.1 | MIT | 173–1,567 KB, 773 KB gzip | yes | SVG | Megabytes, and no school pictures |

Nothing small and maintained exists for school visual models (number lines, tape models, clock faces, jugs, Venn diagrams with fixed regions). The Venn libraries draw areas in proportion, where a task needs every region shown even when its count is 0.

## Size

Trial builds of the widget in a scratch copy, with the repository's own Vite settings (one HTML file):

| Build | Raw | gzip -9 |
|---|---:|---:|
| As it is | 825.3 KB | 220.1 KB |
| + the twelve functions | +11.7 KB | +4.6 KB (+2.1 %) |
| + Rough.js as well | +27.6 KB | +9.2 KB |
| + Konva as well | +188.2 KB | +55.8 KB |

The page has no size budget (SPEC 8.6, remark 18). SPEC 8.6 still quotes 543 KB; the page is 825 KB today, 56 % of it the dictionaries.

## Rendering on every device

**Method.** Each description's SVG went where the drawing stands in the preview's task card. The card ran in the pinned Playwright 1.63 image, in Chromium and WebKit, with the fonts of three platforms in place of the system's:

- Android: Droid Sans Mono, Roboto and Noto Sans Symbols from AOSP `android-16.0.0_r1`;
- Mac: DejaVu Sans Mono 2.37;
- Windows: Liberation Mono.

The card was 320, 360 and 736 px wide, in the light and the dark theme, in Russian and in Arabic.

**Results.** 151 pictures in 16 combinations:

| Measure | Result |
|---|---|
| Measurements | 2,416 |
| A picture scrolling sideways | 0 |
| A label running out of its picture | 0 |
| Smallest label on a 320 px card | 11 px (the ring's station numbers) |
| Same measures in Chromium and WebKit, and on every font | yes |

By comparison, four text drawings of the run scroll sideways on Android's 320 px card.

**Found and fixed in the prototype.** In a right-to-left card the lines of equalities under bars ran out of the picture by 35–88 px. An SVG's `dir` attribute does not reach its `<text>`, and `text-anchor="start"` flips under an inherited `direction: rtl`. The `direction` property on the `<svg>` fixes it, and the widget's version needs a test that holds it.

## What stays of the text drawing

- **Text mode is required (О-10).** Without a card, the model reads the task out of the result's `content`.
  - The service writes the text form from the description itself.
  - Clocks are written in words: "a clock showing 9:40".
  - Every other kind is drawn in ASCII alone: `+`, `-`, `|`, letters and digits. Droid Sans Mono has all of them, and the site's ASCII grids part by 0 px on Android.
- **Checks.**
  - **`drawing_format` becomes a check of the description.** It reads the kind, the fields and their ranges, for example a time from 0:00 to 23:59, cells inside the grid, and parts no fewer than those shaded.
  - **`drawing_mismatch` keeps its rule.** Every label the picture shows is named in the question, and every capital the question names is in the picture.
- **The site.** `site/cards_test.go` holds the site cards' drawings to `checks.DrawingFormat`; it moves to the description's check.
- **Old tasks.** A profile's old tasks keep their text, and `Diagram` keeps drawing a string as it does now.

## The hidden answer

The card draws only the description, and the description is all the child sees before answering. That makes three needs:

- **A check.** No field of the description may equal the correct option unless the question gives that number.
- **Numbers from the question.** Every number in a description must appear in the question, as a label the question names or as a number it gives.
- **The profile.** Today the profile keeps only the drawing's text and drops its structure (`submittask.go:496-505`). It keeps the description instead. The description is no secret: the model wrote it and sees it.

## The model

- **The guide's section.** It is rewritten from 966 tokens to 423. It keeps the rule on when to draw and when not to (R228), and says to describe the picture instead of drawing it.
- **The frames.** They become one example a kind, 87 tokens against a frame's 332.
- **Fewer refusals.** The run had 6 of 45 hand-ins refused over the drawing. Two of them gave a value as text. A typed field with a clear refusal ("`time` is H:MM") should end those, and the label check stays as it is.
- **One name for each kind.** The model's own kinds (`bar`, `bars`, `timeline`, `box`, `rows`, `groups`) collapse into the closed list.
- **Less thinking.** Clocks no longer need the model to work out where a slash points, which it got wrong in 5 of the run's 10 clock faces. How much thinking that saves is for a live run to show.

## Tokens

**Method.** Every count is exact, on `claude-opus-5-5`.

- **The count.** A `claude -p` call with no tools from an empty folder, asking for the word "ok": the input tokens with the text, cache reads and writes included, less those without it. Three empty calls agreed to 1 token in 2,833.
- **The sources.** The packages and hand-ins are the run's own: 38 packages and 36 drawings handed out.
- **The prices.** The price of each kind of token was solved from the run's 19 turns, cost against tokens, to within $0.0004 a turn: $8.01 a million tokens written to the cache, $0.20 read from it, $20.09 output. Each written token was read back 16 times in the run's lessons.

| Per task | Today | Description | Change |
|---|---:|---:|---:|
| Package: the guide's drawing section | 966 | 423 | |
| Package: the list of drawing characters | 150 | 0 | |
| Package: a frame, or a kind's example | 332 | 87 | |
| Package: a reference task's drawing | 126 | 47 | |
| **Package, all of the drawing (2 frames and 1.26 reference drawings on average)** | **1,939** | **629** | **−68 %** |
| **What the model writes for a drawing** | **129** | **51** | **−60 %** |
| The drawing in the service's answer, read again | 55 | 51 | −7 % |

**Money.** A task with a drawing saves 1,311 tokens of input and 77 of output.

- At the run's prices, with every written token read back 16 times, that is about $0.017 a task. A task cost $0.205 in the run, so the saving is 8 %.
- The guide's section and the list of characters are in every package, so tasks without a drawing save part of it too.

**Not counted:**

- **Thinking tokens.** The transcripts redact thinking, so only a live run can compare how long the model thinks over a text drawing and over a description.
- **Refusals.** Each refusal costs a whole hand-in, about 2,500 output tokens with thinking.

## The mock-up

- **The pages.** The study's page, https://claude.ai/artifact/Ae1JjfT9JUQ9CWUZmPeu6h, shows every drawing three ways:
  - the text as the card sets it, in the viewer's own system font, so on an Android phone it shows what a child sees;
  - the text on Android's fonts, from the stand;
  - the picture from its description, as live SVG.

  It also draws one clock with our functions, with Rough.js and with Konva, the two libraries loaded from jsDelivr for the comparison alone. It switches theme, card width (320 or 353 px) and writing direction. The gallery of the run is https://claude.ai/artifact/UgRdUS2pKc2qTofxahv624. Both pages are private to the author.
- **The prototype.** It is not committed: the functions, the parsers and the measuring scripts live in the session's scratch folder, with a copy of the repository at `c1276c3`.
  - `web/src/picture/picture.ts`, the twelve functions, on `h` from Preact;
  - `render.ts`, which renders descriptions with `preact-render-to-string`;
  - `specs.py`, the parsers and the hand descriptions;
  - `measure/pictures.mjs`, the stand, which imports `served`, `settled` and `stillClock` from `web/scripts/drive.ts` and runs in the Playwright image with the three font sets mounted at `~/.local/share/fonts`.

  T70.3 rewrites them with tests.

## Live checks

Each of these is only with the author's consent:

1. **The study's page in Claude on the author's Android phone.** Free. The left column shows the phone's own font, the right the phone's own SVG. It confirms or refutes the Droid Sans Mono finding on a real device.
2. **The prototype's card in Claude, on the web and on Android.** A scratch server behind a tunnel with a few tasks drawn from descriptions. A few messages of the subscription.
3. **The same in ChatGPT, on the web and on a phone.** A few messages; it belongs with T63.
4. **A lesson run on the prototype**, against the run of 2026-10-08. It measures:
   - output and thinking tokens;
   - refusals;
   - the time to a task.

   About $8 of the subscription, the cost of that run.

## Recommendation and draft tasks

Draw every task picture from a description, with our own SVG functions, and keep text only for text mode. In order:

- **T70.1 The format.**
  - The closed list of twelve kinds and the JSON schema of `picture`, replacing `drawing` and `drawing_structure` for new tasks.
  - SPEC 4.4 and 5.4, a decision in the log, and О-70 settled.
- **T70.2 The service.**
  - The check of the description: schema, ranges, labels named, no answer.
  - The text form for `content`.
  - The description kept in the profile.
  - The card's payload.
- **T70.3 The card.**
  - The functions in `web/src/design/`, used by `Diagram` for the widget and the site.
  - Tests: rendered with `preact-render-to-string`, the right-to-left case, and every kind in `just web-layout`.
- **T70.4 The content.**
  - The frames become examples.
  - The 95 reference drawings become descriptions, with the study's parsers, each one looked at.
  - The guide's section is rewritten, and the instructions' version moves.
- **T70.5 The solution card.**
  - The solution comes on a card of its own rather than in a block nested in the task's card.
  - Its steps, its total and its traps are shown in pictures, described by the model with the task and sealed with the solution until the child answers.
  - The author's draft `draft/Картинка в разборе.html` shows four ways to place the picture.
- **T70.6 The live run and version 0.5.0.**
  - Check 4 above, with T63 for ChatGPT.
  - Once the series is in, the release line goes to 0.5, and the release that ships it is v0.5.0.

The tasks are queued in RUN.md, first among the tasks not done (the author's decision of 2026-10-08).

## Open questions

- **О-70 (PRODUCT-V1 12.2).** Whether v1 draws task pictures from descriptions. Section 2 of PRODUCT-V1 keeps "real images (SVG, raster)" out of v1 and lists "text drawings made of characters", which this changes. **Settled on 2026-10-08:** pictures from descriptions, and no text drawings (R269).

## Remarks

1. **T58's premise.** T58 took Liberation Mono's set of characters to be Droid Sans Mono's (R138). Droid Sans Mono has 873 characters and, of the 65 a drawing may use, only `◊` (`docs/live/21-crooked-drawings.md`).
2. **Frames.** T70's text counts eleven frames; there are thirteen (SPEC 4.4, R66).
3. **Size in SPEC.** SPEC 8.6's 543 KB for the page is out of date: it is 825 KB.
