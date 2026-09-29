import type { ComponentChildren } from "preact";
import { useRef } from "preact/hooks";
import { classes } from "../design/classes";
import { MessageHeader } from "../design/thread";
import { useWide } from "../design/wide";
import { useWords } from "./words";

/**
 * CardRoot is the card itself: the width it lays out for, which it measures
 * and hands to what it holds, and the way the card's words run.
 */
export function CardRoot({
	children,
}: {
	children: (wide: boolean) => ComponentChildren;
}) {
	const words = useWords();
	const root = useRef<HTMLDivElement>(null);
	const wide = useWide(root);
	return (
		<div
			ref={root}
			class={classes(
				"mt",
				"mt-widget",
				wide && "mt-wide",
				words.dir === "rtl" && "mt-rtl",
			)}
		>
			{children(wide)}
		</div>
	);
}

/**
 * CardHeader heads what a card shows with MathTrail's name and, when the card
 * knows whose it is, the grade the child is coached at.
 */
export function CardHeader({
	grade,
	wide,
}: {
	grade: number | undefined;
	wide: boolean;
}) {
	const words = useWords();
	return (
		<MessageHeader
			author="app"
			name={words.text("app.name")}
			badge={
				grade === undefined ? undefined : words.text("app.badge", { grade })
			}
			wide={wide}
		/>
	);
}
