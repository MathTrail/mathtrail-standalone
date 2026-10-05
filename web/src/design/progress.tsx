import type { ComponentChildren } from "preact";
import { classes } from "./classes";
import { Icon, type IconName } from "./icons";
import { useScopedId } from "./ids";
import { type Linking, PageLink } from "./links";

/**
 * SegmentTone is the ink a course is filled in: a step of the ramp a rank is
 * drawn in, from 1 at the first ranks to 5 at the last, or the plain ink of a
 * course that is no topic's rank.
 */
export type SegmentTone = "ink" | 1 | 2 | 3 | 4 | 5;

/**
 * MoveWay is which way a rank moved over a while, as a course and the words
 * beside it draw it: a gain, or a step back.
 */
export type MoveWay = "gain" | "loss";

/**
 * SegmentsAt is where a course stood: how many steps were behind it, and how
 * far through the next one it had come.
 */
export type SegmentsAt = { filled: number; part: number };

/**
 * Segments is a course of equal steps drawn as a row: the steps behind filled,
 * the one under way filled as far as it has come, the rest empty — or, where
 * nothing is measured yet, every step drawn open. A course that moved since
 * it stood where was says is filled as far as the lower of the two, and
 * striped on to the higher: in the colour of a gain when it moved up, of a
 * step back when it moved down. A label says in words, for a screen reader
 * alone, what the drawing shows; without one the drawing is a picture of what
 * the words beside it say. The lengths are attributes of the drawing rather
 * than styles: a card's page is held to its host's policy, which promises
 * nothing for a style set on an element.
 */
export function Segments({
	of,
	filled,
	part = 0,
	tone = "ink",
	open = false,
	label,
	was,
}: {
	of: number;
	filled: number;
	part?: number;
	tone?: SegmentTone;
	open?: boolean;
	label?: string;
	was?: SegmentsAt;
}) {
	return (
		<div class="mt-segments" data-tone={String(tone)}>
			{label !== undefined && <span class="mt-vh">{label}</span>}
			{counted(of).map((step) => {
				if (open) {
					return (
						<span
							key={step}
							class="mt-segment mt-segment-open"
							aria-hidden="true"
						/>
					);
				}
				const drawn = drawnStep(step, filled, part, was);
				return drawn.stripes > drawn.fill ? (
					<StripedStep key={step} {...drawn} />
				) : (
					<svg key={step} class="mt-segment" height="6" aria-hidden="true">
						<rect class="mt-segment-track" width="100%" height="6" rx="3" />
						<rect
							class="mt-segment-fill"
							width={lengthOf(drawn.fill)}
							height="6"
							rx="3"
						/>
					</svg>
				);
			})}
		</div>
	);
}

// StripedStep is a step a move is drawn across: its track, the stripes of the
// move under its fill, reaching as far as the move does, and the fill over
// them. The stripes are a pattern of the step's own, by a name no other
// drawing on the page has, and take their colour from the move's way.
function StripedStep({
	fill,
	stripes,
	way,
}: {
	fill: number;
	stripes: number;
	way: MoveWay;
}) {
	const pattern = useScopedId();
	return (
		<svg class="mt-segment" height="6" aria-hidden="true">
			<defs>
				<pattern
					id={pattern}
					class={`mt-stripes mt-stripes-${way}`}
					patternUnits="userSpaceOnUse"
					width="4"
					height="4"
					patternTransform="rotate(45)"
				>
					<rect width="2" height="4" />
				</pattern>
			</defs>
			<rect class="mt-segment-track" width="100%" height="6" rx="3" />
			<rect
				class="mt-segment-stripes"
				width={lengthOf(stripes)}
				height="6"
				rx="3"
				fill={`url(#${pattern})`}
			/>
			<rect class="mt-segment-fill" width={lengthOf(fill)} height="6" rx="3" />
		</svg>
	);
}

// leastShown is the least part of a step a move is drawn over: a step is about
// 22 px wide on the narrowest card, 320 px, and a seventh of it is the 3 px a
// move takes to be seen at all.
const leastShown = 0.14;

// drawnStep is how the step at place step is drawn: how much of it is filled,
// and how far the stripes of a move reach under the fill, from where the
// course stood at was to where it stands — none when it did not move. A move
// too small to be seen, over all the steps it crosses, takes the least part of
// a step that is, at the edge of the step the course is under way in, so that
// the drawing never says less than the words beside it: a gain ends at the
// edge, and a step back starts there.
function drawnStep(
	step: number,
	filled: number,
	part: number,
	was: SegmentsAt | undefined,
): { fill: number; stripes: number; way: MoveWay } {
	if (was === undefined) {
		return { fill: fillOf(step, filled, part), stripes: 0, way: "gain" };
	}
	const now = filled + bounded(part);
	const then = was.filled + bounded(was.part);
	const way = now >= then ? "gain" : "loss";
	const within = (at: number) => bounded(at - (step - 1));
	let fill = within(Math.min(now, then));
	let stripes = within(Math.max(now, then));
	if (step === filled + 1 && Math.abs(now - then) < leastShown) {
		const edge = within(now);
		if (way === "gain") {
			stripes = Math.max(edge, leastShown);
			fill = Math.min(fill, stripes - leastShown);
		} else {
			stripes = Math.min(Math.max(stripes, edge + leastShown), 1);
			fill = Math.min(fill, stripes - leastShown);
		}
	}
	return { fill, stripes, way };
}

// bounded is a part of a step, held from nothing to all of it.
function bounded(part: number): number {
	return Math.min(Math.max(part, 0), 1);
}

// lengthOf is a part of a step as the length of a drawing within it, in whole
// percents.
function lengthOf(part: number): string {
	return `${Math.round(part * 100)}%`;
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
	return step === filled + 1 ? bounded(part) : 0;
}

/**
 * RankSummary is where the child stands, large: the name of the step reached,
 * the line that places it — the step out of how many —, the course drawn under
 * them, how it moved where that is told, and a line of what comes next. A
 * label names it for a screen reader where the name does not.
 */
export function RankSummary({
	label,
	name,
	meta,
	line,
	move,
	children,
}: {
	label?: string;
	name: string;
	meta: string;
	line: string;
	move?: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<section class="mt-rank" aria-label={label}>
			<div class="mt-rank-head">
				<span class="mt-rank-name">{name}</span>
				<span class="mt-meta">{meta}</span>
			</div>
			{children}
			{move}
			<p class="mt-rank-line">{line}</p>
		</section>
	);
}

// The arrow drawn before the words of a move, a picture of them.
const moveArrows: Record<MoveWay, string> = { gain: "↑", loss: "↓" };

/**
 * MoveLine is how a rank moved over a while, in a line: in the colour of the
 * move's way and after its arrow, or plain where it did not move or cannot be
 * told. The arrow is a picture of the words and is hidden from a screen
 * reader, which reads the line out again whenever it changes; a line with
 * nothing to say takes no room.
 */
export function MoveLine({ way, text }: { way?: MoveWay; text?: string }) {
	return (
		<p class="mt-rank-move" data-way={way} aria-live="polite">
			{way !== undefined && text !== undefined && (
				<span class="mt-move-arrow" aria-hidden="true">
					{moveArrows[way]}
				</span>
			)}
			{text}
		</p>
	);
}

/**
 * MoveCounts is how many of a list's rows moved up over a while and how many
 * moved back, an arrow and a number for each way some did, and the same in
 * words for a screen reader alone. It is a summary: each row says its own move
 * where the summary leads.
 */
export function MoveCounts({
	up,
	down,
	label,
}: {
	up?: string;
	down?: string;
	label: string;
}) {
	return (
		<span class="mt-move-counts">
			<span class="mt-move-counted" aria-hidden="true">
				{up !== undefined && (
					<span data-way="gain">{`${moveArrows.gain} ${up}`}</span>
				)}
				{down !== undefined && (
					<span data-way="loss">{`${moveArrows.loss} ${down}`}</span>
				)}
			</span>
			<span class="mt-vh">{label}</span>
		</span>
	);
}

/**
 * MoveLegend says what the stripes on the courses are: a gain and a step back,
 * each beside a sample of its stripes, and the while they are drawn over.
 */
export function MoveLegend({
	gain,
	loss,
	period,
}: {
	gain: string;
	loss: string;
	period: string;
}) {
	return (
		<p class="mt-move-legend">
			<span class="mt-move-key">
				<span class="mt-move-sample" data-way="gain" aria-hidden="true" />
				{gain}
			</span>
			<span class="mt-move-key">
				<span class="mt-move-sample" data-way="loss" aria-hidden="true" />
				{loss}
			</span>
			<span class="mt-move-period">
				<span aria-hidden="true">· </span>
				{period}
			</span>
		</p>
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
 * StatusDots is how a run of entries went, a dot for each in its tone's
 * colour, the first entry first, and the same in words for a screen reader
 * alone. It is a summary: the entries are told, each with its mark and its
 * words, where the summary leads.
 */
export function StatusDots({
	tones,
	label,
}: {
	tones: readonly StatusTone[];
	label: string;
}) {
	return (
		<span class="mt-status-dots">
			<span class="mt-dots" aria-hidden="true">
				{counted(tones.length).map((place) => (
					<span key={place} class="mt-dot" data-tone={tones[place - 1]} />
				))}
			</span>
			<span class="mt-vh">{label}</span>
		</span>
	);
}

/**
 * RankRow is one topic of a list of ranks: what it is about, a mark when it is
 * mastered, a word of how it stands — or of how it moved, in the colour of the
 * move's way and after its arrow — and the name of its rank at the line's end,
 * and its course under them; and the address of the topic's page, when it has
 * one to link to. Its id tells it apart from the others.
 */
export type RankRow = {
	id: string;
	label: string;
	href?: string;
	mark?: { tone: StatusTone; label: string };
	word?: string;
	way?: MoveWay;
	name?: string;
	segments: Parameters<typeof Segments>[0];
};

/**
 * RankList is a list of topics, each with its own course, under a line that
 * says what the courses show when there is one, and what their stripes are
 * when they are drawn with any. It carries its label, unless what it stands in
 * names it already — the title of a part folded away. A topic with a page is
 * named by a link to it, where the screen links pages at all.
 */
export function RankList({
	label,
	note,
	legend,
	rows,
	linking,
}: {
	label?: string;
	note?: string;
	legend?: ComponentChildren;
	rows: readonly RankRow[];
	linking?: Linking;
}) {
	return (
		<section class="mt-list">
			{label !== undefined && <h3 class="mt-section-label">{label}</h3>}
			{note !== undefined && <p class="mt-list-lead">{note}</p>}
			{legend}
			<ul>
				{rows.map((row) => (
					<li key={row.id} class="mt-rank-row">
						<div class="mt-rank-row-top">
							<span class="mt-rank-row-start">
								<span class="mt-row-label">
									{row.href !== undefined && linking !== undefined ? (
										<PageLink
											label={row.label}
											href={row.href}
											linking={linking}
										/>
									) : (
										row.label
									)}
								</span>
								{row.mark !== undefined && (
									<StatusMark tone={row.mark.tone} label={row.mark.label} />
								)}
							</span>
							<span class="mt-rank-row-end">
								{row.word !== undefined && (
									<span class="mt-rank-word" data-way={row.way}>
										{row.way !== undefined && (
											<span class="mt-move-arrow" aria-hidden="true">
												{moveArrows[row.way]}
											</span>
										)}
										{row.word}
									</span>
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
 * StatList is a list of lines, each with what it is about at its start and
 * how it stands at its end, and a note under it when there is more to say of
 * the whole list; a list set apart is drawn in a frame of its own. It carries
 * its label, unless what it stands in names it already — the title of a part
 * folded away.
 */
export function StatList({
	label,
	rows,
	note,
	framed = false,
}: {
	label?: string;
	rows: readonly StatRow[];
	note?: string;
	framed?: boolean;
}) {
	return (
		<section class={classes("mt-list", framed && "mt-list-framed")}>
			{label !== undefined && <h3 class="mt-section-label">{label}</h3>}
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
 * FieldsFrame is a frame for a group of a profile's fields: at its head their
 * label — unless what the frame stands in names it already, the title of a
 * part folded away — and what can be done about them, what became of it under
 * the head, and the fields — told, or a form to change them — below. A frame
 * with nothing for its head has none.
 */
export function FieldsFrame({
	label,
	action,
	status,
	children,
}: {
	label?: string;
	action?: ComponentChildren;
	status?: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<section class="mt-fields">
			{(label !== undefined || action !== undefined) && (
				<div class="mt-fields-head">
					{label !== undefined && <h3 class="mt-section-label">{label}</h3>}
					{action}
				</div>
			)}
			{status}
			{children}
		</section>
	);
}

/**
 * ProfileFields is a group of a profile's fields in a frame, each named, with
 * its label, unless what it stands in names it already, and what can be done
 * about them at its head, and what became of it under the head.
 */
export function ProfileFields({
	label,
	fields,
	action,
	status,
}: {
	label?: string;
	fields: readonly Field[];
	action?: ComponentChildren;
	status?: ComponentChildren;
}) {
	return (
		<FieldsFrame label={label} action={action} status={status}>
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
		</FieldsFrame>
	);
}
