import type { ComponentChildren } from "preact";
import { Box, Key, Line, Op, Say, Stack, Versus } from "../Sketch";
import type { TopicArt } from "./art";

// Part is one part of the picture of two groups: those in the left group
// alone, those in both, those in the right one alone, and those in neither.
type Part = "left" | "both" | "right" | "out";

// Ring is how a part's number is ringed: the one a step fills in, a wrong
// answer, or the right one.
type Ring = "picked" | "wrong" | "right";

// Geometry is the circles of a picture of two groups, in its own pixels:
// its width and height, the circles' radius, how far apart their middles
// stand, and how large a number's disc is.
type Geometry = {
	width: number;
	height: number;
	radius: number;
	apart: number;
	disc: number;
};

const geometries = {
	md: { width: 160, height: 100, radius: 44, apart: 40, disc: 11 },
	sm: { width: 100, height: 64, radius: 27, apart: 26, disc: 8 },
	little: { width: 84, height: 50, radius: 21, apart: 42, disc: 0 },
	much: { width: 64, height: 50, radius: 21, apart: 14, disc: 0 },
} as const satisfies Record<string, Geometry>;

// Venn is two groups as circles that overlap, blue on the left and orange
// on the right, the number of each part on a disc in it, ringed as a step
// rings it and pale where it is not the step's; their names over them; and,
// where it counts everyone, the box of all with its name and the number of
// those in neither in its corner.
function Venn({
	shape = "md",
	names,
	counts = {},
	rings = {},
	pale = [],
	all,
	scale,
}: {
	shape?: keyof typeof geometries;
	names?: readonly [ComponentChildren, ComponentChildren];
	counts?: Readonly<Partial<Record<Part, string>>>;
	rings?: Readonly<Partial<Record<Part, Ring>>>;
	pale?: readonly Part[];
	all?: ComponentChildren | true;
	scale?: "hero";
}) {
	const g: Geometry = geometries[shape];
	const middle = g.width / 2;
	const left = middle - g.apart / 2;
	const right = middle + g.apart / 2;
	const y = g.height / 2;
	// Each part's number stands in the middle of its part along the line
	// through both circles' middles.
	const spots: Record<Exclude<Part, "out">, number> = {
		left: middle - g.radius,
		both: middle,
		right: middle + g.radius,
	};
	return (
		<span
			class="s-ovl-venn"
			data-shape={shape}
			data-boxed={all === undefined ? undefined : ""}
			data-scale={scale}
		>
			{names !== undefined && (
				<span class="s-ovl-names">
					<span data-set="left">{names[0]}</span>
					<span data-set="right">{names[1]}</span>
				</span>
			)}
			<svg
				class="s-ovl-circles"
				width={g.width}
				height={g.height}
				viewBox={`0 0 ${g.width} ${g.height}`}
			>
				<circle class="s-ovl-set" data-set="left" cx={left} cy={y} r={g.radius} />
				<circle class="s-ovl-set" data-set="right" cx={right} cy={y} r={g.radius} />
				{(["left", "both", "right"] as const).map((part) => {
					const count = counts[part];
					if (count === undefined) {
						return null;
					}
					return (
						<g
							key={part}
							class="s-ovl-count"
							data-ring={rings[part]}
							data-pale={pale.includes(part) ? "" : undefined}
						>
							<circle cx={spots[part]} cy={y} r={g.disc} />
							<text
								x={spots[part]}
								y={y}
								text-anchor="middle"
								dominant-baseline="central"
								style={{ fontSize: `${g.disc}px` }}
							>
								{count}
							</text>
						</g>
					);
				})}
			</svg>
			{all !== undefined && all !== true && <span class="s-ovl-all">{all}</span>}
			{counts.out !== undefined && (
				<span
					class="s-ovl-out"
					data-ring={rings.out}
					data-pale={pale.includes("out") ? "" : undefined}
				>
					{counts.out}
				</span>
			)}
		</span>
	);
}

// club are the numbers of the key idea's thirty pupils, by part.
const club = { left: "10", both: "4", right: "7", out: "9" } as const;

// Filled is the picture of the key idea's thirty pupils with the part a row
// finds ringed and the others pale.
function Filled({ part }: { part: Part }) {
	return (
		<Venn
			shape="sm"
			counts={club}
			rings={{ [part]: "picked" }}
			pale={(["left", "both", "right", "out"] as const).filter((one) => one !== part)}
			all
		/>
	);
}

// Kids are the children of a group in a row of cells, with the bar of those
// who like apples over them from the left and the bar of those who like
// pears under them; those who like both violet, framed where a step finds
// them.
function Kids({
	names,
	apples,
	pears,
	both,
	framed = false,
}: {
	names: readonly [ComponentChildren, ComponentChildren, ComponentChildren];
	apples: readonly [number, number];
	pears: readonly [number, number];
	both: readonly [number, number];
	framed?: boolean;
}) {
	return (
		<span class="s-ovl-kids">
			<span class="s-ovl-row-name" data-set="left" style={{ gridRow: "1" }}>
				{names[0]}
			</span>
			<span
				class="s-ovl-bar"
				data-set="left"
				style={{ gridRow: "1", gridColumn: `${apples[0] + 1} / ${apples[1] + 2}` }}
			/>
			<span class="s-ovl-row-name" style={{ gridRow: "2" }}>
				{names[1]}
			</span>
			{Array.from({ length: 25 }, (_, kid) => {
				const inBoth = kid + 1 >= both[0] && kid + 1 <= both[1];
				return (
					<span
						key={kid}
						class="s-ovl-kid"
						data-both={inBoth ? "" : undefined}
						data-framed={inBoth && framed ? "" : undefined}
						style={{ gridRow: "2", gridColumn: `${kid + 2}` }}
					/>
				);
			})}
			<span class="s-ovl-row-name" data-set="right" style={{ gridRow: "3" }}>
				{names[2]}
			</span>
			<span
				class="s-ovl-bar"
				data-set="right"
				style={{ gridRow: "3", gridColumn: `${pears[0] + 1} / ${pears[1] + 2}` }}
			/>
		</span>
	);
}

// Three are three subjects as three circles that overlap, each named with
// how many like it, and the number in the middle, those who like all three.
function Three({
	names,
}: {
	names: readonly [ComponentChildren, ComponentChildren, ComponentChildren];
}) {
	return (
		<span class="s-ovl-three">
			<span class="s-ovl-three-name" data-set="left">
				{names[0]}
			</span>
			<span class="s-ovl-three-name" data-set="right">
				{names[1]}
			</span>
			<svg class="s-ovl-circles" width={150} height={136} viewBox="0 0 150 136">
				<circle class="s-ovl-set" data-set="left" cx={50} cy={46} r={40} />
				<circle class="s-ovl-set" data-set="right" cx={100} cy={46} r={40} />
				<circle class="s-ovl-set" data-set="third" cx={75} cy={90} r={40} />
				<g class="s-ovl-count">
					<circle cx={75} cy={64} r={10} />
					<text
						x={75}
						y={64}
						text-anchor="middle"
						dominant-baseline="central"
						style={{ fontSize: "10px" }}
					>
						0
					</text>
				</g>
			</svg>
			<span class="s-ovl-three-name" data-set="third">
				{names[2]}
			</span>
		</span>
	);
}

// Tally is a pupil as a cell, with a dot for each subject they like: blue
// for one, orange for two.
function Tally({ likes }: { likes: 1 | 2 }) {
	return (
		<span class="s-ovl-tally" data-likes={likes}>
			<span />
			{likes === 2 && <span />}
		</span>
	);
}

/**
 * art are the drawings of overlapping groups: two circles that overlap, the
 * number of each part on a disc in it and those in neither in the corner of
 * the box of all, filled in from the middle out; the children of a group in
 * a row, those who like apples and those who like pears; three circles; and
 * pupils counted by the subjects they like.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Venn
				names={[page.text(`${at}.chess`), page.text(`${at}.football`)]}
				counts={{ left: "7", both: "3", right: "9" }}
				scale="hero"
			/>
			<Say>{page.text(`${at}.middle`)}</Say>
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={8}>
				<Venn
					names={[page.text(`${at}.choir`), page.text(`${at}.art`)]}
					counts={{ both: "2×" }}
				/>
				<Say>{page.text(`${at}.twice`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Venn
					counts={club}
					rings={{ out: "picked" }}
					pale={["left", "both", "right"]}
					all={page.text(`${at}.class`)}
				/>
				<Say>{page.text(`${at}.none`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Line gap={20} align="end">
				<Stack gap={6}>
					<Venn shape="little" />
					<Say>{page.text(`${at}.little`)}</Say>
				</Stack>
				<Stack gap={6}>
					<Venn shape="much" />
					<Say>{page.text(`${at}.much`)}</Say>
				</Stack>
			</Line>
		),
	],
	idea: [
		{
			column: 0,
			words: "beside",
			rows: [
				() => <Filled part="both" />,
				() => <Filled part="left" />,
				() => <Filled part="right" />,
				() => <Filled part="out" />,
			],
		},
	],
	ideaAnswers: 1,
	legend: ({ page, at }) => (
		<Key sample={<span class="s-ovl-out" data-ring="picked">6</span>}>
			{page.text(`${at}.filled`)}
		</Key>
	),
	examples: [
		{
			steps: [
				({ page, at }) => (
					<Venn
						names={[page.text(`${at}.sister`), page.text(`${at}.brother`)]}
						counts={{ both: "6" }}
						rings={{ both: "picked" }}
						all={page.text(`${at}.all`)}
					/>
				),
				({ page, at }) => (
					<>
						<Venn
							names={[page.text(`${at}.sister`), page.text(`${at}.brother`)]}
							counts={{ left: "9", both: "6", right: "5" }}
							rings={{ left: "picked", right: "picked" }}
							pale={["both"]}
							all={page.text(`${at}.all`)}
						/>
						<Say>{page.text(`${at}.only`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Venn
							names={[page.text(`${at}.sister`), page.text(`${at}.brother`)]}
							counts={{ left: "9", both: "6", right: "5", out: "4" }}
							rings={{ out: "picked" }}
							pale={["left", "both", "right"]}
							all={page.text(`${at}.all`)}
						/>
						<Say>{page.text(`${at}.out`)}</Say>
					</>
				),
			],
		},
		{
			steps: [
				({ page, at }) => (
					<>
						<Kids
							names={[
								page.text(`${at}.apples`),
								page.text(`${at}.kids`),
								page.text(`${at}.pears`),
							]}
							apples={[1, 18]}
							pears={[10, 25]}
							both={[10, 18]}
						/>
						<Say>{page.text(`${at}.ends`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Kids
							names={[
								page.text(`${at}.apples`),
								page.text(`${at}.kids`),
								page.text(`${at}.pears`),
							]}
							apples={[1, 18]}
							pears={[10, 25]}
							both={[10, 18]}
							framed
						/>
						<Say>{page.text(`${at}.violet`)}</Say>
					</>
				),
				({ page, at }) => (
					<Venn
						names={[page.text(`${at}.apples`), page.text(`${at}.pears`)]}
						counts={{ left: "9", both: "9", right: "7", out: "0" }}
						rings={{ both: "picked" }}
						pale={["left", "right", "out"]}
						all={page.text(`${at}.all`)}
					/>
				),
			],
			note: ({ page, at }) => (
				<>
					<Kids
						names={[
							page.text(`${at}.apples`),
							page.text(`${at}.kids`),
							page.text(`${at}.pears`),
						]}
						apples={[1, 18]}
						pears={[1, 16]}
						both={[1, 16]}
					/>
					<Say>{page.text(`${at}.inside`)}</Say>
				</>
			),
		},
		{
			task: ({ page, at }) => (
				<Three
					names={[
						page.text(`${at}.maths`),
						page.text(`${at}.reading`),
						page.text(`${at}.drawing`),
					]}
				/>
			),
			steps: [
				({ page, at }) => (
					<Line gap={14}>
						<Stack gap={4}>
							<Tally likes={1} />
							<Say>{page.text(`${at}.once`)}</Say>
						</Stack>
						<Stack gap={4}>
							<Tally likes={2} />
							<Say>{page.text(`${at}.twice`)}</Say>
						</Stack>
						<span class="s-ovl-sum">12 + 10 + 9 = 31</span>
					</Line>
				),
				({ page, at }) => (
					<>
						<Line gap={3}>
							{Array.from({ length: 20 }, (_, pupil) => (
								<Tally key={pupil} likes={1} />
							))}
						</Line>
						<Say>{page.text(`${at}.each`)}</Say>
					</>
				),
				() => (
					<>
						<Line gap={3}>
							{Array.from({ length: 9 }, (_, pupil) => (
								<Tally key={`one-${pupil}`} likes={1} />
							))}
							<span class="s-ovl-gap" />
							{Array.from({ length: 11 }, (_, pupil) => (
								<Tally key={`two-${pupil}`} likes={2} />
							))}
						</Line>
						<Say>9 + 2 × 11 = 31</Say>
					</>
				),
			],
		},
	],
	traps: {
		double_count: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="xs">15</Box>
							<Op>+</Op>
							<Box size="xs">11</Box>
							<Op>=</Op>
							<Box size="xs" ring="wrong">
								26
							</Box>
						</>
					}
					right={
						<>
							<Box size="xs">15</Box>
							<Op>+</Op>
							<Box size="xs">11</Box>
							<Op>−</Op>
							<Box size="xs">6</Box>
							<Op>=</Op>
							<Box size="xs" ring="right">
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
						<Venn
							shape="sm"
							counts={{ left: "9", both: "6", right: "5", out: "4" }}
							rings={{ both: "wrong" }}
							pale={["left", "right", "out"]}
							all
						/>
					}
					right={
						<Venn
							shape="sm"
							counts={{ left: "9", both: "6", right: "5", out: "4" }}
							rings={{ out: "right" }}
							pale={["left", "both", "right"]}
							all
						/>
					}
				/>
			</>
		),
		missed_case: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">24</Box>
							<Op>−</Op>
							<Box size="sm">15</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								9
							</Box>
						</>
					}
					right={
						<>
							<Box size="sm">24</Box>
							<Op>−</Op>
							<Box size="sm">20</Box>
							<Op>=</Op>
							<Box size="sm" ring="right">
								4
							</Box>
						</>
					}
				/>
			</>
		),
	},
};
