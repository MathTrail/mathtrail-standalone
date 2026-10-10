import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import { Chip, Key, Line, Row, Say, Stack, Then, Versus } from "../Sketch";
import type { TopicArt } from "./art";

// Paint is the colour a person or a boat is drawn in.
type Paint = "red" | "blue" | "green" | "violet" | "orange" | "teal";

// Ring is how a person on the line is ringed: the one a step adds, one who
// stands wrongly, or one who stands rightly.
type Ring = "picked" | "wrong" | "right";

// Bead is a person as a circle of their colour with their initial, ringed as
// a step finds them; small where they wait or stand in a block.
function Bead({
	paint,
	ring,
	size = "md",
	children,
}: {
	paint: Paint;
	ring?: Ring;
	size?: "md" | "sm";
	children: ComponentChildren;
}) {
	return (
		<span
			class="s-ord-bead"
			data-paint={paint}
			data-ring={ring}
			data-size={size}
		>
			{children}
		</span>
	);
}

// Slot is a place on the line still free.
function Slot() {
	return <span class="s-ord-slot">?</span>;
}

// Block is two people who always stand side by side, in one outline, ringed
// in the accent where a step finds it.
function Block({
	picked = false,
	children,
}: {
	picked?: boolean;
	children: ComponentChildren;
}) {
	return (
		<span class="s-ord-block" data-picked={picked ? "" : undefined}>
			{children}
		</span>
	);
}

// Arrow is the arrow from one to the next on a line, with its word over it.
function Arrow({ children }: { children: ComponentChildren }) {
	return (
		<span class="s-ord-arrow">
			<span class="s-ord-arrow-word">{children}</span>
		</span>
	);
}

// Placed is someone on a line under the numbers of their places, a block
// under two, the winner's number in gold.
function Placed({
	places,
	gold = false,
	children,
}: {
	places: readonly number[];
	gold?: boolean;
	children: ComponentChildren;
}) {
	return (
		<Stack gap={4}>
			<span class="s-ord-places">
				{places.map((place) => (
					<span
						key={place}
						class="s-ord-place"
						data-gold={gold ? "" : undefined}
					>
						{place}
					</span>
				))}
			</span>
			{children}
		</Stack>
	);
}

// Waiting are those still to be placed, apart from the line under what they
// do.
function Waiting({
	word,
	children,
}: {
	word: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<span class="s-ord-waiting">
			<span class="s-ord-place">{word}</span>
			<Line gap={4}>{children}</Line>
		</span>
	);
}

// Stander is one person of a line drawn standing: who they are, their colour,
// how tall they are, their name, and whether they stood on the line before
// the step, pale, or are the one it adds, their name in bold.
type Stander = {
	who: (typeof tall)[number]["who"];
	paint: Paint;
	height: number;
	name: ComponentChildren;
	before?: boolean;
};

// Lineup is people standing in a row from the tallest, on one floor, their
// names under them with a sign of who is taller between each two.
function Lineup({
	people,
	size = "md",
}: {
	people: readonly Stander[];
	size?: "md" | "sm";
}) {
	return (
		<span class="s-ord-lineup" data-size={size}>
			{people.map((one, at) => [
				at > 0 && (
					<span key={`than-${one.who}`} class="s-ord-than-column">
						<span class="s-ord-floor" />
						<span class="s-ord-than">&gt;</span>
					</span>
				),
				<span
					key={one.who}
					class="s-ord-stander"
					data-before={one.before ? "" : undefined}
				>
					<span class="s-ord-floor">
						<span
							class="s-ord-figure"
							data-paint={one.paint}
							style={{ "--s-height": `${one.height}` }}
						/>
					</span>
					<span class="s-ord-name">{one.name}</span>
				</span>,
			])}
		</span>
	);
}

// Boat is a sailing boat in its colour: framed where a step places it, and
// pale where it waits.
function Boat({
	paint,
	framed = false,
	waiting = false,
}: {
	paint: Paint;
	framed?: boolean;
	waiting?: boolean;
}) {
	return (
		<span
			class="s-ord-boat"
			data-paint={paint}
			data-framed={framed ? "" : undefined}
			data-waiting={waiting ? "" : undefined}
		>
			<span class="s-ord-sail" />
			<span class="s-ord-hull" />
		</span>
	);
}

// Clue is a clue as it is told, on a pill.
function Clue({ children }: { children: ComponentChildren }) {
	return <span class="s-ord-clue">{children}</span>;
}

// tall are the four children of the page's lines, by their place from the
// tallest, with their colours and heights.
const tall = [
	{ who: "ann", paint: "red", height: 100 },
	{ who: "ben", paint: "blue", height: 82 },
	{ who: "kate", paint: "green", height: 66 },
	{ who: "dan", paint: "violet", height: 52 },
] as const;

// lineOf is the first so many children of the line, each named in the
// page's words, those on the line before the last so many pale.
function lineOf(page: PageReader, count: number, added = count): Stander[] {
	return tall.slice(0, count).map((one, at) => ({
		who: one.who,
		paint: one.paint,
		height: one.height,
		name: page.text(`art.${one.who}-name`),
		before: at < count - added,
	}));
}

// Child is one of the four children as a bead with their initial.
function Child({
	page,
	who,
	ring,
}: {
	page: PageReader;
	who: (typeof tall)[number]["who"];
	ring?: Ring;
}) {
	const paint = tall.find((one) => one.who === who)?.paint ?? "red";
	return (
		<Bead paint={paint} ring={ring}>
			{page.text(`art.${who}`)}
		</Bead>
	);
}

// queue are the children of the canteen's queue with their colours.
const queue = {
	lena: "orange",
	max: "blue",
	nina: "green",
	oleg: "violet",
	pavel: "red",
} as const satisfies Record<string, Paint>;

// runners are the six runners with their colours.
const runners = {
	ada: "red",
	bob: "blue",
	cid: "teal",
	dot: "green",
	eli: "violet",
	fay: "orange",
} as const satisfies Record<string, Paint>;

// Person is one of an example's people as a bead with their initial, the
// words of their initials under where.
function Person({
	page,
	where,
	who,
	paint,
	ring,
	size,
}: {
	page: PageReader;
	where: string;
	who: string;
	paint: Paint;
	ring?: Ring;
	size?: "md" | "sm";
}) {
	return (
		<Bead paint={paint} ring={ring} size={size}>
			{page.text(`${where}.${who}`)}
		</Bead>
	);
}

// Named are people each over their name.
function Named({
	page,
	where,
	people,
}: {
	page: PageReader;
	where: string;
	people: Readonly<Record<string, Paint>>;
}) {
	return (
		<Line gap={12} align="start">
			{Object.entries(people).map(([who, paint]) => (
				<Stack key={who} gap={4}>
					<Person page={page} where={where} who={who} paint={paint} />
					<Say>{page.text(`${where}.${who}-name`)}</Say>
				</Stack>
			))}
		</Line>
	);
}

// Queue is the canteen's queue, its places numbered, in the order given, a
// free place as "?" and a ring where given.
function Queue({
	page,
	order,
	rings = {},
}: {
	page: PageReader;
	order: readonly (keyof typeof queue | "?")[];
	rings?: Readonly<Partial<Record<keyof typeof queue, Ring>>>;
}) {
	const where = "examples.2.art";
	return (
		<Line gap={6} align="end">
			{order.map((who, at) => {
				const place = at + 1;
				return (
					<Placed key={place} places={[place]}>
						{who === "?" ? (
							<Slot />
						) : (
							<Person
								page={page}
								where={where}
								who={who}
								paint={queue[who]}
								ring={rings[who]}
							/>
						)}
					</Placed>
				);
			})}
		</Line>
	);
}

// Runner is one of the six runners as a bead.
function Runner({
	page,
	who,
	size,
}: {
	page: PageReader;
	who: keyof typeof runners;
	size?: "md" | "sm";
}) {
	return (
		<Person
			page={page}
			where="examples.3.art"
			who={who}
			paint={runners[who]}
			size={size}
		/>
	);
}

// Pair is a block of two runners.
function Pair({
	page,
	first,
	second,
	picked,
	size,
}: {
	page: PageReader;
	first: keyof typeof runners;
	second: keyof typeof runners;
	picked?: boolean;
	size?: "md" | "sm";
}) {
	return (
		<Block picked={picked}>
			<Runner page={page} who={first} size={size} />
			<Runner page={page} who={second} size={size} />
		</Block>
	);
}

// Turned is the words of a clue turned into "taller than", read from its
// cell of the key idea's table, with a note where it had to be turned round.
function Turned({
	page,
	row,
	turned,
}: {
	page: PageReader;
	row: number;
	turned: boolean;
}) {
	return (
		<Line gap={10}>
			<span class="s-ord-turned">{page.text(`idea.table.rows.${row}.2`)}</span>
			{turned && <Chip tone="cool">↻ {page.text("idea.art.turned")}</Chip>}
		</Line>
	);
}

/**
 * art are the drawings of ordering: people standing from the tallest to the
 * shortest, clues as arrows from one person to the next, boats and queues
 * with their places, blocks of two who stand side by side, and the places
 * still free.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => (
		<>
			{page.text(`${at}.line`)}
			<span class="s-ord-clues">
				<Clue>{page.text(`${at}.clue-1`)}</Clue>
				<Clue>{page.text(`${at}.clue-2`)}</Clue>
				<Clue>{page.text(`${at}.clue-3`)}</Clue>
			</span>
			<span class="s-subject-rule">↓ {page.text(`${at}.rule`)}</span>
		</>
	),
	hero: ({ page, at }) => (
		<>
			<span class="s-ord-axis">
				<Say>{page.text(`${at}.top`)}</Say>
				<span class="s-ord-axis-line" />
				<Say>{page.text(`${at}.bottom`)}</Say>
			</span>
			<Lineup people={lineOf(page, 4)} />
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={10}>
				<Line gap={8}>
					<Child page={page} who="ann" />
					<Arrow>{page.text("art.taller")}</Arrow>
					<Child page={page} who="ben" />
				</Line>
				<Say>{page.text(`${at}.clue`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Line gap={16}>
				<Lineup people={lineOf(page, 2)} size="sm" />
				<Stack gap={2}>
					<Say>{page.text(`${at}.clue`)}</Say>
					<Say>=</Say>
					<Say>{page.text(`${at}.turned`)}</Say>
				</Stack>
			</Line>
		),
		({ page, at }) => (
			<span class="s-ord-chain">
				<span class="s-ord-leap">{page.text(`${at}.implied`)}</span>
				<Line gap={6}>
					<Child page={page} who="ann" />
					<Then />
					<Child page={page} who="ben" />
					<Then />
					<Child page={page} who="kate" />
				</Line>
			</span>
		),
	],
	idea: [
		{
			column: 1,
			rows: [
				({ page }) => <Turned page={page} row={1} turned={false} />,
				({ page }) => <Turned page={page} row={2} turned />,
				({ page }) => <Turned page={page} row={3} turned />,
			],
		},
		{
			column: 2,
			rows: [2, 3, 4].map((count) => ({ page }: { page: PageReader }) => (
				<Lineup people={lineOf(page, count, count === 2 ? 2 : 1)} size="sm" />
			)),
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key
				sample={
					<Bead paint="red" ring="picked" size="sm">
						{page.text("art.ann")}
					</Bead>
				}
			>
				{page.text(`${at}.added`)}
			</Key>
			<Key sample={<Slot />}>{page.text(`${at}.free`)}</Key>
			<Key
				sample={
					<Block>
						<Bead paint="red" size="sm">
							{page.text("art.ann")}
						</Bead>
						<Bead paint="blue" size="sm">
							{page.text("art.ben")}
						</Bead>
					</Block>
				}
			>
				{page.text(`${at}.block`)}
			</Key>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={16} align="start">
					{(["red", "blue", "green"] as const).map((paint) => (
						<Stack key={paint} gap={4}>
							<Slot />
							<Boat paint={paint} />
							<Say>{page.text(`${at}.${paint}`)}</Say>
						</Stack>
					))}
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Line gap={8}>
						<Boat paint="blue" framed />
						<Arrow>{page.text(`${at}.before`)}</Arrow>
						<Boat paint="red" framed />
						<span class="s-ord-apart" />
						<Waiting word={page.text(`${at}.waits`)}>
							<Boat paint="green" waiting />
						</Waiting>
					</Line>
				),
				({ page, at }) => (
					<Line gap={8}>
						<Boat paint="green" framed />
						<Arrow>{page.text(`${at}.before`)}</Arrow>
						<Boat paint="blue" />
						<Then />
						<Boat paint="red" />
					</Line>
				),
				() => (
					<Line gap={8} align="end">
						<Placed places={[1]} gold>
							<Boat paint="green" />
						</Placed>
						<span class="s-ord-between">
							<Then />
						</span>
						<Placed places={[2]}>
							<Boat paint="blue" />
						</Placed>
						<span class="s-ord-between">
							<Then />
						</span>
						<Placed places={[3]}>
							<Boat paint="red" />
						</Placed>
					</Line>
				),
			],
		},
		{
			task: ({ page, at }) => <Named page={page} where={at} people={queue} />,
			steps: [
				({ page, at }) => (
					<Line gap={12} align="end">
						<Queue
							page={page}
							order={["?", "?", "nina", "oleg", "?"]}
							rings={{ nina: "picked", oleg: "picked" }}
						/>
						<span class="s-ord-apart" />
						<Waiting word={page.text(`${at}.wait`)}>
							{(["max", "lena", "pavel"] as const).map((who) => (
								<Person
									key={who}
									page={page}
									where={at}
									who={who}
									paint={queue[who]}
									size="sm"
								/>
							))}
						</Waiting>
					</Line>
				),
				({ page, at }) => (
					<Line gap={12} align="end">
						<Queue
							page={page}
							order={["max", "lena", "nina", "oleg", "?"]}
							rings={{ max: "picked", lena: "picked" }}
						/>
						<span class="s-ord-apart" />
						<Waiting word={page.text(`${at}.waits`)}>
							<Person
								page={page}
								where={at}
								who="pavel"
								paint="red"
								size="sm"
							/>
						</Waiting>
					</Line>
				),
				({ page }) => (
					<Queue
						page={page}
						order={["max", "lena", "nina", "oleg", "pavel"]}
						rings={{ pavel: "picked" }}
					/>
				),
			],
			note: ({ page, at }) => (
				<>
					<Queue
						page={page}
						order={["max", "pavel", "nina", "oleg", "lena"]}
						rings={{ lena: "wrong" }}
					/>
					<Say>{page.text(`${at}.other`)}</Say>
				</>
			),
		},
		{
			task: ({ page, at }) => <Named page={page} where={at} people={runners} />,
			steps: [
				({ page, at }) => (
					<Stack gap={10} align="start">
						<Row name={page.text(`${at}.blocks`)} gap={8}>
							<Pair page={page} first="ada" second="bob" picked />
							<Pair page={page} first="fay" second="eli" picked />
						</Row>
						<Row name={page.text(`${at}.order`)} gap={6}>
							<Runner page={page} who="dot" />
							<Then />
							<Runner page={page} who="cid" />
							<Then />
							<Pair page={page} first="ada" second="bob" />
						</Row>
					</Stack>
				),
				({ page, at }) => (
					<Stack gap={10}>
						<Line gap={6}>
							<Slot />
							<Then />
							<Runner page={page} who="dot" />
							<Then />
							<Runner page={page} who="cid" />
							<Then />
							<Pair page={page} first="ada" second="bob" />
						</Line>
						<Line gap={8}>
							<Say>{page.text(`${at}.only`)}</Say>
							<Pair page={page} first="fay" second="eli" picked size="sm" />
						</Line>
					</Stack>
				),
				({ page }) => (
					<Line gap={8} align="end">
						<Placed places={[1, 2]}>
							<Pair page={page} first="fay" second="eli" />
						</Placed>
						<Placed places={[3]}>
							<Runner page={page} who="dot" />
						</Placed>
						<Placed places={[4]}>
							<Runner page={page} who="cid" />
						</Placed>
						<Placed places={[5, 6]}>
							<Pair page={page} first="ada" second="bob" />
						</Placed>
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Row name={page.text(`${at}.six`)} gap={4}>
						{(Object.keys(runners) as (keyof typeof runners)[]).map((who) => (
							<Runner key={who} page={page} who={who} size="sm" />
						))}
					</Row>
					<Row name={page.text(`${at}.four`)} gap={4}>
						<Pair page={page} first="fay" second="eli" size="sm" />
						<Runner page={page} who="dot" size="sm" />
						<Runner page={page} who="cid" size="sm" />
						<Pair page={page} first="ada" second="bob" size="sm" />
					</Row>
				</Stack>
			),
		},
	],
	traps: {
		ignored_condition: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<Queue
							page={page}
							order={["max", "pavel", "nina", "oleg", "lena"]}
							rings={{ lena: "wrong" }}
						/>
					}
					right={
						<Queue
							page={page}
							order={["max", "lena", "nina", "oleg", "pavel"]}
							rings={{ pavel: "right" }}
						/>
					}
				/>
			</>
		),
		reversed_relation: ({ page, at }) => (
			<>
				<Line gap={6} align="start">
					<Clue>{page.text(`${at}.clue`)}</Clue>
				</Line>
				<Versus
					child={
						<>
							<Child page={page} who="kate" />
							<Arrow>{page.text("art.taller")}</Arrow>
							<Child page={page} who="ben" />
							<span class="s-ord-mark" data-way="wrong">
								✕
							</span>
						</>
					}
					right={
						<>
							<Child page={page} who="ben" />
							<Arrow>{page.text("art.taller")}</Arrow>
							<Child page={page} who="kate" />
							<span class="s-ord-mark" data-way="right">
								✓
							</span>
						</>
					}
				/>
			</>
		),
		stopped_early: ({ page, at }) => (
			<>
				<Line gap={6}>
					<Clue>{page.text(`${at}.clue-1`)}</Clue>
					<Clue>{page.text(`${at}.clue-2`)}</Clue>
					<Say>{page.text(`${at}.who`)}</Say>
				</Line>
				<Versus
					child={
						<>
							<Child page={page} who="ann" />
							<Arrow>{page.text("art.taller")}</Arrow>
							<Child page={page} who="ben" ring="wrong" />
							<span class="s-ord-stop" />
						</>
					}
					right={
						<>
							<Child page={page} who="ann" />
							<Arrow>{page.text("art.taller")}</Arrow>
							<Child page={page} who="ben" />
							<Arrow>{page.text("art.taller")}</Arrow>
							<Child page={page} who="kate" ring="right" />
						</>
					}
				/>
			</>
		),
	},
};
