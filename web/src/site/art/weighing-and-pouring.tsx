import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import {
	Box,
	Chip,
	Key,
	Line,
	Op,
	Say,
	Stack,
	Stretch,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";
import { Kind } from "./pictures";

// Jug is a jug of so many litres, as tall as it holds, with the water in it
// and how much there is written on the water, and its size under it; ringed
// in dashed red where it is filled to a mark it does not have.
function Jug({
	page,
	holds,
	capacity,
	wrong = false,
}: {
	page: PageReader;
	holds: number;
	capacity: number;
	wrong?: boolean;
}) {
	return (
		<Stack gap={3}>
			<span
				class="s-wei-jug"
				data-wrong={wrong ? "" : undefined}
				style={{
					"--s-capacity": String(capacity),
					"--s-holds": String(holds),
				}}
			>
				<span class="s-wei-water" />
				<span class="s-wei-litres">{holds}</span>
			</span>
			<span class="s-wei-size">
				{page.text("art.litres", { count: capacity })}
			</span>
		</Stack>
	);
}

// Jugs are what two jugs hold: the first's litres of its size, then the
// second's.
type Jugs = readonly [readonly [number, number], readonly [number, number]];

// State is two jugs side by side, what each holds after a step, with the
// step's number over them where it is counted, framed where it is the goal.
function State({
	page,
	jugs,
	step,
	goal = false,
}: {
	page: PageReader;
	jugs: Jugs;
	step?: number;
	goal?: boolean;
}) {
	return (
		<Stack gap={4}>
			{step !== undefined && (
				<span class="s-wei-step" data-goal={goal ? "" : undefined}>
					{step}
				</span>
			)}
			<span class="s-wei-state" data-goal={goal ? "" : undefined}>
				{jugs.map(([holds, capacity], at) => (
					<Jug key={at} page={page} holds={holds} capacity={capacity} />
				))}
			</span>
		</Stack>
	);
}

// Move is what is done from one state to the next, over its arrow.
function Move({ children }: { children: ComponentChildren }) {
	return (
		<span class="s-wei-move">
			<span>{children}</span>
			<span class="s-wei-arrow">→</span>
		</span>
	);
}

// Coin is a gold coin, small or tiny where many stand together.
function Coin({ size = "md" }: { size?: "md" | "sm" | "xs" }) {
	return <span class="s-wei-coin" data-size={size} />;
}

// Coins are so many coins in a row.
function Coins({ count, size }: { count: number; size?: "md" | "sm" | "xs" }) {
	return (
		<span class="s-wei-coins">
			{Array.from({ length: count }, (_, at) => (
				<Coin key={at} size={size} />
			))}
		</span>
	);
}

// Pan is one pan of a balance and what lies on it: so many coins, or what is
// written on it.
function Pan({ load }: { load: number | string }) {
	return (
		<span class="s-wei-pan">
			{typeof load === "number" ? (
				<Coins count={load} size="sm" />
			) : (
				<span class="s-wei-pan-load">{load}</span>
			)}
		</span>
	);
}

// Balance is a pan balance with what lies on each pan, the heavier pan
// lower where one is, level where they weigh the same, and what it shows
// under it.
function Balance({
	left,
	right,
	lower,
	children,
}: {
	left: number | string;
	right: number | string;
	lower?: "left" | "right";
	children?: ComponentChildren;
}) {
	return (
		<span class="s-wei-balance" data-lower={lower}>
			<span class="s-wei-beam">
				<Pan load={left} />
				<span class="s-wei-post" />
				<Pan load={right} />
			</span>
			<span class="s-wei-foot" />
			{children !== undefined && <span class="s-wei-shows">{children}</span>}
		</span>
	);
}

// Aside are coins set aside, in a dashed box, with what they are under it.
function Aside({
	load,
	children,
}: {
	load: number | string;
	children: ComponentChildren;
}) {
	return (
		<Stack gap={4}>
			<span class="s-wei-aside">
				{typeof load === "number" ? <Coins count={load} size="sm" /> : load}
			</span>
			<Say>{children}</Say>
		</Stack>
	);
}

// Heaps are coins in heaps of three, so many heaps, each heap's coins in a
// row or, where there are nine, in a square.
function Heaps({
	heaps,
	coins,
	size,
}: {
	heaps: number;
	coins: number;
	size: "md" | "sm" | "xs";
}) {
	return (
		<Line gap={6}>
			{Array.from({ length: heaps }, (_, heap) => (
				<span
					key={heap}
					class="s-wei-heap"
					data-square={coins === 9 ? "" : undefined}
				>
					{Array.from({ length: coins }, (_, at) => (
						<Coin key={at} size={size} />
					))}
				</span>
			))}
		</Line>
	);
}

// CooksTable is the table of what the cook's two jugs hold, step by step,
// the litre found in the last row.
function CooksTable({ page }: { page: PageReader }) {
	return (
		<Kind
			locale={page.locale}
			picture={{
				kind: "table",
				header: [
					page.plain("art.step"),
					page.plain("art.litres", { count: 3 }),
					page.plain("art.litres", { count: 5 }),
				],
				rows: [
					["0", "0", "0"],
					["1", "3", "0"],
					["2", "0", "3"],
					["3", "3", "3"],
					["4", "1", "5"],
				],
			}}
			tones={{ "cell 4.1": ["cool", "picked"] }}
		/>
	);
}

/**
 * art are the drawings of weighing and pouring: jugs as tall as they hold
 * and the water in them, step after step; a pan balance whose heavier pan
 * hangs lower, with the coins on its pans and those set aside; heaps of
 * three; and a table of states, as the card draws one.
 */
export const art: TopicArt = {
	pictures: true,
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Stack gap={14} align="start">
				<Line gap={6} align="end">
					<State
						page={page}
						step={0}
						jugs={[
							[5, 5],
							[0, 3],
						]}
					/>
					<Move>{page.text("art.pour")}</Move>
					<State
						page={page}
						step={1}
						jugs={[
							[2, 5],
							[3, 3],
						]}
					/>
					<Move>{page.text("art.empty")}</Move>
					<State
						page={page}
						step={2}
						jugs={[
							[2, 5],
							[0, 3],
						]}
					/>
				</Line>
				<Line gap={6} align="end">
					<State
						page={page}
						step={3}
						jugs={[
							[0, 5],
							[2, 3],
						]}
					/>
					<Move>{page.text("art.fill")}</Move>
					<State
						page={page}
						step={4}
						jugs={[
							[5, 5],
							[2, 3],
						]}
					/>
					<Move>{page.text("art.top")}</Move>
					<State
						page={page}
						step={5}
						jugs={[
							[4, 5],
							[3, 3],
						]}
						goal
					/>
				</Line>
			</Stack>
			<Say>{page.text(`${at}.pair`)}</Say>
		</>
	),
	basis: [
		({ page }) => (
			<Line gap={8} align="end">
				<State
					page={page}
					jugs={[
						[5, 5],
						[0, 3],
					]}
				/>
				<Move>{page.text("art.pour")}</Move>
				<State
					page={page}
					jugs={[
						[2, 5],
						[3, 3],
					]}
				/>
			</Line>
		),
		({ page, at }) => (
			<Line gap={14} align="end">
				<Balance left={2} right={2} lower="left">
					{page.text(`${at}.left`)}
				</Balance>
				<Balance left={2} right={2}>
					{page.text(`${at}.level`)}
				</Balance>
				<Balance left={2} right={2} lower="right">
					{page.text(`${at}.right`)}
				</Balance>
			</Line>
		),
		({ page, at }) => (
			<Stack gap={8} align="start">
				<Line gap={8}>
					<Jug page={page} holds={3} capacity={5} />
					<Say>{page.text(`${at}.no-marks`)}</Say>
				</Line>
				<Line gap={8}>
					<Chip tone="red">✕ {page.text(`${at}.pour-out`)}</Chip>
					<Say>{page.text(`${at}.unless`)}</Say>
				</Line>
				<Line gap={8}>
					<Chip tone="grey">{page.text(`${at}.weight`)}</Chip>
					<Say>{page.text(`${at}.one-pan`)}</Say>
				</Line>
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			words: "under",
			rows: [
				() => <Heaps heaps={1} coins={3} size="md" />,
				() => <Heaps heaps={3} coins={3} size="sm" />,
				() => <Heaps heaps={3} coins={9} size="xs" />,
			],
		},
	],
	legend: ({ page, at }) => (
		<Key sample={<span class="s-wei-goal-sample" />}>
			{page.text(`${at}.goal`)}
		</Key>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={12} align="end">
					<State
						page={page}
						jugs={[
							[0, 3],
							[0, 5],
						]}
					/>
					<Say>{page.text(`${at}.jugs`)}</Say>
				</Line>
			),
			steps: [
				({ page }) => (
					<Line gap={6} align="end">
						<State
							page={page}
							step={0}
							jugs={[
								[0, 3],
								[0, 5],
							]}
						/>
						<Move>{page.text("art.fill")}</Move>
						<State
							page={page}
							step={1}
							jugs={[
								[3, 3],
								[0, 5],
							]}
						/>
						<Move>{page.text("art.pour")}</Move>
						<State
							page={page}
							step={2}
							jugs={[
								[0, 3],
								[3, 5],
							]}
						/>
					</Line>
				),
				({ page }) => (
					<Line gap={6} align="end">
						<State
							page={page}
							step={2}
							jugs={[
								[0, 3],
								[3, 5],
							]}
						/>
						<Move>{page.text("art.fill")}</Move>
						<State
							page={page}
							step={3}
							jugs={[
								[3, 3],
								[3, 5],
							]}
						/>
						<Move>{page.text("art.pour")}</Move>
						<State
							page={page}
							step={4}
							jugs={[
								[1, 3],
								[5, 5],
							]}
							goal
						/>
					</Line>
				),
				({ page, at }) => (
					<>
						<CooksTable page={page} />
						<Say>{page.text(`${at}.rows`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Line gap={6}>
					<Say>{page.text(`${at}.three`)}</Say>
					{["0", "2", "3", "5"].map((litres) => (
						<Box key={litres} size="sm">
							{litres}
						</Box>
					))}
					<Say way="wrong">{page.text(`${at}.no-litre`)}</Say>
				</Line>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={12}>
					<Coins count={8} />
					<Say>{page.text(`${at}.one-lighter`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Line gap={16} align="end">
						<Balance left={3} right={3}>
							{page.text(`${at}.three`)}
						</Balance>
						<Aside load={2}>{page.text(`${at}.aside`)}</Aside>
					</Line>
				),
				({ page, at }) => (
					<Line gap={10} align="end">
						<Balance left={3} right={3} lower="right">
							{page.text(`${at}.lighter`)}
						</Balance>
						<span class="s-wei-then">→</span>
						<Balance left={1} right={1}>
							{page.text(`${at}.one`)}
						</Balance>
						<Say>{page.text(`${at}.found`)}</Say>
					</Line>
				),
				({ page, at }) => (
					<Line gap={10} align="end">
						<Balance left={3} right={3}>
							{page.text(`${at}.level`)}
						</Balance>
						<span class="s-wei-then">→</span>
						<Balance left={1} right={1} lower="right">
							{page.text(`${at}.of-two`)}
						</Balance>
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Line gap={12} align="end">
					<Balance left={4} right={4}>
						{page.text(`${at}.never-level`)}
					</Balance>
					<Say way="wrong">{page.text(`${at}.four-left`)}</Say>
				</Line>
			),
		},
		{
			steps: [
				() => (
					<Line gap={6}>
						<Box size="sm">3 × 3 × 3 × 3</Box>
						<Op>=</Op>
						<Box size="sm">81</Box>
						<Op>&lt;</Op>
						<Box size="sm" ring="wrong">
							100
						</Box>
					</Line>
				),
				({ page, at }) => (
					<Line gap={16} align="end">
						<Balance left="33" right="33" />
						<Aside load="34">{page.text(`${at}.aside`)}</Aside>
					</Line>
				),
				({ page, at }) => (
					<>
						<Line gap={3}>
							{["100", "34", "12", "4", "2"].map((coins, weighing) => [
								<Box key={coins} size="sm">
									{coins}
								</Box>,
								<Stretch key={`weighing-${coins}`} length={26}>
									{weighing + 1}
								</Stretch>,
							])}
							<Box size="sm" ring="picked">
								1
							</Box>
						</Line>
						<Say>{page.text(`${at}.numbers`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Line gap={8}>
						<Chip tone="green">{page.text(`${at}.plan`)}</Chip>
						<Say>{page.text(`${at}.always`)}</Say>
					</Line>
					<Line gap={8}>
						<Chip tone="cool">{page.text(`${at}.bound`)}</Chip>
						<Say>{page.text(`${at}.at-most`)}</Say>
					</Line>
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
						<>
							<Jug page={page} holds={1} capacity={5} wrong />
							<Say way="wrong">{page.text(`${at}.to-mark`)}</Say>
						</>
					}
					right={
						<>
							<State
								page={page}
								jugs={[
									[1, 3],
									[5, 5],
								]}
							/>
							<Say>{page.text(`${at}.after-pouring`)}</Say>
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
							<span class="s-wei-heap-wrong">
								<Coins count={3} size="sm" />
							</span>
							<Say way="wrong">{page.text(`${at}.here`)}</Say>
						</>
					}
					right={
						<>
							<Balance left={1} right={1} />
							<Say>{page.text(`${at}.one-more`)}</Say>
						</>
					}
				/>
			</>
		),
		off_by_one: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<Box size="sm" ring="wrong">
							{page.text(`${at}.three`)}
						</Box>
					}
					right={
						<>
							<Box size="sm" ring="right">
								{page.text(`${at}.four`)}
							</Box>
							<Say>{page.text(`${at}.empty`)}</Say>
						</>
					}
				/>
			</>
		),
	},
};
