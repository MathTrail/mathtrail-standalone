/** CatalogTrap is a trap of the service's catalog: the field the site reads. */
export type CatalogTrap = { readonly id: string };

/**
 * ReferenceTask is a reference task of the service's content as the site reads
 * it: its id, its topic, and the trap each of its wrong options names.
 */
export type ReferenceTask = {
	readonly id: string;
	readonly topic: string;
	readonly distractors: Readonly<Record<string, { readonly trap: string }>>;
};

/**
 * rankTraps is every topic's traps, the most frequent among its reference
 * tasks first. A trap counts once for every wrong option that names it, at
 * every level the topic is taught at, and two traps named as often stand in
 * the catalog's order: the count the service makes of a topic's usual
 * mistakes, taken over the whole topic, since a topic's page is about all of
 * it. A trap the catalog does not have stops the build, since a card could not
 * name it.
 */
export function rankTraps(
	catalog: readonly CatalogTrap[],
	tasks: readonly ReferenceTask[],
): ReadonlyMap<string, readonly string[]> {
	const order = new Map(catalog.map((trap, at) => [trap.id, at]));
	const counts = new Map<string, Map<string, number>>();
	for (const task of tasks) {
		const seen = counts.get(task.topic) ?? new Map<string, number>();
		counts.set(task.topic, seen);
		for (const { trap } of Object.values(task.distractors)) {
			if (!order.has(trap)) {
				throw new Error(
					`the reference task ${task.id} names the trap ${trap}, which the catalog does not have`,
				);
			}
			seen.set(trap, (seen.get(trap) ?? 0) + 1);
		}
	}
	return new Map(
		[...counts].map(([topic, seen]) => [
			topic,
			[...seen.keys()].sort(
				(a, b) =>
					(seen.get(b) ?? 0) - (seen.get(a) ?? 0) ||
					(order.get(a) ?? 0) - (order.get(b) ?? 0),
			),
		]),
	);
}
