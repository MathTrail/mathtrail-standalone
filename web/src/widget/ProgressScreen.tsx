import { useState } from "preact/hooks";
import { Fold, Note } from "../design/blocks";
import {
	RankList,
	type RankRow,
	RankSummary,
	Segments,
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
import { listed, rankCount, rankName, topicName, trapName } from "./names";
import { ParentData, ParentProfile } from "./ProfileScreen";
import type { Details, ProgressReport, Recommendation } from "./payload";
import { type Key, percentText, useWords } from "./words";

/** recentShown is how many of the latest entries the progress lists. */
const recentShown = 5;

/**
 * ProgressCard is the card of the child's progress, drawn when the model
 * reads it. No task stands behind it, so it has no way back to one; its
 * sections stay as they were opened for as long as the card is drawn.
 */
export function ProgressCard({
	report,
	host,
}: {
	report: ProgressReport;
	host: Host;
}) {
	const folds = useFolds();
	return (
		<CardRoot>
			{(wide) => (
				<ProgressScreen report={report} wide={wide} host={host} folds={folds} />
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
 * open is told by folds, kept by whoever outlives the screen. A change the
 * parent saves on its form shows at once, the name and the grade at the top
 * among it, and is handed to onSaved.
 */
export function ProgressScreen({
	report,
	wide,
	host,
	folds,
	onSaved,
}: {
	report: ProgressReport;
	wide: boolean;
	host: Host;
	folds: Folds;
	onSaved?: (details: Details) => void;
}) {
	const words = useWords();
	const { recommendation } = report;
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
					<Standing report={report} />
					{recommendation !== null && (
						<Note label={words.text("progress.next_up")}>
							{nextUp(words, recommendation)}
						</Note>
					)}
				</div>
			)}
			<div class="mt-folds">
				{report.topics.length > 0 && (
					<Fold title={words.text("progress.topics")} {...folding("topics")}>
						<RankList
							note={
								report.trial === null
									? words.text("progress.topics_note")
									: undefined
							}
							rows={topicRows(words, report)}
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
// the course of the ranks filled as far as the rating has come, and the next
// rank to reach — or, while the trial series is still finding where the child
// stands, how many of its tasks are done, with no rank yet to show. The
// rating's number is the model's to say, and the answer's result shows it.
function Standing({ report }: { report: ProgressReport }) {
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
		>
			<Segments
				of={overall.ranks}
				filled={overall.rank - 1}
				part={(overall.share ?? 0) / 100}
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

// nextUp is what comes next, in words: the topic, and that it comes again
// when it is worked over after a mistake.
function nextUp(words: Words<Key>, next: Recommendation): string {
	const topic = topicName(words, next.topic);
	return next.goal === "reinforce"
		? words.text("progress.again", { topic })
		: topic;
}

// topicRows are the topics listed, each with its rank, how that stands to the
// overall one and its course of the ranks in the ramp's step for it; or, with
// no rank yet, its course drawn open and the word why — the trial series still
// running, or no answer in the topic so far. A topic of a progress from
// before the topics had ranks is drawn with an empty course, saying nothing.
function topicRows(words: Words<Key>, report: ProgressReport): RankRow[] {
	const ranks = report.overall?.ranks ?? rankCount;
	return report.topics.map((topic): RankRow => {
		const row = {
			id: topic.topic,
			label: topicName(words, topic.topic),
			mark: topic.mastered
				? { tone: "correct" as const, label: words.text("progress.mastered") }
				: undefined,
		};
		if (topic.rank !== null && topic.rank !== undefined) {
			return {
				...row,
				word: comparedWord(words, topic.compared),
				name: rankName(words, topic.rank),
				segments: {
					of: ranks,
					filled: topic.rank - 1,
					part: (topic.share ?? 0) / 100,
					tone: toneOf(topic.rank, ranks),
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

// comparedWord is how a topic's rank stands to the overall one, in words: a
// rank above it is ahead, one below it behind, and the overall rank itself
// needs no word.
function comparedWord(
	words: Words<Key>,
	compared: string | null | undefined,
): string | undefined {
	switch (compared) {
		case "ahead":
			return words.text("progress.ahead");
		case "behind":
			return words.text("progress.behind");
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
