import { useState } from "preact/hooks";
import type { Host } from "./bridge";
import { ComingCard } from "./ComingCard";
import type { Coming, HandedTask } from "./payload";
import { TaskCard } from "./TaskCard";

/** Shown is what a card of a lesson shows: a task, or a task it waits for. */
export type Shown =
	| { kind: "task"; handed: HandedTask }
	| { kind: "coming"; coming: Coming };

/**
 * LessonCard is the card of a lesson, which starts at what a tool's result
 * drew it with — a task, or the wait for one — and moves on in place as the
 * child takes the next task on it: to the task written ahead, at once, or to
 * the wait for the one still being written.
 */
export function LessonCard({ start, host }: { start: Shown; host: Host }) {
	const [shown, setShown] = useState<Shown>(start);
	switch (shown.kind) {
		case "task":
			return (
				<TaskCard
					key={shown.handed.task.id}
					handed={shown.handed}
					host={host}
					onTaken={setShown}
				/>
			);
		case "coming":
			return (
				<ComingCard
					key={shown.coming.requestId}
					coming={shown.coming}
					host={host}
					onTaken={setShown}
				/>
			);
	}
}
