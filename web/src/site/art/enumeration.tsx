import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import {
	Box,
	Chip,
	Key,
	Line,
	Op,
	Row,
	Say,
	Stack,
	Then,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";
import { Kind, lit } from "./pictures";

// Digit is one of the digits 1, 2 and 3 on a tile of its own colour.
type Digit = 1 | 2 | 3;

// Tile is a digit on its tile, of a size.
function Tile({
	digit,
	size = "md",
}: {
	digit: Digit;
	size?: "lg" | "md" | "sm";
}) {
	return (
		<span class="s-enum-tile" data-digit={digit} data-size={size}>
			{digit}
		</span>
	);
}

// Tiles are a number written in tiles side by side: faded and ringed in
// dashed red where it is not allowed.
function Tiles({
	digits,
	size = "sm",
	banned = false,
}: {
	digits: readonly Digit[];
	size?: "md" | "sm";
	banned?: boolean;
}) {
	return (
		<span class="s-enum-tiles" data-banned={banned ? "" : undefined}>
			{digits.map((digit, at) => (
				<Tile key={at} digit={digit} size={size} />
			))}
		</span>
	);
}

// Branch is one first digit of the tree and the two numbers it begins: the
// digit, a fork to the two digits that may follow it, and under each the
// number the two make.
function Branch({ first }: { first: Digit }) {
	const seconds = ([1, 2, 3] as const).filter((second) => second !== first);
	return (
		<span class="s-enum-branch">
			<span class="s-enum-root">
				<Tile digit={first} size="lg" />
			</span>
			<span class="s-enum-fork" />
			{seconds.map((second) => (
				<span key={second} class="s-enum-twig">
					<Tile digit={second} />
					<span class="s-enum-stem" />
					<Tiles digits={[first, second]} />
				</span>
			))}
		</span>
	);
}

// Who is one of the three children, A, B and C.
type Who = "a" | "b" | "c";

// Child is a child as a circle of their own colour with their letter: large
// where they lead, small where they help.
function Child({
	page,
	who,
	size = "md",
}: {
	page: PageReader;
	who: Who;
	size?: "lg" | "md" | "sm";
}) {
	return (
		<span class="s-enum-child" data-who={who} data-size={size}>
			{page.text(`art.${who}`)}
		</span>
	);
}

// Children are children side by side, as one case of a list: a row, a pair,
// or a leader and a helper; ringed where the order does not count.
function Children({
	page,
	who,
	size = "md",
	ringed = false,
	leader = false,
}: {
	page: PageReader;
	who: readonly Who[];
	size?: "md" | "sm";
	ringed?: boolean;
	leader?: boolean;
}) {
	return (
		<span class="s-enum-children" data-ringed={ringed ? "" : undefined}>
			{who.map((one, at) => (
				<Child
					key={one}
					page={page}
					who={one}
					size={leader ? (at === 0 ? "md" : "sm") : size}
				/>
			))}
		</span>
	);
}

// Person is Lea or Tom as a circle of their own colour with their initial.
function Person({ page, who }: { page: PageReader; who: "lea" | "tom" }) {
	return (
		<span class="s-enum-person" data-who={who}>
			{page.text(`art.${who}`)}
		</span>
	);
}

// Handshake is two people joined by a line: one handshake, whoever is named
// first.
function Handshake({
	page,
	first,
}: {
	page: PageReader;
	first: "lea" | "tom";
}) {
	const second = first === "lea" ? "tom" : "lea";
	return (
		<span class="s-enum-shake">
			<Person page={page} who={first} />
			<span class="s-enum-hand" />
			<Person page={page} who={second} />
		</span>
	);
}

// Queue is two people standing one behind the other.
function Queue({ page, first }: { page: PageReader; first: "lea" | "tom" }) {
	const second = first === "lea" ? "tom" : "lea";
	return (
		<span class="s-enum-queue">
			<Person page={page} who={first} />
			<Person page={page} who={second} />
		</span>
	);
}

// Paint is one of the three colours of the flags' stripes.
type Paint = "red" | "yellow" | "green";
const paints: readonly Paint[] = ["red", "yellow", "green"];

// Swatch is a stroke of one paint.
function Swatch({ paint }: { paint: Paint }) {
	return <span class="s-enum-paint" data-paint={paint} />;
}

// Flag is a flag of two stripes on its pole, the top one first, or a flag
// still to paint where it has no colours: framed in the accent where a step
// adds it, with its number in its corner where it is counted.
function Flag({
	top,
	bottom,
	framed = false,
	number,
}: {
	top?: Paint;
	bottom?: Paint;
	framed?: boolean;
	number?: number;
}) {
	return (
		<span class="s-enum-flag" data-framed={framed ? "" : undefined}>
			<span class="s-enum-pole" />
			<span class="s-enum-cloth">
				<span data-paint={top} />
				<span data-paint={bottom} />
			</span>
			{number !== undefined && (
				<span class="s-enum-flag-number">{number}</span>
			)}
		</span>
	);
}

// Tops are the flags of one colour on top, under a bracket with the colour's
// name, framed where the step adds them.
function Tops({
	top,
	name,
	framed = false,
}: {
	top: Paint;
	name: ComponentChildren;
	framed?: boolean;
}) {
	return (
		<span class="s-enum-tops">
			<span class="s-enum-tops-flags">
				{paints
					.filter((bottom) => bottom !== top)
					.map((bottom) => (
						<Flag key={bottom} top={top} bottom={bottom} framed={framed} />
					))}
			</span>
			<span class="s-enum-bracket" />
			<Say>{name}</Say>
		</span>
	);
}

// FlagTable is every top and bottom colour, the tops down its side and the
// bottoms across its head: the flag they make where they differ, numbered in
// the order the list wrote them, and a cross where they are one colour.
function FlagTable({ corner }: { corner: ComponentChildren }) {
	let counted = 0;
	return (
		<span class="s-enum-table">
			<span class="s-enum-corner">{corner}</span>
			{paints.map((bottom) => (
				<span key={bottom} class="s-enum-head">
					<Swatch paint={bottom} />
				</span>
			))}
			{paints.map((top) => [
				<span key={top} class="s-enum-head">
					<Swatch paint={top} />
				</span>,
				...paints.map((bottom) => {
					if (top === bottom) {
						return (
							<span key={`${top}-${bottom}`} class="s-enum-cross">
								×
							</span>
						);
					}
					counted += 1;
					return (
						<span key={`${top}-${bottom}`} class="s-enum-cell">
							<Flag top={top} bottom={bottom} number={counted} />
						</span>
					);
				}),
			])}
		</span>
	);
}

// Cup is a cup of a colour, with its handle.
function Cup({ color }: { color: "red" | "blue" | "yellow" }) {
	return <span class="s-enum-cup" data-color={color} />;
}

// Saucer is a saucer of a colour.
function Saucer({ color }: { color: "green" | "violet" }) {
	return <span class="s-enum-saucer" data-color={color} />;
}

// Sets are the three cups across and the two saucers down, each cup on each
// saucer where they meet.
function Sets() {
	const cups = ["red", "blue", "yellow"] as const;
	return (
		<span class="s-enum-sets">
			<span />
			{cups.map((cup) => (
				<Cup key={cup} color={cup} />
			))}
			{(["green", "violet"] as const).map((saucer) => [
				<Saucer key={saucer} color={saucer} />,
				...cups.map((cup) => (
					<span key={`${saucer}-${cup}`} class="s-enum-set">
						<Cup color={cup} />
						<Saucer color={saucer} />
					</span>
				)),
			])}
		</span>
	);
}

// Verdict is whether the order matters, on a pill green for yes and grey for
// no, then why: both read from the cell's own words, "Yes: why", which a
// screen reader reads whole.
function Verdict({
	page,
	row,
	yes,
}: {
	page: PageReader;
	row: number;
	yes: boolean;
}) {
	const [word = "", ...why] = page
		.plain(`idea.table.rows.${row}.2`)
		.split(": ");
	return (
		<Stack gap={4} align="start">
			<Chip tone={yes ? "green" : "grey"}>{word}</Chip>
			<span class="s-enum-why">{why.join(": ")}</span>
		</Stack>
	);
}

// Count is how many cases there are: an equals sign and the number.
function Count({ children }: { children: ComponentChildren }) {
	return <span class="s-enum-count">= {children}</span>;
}

// Tally is how many numbers of a length there are, on a chain of tallies:
// the sum it comes from over it, ringed where a step finds it, and the
// length under it.
function Tally({
	value,
	over,
	under,
}: {
	value: string;
	over?: ComponentChildren;
	under: ComponentChildren;
}) {
	return (
		<Stack gap={4}>
			<span class="s-enum-over">{over}</span>
			<Box size="lg" ring={over === undefined ? undefined : "picked"}>
				{value}
			</Box>
			<Say>{under}</Say>
		</Stack>
	);
}

// Mark is a tick or a cross beside a number: allowed or not.
function Mark({ way }: { way: "right" | "wrong" }) {
	return (
		<span class="s-enum-mark" data-way={way}>
			{way === "right" ? "✓" : "✕"}
		</span>
	);
}

// threeApart are the two-digit numbers whose digits are 3 apart, in order.
const threeApart = [
	"14",
	"25",
	"30",
	"36",
	"41",
	"47",
	"52",
	"58",
	"63",
	"69",
	"74",
	"85",
	"96",
];

// noTwoTwos are the four-digit numbers of 1s and 2s with no two 2s side by
// side, in the order the list writes them.
const noTwoTwos: readonly (readonly Digit[])[] = [
	[1, 1, 1, 1],
	[1, 1, 1, 2],
	[1, 1, 2, 1],
	[1, 2, 1, 1],
	[2, 1, 1, 1],
	[1, 2, 1, 2],
	[2, 1, 1, 2],
	[2, 1, 2, 1],
];

/**
 * art are the drawings of enumeration: a tree of the numbers each first
 * digit begins, lists kept in order with the case they miss, children in
 * rows and pairs, flags in a table of their colours, and digits laid out by
 * a rule.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => (
		<>
			{page.text(`${at}.line`)}
			<span class="s-enum-given">
				<Tile digit={1} />
				<Tile digit={2} />
				<Tile digit={3} />
			</span>
			<span class="s-subject-rule">↓ {page.text(`${at}.rule`)}</span>
		</>
	),
	hero: ({ page, at }) => (
		<>
			<span class="s-enum-tree">
				<Branch first={1} />
				<Branch first={2} />
				<Branch first={3} />
			</span>
			<Say>{page.text(`${at}.count`)}</Say>
		</>
	),
	basis: [
		({ page, at }) => (
			<Line gap={16} align="start">
				{(["a", "b", "c"] as const).map((first) => (
					<Stack key={first} gap={6}>
						<Say>{page.text(`${at}.on`, { child: page.plain(`art.${first}`) })}</Say>
						{(["a", "b", "c"] as const)
							.filter((second) => second !== first)
							.map((second) =>
								first === "b" && second === "c" ? (
									<span key={second} class="s-enum-lost">
										{page.plain("art.b")}
										{page.plain("art.c")}?
									</span>
								) : (
									<Children key={second} page={page} who={[first, second]} />
								),
							)}
					</Stack>
				))}
			</Line>
		),
		({ page, at }) => (
			<Stack gap={12} align="start">
				<Line gap={10}>
					<Handshake page={page} first="lea" />
					<Say>{page.text(`${at}.handshake`)}</Say>
				</Line>
				<Line gap={10}>
					<Queue page={page} first="lea" />
					<Queue page={page} first="tom" />
					<Say>{page.text(`${at}.queues`)}</Say>
				</Line>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Sets />
				<Say>{page.text(`${at}.sets`)}</Say>
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			rows: [
				({ page }) => <Verdict page={page} row={1} yes />,
				({ page }) => <Verdict page={page} row={2} yes={false} />,
				({ page }) => <Verdict page={page} row={3} yes />,
			],
		},
		{
			column: 2,
			rows: [
				({ page }) => (
					<Line gap={8}>
						{(
							[
								["a", "b", "c"],
								["a", "c", "b"],
								["b", "a", "c"],
								["b", "c", "a"],
								["c", "a", "b"],
								["c", "b", "a"],
							] as const
						).map((line) => (
							<Children key={line.join("")} page={page} who={line} size="sm" />
						))}
						<Count>6</Count>
					</Line>
				),
				({ page }) => (
					<Line gap={8}>
						{(
							[
								["a", "b"],
								["a", "c"],
								["b", "c"],
							] as const
						).map((pair) => (
							<Children
								key={pair.join("")}
								page={page}
								who={pair}
								size="sm"
								ringed
							/>
						))}
						<Count>3</Count>
					</Line>
				),
				({ page }) => (
					<Line gap={8}>
						{(
							[
								["a", "b"],
								["a", "c"],
								["b", "a"],
								["b", "c"],
								["c", "a"],
								["c", "b"],
							] as const
						).map((pair) => (
							<Children key={pair.join("")} page={page} who={pair} leader />
						))}
						<Count>6</Count>
					</Line>
				),
			],
		},
	],
	legend: ({ page, at }) => (
		<Key sample={<span class="s-enum-frame" />}>{page.text(`${at}.frame`)}</Key>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={16} align="end">
					<Flag />
					<Line gap={6}>
						{paints.map((paint) => (
							<Swatch key={paint} paint={paint} />
						))}
					</Line>
					<Say>{page.text(`${at}.paints`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Stack gap={6}>
						<Line gap={4}>
							<Flag top="red" bottom="yellow" framed />
							<Flag top="red" bottom="green" framed />
						</Line>
						<Say>{page.text(`${at}.red-top`)}</Say>
					</Stack>
				),
				({ page, at }) => (
					<Line gap={12} align="start">
						<Tops top="red" name={page.text(`${at}.red`)} />
						<Tops top="yellow" name={page.text(`${at}.yellow`)} framed />
						<Tops top="green" name={page.text(`${at}.green`)} framed />
					</Line>
				),
				({ page, at }) => (
					<>
						<FlagTable
							corner={
								<>
									<span>{page.text(`${at}.top`)}</span>
									<span>{page.text(`${at}.bottom`)}</span>
								</>
							}
						/>
						<Say>{page.text(`${at}.same`)}</Say>
					</>
				),
			],
		},
		{
			task: ({ page, at }) => (
				<Line gap={16} align="end">
					<Line gap={6}>
						<Box>14</Box>
						<Box>30</Box>
					</Line>
					<Say>{page.text(`${at}.given`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{
								kind: "table",
								header: ["1", "2", "3", "4", "…"],
								rows: [
									["4", "5", "0", "1", ""],
									["", "", "6", "7", ""],
								],
							}}
							tones={{ "cell 0.2": ["warm"] }}
						/>
						<Say>{page.text(`${at}.under`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{
								kind: "table",
								header: ["1", "2", "3", "4", "5", "6", "7", "8", "9"],
								rows: [
									["4", "5", "0", "1", "2", "3", "4", "5", "6"],
									["", "", "6", "7", "8", "9", "", "", ""],
								],
							}}
							tones={lit(
								[2, 3, 4, 5].flatMap((column) => [
									`cell 0.${column}`,
									`cell 1.${column}`,
								]),
								"cool",
							)}
						/>
						<Say>{page.text(`${at}.two`)}</Say>
					</>
				),
				() => (
					<>
						<Line gap={6}>
							{threeApart.map((number) => (
								<Box
									key={number}
									size="sm"
									ring={number === "30" ? "picked" : undefined}
								>
									{number}
								</Box>
							))}
						</Line>
						<Say>4 × 2 + 5 × 1 = 13</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Line gap={14}>
					<Stack gap={6}>
						<Box ring="right">30</Box>
						<Say way="right">{page.text(`${at}.can`)}</Say>
					</Stack>
					<Stack gap={6}>
						<Box ring="wrong">03</Box>
						<Say way="wrong">{page.text(`${at}.not`)}</Say>
					</Stack>
				</Line>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={10}>
					<Tiles digits={[1, 2, 1, 2]} size="md" />
					<Mark way="right" />
					<Tiles digits={[1, 2, 2, 1]} size="md" />
					<Mark way="wrong" />
					<Say>{page.text(`${at}.twos`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Stack gap={8} align="start">
						<Row name={page.text(`${at}.one`)} gap={6}>
							<Tiles digits={[1]} />
							<Tiles digits={[2]} />
						</Row>
						<Row name={page.text(`${at}.two`)} gap={6}>
							<Tiles digits={[1, 1]} />
							<Tiles digits={[1, 2]} />
							<Tiles digits={[2, 1]} />
							<Tiles digits={[2, 2]} banned />
						</Row>
					</Stack>
				),
				({ page, at }) => (
					<Line gap={8}>
						<Tally value="2" under={page.text(`${at}.one`)} />
						<Then />
						<Tally value="3" under={page.text(`${at}.two`)} />
						<Then />
						<Tally value="5" over="3 + 2" under={page.text(`${at}.three`)} />
						<Then />
						<Tally value="8" over="5 + 3" under={page.text(`${at}.four`)} />
					</Line>
				),
				() => (
					<Stack gap={6}>
						<Line gap={8}>
							{noTwoTwos.slice(0, 4).map((digits) => (
								<Tiles key={digits.join("")} digits={digits} />
							))}
						</Line>
						<Line gap={8}>
							{noTwoTwos.slice(4).map((digits) => (
								<Tiles key={digits.join("")} digits={digits} />
							))}
						</Line>
					</Stack>
				),
			],
			note: () => (
				<Line gap={6}>
					{["2", "3", "5", "8"].map((count) => (
						<Box key={count}>{count}</Box>
					))}
					{["13", "21", "34"].map((count) => (
						<span key={count} class="s-enum-next">
							{count}
						</span>
					))}
				</Line>
			),
		},
	],
	traps: {
		missed_case: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">14</Box>
							<Box size="sm">25</Box>
							<Box size="sm" tone="unknown">
								?
							</Box>
							<Box size="sm">36</Box>
							<Op>…</Op>
						</>
					}
					right={
						<>
							<Box size="sm">14</Box>
							<Box size="sm">25</Box>
							<Box size="sm" ring="right">
								30
							</Box>
							<Box size="sm">36</Box>
							<Op>…</Op>
						</>
					}
				/>
			</>
		),
		wrong_operation: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">3</Box>
							<Op>+</Op>
							<Box size="sm">2</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								5
							</Box>
						</>
					}
					right={
						<>
							<Box size="sm">3</Box>
							<Op>×</Op>
							<Box size="sm">2</Box>
							<Op>=</Op>
							<Box size="sm" ring="right">
								6
							</Box>
						</>
					}
				/>
			</>
		),
		double_count: ({ page }) => (
			<Versus
				child={
					<>
						<Handshake page={page} first="lea" />
						<Handshake page={page} first="tom" />
						<Count>2</Count>
					</>
				}
				right={
					<>
						<Handshake page={page} first="lea" />
						<Count>1</Count>
					</>
				}
			/>
		),
	},
};
