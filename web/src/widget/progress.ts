/**
 * Progress is the progress shown in a card over what the card shows, as it is
 * read: being read, read, or not read. Each opening is numbered, so that the
 * reply to one closed since is not taken for the reply to the one open now.
 */
export type Progress =
	| { state: "reading"; opening: number }
	| { state: "read"; opening: number; payload: unknown }
	| { state: "failed"; opening: number };

/** ProgressEvent is something that happens to the progress over a card. */
export type ProgressEvent =
	| { type: "opened"; opening: number }
	| { type: "read"; opening: number; payload: unknown }
	| { type: "failed"; opening: number }
	| { type: "closed" };

/**
 * progressAfter is the progress after event: opened when it is closed, read
 * or failed by the reply to the opening being read — a reply to an opening
 * closed since changes nothing — and closed by the way back. An event that
 * changes nothing leaves the very progress it was given.
 */
export function progressAfter(
	progress: Progress | undefined,
	event: ProgressEvent,
): Progress | undefined {
	if (event.type === "closed") {
		return undefined;
	}
	if (event.type === "opened") {
		return progress ?? { state: "reading", opening: event.opening };
	}
	if (progress?.state !== "reading" || progress.opening !== event.opening) {
		return progress;
	}
	return event.type === "read"
		? { state: "read", opening: event.opening, payload: event.payload }
		: { state: "failed", opening: event.opening };
}
