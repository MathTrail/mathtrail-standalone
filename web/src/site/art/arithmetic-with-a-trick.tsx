import type { ComponentChildren } from "preact";
import {
	Act,
	Box,
	Dots,
	Down,
	Group,
	Key,
	Line,
	Op,
	Row,
	Say,
	Stack,
	Then,
	Total,
	Unknown,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";
import { Kind, lit, pieces } from "./pictures";

// Area is a big product drawn as a rectangle of its parts: the grey block
// both products share, and the strip that is one more row or one more
// column, ringed where the step finds it and with what it adds under it. A
// faded block is one that cancels out.
function Area({
	strip,
	adds,
	picked = false,
	faded = false,
}: {
	strip: "side" | "under";
	adds: ComponentChildren;
	picked?: boolean;
	faded?: boolean;
}) {
	return (
		<span
			class="s-arith-area"
			data-strip={strip}
			data-faded={faded ? "" : undefined}
		>
			<span class="s-arith-block">2026 × 2025</span>
			<span class="s-arith-strip" data-picked={picked ? "" : undefined} />
			<span class="s-arith-adds">{adds}</span>
		</span>
	);
}

// Chain is Masha's chain of actions, from the number to the result, each
// box between two actions what the action gave, or an ellipsis where it is
// not yet known.
function Chain({
	start,
	between = ["…", "…"],
	end,
}: {
	start: ComponentChildren;
	between?: readonly [ComponentChildren, ComponentChildren];
	end?: ComponentChildren;
}) {
	return (
		<Line gap={4}>
			{start}
			<Then />
			<Act>× 2</Act>
			<Then />
			<Box>{between[0]}</Box>
			<Then />
			<Act>+ 7</Act>
			<Then />
			<Box>{between[1]}</Box>
			<Then />
			<Act>: 3</Act>
			<Then />
			{end ?? <Box>9</Box>}
		</Line>
	);
}

// Undone is the same chain undone from the end: the result first on the
// right, each action replaced by the one that undoes it, the arrows back.
function Undone({
	start,
	between = ["…", "…"],
	warm = false,
}: {
	start: ComponentChildren;
	between?: readonly [ComponentChildren, ComponentChildren];
	warm?: boolean;
}) {
	const tone = warm ? "warm" : "plain";
	return (
		<Line gap={4}>
			{start}
			<Then back tone="warm" />
			<Act tone="warm">: 2</Act>
			<Then back tone="warm" />
			<Box tone={tone}>{between[0]}</Box>
			<Then back tone="warm" />
			<Act tone="warm">− 7</Act>
			<Then back tone="warm" />
			<Box tone={tone}>{between[1]}</Box>
			<Then back tone="warm" />
			<Act tone="warm">× 3</Act>
			<Then back tone="warm" />
			<Box>9</Box>
		</Line>
	);
}

// tens are the pairs that make ten in the first example, each in its own
// tone: 6 and 4, 7 and 3, 8 and 2.
const tens = [
	["6", "4", "cool"],
	["7", "3", "warm"],
	["8", "2", "green"],
] as const;

/**
 * art are the drawings of arithmetic with a trick: numbers that make a round
 * one ringed together in one colour, a chain of actions undone from the end,
 * and a big product as a rectangle with a strip more.
 */
export const art: TopicArt = {
	pictures: true,
	heroLine: () => "19 + 23 + 81 + 77",
	hero: ({ page, at }) => (
		<>
			<Line gap={4}>
				<Box tone="cool">19</Box>
				<Op>+</Op>
				<Box tone="warm">23</Box>
				<Op>+</Op>
				<Box tone="cool">81</Box>
				<Op>+</Op>
				<Box tone="warm">77</Box>
			</Line>
			<Down>{page.text(`${at}.pairs`)}</Down>
			<Line gap={10}>
				<Group tone="cool" label="100">
					<Box tone="cool">19</Box>
					<Op>+</Op>
					<Box tone="cool">81</Box>
				</Group>
				<Op>+</Op>
				<Group tone="warm" label="100">
					<Box tone="warm">23</Box>
					<Op>+</Op>
					<Box tone="warm">77</Box>
				</Group>
				<Op>=</Op>
				<Box size="xl">200</Box>
			</Line>
			<Say>{page.text(`${at}.look`)}</Say>
		</>
	),
	basis: [
		() => (
			<Stack gap={12}>
				<Line gap={4}>
					<Box tone="cool">25</Box>
					<Op>×</Op>
					<Box>9</Box>
					<Op>×</Op>
					<Box tone="cool">4</Box>
				</Line>
				<Down />
				<Line gap={6}>
					<Group tone="cool" label="100">
						<Box tone="cool">25</Box>
						<Op>×</Op>
						<Box tone="cool">4</Box>
					</Group>
					<Op>×</Op>
					<Box>9</Box>
				</Line>
			</Stack>
		),
		() => (
			<Line gap={8}>
				<Group tone="cool" label="100">
					<Box tone="cool" size="xs">
						36
					</Box>
					<Op>+</Op>
					<Box tone="cool" size="xs">
						64
					</Box>
				</Group>
				<Group tone="warm" label="100">
					<Box tone="warm" size="xs">
						25
					</Box>
					<Op>×</Op>
					<Box tone="warm" size="xs">
						4
					</Box>
				</Group>
				<Group tone="green" label="1000">
					<Box tone="green" size="xs">
						125
					</Box>
					<Op>×</Op>
					<Box tone="green" size="xs">
						8
					</Box>
				</Group>
			</Line>
		),
		({ page, at }) => (
			<Stack gap={10}>
				<Line gap={6}>
					<Act>+ 3</Act>
					<Say>{page.text(`${at}.undoes`)}</Say>
					<Act tone="warm">− 3</Act>
				</Line>
				<Line gap={6}>
					<Act>× 2</Act>
					<Say>{page.text(`${at}.undoes`)}</Say>
					<Act tone="warm">: 2</Act>
				</Line>
				<Say>{page.text(`${at}.back`)}</Say>
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			rows: [
				() => (
					<Line gap={8}>
						<Group tone="cool" label="100">
							<Box tone="cool" size="xs">
								19
							</Box>
							<Op>+</Op>
							<Box tone="cool" size="xs">
								81
							</Box>
						</Group>
						<Op>+</Op>
						<Group tone="warm" label="100">
							<Box tone="warm" size="xs">
								23
							</Box>
							<Op>+</Op>
							<Box tone="warm" size="xs">
								77
							</Box>
						</Group>
					</Line>
				),
				() => (
					<Line gap={8}>
						<Group tone="cool" label="100">
							<Box tone="cool" size="xs">
								4
							</Box>
							<Op>×</Op>
							<Box tone="cool" size="xs">
								25
							</Box>
						</Group>
						<Op>×</Op>
						<Box size="xs">19</Box>
					</Line>
				),
				({ page, at }) => (
					<Line gap={12}>
						<Stack gap={6}>
							<Line gap={3}>
								{["1", "2", "3", "4", "5"].map((one) => (
									<Box key={one} tone="cool" size="xs">
										{one}
									</Box>
								))}
							</Line>
							<Line gap={3}>
								{["10", "9", "8", "7", "6"].map((one) => (
									<Box key={one} tone="warm" size="xs">
										{one}
									</Box>
								))}
							</Line>
							<Line gap={3}>
								{[1, 2, 3, 4, 5].map((one) => (
									<Total key={one}>11</Total>
								))}
							</Line>
						</Stack>
						<Say>{page.text(`${at}.pairs`)}</Say>
					</Line>
				),
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key
				sample={
					<Line gap={3}>
						<Box tone="cool" size="mini">
							19
						</Box>
						<Box tone="cool" size="mini">
							81
						</Box>
					</Line>
				}
			>
				{page.text(`${at}.pair`)}
			</Key>
			<Key sample={<Act tone="warm">− 7</Act>}>{page.text(`${at}.undo`)}</Key>
		</>
	),
	examples: [
		{
			steps: [
				({ page, at }) => (
					<>
						<Line gap={4}>
							<Box tone="cool">6</Box>
							<Op>+</Op>
							<Box tone="warm">7</Box>
							<Op>+</Op>
							<Box tone="green">8</Box>
							<Op>+</Op>
							<Box tone="green">2</Box>
							<Op>+</Op>
							<Box tone="warm">3</Box>
							<Op>+</Op>
							<Box tone="cool">4</Box>
						</Line>
						<Say>{page.text(`${at}.tens`)}</Say>
					</>
				),
				({ page }) => (
					<Kind
						locale={page.locale}
						picture={{
							kind: "bars",
							bars: tens.map(([one, other]) => ({
								parts: 10,
								shaded: 10,
								value: `${one} + ${other} = 10`,
							})),
						}}
						tones={Object.assign(
							{},
							...tens.map(([, , tone], bar) => lit(pieces(bar, 10), tone)),
						)}
					/>
				),
				() => (
					<Line gap={6}>
						<Box tone="cool">10</Box>
						<Op>+</Op>
						<Box tone="warm">10</Box>
						<Op>+</Op>
						<Box tone="green">10</Box>
						<Op>=</Op>
						<Box size="lg" ring="picked">
							30
						</Box>
					</Line>
				),
			],
			note: () => (
				<Line gap={4}>
					{["6", "13", "21", "23", "26", "30"].map((one, at) => [
						at > 0 && <Then key={`then-${one}`} />,
						<Box key={one} size="sm">
							{one}
						</Box>,
					])}
				</Line>
			),
		},
		{
			task: () => <Chain start={<Unknown />} />,
			steps: [
				({ page, at }) => (
					<Stack gap={10} align="start">
						<Row name={page.text(`${at}.masha`)} size="small" gap={4}>
							<Chain start={<Unknown />} />
						</Row>
						<Row name={page.text(`${at}.back`)} size="small" gap={4}>
							<Undone start={<Box>?</Box>} />
						</Row>
					</Stack>
				),
				() => <Undone start={<Unknown />} between={["20", "27"]} warm />,
				({ page, at }) => (
					<Stack gap={10} align="start">
						<Row name={page.text(`${at}.back`)} size="small" gap={4}>
							<Undone
								start={<Box tone="warm">10</Box>}
								between={["20", "27"]}
							/>
						</Row>
						<Row name={page.text(`${at}.check`)} size="small" gap={4}>
							<Chain
								start={<Box>10</Box>}
								between={["20", "27"]}
								end={<Box ring="right">9</Box>}
							/>
						</Row>
					</Stack>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={10} align="start">
					<Row name={page.text(`${at}.on`)} size="small" gap={8}>
						<Act>{page.text(`${at}.socks`)}</Act>
						<Then />
						<Act>{page.text(`${at}.shoes`)}</Act>
					</Row>
					<Row name={page.text(`${at}.off`)} size="small" gap={8}>
						<Act tone="warm">{page.text(`${at}.shoes`)}</Act>
						<Then tone="warm" />
						<Act tone="warm">{page.text(`${at}.socks`)}</Act>
					</Row>
				</Stack>
			),
		},
		{
			steps: [
				() => (
					<Stack gap={8}>
						<Say>2026 × 2026</Say>
						<Area strip="side" adds="+ 2026" picked />
					</Stack>
				),
				() => (
					<Stack gap={8}>
						<Say>2027 × 2025</Say>
						<Area strip="under" adds="+ 2025" picked />
					</Stack>
				),
				({ page, at }) => (
					<>
						<Line gap={10}>
							<Area strip="side" adds="+ 2026" faded />
							<Op>−</Op>
							<Area strip="under" adds="+ 2025" faded />
							<Op>=</Op>
							<Box size="lg" ring="picked">
								1
							</Box>
						</Line>
						<Say>{page.text(`${at}.cancel`)}</Say>
					</>
				),
			],
			note: () => (
				<Line gap={16}>
					<Stack gap={6}>
						<Dots count={25} columns={5} />
						<Say>5 × 5 = 25</Say>
					</Stack>
					<Op>−</Op>
					<Stack gap={6}>
						<Dots count={24} columns={6} tone="warm" />
						<Say>4 × 6 = 24</Say>
					</Stack>
					<Op>=</Op>
					<Box ring="right">1</Box>
				</Line>
			),
		},
	],
	traps: {
		stopped_early: () => (
			<>
				<Say>6 + 7 + 8 + 2 + 3 + 4</Say>
				<Versus
					child={
						<>
							<Box tone="cool" size="sm">
								10
							</Box>
							<Op>+</Op>
							<Box tone="warm" size="sm">
								10
							</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								20
							</Box>
						</>
					}
					right={
						<>
							<Box tone="cool" size="sm">
								10
							</Box>
							<Op>+</Op>
							<Box tone="warm" size="sm">
								10
							</Box>
							<Op>+</Op>
							<Box tone="green" size="sm">
								10
							</Box>
							<Op>=</Op>
							<Box size="sm" ring="right">
								30
							</Box>
						</>
					}
				/>
			</>
		),
		wrong_operation: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.undo`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">27</Box>
							<Act>+ 7</Act>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								34
							</Box>
						</>
					}
					right={
						<>
							<Box size="sm">27</Box>
							<Act tone="warm">− 7</Act>
							<Op>=</Op>
							<Box size="sm" ring="right">
								20
							</Box>
						</>
					}
				/>
			</>
		),
		number_from_text: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm" ring="wrong">
								9
							</Box>
							<Say>{page.text(`${at}.got`)}</Say>
						</>
					}
					right={
						<>
							<Box size="sm" ring="right">
								10
							</Box>
							<Say>{page.text(`${at}.thought`)}</Say>
						</>
					}
				/>
			</>
		),
	},
};
