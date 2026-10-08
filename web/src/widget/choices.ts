/** letters are the options' letters, in the order the card shows them. */
export const letters = ["A", "B", "C", "D", "E"] as const;

/** Letter names one of a task's five options. */
export type Letter = (typeof letters)[number];

/**
 * dontKnow is the answer "I don't know": a wrong answer that chose no option.
 * A card has no button for it: the adult says it in the chat, and the model
 * records it. A card shows it only as the answer recorded first, told back to
 * an option pressed after it.
 */
export const dontKnow = "?";
