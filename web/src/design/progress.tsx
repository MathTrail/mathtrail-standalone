import type { ComponentChildren } from "preact";
import { classes } from "./classes";
import { Icon, type IconName } from "./icons";

/**
 * RatingSummary is a number the child climbs, large, with the line that
 * places it above it and the words that name it beside it — and, when a
 * course of steps is shown, how many of them are behind. A label names it
 * for a screen reader where the line above does not.
 */
export function RatingSummary({
	label,
	rankLabel,
	rating,
	ratingLabel,
	pips,
}: {
	label?: string;
	rankLabel: string;
	rating: string;
	ratingLabel: string;
	pips?: { on: number; of: number };
}) {
	return (
		<section class="mt-rating" aria-label={label}>
			<span class="mt-rating-rank">{rankLabel}</span>
			<div class="mt-rating-line">
				<span class="mt-rating-num">{rating}</span>
				<span class="mt-meta">{ratingLabel}</span>
			</div>
			{pips !== undefined && (
				<div class="mt-pips" aria-hidden="true">
					{counted(pips.of).map((pip) => (
						<span key={pip} class="mt-pip" data-on={String(pip <= pips.on)} />
					))}
				</div>
			)}
		</section>
	);
}

// counted are the numbers from one to count: the pips, each told apart by
// its place in the row.
function counted(count: number): number[] {
	return Array.from({ length: Math.max(count, 0) }, (_, at) => at + 1);
}

/**
 * StatusTone is how an entry went: right, wrong, or left without an answer.
 */
export type StatusTone = "correct" | "wrong" | "skipped";

// The mark beside each tone's words.
const statusIcons: Record<StatusTone, IconName> = {
	correct: "check",
	wrong: "cross",
	skipped: "dash",
};

/** StatusMark is how an entry went, in words and with its mark. */
export function StatusMark({
	tone,
	label,
}: {
	tone: StatusTone;
	label: string;
}) {
	return (
		<span class={classes("mt-status", `mt-status-${tone}`)}>
			<Icon name={statusIcons[tone]} size={16} />
			{label}
		</span>
	);
}

/**
 * StatRow is one line of a list: what it is about, and how it stands — a
 * mark, a number, or both — or, on a line that has one, how often, with a bar
 * as long as its share of the most. Its id tells it apart from the others,
 * since two lines may say the same.
 */
export type StatRow = {
	id: string;
	label: string;
	status?: { tone: StatusTone; label: string };
	value?: string;
	bar?: { share: number; count: string };
};

/**
 * StatList is a labelled list of lines, each with what it is about at its
 * start and how it stands at its end, and a note under it when there is more
 * to say of the whole list.
 */
export function StatList({
	label,
	rows,
	note,
}: {
	label: string;
	rows: readonly StatRow[];
	note?: string;
}) {
	return (
		<section class="mt-list">
			<h3 class="mt-section-label">{label}</h3>
			<ul>
				{rows.map((row) =>
					row.bar === undefined ? (
						<li key={row.id} class="mt-row">
							<span class="mt-row-label">{row.label}</span>
							<span class="mt-row-spacer" />
							{row.status !== undefined && (
								<StatusMark tone={row.status.tone} label={row.status.label} />
							)}
							{row.value !== undefined && (
								<span class="mt-row-value">{row.value}</span>
							)}
						</li>
					) : (
						<li key={row.id} class="mt-row mt-row-bar">
							<div class="mt-row-bar-top">
								<span class="mt-row-label">{row.label}</span>
								<span class="mt-row-spacer" />
								<span class="mt-row-count">{row.bar.count}</span>
							</div>
							<Bar share={row.bar.share} />
						</li>
					),
				)}
			</ul>
			{note !== undefined && <p class="mt-list-note">{note}</p>}
		</section>
	);
}

// Bar is a share from nothing to all of it, drawn as a filled length of its
// track. The length is an attribute of the drawing rather than a style: a
// card's page is held to its host's policy, which promises nothing for a style
// set on an element.
function Bar({ share }: { share: number }) {
	const filled = Math.round(Math.min(Math.max(share, 0), 1) * 100);
	return (
		<svg class="mt-bar-chart" width="100%" height="4" aria-hidden="true">
			<rect class="mt-bar-track" width="100%" height="4" rx="2" />
			<rect class="mt-bar-fill" width={`${filled}%`} height="4" rx="2" />
		</svg>
	);
}

/**
 * Field is one of a profile's fields: its name, what it holds, and a line
 * under it when what it holds needs a word of explanation.
 */
export type Field = { term: string; value: string; note?: string };

/**
 * ProfileFields is a profile's fields under their label, each named, and
 * what can be done about them under the fields.
 */
export function ProfileFields({
	label,
	fields,
	action,
}: {
	label: string;
	fields: readonly Field[];
	action?: ComponentChildren;
}) {
	return (
		<section class="mt-fields">
			<h3 class="mt-section-label">{label}</h3>
			<dl>
				{fields.map((field) => (
					<div key={field.term}>
						<dt>{field.term}</dt>
						<dd>{field.value}</dd>
						{field.note !== undefined && (
							<dd class="mt-field-note">{field.note}</dd>
						)}
					</div>
				))}
			</dl>
			{action !== undefined && <div class="mt-fields-actions">{action}</div>}
		</section>
	);
}
