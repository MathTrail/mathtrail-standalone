import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import { Box, Chip, Key, Line, Row, Say, Stack, Versus } from "../Sketch";
import type { TopicArt } from "./art";
import { Kind } from "./pictures";

// Pigeon is a pigeon as a round bird with an eye, the number of what it
// stands for on it where it has one; ringed where it is the one that
// decides.
function Pigeon({
	ringed = false,
	size = "md",
	children,
}: {
	ringed?: boolean;
	size?: "md" | "sm";
	children?: ComponentChildren;
}) {
	return (
		<span
			class="s-pig-pigeon"
			data-ringed={ringed ? "" : undefined}
			data-size={size}
			data-numbered={children === undefined ? undefined : ""}
		>
			{children}
		</span>
	);
}

// Pigeons are so many pigeons in a row.
function Pigeons({ count, size }: { count: number; size?: "md" | "sm" }) {
	return (
		<Line gap={size === "sm" ? 4 : 10}>
			{Array.from({ length: count }, (_, at) => (
				<Pigeon key={at} size={size} />
			))}
		</Line>
	);
}

// Cage is a cage and the pigeons in it, with its bars, or plain where it is
// a value a sum can take; lit where it is the one that holds two; with what
// it stands for under it where given.
function Cage({
	pigeons,
	lit = false,
	plain = false,
	size = "md",
	children,
}: {
	pigeons: number;
	lit?: boolean;
	plain?: boolean;
	size?: "md" | "sm";
	children?: ComponentChildren;
}) {
	return (
		<Stack gap={4}>
			<span
				class="s-pig-cage"
				data-lit={lit ? "" : undefined}
				data-plain={plain ? "" : undefined}
				data-size={size}
			>
				{Array.from({ length: pigeons }, (_, at) => (
					<Pigeon key={at} size="sm" />
				))}
			</span>
			{children !== undefined && (
				<span class="s-pig-under" data-lit={lit ? "" : undefined}>
					{children}
				</span>
			)}
		</Stack>
	);
}

// Cages are four cages for five pigeons, two in the first, lit, with what
// the first is said to hold under it.
function Cages({ two }: { two?: ComponentChildren }) {
	return (
		<Line gap={6} align="start">
			<Cage pigeons={2} lit>
				{two}
			</Cage>
			<Cage pigeons={1} />
			<Cage pigeons={1} />
			<Cage pigeons={1} />
		</Line>
	);
}

// weekdays are the short names of the days of the week in a language, from
// Monday, written as a name begins.
function weekdays(locale: string): string[] {
	const format = new Intl.DateTimeFormat(locale, {
		weekday: "short",
		timeZone: "UTC",
	});
	return Array.from({ length: 7 }, (_, day) => {
		const name = format.format(new Date(Date.UTC(2024, 0, 1 + day)));
		return name.charAt(0).toLocaleUpperCase(locale) + name.slice(1);
	});
}

// Week is a cage for each day of the week, a pigeon in each, its day's name
// under it.
function Week({ page, size = "md" }: { page: PageReader; size?: "md" | "sm" }) {
	return (
		<Line gap={3} align="start">
			{weekdays(page.locale).map((day) => (
				<Cage key={day} pigeons={1} size="sm">
					{size === "md" ? day : <span class="s-pig-day">{day}</span>}
				</Cage>
			))}
		</Line>
	);
}

// Ball is a ball, blue or red, ringed where it is the one that decides.
function Ball({
	red = false,
	ringed = false,
}: {
	red?: boolean;
	ringed?: boolean;
}) {
	return (
		<span
			class="s-pig-ball"
			data-red={red ? "" : undefined}
			data-ringed={ringed ? "" : undefined}
		/>
	);
}

// Balls are so many blue balls, then so many red ones, the last of them
// ringed where given.
function Balls({
	blue,
	red = 0,
	ringLast = false,
}: {
	blue: number;
	red?: number;
	ringLast?: boolean;
}) {
	return (
		<Line gap={3}>
			{Array.from({ length: blue }, (_, at) => (
				<Ball key={`blue-${at}`} />
			))}
			{Array.from({ length: red }, (_, at) => (
				<Ball key={`red-${at}`} red ringed={ringLast && at === red - 1} />
			))}
		</Line>
	);
}

// Candy is a wrapped sweet, lemon or orange, ringed where it decides.
function Candy({
	orange = false,
	ringed = false,
}: {
	orange?: boolean;
	ringed?: boolean;
}) {
	return (
		<span
			class="s-pig-candy"
			data-orange={orange ? "" : undefined}
			data-ringed={ringed ? "" : undefined}
		/>
	);
}

// Lemons are so many lemon sweets in a row.
function Lemons({ count }: { count: number }) {
	return (
		<Line gap={2}>
			{Array.from({ length: count }, (_, at) => (
				<Candy key={at} />
			))}
		</Line>
	);
}

// Jar is the jar of sweets, with so many lemon and so many orange in it.
function Jar({ lemons, oranges }: { lemons: number; oranges: number }) {
	return (
		<span class="s-pig-jar">
			{Array.from({ length: lemons }, (_, at) => (
				<Candy key={`lemon-${at}`} />
			))}
			{Array.from({ length: oranges }, (_, at) => (
				<Candy key={`orange-${at}`} orange />
			))}
		</span>
	);
}

// Shade is how a sock looks: black, white or striped.
type Shade = "black" | "white" | "striped";

// Sock is a sock of a shade, ringed where it decides or where it is wrongly
// taken.
function Sock({ kind, ring }: { kind: Shade; ring?: "picked" | "wrong" }) {
	return <span class="s-pig-sock" data-kind={kind} data-ring={ring} />;
}

// Socks are socks of kinds in a row: the counts of each, in order.
function Socks({
	white = 0,
	striped = 0,
	black = 0,
}: {
	white?: number;
	striped?: number;
	black?: number;
}) {
	return (
		<Line gap={2}>
			{Array.from({ length: white }, (_, at) => (
				<Sock key={`white-${at}`} kind="white" />
			))}
			{Array.from({ length: striped }, (_, at) => (
				<Sock key={`striped-${at}`} kind="striped" />
			))}
			{Array.from({ length: black }, (_, at) => (
				<Sock key={`black-${at}`} kind="black" />
			))}
		</Line>
	);
}

// drawer are the socks of the drawer in the order they lie, mixed.
const drawer: readonly Shade[] = [
	"black",
	"white",
	"black",
	"striped",
	"black",
	"white",
	"white",
	"black",
	"black",
	"white",
	"striped",
	"black",
	"white",
	"black",
	"white",
	"black",
];

// Drawer is the drawer of socks, mixed.
function Drawer() {
	return (
		<span class="s-pig-drawer">
			{drawer.map((kind, at) => (
				<Sock key={at} kind={kind} />
			))}
		</span>
	);
}

// Luck is a run of bad luck: what it brings, over a dashed red bracket with
// what it comes to under it.
function Luck({
	label,
	children,
}: {
	label: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<span class="s-pig-luck">
			<span class="s-pig-luck-run">{children}</span>
			<span class="s-pig-luck-label">{label}</span>
		</span>
	);
}

// Decides is the one more that decides the matter, ringed, with "+1" under it.
function Decides({ children }: { children: ComponentChildren }) {
	return (
		<span class="s-pig-decides">
			{children}
			<span class="s-pig-plus">+1</span>
		</span>
	);
}

// Count is a count written out in bold.
function Count({ children }: { children: ComponentChildren }) {
	return <span class="s-pig-count">{children}</span>;
}

// sums are the six sums of the grid 1 3 / 2 2, by the line they are taken
// along.
const sums = [
	["row", "4"],
	["row", "4"],
	["column", "3"],
	["column", "5"],
	["diagonal", "3"],
	["diagonal", "5"],
] as const;

// Grid is the grid of the example, 1 3 over 2 2, as a table the card draws.
function Grid({ page }: { page: PageReader }) {
	return (
		<Kind
			locale={page.locale}
			picture={{
				kind: "table",
				rows: [
					["1", "3"],
					["2", "2"],
				],
			}}
		/>
	);
}

// Values are the five sums two numbers from 1, 2 and 3 can make, each a
// cage, the pigeons the six sums put in each.
function Values({
	pigeons = [0, 0, 0, 0, 0],
	lit = false,
}: {
	pigeons?: readonly number[];
	lit?: boolean;
}) {
	return (
		<Line gap={6} align="start">
			{["2", "3", "4", "5", "6"].map((value, at) => (
				<Cage
					key={value}
					pigeons={pigeons[at] ?? 0}
					lit={lit && (pigeons[at] ?? 0) > 1}
					plain
				>
					{value}
				</Cage>
			))}
		</Line>
	);
}

/**
 * art are the drawings of the pigeonhole principle: pigeons and the cages
 * they sit in, the one with two lit; runs of bad luck over a dashed red
 * bracket and the one more that decides; balls, sweets in a jar and socks in
 * a drawer; and the sums of a grid as pigeons in the cages of their values.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Pigeons count={5} />
			<Say>↓ {page.text(`${at}.seat`)}</Say>
			<Cages two={page.text(`${at}.two`)} />
			<Say>{page.text(`${at}.always`)}</Say>
		</>
	),
	basis: [
		() => (
			<Stack gap={8}>
				<Pigeons count={5} size="sm" />
				<Cages />
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={8} align="start">
				<Row name={page.text(`${at}.luck`)} size="wide" gap={6}>
					<Balls blue={0} red={2} />
					<Chip tone="grey">{page.text(`${at}.not-counted`)}</Chip>
				</Row>
				<Row name={page.text(`${at}.bad-luck`)} size="wide" gap={6}>
					<Balls blue={6} red={1} ringLast />
				</Row>
			</Stack>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Decides>
					<Pigeon ringed />
				</Decides>
				<Week page={page} />
				<Say>{page.text(`${at}.eighth`)}</Say>
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			words: "under",
			rows: [
				({ page }) => <Week page={page} size="sm" />,
				() => <Balls blue={6} />,
				() => <Balls blue={6} red={1} />,
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key sample={<span class="s-pig-luck-sample" />}>
				{page.text(`${at}.bad-luck`)}
			</Key>
			<Key sample={<span class="s-pig-decides-sample" />}>
				{page.text(`${at}.decides`)}
			</Key>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={14} align="end">
					<Jar lemons={8} oranges={2} />
					<Stack gap={4} align="start">
						<Key sample={<Candy />}>{page.text(`${at}.lemon`)}</Key>
						<Key sample={<Candy orange />}>{page.text(`${at}.orange`)}</Key>
					</Stack>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Luck label={page.text(`${at}.unlucky`)}>
						<Lemons count={3} />
						<span class="s-pig-more">…</span>
					</Luck>
				),
				({ page, at }) => (
					<Line gap={10}>
						<Luck label={page.text(`${at}.all-lemons`)}>
							<Lemons count={8} />
						</Luck>
						<Jar lemons={0} oranges={2} />
					</Line>
				),
				({ page, at }) => (
					<>
						<Line gap={6} align="start">
							<Luck label={page.text(`${at}.lemons`)}>
								<Lemons count={8} />
							</Luck>
							<Decides>
								<Candy orange ringed />
							</Decides>
						</Line>
						<Say>8 + 1 = 9</Say>
					</>
				),
			],
		},
		{
			task: ({ page, at }) => (
				<Line gap={14}>
					<Drawer />
					<Stack gap={4} align="start">
						<Key sample={<Sock kind="black" />}>{page.text(`${at}.black`)}</Key>
						<Key sample={<Sock kind="white" />}>{page.text(`${at}.white`)}</Key>
						<Key sample={<Sock kind="striped" />}>
							{page.text(`${at}.striped`)}
						</Key>
					</Stack>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Luck label={page.text(`${at}.no-black`)}>
						<Socks white={6} striped={2} />
					</Luck>
				),
				({ page, at }) => (
					<Line gap={6} align="start">
						<Luck label={page.text(`${at}.not-black`)}>
							<Socks white={6} striped={2} />
						</Luck>
						<Luck label={page.text(`${at}.one`)}>
							<Sock kind="black" />
						</Luck>
					</Line>
				),
				({ page, at }) => (
					<>
						<Line gap={6} align="start">
							<Luck label={page.text(`${at}.not-black`)}>
								<Socks white={6} striped={2} />
							</Luck>
							<Luck label={page.text(`${at}.one`)}>
								<Sock kind="black" />
							</Luck>
							<Decides>
								<Sock kind="black" ring="picked" />
							</Decides>
						</Line>
						<Say>8 + 1 + 1 = 10</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<>
					<Socks white={6} striped={2} black={1} />
					<Say way="wrong">{page.text(`${at}.only-one`)}</Say>
				</>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={14} align="end">
					<Grid page={page} />
					<Say>{page.text(`${at}.example`)}</Say>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<>
						<Values />
						<Say>{page.text(`${at}.cages`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Line gap={12}>
							<Grid page={page} />
							<Line gap={4} align="end">
								{sums.map(([line, value], place) => (
									<Stack key={`${line}-${place}`} gap={4}>
										<span class="s-pig-line">{page.text(`${at}.${line}`)}</span>
										<Pigeon>{value}</Pigeon>
									</Stack>
								))}
							</Line>
						</Line>
						<Say>{page.text(`${at}.pigeons`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Values pigeons={[0, 2, 2, 2, 0]} lit />
						<Say>{page.text(`${at}.two`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Row name={page.text(`${at}.cages-name`)} gap={4}>
						{["2", "3", "4", "5", "6"].map((value) => (
							<Box key={value} size="xs">
								{value}
							</Box>
						))}
					</Row>
					<Row name={page.text(`${at}.pigeons-name`)} gap={4}>
						<Pigeons count={6} size="sm" />
					</Row>
				</Stack>
			),
		},
	],
	traps: {
		best_case_not_worst: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Socks black={2} />
							<Count>= 2</Count>
						</>
					}
					right={
						<>
							<Socks white={6} striped={2} black={2} />
							<Count>= 10</Count>
						</>
					}
				/>
			</>
		),
		ignored_condition: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Line gap={2}>
								<Sock kind="black" />
								<Sock kind="white" />
								<Sock kind="striped" />
								<Sock kind="white" ring="wrong" />
							</Line>
							<Count>= 4</Count>
						</>
					}
					right={
						<>
							<Count>… = 10</Count>
							<Say>{page.text(`${at}.black-pair`)}</Say>
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
							<Lemons count={8} />
							<Count>= 8</Count>
						</>
					}
					right={
						<>
							<Lemons count={8} />
							<Candy orange ringed />
							<Count>= 9</Count>
						</>
					}
				/>
			</>
		),
	},
};
