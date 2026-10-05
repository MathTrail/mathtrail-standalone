import {
	Commit,
	CommitDate,
	Counted,
	Num,
	type NumberForm,
} from "./ResearchNumbers";
import type { PageReader } from "./reader";
import {
	ceilingKinds,
	marks,
	type Research,
	type ResearchFile,
	type Row,
} from "./research";
import { TableFrame } from "./TableFrame";

/** Bench is what the page says of the bench's run. */
type Bench = ResearchFile["bench"];

/** Value is a rule's number of a row, with its interval and its mark. */
type Value = Row["values"]["service"] | Row["values"]["earlier"];

// groups are the groups the table shows its rows in, in its order: the goals
// of the step, those of mastery and those of the screen, and the numbers shown
// for comparison, which have no goal.
const groups = ["step", "mastery", "screen", "context"] as const;

// titleId is the id of the table's heading, which names the table.
const titleId = "student-model-title";

/**
 * chartScale is the number a row's drawing is scaled to: the largest of what
 * it draws, the ends of the intervals, the bound and the ceiling, or one when
 * every one of them is nothing.
 */
function chartScale(row: Row): number {
	const { service, earlier, ceiling } = row.values;
	return (
		Math.max(
			service.high,
			earlier.high,
			row.bound?.value ?? 0,
			ceiling?.high ?? 0,
		) || 1
	);
}

/**
 * StudentModel is the student model against its goals: every row of the
 * bench's table, the rule before and the service as it runs, each number with
 * its interval and its mark, the goal and the ceiling; what the marks and the
 * ceilings mean; and where the numbers come from.
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
	const { bench } = research;
	return (
		<section class="s-wrap s-section" id={id}>
			<div class="s-intro">
				<h2 id={titleId}>{page.text("model.title")}</h2>
				<p class="s-intro-line">{page.text("model.lead")}</p>
			</div>
			<TableFrame labelledBy={titleId}>
				<table class="s-table s-research-goals">
					<thead>
						<tr>
							<th scope="col">{page.text("model.columns.measure")}</th>
							<th scope="col">{page.text("model.columns.earlier")}</th>
							<th scope="col">{page.text("model.columns.service")}</th>
							<th scope="col">{page.text("model.columns.goal")}</th>
							<th scope="col">{page.text("model.columns.ceiling")}</th>
						</tr>
					</thead>
					{groups.map((group) => {
						const rows = bench.rows.filter(
							(row) => (row.criterion ?? "context") === group,
						);
						return rows.length === 0 ? null : (
							<tbody key={group}>
								<tr>
									<th scope="rowgroup" colSpan={5}>
										{page.text(`model.groups.${group}`)}
									</th>
								</tr>
								{rows.map((row) => (
									<GoalRow key={row.id} page={page} row={row} bench={bench} />
								))}
							</tbody>
						);
					})}
				</table>
			</TableFrame>
			<p class="s-research-caption">{page.text("model.earlier")}</p>
			<Legend page={page} />
			<Provenance page={page} research={research} />
		</section>
	);
}

// GoalRow is one row of the table: the measure, the rule before and the
// service, the goal, and the ceiling.
function GoalRow({
	page,
	row,
	bench,
}: {
	page: PageReader;
	row: Row;
	bench: Bench;
}) {
	const form: NumberForm = row.unit;
	const slots = labelSlots(bench);
	return (
		<tr>
			<th scope="row">
				<span class="s-research-measure">
					{page.text(`model.rows.${row.id}.name`, slots)}
				</span>
				<span class="s-research-unit">
					{page.text(`model.rows.${row.id}.unit`)}
				</span>
				<Chart row={row} />
			</th>
			<td>
				<RuleValue page={page} value={row.values.earlier} form={form} />
			</td>
			<td>
				<RuleValue page={page} value={row.values.service} form={form} />
			</td>
			<td>
				<Goal page={page} row={row} form={form} slots={slots} />
			</td>
			<td>
				<Ceiling page={page} row={row} form={form} />
			</td>
		</tr>
	);
}

// labelSlots are what the names and the goals of the rows may say of the run:
// the answers the error is read after, the windows of the screen, and the
// numbers the goals are drawn with. A text fills the slots it names, and each
// row is given slots of its own, since an element is drawn in one place.
function labelSlots(bench: Bench): Parameters<PageReader["text"]>[1] {
	const { early, late } = bench.screen_windows;
	const { lag_share, corridor_share, late_times } = bench.goal_parameters;
	return {
		after: <Counted value={bench.error_after} noun="research.answers" />,
		first: <Num value={late.first} />,
		last: <Num value={late.last} />,
		earlyfirst: <Num value={early.first} />,
		earlylast: <Num value={early.last} />,
		lag: <Num value={lag_share} form="share" />,
		corridor: <Num value={corridor_share} form="share" />,
		times: <Counted value={late_times} noun="research.times" />,
	};
}

// RuleValue is a rule's number of a row, with its interval, and its mark when
// the row is a goal's.
function RuleValue({
	page,
	value,
	form,
}: {
	page: PageReader;
	value: Value;
	form: NumberForm;
}) {
	return (
		<>
			<span class="s-research-number">
				<Num value={value.value} form={form} />
			</span>{" "}
			<span class="s-research-interval">
				{page.text("model.interval", {
					low: <Num value={value.low} form={form} />,
					high: <Num value={value.high} form={form} />,
				})}
			</span>
			{value.mark === null ? null : (
				<span class={`s-pill s-research-mark s-research-${value.mark}`}>
					{page.text(`model.marks.${value.mark}.name`)}
				</span>
			)}
		</>
	);
}

// Goal is a row's goal: the bound on the better side, the rule it is drawn
// by, and whether it is read off the service's own number; none for a number
// shown for comparison.
function Goal({
	page,
	row,
	form,
	slots,
}: {
	page: PageReader;
	row: Row;
	form: NumberForm;
	slots: Parameters<PageReader["text"]>[1];
}) {
	if (row.bound === null) {
		return <>{page.text("model.none")}</>;
	}
	return (
		<>
			<span class="s-research-number">
				{page.text(
					row.better === "lower" ? "model.at_most" : "model.at_least",
					{ bound: <Num value={row.bound.value} form={form} /> },
				)}
			</span>{" "}
			<span class="s-research-rule">
				{page.text(`model.rows.${row.id}.goal`, slots)}
			</span>
			{row.bound.own ? (
				<span class="s-research-own">{page.text("model.own")}</span>
			) : null}
		</>
	);
}

// Ceiling is how far a row could go: the oracle's number, which knows where
// the child stands, perfection, or none, for what the oracle does not show.
function Ceiling({
	page,
	row,
	form,
}: {
	page: PageReader;
	row: Row;
	form: NumberForm;
}) {
	const ceiling = row.values.ceiling;
	if (ceiling === null) {
		return <>{page.text("model.none")}</>;
	}
	if (ceiling.of === "perfect") {
		return <>{page.text("model.ceilings.perfect.name")}</>;
	}
	return (
		<>
			<span class="s-research-number">
				<Num value={ceiling.value} form={form} />
			</span>{" "}
			<span class="s-research-interval">
				{page.text("model.interval", {
					low: <Num value={ceiling.low} form={form} />,
					high: <Num value={ceiling.high} form={form} />,
				})}
			</span>
		</>
	);
}

// Chart draws a row: a bar for each rule with a whisker for its interval, a
// tick for the bound and one for the oracle's ceiling. It is hidden from a
// screen reader, which reads the row's cells.
function Chart({ row }: { row: Row }) {
	const scale = chartScale(row);
	const at = (value: number) => `${(value / scale) * 100}%`;
	const ceiling = row.values.ceiling;
	return (
		<span class="s-research-chart" dir="ltr" aria-hidden="true">
			{(["earlier", "service"] as const).map((rule) => {
				const value = row.values[rule];
				return (
					<span key={rule} class={`s-research-bar s-research-bar-${rule}`}>
						<span
							class="s-research-fill"
							style={{ inlineSize: at(value.value) }}
						/>
						<span
							class="s-research-whisker"
							style={{
								insetInlineStart: at(value.low),
								inlineSize: at(value.high - value.low),
							}}
						/>
					</span>
				);
			})}
			{row.bound === null ? null : (
				<span
					class="s-research-bound"
					style={{ insetInlineStart: at(row.bound.value) }}
				/>
			)}
			{ceiling?.of === "oracle" ? (
				<span
					class="s-research-ceiling"
					style={{ insetInlineStart: at(ceiling.value) }}
				/>
			) : null}
		</span>
	);
}

// Legend says what every mark and every ceiling means, whatever the rows of
// this build hold.
function Legend({ page }: { page: PageReader }) {
	return (
		<dl class="s-research-legend">
			{marks.map((mark) => (
				<div key={mark}>
					<dt>
						<span class={`s-pill s-research-mark s-research-${mark}`}>
							{page.text(`model.marks.${mark}.name`)}
						</span>
					</dt>
					<dd>{page.text(`model.marks.${mark}.means`)}</dd>
				</div>
			))}
			{ceilingKinds.map((kind) => (
				<div key={kind}>
					<dt>{page.text(`model.ceilings.${kind}.name`)}</dt>
					<dd>{page.text(`model.ceilings.${kind}.means`)}</dd>
				</div>
			))}
		</dl>
	);
}

// Provenance is where the numbers come from: the commit they are computed
// from and its date, the run, its intervals, that no child took part, and the
// commit of the paper's own numbers.
function Provenance({
	page,
	research,
}: {
	page: PageReader;
	research: Research;
}) {
	const { bench, built_from, paper } = research;
	return (
		<div class="s-research-provenance">
			<h3>{page.text("model.provenance.title")}</h3>
			<p>
				{page.text("model.provenance.commit", {
					commit: <Commit hash={built_from.commit} />,
					date: <CommitDate date={built_from.date} />,
				})}
			</p>
			<p>
				{page.text("model.provenance.run", {
					children: <Counted value={bench.children} noun="research.children" />,
					answers: <Counted value={bench.answers} noun="research.given" />,
					seed: <Num value={bench.seed} form="seed" />,
				})}
			</p>
			<p>
				{page.text("model.provenance.intervals", {
					interval: <Num value={bench.interval} form="share" />,
					resamples: <Num value={bench.resamples} />,
				})}
			</p>
			<p>{page.text("model.provenance.nobody")}</p>
			<p>
				{page.text("model.provenance.paper", {
					commit: <Commit hash={paper.commit} />,
				})}
			</p>
		</div>
	);
}
