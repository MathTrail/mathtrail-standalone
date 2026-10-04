import type { PageReader } from "./reader";
import { board, lineId, mapOf, nodeId, sayingId } from "./topicmap";
import type { Topic, Topics } from "./topics";

/**
 * TopicMap draws how the topics build on one another: a column for each
 * layer under its step's words, a box for each topic leading to its card on
 * the page, a line for each link, and the line of words each topic says while
 * it is pointed at. The map is laid out from left to right in every language,
 * as a drawing is, and it scrolls sideways inside its panel where the screen
 * is narrower than it. Pointing at a topic, or reaching it with the keyboard,
 * lights what it stands on and what it opens, by the rules the page's head
 * carries; no script runs.
 */
export function TopicMap({
	page,
	topics,
	name,
	names,
}: {
	page: PageReader;
	topics: Topics;
	name: (topic: Topic) => string;
	names: (ids: readonly string[]) => string;
}) {
	const map = mapOf(topics);
	const steps = page.list("map.layers");
	if (steps.length !== map.columns.length) {
		throw new Error(
			`the page's words name ${steps.length} steps of the map, and the topics stand in ${map.columns.length} layers`,
		);
	}
	return (
		<div class="s-map">
			<div class="s-map-scroll">
				<div class="s-map-board" style={{ inlineSize: `${map.width}px` }}>
					<div
						class="s-map-heads"
						style={{
							gridTemplateColumns: `repeat(${map.columns.length}, ${board.width}px)`,
							columnGap: `${board.columnGap}px`,
						}}
					>
						{steps.map((step) => (
							<div key={step} class="s-map-head">
								<p class="s-map-step">{page.text(`${step}.step`)}</p>
								<h3>{page.text(`${step}.name`)}</h3>
								<p class="s-map-head-line">{page.text(`${step}.lead`)}</p>
							</div>
						))}
					</div>
					<div class="s-map-area" style={{ blockSize: `${map.height}px` }}>
						<svg
							class="s-map-lines"
							width={map.width}
							height={map.height}
							viewBox={`0 0 ${map.width} ${map.height}`}
							aria-hidden="true"
						>
							{map.lines.map((line) => (
								<g key={lineId(line)} id={lineId(line)} class="s-map-line">
									<path d={line.path} />
									<circle cx={line.end[0]} cy={line.end[1]} r="3" />
								</g>
							))}
						</svg>
						{map.spots.map(({ topic, x, y }) => (
							<a
								key={topic.id}
								id={nodeId(topic)}
								class="s-map-node"
								href={`#${topic.slug}`}
								style={{
									insetInlineStart: `${x}px`,
									insetBlockStart: `${y}px`,
									inlineSize: `${board.width}px`,
									blockSize: `${board.height}px`,
								}}
							>
								<span dir="auto">{name(topic)}</span>
								<span aria-hidden="true">›</span>
							</a>
						))}
					</div>
				</div>
			</div>
			<p class="s-map-hint">{page.text("map.hint")}</p>
			{topics.all.map((topic) => (
				<p key={topic.id} id={sayingId(topic)} class="s-map-say">
					{page.text("map.say", {
						topic: name(topic),
						before:
							topic.before.length === 0
								? page.plain("map.foundation")
								: names(topic.before),
						after:
							topic.after.length === 0
								? page.plain("map.summit")
								: names(topic.after),
					})}
				</p>
			))}
			<ul class="s-map-legend">
				<li class="s-map-first">{page.text("map.legend.first")}</li>
				<li class="s-map-topic">{page.text("map.legend.topic")}</li>
				<li class="s-map-then">{page.text("map.legend.then")}</li>
			</ul>
		</div>
	);
}
