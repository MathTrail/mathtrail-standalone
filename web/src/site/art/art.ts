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
 * CellWords are where a cell of the key idea's table keeps its own words
 * beside its drawing: under it, beside it, or for a screen reader alone.
 */
export type CellWords = "under" | "beside" | "hidden";

/**
 * IdeaColumn are the drawings of one column of the key idea's table, one in
 * each row, their words in one place, and where the cells keep their own
 * words: for a screen reader alone where the column does not say.
 */
export type IdeaColumn = {
	readonly column: number;
	readonly rows: readonly Art[];
	readonly words?: CellWords;
};

/**
 * TopicArt are the drawings of a topic's page, by where each stands: the
 * first screen's, with the line over it where it has one; one on each card
 * of the ground the topic stands on; those of the columns of the key idea's
 * table, the column its answers stand in where that is not the last, and
 * the key over it; the key under the steps of the move; the drawings of
 * each worked example; and one for each trap the page explains, by the
 * trap's name. A place with no drawing keeps the page's own look.
 */
export type TopicArt = {
	readonly hero: Art;
	readonly heroLine?: Art;
	readonly basis?: readonly Art[];
	readonly idea?: readonly IdeaColumn[];
	readonly ideaAnswers?: number;
	readonly ideaLegend?: Art;
	readonly legend?: Art;
	readonly examples?: readonly ExampleArt[];
	readonly traps?: Readonly<Record<string, Art>>;
};
