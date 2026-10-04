import type { ComponentChildren } from "preact";
import { useEffect, useMemo, useReducer, useRef, useState } from "preact/hooks";
import { Verdict } from "../design/blocks";
import { MessageHeader, ThreadBar } from "../design/thread";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import { FirstRunScreen } from "./FirstRunScreen";
import { type Folds, useFolds } from "./folds";
import { ProgressScreen } from "./ProgressScreen";
import { type Child, readScreen } from "./payload";
import { type Progress, progressAfter } from "./progress";
import { useWords } from "./words";

/**
 * CardFrame is what every card of a lesson has around what it shows: the width
 * it lays out for, handed to what it frames, and — on a card that knows whose
 * it is — the line at its top, which opens the child's progress over the card
 * and leads back to it as it was left, by the way back named in back. The
 * progress is read afresh at each opening, and its sections open as they were
 * left, for as long as the card is drawn. A change the parent saves on the
 * progress's form is the card's from then on — its line and what it frames
 * show the child as the profile now says — until the card is handed another
 * payload, which carries the child as the service has it.
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
	children: (wide: boolean, child: Child | undefined) => ComponentChildren;
}) {
	const words = useWords();
	const [saved, setSaved] = useState<{
		over: Child | undefined;
		child: Child;
	}>();
	const whose =
		saved !== undefined && saved.over === child ? saved.child : child;
	const [progress, dispatch] = useReducer(progressAfter, undefined);
	const folds = useFolds();
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
						{whose !== undefined && (
							<ThreadBar
								name={whose.pseudonym}
								action={words.text("task.profile_action")}
								onClick={openProgress}
								buttonRef={topBar}
							/>
						)}
						{children(wide, whose)}
					</div>
					{progress !== undefined && whose !== undefined && (
						<>
							<ThreadBar
								variant="back"
								label={back}
								onClick={() => dispatch({ type: "closed" })}
								buttonRef={backBar}
							/>
							<ProgressOverCard
								progress={progress}
								child={whose}
								wide={wide}
								host={host}
								folds={folds}
								onSaved={(changed) => setSaved({ over: child, child: changed })}
							/>
						</>
					)}
				</>
			)}
		</CardRoot>
	);
}

// ProgressOverCard is the progress shown in the card over what it showed: the
// screen the reply names — the progress, its sections open as folds says, or
// the first sign-in when the profile has gone since — and, until the reply is
// in or when it does not read, whose progress it is and why none is shown. A
// change saved on its form is handed to onSaved.
function ProgressOverCard({
	progress,
	child,
	wide,
	host,
	folds,
	onSaved,
}: {
	progress: Progress;
	child: Child;
	wide: boolean;
	host: Host;
	folds: Folds;
	onSaved: (child: Child) => void;
}) {
	const words = useWords();
	const shown = useMemo(
		() =>
			progress.state === "read" ? readScreen(progress.payload) : undefined,
		[progress],
	);
	if (shown?.screen === "progress") {
		return (
			<ProgressScreen
				report={shown.report}
				wide={wide}
				host={host}
				folds={folds}
				onSaved={onSaved}
			/>
		);
	}
	if (shown?.screen === "first_run") {
		return <FirstRunScreen firstRun={shown.firstRun} wide={wide} />;
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
