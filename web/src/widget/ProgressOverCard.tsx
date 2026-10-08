import { useMemo } from "preact/hooks";
import { Verdict } from "../design/blocks";
import { LoadingBar, ProgressSkeleton } from "../design/loading";
import { MessageHeader } from "../design/thread";
import type { ProgressOver } from "./CardFrame";
import { FirstRunScreen } from "./FirstRunScreen";
import { ProgressScreen } from "./ProgressScreen";
import { readScreen } from "./payload";
import { useWords } from "./words";

// outlineRuns are the runs of ranks the outline marks under its course while
// the progress is read: the eleven ranks as the service shares them out
// between grades 1–2, 3–4 and 5–6, so that the brackets stand where the
// screen's will. Should the service share them out otherwise, a bracket moves
// as the screen comes, and nothing more.
const outlineRuns = [
	{ first: 1, last: 4 },
	{ first: 5, last: 7 },
	{ first: 8, last: 11 },
];

/**
 * ProgressOverCard is the progress shown in the card over what it showed: the
 * screen the reply names — the progress, its sections open as folds says, or
 * the first sign-in when the profile has gone since. Until the reply is in, it
 * shows whose progress it is over the outline of the screen to come, with a
 * bar running over them; when the reply does not read, whose progress it is
 * and why none is shown. A change saved on its form is handed to onSaved.
 */
export function ProgressOverCard({
	progress,
	child,
	wide,
	host,
	folds,
	period,
	onSaved,
}: ProgressOver) {
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
				period={period}
				onSaved={onSaved}
			/>
		);
	}
	if (shown?.screen === "first_run") {
		return <FirstRunScreen firstRun={shown.firstRun} wide={wide} />;
	}
	const reading = progress.state === "reading";
	return (
		<article aria-label={words.text("progress.label")}>
			{reading && <LoadingBar />}
			<MessageHeader
				author="person"
				name={child.pseudonym}
				badge={words.text("child.grade", { grade: child.grade })}
				wide={wide}
			/>
			{reading ? (
				<ProgressSkeleton
					label={words.text("progress.loading")}
					runs={outlineRuns}
				/>
			) : (
				<div class="mt-progress">
					<Verdict>{words.text("progress.failed")}</Verdict>
				</div>
			)}
		</article>
	);
}
