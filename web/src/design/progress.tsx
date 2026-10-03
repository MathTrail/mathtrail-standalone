import type { ComponentChildren } from "preact";
import { classes } from "./classes";
import { Icon, type IconName } from "./icons";

/**
 * SegmentTone is the ink a course is filled in: a step of the ramp a rank is
 * drawn in, from 1 at the first ranks to 5 at the last, or the plain ink of a
 * course that is no topic's rank.
 */
export type SegmentTone = "ink" | 1 | 2 | 3 | 4 | 5;

/**
 * Segments is a course of equal steps drawn as a row: the steps behind filled,
 * the one under way filled as far as it has come, the rest empty — or, where
 * nothing is measured yet, every step drawn open. A label says in words, for a
 * screen reader alone, what the drawing shows; without one the drawing is a
 * picture of what the words beside it say. The lengths are attributes of the
 * drawing rather than styles: a card's page is held to its host's policy,
 * which promises nothing for a style set on an element.
 */
export function Segments({
	of,
	filled,
	part = 0,
	tone = "ink",
	open = false,
	label,
}: {
	of: number;
	filled: number;
	part?: number;
	tone?: SegmentTone;
	open?: boolean;
	label?: string;
}) {
	return (
		<div class="mt-segments" data-tone={String(tone)}>
			{label !== undefined && <span class="mt-vh">{label}</span>}
			{counted(of).map((step) =>
				open ? (
					<span
						key={step}
						class="mt-segment mt-segment-open"
						aria-hidden="true"
					/>
				) : (
					<svg key={step} class="mt-segment" height="6" aria-hidden="true">
						<rect class="mt-segment-track" width="100%" height="6" rx="3" />
						<rect
							class="mt-segment-fill"
							width={`${Math.round(fillOf(step, filled, part) * 100)}%`}
							height="6"
							rx="3"
						/>
					</svg>
				),
			)}
		</div>
	);
}

// counted are the numbers from one to count: the steps, each told apart by its
// place in the row.
function counted(count: number): number[] {
	return Array.from({ length: Math.max(count, 0) }, (_, at) => at + 1);
}

// fillOf is how much of the step at place step is filled, from nothing to all
// of it, when filled steps are behind and part of the next one is.
function fillOf(step: number, filled: number, part: number): number {
	if (step <= filled) {
		return 1;
	}
	return step === filled + 1 ? Math.min(Math.max(part, 0), 1) : 0;
}

/**
 * RankSummary is where the child stands, large: the name of the step reached,
 * the line that places it — the step out of how many, and the number beside
 * it — the course drawn under them, and a line of what comes next. A label
 * names it for a screen reader where the name does not.
 */
export function RankSummary({
	label,
	name,
	meta,
	line,
	children,
}: {
	label?: string;
	name: string;
	meta: string;
	line: string;
	children: ComponentChildren;
}) {
	return (
		<section class="mt-rank" aria-label={label}>
			<div class="mt-rank-head">
				<span class="mt-rank-name">{name}</span>
				<span class="mt-meta">{meta}</span>
			</div>
			{children}
			<p class="mt-rank-line">{line}</p>
		</section>
	);
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
 * RankRow is one topic of a list of ranks: what it is about, a mark when it is
 * mastered, a word of how it stands and the name of its rank at the line's
 * end, and its course under them. Its id tells it apart from the others.
 */
export type RankRow = {
	id: string;
	label: string;
	mark?: { tone: StatusTone; label: string };
	word?: string;
	name?: string;
	segments: Parameters<typeof Segments>[0];
};

/**
 * RankList is a labelled list of topics, each with its own course, under a
 * line that says what the courses show when there is one.
 */
export function RankList({
	label,
	note,
	rows,
}: {
	label: string;
	note?: string;
	rows: readonly RankRow[];
}) {
	return (
		<section class="mt-list">
			<h3 class="mt-section-label">{label}</h3>
			{note !== undefined && <p class="mt-list-lead">{note}</p>}
			<ul>
				{rows.map((row) => (
					<li key={row.id} class="mt-rank-row">
						<div class="mt-rank-row-top">
							<span class="mt-rank-row-start">
								<span class="mt-row-label">{row.label}</span>
								{row.mark !== undefined && (
									<StatusMark tone={row.mark.tone} label={row.mark.label} />
								)}
							</span>
							<span class="mt-rank-row-end">
								{row.word !== undefined && (
									<span class="mt-rank-word">{row.word}</span>
								)}
								{row.name !== undefined && (
									<span class="mt-rank-row-name">{row.name}</span>
								)}
							</span>
						</div>
						<Segments {...row.segments} />
					</li>
				))}
			</ul>
		</section>
	);
}

/** mostDots is the most dots a line draws, however many times it counts. */
const mostDots = 5;

/**
 * StatRow is one line of a list: what it is about, and how it stands — a mark
 * — or, on a line that counts, how many times, with a dot for each, up to a
 * handful. Its id tells it apart from the others, since two lines may say the
 * same.
 */
export type StatRow = {
	id: string;
	label: string;
	status?: { tone: StatusTone; label: string };
	count?: { times: number; label: string };
};

/**
 * StatList is a labelled list of lines, each with what it is about at its
 * start and how it stands at its end, and a note under it when there is more
 * to say of the whole list; a list set apart is drawn in a frame of its own.
 */
export function StatList({
	label,
	rows,
	note,
	framed = false,
}: {
	label: string;
	rows: readonly StatRow[];
	note?: string;
	framed?: boolean;
}) {
	return (
		<section class={classes("mt-list", framed && "mt-list-framed")}>
			<h3 class="mt-section-label">{label}</h3>
			<ul>
				{rows.map((row) => (
					<li key={row.id} class="mt-row">
						<span class="mt-row-label">{row.label}</span>
						<span class="mt-row-spacer" />
						{row.status !== undefined && (
							<StatusMark tone={row.status.tone} label={row.status.label} />
						)}
						{row.count !== undefined && (
							<span class="mt-row-count">
								<span class="mt-dots" aria-hidden="true">
									{counted(Math.min(row.count.times, mostDots)).map((dot) => (
										<span key={dot} class="mt-dot" />
									))}
								</span>
								{row.count.label}
							</span>
						)}
					</li>
				))}
			</ul>
			{note !== undefined && <p class="mt-list-note">{note}</p>}
		</section>
	);
}

/**
 * Field is one of a profile's fields: its name, what it holds — a text, or a
 * list of names, each drawn apart, a name given twice drawn once — and a line
 * under it when what it holds needs a word of explanation.
 */
export type Field = {
	term: string;
	value: string | readonly string[];
	note?: string;
};

/**
 * ProfileFields is a group of a profile's fields in a frame: their label at
 * its head, with what can be done about them beside it and what became of it
 * under it, and each field named.
 */
export function ProfileFields({
	label,
	fields,
	action,
	status,
}: {
	label: string;
	fields: readonly Field[];
	action?: ComponentChildren;
	status?: ComponentChildren;
}) {
	return (
		<section class="mt-fields">
			<div class="mt-fields-head">
				<h3 class="mt-section-label">{label}</h3>
				{action}
			</div>
			{status}
			<dl>
				{fields.map((field) => (
					<div key={field.term}>
						<dt>{field.term}</dt>
						<dd>
							{typeof field.value === "string" ? (
								field.value
							) : (
								<ul class="mt-chips">
									{[...new Set(field.value)].map((item) => (
										<li key={item} class="mt-chip">
											{item}
										</li>
									))}
								</ul>
							)}
							{field.note !== undefined && (
								<p class="mt-field-note">{field.note}</p>
							)}
						</dd>
					</div>
				))}
			</dl>
		</section>
	);
}
