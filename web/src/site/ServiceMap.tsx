import type { PageReader } from "./reader";
import { ServiceSection } from "./ServiceSection";
import {
	aboutId,
	type Column,
	columns,
	faceId,
	type Group,
	groupOf,
	linkKey,
	neighbours,
	parts,
	pickId,
	quickPicks,
	tools,
	toolsPart,
	wholeMap,
} from "./service";

// partChoice is the name the map's parts share as one group of radio buttons,
// with the choice of none among them.
const partChoice = "service-part";

/**
 * ServiceMap is the map of the service: who talks to it, its own parts, and
 * the services outside its binary, each part a choice. With a part chosen,
 * the parts it talks to stay bright and the rest fade, and the panel under the
 * map, which stays in sight as the map scrolls, says what the part does and
 * what it talks to, each of those a choice too; before any is chosen, it
 * offers a few. The choices are radio buttons, lit by the rules the page's
 * head carries, so no script runs, and every part's details are in the page.
 */
export function ServiceMap({ page }: { page: PageReader }) {
	return (
		<ServiceSection page={page} id="map">
			<fieldset class="s-service-map">
				<legend class="s-hidden">{page.text("map.legend")}</legend>
				<div class="s-service-columns">
					{columns.map((column) => (
						<MapColumn key={column.id} page={page} column={column} />
					))}
				</div>
				<Panel page={page} />
			</fieldset>
		</ServiceSection>
	);
}

// MapColumn is one column of the map: its heading and line, the note beside
// them where the column has one, and its groups of parts.
function MapColumn({ page, column }: { page: PageReader; column: Column }) {
	const words = `map.columns.${column.id}`;
	return (
		<div class={`s-service-column s-service-column-${column.id}`}>
			<div class="s-service-column-head">
				<div>
					<h3>{page.text(`${words}.title`)}</h3>
					<p>{page.text(`${words}.sub`)}</p>
				</div>
				{column.tagged && (
					<p class="s-service-tag">{page.text(`${words}.tag`)}</p>
				)}
			</div>
			{column.groups.map((group) => (
				<MapGroup key={group.id} page={page} group={group} />
			))}
		</div>
	);
}

// MapGroup is parts of a column under their label, in a grid, or alone in a
// row of their own when the group has one.
function MapGroup({ page, group }: { page: PageReader; group: Group }) {
	return (
		<div class="s-service-group">
			<p class="s-service-group-title">
				{page.text(`map.groups.${group.id}.title`)}
			</p>
			<div
				class={
					group.parts.length === 1
						? "s-service-parts s-service-parts-one"
						: "s-service-parts"
				}
			>
				{group.parts.map((part) => (
					<Part key={part} page={page} part={part} />
				))}
			</div>
		</div>
	);
}

// Part is one part of the map as the label of its choice: the part's name, a
// line where it has one, and for the tools every tool by name.
function Part({ page, part }: { page: PageReader; part: string }) {
	const words = `map.parts.${part}`;
	return (
		<label for={pickId(part)} id={faceId(part)} class="s-service-face">
			<span class="s-service-name">{page.text(`${words}.name`)}</span>
			{page.has(`${words}.sub`) && (
				<span class="s-service-sub">{page.text(`${words}.sub`)}</span>
			)}
			{part === toolsPart && (
				<span class="s-service-tools">
					{tools.map((tool) => (
						<code key={tool}>{tool}</code>
					))}
				</span>
			)}
		</label>
	);
}

// Panel is what stays in sight under the map: before a part is chosen, how to
// choose and a few parts to start with; once one is, that part's details. It
// holds the choices themselves, each named by its part's name, and the choice
// of none, which the details' button to clear chooses: wherever a part's
// label is pressed, the choice it moves to is in sight, and the page does not
// scroll to it.
function Panel({ page }: { page: PageReader }) {
	return (
		<div class="s-service-panel">
			{parts.map((part) => (
				<input
					key={part}
					type="radio"
					class="s-service-pick s-hidden"
					name={partChoice}
					id={pickId(part)}
					aria-label={page.plain(`map.parts.${part}.name`)}
				/>
			))}
			<input
				type="radio"
				class="s-service-whole s-hidden"
				name={partChoice}
				id={wholeMap}
				aria-label={page.plain("map.whole")}
				checked
			/>
			<div class="s-service-hint">
				<p>{page.text("map.hint")}</p>
				<p class="s-service-quick">
					{quickPicks.map((part) => (
						<label key={part} for={pickId(part)}>
							{page.text(`map.parts.${part}.name`)}
						</label>
					))}
				</p>
			</div>
			{parts.map((part) => (
				<About key={part} page={page} part={part} />
			))}
		</div>
	);
}

// About is a part's details in the panel: its place, its name and what it
// does, beside every part it talks to, with what passes between them, each a
// choice of that part.
function About({ page, part }: { page: PageReader; part: string }) {
	return (
		<div id={aboutId(part)} class="s-service-about">
			<div class="s-service-about-text">
				<p class="s-service-about-group">
					{page.text(`map.groups.${groupOf(part)}.short`)}
				</p>
				<p class="s-service-about-name">
					{page.text(`map.parts.${part}.name`)}
				</p>
				<p class="s-service-about-what">
					{page.text(`map.parts.${part}.about`)}
				</p>
			</div>
			<div class="s-service-about-links">
				<div class="s-service-about-head">
					<p>{page.text("map.linked")}</p>
					<label for={wholeMap} class="s-service-reset">
						{page.text("map.reset")}
					</label>
				</div>
				<ul class="s-service-neighbours">
					{neighbours(part).map(({ part: other, link }) => (
						<li key={other}>
							<label for={pickId(other)}>
								<span class="s-service-neighbour">
									{page.text(`map.parts.${other}.name`)}
								</span>
								<span class="s-service-passes">
									{page.text(`map.links.${linkKey(link)}`)}
								</span>
							</label>
						</li>
					))}
				</ul>
			</div>
		</div>
	);
}
