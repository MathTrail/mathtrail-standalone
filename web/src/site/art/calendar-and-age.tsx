import type { ComponentChildren } from "preact";
import {
	Box,
	Chip,
	Key,
	Line,
	Op,
	Row,
	Say,
	Stack,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";
import { Kind, lit } from "./pictures";

// Weeks is a block of whole weeks, with the days they make written on it.
function Weeks({
	wide = false,
	children,
}: {
	wide?: boolean;
	children: ComponentChildren;
}) {
	return (
		<span class="s-cal-weeks" data-wide={wide ? "" : undefined}>
			{children}
		</span>
	);
}

// Rest is whole weeks, then, a little apart, the days left over, ringed
// where the step finds them; a wide block holds many weeks at once.
function Rest({
	weeks,
	days,
	picked = false,
	wide = false,
}: {
	weeks: readonly ComponentChildren[];
	days: number;
	picked?: boolean;
	wide?: boolean;
}) {
	return (
		<Line gap={3}>
			{weeks.map((week, at) => (
				<Weeks key={at} wide={wide}>
					{week}
				</Weeks>
			))}
			<span class="s-cal-apart" />
			{Array.from({ length: days }, (_, at) => (
				<span
					key={at}
					class="s-cal-day"
					data-picked={picked ? "" : undefined}
				/>
			))}
		</Line>
	);
}

// Age is one person's age as a bar: their own years, then the years by which
// they are older, warm, and the age written after the bar.
function Age({
	name,
	own,
	more = 0,
	value,
}: {
	name: ComponentChildren;
	own: number;
	more?: number;
	value: string;
}) {
	return (
		<span class="s-cal-age">
			<span class="s-cal-age-name">{name}</span>
			<span class="s-cal-age-bar">
				<span
					class="s-cal-own"
					data-alone={more === 0 ? "" : undefined}
					style={{ "--s-length": `${own}px` }}
				/>
				{more > 0 && (
					<span class="s-cal-more" style={{ "--s-length": `${more}px` }} />
				)}
			</span>
			<span class="s-cal-age-value">{value}</span>
		</span>
	);
}

// Drop is one watering, numbered under its drop: red and empty where it is
// one too many.
function Drop({
	wrong = false,
	children,
}: {
	wrong?: boolean;
	children: ComponentChildren;
}) {
	return (
		<span class="s-cal-watering" data-wrong={wrong ? "" : undefined}>
			<span class="s-cal-drop" />
			<span class="s-cal-drop-name">{children}</span>
		</span>
	);
}

// Gap is the days between two waterings, written over the line between
// them: dashed and red where there is no such gap.
function Gap({
	short = false,
	wrong = false,
	children,
}: {
	short?: boolean;
	wrong?: boolean;
	children: ComponentChildren;
}) {
	return (
		<span
			class="s-cal-gap"
			data-short={short ? "" : undefined}
			data-wrong={wrong ? "" : undefined}
		>
			<span>{children}</span>
		</span>
	);
}

// Waterings are the four waterings and the gaps between them, each gap with
// what is written over it.
function Waterings({
	gap,
	short = false,
}: {
	gap: ComponentChildren;
	short?: boolean;
}) {
	return (
		<Line gap={1} align="start">
			{["1", "2", "3", "4"].map((one, at) => [
				at > 0 && (
					<Gap key={`gap-${one}`} short={short}>
						{gap}
					</Gap>
				),
				<Drop key={one}>{one}</Drop>,
			])}
		</Line>
	);
}

// AgesThen are two people's ages now and some years on, each time the
// length of the younger one's bar and the two ages, the difference between
// them the same warm length.
function AgesThen({
	words,
	young,
	old,
	now,
	then,
	more,
}: {
	words: { now: ComponentChildren; then: ComponentChildren };
	young: ComponentChildren;
	old: ComponentChildren;
	now: readonly [number, string, string];
	then: readonly [number, string, string];
	more: number;
}) {
	return (
		<Stack gap={8} align="start">
			<Say>{words.now}</Say>
			<Age name={young} own={now[0]} value={now[1]} />
			<Age name={old} own={now[0]} more={more} value={now[2]} />
			<Say>{words.then}</Say>
			<Age name={young} own={then[0]} value={then[1]} />
			<Age name={old} own={then[0]} more={more} value={then[2]} />
		</Stack>
	);
}

/**
 * art are the drawings of the calendar and age: a month whose same weekdays
 * stand in one column, whole weeks with the days left over, the waterings
 * and the gaps between them, and two ages whose difference stays the same.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Kind
				locale={page.locale}
				picture={{ kind: "calendar", first: 3, days: 30 }}
				tones={lit(["day 7", "day 14", "day 21", "day 28"], "cool")}
			/>
			<Say>{page.text(`${at}.tuesdays`)}</Say>
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={10}>
				<Rest
					weeks={Array.from({ length: 4 }, () => page.text(`${at}.week`))}
					days={2}
					picked
				/>
				<Say>{page.text(`${at}.month`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={10}>
				<Line gap={3}>
					{["3", "4", "5", "6", "7", "8", "9", "10"].map((day) => (
						<Box key={day} tone="cool" size="xs">
							{day}
						</Box>
					))}
				</Line>
				<Say>{page.text(`${at}.holidays`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<AgesThen
				words={{ now: page.text(`${at}.now`), then: page.text(`${at}.then`) }}
				young={page.text(`${at}.anya`)}
				old={page.text(`${at}.mum`)}
				now={[18, "6", "30"]}
				then={[48, "16", "40"]}
				more={72}
			/>
		),
	],
	idea: [
		{
			column: 1,
			showWords: true,
			rows: [
				({ page, at }) => <Rest weeks={[page.text(`${at}.week`)]} days={3} />,
				({ page, at }) => (
					<Rest
						weeks={[page.text(`${at}.week`), page.text(`${at}.week`)]}
						days={1}
					/>
				),
				({ page, at }) => (
					<Rest
						weeks={Array.from({ length: 4 }, () => page.text(`${at}.week`))}
						days={2}
					/>
				),
				({ page, at }) => (
					<Rest weeks={[page.text(`${at}.weeks`)]} days={2} wide />
				),
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key sample={<Weeks>{page.text(`${at}.week`)}</Weeks>}>
				{page.text(`${at}.whole`)}
			</Key>
			<Key sample={<span class="s-cal-day" />}>{page.text(`${at}.day`)}</Key>
			<Key sample={<span class="s-cal-swatch" />}>
				{page.text(`${at}.difference`)}
			</Key>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={8}>
					<Drop>{page.text(`${at}.monday`)}</Drop>
					<Say>{page.text(`${at}.first`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Waterings gap={page.text(`${at}.three-days`)} />
						<Say>{page.text(`${at}.gaps`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Rest weeks={[page.text(`${at}.week`)]} days={2} picked />
						<Say>9 = 7 + 2</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{
								kind: "calendar",
								first: 1,
								days: 30,
								marks: { "1": "1", "4": "2", "7": "3", "10": "4" },
							}}
							tones={{
								...lit(["day 1", "day 4", "day 7"], "cool"),
								"day 10": ["cool", "picked"],
							}}
						/>
						<Say>{page.text(`${at}.numbers`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<>
					<Line gap={1} align="start">
						<Waterings gap="3" short />
						<Gap short wrong>
							3?
						</Gap>
						<Drop wrong>?</Drop>
					</Line>
					<Say>{page.text(`${at}.no-gap`)}</Say>
				</>
			),
		},
		{
			steps: [
				({ page, at }) => (
					<Stack gap={8} align="start">
						<AgesThen
							words={{
								now: page.text(`${at}.now`),
								then: page.text(`${at}.then`),
							}}
							young={page.text(`${at}.vera`)}
							old={page.text(`${at}.dad`)}
							now={[30, "10", "38"]}
							then={[42, "14", "42"]}
							more={84}
						/>
						<Say>{page.text(`${at}.difference`)}</Say>
					</Stack>
				),
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{
								kind: "bars",
								bars: [
									{ segments: [{ size: 2, label: "?" }] },
									{
										segments: [{ size: 2 }, { size: 2 }, { size: 2 }],
										braces: [{ from: 1, to: 3, label: "28" }],
									},
								],
							}}
							tones={{
								...lit(["piece 0.0", "piece 1.0"], "cool"),
								...lit(["piece 1.1", "piece 1.2"], "warm"),
							}}
						/>
						<Say>{page.text(`${at}.parts`)}</Say>
					</>
				),
				({ page, at }) => (
					<Line gap={10}>
						<Box size="lg">14</Box>
						<Op>−</Op>
						<Box size="lg">4</Box>
						<Op>=</Op>
						<Box size="lg" ring="picked">
							10
						</Box>
						<Say>{page.text(`${at}.check`)}</Say>
					</Line>
				),
			],
		},
		{
			steps: [
				({ page, at }) => (
					<>
						<Line gap={6}>
							<Box tone="cool">2</Box>
							<Box tone="warm">9</Box>
							<Box tone="cool">16</Box>
							<Box tone="warm">23</Box>
							<Box tone="cool">30</Box>
						</Line>
						<Say>{page.text(`${at}.even`)}</Say>
					</>
				),
				({ page, at }) => (
					<Stack gap={10} align="start">
						<Row name={page.text(`${at}.first`)} size="wide" gap={4}>
							<Box>1</Box>
							<Box tone="cool" ring="picked">
								2
							</Box>
							<Box>3</Box>
						</Row>
						<Row name={page.text(`${at}.fifth`)} size="wide" gap={4}>
							<Say>2 + 28 =</Say>
							<Box tone="cool">30</Box>
						</Row>
					</Stack>
				),
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{ kind: "calendar", first: 5, days: 30 }}
							tones={{
								...lit(["day 2", "day 9", "day 16", "day 23", "day 30"], "cool"),
								"day 25": ["warm", "picked"],
							}}
						/>
						<Say>{page.text(`${at}.saturdays`)}</Say>
					</>
				),
			],
			note: () => (
				<Line gap={6}>
					<Box tone="cool">2</Box>
					<Op>+7</Op>
					<Box tone="warm">9</Box>
					<Op>+7</Op>
					<Box tone="cool">16</Box>
				</Line>
			),
		},
	],
	traps: {
		off_by_one: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.gaps`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">4 × 3</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								12
							</Box>
						</>
					}
					right={
						<>
							<Box size="sm">3 × 3</Box>
							<Op>=</Op>
							<Box size="sm" ring="right">
								9
							</Box>
						</>
					}
				/>
			</>
		),
		wrong_operation: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.older`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">30</Box>
							<Op>+</Op>
							<Box size="sm">6</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								36
							</Box>
						</>
					}
					right={
						<>
							<Box size="sm">30</Box>
							<Op>−</Op>
							<Box size="sm">6</Box>
							<Op>=</Op>
							<Box size="sm" ring="right">
								24
							</Box>
						</>
					}
				/>
			</>
		),
		ignored_condition: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.year`)}</Say>
				<Versus
					child={
						<>
							<Weeks wide>{page.text(`${at}.weeks`)}</Weeks>
							<Chip tone="red">{page.text(`${at}.monday`)}</Chip>
						</>
					}
					right={
						<>
							<Rest weeks={[page.text(`${at}.weeks`)]} days={1} wide />
							<Chip tone="green">{page.text(`${at}.tuesday`)}</Chip>
						</>
					}
				/>
			</>
		),
	},
};
