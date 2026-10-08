# Crooked drawings

**Measured: two causes, and the model's own typing is not one of them.** 2026-10-08. The author saw two drawings come out crooked in Claude, on an Android phone and on a computer, the clocks worst of all. To find out why, four lessons were held in which every task asked for a picture. Each drawing was then photographed on the real card in the fonts an Android phone and a computer would pick, and every drawing the project ships was measured again.

- **Android parts the columns.** The card sets a drawing in the system's monospaced font, and on Android that font is Droid Sans Mono. Droid Sans Mono has one of the 65 characters a drawing may use besides ASCII, `◊`. Android draws the other 64 from Noto Sans Symbols, where a box-drawing character is 9.91 px wide against a cell of 8.40 px, and an arrow is 14 px wide. A line that mixes letters or digits with lines therefore drifts against the lines above and below it.
  - Shipped drawings: 86 of the 229 the project ships part by a whole cell or more, among them 11 of the 13 frames and 54 of the 95 reference drawings. On the fonts of a Mac or Windows, none does.
  - Lessons: 16 of the 36 drawings the run handed out part by a whole cell or more, 22 by half a cell or more.
- **A clock face does not fit the grid, on any device.** A cell of 8.4 × 20 px cannot hold twelve numbers on a circle. The frame's numbers stand up to 16.4° off their places. A hand drawn with `/` or `\` can only run 22.8° off the vertical, so no hand can point at 2, 4, 8 or 10. In the run, 5 of the 10 clock faces point a hand at the wrong number. Among the reference clock faces the model learns from, 20 of 33 do.
- **Rows leave gaps.** A row is 20 px high, and a vertical line is drawn 17–18 px tall, so every frame is dashed. A slash is 10–13 px tall, so a slanted hand falls apart into strokes.
- **The model's text was sound.** None of the 43 drawings handed in has a frame, wall or table out of line. The model drew whenever it was asked: in all 19 tasks the parent asked for, and in 17 of the 19 it wrote ahead.
- **A probe of a fix works.** The probe set every character of a drawing in a cell one character wide and closed the rows to 17 px. On Android's fonts the columns then line up and the lines join, in Chromium and WebKit alike, with no font shipped and no drawing changed. It does nothing for the clock faces.

T58 calibrated the drawings on stand-in fonts and took Liberation Mono's set of characters to be Droid Sans Mono's (R138, `06-drawings.md`). That premise was wrong, and it is why the calibration passed what Android breaks.

## What this is

A run at the author's request, outside the RUN.md plan, to see which drawings come out crooked and why. Nothing in the product changed. The choice of a fix is the author's, from the options at the end.

## The run

- **Service.** A clean copy of commit `c1276c3`, served by `go run ./cmd/server` as `just play-server` serves it:
  - port 8081, the development sign-in and the profile in memory;
  - the day's ceilings raised to 200 accepted tasks and 30 failed requests.
- **Chat.** Claude Code 2.1.280 in `-p` mode, through `just play`. The model was `claude-opus-5-5` at `--effort medium`, the setting the author uses in claude.ai (16-task-time).
- **Profile.** One for the whole run: the pseudonym Тест, grade 3, space and football, lessons in Russian.
- **Lessons.** Four of them, each one chat, with the parent's messages sent in turn by `--resume`. Every message asked for a task with a picture and named its topic, level and difficulty. From the second message on, it opened with the child's answer to the task on the card, the right letter. The last task of each lesson was answered by a call to `submit_answer`, as its card would answer it.

| Lesson | Tasks asked for |
|---|---|
| 1 | `time.clocks`: `1-2` at difficulties 2 and 4, `3-4` at 2 and 4, `5-6` at 3. Three messages asked for "a picture of the clock", two for "a picture" |
| 2 | `algorithms.weighing_pouring` `3-4`, `counting.gaps` `1-2`, `logic.ordering` `3-4`, `parity.alternation` `3-4`, `combinatorics.enumeration` `3-4` |
| 3 | At `5-6`: `geometry.grid`, `logic.sets`, `ratio.sharing`, `fractions.parts`, `games.strategy` |
| 4 | Topics with no frame: `time.calendar` `3-4`, `pigeonhole.basic` `3-4`, `logic.knights_liars` `3-4`, `number.divisibility` `5-6` |

**Tasks.** Every turn wrote two tasks: the one the parent asked for, and one ahead for the next. That made 38 tasks in 19 turns, and 45 hand-ins.

**Refusals.** 7 hand-ins were refused, and every request was accepted by its second attempt.

- 4 were for labels (`drawing_mismatch`).
- 2 gave a drawing's `value` as text, which must be a whole number (`bad_structure`).
- 1 was for a sentence of 32 words (`readability`).

**Cost and time.** $7.78 of the author's subscription by Claude Code's count, a little under the $8–15 the plan estimated. A resumed chat reports its cost as a running total of the session, so the run's cost is the sum of each lesson's last turn. 24 minutes of chats. No chat stalled, and the watchdog ended none.

## How the drawings were measured

- **On the card.** Each drawing went into `.mt-diagram` of the task card in the widget's preview (scene `task`), its question and options with it. The card was 320 px wide, the card of a 360 px Android phone in Claude, and 736 px wide, Claude on the web.
- **Browsers.** Chromium, which Android's WebView and Claude's desktop app run on, and WebKit. Both came from the pinned Playwright 1.63 image.
- **Fonts.** Each set took the place of the system's font:

  | Platform | Fonts |
  |---|---|
  | Android | `DroidSansMono.ttf`, `Roboto-Regular.ttf` and `NotoSansSymbols-Regular-Subsetted.ttf` from AOSP `android-16.0.0_r1`, in the order of that release's `fonts.xml`: `monospace` is Droid Sans Mono, Roboto is the first fallback, and Noto Sans Symbols is the first fallback that has these characters |
  | Mac | DejaVu Sans Mono 2.37, whose glyphs have Menlo's sizes |
  | Windows | Liberation Mono, which has Consolas's set of characters |

- **Measures.** Each character's place came from a range over it, against the grid of cells. A column is the same column in every line it holds a mark in, and its spread is how far that mark moves between lines.
  - The widths of the Android characters were also read from the fonts' tables.
  - The same model, run over the text alone, agrees with Chromium to 0.1 px. WebKit agrees with Chromium to 0.8 px.
- **The text itself.** Two checks read it with no browser:
  - the joints of the box-drawing lines, where an arm must meet an arm;
  - for a clock face, the angle of each number and each hand from `●`, on the 8.4 × 20 px cell.

## 1. Android draws the drawing's characters from another font

| Characters | On Android, from | Width | Against the cell of 8.40 px |
|---|---|---:|---:|
| `─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼`, the double lines | Noto Sans Symbols | 9.91 px | +18% |
| `▀ ▄ █ ▌ ▐ ░ ▒` | Noto Sans Symbols | 9.91 px | +18% |
| `▓` | Noto Sans Symbols | 10.21 px | +22% |
| `■ □ ► ◄ ◘ ◙` | Noto Sans Symbols | 10.50 px | +25% |
| `▲ ▼` | Noto Sans Symbols | 13.86 px | +65% |
| `▬ ← →` | Noto Sans Symbols | 14.00 px | +67% |
| `↑ ↓` | Noto Sans Symbols | 7.00 px | −17% |
| `◦` | Noto Sans Symbols | 4.96 px | −41% |
| `● ○` | Noto Sans Symbols | 8.46 px | +1% |
| `◊` | Droid Sans Mono | 8.40 px | 0 |

- **It holds across Android releases.** Droid Sans Mono has 873 characters: Latin, Greek, Cyrillic, and some punctuation and mathematics. Its file is the same byte for byte in Android 12, 14 and 16, and so is Noto Sans Symbols. The fallback order is the same in all three too. So it holds on any phone that keeps the stock fonts.
- **Shipped drawings.** Drawings whose columns part by a whole cell or more, on each platform's fonts:

| Drawings | Count | Android | Mac | Windows |
|---|---:|---:|---:|---:|
| Frames | 13 | 11 | 0 | 0 |
| Filled frames | 13 | 10 | 0 | 0 |
| Reference tasks | 95 | 54 | 0 | 0 |
| Calibration set (T58) | 18 | 7 | 0 | 1 |
| Site, each drawing once | 90 | 4 | 0 | 0 |

The one drawing on Windows is calibration drawing 16, whose characters a drawing may not use.

- **The worst.** The timetable frame parts by 25.7 px, three cells. The grids part by 15–21 px. The Venn frame and its six reference tasks part by 37.8 px, and so do the run's Venn drawings.
- **ASCII holds.** The site draws its grids in ASCII, `+---+` and `|`, and they do not part on Android at all.
- **Width.** The wider characters also make some drawings too wide for the card. On Android, the Venn frame and its six reference drawings are 267.6 px and the number line 281.6 px, against the 260 px a 320 px card leaves a drawing. Both scroll sideways on a phone where they fit on a computer. In the run, the two Venn drawings and the two drawings of boxes of `■` did.

## 2. A clock face does not fit the grid

On the clock frame, each number's angle clockwise from 12 and its distance from `●`, with the cell at 8.4 × 20 px:

| Number | Where it stands | Where it belongs | Off by | Distance |
|---:|---:|---:|---:|---:|
| 1 | 40.0° | 30° | +10.0° | 52.2 px |
| 2 | 68.4° | 60° | +8.4° | 54.2 px |
| 3 | 90.0° | 90° | 0 | 67.2 px |
| 4 | 108.8° | 120° | −11.2° | 62.1 px |
| 5 | 133.6° | 150° | −16.4° | 58.0 px |
| 6 | 180.0° | 180° | 0 | 60.0 px |
| 7 | 226.4° | 210° | +16.4° | 58.0 px |
| 8 | 251.2° | 240° | +11.2° | 62.1 px |
| 9 | 270.0° | 270° | 0 | 67.2 px |
| 10 | 290.1° | 300° | −9.9° | 58.1 px |
| 11 | 316.6° | 330° | −13.4° | 55.0 px |
| 12 | 356.0° | 0° | −4.0° | 60.1 px |

- **The numbers.** The gap between 3 and 4 is 19°, the gap between 5 and 6 is 46°, and the distance from the centre runs from 52 to 67 px. The face is a lopsided diamond rather than a circle, and every one of the 34 reference clock faces inherits it.
- **The hands.**
  - A hand can run in eight directions only. `│` and `─` give 0°, 90°, 180° and 270°. `/` and `\`, stepping a column a row, give 22.8°, 157.2°, 202.8° and 337.2°, since 20 px of height to 8.4 px of width is 67° from the horizontal.
  - An hour hand for 4 o'clock therefore points between 5 and 6 (`clk-12-d1-1`), and one for 10 o'clock at 11 (`clk-12-d1-5`). At 8:20 both hands point near 6 (`clk-12-d2-5`).
  - The frame says a hand points "only roughly" at the numbers that are not 12, 3, 6 or 9 (R228). At 1, 5, 7 and 11 the nearest direction is 7° off. At 2, 4, 8 and 10 it is 37° off, more than the 30° between two numbers.
- **Reference clock faces.** Of the 33 whose time the question names, 20 point a hand more than 15° off, past half the way to the next number, and 18 by 30° or more. 25 draw a hand with `/` or `\`.
- **Lessons.** The model drew a clock face in all ten tasks of lesson 1, in the style of the reference tasks, never the timetable. Five point a hand at the wrong number:

| Task | Time | Off by |
|---|---|---:|
| `l1-1-asked` | 10:00 | 37° |
| `l1-1-ahead` | 4:00 | 37° |
| `l1-2-ahead` | 8:45 | 60° |
| `l1-4-asked` | 10:30 | 22° |
| `l1-5-ahead` | 9:40 | 47° |

  - In 7 of the ten, the hour hand is a single `/` or `\` one row from `●`. It touches neither the centre nor the minute hand.
  - The model knew the drawing was rough. It added "the hands stand as at 9:30" to two questions.

## 3. Rows leave gaps

A row is 20 px high (`line-height: 20px` at 14 px). How tall each glyph is drawn, by canvas measure:

| Fonts | `│ ║ █` | `/ \` |
|---|---:|---:|
| Noto Sans Symbols on Android | 18 px | 10 px, from Droid Sans Mono |
| DejaVu Sans Mono | 18 px | 13 px |
| Liberation Mono | 17 px | 12 px |

- **Frames.** Every vertical line breaks for 2–3 px at each row, everywhere.
- **Slanted hands.** A slanted hand of several slashes breaks for 7–10 px at each row.

## 4. The model's text

- **Joints.** Of the 43 drawings handed in, one has joints the check reads as broken. It is `l3-4-asked`, whose brace sits right under the bar it counts. No frame, wall or table was misaligned in the text.
- **Copying.** The model copied the frames and the style of the reference tasks closely, which is why the reference clock faces' faults came back in every clock face it drew.
- **Words in drawings.** Equalities in drawings came back too, `T = 15`, `K = ?` and `S + K = 40` (06-drawings, remark 3).

## A probe of a fix

The probe was made in the measurement page only. Each character of the drawing became a span of `display: inline-block; width: 1ch; text-align: center`, and the row was set to 17 px.

- **Columns.** On Android's fonts every column lines up: the grids, the timetable, the Venn drawings, the balance and the boxes. A character wider than a cell overflows into its neighbours, so lines join, and the drawing is no wider than its cells: the Venn drawing is 226.8 px on Android too.
- **Rows.** The frames close and the slashes still do not join.
- **What it costs.**
  - At 17 px, rows of blocks stacked with no line between them merge into one: the bars of the bar-model frame and of the ratio drawings. A frame like that needs an empty line between its rows.
  - The 14 px arrows overlap a neighbour by 2.8 px on each side.
- **Clocks.** The clock face is as wrong as before.

## Options

For the author to decide. Nothing of this is done.

1. **Cells in the card.** `Diagram` lays the drawing out a character to a cell and sets the row to the glyphs' height. This fixes Android and the gaps on every platform, with no font shipped and no drawing changed.
   - The frames and reference drawings that stack rows of blocks get a line between them.
   - The arrows `← →` and `▬` can be narrowed or swapped for `->` and `=`.
   - The site's drawings use the same font stack and would want the same treatment.
2. **Clocks, now: no clock face in drawings.** Drop the clock-face frame and the 34 reference clock faces. The guide then says to keep a time in words or in the timetable frame, and never to draw a face.
3. **Clocks, later: a real clock face on the card.** The model gives the time in a structured field, and the card draws an exact face, as the gallery of this run does. This reopened T70, pictures drawn by the widget: its study is [widget-drawings.md](../widget-drawings.md).
4. **A font of the widget's own.** It would need hosts to admit an inlined font. Their content-security policy for fonts is not known, and option 1 does the same with nothing to load.

The set of characters T58 narrowed (R138) can stay as it is: with option 1, its characters line up on Android too.

## Not measured

- **Real phones.** Neither the author's phone nor any other, so neither Samsung's nor any other maker's own fonts. A screenshot of `just drawings`, drawings 12–15, in Claude on the author's Android phone would confirm the substitution or refute it.
- **iPhone.** SF Mono, which `ui-monospace` picks in WebKit there.
- **ChatGPT's card.**
- **The author's own two drawings.** Their text is in the author's Drive profile and was not read.

## How to repeat

The drivers and the measuring scripts lived in the session's scratch folder, outside the repository. What they do:

```sh
# the service, from a clean copy of the commit, on a port of its own
git archive c1276c3 | tar -x -C src && (cd src && go build -o ../server ./cmd/server)
PORT=8081 MATHTRAIL_LOG_FORMAT=json MATHTRAIL_DEV_AUTH=true MATHTRAIL_DAILY_TASKS=200 MATHTRAIL_DAILY_FAILED=30 \
  MATHTRAIL_SEAL_KEY_CURRENT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')" setsid ./server > server.jsonl &

# each message of a lesson, the first with --session-id and the rest with --resume of the same id
printf '%s' "Здравствуйте! Я родитель … Давайте задачу на часы (time.clocks), уровень 1–2, сложность 2 из 5, и обязательно с картинкой часов." \
  | MATHTRAIL_PLAY_DIR=<run>/play MATHTRAIL_PLAY_MODEL=claude-opus-5-5 just LOCAL_MCP=http://localhost:8081/mcp play -p \
      --output-format stream-json --verbose --effort medium --session-id <uuid> > l1-1.jsonl
```

- **Drawings.** Each drawing is the `task.drawing` of a `submit_task` call in the chat's stream. A call is written ahead when its request came from `prepare_task`.
- **Pictures.** A script imports `served`, `settled` and `addressOf` from `web/scripts/drive.ts`. It runs in the pinned Playwright image as the user, with the three sets of fonts mounted at `~/.local/share/fonts`, and puts each drawing into the card of scene `task`.
- **Android's fonts.** Each comes from `https://android.googlesource.com/platform/<repo>/+/refs/tags/android-16.0.0_r1/<path>?format=TEXT`, base64-decoded:
  - `frameworks/base/data/fonts/DroidSansMono.ttf` and `frameworks/base/data/fonts/fonts.xml`;
  - `external/roboto-fonts/Roboto-Regular.ttf`;
  - `external/noto-fonts/notosanssymbols/NotoSansSymbols-Regular-Subsetted.ttf`.
- **Gallery.** Every drawing of the run as the card shows it on Android and on a computer, with the true clock face beside each clock: https://claude.ai/artifact/UgRdUS2pKc2qTofxahv624 (private to the author).

## The drawings of the run

Android shift is how far a column parts between lines on Android's fonts. Scrolls at 320 is whether the drawing scrolls sideways in a 320 px card on Android. The clock column gives the time drawn and how far its worst hand is off.

| Drawing | Topic, level | Fate | Cells | Android shift, px | Scrolls at 320 | Clock |
|---|---|---|---|---:|---|---|
| `l1-1-asked-1` | time.clocks `1-2` | refused | 18×7 | 0.0 | | 10:00, 37° |
| `l1-1-asked-2` | time.clocks `1-2` | accepted | 18×7 | 0.0 | | 10:00, 37° |
| `l1-1-ahead-1` | time.clocks `1-2` | kept | 18×7 | 0.0 | | 4:00, 37° |
| `l1-2-asked-1` | time.clocks `1-2` | accepted | 18×7 | 0.0 | | 5:30, 8° |
| `l1-2-ahead-1` | time.clocks `1-2` | kept | 18×7 | 6.0 | | 8:45, 60° |
| `l1-3-asked-1` | time.clocks `3-4` | accepted | 18×7 | 6.0 | | 9:30, 15° |
| `l1-3-ahead-1` | time.clocks `3-4` | kept | 18×7 | 0.0 | | 12:00, 0° |
| `l1-4-asked-1` | time.clocks `3-4` | accepted | 18×7 | 0.0 | | 10:30, 22° |
| `l1-4-ahead-1` | time.clocks `3-4` | kept | 18×7 | 0.0 | | 7:00, 7° |
| `l1-5-asked-1` | time.clocks `5-6` | accepted | 18×7 | 3.0 | | 9:00, 0° |
| `l1-5-ahead-1` | time.clocks `5-6` | kept | 18×7 | 0.0 | | 9:40, 47° |
| `l2-1-asked-1` | algorithms.weighing_pouring `3-4` | refused | 13×4 | 10.6 | | |
| `l2-1-asked-2` | algorithms.weighing_pouring `3-4` | accepted | 13×4 | 10.6 | | |
| `l2-1-ahead-1` | algorithms.weighing_pouring `3-4` | kept | 22×5 | 16.6 | | |
| `l2-2-asked-1` | counting.gaps `1-2` | accepted | 21×2 | 10.7 | | |
| `l2-3-asked-1` | logic.ordering `3-4` | accepted | 13×2 | 12.3 | | |
| `l2-3-ahead-1` | logic.ordering `3-4` | kept | 13×2 | 12.3 | | |
| `l2-4-asked-1` | parity.alternation `3-4` | refused | 19×4 | 24.2 | | |
| `l2-4-asked-2` | parity.alternation `3-4` | accepted | 19×4 | 24.2 | | |
| `l2-4-ahead-1` | parity.alternation `3-4` | kept | 18×2 | 12.3 | | |
| `l2-5-asked-1` | combinatorics.enumeration `3-4` | accepted | 13×2 | 0.2 | | |
| `l3-1-asked-1` | geometry.grid `5-6` | refused | 15×8 | 15.1 | | |
| `l3-1-asked-2` | geometry.grid `5-6` | accepted | 15×8 | 15.1 | | |
| `l3-1-ahead-1` | geometry.grid `5-6` | kept | 19×10 | 21.2 | | |
| `l3-2-asked-1` | logic.sets `5-6` | accepted | 27×8 | 37.8 | yes | |
| `l3-2-ahead-1` | logic.sets `5-6` | kept | 27×8 | 37.8 | yes | |
| `l3-3-asked-1` | ratio.sharing `5-6` | accepted | 21×4 | 4.5 | | |
| `l3-3-ahead-1` | ratio.sharing `5-6` | kept | 18×4 | 10.6 | | |
| `l3-4-asked-1` | fractions.parts `5-6` | accepted | 27×2 | 4.5 | | |
| `l3-4-ahead-1` | fractions.parts `5-6` | kept | 27×2 | 1.5 | | |
| `l3-5-asked-1` | games.strategy `5-6` | accepted | 22×1 | 0.0 | | |
| `l3-5-ahead-1` | games.strategy `5-6` | kept | 22×1 | 0.0 | | |
| `l4-1-asked-1` | time.calendar `3-4` | refused | 17×2 | 13.7 | | |
| `l4-1-asked-2` | time.calendar `3-4` | accepted | 17×2 | 13.7 | | |
| `l4-1-ahead-1` | time.calendar `3-4` | kept | 23×2 | 1.5 | | |
| `l4-2-asked-1` | pigeonhole.basic `3-4` | accepted | 19×4 | 25.4 | | |
| `l4-2-ahead-1` | pigeonhole.basic `3-4` | refused | 14×3 | 0.0 | | |
| `l4-2-ahead-2` | pigeonhole.basic `3-4` | kept | 14×3 | 0.0 | | |
| `l4-3-asked-1` | logic.knights_liars `3-4` | accepted | 7×2 | 6.2 | | |
| `l4-3-ahead-1` | logic.knights_liars `3-4` | kept | 7×2 | 6.2 | | |
| `l4-4-asked-1` | number.divisibility `5-6` | refused | 29×3 | 8.2 | yes | |
| `l4-4-asked-2` | number.divisibility `5-6` | accepted | 29×4 | 37.5 | yes | |
| `l4-4-ahead-1` | number.divisibility `5-6` | kept | 15×2 | 10.5 | | |
