import { z } from "zod";
import type { Picture } from "../design/picture/model";
import { pictureFormat } from "../widget/picture";
import type { CatalogTopic } from "./topics";

/**
 * markups are the drawings of the site's own a topic's drawing may be, where
 * no kind of picture shows what the page means: islanders and what they say,
 * a product and a sum regrouped, a tree of numbers, a time stepped through the
 * hour, a round table, flags of two stripes, a strip cut in two, the corners
 * of a cube, a count of eggs, a price changed twice, and a daisy.
 */
export const markups = [
	"islanders",
	"product-regrouped",
	"number-tree",
	"through-the-hour",
	"sum-regrouped",
	"round-table",
	"two-stripe-flags",
	"strip-cuts",
	"cube-corners",
	"eggs-backwards",
	"price-changes",
	"daisy",
] as const;

/** Markup is the name of a drawing of the site's own. */
export type Markup = (typeof markups)[number];

/**
 * Drawing is what a topic's card or page draws: a picture of one of the kinds
 * a task's picture has, or a drawing of the site's own, by its name.
 */
export type Drawing =
	| { readonly picture: Picture }
	| { readonly markup: Markup };

/**
 * TopicDrawings are the drawings of a topic: its card's on the page of the
 * topics, and its first screen's once its page is published.
 */
export type TopicDrawings = {
	readonly card: Drawing;
	readonly hero?: Drawing;
};

/**
 * drawingFile is a drawing as the site's data writes it: a picture the card's
 * reader accepts, or the name of a drawing of the site's own — one of the two,
 * and nothing beside it. It is one object with two members rather than a union
 * of two, which zod would refuse as an invalid input with no member named: a
 * name it does not know, or a picture short of a member, is named where it
 * stands.
 */
export const drawingFile = z
	.strictObject({
		picture: pictureFormat.optional(),
		markup: z.enum(markups).optional(),
	})
	.transform((drawing, problems): Drawing => {
		if (drawing.picture !== undefined && drawing.markup === undefined) {
			return { picture: drawing.picture };
		}
		if (drawing.markup !== undefined && drawing.picture === undefined) {
			return { markup: drawing.markup };
		}
		problems.addIssue({
			code: "custom",
			message:
				"a drawing is a picture or a drawing of the site's own, one of the two",
		});
		return z.NEVER;
	});

/** topicDrawingsFile is a topic's drawings as the site's data writes them. */
export const topicDrawingsFile = z.strictObject({
	card: drawingFile,
	hero: drawingFile.optional(),
});

/**
 * readDrawings reads the drawings of the topics: each of a topic of the
 * catalog, and a first screen's only for a topic whose page is published.
 */
export function readDrawings(
	catalog: readonly CatalogTopic[],
	drawings: Readonly<Record<string, z.infer<typeof topicDrawingsFile>>>,
): ReadonlyMap<string, TopicDrawings> {
	const byId = new Map(catalog.map((topic) => [topic.id, topic]));
	return new Map(
		Object.entries(drawings).map(([id, { card, hero }]) => {
			const topic = byId.get(id);
			if (topic === undefined) {
				throw new Error(
					`the drawings of ${id} are of a topic the catalog does not have`,
				);
			}
			if (hero === undefined) {
				return [id, { card }];
			}
			if (!topic.site_page) {
				throw new Error(
					`the first screen of ${id} has a drawing, and the catalog does not publish its page`,
				);
			}
			return [id, { card, hero }];
		}),
	);
}
