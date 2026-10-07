/**
 * The worked task of the page "Research", which its text chooses rather than
 * reports: a fence of length metres with a post every gap metres, its ends
 * included, and the five options a card offers for it, the right one at
 * answer. The first screen draws a task's way with it, and thesis 02 the
 * solver's two runs over its options.
 */
export const fence = { length: 12, gap: 3 } as const;

/** example are the values of the worked task's options, in letter order. */
export const example = [3, 4, 5, 6, 12] as const;

/** answer is the place of the worked task's right option. */
export const answer = 2;

/** fencePosts is how many posts the worked task's fence has. */
export const fencePosts = fence.length / fence.gap + 1;
