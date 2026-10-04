import type { Topic, Topics } from "./topics";

/**
 * board is the size of what the map draws, in pixels: a topic's box, the gaps
 * between the columns and the rows, and the lanes a line that skips a column
 * runs in under all of them. The map is laid out once, when the site is
 * built, so a page needs no script to draw its lines.
 */
export const board = {
	width: 248,
	height: 64,
	columnGap: 96,
	rowGap: 12,
	lane: 10,
	under: 28,
	corner: 10,
} as const;

/** Spot is where a topic's box stands on the map: its column and its row. */
export type Spot = {
	readonly topic: Topic;
	readonly x: number;
	readonly y: number;
};

/**
 * Line is one link of the map, from the topic built on to the topic built on
 * it, as the path an SVG draws, and where it ends.
 */
export type Line = {
	readonly from: Topic;
	readonly to: Topic;
	readonly path: string;
	readonly end: readonly [number, number];
};

/**
 * TopicMap is the map of how the topics link: a column for each layer, the
 * foundations first, each topic's box, every link a line, and the size of it
 * all.
 */
export type TopicMap = {
	readonly columns: readonly (readonly Topic[])[];
	readonly spots: readonly Spot[];
	readonly lines: readonly Line[];
	readonly width: number;
	readonly height: number;
};

/**
 * mapOf lays the topics out: each in the column of its layer, the foundations
 * in the catalog's order and every other column by where its topics' bases
 * stand, so that lines cross as little as a simple rule allows. A link to the
 * next column is a curve between the two boxes; a link that skips a column
 * runs down a gutter, along a lane of its own under every box, and up the next
 * gutter, as the author's draft draws it.
 */
export function mapOf(topics: Topics): TopicMap {
	const columns = columnsOf(topics);
	const rows = new Map(
		columns.flatMap((column) => column.map((topic, row) => [topic.id, row])),
	);
	const spotOf = (topic: Topic): Spot => ({
		topic,
		x: topic.layer * (board.width + board.columnGap),
		y: (rows.get(topic.id) ?? 0) * (board.height + board.rowGap),
	});
	const spots = columns.flat().map(spotOf);
	const tallest = Math.max(...columns.map((column) => column.length));
	const bottom = tallest * (board.height + board.rowGap) - board.rowGap;
	const links = topics.all.flatMap((to) =>
		to.bases.flatMap((base) => {
			const from = topics.byId.get(base);
			return from === undefined ? [] : [{ from, to }];
		}),
	);
	const skipping = links.filter(({ from, to }) => to.layer - from.layer > 1);
	const lines = links.map(({ from, to }): Line => {
		const [x1, y1] = rightOf(spotOf(from));
		const [x2, y2] = leftOf(spotOf(to));
		const lane = skipping.findIndex(
			(link) => link.from === from && link.to === to,
		);
		return {
			from,
			to,
			path:
				lane < 0
					? curve(x1, y1, x2, y2)
					: underpass(x1, y1, x2, y2, bottom, lane, skipping.length),
			end: [x2, y2],
		};
	});
	return {
		columns,
		spots,
		lines,
		width:
			columns.length * board.width + (columns.length - 1) * board.columnGap,
		height:
			bottom +
			(skipping.length === 0 ? 0 : board.under + skipping.length * board.lane),
	};
}

// columnsOf puts every topic in the column of its layer: the foundations in
// the catalog's order, and each later column by the mean row of its topics'
// bases, ties kept in the catalog's order.
function columnsOf(topics: Topics): Topic[][] {
	const depth = Math.max(...topics.all.map((topic) => topic.layer));
	const rows = new Map<string, number>();
	const columns: Topic[][] = [];
	for (let layer = 0; layer <= depth; layer++) {
		const mean = (topic: Topic) =>
			topic.bases.reduce((sum, base) => sum + (rows.get(base) ?? 0), 0) /
			Math.max(1, topic.bases.length);
		const column = topics.all
			.filter((topic) => topic.layer === layer)
			.map((topic, order) => ({ topic, order, at: mean(topic) }))
			.sort((a, b) => a.at - b.at || a.order - b.order)
			.map(({ topic }) => topic);
		for (const [row, topic] of column.entries()) {
			rows.set(topic.id, row);
		}
		columns.push(column);
	}
	return columns;
}

// rightOf and leftOf are the middles of a box's two sides, where lines meet it.
function rightOf({ x, y }: Spot): [number, number] {
	return [x + board.width, y + board.height / 2];
}

function leftOf({ x, y }: Spot): [number, number] {
	return [x, y + board.height / 2];
}

// curve joins two boxes in neighbouring columns.
function curve(x1: number, y1: number, x2: number, y2: number): string {
	const bend = (x2 - x1) / 2;
	return `M${x1} ${y1}C${x1 + bend} ${y1} ${x2 - bend} ${y2} ${x2} ${y2}`;
}

// underpass leads a line past the column between its two boxes: down the
// gutter after the first, along a lane of its own under every box, and up the
// gutter before the second, its corners rounded.
function underpass(
	x1: number,
	y1: number,
	x2: number,
	y2: number,
	bottom: number,
	lane: number,
	lanes: number,
): string {
	const shift = (lane - (lanes - 1) / 2) * board.lane;
	const down = x1 + board.columnGap / 2 + shift;
	const up = x2 - board.columnGap / 2 + shift;
	const along = bottom + board.under + lane * board.lane;
	const r = board.corner;
	return [
		`M${x1} ${y1}`,
		`H${down - r}`,
		`Q${down} ${y1} ${down} ${y1 + r}`,
		`V${along - r}`,
		`Q${down} ${along} ${down + r} ${along}`,
		`H${up - r}`,
		`Q${up} ${along} ${up} ${along - r}`,
		`V${y2 + r}`,
		`Q${up} ${y2} ${up + r} ${y2}`,
		`H${x2}`,
	].join("");
}

/** nodeId, lineId and sayingId are the ids the map gives a topic's parts. */
export const nodeId = (topic: Topic) => `map-${topic.slug}`;
export const lineId = (line: Line) => `line-${line.from.slug}-${line.to.slug}`;
export const sayingId = (topic: Topic) => `say-${topic.slug}`;

/**
 * litBy are the lines a topic pointed at lights: the links among the topics
 * below it and into it, and among the topics above it and out of it.
 */
export function litBy(topic: Topic, lines: readonly Line[]): Line[] {
	const below = new Set([...topic.before, topic.id]);
	const above = new Set([...topic.after, topic.id]);
	return lines.filter(
		({ from, to }) =>
			(topic.before.includes(from.id) && below.has(to.id)) ||
			(above.has(from.id) && topic.after.includes(to.id)),
	);
}

/**
 * lightRules are the stylesheet that lights the map while a topic is pointed
 * at or reached with the keyboard: the topics below it as the first to learn,
 * those above it as what comes after, the lines between them, and its own
 * line of words under the map. A rule names its topics by their ids, so the
 * rules are written from the catalog's links when the site is built, and need
 * no script; a browser that cannot read them shows the map unlit.
 */
export function lightRules(topics: Topics, map: TopicMap): string {
	const pointed = (topic: Topic) =>
		`.s-map:has(#${nodeId(topic)}:is(:hover,:focus-visible))`;
	const byId = (id: string) => topics.byId.get(id);
	const first = topics.all.flatMap((topic) =>
		topic.before.flatMap((id) => {
			const below = byId(id);
			return below ? [`${pointed(topic)} #${nodeId(below)}`] : [];
		}),
	);
	const then = topics.all.flatMap((topic) =>
		topic.after.flatMap((id) => {
			const above = byId(id);
			return above ? [`${pointed(topic)} #${nodeId(above)}`] : [];
		}),
	);
	const lit = topics.all.flatMap((topic) =>
		litBy(topic, map.lines).map((line) => `${pointed(topic)} #${lineId(line)}`),
	);
	const said = topics.all.map(
		(topic) => `${pointed(topic)} #${sayingId(topic)}`,
	);
	return [
		rule(
			first,
			"background:var(--s-accent-tint);border-color:var(--s-accent-tint);color:var(--s-on-tint)",
		),
		rule(
			then,
			"background:var(--s-page);border-color:var(--s-accent);color:var(--s-accent)",
		),
		rule(lit, "color:var(--s-accent);stroke-width:2.5"),
		rule(said, "display:block"),
	]
		.filter((written) => written !== "")
		.join("\n");
}

// rule is one rule of selectors, or nothing when there are none.
function rule(selectors: readonly string[], declarations: string): string {
	return selectors.length === 0
		? ""
		: `${selectors.join(",\n")}{${declarations}}`;
}
