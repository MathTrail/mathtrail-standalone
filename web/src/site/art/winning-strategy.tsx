import type { ComponentChildren } from "preact";
import type { Tone, Tones } from "../../design/picture/tones";
import type { PageReader } from "../reader";
import { Box, Chip, Key, Line, Op, Say, Stack, Then, Versus } from "../Sketch";
import type { TopicArt } from "./art";
import { Kind } from "./pictures";

// Positions are the positions of a game from one count to another on a
// number line the card draws: each that loses for the player about to move
// lit as struck, each that wins green, and the one a step finds ringed.
function Positions({
	page,
	from,
	to,
	loses,
	ringed,
}: {
	page: PageReader;
	from: number;
	to: number;
	loses: (count: number) => boolean;
	ringed?: number;
}) {
	const tones: Tones = Object.fromEntries(
		Array.from({ length: to - from + 1 }, (_, at) => {
			const count = from + at;
			const tone: Tone[] = [loses(count) ? "struck" : "green"];
			return [`tick ${count}`, count === ringed ? [...tone, "picked"] : tone];
		}),
	);
	return (
		<Kind
			locale={page.locale}
			picture={{ kind: "number_line", from, to }}
			tones={tones}
		/>
	);
}

// everyFour and everyFive say whether a count loses for the player about to
// move, where a move takes 1 to 3, or 1 to 4.
const everyFour = (count: number) => count % 4 === 0;
const everyFive = (count: number) => count % 5 === 0;

// Tile is a position as a tile, red where it loses for the player about to
// move and green where it wins, or plain where it is not yet known; ringed
// as a step rings it.
function Tile({
	count,
	loses,
	ring,
	size = "xs",
}: {
	count: number;
	loses?: boolean;
	ring?: "picked" | "wrong" | "right";
	size?: "sm" | "xs";
}) {
	const tone = loses === undefined ? "plain" : loses ? "red" : "green";
	return (
		<Box tone={tone} size={size} ring={ring}>
			{count}
		</Box>
	);
}

// Tiles are the positions from one count to another, as tiles going on to
// the next line where there are many, those of the counts given ringed.
function Tiles({
	from,
	to,
	loses,
	ringed = [],
}: {
	from: number;
	to: number;
	loses: (count: number) => boolean;
	ringed?: readonly number[];
}) {
	return (
		<span class="s-win-tiles">
			{Array.from({ length: to - from + 1 }, (_, at) => {
				const count = from + at;
				return (
					<Tile
						key={count}
						count={count}
						loses={loses(count)}
						ring={ringed.includes(count) ? "picked" : undefined}
					/>
				);
			})}
		</span>
	);
}

// Move is a move from one position to the next, what it takes over its
// arrow, red where the rules forbid it.
function Move({
	wrong = false,
	children,
}: {
	wrong?: boolean;
	children?: ComponentChildren;
}) {
	return (
		<span class="s-win-move" data-wrong={wrong ? "" : undefined}>
			<span>{children}</span>
			<span class="s-win-arrow">→</span>
		</span>
	);
}

// Matches are so many matches standing in a row, their heads up.
function Matches({ count }: { count: number }) {
	return (
		<span class="s-win-matches">
			{Array.from({ length: count }, (_, at) => (
				<span key={at} />
			))}
		</span>
	);
}

// Sweets are so many sweets, as pink dots.
function Sweets({ count }: { count: number }) {
	return (
		<span class="s-win-sweets">
			{Array.from({ length: count }, (_, at) => (
				<span key={at} />
			))}
		</span>
	);
}

// Spot is the sample of a key: a pink dot for a position that loses for the
// player about to move, a green one for one that wins.
function Spot({ loses }: { loses: boolean }) {
	return <span class="s-win-spot" data-loses={loses ? "" : undefined} />;
}

// Keys are the key of the positions' colours.
function Keys({ page }: { page: PageReader }) {
	return (
		<>
			<Key sample={<Spot loses />}>{page.text("art.loses")}</Key>
			<Key sample={<Spot loses={false} />}>{page.text("art.wins")}</Key>
		</>
	);
}

// Board is the king's board of so many cells a side, checkered, the flag in
// its top right corner: the cells one move from the flag green where the
// step shows them, a red dot on each cell that loses for the player about
// to move where it shows those, and the king in the bottom left corner
// where it stands there.
function Board({
	size = 5,
	near = false,
	dots = false,
	king = false,
}: {
	size?: number;
	near?: boolean;
	dots?: boolean;
	king?: boolean;
}) {
	const last = size - 1;
	return (
		<span class="s-win-board" style={{ "--s-cells": String(size) }}>
			{Array.from({ length: size }, (_, row) =>
				Array.from({ length: size }, (_, column) => {
					const up = row;
					const across = last - column;
					const flag = up === 0 && across === 0;
					const green = near && !flag && up <= 1 && across <= 1;
					const loses = dots && !flag && up % 2 === 0 && across % 2 === 0;
					const kingHere = king && row === last && column === 0;
					return (
						<span
							key={`${row}.${column}`}
							class="s-win-cell"
							data-dark={(row + column) % 2 === 1 ? "" : undefined}
							data-green={green ? "" : undefined}
						>
							{flag && <span class="s-win-flag" />}
							{kingHere ? (
								<span class="s-win-king">K</span>
							) : (
								loses && <span class="s-win-dot" />
							)}
						</span>
					);
				}),
			)}
		</span>
	);
}

// Petal is how a petal of the daisy looks: still there, torn off, or in one
// of the two rows the second player makes.
type Petal = "plain" | "torn" | "blue" | "orange";

// Daisy is a daisy of twelve petals round its middle, the first at the top
// and the rest clockwise, each drawn as it is.
function Daisy({
	petals,
	size = 110,
}: {
	petals: readonly Petal[];
	size?: number;
}) {
	return (
		<svg
			aria-hidden="true"
			class="s-win-daisy"
			width={size}
			height={size}
			viewBox="0 0 120 120"
		>
			{petals.map((petal, at) => (
				<ellipse
					key={at}
					class="s-win-petal"
					data-petal={petal}
					cx={60}
					cy={27}
					rx={7}
					ry={16}
					transform={`rotate(${(360 * at) / petals.length} 60 60)`}
				/>
			))}
			<circle class="s-win-middle" cx={60} cy={60} r={15} />
		</svg>
	);
}

// daisy is the daisy's twelve petals, all plain but those given.
function daisy(changed: Readonly<Record<number, Petal>> = {}): Petal[] {
	return Array.from({ length: 12 }, (_, at) => changed[at] ?? "plain");
}

// rows is the daisy after the second player's first move: the petals torn
// at its top and its bottom, a blue row of five on the right and an orange
// one on the left.
function rows(changed: Readonly<Record<number, Petal>> = {}): Petal[] {
	return Array.from(
		{ length: 12 },
		(_, at) =>
			changed[at] ??
			(at === 0 || at === 6 ? "torn" : at < 6 ? "blue" : "orange"),
	);
}

/**
 * art are the drawings of games with a winning strategy: positions on a
 * number line, red where they lose for the player about to move and green
 * where they win; tiles of positions and the moves between them; matches and
 * sweets; the king's board with the flag, its losing cells dotted red; and
 * the daisy, torn into two rows that mirror each other.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page }) => (
		<>
			<Matches count={13} />
			<span class="s-win-line">
				<Positions page={page} from={0} to={8} loses={everyFour} />
			</span>
			<Line gap={16}>
				<Keys page={page} />
			</Line>
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={8}>
				<Positions page={page} from={0} to={4} loses={everyFour} />
				<Say>{page.text(`${at}.back`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Line gap={8}>
				<Tile count={6} loses={false} size="sm" />
				<Move>{page.text(`${at}.take`)}</Move>
				<Tile count={4} loses size="sm" ring="picked" />
				<Say>{page.text(`${at}.losing`)}</Say>
			</Line>
		),
		({ page, at }) => (
			<Stack gap={6}>
				{[
					[1, 3],
					[2, 2],
					[3, 1],
				].map(([theirs, ours]) => (
					<Line key={theirs} gap={6}>
						<Say>{page.text(`${at}.they`)}</Say>
						<Box size="xs">{theirs}</Box>
						<Op>+</Op>
						<Say>{page.text(`${at}.answer`)}</Say>
						<Box size="xs" tone="cool">
							{ours}
						</Box>
						<Op>= 4</Op>
					</Line>
				))}
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			words: "under",
			rows: [
				() => <Tiles from={0} to={3} loses={everyFour} ringed={[1, 2, 3]} />,
				() => <Tiles from={0} to={4} loses={everyFour} ringed={[4]} />,
				() => <Tiles from={4} to={7} loses={everyFour} ringed={[5, 6, 7]} />,
				() => <Tiles from={4} to={8} loses={everyFour} ringed={[8]} />,
			],
		},
	],
	legend: ({ page }) => <Keys page={page} />,
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={14}>
					<Sweets count={23} />
					<Say>{page.text(`${at}.sweets`)}</Say>
				</Line>
			),
			steps: [
				({ page }) => (
					<Positions page={page} from={0} to={5} loses={everyFive} ringed={5} />
				),
				() => <Tiles from={0} to={23} loses={everyFive} />,
				({ page, at }) => (
					<Stack gap={10} align="start">
						<Line gap={6}>
							<Tile count={23} loses={false} size="sm" />
							<Move>{page.text(`${at}.kim`)}</Move>
							<Tile count={20} loses size="sm" ring="picked" />
						</Line>
						<Line gap={4}>
							{[20, 15, 10, 5, 0].map((count, place) => [
								place > 0 && <Then key={`then-${count}`} />,
								<Tile key={count} count={count} loses size="sm" />,
							])}
						</Line>
						<Say>{page.text(`${at}.answer`)}</Say>
					</Stack>
				),
			],
			note: ({ page, at }) => (
				<Line gap={6}>
					<Tile count={23} loses={false} size="sm" />
					<Then />
					<Tile count={19} loses={false} size="sm" ring="wrong" />
					<Move>{page.text(`${at}.leo`)}</Move>
					<Tile count={15} loses size="sm" />
					<Say way="wrong">{page.text(`${at}.lost`)}</Say>
				</Line>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={14} align="end">
					<Board king />
					<Say>{page.text(`${at}.moves`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Board near />
						<Say>{page.text(`${at}.green`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Board dots />
						<Say>{page.text(`${at}.red`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Board dots king />
						<Say>{page.text(`${at}.start`)}</Say>
					</>
				),
			],
		},
		{
			task: () => <Daisy petals={daisy()} />,
			steps: [
				({ page, at }) => (
					<>
						<Daisy petals={daisy({ 0: "torn" })} />
						<Say>{page.text(`${at}.eleven`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Daisy petals={rows()} />
						<Say>{page.text(`${at}.two-rows`)}</Say>
					</>
				),
				({ page, at }) => (
					<Line gap={12}>
						<Daisy petals={rows({ 2: "torn", 10: "torn" })} size={96} />
						<Say>{page.text(`${at}.mirror`)}</Say>
					</Line>
				),
			],
		},
	],
	traps: {
		ignored_condition: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Tile count={23} size="sm" />
							<Move wrong>{page.text(`${at}.five`)}</Move>
							<Tile count={18} size="sm" />
						</>
					}
					right={
						<>
							<Tile count={23} loses={false} size="sm" />
							<Move>{page.text(`${at}.three`)}</Move>
							<Tile count={20} loses size="sm" />
						</>
					}
				/>
			</>
		),
		first_move_assumed: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={<Chip tone="red">{page.text(`${at}.first`)}</Chip>}
					right={
						<>
							<Board size={3} dots king />
							<Chip tone="green">{page.text(`${at}.second`)}</Chip>
						</>
					}
				/>
			</>
		),
		best_case_not_worst: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Move>−4</Move>
							<Tile count={19} loses={false} size="sm" ring="wrong" />
							<Move>{page.text(`${at}.leo`)}</Move>
							<Tile count={15} loses size="sm" />
						</>
					}
					right={
						<>
							<Move>−3</Move>
							<Tile count={20} loses size="sm" ring="right" />
						</>
					}
				/>
			</>
		),
	},
};
