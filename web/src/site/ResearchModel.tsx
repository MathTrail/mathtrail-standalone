import type { ComponentChildren } from "preact";
import {
	Commit,
	CommitDate,
	Counted,
	Num,
	type NumberForm,
} from "./ResearchNumbers";
import type { PageReader } from "./reader";
import { type Mark, marks, type Research, type Row } from "./research";

// groups are the groups the board shows its rows in, in its order: the goals
// of the step, those of mastery and those of the screen, and the numbers shown
// for comparison, which have no goal.
const groups = ["step", "mastery", "screen", "context"] as const;

/**
 * barScale is the number a row's bar runs to from nothing: all, for a share;
 * otherwise one, or past one the least of one, two and five times a power of
 * ten that holds the top of the service's interval, the goal and the ceiling,
 * so that a row reads off a round scale.
 */
export function barScale(row: Row): number {
	if (row.unit === "share") {
		return 1;
	}
	const most = Math.max(
		row.values.service.high,
		row.bound?.value ?? 0,
		row.values.ceiling?.high ?? 0,
	);
	if (most <= 1) {
		return 1;
	}
	const power = 10 ** Math.floor(Math.log10(most));
	const step = [1, 2, 5, 10].find((times) => times * power >= most) ?? 10;
	return step * power;
}

/**
 * Place is where a label of a bar stands: on which side of its line, and
 * whether on a second row under the first, out of the other label's way.
 */
export type Place = { readonly side: "start" | "end"; readonly low: boolean };

/**
 * linePlaces are where the labels of a bar's goal and ceiling stand, given
 * where their lines stand on the bar, from nothing to one, either of them
 * missing. A label alone stands after its line, or before it past the bar's
 * middle, where the room is. Two labels turn away from each other, the left
 * one before its line and the right one after its, when each has three tenths
 * of the bar to stand in; otherwise each takes its side as a label alone does,
 * and the ceiling's drops to the second row, out of the goal's way.
 */
export function linePlaces(
	goal: number | undefined,
	ceiling: number | undefined,
): { goal?: Place; ceiling?: Place } {
	const alone = (at: number): Place => ({
		side: at > 0.5 ? "start" : "end",
		low: false,
	});
	if (goal === undefined || ceiling === undefined) {
		return {
			goal: goal === undefined ? undefined : alone(goal),
			ceiling: ceiling === undefined ? undefined : alone(ceiling),
		};
	}
	const [left, right] = goal <= ceiling ? [goal, ceiling] : [ceiling, goal];
	if (left >= 0.3 && right <= 0.7) {
		const before: Place = { side: "start", low: false };
		const after: Place = { side: "end", low: false };
		return goal <= ceiling
			? { goal: before, ceiling: after }
			: { goal: after, ceiling: before };
	}
	return { goal: alone(goal), ceiling: { ...alone(ceiling), low: true } };
}

/**
 * drawnCeiling is the ceiling a row's bar draws: the oracle's, where it stands
 * off nothing. Perfection, and an oracle at nothing, sit where the bar starts
 * and draw no line.
 */
function drawnCeiling(row: Row) {
	const { ceiling } = row.values;
	return ceiling?.of === "oracle" && ceiling.value > 0 ? ceiling : undefined;
}

/**
 * StudentModel is the student model against its goals: what the board shows
 * and the run its numbers come from, with a line on the paper's own; each
 * measure, in its group, with the service's number and its interval drawn as a
 * bar against the goal and the ceiling, and its mark; and what the marks the
 * rows hold mean.
 */
export function StudentModel({
	id,
	page,
	research,
}: {
	id: string;
	page: PageReader;
	research: Research;
}) {
	const { bench, built_from, paper } = research;
	const shown = marks.filter((mark) =>
		bench.rows.some((row) => row.values.service.mark === mark),
	);
	for (const mark of marks.filter((mark) => !shown.includes(mark))) {
		page.leaveOut(`model.marks.${mark}`);
	}
	const drawsCeiling = bench.rows.some(
		(row) => drawnCeiling(row) !== undefined,
	);
	if (!drawsCeiling) {
		page.leaveOut("model.key.ceiling");
		page.leaveOut("model.ceiling");
	}
	for (const [key, better] of [
		["model.goal_at_most", "lower"],
		["model.goal_at_least", "higher"],
	] as const) {
		if (
			!bench.rows.some((row) => row.bound !== null && row.better === better)
		) {
			page.leaveOut(key);
		}
	}
	return (
		<section class="s-wrap s-research-wrap s-research-section" id={id}>
			<h2>{page.text("model.title")}</h2>
			<p class="s-research-intro">{page.text("model.lead")}</p>
			<p class="s-research-run-line">
				<span>
					{page.text("model.run.commit", {
						commit: <Commit hash={built_from.commit} />,
						date: <CommitDate date={built_from.date} />,
					})}
				</span>{" "}
				<span>
					{page.text("model.run.size", {
						children: (
							<Counted value={bench.children} noun="research.children" />
						),
						answers: <Counted value={bench.answers} noun="research.given" />,
					})}
				</span>{" "}
				<span>
					{page.text("model.run.seed", {
						seed: <Num value={bench.seed} form="seed" />,
					})}
				</span>
			</p>
			<p class="s-research-paper-note">
				{page.text("model.paper", { commit: <Commit hash={paper.commit} /> })}
			</p>
			<div class="s-research-board">
				<ul class="s-research-key">
					<li>
						<span class="s-research-swatch s-research-swatch-service" />
						<span>{page.text("model.key.service")}</span>
					</li>
					<li>
						<span class="s-research-swatch s-research-swatch-interval" />
						<span>
							{page.text("model.key.interval", {
								interval: <Num value={bench.interval} form="share" />,
							})}
						</span>
					</li>
					<li>
						<span class="s-research-swatch s-research-swatch-goal" />
						<span>{page.text("model.key.goal")}</span>
					</li>
					{drawsCeiling ? (
						<li>
							<span class="s-research-swatch s-research-swatch-ceiling" />
							<span>{page.text("model.key.ceiling")}</span>
						</li>
					) : null}
				</ul>
				{groups.map((group) => {
					const rows = bench.rows.filter(
						(row) => (row.criterion ?? "context") === group,
					);
					return rows.length === 0 ? null : (
						<Group key={group} page={page} research={research} group={group}>
							{rows.map((row) => (
								<Measure
									key={row.id}
									page={page}
									research={research}
									row={row}
								/>
							))}
						</Group>
					);
				})}
			</div>
			<dl class="s-research-marks">
				{shown.map((mark) => (
					<div key={mark}>
						<dt>
							<MarkPill page={page} mark={mark} />
						</dt>
						<dd>{page.text(`model.marks.${mark}.means`)}</dd>
					</div>
				))}
			</dl>
		</section>
	);
}

// Group is a group of the board's rows under its heading; the screen's says
// which answers its measures are read over and which its goals are drawn from.
function Group({
	page,
	research,
	group,
	children,
}: {
	page: PageReader;
	research: Research;
	group: (typeof groups)[number];
	children: ComponentChildren;
}) {
	const { early, late } = research.bench.screen_windows;
	const headingId = `student-model-${group}`;
	return (
		<section class="s-research-group" aria-labelledby={headingId}>
			<h3 id={headingId}>
				<span>{page.text(`model.groups.${group}`)}</span>
				{group === "screen" ? (
					<span class="s-research-group-note">
						{page.text("model.groups.screen_note", {
							first: <Num value={late.first} />,
							last: <Num value={late.last} />,
							earlyfirst: <Num value={early.first} />,
							earlylast: <Num value={early.last} />,
						})}
					</span>
				) : null}
			</h3>
			<ol class="s-research-measures">{children}</ol>
		</section>
	);
}

// Measure is one row of the board: the measure's name, its unit and its mark;
// its bar; and the service's number with its interval.
function Measure({
	page,
	research,
	row,
}: {
	page: PageReader;
	research: Research;
	row: Row;
}) {
	const form: NumberForm = row.unit;
	const { service } = row.values;
	return (
		<li
			class={
				row.kind === "context"
					? "s-research-measure s-research-context"
					: "s-research-measure"
			}
		>
			<div class="s-research-measure-words">
				<p class="s-research-measure-name">
					{page.text(`model.rows.${row.id}.name`, nameSlots(research))}
				</p>
				<p class="s-research-measure-unit">
					{page.text(`model.rows.${row.id}.unit`, unitSlots(row, form))}
				</p>
				{service.mark === null ? null : (
					<p>
						<MarkPill page={page} mark={service.mark} />
					</p>
				)}
			</div>
			<Bar page={page} row={row} form={form} />
			<p class="s-research-measure-value">
				<span class="s-research-number">
					<Num value={service.value} form={form} />
				</span>
				<span class="s-research-interval">
					{page.text("model.interval", {
						low: <Num value={service.low} form={form} />,
						high: <Num value={service.high} form={form} />,
					})}
				</span>
			</p>
		</li>
	);
}

// nameSlots are what the names of the rows may say of the run and the
// product: the answers the error is read after, and the corridor's ends. A
// name fills the slots it names, and each row is given slots of its own, since
// an element is drawn in one place.
function nameSlots(research: Research): Parameters<PageReader["text"]>[1] {
	const { bench, product } = research;
	return {
		after: <Counted value={bench.error_after} noun="research.answers" />,
		low: <Num value={product.corridor.low} form="share" />,
		high: <Num value={product.corridor.high} form="share" />,
	};
}

// unitSlots are what the unit of a row may say of its numbers: the best any
// rule could do, the oracle's, where the row has it.
function unitSlots(
	row: Row,
	form: NumberForm,
): Parameters<PageReader["text"]>[1] {
	const ceiling = row.values.ceiling;
	return ceiling?.of === "oracle"
		? { best: <Num value={ceiling.value} form={form} /> }
		: {};
}

// MarkPill is a mark in the colours of its kind.
function MarkPill({ page, mark }: { page: PageReader; mark: Mark }) {
	return (
		<span class={`s-research-mark s-research-${mark}`}>
			{page.text(`model.marks.${mark}.name`)}
		</span>
	);
}

// Bar draws a row on a scale from nothing: the service's number as a fill,
// its interval over it, the goal as a dashed line and the oracle's ceiling,
// where it stands off nothing, as a dotted one, each line with its label. The
// fill and the lines are hidden from a screen reader, which reads the labels
// and the number beside the bar.
function Bar({
	page,
	row,
	form,
}: {
	page: PageReader;
	row: Row;
	form: NumberForm;
}) {
	const scale = barScale(row);
	const share = (value: number) => Math.min(1, Math.max(0, value / scale));
	const at = (value: number) => `${share(value) * 100}%`;
	const { service } = row.values;
	const oracle = drawnCeiling(row);
	const places = linePlaces(
		row.bound === null ? undefined : share(row.bound.value),
		oracle === undefined ? undefined : share(oracle.value),
	);
	return (
		<div
			class={
				places.ceiling?.low
					? "s-research-bar s-research-bar-two"
					: "s-research-bar"
			}
			dir="ltr"
		>
			{row.bound === null || places.goal === undefined ? null : (
				<Line kind="goal" at={at(row.bound.value)} place={places.goal}>
					{page.text(
						row.better === "lower"
							? "model.goal_at_most"
							: "model.goal_at_least",
						{ bound: <Num value={row.bound.value} form={form} /> },
					)}
				</Line>
			)}
			{oracle === undefined || places.ceiling === undefined ? null : (
				<Line kind="ceiling" at={at(oracle.value)} place={places.ceiling}>
					{page.text("model.ceiling", {
						value: <Num value={oracle.value} form={form} />,
					})}
				</Line>
			)}
			<span class="s-research-track" aria-hidden="true">
				<span
					class="s-research-fill"
					style={{ inlineSize: at(service.value) }}
				/>
				<span
					class="s-research-band"
					style={{
						insetInlineStart: at(service.low),
						inlineSize: `${(share(service.high) - share(service.low)) * 100}%`,
					}}
				/>
			</span>
		</div>
	);
}

// Line is the goal's or the ceiling's line across a bar, at its place, with
// its label where it is placed.
function Line({
	kind,
	at,
	place,
	children,
}: {
	kind: "goal" | "ceiling";
	at: string;
	place: Place;
	children: ComponentChildren;
}) {
	return (
		<>
			<span
				class={`s-research-line s-research-line-${kind}`}
				style={{ insetInlineStart: at }}
				aria-hidden="true"
			/>
			<span
				class={`s-research-line-label s-research-line-${kind} s-research-side-${place.side}${place.low ? " s-research-line-low" : ""}`}
				style={{ insetInlineStart: at }}
			>
				{children}
			</span>
		</>
	);
}
