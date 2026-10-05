/** letters are the options' letters, in the order the card shows them. */
export const letters = ["A", "B", "C", "D", "E"] as const;

/** Letter names one of a task's five options. */
export type Letter = (typeof letters)[number];

/** dontKnow is the answer "I don't know": a wrong answer that chose no option. */
export const dontKnow = "?";

/** Choice is an answer the child can give: an option, or "I don't know". */
export type Choice = Letter | typeof dontKnow;

/** choices are every answer the child can give, the options in their order. */
export const choices: readonly Choice[] = [...letters, dontKnow];
