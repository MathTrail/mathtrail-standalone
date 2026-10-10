import type { ComponentChildren } from "preact";
import type { Bar } from "../../design/picture/model";
import type { Tone, Tones } from "../../design/picture/tones";
import type { PageReader } from "../reader";
import { Box, Key, Line, Op, Row, Say, Stack, Versus } from "../Sketch";
import type { TopicArt } from "./art";
import { Kind } from "./pictures";

// Share is one row of parts: who it is, if anyone, how many parts, what
// each holds, if anything, the tone of its parts, the parts the step picks
// counted from its end, what stands after it, and the brace under its
// picked parts.
type Share = {
	name?: string;
	parts: number;
	each?: string;
	tone: Tone;
	picked?: number;
	value?: string;
	brace?: string;
};

// barOf is a share as a bar of a kind the card draws.
function barOf(share: Share): Bar {
	const picked = share.picked ?? 0;
	return {
		label: share.name,
		segments: Array.from({ length: share.parts }, () => ({
			size: 1,
			label: share.each,
		})),
		value: share.value,
		braces:
			share.brace === undefined
				? undefined
				: [
						{
							from: share.parts - picked,
							to: share.parts,
							label: share.brace,
						},
					],
	};
}

// tonesOf are the tones of the shares' parts: each share's own, and picked
// on the parts the step picks.
function tonesOf(shares: readonly Share[]): Tones {
	return Object.fromEntries(
		shares.flatMap((share, place) =>
			Array.from({ length: share.parts }, (_, at) => [
				`piece ${place}.${at}`,
				at >= share.parts - (share.picked ?? 0)
					? [share.tone, "picked" as Tone]
					: [share.tone],
			]),
		),
	);
}

// Shares are rows of equal parts, one under another and to one scale, as
// bars of a kind the card draws; small where a table's cell holds them.
function Shares({
	page,
	shares,
	small = false,
}: {
	page: PageReader;
	shares: readonly Share[];
	small?: boolean;
}) {
	const picture = (
		<Kind
			locale={page.locale}
			picture={{ kind: "bars", bars: shares.map(barOf) }}
			tones={tonesOf(shares)}
		/>
	);
	if (!small) {
		return picture;
	}
	const units = Math.max(...shares.map((share) => share.parts));
	return (
		<span class="s-rat-small" style={{ "--s-units": String(units) }}>
			{picture}
		</span>
	);
}

// Person is one who shares, as a circle of their colour with their initial.
function Person({
	tone,
	children,
}: {
	tone: "warm" | "cool";
	children: ComponentChildren;
}) {
	return (
		<span class="s-rat-person" data-tone={tone}>
			{children}
		</span>
	);
}

// Ball is one part of a box's balls, a tile with what it holds, if anything.
function Ball({ children }: { children?: ComponentChildren }) {
	return <span class="s-rat-ball">{children}</span>;
}

// Crate is a box of balls, its tiles in it, with its name or its count
// under it.
function Crate({
	balls,
	each,
	children,
}: {
	balls: number;
	each?: string;
	children: ComponentChildren;
}) {
	return (
		<Stack gap={4}>
			<span class="s-rat-crate">
				{Array.from({ length: balls }, (_, at) => (
					<Ball key={at}>{each}</Ball>
				))}
			</span>
			<Say>{children}</Say>
		</Stack>
	);
}

// Crates are the two boxes side by side: their balls, each holding so many,
// and what is written under each.
function Crates({
	a,
	b,
	each,
	under,
	between,
}: {
	a: number;
	b: number;
	each?: string;
	under: readonly [ComponentChildren, ComponentChildren];
	between?: ComponentChildren;
}) {
	return (
		<Line gap={10} align="start">
			<Crate balls={a} each={each}>
				{under[0]}
			</Crate>
			{between}
			<Crate balls={b} each={each}>
				{under[1]}
			</Crate>
		</Line>
	);
}

// Moved is the part moved from one box to the other, on its arrow.
function Moved({ children }: { children: ComponentChildren }) {
	return (
		<span class="s-rat-moved">
			<span>{children}</span>
			<span class="s-rat-moved-arrow">→</span>
		</span>
	);
}

// Tile is a part as a tile, for a key: holding what one part holds, or
// framed as a part the task speaks of.
function Tile({
	framed = false,
	children,
}: {
	framed?: boolean;
	children?: ComponentChildren;
}) {
	return (
		<span class="s-rat-tile" data-framed={framed ? "" : undefined}>
			{children}
		</span>
	);
}

// Big is a sum written out large.
function Big({ children }: { children: ComponentChildren }) {
	return <span class="s-rat-big">{children}</span>;
}

/**
 * art are the drawings of ratios and sharing: each one's parts as a row of
 * equal parts to one scale, as bars of a kind the card draws, those the task
 * speaks of ringed; the two people who share; and two boxes of balls, before
 * a move and after it.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Shares
				page={page}
				shares={[
					{ name: page.plain("art.ann"), parts: 3, each: "5", tone: "warm" },
					{ name: page.plain("art.ben"), parts: 4, each: "5", tone: "cool" },
				]}
			/>
			<Big>{page.text(`${at}.parts`)}</Big>
			<Say>{page.text(`${at}.shares`)}</Say>
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={8}>
				<Shares
					page={page}
					shares={[
						{ parts: 3, tone: "warm" },
						{ parts: 4, tone: "cool" },
					]}
				/>
				<Say>{page.text(`${at}.seven`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Shares
					page={page}
					shares={[
						{ parts: 3, each: "5", tone: "warm", value: "= 15" },
						{ parts: 4, each: "5", tone: "cool", value: "= 20" },
					]}
				/>
				<Say>{page.text(`${at}.one`)}</Say>
			</Stack>
		),
		({ page }) => (
			<Shares
				page={page}
				shares={[
					{ name: "3 : 4", parts: 3, tone: "warm" },
					{ parts: 4, tone: "cool" },
					{ name: "6 : 8", parts: 6, tone: "warm" },
					{ parts: 8, tone: "cool" },
				]}
			/>
		),
	],
	idea: [
		{
			column: 1,
			words: "beside",
			rows: [
				({ page }) => (
					<Shares
						page={page}
						small
						shares={[
							{ parts: 3, tone: "warm" },
							{ parts: 4, tone: "cool" },
						]}
					/>
				),
				({ page }) => (
					<Shares
						page={page}
						small
						shares={[
							{ parts: 1, tone: "warm" },
							{ parts: 2, tone: "cool" },
							{ parts: 3, tone: "green" },
						]}
					/>
				),
				({ page }) => (
					<Shares
						page={page}
						small
						shares={[
							{ parts: 2, tone: "warm" },
							{ parts: 7, tone: "cool" },
						]}
					/>
				),
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key sample={<Tile>5</Tile>}>{page.text(`${at}.one`)}</Key>
			<Key sample={<Tile framed />}>{page.text(`${at}.spoken`)}</Key>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={8}>
					<Person tone="cool">{page.text(`${at}.max`)}</Person>
					<Say>4</Say>
					<Op>:</Op>
					<Person tone="warm">{page.text(`${at}.lena`)}</Person>
					<Say>5</Say>
					<Say>{page.text(`${at}.together`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Shares
							page={page}
							shares={[
								{ name: page.plain(`${at}.max-name`), parts: 4, tone: "cool" },
								{ name: page.plain(`${at}.lena-name`), parts: 5, tone: "warm" },
							]}
						/>
						<Say>{page.text(`${at}.nine`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Shares
							page={page}
							shares={[
								{
									name: page.plain(`${at}.max-name`),
									parts: 4,
									each: "5",
									tone: "cool",
								},
								{
									name: page.plain(`${at}.lena-name`),
									parts: 5,
									each: "5",
									tone: "warm",
								},
							]}
						/>
						<Say>{page.text(`${at}.each`)}</Say>
					</>
				),
				({ page, at }) => (
					<Shares
						page={page}
						shares={[
							{
								name: page.plain(`${at}.max-name`),
								parts: 4,
								each: "5",
								tone: "cool",
								value: "= 20",
							},
							{
								name: page.plain(`${at}.lena-name`),
								parts: 5,
								each: "5",
								tone: "warm",
								picked: 5,
								value: "= 25",
							},
						]}
					/>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Line gap={6}>
						<Box size="sm">45 : 5</Box>
						<Op>=</Op>
						<Box size="sm" ring="wrong">
							9
						</Box>
						<Say way="wrong">{page.text(`${at}.only-lena`)}</Say>
					</Line>
					<Line gap={6}>
						<Box size="sm">45 : 9</Box>
						<Op>=</Op>
						<Box size="sm" ring="right">
							5
						</Box>
					</Line>
				</Stack>
			),
		},
		{
			steps: [
				({ page, at }) => (
					<Shares
						page={page}
						shares={[
							{ name: page.plain(`${at}.boys`), parts: 4, tone: "cool" },
							{ name: page.plain(`${at}.girls`), parts: 7, tone: "warm", picked: 3 },
						]}
					/>
				),
				({ page, at }) => (
					<Shares
						page={page}
						shares={[
							{ name: page.plain(`${at}.boys`), parts: 4, tone: "cool" },
							{
								name: page.plain(`${at}.girls`),
								parts: 7,
								tone: "warm",
								picked: 3,
								brace: page.plain(`${at}.nine`),
							},
						]}
					/>
				),
				({ page, at }) => (
					<>
						<Shares
							page={page}
							shares={[
								{
									name: page.plain(`${at}.boys`),
									parts: 4,
									each: "3",
									tone: "cool",
									value: "= 12",
								},
								{
									name: page.plain(`${at}.girls`),
									parts: 7,
									each: "3",
									tone: "warm",
									value: "= 21",
								},
							]}
						/>
						<Say>{page.text(`${at}.choir`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Shares
					page={page}
					shares={[
						{
							name: page.plain(`${at}.girls`),
							parts: 7,
							tone: "warm",
							picked: 3,
							brace: page.plain(`${at}.this-nine`),
						},
					]}
				/>
			),
		},
		{
			task: ({ page, at }) => (
				<Crates
					a={3}
					b={1}
					under={[page.text(`${at}.box-a`), page.text(`${at}.box-b`)]}
				/>
			),
			steps: [
				({ page, at }) => (
					<Line gap={12}>
						<Crates
							a={3}
							b={1}
							under={[page.text("art.a"), page.text("art.b")]}
						/>
						<Say>{page.text(`${at}.four`)}</Say>
					</Line>
				),
				({ page, at }) => (
					<Stack gap={10} align="start">
						<Row name={page.text(`${at}.before`)} size="small" gap={10}>
							<Crates
								a={3}
								b={1}
								under={[page.text("art.a"), page.text("art.b")]}
							/>
						</Row>
						<Row name={page.text(`${at}.after`)} size="small" gap={10}>
							<Crates
								a={2}
								b={2}
								under={[page.text("art.a"), page.text("art.b")]}
								between={<Moved>{page.text(`${at}.part`)}</Moved>}
							/>
						</Row>
					</Stack>
				),
				({ page, at }) => (
					<Stack gap={10} align="start">
						<Row name={page.text(`${at}.before`)} size="small" gap={10}>
							<Crates a={3} b={1} each="10" under={["30", "10"]} />
						</Row>
						<Row name={page.text(`${at}.after`)} size="small" gap={10}>
							<Crates a={2} b={2} each="10" under={["20", "20"]} />
						</Row>
					</Stack>
				),
			],
		},
	],
	traps: {
		ratio_total_confusion: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">45 : 5</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								9
							</Box>
						</>
					}
					right={
						<>
							<Box size="sm">45 : 9</Box>
							<Op>=</Op>
							<Box size="sm" ring="right">
								5
							</Box>
							<Say>{page.text(`${at}.one`)}</Say>
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
							<Box size="sm">3 + 6</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								9
							</Box>
						</>
					}
					right={
						<>
							<Box size="sm">3 × 4</Box>
							<Op>=</Op>
							<Box size="sm" ring="right">
								12
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
							<Person tone="cool">{page.text(`${at}.max`)}</Person>
							<Box size="sm" ring="wrong">
								20
							</Box>
						</>
					}
					right={
						<>
							<Person tone="warm">{page.text(`${at}.lena`)}</Person>
							<Box size="sm" ring="right">
								25
							</Box>
						</>
					}
				/>
			</>
		),
	},
};
