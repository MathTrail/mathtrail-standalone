import { describe, expect, test } from "vitest";
import {
	aboutId,
	blocks,
	columns,
	coreParts,
	faceId,
	forward,
	gridColumn,
	groupOf,
	lifeScale,
	lifetimes,
	linkKey,
	links,
	logShare,
	neighbours,
	numberedSteps,
	parts,
	partsOf,
	pickId,
	quickPicks,
	reachOf,
	scenarioId,
	scenarios,
	serviceRules,
	shareOf,
	sizeScale,
	stepsId,
	stores,
	taskStates,
	tools,
	toolsPart,
	totalOf,
	turnName,
	turns,
} from "./service";

// rulesOf are the selectors of the rule of the page's head that sets what,
// one by one.
function rulesOf(what: string): string[] {
	const rules = serviceRules();
	const end = rules.indexOf(`{${what}}`);
	if (end < 0) {
		return [];
	}
	return rules
		.slice(rules.lastIndexOf("}", end) + 1, end)
		.split(",")
		.map((selector) => selector.trim())
		.filter((selector) => selector !== "");
}

const ringed = rulesOf("opacity:1;box-shadow:0 0 0 1.5px var(--s-accent)");

// chosen is the part of the selector of a rule of the page's head that holds
// with part chosen on the map.
const chosen = (part: string) => `#map:has(#${pickId(part)}:checked) `;

describe("the map of the service", () => {
	test("names every part once, by an id its words and ids are made of", () => {
		expect(new Set(parts).size).toBe(parts.length);
		for (const part of parts) {
			expect(part).toMatch(/^[a-z][a-z0-9]*$/);
		}
		expect(parts).not.toContain("whole");
	});

	test("stands every part in one group of one column, and names every group once", () => {
		const groups = columns.flatMap((column) => column.groups);

		expect(new Set(groups.map(({ id }) => id)).size).toBe(groups.length);
		expect(groupOf("cimd")).toBe("host");
		expect(groupOf("tools")).toBe("tools");
		expect(() => groupOf("nowhere")).toThrow(
			"the map of the service has no part nowhere",
		);
		expect(partsOf("people")).toEqual(["adult", "kid"]);
		expect(() => partsOf("nowhere")).toThrow(
			"the map of the service has no group nowhere",
		);
	});

	test("links two different parts it has, each pair once, and leaves no part unlinked", () => {
		const pairs = links.map(([from, to]) => [from, to].sort().join(" "));

		expect(new Set(pairs).size).toBe(links.length);
		for (const [from, to] of links) {
			expect(from).not.toBe(to);
			expect(parts).toContain(from);
			expect(parts).toContain(to);
		}
		for (const part of parts) {
			expect(neighbours(part), part).not.toEqual([]);
		}
	});

	test("names a link's words by its two parts, the first the one that starts the talk", () => {
		expect(linkKey(["authsrv", "cimd"])).toBe("authsrv-cimd");
		expect(neighbours("cimd")).toEqual([
			{ part: "authsrv", link: ["authsrv", "cimd"] },
		]);
	});

	test("offers the parts it has for a first choice and for its tools, and names its group of pure computation as its core", () => {
		for (const part of [...quickPicks, toolsPart]) {
			expect(parts).toContain(part);
		}
		expect(coreParts).toEqual([
			"rule",
			"rating",
			"checks",
			"solver",
			"content",
		]);
	});

	test("lists every tool once", () => {
		expect(tools).toHaveLength(11);
		expect(new Set(tools).size).toBe(tools.length);
	});

	test("rings, with a part chosen, exactly the parts it talks to", () => {
		for (const part of parts) {
			expect(
				ringed
					.filter((selector) => selector.startsWith(chosen(part)))
					.map((selector) => selector.slice(chosen(part).length)),
				part,
			).toEqual(neighbours(part).map(({ part: other }) => `#${faceId(other)}`));
		}
		expect(ringed).toHaveLength(links.length * 2);
	});

	test("fills the part chosen, and outlines the part whose choice the keyboard is on", () => {
		expect(
			rulesOf(
				"opacity:1;background:var(--s-accent-tint);box-shadow:0 0 0 2px var(--s-accent)",
			),
		).toEqual(parts.map((part) => `${chosen(part)}#${faceId(part)}`));
		expect(
			rulesOf("outline:2px solid var(--s-accent);outline-offset:2px"),
		).toEqual(
			parts.map(
				(part) => `#map:has(#${pickId(part)}:focus-visible) #${faceId(part)}`,
			),
		);
	});

	test("shows, with a part chosen, its details alone, and with a scenario chosen, its steps alone", () => {
		expect(rulesOf("display:grid")).toEqual(
			parts.map((part) => `${chosen(part)}#${aboutId(part)}`),
		);
		expect(rulesOf("display:block")).toEqual(
			scenarios.map(
				({ id }) => `#flow:has(#${scenarioId(id)}:checked) #${stepsId(id)}`,
			),
		);
	});

	test("only ever shows with the rules of the page's head, and hides nothing", () => {
		expect(serviceRules()).not.toMatch(/display:none|opacity:0[;}]/);
	});
});

describe("a lesson, call by call", () => {
	test("keeps every step between two actors of its scenario, an arrow from one to another", () => {
		for (const { id, actors, steps } of scenarios) {
			for (const step of steps) {
				for (const end of [step.from, step.to]) {
					expect(end, id).toBeGreaterThanOrEqual(1);
					expect(end, id).toBeLessThanOrEqual(actors.length);
				}
				if (step.kind === "note") {
					expect(step.from, id).toBeLessThanOrEqual(step.to);
				} else {
					expect(step.from, id).not.toBe(step.to);
				}
			}
		}
	});

	test("draws a call between the middles of its actors, either way, and a note across both of them whole", () => {
		expect(gridColumn({ kind: "call", from: 1, to: 3 })).toBe("2 / 6");
		expect(gridColumn({ kind: "reply", from: 4, to: 2 })).toBe("4 / 8");
		expect(gridColumn({ kind: "note", from: 4, to: 5, tone: "service" })).toBe(
			"7 / 11",
		);
	});

	test("numbers a scenario's arrows from one, and its notes not at all", () => {
		expect(
			numberedSteps([
				{ kind: "call", from: 1, to: 2 },
				{ kind: "note", from: 1, to: 2, tone: "host" },
				{ kind: "reply", from: 2, to: 1 },
				{ kind: "host", from: 1, to: 2 },
			]).map(({ number }) => number),
		).toEqual([1, 0, 2, 3]);
	});

	test("takes the task on the card forward by a call between each two states, and names its other ways from one state to another", () => {
		expect(taskStates).toEqual(["none", "requested", "issued", "answered"]);
		for (const { tool } of forward) {
			expect(tools).toContain(tool);
		}
		for (const turn of turns) {
			expect(taskStates).toContain(turn.from);
			expect(taskStates).toContain(turn.to);
		}
		expect(new Set(turns.map(turnName)).size).toBe(turns.length);
		expect(
			turnName({ from: "answered", to: "answered", tone: "answered" }),
		).toBe("answered ↺");
		expect(turnName({ from: "issued", to: "none", tone: "issued" })).toBe(
			"issued → none",
		);
	});
});

describe("the profile's file", () => {
	test("keeps each block's usual size within its cap, and the file near its caps within the bar, past the size it was meant to keep within", () => {
		for (const block of blocks) {
			expect(block.typical, block.id).toBeLessThanOrEqual(block.cap);
		}
		expect(totalOf("typical")).toBeLessThanOrEqual(sizeScale.goal);
		expect(totalOf("cap")).toBeGreaterThan(sizeScale.goal);
		expect(totalOf("cap")).toBeLessThanOrEqual(sizeScale.most);
	});

	test("adds the file up from its blocks, in whole KB", () => {
		expect(totalOf("typical")).toBe(61);
		expect(totalOf("cap")).toBe(79);
	});

	test("names among the writers of a block the service's tools alone", () => {
		for (const block of blocks) {
			for (const tool of block.writers) {
				expect(tools, block.id).toContain(tool);
			}
		}
	});

	test("draws a size as its share of the bar", () => {
		expect(shareOf(sizeScale.goal)).toBe("80%");
		expect(shareOf(27)).toBe("33.75%");
		expect(shareOf(0.6)).toBe("0.75%");
	});

	test("keeps the profile's file in the parent's Drive, by the name a store has for it", () => {
		expect(
			stores.flatMap(({ rows }) => rows.flatMap(({ code }) => code ?? [])),
		).toEqual(["mathtrail-profile.json"]);
	});
});

describe("the lifetimes", () => {
	test("puts a minute at the start of the scale, half a year at its end, and an hour, a day and a week where the draft does", () => {
		expect(lifeScale.ticks.map(logShare)).toEqual([
			0, 32.85, 58.34, 73.95, 100,
		]);
	});

	test("keeps what lives longer than the scale at its end, and what lives less at its start", () => {
		expect(logShare(lifeScale.to * 2)).toBe(100);
		expect(logShare(1)).toBe(0);
	});

	test("reaches a token that slides as far as it can be renewed, and anything else as far as it lives", () => {
		expect(lifetimes.map(reachOf)).toEqual([
			{ lives: 0, more: 0 },
			{ lives: 18.47, more: 0 },
			{ lives: 21.72, more: 0 },
			{ lives: 85.63, more: 8.81 },
			{ lives: 94.44, more: 0 },
			{ lives: 100, more: 0 },
		]);
	});
});
