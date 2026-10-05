import { useMemo } from "preact/hooks";
import { Verdict } from "../design/blocks";
import { MessageHeader } from "../design/thread";
import type { ProgressOver } from "./CardFrame";
import { FirstRunScreen } from "./FirstRunScreen";
import { ProgressScreen } from "./ProgressScreen";
import { readScreen } from "./payload";
import { useWords } from "./words";

/**
 * ProgressOverCard is the progress shown in the card over what it showed: the
 * screen the reply names — the progress, its sections open as folds says, or
 * the first sign-in when the profile has gone since — and, until the reply is
 * in or when it does not read, whose progress it is and why none is shown. A
 * change saved on its form is handed to onSaved.
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
