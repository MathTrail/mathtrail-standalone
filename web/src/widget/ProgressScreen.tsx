import { useState } from "preact/hooks";
import { Fold, Note } from "../design/blocks";
import { ViewSwitch } from "../design/controls";
import {
	MoveCounts,
	MoveLegend,
	MoveLine,
	type MoveWay,
	RankList,
	type RankRow,
	RankSummary,
	Segments,
	type SegmentsAt,
	type SegmentTone,
	StatList,
	type StatRow,
	StatusDots,
} from "../design/progress";
import { MessageHeader } from "../design/thread";
import type { Words } from "../i18n/words";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import { type Folds, type Section, useFolds } from "./folds";
import {
	moveOver,
	movesCounted,
	type Period,
	type PeriodChoice,
	periodShown,
	type Reading,
	readingOf,
	usePeriod,
} from "./moves";
import { listed, rankCount, rankName, topicName, trapName } from "./names";
import { ParentData, ParentProfile } from "./ProfileScreen";
import type { Details, Moves, ProgressReport, Recommendation } from "./payload";
import { countText, type Key, percentText, useWords } from "./words";

/** recentShown is how many of the latest entries the progress lists. */
const recentShown = 5;

/**
 * ProgressCard is the card of the child's progress, drawn when the model
 * reads it. No task stands behind it, so it has no way back to one; its
 * sections stay as they were opened, and its moves over the while chosen,
 * for as long as the card is drawn.
 */
export function ProgressCard({
	report,
	host,
}: {
	report: ProgressReport;
	host: Host;
}) {
	const folds = useFolds();
	const period = usePeriod();
	return (
		<CardRoot>
			{(wide) => (
				<ProgressScreen
					report={report}
					wide={wide}
					host={host}
					folds={folds}
					period={period}
				/>
			)}
		</CardRoot>
	);
}

/**
 * ProgressScreen is where the child stands: the rank reached and how far
 * through it — or, while the trial series runs, how far the series has got —
 * and what comes next, over the sections that fold away under their titles:
 * each topic met or within reach with a rank of its own, the mistakes that
 * keep coming back, the latest answers with how many tasks were left without
 * one, and the child's profile for the parent with what the parent can do
 * with the data. A section with nothing in it is not drawn; which sections are
 * open is told by folds, kept by whoever outlives the screen. Once the trial
 * series is over, a switch over the rank chooses the while the moves are drawn
 * over — since the last task, or over the week — as period tells, kept by
 * whoever outlives the screen too: how the rank moved, in stripes on its
 * course and in a line under it, and each topic by its own answers, with how
 * many moved each way by the title of the topics. Where nobody outlives the
 * screen — a card drawn once — it keeps the while chosen itself. A
 * change the parent saves on its form shows at once, the name and the grade at
 * the top among it, and is handed to onSaved.
 */
export function ProgressScreen({
	report,
	wide,
	host,
	folds,
	period,
	onSaved,
}: {
	report: ProgressReport;
	wide: boolean;
	host: Host;
	folds: Folds;
	period?: PeriodChoice;
	onSaved?: (details: Details) => void;
}) {
	const words = useWords();
	const own = usePeriod();
	const choice = period ?? own;
	const { recommendation } = report;
	const moves =
		report.trial === null ? (report.overall?.change ?? undefined) : undefined;
	const shown =
		moves === undefined ? undefined : periodShown(choice.chosen, moves);
	const rows = topicRows(words, report, shown);
	const [profile, setProfile] = useState(report.profile);
	const skipped =
		report.skipped ??
		report.topics.reduce((sum, topic) => sum + topic.skipped, 0);
	const latest = latestOf(words, report.recent);
	const folding = (section: Section) => ({
		open: folds.open.has(section),
		onToggle: () => folds.toggle(section),
	});
	return (
		<article aria-label={words.text("progress.label")}>
			<MessageHeader
				author="person"
				name={profile.pseudonym}
				badge={words.text("child.grade", { grade: profile.grade })}
				wide={wide}
			/>
			{(report.trial !== null ||
				report.overall !== null ||
				recommendation !== null) && (
				<div class="mt-progress">
					{shown !== undefined && (
						<ViewSwitch
							legend={words.text("progress.period")}
							options={[
								{
									value: "last_task",
									label: words.text("progress.period_task"),
								},
								{ value: "week", label: words.text("progress.period_week") },
							]}
							value={shown}
							onChange={choice.choose}
						/>
					)}
					<Standing report={report} moves={moves} shown={shown} />
					{recommendation !== null && (
						<Note label={words.text("progress.next_up")}>
							{nextUp(words, recommendation)}
						</Note>
					)}
				</div>
			)}
			<div class="mt-folds">
				{report.topics.length > 0 && (
					<Fold
						title={words.text("progress.topics")}
						summary={rowsMoved(words, rows)}
						{...folding("topics")}
					>
						<RankList
							note={
								report.trial === null
									? words.text("progress.topics_note")
									: undefined
							}
							legend={
								shown !== undefined && (
									<MoveLegend
										gain={words.text("progress.legend_gain")}
										loss={words.text("progress.legend_loss")}
										period={words.text(
											shown === "week"
												? "progress.legend_week"
												: "progress.legend_task",
										)}
									/>
								)
							}
							rows={rows}
						/>
					</Fold>
				)}
				{report.mistakes.length > 0 && (
					<Fold
						title={words.text("progress.mistakes")}
						{...folding("mistakes")}
					>
						<StatList rows={mistakeRows(words, report.mistakes)} framed />
					</Fold>
				)}
				{latest.length > 0 && (
					<Fold
						title={words.text("progress.recent")}
						summary={
							<StatusDots
								tones={latest.map((entry) => entry.status.tone)}
								label={listed(
									words,
									latest.map((entry) => entry.status.label),
								)}
							/>
						}
						{...folding("recent")}
					>
						<StatList
							rows={recentRows(words, latest)}
							note={
								skipped > 0
									? words.text("progress.skipped_count", { count: skipped })
									: undefined
							}
						/>
					</Fold>
				)}
				<Fold
					title={words.text("profile.card_label")}
					summary={words.text("progress.for_parent")}
					{...folding("profile")}
				>
					<ParentProfile
						details={profile}
						host={host}
						onSaved={(saved) => {
							setProfile(saved);
							onSaved?.(saved);
						}}
					/>
					{report.location !== undefined && (
						<ParentData location={report.location} />
					)}
				</Fold>
			</div>
		</article>
	);
}

// Standing is the rank the child climbs: its name, the rank out of how many,
// the course of the ranks filled as far as the rating has come — striped
// where it moved over the while shown, with a line of how —, and the next
// rank to reach; or, while the trial series is still finding where the child
// stands, how many of its tasks are done, with no rank yet to show. The
// rating's number is the model's to say, and the answer's result shows it.
function Standing({
	report,
	moves,
	shown,
}: {
	report: ProgressReport;
	moves: Moves | undefined;
	shown: Period | undefined;
}) {
	const words = useWords();
	const { trial, overall } = report;
	if (trial !== null) {
		return (
			<RankSummary
				name={words.text("result.trial")}
				meta={words.text("result.trial_progress", {
					answered: trial.answered,
					total: trial.of,
				})}
				line={words.text("progress.trial_ranks", { total: trial.of })}
			>
				<Segments of={trial.of} filled={trial.answered} />
			</RankSummary>
		);
	}
	if (overall === null) {
		return null;
	}
	const top = overall.rank >= overall.ranks;
	const reading =
		moves === undefined || shown === undefined
			? undefined
			: readingOf(moveOver(moves, shown));
	return (
		<RankSummary
			label={words.text("progress.overall_label")}
			name={rankName(words, overall.rank)}
			meta={words.text("progress.rank_line", {
				rank: overall.rank,
				total: overall.ranks,
			})}
			line={
				top
					? words.text("progress.top_rank")
					: words.text("progress.next_rank", {
							name: rankName(words, overall.rank + 1),
						})
			}
			move={
				reading !== undefined &&
				shown !== undefined && (
					<MoveLine {...moveLineOf(words, shown, reading, overall.rank)} />
				)
			}
		>
			<Segments
				of={overall.ranks}
				filled={overall.rank - 1}
				part={(overall.share ?? 0) / 100}
				was={wasOf(reading)}
				label={
					top || overall.share === undefined
						? undefined
						: words.text("progress.share", {
								share: percentText(words, overall.share),
							})
				}
			/>
		</RankSummary>
	);
}

// MoveKind is what a line says of a while: how the rank moved, or that the
// while cannot be told.
type MoveKind =
	| "rank_up"
	| "forward"
	| "same"
	| "back"
	| "rank_down"
	| "untold";

// The words of each while's line, by what it says.
const moveLines: Record<Period, Record<MoveKind, Key>> = {
	week: {
		rank_up: "progress.week_rank_up",
		forward: "progress.week_forward",
		same: "progress.week_same",
		back: "progress.week_back",
		rank_down: "progress.week_rank_down",
		untold: "progress.week_untold",
	},
	last_task: {
		rank_up: "progress.task_rank_up",
		forward: "progress.task_forward",
		same: "progress.task_same",
		back: "progress.task_back",
		rank_down: "progress.task_rank_down",
		untold: "progress.task_untold",
	},
};

// moveLineOf is the line of how the overall rank moved over the while shown,
// rank being where it stands now: which way, and the words — from which rank
// to which when the rank changed, that it stayed when it did not move, that
// there is nothing to measure from yet when the while cannot be told, and
// nothing at all for a word this card does not know.
function moveLineOf(
	words: Words<Key>,
	period: Period,
	reading: Reading,
	rank: number,
): { way?: MoveWay; text?: string } {
	const said = moveLines[period];
	switch (reading.kind) {
		case "untold":
			return { text: words.text(said.untold) };
		case "same":
			return { text: words.text(said.same) };
		case "moved": {
			const gain = reading.way === "gain";
			if (!reading.ranked) {
				return {
					way: reading.way,
					text: words.text(gain ? said.forward : said.back),
				};
			}
			return {
				way: reading.way,
				text: words.text(gain ? said.rank_up : said.rank_down, {
					from: rankName(words, reading.rank),
					to: rankName(words, rank),
				}),
			};
		}
		default:
			return {};
	}
}

// wasOf is where a course stood before a move it is drawn with, or undefined
// when it is drawn with none.
function wasOf(reading: Reading | undefined): SegmentsAt | undefined {
	return reading?.kind === "moved"
		? { filled: reading.rank - 1, part: reading.share / 100 }
		: undefined;
}

// nextUp is what comes next, in words: the topic, and that it comes again
// when it is worked over after a mistake.
function nextUp(words: Words<Key>, next: Recommendation): string {
	const topic = topicName(words, next.topic);
	return next.goal === "reinforce"
		? words.text("progress.again", { topic })
		: topic;
}

// TopicRow is a topic as the list draws it, with how its rank moved over the
// while shown, where it is drawn with a move.
type TopicRow = RankRow & { reading?: Reading };

// topicRows are the topics listed, each with its rank, how that stands to the
// overall one — or how it moved by its own answers over the while shown, when
// a while is — and its course of the ranks in the ramp's step for it, striped
// where it moved; or, with no rank yet, its course drawn open and the word why
// — the trial series still running, or no answer in the topic so far. A topic
// of a progress from before the topics had ranks is drawn with an empty
// course, saying nothing.
function topicRows(
	words: Words<Key>,
	report: ProgressReport,
	shown: Period | undefined,
): TopicRow[] {
	const ranks = report.overall?.ranks ?? rankCount;
	return report.topics.map((topic): TopicRow => {
		const row = {
			id: topic.topic,
			label: topicName(words, topic.topic),
			mark: topic.mastered
				? { tone: "correct" as const, label: words.text("progress.mastered") }
				: undefined,
		};
		if (topic.rank !== null && topic.rank !== undefined) {
			const reading =
				shown === undefined || topic.change === undefined
					? undefined
					: readingOf(moveOver(topic.change, shown));
			const moved = movedWord(words, reading);
			return {
				...row,
				reading,
				word: moved?.word ?? comparedWord(words, topic.compared),
				way: moved?.way,
				name: rankName(words, topic.rank),
				segments: {
					of: ranks,
					filled: topic.rank - 1,
					part: (topic.share ?? 0) / 100,
					tone: toneOf(topic.rank, ranks),
					was: wasOf(reading),
				},
			};
		}
		if (report.trial !== null) {
			return {
				...row,
				word: words.text("progress.no_rank"),
				segments: { of: ranks, filled: 0, open: true },
			};
		}
		if (topic.answers === 0) {
			return {
				...row,
				word: words.text("progress.no_answers"),
				segments: { of: ranks, filled: 0, open: true },
			};
		}
		return { ...row, segments: { of: ranks, filled: 0 } };
	});
}

// movedWord is the word of how a topic moved by its own answers, with the
// move's way: a rank up or down, forward or back within its rank, or new to
// the while; none where it did not move or the while cannot be told.
function movedWord(
	words: Words<Key>,
	reading: Reading | undefined,
): { word: string; way: MoveWay } | undefined {
	if (reading?.kind === "new") {
		return { word: words.text("progress.note_new"), way: "gain" };
	}
	if (reading?.kind !== "moved") {
		return undefined;
	}
	const gain = reading.way === "gain";
	const key: Key = reading.ranked
		? gain
			? "progress.note_rank_up"
			: "progress.note_rank_down"
		: gain
			? "progress.note_forward"
			: "progress.note_back";
	return { word: words.text(key), way: reading.way };
}

// rowsMoved is how many topics moved up over the while shown and how many
// moved back, for the title of the topics, or nothing when none moved or no
// while is shown.
function rowsMoved(words: Words<Key>, rows: readonly TopicRow[]) {
	const { up, down } = movesCounted(
		rows.flatMap((row) => (row.reading === undefined ? [] : [row.reading])),
	);
	if (up === 0 && down === 0) {
		return undefined;
	}
	const said: string[] = [];
	if (up > 0) {
		said.push(words.text("progress.moved_up", { count: up }));
	}
	if (down > 0) {
		said.push(words.text("progress.moved_down", { count: down }));
	}
	return (
		<MoveCounts
			up={up > 0 ? countText(words, up) : undefined}
			down={down > 0 ? countText(words, down) : undefined}
			label={listed(words, said)}
		/>
	);
}

// comparedWord is how a topic's rank stands to the overall one, in words: a
// rank above it is ahead, one below it behind, and the overall rank itself
// even with it.
function comparedWord(
	words: Words<Key>,
	compared: string | null | undefined,
): string | undefined {
	switch (compared) {
		case "ahead":
			return words.text("progress.ahead");
		case "behind":
			return words.text("progress.behind");
		case "even":
			return words.text("progress.even");
		default:
			return undefined;
	}
}

// toneOf is the step of the ramp a rank is drawn in: five steps over the
// ranks there are, the first ranks in the first.
function toneOf(rank: number, ranks: number): SegmentTone {
	const step = 1 + Math.floor(((rank - 1) * 5) / ranks);
	return Math.min(Math.max(step, 1), 5) as SegmentTone;
}

// Status is how an entry went, in words and as the tone of its mark.
type Status = NonNullable<StatRow["status"]>;

// Latest is one of the latest entries the progress shows: its topic, and how
// it went.
type Latest = { topic: string; status: Status };

// latestOf are the latest entries the progress shows, the newest first, each
// with how it went.
function latestOf(
	words: Words<Key>,
	recent: ProgressReport["recent"],
): Latest[] {
	return recent
		.slice(0, recentShown)
		.map((entry) => ({ topic: entry.topic, status: statusOf(words, entry) }));
}

// recentRows are the latest entries as lines: each answer right or wrong, and
// each task left without one.
function recentRows(words: Words<Key>, latest: readonly Latest[]): StatRow[] {
	return latest.map(
		(entry, place): StatRow => ({
			id: `${place}`,
			label: topicName(words, entry.topic),
			status: entry.status,
		}),
	);
}

// statusOf is how an entry went, in words.
function statusOf(
	words: Words<Key>,
	entry: ProgressReport["recent"][number],
): Status {
	if (entry.skipped || entry.correct === null) {
		return { tone: "skipped", label: words.text("progress.skipped") };
	}
	return entry.correct
		? { tone: "correct", label: words.text("progress.right") }
		: { tone: "wrong", label: words.text("progress.wrong") };
}

// mistakeRows are the mistakes that keep coming back, the most frequent
// first: each with how many times, and a dot for each time.
function mistakeRows(
	words: Words<Key>,
	mistakes: ProgressReport["mistakes"],
): StatRow[] {
	return mistakes.map(
		(mistake): StatRow => ({
			id: mistake.trap,
			label: trapName(words, mistake.trap),
			count: {
				times: mistake.times,
				label: words.text("progress.times", { count: mistake.times }),
			},
		}),
	);
}
