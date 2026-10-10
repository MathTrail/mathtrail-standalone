import type { ComponentChildren } from "preact";
import type { Bar } from "../../design/picture/model";
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
	Stretch,
	Swatch,
	Then,
	Versus,
} from "../Sketch";
import type { TopicArt } from "./art";
import { Kind, lit, pieces } from "./pictures";

// lasts is a length of time in the page's words: minutes, whole hours, or
// both.
function lasts(page: PageReader, minutes: number): string {
	const hours = Math.floor(minutes / 60);
	const left = minutes % 60;
	if (hours === 0) {
		return page.plain("art.minutes", { count: left });
	}
	if (left === 0) {
		return page.plain("art.hours", { count: hours });
	}
	return page.plain("art.both", { hours, minutes: left });
}

// Face is a clock face as the card draws one, showing a time, its rim in the
// accent where the time is a whole hour.
function Face({
	page,
	time,
	size = "md",
	whole = false,
}: {
	page: PageReader;
	time: string;
	size?: "lg" | "md" | "sm";
	whole?: boolean;
}) {
	return (
		<span
			class="s-clock-face"
			data-size={size}
			data-whole={whole ? "" : undefined}
		>
			<Kind locale={page.locale} picture={{ kind: "clock", time }} />
		</span>
	);
}

// Times are moments on a line of time, the earlier first, with the arrows of
// the time between them: in one row, its arrows growing shorter where a
// narrow screen leaves no room.
function Times({ children }: { children: ComponentChildren }) {
	return <span class="s-clock-times">{children}</span>;
}

// Across is two moments, each over its words, and what goes from one to the
// other between them, level with the moments.
function Across({
	from,
	between,
	to,
}: {
	from: readonly [ComponentChildren, ComponentChildren];
	between: ComponentChildren;
	to: readonly [ComponentChildren, ComponentChildren];
}) {
	return (
		<span class="s-clock-across">
			{from[0]}
			{between}
			{to[0]}
			<Say>{from[1]}</Say>
			<span />
			<Say>{to[1]}</Say>
		</span>
	);
}

// tournament is a tournament's five rounds of 50 minutes and the four breaks
// of 15 between them, as one bar of a kind the card draws; rounds are the
// names of the rounds' pieces, and breaks of the breaks'.
const tournament: Bar = {
	segments: [50, 15, 50, 15, 50, 15, 50, 15, 50].map((size) => ({
		size,
		label: String(size),
	})),
};
const rounds = pieces(0, 9).filter((_, at) => at % 2 === 0);
const breaks = pieces(0, 9).filter((_, at) => at % 2 === 1);

/**
 * art are the drawings of clocks: faces that show the times, and lines of
 * time that stop at the whole hour, the length of each step written over its
 * arrow.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Line gap={8}>
				<Stack gap={6}>
					<Face page={page} time="4:50" size="lg" />
					<Box size="xs">4:50</Box>
				</Stack>
				<Then />
				<Stack gap={6}>
					<Face page={page} time="5:00" size="lg" whole />
					<Box tone="cool" size="xs">
						5:00
					</Box>
				</Stack>
				<Then />
				<Stack gap={6}>
					<Face page={page} time="5:15" size="lg" />
					<Box size="xs">5:15</Box>
				</Stack>
			</Line>
			<Times>
				<Box size="sm">4:50</Box>
				<Stretch length={76}>{lasts(page, 10)}</Stretch>
				<Box tone="cool" size="sm">
					5:00
				</Box>
				<Stretch length={96}>{lasts(page, 15)}</Stretch>
				<Box size="sm">5:15</Box>
			</Times>
			<Say>{page.text(`${at}.whole`)}</Say>
		</>
	),
	basis: [
		({ page }) => (
			<Stack gap={12}>
				<Line gap={8}>
					<Face page={page} time="8:50" size="sm" />
					<Op>+</Op>
					<Chip tone="cool">{lasts(page, 20)}</Chip>
					<Op>=</Op>
					<Face page={page} time="9:10" size="sm" />
				</Line>
				<Line gap={10}>
					<Box size="sm" ring="right">
						9:10
					</Box>
					<Box size="sm" ring="wrong">
						8:70
					</Box>
				</Line>
			</Stack>
		),
		({ page }) => (
			<Times>
				<Box size="sm">8:50</Box>
				<Stretch length={70}>{lasts(page, 10)}</Stretch>
				<Box tone="cool" size="sm">
					9:00
				</Box>
				<Stretch length={70}>{lasts(page, 10)}</Stretch>
				<Box size="sm">9:10</Box>
			</Times>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Times>
					<Box size="sm">22:00</Box>
					<Stretch length={56}>{lasts(page, 120)}</Stretch>
					<Box tone="cool" size="sm">
						0:00
					</Box>
					<Stretch length={96}>{lasts(page, 360)}</Stretch>
					<Box size="sm">6:00</Box>
				</Times>
				<Say>{page.text(`${at}.midnight`)}</Say>
			</Stack>
		),
	],
	idea: [
		{
			column: 1,
			rows: [
				({ page }) => (
					<Times>
						<Box size="sm">8:40</Box>
						<Stretch length={70}>{lasts(page, 20)}</Stretch>
						<Box tone="cool" size="sm">
							9:00
						</Box>
						<Stretch length={58}>{lasts(page, 15)}</Stretch>
						<Box size="sm">9:15</Box>
					</Times>
				),
				({ page }) => (
					<Times>
						<Box size="sm">10:50</Box>
						<Stretch length={54}>{lasts(page, 10)}</Stretch>
						<Box tone="cool" size="sm">
							11:00
						</Box>
						<Stretch length={110}>{lasts(page, 90)}</Stretch>
						<Box size="sm">12:30</Box>
					</Times>
				),
				({ page }) => (
					<Times>
						<Box size="sm">22:30</Box>
						<Stretch length={90}>{lasts(page, 90)}</Stretch>
						<Box tone="cool" size="sm">
							0:00
						</Box>
						<Stretch length={110}>{lasts(page, 360)}</Stretch>
						<Box size="sm">6:00</Box>
					</Times>
				),
			],
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key
				sample={
					<Box tone="cool" size="mini">
						9:00
					</Box>
				}
			>
				{page.text(`${at}.whole`)}
			</Key>
			<Key sample={<span class="s-clock-swatch" />}>
				{page.text(`${at}.lasts`)}
			</Key>
		</>
	),
	examples: [
		{
			task: ({ page, at }) => (
				<Line gap={16} align="end">
					<Stack gap={6}>
						<Face page={page} time="3:45" />
						<Box size="xs">3:45</Box>
					</Stack>
					<Say>{page.text(`${at}.start`)}</Say>
				</Line>
			),
			steps: [
				({ page }) => (
					<Line gap={14}>
						<Face page={page} time="4:00" whole />
						<Times>
							<Box size="sm">3:45</Box>
							<Stretch length={80}>{lasts(page, 15)}</Stretch>
							<Box tone="cool" size="sm" ring="picked">
								4:00
							</Box>
						</Times>
					</Line>
				),
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{
								kind: "bars",
								bars: [
									{
										segments: [
											{ size: 15, label: "15" },
											{ size: 25, label: "25" },
										],
									},
								],
							}}
							tones={{ "piece 0.1": ["cool", "picked"] }}
						/>
						<Say>{page.text(`${at}.split`)}</Say>
					</>
				),
				({ page }) => (
					<Line gap={14}>
						<Times>
							<Box size="sm">3:45</Box>
							<Stretch length={64}>{lasts(page, 15)}</Stretch>
							<Box tone="cool" size="sm">
								4:00
							</Box>
							<Stretch length={90}>{lasts(page, 25)}</Stretch>
							<Box size="sm" ring="picked">
								4:25
							</Box>
						</Times>
						<Face page={page} time="4:25" />
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Line gap={14}>
					<Stack gap={6}>
						<Box size="lg" ring="wrong">
							3:85
						</Box>
						<Say way="wrong">{page.text(`${at}.never`)}</Say>
					</Stack>
					<Stack gap={6}>
						<Box size="lg" ring="right">
							4:25
						</Box>
						<Say way="right">{page.text(`${at}.right`)}</Say>
					</Stack>
				</Line>
			),
		},
		{
			task: ({ page, at }) => (
				<Stack gap={8} align="start">
					<Kind
						locale={page.locale}
						picture={{ kind: "bars", bars: [tournament] }}
						tones={{ ...lit(rounds, "cool"), ...lit(breaks, "warm") }}
					/>
					<Line gap={14}>
						<Key sample={<Swatch />}>{page.text(`${at}.round`)}</Key>
						<Key sample={<Swatch tone="warm" length={10} />}>
							{page.text(`${at}.break`)}
						</Key>
					</Line>
				</Stack>
			),
			steps: [
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{ kind: "bars", bars: [tournament] }}
							tones={lit(rounds, "cool", "picked")}
						/>
						<Say>{page.text(`${at}.rounds`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Kind
							locale={page.locale}
							picture={{ kind: "bars", bars: [tournament] }}
							tones={{
								...lit(rounds, "cool"),
								...lit(breaks, "warm", "picked"),
							}}
						/>
						<Say>{page.text(`${at}.breaks`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Times>
							<Box tone="cool" size="sm">
								10:00
							</Box>
							<Stretch length={150}>{lasts(page, 300)}</Stretch>
							<Box tone="cool" size="sm">
								15:00
							</Box>
							<Stretch length={50}>{lasts(page, 10)}</Stretch>
							<Box size="sm" ring="picked">
								15:10
							</Box>
						</Times>
						<Say>{page.text(`${at}.total`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={8}>
					<Kind
						locale={page.locale}
						picture={{
							kind: "bars",
							bars: [
								{
									segments: [60, 60, 60, 60, 60, 10].map((size) => ({
										size,
										label: String(size),
									})),
								},
							],
						}}
						tones={{ ...lit(pieces(0, 5), "cool"), "piece 0.5": ["warm"] }}
					/>
					<Say>{page.text(`${at}.hours`)}</Say>
				</Stack>
			),
		},
		{
			task: ({ page, at }) => (
				<Line gap={16} align="end">
					<Stack gap={6}>
						<Face page={page} time="8:00" />
						<Say>{page.text(`${at}.take-off`)}</Say>
					</Stack>
					<Stack gap={6}>
						<Face page={page} time="8:30" />
						<Say>{page.text(`${at}.landing`)}</Say>
					</Stack>
				</Line>
			),
			steps: [
				({ page, at }) => (
					<Across
						from={[<Box key="8:00">8:00</Box>, page.text(`${at}.grandma-clock`)]}
						between={<Op>≠</Op>}
						to={[
							<Box key="8:30" tone="pale">
								8:30
							</Box>,
							page.text(`${at}.ann-clock`),
						]}
					/>
				),
				({ page, at }) => (
					<Across
						from={[
							<Box key="8:30" tone="pale">
								8:30
							</Box>,
							page.text(`${at}.at-ann`),
						]}
						between={
							<Stretch length={60}>{`+ ${lasts(page, 240)}`}</Stretch>
						}
						to={[
							<Box key="12:30" ring="picked">
								12:30
							</Box>,
							page.text(`${at}.at-grandma`),
						]}
					/>
				),
				({ page, at }) => (
					<Stack gap={12} align="start">
						<Row name={page.text(`${at}.grandma`)} gap={4}>
							<Times>
								<Box size="sm">8:00</Box>
								<Stretch length={130}>{lasts(page, 270)}</Stretch>
								<Box size="sm" ring="picked">
									12:30
								</Box>
							</Times>
						</Row>
						<Row name={page.text(`${at}.ann`)} gap={4}>
							<Times>
								<Box tone="pale" size="sm">
									4:00
								</Box>
								<Stretch length={130}>{lasts(page, 270)}</Stretch>
								<Box tone="pale" size="sm">
									8:30
								</Box>
							</Times>
						</Row>
					</Stack>
				),
			],
			note: ({ page, at }) => (
				<Line gap={10}>
					<Box tone="pale">8:30</Box>
					<Op>−</Op>
					<Box>8:00</Box>
					<Op>=</Op>
					<Box ring="wrong">0:30</Box>
					<Say way="wrong">{page.text(`${at}.other`)}</Say>
				</Line>
			),
		},
	],
	traps: {
		wrong_operation: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Box size="sm">3:00</Box>
							<Op>+</Op>
							<Box size="sm">5:00</Box>
							<Op>=</Op>
							<Box size="sm" ring="wrong">
								8:00
							</Box>
						</>
					}
					right={
						<Times>
							<Box size="sm">3:00</Box>
							<Stretch length={64}>{lasts(page, 120)}</Stretch>
							<Box size="sm">5:00</Box>
						</Times>
					}
				/>
			</>
		),
		time_unit_mixup: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<Box size="sm" ring="wrong">
							3:85
						</Box>
					}
					right={
						<Times>
							<Box size="sm">3:45</Box>
							<Stretch length={36}>15</Stretch>
							<Box tone="cool" size="sm">
								4:00
							</Box>
							<Stretch length={44}>25</Stretch>
							<Box size="sm" ring="right">
								4:25
							</Box>
						</Times>
					}
				/>
			</>
		),
		reversed_relation: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Times>
								<Box size="sm" ring="wrong">
									4:30
								</Box>
								<Stretch length={50} back wrong />
								<Box size="sm">8:00</Box>
							</Times>
							<Say way="wrong">{page.text(`${at}.before`)}</Say>
						</>
					}
					right={
						<Times>
							<Box size="sm">8:00</Box>
							<Stretch length={50} />
							<Box size="sm" ring="right">
								12:30
							</Box>
						</Times>
					}
				/>
			</>
		),
	},
};
