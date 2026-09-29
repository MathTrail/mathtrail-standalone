import type { ComponentChildren } from "preact";
import { useEffect, useMemo, useReducer, useRef } from "preact/hooks";
import { Verdict } from "../design/blocks";
import { MessageHeader, ThreadBar } from "../design/thread";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import { FirstRunScreen } from "./FirstRunScreen";
import { ProgressScreen } from "./ProgressScreen";
import { type Child, readScreen } from "./payload";
import { type Progress, progressAfter } from "./progress";
import { useWords } from "./words";

/**
 * CardFrame is what every card of a lesson has around what it shows: the width
 * it lays out for, handed to what it frames, and — on a card that knows whose
 * it is — the line at its top, which opens the child's progress over the card
 * and leads back to it as it was left, by the way back named in back.
 */
export function CardFrame({
	child,
	back,
	host,
	children,
}: {
	child: Child | undefined;
	back: string;
	host: Host;
	children: (wide: boolean) => ComponentChildren;
}) {
	const words = useWords();
	const [progress, dispatch] = useReducer(progressAfter, undefined);
	const topBar = useRef<HTMLButtonElement>(null);
	const backBar = useRef<HTMLButtonElement>(null);
	const openings = useRef(0);

	async function openProgress() {
		openings.current += 1;
		const opening = openings.current;
		dispatch({ type: "opened", opening });
		try {
			const read = await host.callTool("read_progress", {});
			dispatch(
				read.isError === true
					? { type: "failed", opening }
					: { type: "read", opening, payload: read.structuredContent },
			);
		} catch (error: unknown) {
			console.error("widget: the progress did not arrive", error);
			dispatch({ type: "failed", opening });
		}
	}

	useFocusFollowsProgress(progress !== undefined, topBar, backBar);

	return (
		<CardRoot>
			{(wide) => (
				<>
					<div hidden={progress !== undefined}>
						{child !== undefined && (
							<ThreadBar
								name={child.pseudonym}
								action={words.text("task.profile_action")}
								onClick={openProgress}
								buttonRef={topBar}
							/>
						)}
						{children(wide)}
					</div>
					{progress !== undefined && child !== undefined && (
						<>
							<ThreadBar
								variant="back"
								label={back}
								onClick={() => dispatch({ type: "closed" })}
								buttonRef={backBar}
							/>
							<ProgressOverCard
								progress={progress}
								child={child}
								wide={wide}
								host={host}
							/>
						</>
					)}
				</>
			)}
		</CardRoot>
	);
}

// ProgressOverCard is the progress shown in the card over what it showed: the
// screen the reply names — the progress, or the first sign-in when the
// profile has gone since — and, until the reply is in or when it does not
// read, whose progress it is and why none is shown.
function ProgressOverCard({
	progress,
	child,
	wide,
	host,
}: {
	progress: Progress;
	child: Child;
	wide: boolean;
	host: Host;
}) {
	const words = useWords();
	const shown = useMemo(
		() =>
			progress.state === "read" ? readScreen(progress.payload) : undefined,
		[progress],
	);
	if (shown?.screen === "progress") {
		return <ProgressScreen report={shown.report} wide={wide} host={host} />;
	}
	if (shown?.screen === "first_run") {
		return <FirstRunScreen firstRun={shown.firstRun} wide={wide} host={host} />;
	}
	return (
		<article
			aria-label={words.text("progress.label")}
			aria-busy={progress.state === "reading"}
		>
			<MessageHeader
				author="person"
				name={child.pseudonym}
				badge={words.text("child.grade", { grade: child.grade })}
				wide={wide}
			/>
			<div class="mt-progress">
				{progress.state !== "reading" && (
					<Verdict>{words.text("progress.failed")}</Verdict>
				)}
			</div>
		</article>
	);
}

// useFocusFollowsProgress moves the focus with the card between what it shows
// and the progress: to the way back as the progress opens, to the way there
// as it closes. The first draw moves nothing.
function useFocusFollowsProgress(
	open: boolean,
	topBar: { current: HTMLButtonElement | null },
	backBar: { current: HTMLButtonElement | null },
): void {
	const wasOpen = useRef(open);
	useEffect(() => {
		if (wasOpen.current === open) {
			return;
		}
		wasOpen.current = open;
		(open ? backBar : topBar).current?.focus({ preventScroll: true });
	}, [open, topBar, backBar]);
}
