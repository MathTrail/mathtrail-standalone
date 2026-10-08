import { rule } from "./headrules";

/**
 * serviceSections are the ids of the parts of the page "Service", in the
 * order the page draws them: the map of the service's parts, a lesson call by
 * call, where the state lives, the profile's file, and what is sealed.
 */
export const serviceSections = [
	"map",
	"flow",
	"state",
	"profile",
	"seal",
] as const;

/** SectionId is the id of one part of the page. */
export type SectionId = (typeof serviceSections)[number];

/** titleId is the id of the heading of a part of the page. */
export const titleId = (section: SectionId): string => `${section}-title`;

/**
 * Group is parts of the map drawn together under one label, in the order
 * they stand.
 */
export type Group = { readonly id: string; readonly parts: readonly string[] };

/**
 * Column is a column of the map: the family's chat, the service itself, or
 * what stands outside it; tagged when its heading carries a note beside it.
 */
export type Column = {
	readonly id: "family" | "service" | "outside";
	readonly tagged: boolean;
	readonly groups: readonly Group[];
};

/**
 * columns are the map of the service, column by column: who talks to it, its
 * own parts, and the services outside its binary. Each part's id names its
 * words, its choice on the map and its details under the map.
 */
export const columns: readonly Column[] = [
	{
		id: "family",
		tagged: false,
		groups: [
			{ id: "people", parts: ["adult", "kid"] },
			{ id: "host", parts: ["model", "appsrt", "cimd"] },
		],
	},
	{
		id: "service",
		tagged: true,
		groups: [
			{ id: "entry", parts: ["router", "authsrv", "mcpsrv", "limits"] },
			{ id: "tools", parts: ["tools"] },
			{ id: "core", parts: ["rule", "rating", "checks", "solver", "content"] },
			{ id: "around", parts: ["widget", "store", "seal", "logs", "boot"] },
		],
	},
	{
		id: "outside",
		tagged: false,
		groups: [
			{ id: "account", parts: ["goauth", "gdrive"] },
			{ id: "cloud", parts: ["sm", "ar", "clog", "bq"] },
			{ id: "github", parts: ["ga", "pages"] },
		],
	},
];

// groups are the map's groups of parts, column by column.
const groups: readonly Group[] = columns.flatMap((column) => column.groups);

/** parts are the ids of every part of the map, in the order it draws them. */
export const parts: readonly string[] = groups.flatMap((group) => group.parts);

/** groupOf is the id of the group part stands in. */
export function groupOf(part: string): string {
	const group = groups.find((candidate) => candidate.parts.includes(part));
	if (group === undefined) {
		throw new Error(`the map of the service has no part ${part}`);
	}
	return group.id;
}

/**
 * hosts are the chats the first screen names by name; wires are the
 * protocols it names between its tiers: the chat's with the service, and the
 * service's with the parent's Drive.
 */
export const hosts: readonly string[] = ["Claude", "ChatGPT"];
export const wires = { chat: "MCP", drive: "Drive API v3" } as const;

/**
 * toolsPart is the part of the map that lists the service's tools by name;
 * coreParts are the parts of the map's group of pure computation, which the
 * first screen names as the core of the service.
 */
export const toolsPart = "tools";
export const coreParts: readonly string[] = partsOf("core");

/** partsOf are the parts of the map's group id, in the order they stand. */
export function partsOf(id: string): readonly string[] {
	const group = groups.find((candidate) => candidate.id === id);
	if (group === undefined) {
		throw new Error(`the map of the service has no group ${id}`);
	}
	return group.parts;
}

/** tools are the names of the service's tools, as a chat calls them. */
export const tools: readonly string[] = [
	"next_task",
	"prepare_task",
	"get_package",
	"submit_task",
	"read_task",
	"submit_answer",
	"get_progress",
	"read_progress",
	"get_profile",
	"save_profile",
	"edit_profile",
];

/**
 * Link is two parts of the map that talk to each other, the first the one
 * that starts the talk.
 */
export type Link = readonly [from: string, to: string];

/** links are every two parts of the map that talk, each pair once. */
export const links: readonly Link[] = [
	["adult", "model"],
	["kid", "appsrt"],
	["adult", "authsrv"],
	["model", "router"],
	["appsrt", "router"],
	["mcpsrv", "appsrt"],
	["authsrv", "cimd"],
	["router", "authsrv"],
	["router", "mcpsrv"],
	["router", "limits"],
	["mcpsrv", "tools"],
	["widget", "mcpsrv"],
	["content", "mcpsrv"],
	["tools", "rule"],
	["tools", "rating"],
	["tools", "checks"],
	["tools", "store"],
	["tools", "content"],
	["tools", "limits"],
	["tools", "seal"],
	["rule", "content"],
	["checks", "solver"],
	["checks", "content"],
	["authsrv", "seal"],
	["boot", "router"],
	["tools", "logs"],
	["authsrv", "logs"],
	["store", "logs"],
	["limits", "logs"],
	["store", "gdrive"],
	["authsrv", "goauth"],
	["sm", "boot"],
	["sm", "seal"],
	["ar", "boot"],
	["logs", "clog"],
	["clog", "bq"],
	["ga", "ar"],
	["ga", "pages"],
	["authsrv", "pages"],
	["appsrt", "pages"],
];

/** linkKey is the name a link's words go by: its two parts, from-to. */
export function linkKey([from, to]: Link): string {
	return `${from}-${to}`;
}

/**
 * neighbours are the parts part talks to, each with the link between them, in
 * the order of the links.
 */
export function neighbours(
	part: string,
): { readonly part: string; readonly link: Link }[] {
	return links.flatMap((link) => {
		const [from, to] = link;
		if (from === part) {
			return [{ part: to, link }];
		}
		return to === part ? [{ part: from, link }] : [];
	});
}

/**
 * quickPicks are the parts the map offers to choose before any is chosen:
 * the tools, the seal and the store, where most of what the page tells meets.
 */
export const quickPicks: readonly string[] = ["tools", "seal", "store"];

/**
 * ActorKind is who an actor of a lesson is: a person, the host's side, the
 * service, or Google's side.
 */
export type ActorKind = "person" | "host" | "service" | "google";

/** Actor is one column of a scenario: who calls and answers there. */
export type Actor = { readonly id: string; readonly kind: ActorKind };

/**
 * Step is one row of a scenario, between two of its actors, counted from 1:
 * a call, its answer, a message inside the host, or a note across the actors
 * from one to another, in the host's colours or the service's.
 */
export type Step =
	| {
			readonly kind: "call" | "reply" | "host";
			readonly from: number;
			readonly to: number;
	  }
	| {
			readonly kind: "note";
			readonly from: number;
			readonly to: number;
			readonly tone: "host" | "service";
	  };

/** Scenario is one way a lesson goes, its actors and its steps in order. */
export type Scenario = {
	readonly id: "task" | "answer" | "signin";
	readonly actors: readonly Actor[];
	readonly steps: readonly Step[];
};

// asking are the actors of a task asked for: the adult, who asks for it in the
// chat, the card, the host's model, the service and the parent's Drive.
const asking: readonly Actor[] = [
	{ id: "adult", kind: "person" },
	{ id: "card", kind: "host" },
	{ id: "model", kind: "host" },
	{ id: "service", kind: "service" },
	{ id: "drive", kind: "google" },
];

// answering are the actors of an answer: the adult, who asks about it in the
// chat, and the child beside them, who answers on the card, then the card,
// the host's model, the service and the parent's Drive.
const answering: readonly Actor[] = [
	{ id: "adult", kind: "person" },
	{ id: "kid", kind: "person" },
	{ id: "card", kind: "host" },
	{ id: "model", kind: "host" },
	{ id: "service", kind: "service" },
	{ id: "drive", kind: "google" },
];

// call, reply, inHost and note are a step of each kind, between two actors.
const call = (from: number, to: number): Step => ({ kind: "call", from, to });
const reply = (from: number, to: number): Step => ({ kind: "reply", from, to });
const inHost = (from: number, to: number): Step => ({ kind: "host", from, to });
const note = (from: number, to: number, tone: "host" | "service"): Step => ({
	kind: "note",
	from,
	to,
	tone,
});

/**
 * scenarios are a lesson's three ways, call by call: a task, from the adult's
 * asking to the card that shows it; an answer, from the child's press to the
 * chat's explaining to the adult; and the parent's one sign-in.
 */
export const scenarios: readonly Scenario[] = [
	{
		id: "task",
		actors: asking,
		steps: [
			call(1, 3),
			call(3, 4),
			note(1, 3, "host"),
			note(4, 5, "service"),
			call(4, 5),
			reply(4, 3),
			call(3, 4),
			reply(4, 3),
			note(2, 4, "host"),
			call(3, 4),
			note(4, 5, "service"),
			call(4, 5),
			call(2, 4),
			reply(4, 2),
		],
	},
	{
		id: "answer",
		actors: answering,
		steps: [
			call(2, 3),
			call(3, 5),
			note(5, 6, "service"),
			call(5, 6),
			reply(5, 3),
			inHost(3, 4),
			call(1, 4),
			reply(4, 1),
		],
	},
	{
		id: "signin",
		actors: [
			{ id: "browser", kind: "person" },
			{ id: "host", kind: "host" },
			{ id: "service", kind: "service" },
			{ id: "google", kind: "google" },
			{ id: "drive", kind: "google" },
		],
		steps: [
			call(2, 3),
			reply(3, 2),
			call(2, 1),
			call(1, 3),
			note(2, 4, "service"),
			call(1, 4),
			reply(4, 1),
			call(1, 3),
			call(3, 4),
			call(3, 5),
			note(3, 5, "service"),
			reply(3, 1),
			call(1, 2),
			call(2, 3),
			reply(3, 2),
		],
	},
];

/**
 * gridColumn is where a step stands in its scenario's grid of two columns an
 * actor: an arrow from the middle of one actor to the middle of the other, a
 * note across both of them whole.
 */
export function gridColumn({ kind, from, to }: Step): string {
	if (kind === "note") {
		return `${2 * from - 1} / ${2 * to + 1}`;
	}
	return `${2 * Math.min(from, to)} / ${2 * Math.max(from, to)}`;
}

/**
 * numberedSteps are a scenario's steps in order, each with the number it is
 * shown under: the arrows counted from 1, and 0 for a note, which shows none.
 */
export function numberedSteps(
	steps: readonly Step[],
): { readonly step: Step; readonly number: number }[] {
	let arrows = 0;
	return steps.map((step) => {
		if (step.kind === "note") {
			return { step, number: 0 };
		}
		arrows += 1;
		return { step, number: arrows };
	});
}

/** TaskState is a state of the task on the card, as the profile keeps it. */
export type TaskState = "none" | "requested" | "issued" | "answered";

/**
 * Forward is a step of the task on the card on to its next state: the call
 * that takes it there, ticked where the task has passed its checks to get
 * there, and the state it reaches.
 */
export type Forward = {
	readonly tool: string;
	readonly checked: boolean;
	readonly to: TaskState;
};

/**
 * firstState is the state the task on the card starts from, and forward the
 * steps that take it on, in the order a task goes through them; taskStates
 * are the states that order passes.
 */
export const firstState: TaskState = "none";
export const forward: readonly Forward[] = [
	{ tool: "next_task", checked: false, to: "requested" },
	{ tool: "submit_task", checked: true, to: "issued" },
	{ tool: "submit_answer", checked: false, to: "answered" },
];
export const taskStates: readonly TaskState[] = [
	firstState,
	...forward.map(({ to }) => to),
];

/**
 * Turn is a way the task on the card goes other than forward: from a state to
 * another or back to itself, in the colour of the state it starts from or in
 * grey.
 */
export type Turn = {
	readonly from: TaskState;
	readonly to: TaskState;
	readonly tone: "requested" | "issued" | "answered" | "muted";
};

/** turns are the ways the task on the card goes other than forward. */
export const turns: readonly Turn[] = [
	{ from: "requested", to: "requested", tone: "requested" },
	{ from: "requested", to: "none", tone: "muted" },
	{ from: "issued", to: "requested", tone: "issued" },
	{ from: "answered", to: "answered", tone: "answered" },
	{ from: "answered", to: "requested", tone: "answered" },
	{ from: "answered", to: "issued", tone: "answered" },
	{ from: "issued", to: "none", tone: "issued" },
];

/** turnName is how the page writes a turn: from → to, or from ↺. */
export function turnName({ from, to }: Turn): string {
	return from === to ? `${from} ↺` : `${from} → ${to}`;
}

/**
 * Store is a place the service's state lives in when the service keeps none:
 * its rows, each with a name in code when it is a file's, in a card of the
 * tone of whoever keeps it.
 */
export type Store = {
	readonly id: string;
	readonly tone: "tint" | "paper" | "outline";
	readonly rows: readonly { readonly id: string; readonly code?: string }[];
};

/**
 * profileFile is the profile's file in the parent's Drive, by its folder and
 * by its name.
 */
export const profileFile = {
	folder: "MathTrail",
	name: "mathtrail-profile.json",
} as const;

/**
 * stores are where the state lives that a server with state would keep: the
 * chat's tokens, the parent's browser, the parent's Drive, the secrets, and
 * the instance's memory, which keeps caches alone.
 */
export const stores: readonly Store[] = [
	{ id: "tokens", tone: "tint", rows: [{ id: "access" }, { id: "refresh" }] },
	{ id: "browser", tone: "tint", rows: [{ id: "approvals" }, { id: "csrf" }] },
	{
		id: "drive",
		tone: "paper",
		rows: [{ id: "profile", code: profileFile.name }],
	},
	{ id: "secrets", tone: "paper", rows: [{ id: "keys" }, { id: "client" }] },
	{
		id: "memory",
		tone: "outline",
		rows: [{ id: "rates" }, { id: "location" }, { id: "clients" }],
	},
];

/**
 * Block is a block of the profile's file: its size in KB as a file usually
 * has it and near its caps, the colour of its stretch of the bar — a colour of
 * the site's palette by name — and the tools that write it, or every write.
 */
export type Block = {
	readonly id: string;
	readonly typical: number;
	readonly cap: number;
	readonly tone: string;
	readonly writers: readonly string[];
	readonly everyWrite: boolean;
};

/**
 * blocks are the profile's file block by block, the largest first, with the
 * sizes the design of the file estimates for them.
 */
export const blocks: readonly Block[] = [
	{
		id: "fingerprints",
		typical: 27,
		cap: 29,
		tone: "--s-accent",
		writers: ["next_task", "submit_task"],
		everyWrite: false,
	},
	{
		id: "history",
		typical: 14,
		cap: 15,
		tone: "--s-bar-history",
		writers: ["submit_answer"],
		everyWrite: false,
	},
	{
		id: "ratings",
		typical: 8,
		cap: 14,
		tone: "--s-bar-ratings",
		writers: ["next_task", "submit_task", "submit_answer"],
		everyWrite: false,
	},
	{
		id: "recent",
		typical: 4,
		cap: 5,
		tone: "--s-bar-recent",
		writers: ["next_task", "submit_answer"],
		everyWrite: false,
	},
	{
		id: "sealed",
		typical: 4,
		cap: 8,
		tone: "--s-bar-sealed",
		writers: ["next_task", "submit_task", "submit_answer"],
		everyWrite: false,
	},
	{
		id: "open",
		typical: 2,
		cap: 4,
		tone: "--s-right",
		writers: ["next_task", "submit_task", "submit_answer"],
		everyWrite: false,
	},
	{
		id: "request",
		typical: 1.5,
		cap: 2,
		tone: "--s-bar-request",
		writers: ["next_task", "prepare_task", "submit_task"],
		everyWrite: false,
	},
	{
		id: "service",
		typical: 0.6,
		cap: 1.5,
		tone: "--s-line-strong",
		writers: ["save_profile", "edit_profile"],
		everyWrite: true,
	},
];

/**
 * sizeScale is the bar the blocks stand on: up to most KB, a mark every step
 * KB, and the size the file was meant to keep within.
 */
export const sizeScale = { most: 80, step: 20, goal: 64 } as const;

/** Size is which of a block's two sizes the page shows. */
export type Size = "typical" | "cap";

/** totalOf is the file's size, in whole KB, by the size of its blocks. */
export function totalOf(size: Size): number {
	return Math.round(blocks.reduce((sum, block) => sum + block[size], 0));
}

/** shareOf is how much of the bar KB of the file take, in per cent. */
export function shareOf(kb: number): string {
	return `${round((kb / sizeScale.most) * 100)}%`;
}

/**
 * purposes are the purposes a seal is cut for, by the letter a token carries:
 * a code, an access token, a refresh token, a client, the state of a sign-in,
 * a consent and the answer of a task.
 */
export const purposes: readonly string[] = ["c", "a", "r", "d", "s", "k", "t"];

/**
 * tokenSample is a sealed access token as the page draws it: its version, its
 * purpose, the id of its key and the start of its ciphertext.
 */
export const tokenSample = {
	version: "mt1",
	purpose: "a",
	key: "kQ7fWx",
	sealed: "5rJ0Yb2mQ8vK1nT4pXc9…",
} as const;

const minute = 60;
const hour = 60 * minute;
const day = 24 * hour;
const week = 7 * day;

/**
 * Lifetime is how long something sealed or kept lives, in seconds, and for a
 * token that slides, the most it can be renewed to; a key's lifetime is the
 * rotation it waits for, drawn in grey.
 */
export type Lifetime = {
	readonly id: string;
	readonly seconds: number;
	readonly upTo?: number;
	readonly key?: true;
};

/**
 * lifetimes are how long the sign-in's pieces live, the shortest first, and
 * the key that seals them.
 */
export const lifetimes: readonly Lifetime[] = [
	{ id: "code", seconds: minute },
	{ id: "signin", seconds: 10 * minute },
	{ id: "access", seconds: 15 * minute },
	{ id: "refresh", seconds: 30 * day, upTo: 90 * day },
	{ id: "key", seconds: 90 * day, key: true },
	{ id: "approvals", seconds: 180 * day },
];

/**
 * lifeScale is the scale the lifetimes are laid on, from a minute to half a
 * year, by the logarithm of the time; ticks are the times it marks.
 */
export const lifeScale = {
	from: minute,
	to: 180 * day,
	ticks: [minute, hour, day, week, 180 * day],
} as const;

/**
 * logShare is where seconds stand on the scale of the lifetimes, in per cent:
 * 0 for a minute, 100 for half a year, on the logarithm of the time between.
 */
export function logShare(seconds: number): number {
	const share =
		Math.log(seconds / lifeScale.from) /
		Math.log(lifeScale.to / lifeScale.from);
	return round(Math.min(1, Math.max(0, share)) * 100);
}

/**
 * reachOf is how far a lifetime reaches on its scale, in per cent: as far as
 * it lives, and for a token that slides, how much further it can be renewed.
 */
export function reachOf(life: Lifetime): { lives: number; more: number } {
	const lives = logShare(life.seconds);
	return {
		lives,
		more: life.upTo === undefined ? 0 : round(logShare(life.upTo) - lives),
	};
}

// round is a share in per cent to two places, as a stylesheet is given it.
function round(percent: number): number {
	return Math.round(percent * 100) / 100;
}

/**
 * pickId is the id of the choice of part on the map, and wholeMap the choice
 * of none; aboutId is the id of the part's details under the map.
 */
export const pickId = (part: string): string => `pick-${part}`;
export const wholeMap = "pick-whole";
export const aboutId = (part: string): string => `about-${part}`;

/** faceId is the id of the part on the map, the label of its choice. */
export const faceId = (part: string): string => `face-${part}`;

/**
 * scenarioId is the id of the choice of a scenario, and stepsId the id of the
 * scenario's steps.
 */
export const scenarioId = (scenario: string): string => `scenario-${scenario}`;
export const stepsId = (scenario: string): string => `flow-${scenario}`;

/**
 * serviceRules are the rules of style the page's head carries, written from
 * the map's links and parts and from the scenarios when the site is built:
 * with a part chosen on the map, the part filled, the parts it talks to ringed
 * and its details shown under the map, and the part whose choice the keyboard
 * is on outlined; with a scenario chosen, its steps shown. They only ever
 * show, so a browser that cannot read them, which the stylesheet then hides
 * nothing from, shows every part's details and every scenario.
 */
export function serviceRules(): string {
	const map: SectionId = "map";
	const flow: SectionId = "flow";
	const chosen = (part: string) => `#${map}:has(#${pickId(part)}:checked)`;
	const filled = parts.map((part) => `${chosen(part)} #${faceId(part)}`);
	const ringed = links.flatMap(([from, to]) => [
		`${chosen(from)} #${faceId(to)}`,
		`${chosen(to)} #${faceId(from)}`,
	]);
	const outlined = parts.map(
		(part) => `#${map}:has(#${pickId(part)}:focus-visible) #${faceId(part)}`,
	);
	const told = parts.map((part) => `${chosen(part)} #${aboutId(part)}`);
	const shown = scenarios.map(
		({ id }) => `#${flow}:has(#${scenarioId(id)}:checked) #${stepsId(id)}`,
	);
	return [
		rule(
			filled,
			"opacity:1;background:var(--s-accent-tint);box-shadow:0 0 0 2px var(--s-accent)",
		),
		rule(ringed, "opacity:1;box-shadow:0 0 0 1.5px var(--s-accent)"),
		rule(outlined, "outline:2px solid var(--s-accent);outline-offset:2px"),
		rule(told, "display:grid"),
		rule(shown, "display:block"),
	]
		.filter((written) => written !== "")
		.join("\n");
}
