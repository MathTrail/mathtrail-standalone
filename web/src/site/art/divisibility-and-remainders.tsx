import type { ComponentChildren } from "preact";
import type { Tones } from "../../design/picture/tones";
import type { PageReader } from "../reader";
import {
	Box,
	Chip,
	Key,
	Line,
	Op,
	Say,
	Stack,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";
import { Kind, lit } from "./pictures";

// Pack is one group of a division, its counters as dots: a full group, or
// what is left over, dashed and warm; ringed where the question asks for it.
function Pack({
	count,
	rest = false,
	picked = false,
	size = "sm",
}: {
	count: number;
	rest?: boolean;
	picked?: boolean;
	size?: "md" | "sm";
}) {
	return (
		<span
			class="s-div-pack"
			data-rest={rest ? "" : undefined}
			data-picked={picked ? "" : undefined}
			data-size={size}
		>
			{Array.from({ length: count }, (_, at) => (
				<span key={at} />
			))}
		</span>
	);
}

// Packs are so many full groups of a size and what is left over, ringed as
// the question asks: the groups, what is left over, or both.
function Packs({
	groups,
	size,
	rest,
	picked = [],
}: {
	groups: number;
	size: number;
	rest: number;
	picked?: readonly ("groups" | "rest")[];
}) {
	return (
		<Line gap={4}>
			{Array.from({ length: groups }, (_, at) => (
				<Pack key={at} count={size} picked={picked.includes("groups")} />
			))}
			<Pack count={rest} rest picked={picked.includes("rest")} />
		</Line>
	);
}

// Car is a car of four seats seen from the side, its seats taken by so many
// pupils; ringed where it is the car the remainder needs.
function Car({ taken, picked = false }: { taken: number; picked?: boolean }) {
	return (
		<span class="s-div-car" data-picked={picked ? "" : undefined}>
			<span class="s-div-car-body">
				{[0, 1, 2, 3].map((seat) => (
					<span key={seat} data-taken={seat < taken ? "" : undefined} />
				))}
			</span>
			<span class="s-div-wheel" />
			<span class="s-div-wheel" />
		</span>
	);
}

// Bus is a bus of 40 seats seen from the side, filled as far as its pupils
// fill it, with how many ride in it on its side: cool where it is full, warm
// where it is not; ringed where it is the bus the remainder needs, and small
// in a trap's row.
function Bus({
	pupils,
	picked = false,
	size = "md",
}: {
	pupils: number;
	picked?: boolean;
	size?: "md" | "sm";
}) {
	return (
		<span
			class="s-div-bus"
			data-full={pupils === 40 ? "" : undefined}
			data-picked={picked ? "" : undefined}
			data-size={size}
		>
			<span class="s-div-bus-body">
				<span
					class="s-div-bus-fill"
					style={{ "--s-share": `${(pupils / 40) * 100}%` }}
				/>
				<span class="s-div-bus-pupils">{pupils}</span>
			</span>
			<span class="s-div-wheel" />
			<span class="s-div-wheel" />
		</span>
	);
}

// Blank is a digit the task hides behind a star, or the one a step puts in
// its place, in a dashed box.
function Blank({ children }: { children: ComponentChildren }) {
	return <span class="s-div-blank">{children}</span>;
}

// Digits are the digits of 38*6, the third one given.
function Digits({ third }: { third: ComponentChildren }) {
	return (
		<Line gap={4}>
			<Box size="lg">3</Box>
			<Box size="lg">8</Box>
			<Blank>{third}</Blank>
			<Box size="lg">6</Box>
		</Line>
	);
}

// Sums is the sums of digits 38*6 can have, from 17 to 26, on a number line,
// the ones a step finds lit.
function Sums({ page, tones }: { page: PageReader; tones?: Tones }) {
	return (
		<Kind
			locale={page.locale}
			picture={{ kind: "number_line", from: 17, to: 26 }}
			tones={tones}
		/>
	);
}

// Cycle is one round of the last digits of the powers of two, 2, 4, 8 and 6,
// the last round outlined and its 6 ringed.
function Cycle({ last = false }: { last?: boolean }) {
	return (
		<span class="s-div-cycle" data-last={last ? "" : undefined}>
			{["2", "4", "8", "6"].map((digit) => (
				<Box
					key={digit}
					size="mini"
					ring={last && digit === "6" ? "picked" : undefined}
				>
					{digit}
				</Box>
			))}
		</span>
	);
}

// Sum is a division written out large: a number as full groups and what is
// left over, the remainder warm where it is drawn so.
function Sum({
	size = "md",
	children,
}: {
	size?: "lg" | "md";
	children: ComponentChildren;
}) {
	return (
		<span class="s-div-sum" data-size={size}>
			{children}
		</span>
	);
}

// Rule is one rule of divisibility: what it divides by on a pill, then the
// rule with an example, the digits it looks at in the accent.
function Rule({
	by,
	children,
}: {
	by: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<span class="s-div-rule">
			<Chip tone="grey">{by}</Chip>
			<span>{children}</span>
		</span>
	);
}

/**
 * art are the drawings of divisibility and remainders: counters in full
 * groups and what is left over, cars and buses with the one the remainder
 * needs, the sums of digits on a number line, and the last digits of the
 * powers of two, which go round by four.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Line gap={4}>
				{[0, 1, 2, 3].map((group) => (
					<Pack key={group} count={3} size="md" />
				))}
				<Pack count={2} rest size="md" />
			</Line>
			<Sum size="lg">
				14 = 4 × 3 + <span class="s-div-rest">2</span>
			</Sum>
			<Say>{page.text(`${at}.groups`)}</Say>
		</>
	),
	basis: [
		({ page, at }) => (
			<Stack gap={10}>
				<Kind
					locale={page.locale}
					picture={{ kind: "number_line", from: 0, to: 7 }}
					tones={{
						...lit(
							["tick 0", "tick 1", "tick 2", "tick 3", "tick 4", "tick 5", "tick 6"],
							"cool",
						),
						"tick 7": ["struck"],
					}}
				/>
				<Say>{page.text(`${at}.remainders`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={10}>
				<Line gap={4}>
					<Line gap={4}>
						{[0, 1, 2, 3].map((car) => (
							<Car key={car} taken={4} />
						))}
					</Line>
					<Line gap={4}>
						{[4, 5, 6].map((car) => (
							<Car key={car} taken={4} />
						))}
						<Car taken={1} picked />
					</Line>
				</Line>
				<Say>{page.text(`${at}.cars`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={10} align="start">
				<Rule by={page.text(`${at}.by-two`)}>{page.text(`${at}.even`)}</Rule>
				<Rule by={page.text(`${at}.by-five`)}>
					{page.text(`${at}.five`)}
				</Rule>
				<Rule by={page.text(`${at}.by-three`)}>
					{page.text(`${at}.digits`)}
				</Rule>
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			words: "under",
			rows: [
				() => <Packs groups={7} size={4} rest={1} picked={["groups", "rest"]} />,
				() => <Packs groups={7} size={4} rest={1} picked={["groups"]} />,
				() => <Packs groups={7} size={4} rest={1} picked={["rest"]} />,
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key sample={<Pack count={3} />}>{page.text(`${at}.group`)}</Key>
			<Key sample={<Pack count={2} rest />}>{page.text(`${at}.rest`)}</Key>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={16} align="end">
					<Stack gap={4}>
						<Bus pupils={40} />
						<Say>{page.text(`${at}.seats`)}</Say>
					</Stack>
					<Say>{page.text(`${at}.pupils`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Line gap={8} align="end">
						<Bus pupils={40} />
						<Bus pupils={40} />
						<Bus pupils={40} />
						<Stack gap={4}>
							<span class="s-div-left">30</span>
							<Say>{page.text(`${at}.left`)}</Say>
						</Stack>
					</Line>
				),
				({ page, at }) => (
					<Stack gap={4}>
						<Bus pupils={30} picked />
						<Say>{page.text(`${at}.part`)}</Say>
					</Stack>
				),
				() => (
					<>
						<Line gap={8}>
							<Bus pupils={40} />
							<Bus pupils={40} />
							<Bus pupils={40} />
							<Bus pupils={30} />
						</Line>
						<Say>[40] [40] [40] [30]</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Line gap={14}>
					<Stack gap={6}>
						<Box size="lg" ring="wrong">
							3
						</Box>
						<Say way="wrong">{page.text(`${at}.full`)}</Say>
					</Stack>
					<Stack gap={6}>
						<Box size="lg" ring="wrong">
							30
						</Box>
						<Say way="wrong">{page.text(`${at}.last`)}</Say>
					</Stack>
					<Stack gap={6}>
						<Box size="lg" ring="right">
							4
						</Box>
						<Say way="right">{page.text(`${at}.all`)}</Say>
					</Stack>
				</Line>
			),
		},
		{
			task: () => <Digits third="*" />,
			steps: [
				({ page, at }) => (
					<>
						<Line gap={6}>
							<Box size="lg">3</Box>
							<Op>+</Op>
							<Box size="lg">8</Box>
							<Op>+</Op>
							<Blank>*</Blank>
							<Op>+</Op>
							<Box size="lg">6</Box>
						</Line>
						<Say>{page.text(`${at}.sum`)}</Say>
					</>
				),
				({ page }) => (
					<>
						<Sums page={page} />
						<Say>17 + 0 … 17 + 9</Say>
					</>
				),
				({ page }) => (
					<>
						<Sums page={page} tones={{ "tick 18": ["cool", "picked"] }} />
						<Digits third="1" />
					</>
				),
			],
			note: ({ page }) => (
				<Sums page={page} tones={lit(["tick 18", "tick 21", "tick 24"], "green")} />
			),
		},
		{
			steps: [
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{
								kind: "table",
								header: ["1", "2", "3", "4", "5", "6", "7", "8"],
								rows: [
									["2", "4", "8", "16", "32", "64", "128", "256"],
									["2", "4", "8", "6", "2", "4", "8", "6"],
								],
							}}
							tones={lit(["cell 1.0", "cell 1.1", "cell 1.2", "cell 1.3"], "cool")}
						/>
						<Say>{page.text(`${at}.first`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Line gap={6}>
							<Cycle />
							<Cycle />
							<Cycle />
							<span class="s-div-more">…</span>
							<Cycle last />
						</Line>
						<Say>{page.text(`${at}.rounds`)}</Say>
					</>
				),
				({ page }) => (
					<Kind
						locale={page.locale}
						picture={{
							kind: "table",
							header: ["1", "2", "3", "4"],
							rows: [["2", "4", "8", "6"]],
						}}
						tones={{ "cell 0.3": ["cool", "picked"] }}
					/>
				),
			],
			note: ({ page, at }) => (
				<>
					<Sum>100 = 25 × 4 + 0</Sum>
					<Say>{page.text(`${at}.place`)}</Say>
				</>
			),
		},
	],
	traps: {
		remainder_vs_quotient: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm" ring="wrong">
								30
							</Box>
							<Say>{page.text(`${at}.left`)}</Say>
						</>
					}
					right={
						<>
							<Box size="sm" ring="right">
								4
							</Box>
							<Say>{page.text(`${at}.buses`)}</Say>
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
						<>
							<Bus pupils={40} size="sm" />
							<Bus pupils={40} size="sm" />
							<Bus pupils={40} size="sm" />
							<Sum>= 3</Sum>
						</>
					}
					right={
						<>
							<Bus pupils={40} size="sm" />
							<Bus pupils={40} size="sm" />
							<Bus pupils={40} size="sm" />
							<Bus pupils={30} size="sm" />
							<Sum>= 4</Sum>
						</>
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
							<Blank>1</Blank>
							<Say way="wrong">{page.text(`${at}.more`)}</Say>
						</>
					}
					right={
						<>
							<Blank>1</Blank>
							<Blank>4</Blank>
							<Blank>7</Blank>
						</>
					}
				/>
			</>
		),
	},
};
