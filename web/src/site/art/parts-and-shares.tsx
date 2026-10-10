import type { ComponentChildren } from "preact";
import type { Bar } from "../../design/picture/model";
import type { Tone, Tones } from "../../design/picture/tones";
import type { PageReader } from "../reader";
import { Box, Key, Line, Op, Say, Stack, Then, Versus } from "../Sketch";
import type { TopicArt } from "./art";
import { Kind } from "./pictures";

// Paint is the tone of each piece of a strip, by its place: the pieces
// given away struck, the ones counted cool or warm, the ones sought green,
// and those left plain.
type Paint = Tone | undefined;

// tonesOf are the tones of a strip's pieces, one paint for each piece.
function tonesOf(paints: readonly Paint[]): Tones {
	return Object.fromEntries(
		paints.flatMap((paint, at) =>
			paint === undefined ? [] : [[`piece 0.${at}`, [paint]]],
		),
	);
}

// Strip is a whole as a strip of a kind the card draws, cut into pieces of
// their own sizes and labels, each painted as given, with the braces under
// it, the length of the whole and what stands after its end.
function Strip({
	page,
	sizes,
	labels = [],
	paints = [],
	braces,
	span,
	value,
}: {
	page: PageReader;
	sizes: readonly number[];
	labels?: readonly (string | undefined)[];
	paints?: readonly Paint[];
	braces?: Bar["braces"];
	span?: string;
	value?: string;
}) {
	return (
		<Kind
			locale={page.locale}
			picture={{
				kind: "bars",
				bars: [
					{
						segments: sizes.map((size, at) => ({ size, label: labels[at] })),
						braces,
						span,
						value,
					},
				],
			}}
			tones={tonesOf(paints)}
		/>
	);
}

// equal are so many equal pieces of a size, each labelled with it where a
// label is given.
function equal(count: number, size: number, label?: string) {
	return {
		sizes: Array.from({ length: count }, () => size),
		labels: Array.from({ length: count }, () => label),
	};
}

// Swatch is a sample of a paint of the strips, for a key.
function Swatch({ paint }: { paint: "struck" | "green" | "cool" | "warm" }) {
	return <span class="s-pts-swatch" data-paint={paint} />;
}

// Egg is an egg, orange where the buyer takes it and pale where it stays.
function Egg({ taken = false }: { taken?: boolean }) {
	return <span class="s-pts-egg" data-taken={taken ? "" : undefined} />;
}

// Eggs are so many eggs, the first so many taken.
function Eggs({ count, taken }: { count: number; taken: number }) {
	return (
		<Line gap={2}>
			{Array.from({ length: count }, (_, at) => (
				<Egg key={at} taken={at < taken} />
			))}
		</Line>
	);
}

// Chain is the eggs Granny has before each buyer and after the last, as
// boxes joined by arrows, forwards or back, the one a step finds ringed.
function Chain({
	values,
	back = false,
	ring,
}: {
	values: readonly string[];
	back?: boolean;
	ring?: { at: number; way: "picked" | "right" };
}) {
	return (
		<Line gap={4}>
			{values.map((value, at) => [
				at > 0 && <Then key={`then-${at}`} back={back} />,
				<Box
					key={value}
					size="sm"
					ring={ring?.at === at ? ring.way : undefined}
				>
					{value}
				</Box>,
			])}
		</Line>
	);
}

// Buyer is what one buyer finds and takes: the eggs, taken orange, and the
// words of the step beside them.
function Buyer({
	page,
	at,
	who,
	count,
	taken,
}: {
	page: PageReader;
	at: string;
	who: string;
	count: number;
	taken: number;
}) {
	return (
		<Stack gap={6} align="start">
			<Say>{page.text(`${at}.${who}`)}</Say>
			<Line gap={8}>
				<Eggs count={count} taken={taken} />
				<Say>{page.text(`${at}.${who}-took`)}</Say>
			</Line>
		</Stack>
	);
}

// Big is a sum written out large.
function Big({ children }: { children: ComponentChildren }) {
	return <span class="s-pts-big">{children}</span>;
}

/**
 * art are the drawings of parts and shares: wholes as strips of a kind the
 * card draws, cut into equal parts, the parts given away struck, those
 * counted cool or warm and those sought green, with braces under them; and
 * the eggs of a market, worked backwards.
 */
export const art: TopicArt = {
	pictures: true,
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Strip
				page={page}
				{...equal(4, 3, "3")}
				paints={["cool", "cool", "cool"]}
				braces={[{ from: 0, to: 3, label: "9" }]}
			/>
			<Big>{page.text(`${at}.sum`)}</Big>
			<Say>{page.text(`${at}.one`)}</Say>
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={6} align="start">
				<Say>{page.text(`${at}.whole`)}</Say>
				<Strip
					page={page}
					{...equal(4, 1.6)}
					labels={["1/4"]}
					paints={["cool"]}
				/>
				<Say>{page.text(`${at}.rest`)}</Say>
				<Strip
					page={page}
					sizes={[3.2, 0.8, 0.8, 0.8, 0.8]}
					paints={["struck", "warm"]}
				/>
			</Stack>
		),
		({ page }) => (
			<Strip
				page={page}
				{...equal(4, 5, "5")}
				paints={["cool", "cool", "cool"]}
				braces={[{ from: 0, to: 3, label: "15" }]}
				span="20"
			/>
		),
		({ page }) => (
			<Strip
				page={page}
				{...equal(12, 1)}
				paints={[
					...Array<Paint>(4).fill("cool"),
					...Array<Paint>(3).fill("warm"),
					...Array<Paint>(5).fill("green"),
				]}
				braces={[
					{ from: 0, to: 4, label: "1/3" },
					{ from: 4, to: 7, label: "1/4" },
					{ from: 7, to: 12, label: "5/12" },
				]}
			/>
		),
	],
	idea: [
		{
			column: 1,
			words: "under",
			rows: [
				({ page }) => (
					<Strip
						page={page}
						{...equal(4, 5, "5")}
						paints={["cool", "cool", "cool"]}
						braces={[{ from: 0, to: 3, label: "15" }]}
						span="20"
					/>
				),
				({ page }) => (
					<Strip
						page={page}
						{...equal(4, 5, "5")}
						paints={["cool", "cool", "cool", "green"]}
						braces={[{ from: 0, to: 3, label: "15" }]}
						span="20"
					/>
				),
				({ page }) => (
					<Strip
						page={page}
						{...equal(5, 7, "7")}
						paints={["cool", "cool", "green", "green", "green"]}
						braces={[{ from: 0, to: 2, label: "14" }]}
						span="35"
					/>
				),
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key sample={<Swatch paint="struck" />}>{page.text(`${at}.given`)}</Key>
			<Key sample={<Swatch paint="green" />}>{page.text(`${at}.sought`)}</Key>
		</>
	),
	examples: [
		{
			task: ({ page }) => <Strip page={page} sizes={[8]} labels={["60"]} />,
			steps: [
				({ page, at }) => (
					<>
						<Strip
							page={page}
							sizes={[30, 30]}
							labels={["30", "30"]}
							paints={["struck", "cool"]}
							braces={[{ from: 1, to: 2, label: "30" }]}
						/>
						<Say>{page.text(`${at}.rest`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Strip
							page={page}
							{...equal(3, 10, "10")}
							paints={["warm", "warm", "cool"]}
							braces={[{ from: 0, to: 2, label: "20" }]}
						/>
						<Say>{page.text(`${at}.evening`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Strip
							page={page}
							{...equal(6, 10, "10")}
							paints={["struck", "struck", "struck", "warm", "warm", "green"]}
							braces={[
								{ from: 0, to: 3, label: "30" },
								{ from: 3, to: 5, label: "20" },
								{ from: 5, to: 6, label: "10" },
							]}
						/>
						<Say>{page.text(`${at}.all`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<>
					<Strip
						page={page}
						sizes={[30, 30]}
						labels={["30", "30"]}
						paints={["struck", "warm"]}
						value="+10"
					/>
					<Say way="wrong">{page.text(`${at}.over`)}</Say>
				</>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={12}>
					<Strip page={page} {...equal(7, 1)} />
					<Say>{page.text(`${at}.sevenths`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Strip
							page={page}
							{...equal(7, 1)}
							paints={["cool", "cool", "warm", "warm", "warm"]}
							braces={[
								{ from: 0, to: 5, label: "5/7" },
								{ from: 5, to: 7, label: "2/7" },
							]}
						/>
						<Line gap={14}>
							<Key sample={<Swatch paint="cool" />}>
								{page.text(`${at}.monday`)}
							</Key>
							<Key sample={<Swatch paint="warm" />}>
								{page.text(`${at}.tuesday`)}
							</Key>
						</Line>
					</>
				),
				({ page }) => (
					<Strip
						page={page}
						{...equal(7, 1)}
						labels={[
							undefined,
							undefined,
							undefined,
							undefined,
							undefined,
							"20",
							"20",
						]}
						paints={[...Array<Paint>(5).fill("struck"), "green", "green"]}
						braces={[{ from: 5, to: 7, label: "40" }]}
					/>
				),
				({ page }) => (
					<>
						<Strip
							page={page}
							{...equal(7, 1, "20")}
							paints={Array<Paint>(7).fill("green")}
							braces={[{ from: 0, to: 7, label: "140" }]}
						/>
						<Say>7 × 20 = 140</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<>
					<Strip
						page={page}
						{...equal(7, 1, "8")}
						paints={Array<Paint>(5).fill("struck")}
						braces={[{ from: 0, to: 5, label: "40" }]}
					/>
					<Say way="wrong">{page.text(`${at}.read`)}</Say>
				</>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={10}>
					<Eggs count={3} taken={0} />
					<Op>…</Op>
					<Say>{page.text(`${at}.basket`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Buyer page={page} at={at} who="third" count={1} taken={1} />
				),
				({ page, at }) => (
					<Buyer page={page} at={at} who="second" count={3} taken={2} />
				),
				({ page, at }) => (
					<Stack gap={10}>
						<Buyer page={page} at={at} who="first" count={7} taken={4} />
						<Chain
							values={["7", "3", "1", "0"]}
							ring={{ at: 0, way: "picked" }}
						/>
					</Stack>
				),
			],
			note: ({ page, at }) => (
				<Line gap={10}>
					<Chain
						values={["0", "1", "3", "7"]}
						back
						ring={{ at: 3, way: "right" }}
					/>
					<Say>{page.text(`${at}.back`)}</Say>
				</Line>
			),
		},
	],
	traps: {
		part_whole_swap: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Say>{page.text(`${at}.of-all`)}</Say>
							<Box size="sm" ring="wrong">
								40
							</Box>
						</>
					}
					right={
						<>
							<Say>{page.text(`${at}.of-rest`)}</Say>
							<Box size="sm" ring="right">
								20
							</Box>
						</>
					}
				/>
			</>
		),
		answered_other_question: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm" ring="wrong">
								20
							</Box>
							<Say>{page.text(`${at}.evening`)}</Say>
						</>
					}
					right={
						<>
							<Box size="sm" ring="right">
								10
							</Box>
							<Say>{page.text(`${at}.left`)}</Say>
						</>
					}
				/>
			</>
		),
		stopped_early: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm" tone="red">
								20
							</Box>
							<span class="s-pts-total">= 20</span>
						</>
					}
					right={
						<>
							<span class="s-pts-book">
								{Array.from({ length: 7 }, (_, at) => (
									<span key={at} />
								))}
							</span>
							<span class="s-pts-total">= 140</span>
						</>
					}
				/>
			</>
		),
	},
};
