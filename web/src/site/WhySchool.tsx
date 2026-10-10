import type { ComponentChildren } from "preact";
import type { PageReader } from "./reader";
import { Stroke } from "./WhyParts";

// schoolRows are the rows of how a lesson and an olympiad problem differ, by
// the names their words have, each with the mark before its label: how the
// problem is built, what counts, what a mistake means and what it teaches.
const schoolRows = [
	["shape", "lines"],
	["valued", "star"],
	["mistake", "warning"],
	["teaches", "wrench"],
] as const;

/**
 * School is how a school lesson and an olympiad problem differ: an example of
 * each side by side — a school sum with its answer left blank, and the fence
 * with its posts — then, row by row, the lesson's way in grey and the
 * olympiad's ticked in blue. The rows are a list of terms, each told on both
 * sides, which say which side they are on to a screen reader, since to the eye
 * the examples above them say it.
 */
export function School({ page }: { page: PageReader }) {
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2 id="school-title">{page.text("school.title")}</h2>
				<p class="s-intro-line">{page.text("school.lead")}</p>
			</div>
			<div class="s-why-sides">
				<div class="s-why-side s-why-side-lesson">
					<p class="s-why-side-name">
						<span class="s-why-badge">
							<Stroke size={20}>
								<path d="M4 5a2 2 0 0 1 2-2h13v16H6a2 2 0 0 0-2 2z" />
								<path d="M4 19V5" />
							</Stroke>
						</span>
						{page.text("school.lesson")}
					</p>
					<div class="s-why-side-example">
						<p class="s-why-side-label">{page.text("school.example")}</p>
						<p class="s-why-sum">
							{page.text("school.sum")}
							<span class="s-why-blank" aria-hidden="true" />
						</p>
						<p class="s-why-side-note">{page.text("school.done")}</p>
					</div>
				</div>
				<div class="s-why-side s-why-side-olympiad">
					<p class="s-why-side-name">
						<span class="s-why-badge">
							<Stroke size={20}>
								<path d="M9 18h6M10 22h4M15.09 14c.18-.98.65-1.74 1.41-2.5A4.65 4.65 0 0 0 18 8 6 6 0 0 0 6 8c0 1 .23 2.23 1.5 3.5A4.61 4.61 0 0 1 8.91 14" />
							</Stroke>
						</span>
						{page.text("school.olympiad")}
					</p>
					<div class="s-why-side-example">
						<p class="s-why-side-label">{page.text("school.example")}</p>
						<p class="s-why-fence">{page.text("school.fence")}</p>
						<span class="s-why-posts" aria-hidden="true">
							<span />
							<i />
							<span />
							<i />
							<span />
							<i />
							<span />
							<i />
							<span />
						</span>
						<p class="s-why-side-note">{page.text("school.found")}</p>
					</div>
				</div>
			</div>
			<dl class="s-why-rows" aria-labelledby="school-title">
				{schoolRows.map(([row, mark]) => (
					<div key={row} class="s-why-row">
						<dt>
							<span class="s-why-badge">
								<Stroke size={20}>{rowMarks[mark]()}</Stroke>
							</span>
							{page.text(`school.rows.${row}.label`)}
						</dt>
						<dd class="s-why-row-lesson">
							<span class="s-hidden">{page.text("school.lesson")}: </span>
							{page.text(`school.rows.${row}.lesson`)}
						</dd>
						<dd class="s-why-row-olympiad">
							<span class="s-why-tick" aria-hidden="true">
								✓
							</span>
							<span>
								<span class="s-hidden">{page.text("school.olympiad")}: </span>
								{page.text(`school.rows.${row}.olympiad`)}
							</span>
						</dd>
					</div>
				))}
			</dl>
		</section>
	);
}

// rowMarks are the marks before the rows' labels: lines of text, a star, a
// warning sign and a wrench.
const rowMarks: Readonly<
	Record<(typeof schoolRows)[number][1], () => ComponentChildren>
> = {
	lines: () => <path d="M4 6h16M4 12h10M4 18h7" />,
	star: () => (
		<path d="m12 3 2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1L3.2 9.5l6.1-.9z" />
	),
	warning: () => (
		<>
			<path d="M12 3 2.5 20h19z" />
			<path d="M12 10v4M12 17h.01" />
		</>
	),
	wrench: () => (
		<path d="M14.7 6.3a4 4 0 0 0-5.4 5.4L3 18l3 3 6.3-6.3a4 4 0 0 0 5.4-5.4l-2.5 2.5-2.4-.6-.6-2.4z" />
	),
};

/**
 * NoTime is why school has no time for olympiad problems, card by card, each
 * under its mark and number, with a small drawing at its foot: a lesson's
 * minutes taken by the syllabus, a test that reads only the answer, and a row
 * of schools with one circle among them. The answer is the fence's, from the
 * site's data.
 */
export function NoTime({ page, answer }: { page: PageReader; answer: string }) {
	const numbers = new Intl.NumberFormat(page.locale, {
		minimumIntegerDigits: 2,
	});
	const at = "time.cards";
	return (
		<section class="s-wrap s-section">
			<h2 class="s-title">{page.text("time.title")}</h2>
			<ol class="s-why-tiles">
				<NoTimeCard
					page={page}
					at={`${at}.syllabus`}
					number={numbers.format(1)}
					mark={<TimeMark name="clock" />}
				>
					<div class="s-why-lesson">
						<span class="s-why-lesson-ends">
							<span>{page.text(`${at}.syllabus.lesson`)}</span>
							<span>{page.text(`${at}.syllabus.length`)}</span>
						</span>
						<span class="s-why-lesson-bar">
							<span />
							<span />
						</span>
						<span class="s-why-lesson-ends">
							<span>{page.text(`${at}.syllabus.all`)}</span>
							<span class="s-why-lesson-extra">
								{page.text(`${at}.syllabus.extra`)}
							</span>
						</span>
					</div>
				</NoTimeCard>
				<NoTimeCard
					page={page}
					at={`${at}.answer`}
					number={numbers.format(2)}
					mark={<TimeMark name="check" />}
				>
					<div class="s-why-checked">
						<span class="s-why-checked-lines">
							<span />
							<span />
							<span />
						</span>
						<span class="s-why-checked-steps">
							{page.text(`${at}.answer.steps`)}
						</span>
						<span class="s-why-checked-answer">
							{page.text(`${at}.answer.answer`, { answer })}
							<span class="s-why-tick">✓</span>
						</span>
					</div>
				</NoTimeCard>
				<NoTimeCard
					page={page}
					at={`${at}.circle`}
					number={numbers.format(3)}
					mark={<TimeMark name="school" />}
				>
					<ul class="s-why-schools">
						{[false, false, true, false, false].map((circle, place) => (
							<li key={place} data-circle={circle ? "" : undefined}>
								<span class="s-why-school" />
								{page.text(`${at}.circle.${circle ? "circle" : "none"}`)}
							</li>
						))}
					</ul>
				</NoTimeCard>
			</ol>
		</section>
	);
}

// NoTimeCard is one card of why school has no time: its mark and number, its
// heading and words, and the drawing at its foot, which shows to the eye what
// the words say.
function NoTimeCard({
	page,
	at,
	number,
	mark,
	children,
}: {
	page: PageReader;
	at: string;
	number: string;
	mark: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<li class="s-why-tile">
			<span class="s-why-tile-top">
				<span class="s-why-badge">{mark}</span>
				<span class="s-why-tile-number" aria-hidden="true">
					{number}
				</span>
			</span>
			<h3>{page.text(`${at}.title`)}</h3>
			<p class="s-why-tile-text">{page.text(`${at}.text`)}</p>
			<div class="s-why-tile-foot" aria-hidden="true">
				{children}
			</div>
		</li>
	);
}

// TimeMark is the mark of a card of why school has no time: a clock, a box
// ticked, or a school.
function TimeMark({ name }: { name: "clock" | "check" | "school" }) {
	switch (name) {
		case "clock":
			return (
				<Stroke size={22}>
					<circle cx="12" cy="12" r="9" />
					<path d="M12 7v5l3 2" />
				</Stroke>
			);
		case "check":
			return (
				<Stroke size={22}>
					<path d="M9 11l3 3 8-8" />
					<path d="M20 12v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h9" />
				</Stroke>
			);
		case "school":
			return (
				<Stroke size={22}>
					<path d="M3 21h18M5 21V10l7-5 7 5v11" />
					<path d="M10 21v-5h4v5" />
				</Stroke>
			);
	}
}
