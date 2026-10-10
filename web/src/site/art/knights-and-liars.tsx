import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import { Line, Round, Say, Stack, Then } from "../Sketch";
import type { TopicArt } from "./art";

// Who is one of the islanders, A, B and C.
type Who = "a" | "b" | "c";

// Kind is what an islander turns out to be, or is supposed to be.
type Kind = "knight" | "liar";

// Islander is an islander as a circle with their letter: dashed while it is
// not known what they are, green for a knight and red for a liar; ringed in
// the accent where a step picks them, and small in a row of words.
function Islander({
	page,
	who,
	is,
	picked = false,
	size = "md",
}: {
	page: PageReader;
	who: Who;
	is?: Kind;
	picked?: boolean;
	size?: "md" | "sm";
}) {
	return (
		<span
			class="s-kl-islander"
			data-is={is}
			data-picked={picked ? "" : undefined}
			data-size={size}
		>
			{page.text(`art.${who}`)}
		</span>
	);
}

// Sign is the letter of a knight or of a liar on a circle of its colour.
function Sign({ page, is }: { page: PageReader; is: Kind }) {
	return (
		<span class="s-kl-islander" data-is={is} data-size="sm">
			{page.text(`art.${is}`)}
		</span>
	);
}

// Said is what an islander says, in a bubble.
function Said({ children }: { children: ComponentChildren }) {
	return <span class="s-kl-said">{children}</span>;
}

// Verdict is how a supposition ends: a cross and a contradiction in red, or
// a tick and that everything fits in green.
function Verdict({
	fits,
	children,
}: {
	fits: boolean;
	children: ComponentChildren;
}) {
	return (
		<span class="s-kl-verdict" data-fits={fits ? "" : undefined}>
			{fits ? "✓" : "✕"} {children}
		</span>
	);
}

// Supposition is one case of a supposition on a band of how it ends: the
// islander as supposed, the reasoning, and the verdict at its end.
function Supposition({
	page,
	who,
	is,
	fits,
	verdict,
	children,
}: {
	page: PageReader;
	who: Who;
	is: Kind;
	fits: boolean;
	verdict: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<span class="s-kl-supposition" data-fits={fits ? "" : undefined}>
			<Islander page={page} who={who} is={is} size="sm" />
			<span class="s-kl-reason">{children}</span>
			<Verdict fits={fits}>{verdict}</Verdict>
		</span>
	);
}

// Speaker is an islander and what they say, side by side.
function Speaker({
	page,
	who,
	is,
	picked,
	children,
}: {
	page: PageReader;
	who: Who;
	is?: Kind;
	picked?: boolean;
	children: ComponentChildren;
}) {
	return (
		<Line gap={10}>
			<Islander page={page} who={who} is={is} picked={picked} />
			<Said>{children}</Said>
		</Line>
	);
}

// Checked is one line of the check: the islander as found, what they said,
// whether that is true and who says it, and a tick.
function Checked({
	page,
	who,
	is,
	said,
	children,
}: {
	page: PageReader;
	who: Who;
	is: Kind;
	said: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<span class="s-kl-checked">
			<Islander page={page} who={who} is={is} size="sm" />
			<span class="s-kl-reason">
				{said}
				<span class="s-kl-why">{children}</span>
			</span>
			<span class="s-kl-tick">✓</span>
		</span>
	);
}

// Table is islanders round a table, knights and liars taking turns from the
// top clockwise, what they make written in the middle; where the table has
// an odd number of seats, the last and the first are one kind side by side,
// ringed in red.
function Table({
	page,
	seats,
	middle,
}: {
	page: PageReader;
	seats: number;
	middle?: ComponentChildren;
}) {
	return (
		<Round size={220} middle={middle}>
			{Array.from({ length: seats }, (_, seat) => {
				const is: Kind = seat % 2 === 0 ? "knight" : "liar";
				const clash = seats % 2 === 1 && (seat === 0 || seat === seats - 1);
				return (
					<span
						key={seat}
						class="s-kl-islander s-kl-seated"
						data-is={is}
						data-clash={clash ? "" : undefined}
					>
						{page.text(`art.${is}`)}
					</span>
				);
			})}
		</Round>
	);
}

/**
 * art are the drawings of knights and liars: islanders as circles, dashed
 * until it is known what they are, green for a knight and red for a liar;
 * what they say in bubbles; each supposition on a band of how it ends; and
 * islanders round a table, taking turns.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => (
		<span class="s-kl-legend">
			<span class="s-kl-sign">
				<Sign page={page} is="knight" />
				{page.text(`${at}.knight`)}
			</span>
			<span class="s-kl-sign">
				<Sign page={page} is="liar" />
				{page.text(`${at}.liar`)}
			</span>
		</span>
	),
	hero: ({ page, at }) => (
		<>
			<Line gap={12}>
				<Islander page={page} who="a" />
				<Said>{page.text(`${at}.says`)}</Said>
				<Islander page={page} who="b" />
			</Line>
			<Say>↓ {page.text(`${at}.rule`)}</Say>
			<Stack gap={8}>
				<Supposition
					page={page}
					who="a"
					is="knight"
					fits={false}
					verdict={page.text(`${at}.contradiction`)}
				>
					{page.text(`${at}.if-knight`)}
				</Supposition>
				<Supposition
					page={page}
					who="a"
					is="liar"
					fits
					verdict={page.text(`${at}.fits`)}
				>
					{page.text(`${at}.if-liar`)}
				</Supposition>
			</Stack>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={12}>
					<Islander page={page} who="a" />
					<Said>{page.text(`${at}.says`)}</Said>
					<Islander page={page} who="b" />
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Line gap={10}>
						<Islander page={page} who="a" is="knight" />
						<Then />
						<Said>{page.text(`${at}.says`)}</Said>
						<Verdict fits={false}>{page.text(`${at}.contradiction`)}</Verdict>
					</Line>
				),
				({ page, at }) => (
					<Line gap={10}>
						<Islander page={page} who="a" is="liar" />
						<Then />
						<Islander page={page} who="b" is="liar" />
						<Verdict fits>{page.text(`${at}.fits`)}</Verdict>
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Line gap={10}>
						<Said>{page.text(`${at}.alone`)}</Said>
						<Say way="wrong">{page.text(`${at}.nobody`)}</Say>
					</Line>
					<Line gap={10}>
						<Said>{page.text(`${at}.says`)}</Said>
						<Say way="right">{page.text(`${at}.liar-can`)}</Say>
					</Line>
				</Stack>
			),
		},
		{
			task: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Speaker page={page} who="a">
						{page.text(`${at}.a-says`)}
					</Speaker>
					<Speaker page={page} who="b">
						{page.text(`${at}.b-says`)}
					</Speaker>
					<Speaker page={page} who="c">
						{page.text(`${at}.c-says`)}
					</Speaker>
				</Stack>
			),
			steps: [
				({ page, at }) => (
					<>
						<Speaker page={page} who="c" picked>
							{page.text(`${at}.c-says`)}
						</Speaker>
						<Say>{page.text(`${at}.most`)}</Say>
					</>
				),
				({ page, at }) => (
					<Line gap={10}>
						<Islander page={page} who="c" is="knight" />
						<Then />
						<Islander page={page} who="a" is="liar" />
						<Islander page={page} who="b" is="liar" />
						<Verdict fits={false}>{page.text(`${at}.contradiction`)}</Verdict>
					</Line>
				),
				({ page }) => (
					<Line gap={10}>
						<Islander page={page} who="c" is="liar" />
						<Then />
						<Islander page={page} who="b" is="liar" />
					</Line>
				),
				({ page, at }) => (
					<Speaker page={page} who="a" is="knight">
						{page.text(`${at}.a-says`)}
					</Speaker>
				),
			],
			note: ({ page, at }) => (
				<span class="s-kl-check">
					<Checked
						page={page}
						who="a"
						is="knight"
						said={page.text(`${at}.a-says`)}
					>
						{page.text(`${at}.true-knight`)}
					</Checked>
					<Checked
						page={page}
						who="b"
						is="liar"
						said={page.text(`${at}.b-says`)}
					>
						{page.text(`${at}.false-liar`)}
					</Checked>
					<Checked
						page={page}
						who="c"
						is="liar"
						said={page.text(`${at}.c-says`)}
					>
						{page.text(`${at}.false-liar`)}
					</Checked>
				</span>
			),
		},
		{
			steps: [
				({ page, at }) => (
					<Line gap={10}>
						<Sign page={page} is="knight" />
						<Then />
						<Sign page={page} is="liar" />
						<Say>{page.text(`${at}.right-liar`)}</Say>
					</Line>
				),
				({ page, at }) => (
					<Line gap={10}>
						<Sign page={page} is="liar" />
						<Then />
						<Sign page={page} is="knight" />
						<Say>{page.text(`${at}.right-knight`)}</Say>
					</Line>
				),
				({ page, at }) => (
					<>
						<Table
							page={page}
							seats={10}
							middle={
								<>
									<span class="s-kl-half">{page.text(`${at}.half`)}</span>
									{page.text(`${at}.kinds`)}
								</>
							}
						/>
						<Say>{page.text(`${at}.turns`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<>
					<Table page={page} seats={9} />
					<Say way="wrong">{page.text(`${at}.clash`)}</Say>
				</>
			),
		},
	],
};
