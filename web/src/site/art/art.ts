import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";

/**
 * ArtProps are what a drawing of a topic's page is drawn from: the page, and
 * the place of the drawing's words among the page's words.
 */
export type ArtProps = { readonly page: PageReader; readonly at: string };

/** Art is one drawing of a topic's page. */
export type Art = (props: ArtProps) => ComponentChildren;

/**
 * ExampleArt are the drawings of a worked example: the one under its task,
 * the one beside each step, counted from 0, which a step with none leaves
 * out, and the one beside its note. Their words share one place, the
 * example's own.
 */
export type ExampleArt = {
	readonly task?: Art;
	readonly steps?: readonly (Art | undefined)[];
	readonly note?: Art;
};

/**
 * IdeaColumn are the drawings of one column of the key idea's table, one in
 * each row, their words in one place: in the first column before the row's
 * name, and in any other with the cell's own words kept under the drawing or
 * for a screen reader alone.
 */
export type IdeaColumn = {
	readonly column: number;
	readonly rows: readonly Art[];
	readonly showWords?: boolean;
};

/**
 * TopicArt are the drawings of a topic's page, by where each stands: the
 * first screen's, with the line over it where it has one; one on each card
 * of the ground the topic stands on; those of the columns of the key idea's
 * table; the key under the steps of the move; the drawings of each worked
 * example; and one for each trap the page explains, by the trap's name. A
 * place with no drawing keeps the page's own look.
 */
export type TopicArt = {
	readonly hero: Art;
	readonly heroLine?: Art;
	readonly basis?: readonly Art[];
	readonly idea?: readonly IdeaColumn[];
	readonly legend?: Art;
	readonly examples?: readonly ExampleArt[];
	readonly traps?: Readonly<Record<string, Art>>;
};
