import {
	type ComponentChildren,
	type ComponentType,
	createContext,
} from "preact";
import {
	useContext,
	useEffect,
	useReducer,
	useRef,
	useState,
} from "preact/hooks";
import { NameBar, ThreadBar } from "../design/thread";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import { type Folds, useFolds } from "./folds";
import { type PeriodChoice, usePeriod } from "./moves";
import type { Child } from "./payload";
import { type Progress, progressAfter } from "./progress";
import { useService } from "./service";
import { useWords } from "./words";

/**
 * ProgressOver is what the progress drawn over a card is given: the reply to
 * the card's reading of the progress as it stands, whose card it is, the width
 * the card lays out for, its host, the sections open and the while chosen as
 * the card keeps them, and where a change saved on the progress's form goes.
 */
export type ProgressOver = {
	progress: Progress;
	child: Child;
	wide: boolean;
	host: Host;
	folds: Folds;
	period: PeriodChoice;
	onSaved: (child: Child) => void;
};

/**
 * OpensProgress is what draws the child's progress over a card, given by the
 * place the card is drawn in. The widget's page gives it; a page that shows a
 * card has no profile to open, and gives none.
 */
export const OpensProgress = createContext<
	ComponentType<ProgressOver> | undefined
>(undefined);

/**
 * CardFrame is what every card of a lesson has around what it shows: the width
 * it lays out for, handed to what it frames, and — on a card that knows whose
 * it is — the line at its top, which says whose card it is and, where the place
 * the card is drawn in opens the progress, opens the child's progress over the
 * card and leads back to it as it was left, by the way back named in back. The
 * progress is read afresh at each opening, and its sections open as they were
 * left, for as long as the card is drawn. A change the parent saves on the
 * progress's form is the card's from then on — its line and what it frames
 * show the child as the profile now says — until the card is handed another
 * payload, which carries the child as the service has it. A card handed the
 * child hands what it frames a child too.
 */
export function CardFrame<C extends Child | undefined>({
	child,
	back,
	host,
	children,
}: {
	child: C;
	back: string;
	host: Host;
	children: (wide: boolean, child: C | Child) => ComponentChildren;
}) {
	const words = useWords();
	const ProgressView = useContext(OpensProgress);
	const service = useService();
	const [saved, setSaved] = useState<{
		over: Child | undefined;
		child: Child;
	}>();
	const whose =
		saved !== undefined && saved.over === child ? saved.child : child;
	const [progress, dispatch] = useReducer(progressAfter, undefined);
	const folds = useFolds();
	const period = usePeriod();
	const topBar = useRef<HTMLButtonElement>(null);
	const backBar = useRef<HTMLButtonElement>(null);
	const openings = useRef(0);

	async function openProgress() {
		openings.current += 1;
		const opening = openings.current;
		dispatch({ type: "opened", opening });
		const read = await service.readProgress();
		dispatch(
			read.kind === "read"
				? { type: "read", opening, payload: read.payload }
				: { type: "failed", opening },
		);
	}

	useFocusFollowsProgress(progress !== undefined, topBar, backBar);

	return (
		<CardRoot>
			{(wide) => (
				<>
					<div hidden={progress !== undefined}>
						{whose !== undefined &&
							(ProgressView === undefined ? (
								<NameBar name={whose.pseudonym} />
							) : (
								<ThreadBar
									name={whose.pseudonym}
									action={words.text("task.profile_action")}
									onClick={openProgress}
									buttonRef={topBar}
								/>
							))}
						{children(wide, whose)}
					</div>
					{progress !== undefined &&
						whose !== undefined &&
						ProgressView !== undefined && (
							<>
								<ThreadBar
									variant="back"
									label={back}
									onClick={() => dispatch({ type: "closed" })}
									buttonRef={backBar}
								/>
								<ProgressView
									progress={progress}
									child={whose}
									wide={wide}
									host={host}
									folds={folds}
									period={period}
									onSaved={(changed) =>
										setSaved({ over: child, child: changed })
									}
								/>
							</>
						)}
				</>
			)}
		</CardRoot>
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
