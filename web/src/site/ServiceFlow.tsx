import { type ComponentChildren, Fragment } from "preact";
import type { PageReader } from "./reader";
import { ServiceSection } from "./ServiceSection";
import { ServiceSwitch } from "./ServiceSwitch";
import {
	firstState,
	forward,
	gridColumn,
	numberedSteps,
	type Scenario,
	type Step,
	scenarioId,
	scenarios,
	stepsId,
	turnName,
	turns,
} from "./service";
import { TableFrame } from "./TableFrame";

// lineKinds are the kinds of arrow the steps are drawn with, in the order the
// legend names them.
const lineKinds = ["call", "reply", "host"] as const;

/**
 * ServiceFlow is a lesson call by call: a switch between its three scenarios
 * — a task, an answer, the sign-in — and the legend of the arrows, each
 * scenario's steps between its actors with a line under them, and then the
 * states the task on the card goes through. Every scenario is in the page;
 * the rules of the page's head show the one chosen.
 */
export function ServiceFlow({ page }: { page: PageReader }) {
	return (
		<ServiceSection page={page} id="flow">
			<div class="s-service-flow-bar">
				<ServiceSwitch
					legend={page.text("flow.legend")}
					name="service-scenario"
					options={scenarios.map((scenario, at) => ({
						value: scenario.id,
						id: scenarioId(scenario.id),
						label: page.text(`flow.scenarios.${scenario.id}`),
						chosen: at === 0,
					}))}
				/>
				<ul class="s-service-kinds">
					{lineKinds.map((kind) => (
						<li key={kind} class={`s-service-kind-${kind}`}>
							{page.text(`flow.kinds.${kind}`)}
						</li>
					))}
				</ul>
			</div>
			{scenarios.map((scenario) => (
				<Steps key={scenario.id} page={page} scenario={scenario} />
			))}
			<Machine page={page} />
		</ServiceSection>
	);
}

// Steps is one scenario: its actors in a row over their lifelines, and its
// steps one under another, each between the actors it joins, the arrows
// numbered and the notes not, then the line that closes it.
function Steps({ page, scenario }: { page: PageReader; scenario: Scenario }) {
	const name = `${stepsId(scenario.id)}-name`;
	const actorOf = (at: number) =>
		page.plain(`flow.actors.${scenario.actors[at - 1]?.id}`);
	const rows = numberedSteps(scenario.steps).map(({ step, number }, at) => ({
		step,
		number,
		words: `flow.${scenario.id}.rows.${at + 1}`,
	}));
	return (
		<div id={stepsId(scenario.id)} class="s-service-steps">
			<h3 id={name} class="s-hidden">
				{page.text(`flow.scenarios.${scenario.id}`)}
			</h3>
			<TableFrame labelledBy={name} class="s-service-frame">
				<div
					class="s-service-diagram"
					dir="ltr"
					style={{
						"--actors": String(scenario.actors.length),
						"--halves": String(scenario.actors.length * 2),
					}}
				>
					<div class="s-service-lifelines" aria-hidden="true">
						{scenario.actors.map((actor) => (
							<span key={actor.id} />
						))}
					</div>
					<ul class="s-service-actors">
						{scenario.actors.map((actor) => (
							<li key={actor.id}>
								<span class={`s-service-actor s-service-actor-${actor.kind}`}>
									{page.text(`flow.actors.${actor.id}`)}
								</span>
							</li>
						))}
					</ul>
					<ol class="s-service-rows">
						{rows.map(({ step, words, number }) => (
							<li key={words} class="s-service-row">
								{step.kind === "note" ? (
									<p
										class={`s-service-note s-service-note-${step.tone}`}
										style={{ gridColumn: gridColumn(step) }}
									>
										{page.text(words)}
									</p>
								) : (
									<Arrow
										step={step}
										number={number}
										said={`${actorOf(step.from)} → ${actorOf(step.to)}: `}
									>
										{page.text(words)}
									</Arrow>
								)}
							</li>
						))}
					</ol>
				</div>
			</TableFrame>
			<p class="s-service-caption">
				{page.text(`flow.${scenario.id}.caption`)}
			</p>
		</div>
	);
}

// Arrow is a numbered step from one actor to another: its number and words
// over a line of its kind, pointing the way the step goes. A screen reader,
// which does not see the columns, is told who the step goes from and to.
function Arrow({
	step,
	number,
	said,
	children,
}: {
	step: Step;
	number: number;
	said: string;
	children: ComponentChildren;
}) {
	const back = step.to < step.from;
	return (
		<div
			class={`s-service-arrow s-service-arrow-${step.kind}`}
			style={{ gridColumn: gridColumn(step) }}
		>
			<p class="s-service-arrow-words">
				<span class="s-service-step">{number}</span>
				<span>
					<span class="s-hidden">{said}</span>
					{children}
				</span>
			</p>
			<span
				class={back ? "s-service-line s-service-line-back" : "s-service-line"}
				aria-hidden="true"
			/>
		</div>
	);
}

// Machine is the task on the card as a machine of states: the states in the
// order a task goes through them with the call between each two, then every
// other way the task goes, each with what makes it go so.
function Machine({ page }: { page: PageReader }) {
	const name = "machine-name";
	return (
		<div class="s-service-machine">
			<h3 id={name}>{page.text("flow.machine.title")}</h3>
			<p class="s-service-machine-lead">{page.text("flow.machine.lead")}</p>
			<TableFrame labelledBy={name} class="s-service-frame">
				<div class="s-service-machine-board" dir="ltr">
					<p class="s-service-states">
						<code class={`s-service-state s-service-state-${firstState}`}>
							{firstState}
						</code>
						{forward.map(({ tool, checked, to }, at) => (
							<Fragment key={tool}>
								<span class="s-service-forward">
									<code>{checked ? `${tool} ✓` : tool}</code>
									<span class="s-service-forward-line" aria-hidden="true" />
									<span class="s-service-forward-words">
										{page.text(`flow.machine.forward.${at + 1}`)}
									</span>
								</span>
								<code class={`s-service-state s-service-state-${to}`}>
									{to}
								</code>
							</Fragment>
						))}
					</p>
					<ul class="s-service-turns">
						{turns.map((turn, at) => (
							<li key={turnName(turn)}>
								<code class={`s-service-turn-${turn.tone}`}>
									{turnName(turn)}
								</code>
								<span>{page.text(`flow.machine.turns.${at + 1}`)}</span>
							</li>
						))}
					</ul>
				</div>
			</TableFrame>
		</div>
	);
}
