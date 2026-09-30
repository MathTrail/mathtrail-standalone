import { Note } from "../design/blocks";
import { RatingSummary, StatList, type StatRow } from "../design/progress";
import { MessageHeader } from "../design/thread";
import type { Words } from "../i18n/words";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import { rankName, topicName, trapName } from "./names";
import { ParentProfile } from "./ProfileScreen";
import type { ProgressReport, Recommendation } from "./payload";
import { type Key, ratingText, useWords } from "./words";

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
 * ProgressScreen is where the child stands: the overall rating with its rank
 * above it — or, while the trial series runs, how far the series has got —
 * what comes next, the topics met with their ratings, the latest answers and
 * how many tasks were left without one, the mistakes that keep coming back,
 * and the child's profile for the parent.
 */
export function ProgressScreen({
	report,
	wide,
	host,
}: {
	report: ProgressReport;
	wide: boolean;
	host: Host;
}) {
	const words = useWords();
	const { profile, recommendation } = report;
	const skipped = report.topics.reduce((sum, topic) => sum + topic.skipped, 0);
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
					<StatList
						label={words.text("progress.topics")}
						rows={topicRows(words, report.topics)}
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
					/>
				)}
				<ParentProfile details={profile} host={host} />
			</div>
		</article>
	);
}

// Standing is the number the child climbs: the overall rating under its
// rank, or — while the trial series is still finding where the child stands
// — how many of its tasks are done, with no rating yet to show.
function Standing({ report }: { report: ProgressReport }) {
	const words = useWords();
	const { trial, overall } = report;
	if (trial !== null) {
		return (
			<RatingSummary
				rankLabel={words.text("result.trial")}
				rating={words.text("result.trial_progress", {
					answered: trial.answered,
					total: trial.of,
				})}
				ratingLabel={words.text("progress.trial_after", { total: trial.of })}
				pips={{ on: trial.answered, of: trial.of }}
			/>
		);
	}
	if (overall === null) {
		return null;
	}
	return (
		<RatingSummary
			label={words.text("progress.overall_label")}
			rankLabel={words.text("progress.rank", {
				rank: overall.rank,
				total: overall.ranks,
				name: rankName(words, overall.rank),
			})}
			rating={ratingText(words, overall.rating)}
			ratingLabel={words.text("progress.overall")}
			pips={{ on: overall.rank, of: overall.ranks }}
		/>
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

// topicRows are the topics met, each with its rating once it has one, and
// marked when it is mastered.
function topicRows(
	words: Words<Key>,
	topics: ProgressReport["topics"],
): StatRow[] {
	return topics.map(
		(topic): StatRow => ({
			id: topic.topic,
			label: topicName(words, topic.topic),
			status: topic.mastered
				? { tone: "correct", label: words.text("progress.mastered") }
				: undefined,
			value:
				topic.rating === null ? undefined : ratingText(words, topic.rating),
		}),
	);
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
// first: each with how many times, and a bar as long as its share of the most
// frequent.
function mistakeRows(
	words: Words<Key>,
	mistakes: ProgressReport["mistakes"],
): StatRow[] {
	const most = Math.max(...mistakes.map((mistake) => mistake.times));
	return mistakes.map(
		(mistake): StatRow => ({
			id: mistake.trap,
			label: trapName(words, mistake.trap),
			bar: {
				share: mistake.times / most,
				count: words.text("progress.times", { count: mistake.times }),
			},
		}),
	);
}
