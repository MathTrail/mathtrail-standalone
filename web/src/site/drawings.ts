import { z } from "zod";
import type { CatalogTopic } from "./topics";

/**
 * thumbs are the small coloured drawings the cards of the page of the topics
 * show, one a topic, each the sign of what its tasks are about: people lined
 * up by height, a knight and a liar, two circles that overlap, pairs of
 * digits, a fence of posts, hutches with rabbits in them, cells of a grid, a
 * clock's face, two weeks with a day marked in each, a product that makes a
 * hundred, dots that alternate, groups of dots and the one left over, three
 * quarters, one tenth, two parts to three, matches, and a balance.
 */
export const thumbs = [
	"lined-up",
	"knight-and-liar",
	"two-circles",
	"digit-pairs",
	"fence",
	"hutches",
	"grid-cells",
	"clock-face",
	"two-weeks",
	"hundred-product",
	"alternating-dots",
	"groups-and-remainder",
	"three-quarters",
	"one-in-ten",
	"two-to-three",
	"matches",
	"balance",
] as const;

/** Thumb is the name of a card's small drawing. */
export type Thumb = (typeof thumbs)[number];

/**
 * Drawing is what a topic's card draws: one of the cards' small drawings, by
 * its name. A topic's page draws drawings of its own, under art/.
 */
export type Drawing = { readonly markup: Thumb };

/**
 * TopicDrawings are the drawings of a topic the site's data gives: its card's
 * on the page of the topics.
 */
export type TopicDrawings = {
	readonly card: Drawing;
};

/**
 * drawingFile is a drawing as the site's data writes it: the name of a card's
 * small drawing, and nothing beside it.
 */
export const drawingFile = z.strictObject({ markup: z.enum(thumbs) });

/** topicDrawingsFile is a topic's drawings as the site's data writes them. */
export const topicDrawingsFile = z.strictObject({
	card: drawingFile,
});

/**
 * readDrawings reads the drawings of the topics: each of a topic of the
 * catalog.
 */
export function readDrawings(
	catalog: readonly CatalogTopic[],
	drawings: Readonly<Record<string, z.infer<typeof topicDrawingsFile>>>,
): ReadonlyMap<string, TopicDrawings> {
	const known = new Set(catalog.map((topic) => topic.id));
	return new Map(
		Object.entries(drawings).map(([id, { card }]) => {
			if (!known.has(id)) {
				throw new Error(
					`the drawings of ${id} are of a topic the catalog does not have`,
				);
			}
			return [id, { card }];
		}),
	);
}
