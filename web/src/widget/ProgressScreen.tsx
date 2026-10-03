import { useState } from "preact/hooks";
import { Note } from "../design/blocks";
import {
	RankList,
	type RankRow,
	RankSummary,
	Segments,
	type SegmentTone,
	StatList,
	type StatRow,
} from "../design/progress";
import { MessageHeader } from "../design/thread";
import type { Words } from "../i18n/words";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import { rankCount, rankName, topicName, trapName } from "./names";
import { ParentData, ParentProfile } from "./ProfileScreen";
import type { Details, ProgressReport, Recommendation } from "./payload";
import { type Key, percentText, ratingText, useWords } from "./words";

/** recentShown is how many of the latest entries the progress lists. */
const recentShown = 5;

/**
 * ProgressCard is the card of the child's progress, drawn when the model
 * reads it. No task stands behind it, so it has no way back to one.
 */
export function ProgressCard({
	report,
	host,
}: {
	report: ProgressReport;
	host: Host;
}) {
	return (
		<CardRoot>
			{(wide) => <ProgressScreen report={report} wide={wide} host={host} />}
		</CardRoot>
	);
}

/**
 * ProgressScreen is where the child stands: the rank reached and how far
 * through it — or, while the trial series runs, how far the series has got —
 * what comes next, each topic met or within reach with a rank of its own, the
 * latest answers and how many tasks were left without one, the mistakes that
 * keep coming back, and the child's profile for the parent with what the
 * parent can do with the data. A change the parent saves on its form shows at
 * once, the name and the grade at the top among it, and is handed to onSaved.
 */
export function ProgressScreen({
	report,
	wide,
	host,
	onSaved,
}: {
	report: ProgressReport;
	wide: boolean;
	host: Host;
	onSaved?: (details: Details) => void;
}) {
	const words = useWords();
	const { recommendation } = report;
	const [profile, setProfile] = useState(report.profile);
	const skipped =
		report.skipped ??
		report.topics.reduce((sum, topic) => sum + topic.skipped, 0);
	return (
		<article aria-label={words.text("progress.label")}>
			<MessageHeader
				author="person"
				name={profile.pseudonym}
				badge={words.text("child.grade", { grade: profile.grade })}
				wide={wide}
			/>
			<div class="mt-progress">
				<Standing report={report} />
				{recommendation !== null && (
					<Note label={words.text("progress.next_up")}>
						{nextUp(words, recommendation)}
					</Note>
				)}
				{report.topics.length > 0 && (
					<RankList
						label={words.text("progress.topics")}
						note={
							report.trial === null
								? words.text("progress.topics_note")
								: undefined
						}
						rows={topicRows(words, report)}
					/>
				)}
				{report.recent.length > 0 && (
					<StatList
						label={words.text("progress.recent")}
						rows={recentRows(words, report.recent)}
						note={
							skipped > 0
								? words.text("progress.skipped_count", { count: skipped })
								: undefined
						}
					/>
				)}
				{report.mistakes.length > 0 && (
					<StatList
						label={words.text("progress.mistakes")}
						rows={mistakeRows(words, report.mistakes)}
						framed
					/>
				)}
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
			</div>
		</article>
	);
}

// Standing is the rank the child climbs: its name, the rank out of how many
// with the rating beside it, the course of the ranks filled as far as the
// rating has come, and the next rank to reach — or, while the trial series is
// still finding where the child stands, how many of its tasks are done, with
// no rank yet to show.
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
				rating: ratingText(words, overall.rating),
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

// recentRows are the latest entries, the newest first: each answer right or
// wrong, and each task left without one.
function recentRows(
	words: Words<Key>,
	recent: ProgressReport["recent"],
): StatRow[] {
	return recent.slice(0, recentShown).map(
		(entry, place): StatRow => ({
			id: `${place}`,
			label: topicName(words, entry.topic),
			status: statusOf(words, entry),
		}),
	);
}

// statusOf is how an entry went, in words.
function statusOf(
	words: Words<Key>,
	entry: ProgressReport["recent"][number],
): StatRow["status"] {
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
