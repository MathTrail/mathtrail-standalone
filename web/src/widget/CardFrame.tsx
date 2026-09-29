import type { ComponentChildren, Ref } from "preact";
import { useEffect, useReducer, useRef } from "preact/hooks";
import { Verdict } from "../design/blocks";
import { classes } from "../design/classes";
import { MessageHeader, ThreadBar } from "../design/thread";
import { useWide } from "../design/wide";
import type { Host } from "./bridge";
import type { Child } from "./payload";
import { type Progress, progressAfter } from "./progress";
import { StubCard } from "./StubCard";
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
	const root = useRef<HTMLDivElement>(null);
	const wide = useWide(root);
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
		<div
			ref={root}
			class={classes(
				"mt",
				"mt-widget",
				wide && "mt-wide",
				words.dir === "rtl" && "mt-rtl",
			)}
		>
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
				<ProgressOverCard
					progress={progress}
					child={child}
					wide={wide}
					back={back}
					backBar={backBar}
					onBack={() => dispatch({ type: "closed" })}
				/>
			)}
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

// ProgressOverCard is the progress shown in the card over what it showed: the
// way back, whose progress it is, and what reading it gave — shown as it
// arrived until the progress has a screen of its own.
function ProgressOverCard({
	progress,
	child,
	wide,
	back,
	backBar,
	onBack,
}: {
	progress: Progress;
	child: Child;
	wide: boolean;
	back: string;
	backBar: Ref<HTMLButtonElement>;
	onBack: () => void;
}) {
	const words = useWords();
	return (
		<>
			<ThreadBar
				variant="back"
				label={back}
				onClick={onBack}
				buttonRef={backBar}
			/>
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
					{progress.state === "read" && <StubCard payload={progress.payload} />}
					{progress.state === "failed" && (
						<Verdict>{words.text("progress.failed")}</Verdict>
					)}
				</div>
			</article>
		</>
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
