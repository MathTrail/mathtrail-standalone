import type { ComponentChildren } from "preact";
import { Counted, Given, Month, Num } from "./ResearchNumbers";
import type { PageReader } from "./reader";
import type {
	AnswersCell,
	ChanceCell,
	Live,
	LiveShown,
	Research,
} from "./research";

/** Corridor is the corridor of chances the rule chooses tasks in. */
type Corridor = Research["product"]["corridor"];

// measures are the three measures of the block, in its order: the answers
// against the chance promised, how far they came out from it by the range of
// the child's answers, and the share right over every range.
const measures = ["chances", "kept_up", "share"] as const;

// titleOf is the id of a measure's heading, which names its table.
const titleOf = (measure: (typeof measures)[number]) =>
	`live-${measure.replace("_", "-")}`;

// clamped is x held between low and high, so that a drawing stays inside its
// frame whatever its numbers are.
const clamped = (x: number, low = 0, high = 1) =>
	Math.min(high, Math.max(low, x));

/**
 * LiveNumbers is the block of the live numbers: whether the model's promises
 * come true for real children, in the latest month counted whole, each of its
 * three measures in a column of its own. Before a month is counted whole, and
 * for a month of too few children, the columns name the measures and the
 * block says when their numbers come; for a month shown, each column draws its
 * measure, beside a table a screen reader reads in place of the drawing. In
 * every state it says the rule a range is shown by.
 */
export function LiveNumbers({
	id,
	page,
	research,
}: {
	id: string;
	page: PageReader;
	research: Research;
}) {
	const { live, product } = research;
	return (
		<section class="s-wrap s-research-wrap s-research-live" id={id}>
			<h2>{page.text("live.title")}</h2>
			<p class="s-research-intro">{page.text("live.lead")}</p>
			{live.state === "ready" ? (
				<MonthShown page={page} live={live} product={product} />
			) : (
				<Frame page={page} live={live} corridor={product.corridor} />
			)}
			<p class="s-research-note">
				{page.text("live.rule", {
					children: (
						<Counted value={live.rule.learners} noun="research.children" />
					),
					answers: <Counted value={live.rule.answers} noun="research.given" />,
					rounded: <Num value={live.rule.rounded_to} />,
				})}
			</p>
		</section>
	);
}

// Column is one measure's column, under its title.
function Column({
	page,
	measure,
	children,
}: {
	page: PageReader;
	measure: (typeof measures)[number];
	children?: ComponentChildren;
}) {
	return (
		<div class="s-research-live-column">
			<h3 id={titleOf(measure)}>{page.text(`live.measures.${measure}`)}</h3>
			{children}
		</div>
	);
}

// Frame is what the block shows while it shows no numbers: a badge, the three
// measures to come, each drawn empty, at nothing, and when their numbers
// come, or that the latest month counted whole had too few children for any
// of them.
function Frame({
	page,
	live,
	corridor,
}: {
	page: PageReader;
	live: Exclude<Live, LiveShown>;
	corridor: Corridor;
}) {
	page.leaveOut("live.ready");
	page.leaveOut(
		live.state === "coming" ? "live.frame.too_few" : "live.frame.coming",
	);
	const at = `live.frame.${live.state}`;
	return (
		<>
			<p class="s-research-soon">{page.text(`${at}.badge`)}</p>
			<div class="s-research-live-columns">
				<Column page={page} measure="chances">
					<Plot page={page} cells={[]} corridor={corridor} />
				</Column>
				<Column page={page} measure="kept_up">
					<div class="s-research-margins" dir="ltr" aria-hidden="true">
						<MarginScale scale={0.1} />
						<ol>
							{["first", "second", "third", "fourth"].map((range) => (
								<li key={range}>
									<span class="s-research-margins-range s-research-skeleton">
										<span />
										<span />
									</span>
									<span class="s-research-margin" />
									<span class="s-research-margins-value">
										<Given value={0} form="signed" />
									</span>
								</li>
							))}
						</ol>
					</div>
				</Column>
				<Column page={page} measure="share">
					<ShareDrawing
						page={page}
						corridor={corridor}
						share={0}
						value={<Given value={0} form="chance" />}
					/>
				</Column>
			</div>
			<p class="s-research-note">
				{page.text(
					`${at}.when`,
					live.month === null ? {} : { month: <Month month={live.month} /> },
				)}
			</p>
		</>
	);
}

// MonthShown is the numbers of a month: how many children and answers stand
// behind them, its three measures, which answers are weighed, and what live
// answers cannot measure.
function MonthShown({
	page,
	live,
	product,
}: {
	page: PageReader;
	live: LiveShown;
	product: Research["product"];
}) {
	page.leaveOut("live.frame");
	return (
		<>
			<p class="s-research-month">
				{page.text("live.ready.month", {
					month: <Month month={live.month} />,
					children: (
						<Counted value={live.total.learners} noun="research.children" />
					),
					answers: <Counted value={live.total.answers} noun="research.given" />,
				})}
			</p>
			<div class="s-research-live-columns">
				<Column page={page} measure="chances">
					<Chances
						page={page}
						cells={live.chances}
						corridor={product.corridor}
					/>
				</Column>
				<Column page={page} measure="kept_up">
					<KeptUp page={page} cells={live.kept_up} />
				</Column>
				<Column page={page} measure="share">
					<Share page={page} live={live} corridor={product.corridor} />
				</Column>
			</div>
			<p class="s-research-note">
				{page.text("live.ready.weighed", {
					trial: (
						<Counted value={product.trial_answers} noun="research.trial" />
					),
				})}
			</p>
			<p class="s-research-note">{page.text("live.ready.corridor")}</p>
		</>
	);
}

// NoRange says a measure's month shows none of its ranges, and leaves out the
// words of the ranges it would have shown.
function NoRange({ page, at }: { page: PageReader; at: string }) {
	page.leaveOut(`${at}.shown`);
	return <p class="s-research-caption">{page.text(`${at}.none`)}</p>;
}

// Chances is the answers against the chance promised: a point for each range
// of that chance, and the table of the ranges.
function Chances({
	page,
	cells,
	corridor,
}: {
	page: PageReader;
	cells: readonly ChanceCell[];
	corridor: Corridor;
}) {
	const at = "live.ready.chances";
	if (cells.length === 0) {
		return <NoRange page={page} at={at} />;
	}
	page.leaveOut(`${at}.none`);
	return (
		<>
			<Plot page={page} cells={cells} corridor={corridor} />
			<p class="s-research-caption">{page.text(`${at}.shown.caption`)}</p>
			<div class="s-hidden">
				<table aria-labelledby={titleOf("chances")}>
					<thead>
						<tr>
							<th scope="col">{page.text(`${at}.shown.heading`)}</th>
							<Columns
								page={page}
								keys={["promised", "came_true", "answers", "children"]}
							/>
						</tr>
					</thead>
					<tbody>
						{cells.map((cell) => (
							<tr key={cell.from}>
								<th scope="row">
									{page.text(`${at}.shown.range`, {
										from: <Num value={cell.from} form="hundredths" />,
										to: <Num value={cell.to} form="hundredths" />,
									})}
								</th>
								<td>
									<Num value={cell.promised_mean} form="chance" />
								</td>
								<td>
									<Num value={cell.correct_share} form="chance" />
								</td>
								<td>
									<Num value={cell.answers} />
								</td>
								<td>
									<Num value={cell.learners} />
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</>
	);
}

// Columns is the headings of the columns a measure's table shares with the
// others, under the keys given.
function Columns({
	page,
	keys,
}: {
	page: PageReader;
	keys: readonly ("promised" | "came_true" | "answers" | "children")[];
}) {
	return (
		<>
			{keys.map((key) => (
				<th key={key} scope="col">
					{page.text(`live.ready.columns.${key}`)}
				</th>
			))}
		</>
	);
}

// Plot draws a point for each range of chance: across, the chance it promised
// on average, and up, the share answered right, the larger the more answers
// it holds; over the dashed diagonal where a promise comes true and the band
// of the corridor. Both axes run from four tenths, or from the lowest tenth
// it draws when that is lower, to certainty, and share the corner where they
// start, which the upright axis names. It is hidden from a screen reader,
// which reads the table under it.
function Plot({
	page,
	cells,
	corridor,
}: {
	page: PageReader;
	cells: readonly ChanceCell[];
	corridor: Corridor;
}) {
	const lowest = Math.min(
		...cells.flatMap((cell) => [cell.promised_mean, cell.correct_share]),
	);
	const low = Math.min(0.4, Math.floor(10 * lowest) / 10);
	const at = (x: number) => clamped((x - low) / (1 - low)) * 100;
	const most = Math.max(...cells.map((cell) => cell.answers)) || 1;
	const words = "live.ready.chances.shown";
	return (
		<div class="s-research-plot-frame" dir="ltr" aria-hidden="true">
			<p class="s-research-plot-up">{page.text(`${words}.up`)}</p>
			<div class="s-research-plot">
				<span
					class="s-research-plot-band"
					style={{
						insetInlineStart: `${at(corridor.low)}%`,
						inlineSize: `${at(corridor.high) - at(corridor.low)}%`,
					}}
				>
					{page.text(`${words}.band`)}
				</span>
				<svg
					viewBox="0 0 100 100"
					preserveAspectRatio="none"
					aria-hidden="true"
				>
					<line
						class="s-research-plot-diagonal"
						x1={0}
						y1={100}
						x2={100}
						y2={0}
						vector-effect="non-scaling-stroke"
					/>
					{runsOf(cells)
						.filter((run) => run.length > 1)
						.map((run) => (
							<polyline
								key={run[0]?.from}
								class="s-research-plot-path"
								points={run
									.map(
										(cell) =>
											`${at(cell.promised_mean)},${100 - at(cell.correct_share)}`,
									)
									.join(" ")}
								vector-effect="non-scaling-stroke"
							/>
						))}
				</svg>
				{cells.map((cell) => (
					<span
						key={cell.from}
						class="s-research-plot-point"
						style={{
							insetInlineStart: `${at(cell.promised_mean)}%`,
							insetBlockEnd: `${at(cell.correct_share)}%`,
							"--size": `${10 + 10 * Math.sqrt(clamped(cell.answers / most))}px`,
						}}
					/>
				))}
				<span class="s-research-plot-ideal">{page.text(`${words}.ideal`)}</span>
				<span class="s-research-y s-research-y-low">
					<Given value={low} form="chance" />
				</span>
				<span class="s-research-y s-research-y-high">
					<Given value={1} form="chance" />
				</span>
				<span class="s-research-x-high">
					<Given value={1} form="chance" />
				</span>
			</div>
			<p class="s-research-plot-across">{page.text(`${words}.across`)}</p>
		</div>
	);
}

/**
 * runsOf are the ranges of chance a month shows, in runs of ranges that meet:
 * a range is named by the hundredths it spans, both ends included, so the next
 * begins a hundredth past the end of the one before. The plot joins the points
 * of a run, and never across a range the month leaves out.
 */
export function runsOf(cells: readonly ChanceCell[]): ChanceCell[][] {
	const hundredths = (chance: number) => Math.round(chance * 100);
	const runs: ChanceCell[][] = [];
	for (const cell of cells) {
		const run = runs.at(-1);
		const last = run?.at(-1);
		if (
			run !== undefined &&
			last !== undefined &&
			hundredths(cell.from) === hundredths(last.to) + 1
		) {
			run.push(cell);
		} else {
			runs.push([cell]);
		}
	}
	return runs;
}

// KeptUp is how far the answers came out from the chance promised, by the
// range of the child's answers: a row for each range, its margin drawn as a
// bar from nothing with a whisker of its standard error either way, and the
// table of the ranges.
function KeptUp({
	page,
	cells,
}: {
	page: PageReader;
	cells: readonly AnswersCell[];
}) {
	const at = "live.ready.kept_up";
	if (cells.length === 0) {
		return <NoRange page={page} at={at} />;
	}
	page.leaveOut(`${at}.none`);
	if (cells.every((cell) => cell.last !== null)) {
		page.leaveOut(`${at}.shown.open`);
	}
	if (cells.every((cell) => cell.last === null)) {
		page.leaveOut(`${at}.shown.range`);
	}
	const scale = Math.max(
		0.1,
		...cells.map(
			(cell) => Math.abs(cell.came_true_less_promised) + cell.standard_error,
		),
	);
	const rangeOf = (cell: AnswersCell) =>
		cell.last === null
			? page.text(`${at}.shown.open`, { first: <Num value={cell.first} /> })
			: page.text(`${at}.shown.range`, {
					first: <Num value={cell.first} />,
					last: <Num value={cell.last} />,
				});
	return (
		<>
			<div class="s-research-margins" dir="ltr" aria-hidden="true">
				<MarginScale scale={scale} />
				<ol>
					{cells.map((cell) => (
						<li key={cell.first}>
							<span class="s-research-margins-range">
								<span>{rangeOf(cell)}</span>
								<span>
									<Counted value={cell.answers} noun="research.answers" />
								</span>
							</span>
							<MarginBar cell={cell} scale={scale} />
							<span class="s-research-margins-value">
								<Num value={cell.came_true_less_promised} form="signed" />
							</span>
						</li>
					))}
				</ol>
			</div>
			<p class="s-research-caption">{page.text(`${at}.shown.caption`)}</p>
			<div class="s-hidden">
				<table aria-labelledby={titleOf("kept_up")}>
					<thead>
						<tr>
							<th scope="col">{page.text(`${at}.shown.heading`)}</th>
							<th scope="col">{page.text(`${at}.shown.margin`)}</th>
							<th scope="col">{page.text(`${at}.shown.error`)}</th>
							<Columns page={page} keys={["answers", "children"]} />
						</tr>
					</thead>
					<tbody>
						{cells.map((cell) => (
							<tr key={cell.first}>
								<th scope="row">{rangeOf(cell)}</th>
								<td>
									<Num value={cell.came_true_less_promised} form="signed" />
								</td>
								<td>
									<Num value={cell.standard_error} form="chance" />
								</td>
								<td>
									<Num value={cell.answers} />
								</td>
								<td>
									<Num value={cell.learners} />
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</>
	);
}

// MarginScale names the ends and the middle of the scale the margins are
// drawn on, from minus scale to scale.
function MarginScale({ scale }: { scale: number }) {
	return (
		<p class="s-research-margins-scale">
			<Given value={-scale} form="signed" />
			<Given value={0} form="signed" />
			<Given value={scale} form="signed" />
		</p>
	);
}

// MarginBar draws a range's margin as a bar from nothing, which stands in the
// middle as a line: to the right where the answers came out better than
// promised, to the left where worse, with a whisker of its standard error
// either way, on a scale from minus scale to scale.
function MarginBar({ cell, scale }: { cell: AnswersCell; scale: number }) {
	const at = (x: number) => 50 + 50 * clamped(x / scale, -1, 1);
	const margin = at(cell.came_true_less_promised);
	const low = at(cell.came_true_less_promised - cell.standard_error);
	const high = at(cell.came_true_less_promised + cell.standard_error);
	return (
		<span class="s-research-margin">
			<span
				class="s-research-margin-fill"
				style={{
					insetInlineStart: `${Math.min(50, margin)}%`,
					inlineSize: `${Math.abs(margin - 50)}%`,
				}}
			/>
			<span
				class="s-research-margin-whisker"
				style={{
					insetInlineStart: `${Math.min(low, high)}%`,
					inlineSize: `${Math.abs(high - low)}%`,
				}}
			/>
		</span>
	);
}

// ShareDrawing draws a share right, value written large, beside what it is
// expected to be, the corridor's middle; and as a point at share on a scale
// from nothing to certain, over the corridor and a dashed line at its middle.
// It is hidden from a screen reader, which reads the table beside it.
function ShareDrawing({
	page,
	corridor,
	share,
	value,
}: {
	page: PageReader;
	corridor: Corridor;
	share: number;
	value: ComponentChildren;
}) {
	const at = (x: number) => `${clamped(x) * 100}%`;
	return (
		<div class="s-research-share" dir="ltr" aria-hidden="true">
			<p class="s-research-share-head">
				<span class="s-research-share-value">{value}</span>
				<span>
					{page.text("live.ready.share.expected", {
						middle: <Num value={corridor.middle} form="chance" />,
					})}
				</span>
			</p>
			<span class="s-research-share-track">
				<span
					class="s-research-share-corridor"
					style={{
						insetInlineStart: at(corridor.low),
						inlineSize: at(corridor.high - corridor.low),
					}}
				/>
				<span
					class="s-research-share-middle"
					style={{ insetInlineStart: at(corridor.middle) }}
				/>
				<span
					class="s-research-share-point"
					style={{ insetInlineStart: at(share) }}
				/>
			</span>
			<p class="s-research-share-ends">
				<Given value={0} />
				<Given value={1} />
			</p>
		</div>
	);
}

// Share is the share answered right over every range of chance, large, and
// what it is expected to be, the corridor's middle; drawn as a point on a
// scale from nothing to certain, over the corridor and a dashed line at its
// middle; and the table of the month's numbers.
function Share({
	page,
	live,
	corridor,
}: {
	page: PageReader;
	live: LiveShown;
	corridor: Corridor;
}) {
	const { total } = live;
	return (
		<>
			<ShareDrawing
				page={page}
				corridor={corridor}
				share={total.correct_share}
				value={<Num value={total.correct_share} form="chance" />}
			/>
			<p class="s-research-caption">
				{page.text("live.ready.share.caption", {
					middle: <Num value={corridor.middle} form="chance" />,
					low: <Num value={corridor.low} form="chance" />,
					high: <Num value={corridor.high} form="chance" />,
				})}
			</p>
			<div class="s-hidden">
				<table aria-labelledby={titleOf("share")}>
					<thead>
						<tr>
							<Columns
								page={page}
								keys={["came_true", "promised", "answers", "children"]}
							/>
						</tr>
					</thead>
					<tbody>
						<tr>
							<td>
								<Num value={total.correct_share} form="chance" />
							</td>
							<td>
								<Num value={total.promised_mean} form="chance" />
							</td>
							<td>
								<Num value={total.answers} />
							</td>
							<td>
								<Num value={total.learners} />
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</>
	);
}
