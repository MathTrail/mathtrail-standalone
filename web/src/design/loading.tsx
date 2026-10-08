import { useEffect, useState } from "preact/hooks";
import { GradeRuns, Segments } from "./progress";

/**
 * LoadingBar is the thin track over a screen being read, along which a stretch
 * of the accent runs for as long as the reading lasts, from where the card's
 * lines start to where they end. It takes no room of its own: it lies over the
 * top of what comes under it, so that nothing moves when it goes. It is a
 * picture; the screen says in words that it is being read.
 */
export function LoadingBar() {
	return (
		<span class="mt-loading-bar" aria-hidden="true">
			<span class="mt-loading-run" />
		</span>
	);
}

// The titles of the progress's sections as the outline draws them, folded,
// each with its summary: the topics, the review, the latest answers — a run
// of dots — and the profile. Each is as long as the words it stands for.
const sections: readonly { title: number; summary: number; thick: number }[] = [
	{ title: 56, summary: 44, thick: 10 },
	{ title: 64, summary: 150, thick: 10 },
	{ title: 140, summary: 56, thick: 8 },
	{ title: 72, summary: 84, thick: 10 },
];

/**
 * ProgressSkeleton is the progress drawn before it is read: the outline of the
 * screen it will be, in the screen's own frames — the switch of the while, the
 * rank over its course, as long as its runs and with them marked under it,
 * the lines under them, what comes next, and the titles of four sections,
 * folded. A faint shape stands wherever the screen's words and marks will,
 * each in a line as high as theirs, so that the screen comes in close to
 * where its outline stood, and a sheen passes over the outline. It takes the
 * classes of the screen's frames, never those of its words or its controls,
 * which whoever waits for the screen looks for. A status tells a screen
 * reader the label, and the shapes are hidden from it. The status is drawn
 * empty and given the label once it is on the page: a screen reader says
 * what changes in a status, and passes over what one says as it comes.
 */
export function ProgressSkeleton({
	label,
	runs,
}: {
	label: string;
	runs: readonly { first: number; last: number }[];
}) {
	const [said, say] = useState("");
	useEffect(() => say(label), [label]);
	const steps = runs.reduce((end, run) => Math.max(end, run.last), 0);
	return (
		<div class="mt-skeleton">
			<output class="mt-vh">{said}</output>
			<div aria-hidden="true">
				<div class="mt-progress">
					<div class="mt-switch">
						{[120, 80].map((width) => (
							<span key={width} class="mt-switch-option">
								<span class="mt-switch-label">
									<Shape width={width} height={10} />
								</span>
							</span>
						))}
					</div>
					<div class="mt-rank">
						<div class="mt-rank-head">
							<Shape width={132} height={30} line={38} />
							<Shape width={36} height={10} line={18} />
						</div>
						<Segments of={steps} filled={0} />
						<div class="mt-grades">
							<GradeRuns
								of={steps}
								runs={runs.map((run) => ({
									...run,
									label: <Shape width={44} height={9} line={16} />,
								}))}
							/>
							<div>
								<Shape width="94%" height={9} line={17} />
								<Shape width="58%" height={9} line={17} />
							</div>
						</div>
						<Shape width={240} height={13} line={22} />
						<div>
							<Shape width="100%" height={10} line={22} />
							<Shape width="42%" height={10} line={22} />
						</div>
					</div>
					<div class="mt-note mt-note-plain">
						<Shape width={52} height={9} line={18} />
						<Shape width={220} height={13} line={22} />
					</div>
				</div>
				<div class="mt-folds">
					{sections.map((section) => (
						<div key={section.title} class="mt-fold">
							<div class="mt-skeleton-row">
								<Shape width={section.title} height={13} line={22} />
								<Shape width={section.summary} height={section.thick} />
								<Shape width={12} height={12} rounding={3} />
							</div>
						</div>
					))}
				</div>
			</div>
			<span class="mt-sheen" aria-hidden="true" />
		</div>
	);
}

// Shape is one faint shape of the outline: a bar height pixels thick and as
// long as width, in the middle of a line as high as line — the line of the
// words it stands for —, its ends rounded. Its sizes are attributes of the
// drawing rather than styles: a card's page is held to its host's policy,
// which promises nothing for a style set on an element.
function Shape({
	width,
	height,
	line = height,
	rounding = Math.min(height / 2, 7),
}: {
	width: number | string;
	height: number;
	line?: number;
	rounding?: number;
}) {
	return (
		<svg class="mt-shape" width={width} height={line} aria-hidden="true">
			<rect
				y={(line - height) / 2}
				width="100%"
				height={height}
				rx={rounding}
			/>
		</svg>
	);
}
