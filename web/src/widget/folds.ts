import { useReducer } from "preact/hooks";

/**
 * Section is a part of the progress that folds away under its title: the
 * topics, the mistakes that repeat, the latest answers, and the profile.
 */
export type Section = "topics" | "mistakes" | "recent" | "profile";

/** Open is which sections of the progress are open. */
export type Open = ReadonlySet<Section>;

/** allFolded is the progress as it first opens: every section folded. */
export const allFolded: Open = new Set();

/**
 * foldsAfter is which sections are open once section is pressed: opened when
 * it was folded, folded when it was open. Each press is applied to what the
 * press before it left, so that presses that come together all count.
 */
export function foldsAfter(open: Open, section: Section): Open {
	const after = new Set(open);
	if (!after.delete(section)) {
		after.add(section);
	}
	return after;
}

/** Folds is which sections of the progress are open, and how to press one. */
export type Folds = {
	open: Open;
	toggle: (section: Section) => void;
};

/**
 * useFolds is which sections of the progress are open, kept for as long as
 * the component that calls it is drawn: a card keeps them while the progress
 * over it is closed and opened again.
 */
export function useFolds(): Folds {
	const [open, toggle] = useReducer(foldsAfter, allFolded);
	return { open, toggle };
}
