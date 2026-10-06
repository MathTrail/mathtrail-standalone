import { Counted, Given, Month, Num } from "./ResearchNumbers";
import type { PageReader } from "./reader";
import type {
	AnswersCell,
	ChanceCell,
	Live,
	LiveShown,
	Research,
} from "./research";
import { TableFrame } from "./TableFrame";

/** Corridor is the corridor of chances the rule chooses tasks in. */
type Corridor = Research["product"]["corridor"];

// The ids of the measures' headings, which name their tables.
const chancesTitle = "live-chances";
const keptUpTitle = "live-kept-up";
const shareTitle = "live-share";

// clamped is x held between low and high, so that a drawing stays inside its
// frame whatever its numbers are.
const clamped = (x: number, low = 0, high = 1) =>
	Math.min(high, Math.max(low, x));

/**
 * LiveNumbers is the block of the live numbers: whether the model's promises
 * come true for real children, in the latest month counted whole. Before a
 * month is counted whole, and for a month of too few children, it says what
 * it will show; for a month shown, it draws each measure beside its table. In
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
		<section class="s-wrap s-section" id={id}>
			<div class="s-intro">
				<h2>{page.text("live.title")}</h2>
				<p class="s-intro-line">{page.text("live.lead")}</p>
				{live.state === "ready" ? (
					<p class="s-intro-line">
						{page.text("live.ready.month", {
							month: <Month month={live.month} />,
							children: (
								<Counted value={live.total.learners} noun="research.children" />
							),
							answers: (
								<Counted value={live.total.answers} noun="research.given" />
							),
						})}
					</p>
				) : null}
			</div>
			{live.state === "ready" ? (
				<MonthShown page={page} live={live} product={product} />
			) : (
				<Frame page={page} live={live} />
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

// Frame is what the block says while it shows no numbers: what it will show,
// and when they come, or that the latest month counted whole had too few
// children for any of them.
function Frame({
	page,
	live,
}: {
	page: PageReader;
	live: Exclude<Live, LiveShown>;
}) {
	page.leaveOut("live.ready");
	page.leaveOut(
		live.state === "coming" ? "live.frame.too_few" : "live.frame.coming",
	);
	const at = `live.frame.${live.state}`;
	return (
		<div class="s-panel s-research-live">
			<p class="s-badge">{page.text(`${at}.badge`)}</p>
			<ul>
				{page.list("live.frame.measures").map((key) => (
					<li key={key}>{page.text(key)}</li>
				))}
			</ul>
			<p>
				{page.text(
					`${at}.when`,
					live.month === null ? {} : { month: <Month month={live.month} /> },
				)}
			</p>
		</div>
	);
}

// MonthShown is the numbers of a month: its answers against the chance
// promised, by the range of that chance; how far they came out from it, by
// the range of the child's answers; and the share right over every range;
// with which answers are weighed, and what live answers cannot measure.
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
		<div class="s-research-live">
			<Chances page={page} cells={live.chances} corridor={product.corridor} />
			<KeptUp page={page} cells={live.kept_up} />
			<Share page={page} live={live} corridor={product.corridor} />
			<p class="s-research-note">
				{page.text("live.ready.weighed", {
					trial: (
						<Counted value={product.trial_answers} noun="research.trial" />
					),
				})}
			</p>
			<p class="s-research-note">{page.text("live.ready.corridor")}</p>
		</div>
	);
}

// NoRange says a measure's month shows none of its ranges, and leaves out the
// words of the ranges it would have shown.
function NoRange({ page, at }: { page: PageReader; at: string }) {
	page.leaveOut(`${at}.shown`);
	return <p class="s-research-caption">{page.text(`${at}.none`)}</p>;
}

// Chances is the answers against the chance promised: a point for each range
// of that chance, beside the table of the ranges.
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
	return (
		<div>
			<h3 id={chancesTitle}>{page.text(`${at}.title`)}</h3>
			{cells.length === 0 ? (
				<NoRange page={page} at={at} />
			) : (
				<ChancesShown page={page} cells={cells} corridor={corridor} />
			)}
		</div>
	);
}

// ChancesShown is the ranges of chance a month shows, drawn and listed.
function ChancesShown({
	page,
	cells,
	corridor,
}: {
	page: PageReader;
	cells: readonly ChanceCell[];
	corridor: Corridor;
}) {
	const at = "live.ready.chances";
	page.leaveOut(`${at}.none`);
	return (
		<>
			<p class="s-research-caption">{page.text(`${at}.shown.caption`)}</p>
			<div class="s-research-figure">
				<div>
					<Plot cells={cells} corridor={corridor} />
					<p class="s-research-caption">{page.text(`${at}.shown.axes`)}</p>
				</div>
				<TableFrame labelledBy={chancesTitle}>
					<table class="s-table">
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
				</TableFrame>
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
// which reads the table beside it.
function Plot({
	cells,
	corridor,
}: {
	cells: readonly ChanceCell[];
	corridor: Corridor;
}) {
	const lowest = Math.min(
		...cells.flatMap((cell) => [cell.promised_mean, cell.correct_share]),
	);
	const low = Math.min(0.4, Math.floor(10 * lowest) / 10);
	const at = (x: number) => clamped((x - low) / (1 - low)) * 100;
	const most = Math.max(...cells.map((cell) => cell.answers)) || 1;
	return (
		<div class="s-research-plot" dir="ltr" aria-hidden="true">
			<svg viewBox="0 0 100 100" width="100%" height="100%" aria-hidden="true">
				<rect
					x={at(corridor.low)}
					y={0}
					width={at(corridor.high) - at(corridor.low)}
					height={100}
					fill="currentColor"
					fill-opacity={0.12}
				/>
				<line
					x1={0}
					y1={100}
					x2={100}
					y2={0}
					stroke="currentColor"
					stroke-opacity={0.6}
					stroke-dasharray="4 3"
					vector-effect="non-scaling-stroke"
				/>
				{cells.map((cell) => (
					<circle
						key={cell.from}
						cx={at(cell.promised_mean)}
						cy={100 - at(cell.correct_share)}
						r={1.5 + 2.5 * Math.sqrt(clamped(cell.answers / most))}
						fill="currentColor"
					/>
				))}
			</svg>
			<span class="s-research-tick" style={{ insetInlineStart: "100%" }}>
				<Given value={1} form="chance" />
			</span>
			<span class="s-research-y" style={{ insetBlockStart: "100%" }}>
				<Given value={low} form="chance" />
			</span>
			<span class="s-research-y" style={{ insetBlockStart: "0%" }}>
				<Given value={1} form="chance" />
			</span>
		</div>
	);
}

// KeptUp is how far the answers came out from the chance promised, by the
// range of the child's answers: a table whose every row draws its margin as a
// bar from nothing, with a whisker of its standard error either way.
function KeptUp({
	page,
	cells,
}: {
	page: PageReader;
	cells: readonly AnswersCell[];
}) {
	const at = "live.ready.kept_up";
	return (
		<div>
			<h3 id={keptUpTitle}>{page.text(`${at}.title`)}</h3>
			{cells.length === 0 ? (
				<NoRange page={page} at={at} />
			) : (
				<KeptUpShown page={page} cells={cells} />
			)}
		</div>
	);
}

// KeptUpShown is the ranges of the child's answers a month shows, each named
// by its first answer and its last, or by its first alone for the range of
// every answer past the others.
function KeptUpShown({
	page,
	cells,
}: {
	page: PageReader;
	cells: readonly AnswersCell[];
}) {
	const at = "live.ready.kept_up";
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
	return (
		<>
			<p class="s-research-caption">{page.text(`${at}.shown.caption`)}</p>
			<TableFrame labelledBy={keptUpTitle}>
				<table class="s-table">
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
								<th scope="row">
									{cell.last === null
										? page.text(`${at}.shown.open`, {
												first: <Num value={cell.first} />,
											})
										: page.text(`${at}.shown.range`, {
												first: <Num value={cell.first} />,
												last: <Num value={cell.last} />,
											})}
									<MarginBar cell={cell} scale={scale} />
								</th>
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
			</TableFrame>
		</>
	);
}

// MarginBar draws a range's margin as a bar from nothing, which stands in the
// middle as a dashed line: to the right where the answers came out better
// than promised, to the left where worse, with a whisker of its standard
// error either way, on a scale from minus scale to scale. It is hidden from a
// screen reader, which reads the row's cells.
function MarginBar({ cell, scale }: { cell: AnswersCell; scale: number }) {
	const at = (x: number) => 50 + 50 * clamped(x / scale, -1, 1);
	const margin = at(cell.came_true_less_promised);
	const low = at(cell.came_true_less_promised - cell.standard_error);
	const high = at(cell.came_true_less_promised + cell.standard_error);
	return (
		<span class="s-research-chart" dir="ltr" aria-hidden="true">
			<span class="s-research-bar s-research-bar-service">
				<span
					class="s-research-fill"
					style={{
						insetInlineStart: `${Math.min(50, margin)}%`,
						inlineSize: `${Math.abs(margin - 50)}%`,
					}}
				/>
				<span
					class="s-research-whisker"
					style={{
						insetInlineStart: `${Math.min(low, high)}%`,
						inlineSize: `${Math.abs(high - low)}%`,
					}}
				/>
			</span>
			<span class="s-research-bound" style={{ insetInlineStart: "50%" }} />
		</span>
	);
}

// Share is the share answered right over every range of chance: one bar over
// the corridor, with a dashed tick at the chance promised on average, beside
// its row of numbers.
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
	const at = (x: number) => `${clamped(x) * 100}%`;
	return (
		<div>
			<h3 id={shareTitle}>{page.text("live.ready.share.title")}</h3>
			<p class="s-research-caption">
				{page.text("live.ready.share.caption", {
					middle: <Num value={corridor.middle} form="chance" />,
					low: <Num value={corridor.low} form="chance" />,
					high: <Num value={corridor.high} form="chance" />,
				})}
			</p>
			<div class="s-research-figure">
				<div class="s-research-axis" dir="ltr" aria-hidden="true">
					<span
						class="s-research-band"
						style={{
							insetInlineStart: at(corridor.low),
							inlineSize: at(corridor.high - corridor.low),
						}}
					/>
					<span
						class="s-research-fill"
						style={{ insetBlock: "10px", inlineSize: at(total.correct_share) }}
					/>
					<span
						class="s-research-bound"
						style={{ insetInlineStart: at(total.promised_mean) }}
					/>
					<span class="s-research-tick" style={{ insetInlineStart: "0%" }}>
						<Given value={0} form="chance" />
					</span>
					<span class="s-research-tick" style={{ insetInlineStart: "100%" }}>
						<Given value={1} form="chance" />
					</span>
				</div>
				<TableFrame labelledBy={shareTitle}>
					<table class="s-table">
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
				</TableFrame>
			</div>
		</div>
	);
}
