import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import {
	Box,
	Chip,
	Key,
	Line,
	Op,
	Round,
	Row,
	Say,
	Stack,
	Then,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";

// Child is a boy or a girl as a dot of their colour, ringed in dashed red
// where two of a kind stand side by side.
function Child({
	girl = false,
	clash = false,
	size = "md",
}: {
	girl?: boolean;
	clash?: boolean;
	size?: "md" | "sm";
}) {
	return (
		<span
			class="s-par-child"
			data-girl={girl ? "" : undefined}
			data-clash={clash ? "" : undefined}
			data-size={size}
		/>
	);
}

// Circle is so many children round a circle, boys and girls taking turns
// from the top clockwise; where their number is odd, the last and the first
// are both boys, side by side, ringed.
function Circle({ count }: { count: number }) {
	return (
		<Round size={96} track="line">
			{Array.from({ length: count }, (_, at) => (
				<Child
					key={at}
					girl={at % 2 === 1}
					clash={count % 2 === 1 && (at === 0 || at === count - 1)}
				/>
			))}
		</Round>
	);
}

// Value is a number on a box: blue where it is even, orange where it is
// odd; ringed as a step rings it.
function Value({
	value,
	ring,
	size = "sm",
}: {
	value: number;
	ring?: "picked" | "wrong" | "right";
	size?: "md" | "sm" | "xs" | "xxs";
}) {
	return (
		<Box tone={value % 2 === 0 ? "cool" : "warm"} size={size} ring={ring}>
			{value}
		</Box>
	);
}

// Sum is numbers added up, each on its box, and what they make where it is
// given.
function Sum({
	values,
	total,
	ring,
	size = "sm",
}: {
	values: readonly number[];
	total?: number;
	ring?: "picked" | "wrong" | "right";
	size?: "md" | "sm" | "xs";
}) {
	return (
		<Line gap={4}>
			{values.map((value, at) => [
				at > 0 && <Op key={`plus-${at}`}>+</Op>,
				<Value key={at} value={value} size={size} />,
			])}
			{total !== undefined && (
				<>
					<Op>=</Op>
					<Value value={total} size={size} ring={ring} />
				</>
			)}
		</Line>
	);
}

// Parity is the word for even or for odd on its pill: blue for even, amber
// for odd.
function Parity({ page, even }: { page: PageReader; even: boolean }) {
	return (
		<span class="s-par-parity" data-even={even ? "" : undefined}>
			{page.text(even ? "art.even" : "art.odd")}
		</span>
	);
}

// Pairs are so many counters set out in pairs, the one left over, if any,
// apart and ringed.
function Pairs({ count }: { count: number }) {
	return (
		<Line gap={4}>
			{Array.from({ length: Math.floor(count / 2) }, (_, pair) => (
				<span key={pair} class="s-par-pair">
					<span class="s-par-dot" />
					<span class="s-par-dot" />
				</span>
			))}
			{count % 2 === 1 && <span class="s-par-dot" data-alone="" />}
		</Line>
	);
}

// River is a river between its two banks with the ferry at one of them,
// ringed red where it is wrongly there and green where it rightly is.
function River({
	at,
	ring,
	size = "md",
}: {
	at: "left" | "right";
	ring?: "wrong" | "right";
	size?: "md" | "sm";
}) {
	return (
		<span class="s-par-river" data-size={size}>
			<span class="s-par-ferry" data-at={at} data-ring={ring} />
		</span>
	);
}

// Target is the target of the darts, its rings worth 1, 3, 5 and 7 from the
// outside in.
function Target() {
	return (
		<svg
			aria-hidden="true"
			class="s-par-target"
			width={90}
			height={90}
			viewBox="0 0 90 90"
		>
			{[42, 32, 22].map((radius) => (
				<circle key={radius} class="s-par-ring" cx={45} cy={45} r={radius} />
			))}
			<circle class="s-par-bull" cx={45} cy={45} r={10} />
			{[
				["1", 8],
				["3", 18],
				["5", 28],
			].map(([value, y]) => (
				<text
					key={value}
					class="s-par-score"
					x={45}
					y={y}
					text-anchor="middle"
					dominant-baseline="central"
				>
					{value}
				</text>
			))}
			<text
				class="s-par-score"
				data-bull=""
				x={45}
				y={45}
				text-anchor="middle"
				dominant-baseline="central"
			>
				7
			</text>
		</svg>
	);
}

// Dart is one dart flying at the target.
function Dart() {
	return <span class="s-par-dart" />;
}

// Corner is a corner of the cube, white or black, or still uncoloured.
type Corner = "white" | "black" | "plain";

// corners are the cube's corners by their places across, up and back, each
// drawn at its point, the front square first.
const corners = [0, 1].flatMap((back) =>
	[0, 1].flatMap((up) =>
		[0, 1].map((across) => ({
			across,
			up,
			back,
			x: 10 + 60 * across + 30 * back,
			y: 100 - 60 * up - 30 * back,
		})),
	),
);

// colourOf is the colour of a corner of the coloured cube: white and black by
// turns, the start white.
function colourOf(one: (typeof corners)[number]): Corner {
	return (one.across + one.up + one.back) % 2 === 1 ? "white" : "black";
}

// edges are the cube's twelve edges, as pairs of corners one step apart.
const edges = corners.flatMap((one, at) =>
	corners
		.slice(at + 1)
		.flatMap((other) =>
			Math.abs(one.across - other.across) +
				Math.abs(one.up - other.up) +
				Math.abs(one.back - other.back) ===
			1
				? [{ one, other, back: one.back + other.back === 2 }]
				: [],
		),
);

// Cube is the wire cube, its back edges pale, the spider's starting corner
// ringed and named; its corners coloured white and black by turns where it
// is coloured, the start white.
function Cube({
	page,
	coloured = false,
}: {
	page: PageReader;
	coloured?: boolean;
}) {
	return (
		<span class="s-par-cube">
			<span class="s-par-start-name">{page.text("art.start")}</span>
			<svg aria-hidden="true" width={110} height={110} viewBox="0 0 110 110">
				{edges.map(({ one, other, back }) => (
					<line
						key={`${one.x}.${one.y}-${other.x}.${other.y}`}
						class="s-par-edge"
						data-back={back ? "" : undefined}
						x1={one.x}
						y1={one.y}
						x2={other.x}
						y2={other.y}
					/>
				))}
				{corners.map((one) => {
					const start = one.across === 0 && one.up === 1 && one.back === 0;
					const corner: Corner = coloured ? colourOf(one) : "plain";
					return (
						<g key={`${one.x}.${one.y}`}>
							{start && (
								<circle class="s-par-start" cx={one.x} cy={one.y} r={9} />
							)}
							<circle
								class="s-par-corner"
								data-corner={corner}
								cx={one.x}
								cy={one.y}
								r={5.5}
							/>
						</g>
					);
				})}
			</svg>
		</span>
	);
}

// Spot is a corner of the cube as a dot: white or black.
function Spot({ black = false }: { black?: boolean }) {
	return <span class="s-par-spot" data-black={black ? "" : undefined} />;
}

// Mark is a tick or a cross: possible or not.
function Mark({ way }: { way: "right" | "wrong" }) {
	return (
		<span class="s-par-mark" data-way={way}>
			{way === "right" ? "✓" : "✕"}
		</span>
	);
}

// Count is a count written out in bold.
function Count({ children }: { children: ComponentChildren }) {
	return <span class="s-par-count">{children}</span>;
}

// Kinds are the key of the numbers' colours: blue for even, orange for odd.
function Kinds({ page }: { page: PageReader }) {
	return (
		<>
			<Key sample={<Value value={2} size="xxs" />}>{page.text("art.even")}</Key>
			<Key sample={<Value value={3} size="xxs" />}>{page.text("art.odd")}</Key>
		</>
	);
}

// Odds are as many orange dots as a sum has odd numbers.
function Odds({ count }: { count: number }) {
	return (
		<span class="s-par-odds">
			{Array.from({ length: count }, (_, at) => (
				<span key={at} />
			))}
		</span>
	);
}

/**
 * art are the drawings of parity and alternation: boys and girls taking
 * turns round a circle, numbers blue where they are even and orange where
 * they are odd, counters in pairs, a ferry between two banks, the target of
 * the darts, and a wire cube whose corners take turns in colour.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => (
		<>
			{page.text(`${at}.line`)}
			<span class="s-par-key">
				<Key sample={<Child size="sm" />}>{page.text(`${at}.boy`)}</Key>
				<Key sample={<Child girl size="sm" />}>{page.text(`${at}.girl`)}</Key>
			</span>
		</>
	),
	hero: ({ page, at }) => (
		<>
			<Line gap={24} align="start">
				<Stack gap={8}>
					<Circle count={6} />
					<Chip tone="green">{page.text(`${at}.six`)}</Chip>
				</Stack>
				<Stack gap={8}>
					<Circle count={5} />
					<Chip tone="red">{page.text(`${at}.five`)}</Chip>
				</Stack>
			</Line>
			<Say>{page.text(`${at}.even`)}</Say>
		</>
	),
	basis: [
		() => (
			<Stack gap={8} align="start">
				<Row name="6" size="small" gap={4}>
					<Pairs count={6} />
				</Row>
				<Row name="7" size="small" gap={4}>
					<Pairs count={7} />
				</Row>
			</Stack>
		),
		() => (
			<Stack gap={6}>
				<Sum values={[4, 6]} total={10} />
				<Sum values={[3, 5]} total={8} />
				<Sum values={[4, 3]} total={7} />
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Line gap={4}>
					{[1, 2, 3, 4, 5, 6].map((place) => (
						<Stack key={place} gap={4}>
							<span class="s-par-place">{place}</span>
							<Child girl={place % 2 === 0} />
						</Stack>
					))}
				</Line>
				<Say>{page.text(`${at}.places`)}</Say>
			</Stack>
		),
	],
	idea: [
		{
			column: 0,
			rows: [
				() => <Sum values={[4, 6, 8]} size="xs" />,
				() => <Sum values={[3, 5]} size="xs" />,
				() => <Sum values={[3, 5, 7]} size="xs" />,
				() => <Sum values={[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]} size="xs" />,
			],
		},
		{
			column: 1,
			words: "beside",
			rows: [
				() => null,
				() => <Odds count={2} />,
				() => <Odds count={3} />,
				() => <Odds count={5} />,
			],
		},
		{
			column: 2,
			rows: [1, 2, 3, 4].map((row) => ({ page }: { page: PageReader }) => (
				<span class="s-par-parity" data-even={row <= 2 ? "" : undefined}>
					{page.text(`idea.table.rows.${row}.3`)}
				</span>
			)),
		},
	],
	ideaLegend: ({ page }) => <Kinds page={page} />,
	legend: ({ page }) => <Kinds page={page} />,
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={12}>
					<River at="left" />
					<Say>{page.text(`${at}.start`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Stack gap={8} align="start">
						<Row name={page.text(`${at}.none`)} gap={4}>
							<River at="left" size="sm" />
						</Row>
						<Row name={page.text(`${at}.one`)} gap={4}>
							<River at="right" size="sm" />
						</Row>
						<Row name={page.text(`${at}.two`)} gap={4}>
							<River at="left" size="sm" />
						</Row>
					</Stack>
				),
				({ page, at }) => (
					<Line gap={20} align="start">
						<Stack gap={6}>
							<Say>{page.text(`${at}.left`)}</Say>
							<Line gap={4}>
								{[0, 2, 4, 6].map((value) => (
									<Value key={value} value={value} />
								))}
							</Line>
						</Stack>
						<Stack gap={6}>
							<Say>{page.text(`${at}.right`)}</Say>
							<Line gap={4}>
								{[1, 3, 5].map((value) => (
									<Value key={value} value={value} />
								))}
							</Line>
						</Stack>
					</Line>
				),
				() => (
					<Line gap={10}>
						<Value value={7} ring="picked" size="md" />
						<River at="right" />
					</Line>
				),
			],
			note: ({ page }) => (
				<Line gap={10}>
					<Value value={100} size="md" />
					<Parity page={page} even />
					<Then />
					<River at="left" />
				</Line>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={14} align="end">
					<Target />
					<Stack gap={6} align="start">
						<Line gap={6}>
							<Dart />
							<Dart />
							<Dart />
							<Dart />
						</Line>
						<Say>{page.text(`${at}.darts`)}</Say>
					</Stack>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Line gap={12}>
							<Target />
							<Line gap={4}>
								{[1, 3, 5, 7].map((value) => (
									<Value key={value} value={value} />
								))}
							</Line>
						</Line>
						<Say>{page.text(`${at}.rings`)}</Say>
					</>
				),
				({ page }) => (
					<Line gap={8} align="start">
						<Stack gap={4}>
							<span class="s-par-pairing">
								<Value value={7} />
								<Value value={3} />
							</span>
							<Parity page={page} even />
						</Stack>
						<span class="s-par-sign">+</span>
						<Stack gap={4}>
							<span class="s-par-pairing">
								<Value value={5} />
								<Value value={1} />
							</span>
							<Parity page={page} even />
						</Stack>
						<span class="s-par-sign">=</span>
						<Stack gap={4}>
							<Value value={16} />
							<Parity page={page} even />
						</Stack>
					</Line>
				),
				({ page, at }) => (
					<Line gap={12}>
						<Stack gap={4}>
							<Value value={15} ring="wrong" />
							<Parity page={page} even={false} />
						</Stack>
						<Say way="wrong">{page.text(`${at}.only-even`)}</Say>
					</Line>
				),
			],
			note: () => (
				<Stack gap={6}>
					<Sum values={[5, 5, 5]} total={15} ring="right" />
					<Sum values={[7, 7, 1]} total={15} ring="right" />
				</Stack>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={14} align="end">
					<Cube page={page} />
					<Say>{page.text(`${at}.edge`)}</Say>
				</Line>
			),
			steps: [
				({ page }) => <Cube page={page} coloured />,
				({ page, at }) => (
					<>
						<Line gap={8}>
							{[0, 1, 2, 3, 4, 5, 6, 7].map((minute) => (
								<Stack key={minute} gap={4}>
									<span class="s-par-place">{minute}</span>
									<Spot black={minute % 2 === 1} />
								</Stack>
							))}
						</Line>
						<Say>{page.text(`${at}.minutes`)}</Say>
					</>
				),
				({ page, at }) => (
					<Line gap={16}>
						<Stack gap={4}>
							<Say>{page.text("art.start")}</Say>
							<Spot />
						</Stack>
						<Stack gap={4}>
							<Say>{page.text(`${at}.seven`)}</Say>
							<Spot black />
						</Stack>
						<Mark way="wrong" />
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Line gap={8}>
					<Spot />
					<Stack gap={2}>
						<Say>{page.text(`${at}.back`)}</Say>
						<span class="s-par-sign">⇄</span>
					</Stack>
					<Spot black />
					<Then />
					<Stack gap={2}>
						<Say>{page.text(`${at}.eight`)}</Say>
						<Spot />
					</Stack>
					<Mark way="right" />
				</Line>
			),
		},
	],
	traps: {
		wrong_parity: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={<River at="left" ring="wrong" size="sm" />}
					right={<River at="right" ring="right" size="sm" />}
				/>
			</>
		),
		ignored_condition: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Line gap={3}>
								<Value value={5} size="xs" />
								<Value value={5} size="xs" />
								<Value value={5} size="xs" />
							</Line>
							<Count>= 15</Count>
							<Say way="wrong">{page.text(`${at}.three`)}</Say>
						</>
					}
					right={
						<>
							<Line gap={3}>
								<Value value={7} size="xs" />
								<Value value={5} size="xs" />
								<Value value={1} size="xs" />
								<Value value={1} size="xs" />
							</Line>
							<Count>= 14</Count>
							<Say>{page.text(`${at}.always`)}</Say>
						</>
					}
				/>
			</>
		),
		off_by_one: ({ page, at }) => (
			<>
				<Line gap={4} align="start">
					{[0, 1, 2, 3, 4].map((place) => (
						<Child key={place} girl={place % 2 === 1} />
					))}
				</Line>
				<Versus
					child={<Say>{page.text(`${at}.equal`)}</Say>}
					right={
						<>
							<Child size="sm" />
							<Count>3</Count>
							<Child girl size="sm" />
							<Count>2</Count>
						</>
					}
				/>
			</>
		),
	},
};
