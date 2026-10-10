import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import { Box, Key, Line, Op, Row, Say, Stack, Versus } from "../Sketch";
import type { TopicArt } from "./art";

// Thing is what stands in a row: a fence's post, a tree, or a bus stop.
type Thing = "post" | "tree" | "stop";

// Track is things in a row joined by a line, each thing's number over it
// where they are numbered and under each gap what is written there, the line
// in the accent where its gaps are what is counted; and, after the last
// thing, the gap that is not there, dashed red, to a stop that is not there.
function Track({
	thing,
	count,
	numbered = false,
	gaps,
	lit = false,
	beyond,
	size = "md",
}: {
	thing: Thing;
	count: number;
	numbered?: boolean;
	gaps?: (gap: number) => ComponentChildren;
	lit?: boolean;
	beyond?: ComponentChildren;
	size?: "md" | "sm";
}) {
	const columns: ComponentChildren[] = [];
	for (let place = 1; place <= count; place++) {
		columns.push(
			<span key={`over-${place}`} class="s-gap-over">
				{numbered ? place : ""}
			</span>,
			<span key={`thing-${place}`} class="s-gap-thing" />,
			<span key={`under-${place}`} class="s-gap-under" />,
		);
		if (place < count) {
			columns.push(
				<span key={`over-gap-${place}`} class="s-gap-over" />,
				<span key={`gap-${place}`} class="s-gap-line" />,
				<span key={`under-gap-${place}`} class="s-gap-under">
					{gaps?.(place)}
				</span>,
			);
		}
	}
	if (beyond !== undefined) {
		columns.push(
			<span key="over-beyond" class="s-gap-over" />,
			<span key="beyond" class="s-gap-line" data-beyond="" />,
			<span key="under-beyond" class="s-gap-under" data-beyond="">
				{beyond}
			</span>,
			<span key="over-ghost" class="s-gap-over" />,
			<span key="ghost" class="s-gap-thing" data-ghost="" />,
			<span key="under-ghost" class="s-gap-under" />,
		);
	}
	return (
		<span
			class="s-gap-track"
			data-thing={thing}
			data-lit={lit ? "" : undefined}
			data-size={size}
		>
			{columns}
		</span>
	);
}

// numbers are the gaps of a track numbered from 1.
const numbers = (gap: number) => gap;

// Bushes are bushes around a flowerbed, the gaps between them numbered
// where they are counted.
function Bushes({
	count,
	numbered = false,
	size = 120,
}: {
	count: number;
	numbered?: boolean;
	size?: number;
}) {
	const middle = size / 2;
	const ring = size * 0.3;
	const at = (turn: number, radius: number) => ({
		x: middle + radius * Math.sin(turn * 2 * Math.PI),
		y: middle - radius * Math.cos(turn * 2 * Math.PI),
	});
	return (
		<svg
			class="s-gap-bushes"
			width={size}
			height={size}
			viewBox={`0 0 ${size} ${size}`}
		>
			<circle class="s-gap-path" cx={middle} cy={middle} r={ring} />
			<circle class="s-gap-bed" cx={middle} cy={middle} r={ring * 0.5} />
			{Array.from({ length: count }, (_, place) => {
				const bush = at(place / count, ring);
				const gap = at((place + 0.5) / count, ring + size * 0.13);
				return (
					<g key={place}>
						<circle class="s-gap-bush" cx={bush.x} cy={bush.y} r={size * 0.055} />
						{numbered && (
							<text
								class="s-gap-number"
								x={gap.x}
								y={gap.y}
								text-anchor="middle"
								dominant-baseline="central"
							>
								{place + 1}
							</text>
						)}
					</g>
				);
			})}
		</svg>
	);
}

// Pages are the pages of a book from 1 to 8, the ones read, 3 to 7, lit,
// and over them, where steps are counted, the four steps from the first to
// the last with what they make.
function Pages({
	steps,
	size = "md",
}: {
	steps?: ComponentChildren;
	size?: "md" | "sm";
}) {
	return (
		<span class="s-gap-pages" data-size={size}>
			{steps !== undefined && (
				<span class="s-gap-hops">
					{[0, 1, 2, 3].map((hop) => (
						<span key={hop} class="s-gap-hop" />
					))}
					<span class="s-gap-hops-name">{steps}</span>
				</span>
			)}
			<span class="s-gap-book">
				{[1, 2, 3, 4, 5, 6, 7, 8].map((page) => (
					<span key={page} data-read={page >= 3 && page <= 7 ? "" : undefined}>
						{page}
					</span>
				))}
			</span>
		</span>
	);
}

// Log is a log sawn into so many pieces, a red cut between each two.
function Log({ pieces, size = "md" }: { pieces: number; size?: "md" | "sm" }) {
	return (
		<span class="s-gap-log" data-size={size}>
			{Array.from({ length: pieces }, (_, piece) => [
				piece > 0 && <span key={`cut-${piece}`} class="s-gap-cut" />,
				<span key={piece} class="s-gap-piece" />,
			])}
		</span>
	);
}

// Logs are logs one under another, each cut so many times.
function Logs({ cuts }: { cuts: readonly number[] }) {
	return (
		<Stack gap={6} align="start">
			{cuts.map((one, place) => (
				<Log key={place} pieces={one + 1} size="sm" />
			))}
		</Stack>
	);
}

// Girl is Masha or Lena as a circle of her own colour with her initial.
function Girl({ page, at, who }: { page: PageReader; at: string; who: "masha" | "lena" }) {
	return (
		<span class="s-gap-girl" data-who={who}>
			{page.text(`${at}.${who}`)}
		</span>
	);
}

// Climb is a climb up the floors of a house from the first to a top floor,
// the line of its flights running up through them, the top one lit; with
// whose climb it is under it, and beside it how many flights it takes.
function Climb({
	top,
	name,
	flights,
	unit,
}: {
	top: number;
	name: ComponentChildren;
	flights: number;
	unit: ComponentChildren;
}) {
	return (
		<Line gap={14} align="start">
			<Stack gap={6}>
				<span class="s-gap-floors">
					{Array.from({ length: top }, (_, below) => top - below).map((floor) => (
						<span key={floor} data-top={floor === top ? "" : undefined}>
							{floor}
						</span>
					))}
				</span>
				<Say>{name}</Say>
			</Stack>
			<Stack gap={2} align="start">
				<span class="s-gap-flights">{flights}</span>
				<Say>{unit}</Say>
			</Stack>
		</Line>
	);
}

// Stairs are Masha's two flights of stairs on a pill: one time Lena climbs
// as much.
function Stairs() {
	return (
		<span class="s-gap-stairs">
			{[0, 1].map((flight) => (
				<span key={flight} class="s-gap-stair">
					<span />
					<span />
					<span />
				</span>
			))}
		</span>
	);
}

// Count is a count written out in bold.
function Count({ children }: { children: ComponentChildren }) {
	return <span class="s-gap-count">{children}</span>;
}

/**
 * art are the drawings of gaps and boundaries: posts, trees and stops in a
 * row with the gaps between them, bushes around a bed, the pages read from
 * one to another, logs and their cuts, and the floors a stair climbs.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => (
		<>
			{page.text(`${at}.line`)}
			<span class="s-subject-rule">↓ {page.text(`${at}.rule`)}</span>
		</>
	),
	hero: ({ page, at }) => (
		<>
			<Track thing="post" count={4} numbered gaps={numbers} lit />
			<Line gap={16}>
				<Say>{page.text(`${at}.posts`)}</Say>
				<span class="s-gap-counted">{page.text(`${at}.spans`)}</span>
			</Line>
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={10}>
				<Track thing="tree" count={6} gaps={numbers} lit />
				<Say>{page.text(`${at}.trees`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={6}>
				<Bushes count={8} numbered />
				<Say>{page.text(`${at}.bushes`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Pages steps={page.text(`${at}.steps`)} />
				<Say>{page.text(`${at}.pages`)}</Say>
			</Stack>
		),
	],
	idea: [
		{
			column: 0,
			words: "beside",
			rows: [
				() => <Track thing="tree" count={6} gaps={numbers} lit size="sm" />,
				() => <Log pieces={5} />,
				() => <Bushes count={8} size={74} />,
				() => <Pages size="sm" />,
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key sample={<span class="s-gap-sample" data-of="counted" />}>
				{page.text(`${at}.counted`)}
			</Key>
			<Key sample={<span class="s-gap-sample" data-of="cut" />}>
				{page.text(`${at}.cut`)}
			</Key>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={12}>
					<span class="s-gap-bus" />
					<Track thing="stop" count={6} />
					<Say>{page.text(`${at}.stops`)}</Say>
				</Line>
			),
			steps: [
				() => <Track thing="stop" count={6} numbered />,
				() => <Track thing="stop" count={6} numbered gaps={numbers} lit />,
				({ page, at }) => (
					<>
						<Track
							thing="stop"
							count={6}
							numbered
							gaps={() => page.text(`${at}.minutes`, { count: 3 })}
							lit
						/>
						<Say>{page.text(`${at}.total`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<>
					<Track
						thing="stop"
						count={6}
						numbered
						gaps={() => 3}
						beyond="3?"
						size="sm"
					/>
					<Say way="wrong">{page.text(`${at}.end`)}</Say>
				</>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={14} align="end">
					<Stack gap={4}>
						<Girl page={page} at={at} who="masha" />
						<Say>{page.text(`${at}.third`)}</Say>
					</Stack>
					<Stack gap={4}>
						<Girl page={page} at={at} who="lena" />
						<Say>{page.text(`${at}.ninth`)}</Say>
					</Stack>
					<Say>{page.text(`${at}.from`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Climb
						top={3}
						name={page.text(`${at}.masha-name`)}
						flights={2}
						unit={page.text(`${at}.two`)}
					/>
				),
				({ page, at }) => (
					<Climb
						top={9}
						name={page.text(`${at}.lena-name`)}
						flights={8}
						unit={page.text(`${at}.eight`)}
					/>
				),
				({ page, at }) => (
					<>
						<Line gap={6}>
							{[0, 1, 2, 3].map((time) => (
								<Stairs key={time} />
							))}
						</Line>
						<Say>{page.text(`${at}.times`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Row name={page.text(`${at}.floors`)} gap={4}>
						<Box size="sm">9</Box>
						<Op>:</Op>
						<Box size="sm">3</Box>
						<Op>=</Op>
						<Box size="sm" ring="wrong">
							3
						</Box>
					</Row>
					<Row name={page.text(`${at}.flights`)} gap={4}>
						<Box size="sm">8</Box>
						<Op>:</Op>
						<Box size="sm">2</Box>
						<Op>=</Op>
						<Box size="sm" ring="right">
							4
						</Box>
					</Row>
				</Stack>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={12}>
					<Line gap={6}>
						<span class="s-gap-cut" />
						<Say>{page.text(`${at}.cuts`)}</Say>
					</Line>
					<Line gap={6}>
						<span class="s-gap-piece" />
						<Say>{page.text(`${at}.pieces`)}</Say>
					</Line>
					<Say>{page.text(`${at}.logs`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Stack gap={8} align="start">
						<Line gap={10}>
							<Log pieces={2} />
							<Say>{page.text(`${at}.one`)}</Say>
						</Line>
						<Line gap={10}>
							<Log pieces={3} />
							<Say>{page.text(`${at}.two`)}</Say>
						</Line>
					</Stack>
				),
				({ page, at }) => (
					<Stack gap={8} align="start">
						<Line gap={10}>
							<Log pieces={3} size="sm" />
							<Say>2 → 3</Say>
						</Line>
						<Line gap={10}>
							<Log pieces={2} size="sm" />
							<Say>1 → 2</Say>
						</Line>
						<Line gap={10}>
							<Log pieces={4} size="sm" />
							<Say>3 → 4</Say>
						</Line>
						<Say>{page.text(`${at}.sum`)}</Say>
					</Stack>
				),
				({ page, at }) => (
					<Stack gap={8}>
						<Logs cuts={[3, 2, 2, 1, 1]} />
						<Say>{page.text(`${at}.all`)}</Say>
					</Stack>
				),
			],
			note: () => (
				<Line gap={24} align="start">
					<Stack gap={8} align="start">
						<Logs cuts={[5, 1, 1, 1, 1]} />
						<Say>5 + 1 + 1 + 1 + 1</Say>
					</Stack>
					<Stack gap={8} align="start">
						<Logs cuts={[2, 2, 2, 2, 1]} />
						<Say>2 + 2 + 2 + 2 + 1</Say>
					</Stack>
				</Line>
			),
		},
	],
	traps: {
		off_by_one: () => (
			<Versus
				child={
					<>
						<Track
							thing="stop"
							count={6}
							gaps={() => 3}
							beyond="3"
							size="sm"
						/>
						<Count>18</Count>
					</>
				}
				right={
					<>
						<Track thing="stop" count={6} gaps={() => 3} lit size="sm" />
						<Count>15</Count>
					</>
				}
			/>
		),
		wrong_operation: () => (
			<Versus
				child={
					<>
						<Box size="sm">6</Box>
						<Op>+</Op>
						<Box size="sm">3</Box>
						<Op>=</Op>
						<Box size="sm" ring="wrong">
							9
						</Box>
					</>
				}
				right={
					<>
						<Box size="sm">5</Box>
						<Op>×</Op>
						<Box size="sm">3</Box>
						<Op>=</Op>
						<Box size="sm" ring="right">
							15
						</Box>
					</>
				}
			/>
		),
		number_from_text: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm" ring="wrong">
								6
							</Box>
							<Say>{page.text(`${at}.stops`)}</Say>
						</>
					}
					right={
						<>
							<Box size="sm" ring="right">
								15
							</Box>
							<Say>{page.text(`${at}.minutes`)}</Say>
						</>
					}
				/>
			</>
		),
	},
};
