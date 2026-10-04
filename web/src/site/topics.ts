/**
 * CatalogTopic is a topic as the catalog the service is built with writes it:
 * the fields the site reads of it.
 */
export type CatalogTopic = {
	readonly id: string;
	readonly slug: string;
	readonly grade_levels: readonly string[];
	readonly builds_on: readonly string[];
	readonly site_page: boolean;
};

/**
 * Topic is a topic of the catalog as the site shows it: its address, its
 * grades, its place on the map, and the topics it stands on and that stand on
 * it — directly, and all the way down and up.
 */
export type Topic = {
	readonly id: string;
	readonly slug: string;
	/** sitePage reports whether the topic's own page is published. */
	readonly sitePage: boolean;
	/** grades are the first and the last grade the topic is taught at. */
	readonly grades: readonly [number, number];
	/** layer is the longest chain of bases below it: 0 for a foundation. */
	readonly layer: number;
	/** bases are the topics it builds on, in the catalog's order. */
	readonly bases: readonly string[];
	/** opens are the topics that build on it, in the catalog's order. */
	readonly opens: readonly string[];
	/** before are every topic below it, through its bases and theirs. */
	readonly before: readonly string[];
	/** after are every topic above it, through what it opens and onwards. */
	readonly after: readonly string[];
};

/** Group is a set of topics the site shows together, by the topics' ids. */
export type Group = { readonly id: string; readonly topics: readonly string[] };

/**
 * Topics are the catalog's topics as the site shows them, in the catalog's
 * order, and the groups they are shown in.
 */
export type Topics = {
	readonly all: readonly Topic[];
	readonly byId: ReadonlyMap<string, Topic>;
	readonly groups: readonly Group[];
};

/**
 * anchorName matches what a part of a page that an address reaches may be
 * called, as a group of topics or a technique is: lowercase words joined by
 * dashes, which no address has to escape.
 */
export const anchorName = /^[a-z]+(?:-[a-z]+)*$/;

/**
 * readTopics reads the catalog's topics and the groups the site's data puts
 * them in. A group is refused that names a topic the catalog lacks, holds
 * none, shares its name with another or with a topic's slug — both are
 * anchors of one page — and so is a topic in no group or in two. The catalog
 * itself is the service's, which refuses a broken one before the site reads
 * it: an unknown base, a circle.
 */
export function readTopics(
	catalog: readonly CatalogTopic[],
	groups: readonly Group[],
): Topics {
	checkGroups(catalog, groups);
	const byId = new Map(catalog.map((topic) => [topic.id, topic]));
	const order = new Map(catalog.map((topic, at) => [topic.id, at]));
	const inOrder = (ids: Iterable<string>) =>
		[...ids].sort((a, b) => (order.get(a) ?? 0) - (order.get(b) ?? 0));
	const opens = new Map(
		catalog.map((topic) => [
			topic.id,
			catalog.filter((other) => other.builds_on.includes(topic.id)),
		]),
	);
	const layers = new Map<string, number>();
	const layerOf = (id: string): number => {
		let layer = layers.get(id);
		if (layer === undefined) {
			const bases = byId.get(id)?.builds_on ?? [];
			layer = bases.length === 0 ? 0 : 1 + Math.max(...bases.map(layerOf));
			layers.set(id, layer);
		}
		return layer;
	};
	const reached = (id: string, next: (id: string) => readonly string[]) => {
		const seen = new Set<string>();
		const walk = (from: string) => {
			for (const to of next(from)) {
				if (!seen.has(to)) {
					seen.add(to);
					walk(to);
				}
			}
		};
		walk(id);
		return inOrder(seen);
	};
	const basesOf = (id: string) => byId.get(id)?.builds_on ?? [];
	const opensOf = (id: string) =>
		(opens.get(id) ?? []).map((topic) => topic.id);
	const all = catalog.map(
		(topic): Topic => ({
			id: topic.id,
			slug: topic.slug,
			sitePage: topic.site_page,
			grades: gradesOf(topic.grade_levels),
			layer: layerOf(topic.id),
			bases: inOrder(topic.builds_on),
			opens: opensOf(topic.id),
			before: reached(topic.id, basesOf),
			after: reached(topic.id, opensOf),
		}),
	);
	return {
		all,
		byId: new Map(all.map((topic) => [topic.id, topic])),
		groups,
	};
}

/**
 * gradesOf is the first and the last grade of levels, each written as a span
 * of grades, as the catalog writes the levels a topic is taught at: 3-4.
 */
export function gradesOf(levels: readonly string[]): [number, number] {
	const grades = levels.flatMap((level) => level.split("-").map(Number));
	return [Math.min(...grades), Math.max(...grades)];
}

// checkGroups refuses groups that leave a topic out, name one twice or name
// one the catalog lacks, and a group whose name another anchor of the page
// has.
function checkGroups(
	catalog: readonly CatalogTopic[],
	groups: readonly Group[],
): void {
	const known = new Set(catalog.map((topic) => topic.id));
	const anchors = new Set(catalog.map((topic) => topic.slug));
	const groupOf = new Map<string, string>();
	for (const { id, topics } of groups) {
		checkGroupName(id, anchors);
		anchors.add(id);
		if (topics.length === 0) {
			throw new Error(`the group ${id} holds no topic`);
		}
		for (const topic of topics) {
			if (!known.has(topic)) {
				throw new Error(
					`the group ${id} holds ${topic}, which the catalog does not have`,
				);
			}
			const other = groupOf.get(topic);
			if (other !== undefined) {
				throw new Error(
					`the topic ${topic} is in the group ${other} and in the group ${id}`,
				);
			}
			groupOf.set(topic, id);
		}
	}
	const left = catalog.find((topic) => !groupOf.has(topic.id));
	if (left !== undefined) {
		throw new Error(`the topic ${left.id} is in no group`);
	}
}

// checkGroupName refuses a group's name that is no anchor, and one that an
// anchor of the page of the topics already has: a topic's or another group's.
function checkGroupName(id: string, anchors: ReadonlySet<string>): void {
	if (!anchorName.test(id)) {
		throw new Error(
			`the group ${JSON.stringify(id)} is no anchor: a group is called by lowercase words joined by dashes`,
		);
	}
	if (anchors.has(id)) {
		throw new Error(
			`the group ${id} shares its name with another anchor of the page of the topics`,
		);
	}
}
