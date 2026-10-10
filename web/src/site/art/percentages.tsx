import type { ComponentChildren } from "preact";
import type { Tone } from "../../design/picture/tones";
import type { PageReader } from "../reader";
import {
	Box,
	Key,
	Line,
	Op,
	Say,
	Stack,
	Stretch,
	Then,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";
import { Kind } from "./pictures";

// Hundred is a hundred cells, a hundred percent, filled in reading order:
// so many in the accent, then so many orange, then so many blue, and the
// rest left grey; framed in the accent where it is the whole a step names.
function Hundred({
	accent = 0,
	orange = 0,
	blue = 0,
	framed = false,
	size = "md",
}: {
	accent?: number;
	orange?: number;
	blue?: number;
	framed?: boolean;
	size?: "md" | "sm";
}) {
	const paintOf = (cell: number) => {
		if (cell < accent) {
			return "accent";
		}
		if (cell < accent + orange) {
			return "orange";
		}
		if (cell < accent + orange + blue) {
			return "blue";
		}
		return undefined;
	};
	return (
		<span
			class="s-pct-hundred"
			data-framed={framed ? "" : undefined}
			data-size={size}
		>
			{Array.from({ length: 100 }, (_, cell) => (
				<span key={cell} data-paint={paintOf(cell)} />
			))}
		</span>
	);
}

// Tenths are a whole as a strip of a kind the card draws, cut into ten equal
// parts each worth a value, the first so many lit, with the brace under them
// and what they make.
function Tenths({
	page,
	value,
	lit,
	total,
}: {
	page: PageReader;
	value: string;
	lit: number;
	total: string;
}) {
	return (
		<Kind
			locale={page.locale}
			picture={{
				kind: "bars",
				bars: [
					{
						segments: Array.from({ length: 10 }, () => ({
							size: 1,
							label: value,
						})),
						braces: [{ from: 0, to: lit, label: total }],
					},
				],
			}}
			tones={Object.fromEntries(
				Array.from({ length: lit }, (_, at) => [
					`piece 0.${at}`,
					["cool"] as Tone[],
				]),
			)}
		/>
	);
}

// Price is a price as a strip to one scale, a price of 200 drawn 220 px
// long: what is kept, blue, and what is taken off, hatched red, each with
// its number.
function Price({ kept, off }: { kept: number; off: number }) {
	return (
		<span
			class="s-pct-price"
			style={{ "--s-length": `${((kept + off) / 200) * 220}px` }}
		>
			<span class="s-pct-kept" style={{ flexGrow: kept }}>
				{kept}
			</span>
			<span class="s-pct-off" style={{ flexGrow: off }}>
				−{off}
			</span>
		</span>
	);
}

// Fifths are a price as a strip of a kind the card draws in parts of 100,
// the last one added, green, or taken off, struck.
function Fifths({ page, last }: { page: PageReader; last: "added" | "taken" }) {
	return (
		<Kind
			locale={page.locale}
			picture={{
				kind: "bars",
				bars: [
					{
						segments: [
							...Array.from({ length: 4 }, () => ({ size: 2, label: "100" })),
							{ size: 2, label: last === "added" ? "+100" : "−100" },
						],
					},
				],
			}}
			tones={{
				"piece 0.0": ["cool"],
				"piece 0.1": ["cool"],
				"piece 0.2": ["cool"],
				"piece 0.3": ["cool"],
				"piece 0.4": [last === "added" ? "green" : "struck"],
			}}
		/>
	);
}

// Melon is a watermelon.
function Melon() {
	return <span class="s-pct-melon" />;
}

// Water is the weight of the melons as a strip to one scale: the water,
// blue, and what is not water, an orange sliver, with its percentage after
// the strip.
function Water({
	length,
	water,
	dry,
}: {
	length: number;
	water: ComponentChildren;
	dry: string;
}) {
	return (
		<Line gap={6}>
			<span class="s-pct-water" style={{ "--s-length": `${length}px` }}>
				<span>{water}</span>
			</span>
			<span class="s-pct-dry">{dry}</span>
		</Line>
	);
}

// Colour is a percentage in the colour of its cells, and what it counts.
function Colour({
	paint,
	value,
	children,
}: {
	paint: "orange" | "blue";
	value: string;
	children: ComponentChildren;
}) {
	return (
		<span class="s-pct-colour">
			<strong data-paint={paint}>{value}</strong> {children}
		</span>
	);
}

// Step is a change of the price on its arrow: up, green, or down, red.
function Step({ children, up }: { children: ComponentChildren; up: boolean }) {
	return (
		<span class="s-pct-step" data-up={up ? "" : undefined}>
			<Stretch length={34}>{children}</Stretch>
		</span>
	);
}

/**
 * art are the drawings of percentages: a hundred cells as a hundred percent,
 * wholes as strips of a kind the card draws cut into tenths or into parts of
 * 100, prices as strips to one scale with what is taken off, and the water
 * of the melons.
 */
export const art: TopicArt = {
	pictures: true,
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page }) => (
		<>
			<Tenths page={page} value="5" lit={3} total="15" />
			<span class="s-pct-facts">
				<span>100% = 50</span>
				<span>10% = 5</span>
				<span data-lit="">30% = 15</span>
			</span>
		</>
	),
	basis: [
		({ page, at }) => (
			<Line gap={16} align="start">
				<Stack gap={6}>
					<Hundred accent={1} />
					<Say>1%</Say>
				</Stack>
				<Stack gap={6}>
					<Hundred blue={50} />
					<Say>{page.text(`${at}.half`)}</Say>
				</Stack>
			</Line>
		),
		({ page, at }) => (
			<Stack gap={6}>
				<Kind
					locale={page.locale}
					picture={{
						kind: "bars",
						bars: [
							{ parts: 10, length: 200 },
							{ parts: 10, length: 50 },
						],
					}}
					tones={{ "piece 0.0": ["cool"], "piece 1.0": ["cool"] }}
				/>
				<Say>{page.text(`${at}.tens`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={6}>
				<Say>{page.text(`${at}.eighty`)}</Say>
				<Tenths page={page} value="8" lit={3} total="24" />
				<Say>30% = 3 × 8 = 24</Say>
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			words: "under",
			rows: [
				() => <Price kept={160} off={40} />,
				() => <Price kept={128} off={32} />,
				() => <Price kept={128} off={72} />,
			],
		},
	],
	legend: ({ page, at }) => (
		<Key sample={<Hundred size="sm" />}>{page.text(`${at}.hundred`)}</Key>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={14} align="end">
					<Hundred orange={35} />
					<Say>{page.text(`${at}.go`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Line gap={14}>
						<Hundred accent={1} framed />
						<Stack gap={2} align="start">
							<Say>{page.text(`${at}.grid`)}</Say>
							<Say>{page.text(`${at}.cell`)}</Say>
						</Stack>
					</Line>
				),
				({ page, at }) => (
					<Line gap={14}>
						<Hundred orange={35} blue={65} />
						<Stack gap={4} align="start">
							<Colour paint="orange" value="35%">
								{page.text(`${at}.go-short`)}
							</Colour>
							<Colour paint="blue" value="65%">
								{page.text(`${at}.do-not`)}
							</Colour>
						</Stack>
					</Line>
				),
				() => (
					<Line gap={8}>
						<Box size="sm" ring="picked">
							65
						</Box>
						<Op>×</Op>
						<Box size="sm">3</Box>
						<Op>=</Op>
						<Box size="sm">195</Box>
					</Line>
				),
			],
			note: () => (
				<Line gap={6}>
					<Box size="sm">35 × 3</Box>
					<Op>=</Op>
					<Box size="sm">105</Box>
					<Then />
					<Box size="sm">300 − 105</Box>
					<Op>=</Op>
					<Box size="sm" ring="right">
						195
					</Box>
				</Line>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={8}>
					<Box size="sm">400</Box>
					<Say>{page.text(`${at}.coins`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Stack gap={6} align="start">
						<Say>{page.text(`${at}.quarters`)}</Say>
						<Fifths page={page} last="added" />
						<Say>{page.text(`${at}.five-hundred`)}</Say>
					</Stack>
				),
				({ page, at }) => (
					<Stack gap={6} align="start">
						<Say>{page.text(`${at}.fifths`)}</Say>
						<Fifths page={page} last="taken" />
					</Stack>
				),
				() => (
					<Line gap={6}>
						<Box size="sm">400</Box>
						<Step up>+25%</Step>
						<Box size="sm">500</Box>
						<Step up={false}>−20%</Step>
						<Box size="sm" ring="picked">
							400
						</Box>
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Line gap={6}>
						<Say>{page.text(`${at}.of-old`)}</Say>
						<Box size="sm">500 − 80</Box>
						<Op>=</Op>
						<Box size="sm" ring="wrong">
							420
						</Box>
					</Line>
					<Line gap={6}>
						<Say>{page.text(`${at}.of-new`)}</Say>
						<Box size="sm">500 − 100</Box>
						<Op>=</Op>
						<Box size="sm" ring="right">
							400
						</Box>
					</Line>
				</Stack>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={10}>
					<Melon />
					<Say>{page.text(`${at}.melons`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Line gap={6}>
							<Hundred orange={1} blue={99} />
							<Hundred orange={1} blue={99} />
						</Line>
						<Say>{page.text(`${at}.dry`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Hundred orange={2} blue={98} framed />
						<Say>{page.text(`${at}.same`)}</Say>
					</>
				),
				({ page, at }) => (
					<Line gap={12}>
						<Stack gap={6}>
							<Line gap={4}>
								<Hundred orange={1} blue={99} size="sm" />
								<Hundred orange={1} blue={99} size="sm" />
							</Line>
							<Say>{page.text(`${at}.was`)}</Say>
						</Stack>
						<Then />
						<Stack gap={6}>
							<Hundred orange={2} blue={98} size="sm" />
							<Say>{page.text(`${at}.became`)}</Say>
						</Stack>
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={6} align="start">
					<Water length={200} water={page.text(`${at}.water`)} dry="1%" />
					<Water length={100} water="98%" dry="2%" />
				</Stack>
			),
		},
	],
	traps: {
		percent_wrong_base: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Say>{page.text(`${at}.of-old`)}</Say>
							<Box size="sm" ring="wrong">
								80
							</Box>
						</>
					}
					right={
						<>
							<Say>{page.text(`${at}.of-new`)}</Say>
							<Box size="sm" ring="right">
								100
							</Box>
						</>
					}
				/>
			</>
		),
		part_whole_swap: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm" ring="wrong">
								{page.text(`${at}.dry`)}
							</Box>
							<Say>{page.text(`${at}.only-dry`)}</Say>
						</>
					}
					right={
						<>
							<Box size="sm" ring="right">
								{page.text(`${at}.all`)}
							</Box>
							<Say>{page.text(`${at}.weight`)}</Say>
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
								105
							</Box>
							<Say>{page.text(`${at}.go`)}</Say>
						</>
					}
					right={
						<>
							<Box size="sm" ring="right">
								195
							</Box>
							<Say>{page.text(`${at}.do-not`)}</Say>
						</>
					}
				/>
			</>
		),
	},
};
