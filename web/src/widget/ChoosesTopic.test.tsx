import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import type { Host } from "./bridge";
import { cardWords } from "./dictionaries";
import { type HandedTask, readHandedTask } from "./payload";
import { type Service, ServiceContext } from "./service";
import { TaskCard } from "./TaskCard";
import { ChoosesTopic } from "./TopicChoice";
import { press } from "./testing/card";
import { fence, withTopicChoice } from "./testing/lesson";
import { WordsContext } from "./words";

afterEach(() => {
	for (const root of document.body.children) {
		act(() => render(null, root));
	}
	document.body.innerHTML = "";
});

// Heard is what a card did with the world: the changes it asked the service
// to save, and the messages it put in the chat.
type Heard = { saved: Record<string, unknown>[]; sent: string[] };

// givenOnTheChoice is the fence handed out on the topic chosen: its card marks
// the topic above the task, with a cross that gives the choice back.
function givenOnTheChoice(): HandedTask {
	const handed = readHandedTask(
		withTopicChoice(fence, { chosen: fence.task.topic }),
	);
	if (handed === undefined) {
		throw new Error("the example is no task handed out");
	}
	return handed;
}

// drawn draws the card of handed where a choice of the topic can be acted on,
// or not, as chooses says, and returns its root and what it did.
function drawn(handed: HandedTask, chooses: boolean) {
	const heard: Heard = { saved: [], sent: [] };
	const service: Service = {
		recordAnswer: () => Promise.resolve({ kind: "failed" }),
		taskStatus: () => Promise.resolve({ kind: "unknown" }),
		readProgress: () => Promise.resolve({ kind: "failed" }),
		saveEdit: (changes) => {
			heard.saved.push(changes);
			return Promise.resolve({ kind: "failed" });
		},
	};
	const host: Host = {
		callTool: () => Promise.reject(new Error("the test calls no tool")),
		sendMessage: (text) => {
			heard.sent.push(text);
			return Promise.resolve();
		},
		tellModel: () => Promise.resolve(),
		canOpenLinks: () => false,
		openLink: () => Promise.resolve(false),
	};
	const root = document.createElement("div");
	document.body.append(root);
	act(() =>
		render(
			<WordsContext.Provider value={cardWords("en", undefined)}>
				<ServiceContext.Provider value={service}>
					<ChoosesTopic.Provider value={chooses}>
						<TaskCard handed={handed} host={host} />
					</ChoosesTopic.Provider>
				</ServiceContext.Provider>
			</WordsContext.Provider>,
			root,
		),
	);
	return { root, heard };
}

// cross is the cross of the mark of the topic chosen above a card's task.
function cross(root: HTMLElement): HTMLButtonElement {
	const found = root.querySelector<HTMLButtonElement>(
		".mt-topic-chip .mt-chip-remove",
	);
	if (found === null) {
		throw new Error("the card marks no topic chosen");
	}
	return found;
}

describe("a card drawn where a choice of the topic cannot be acted on", () => {
	test("shows the button of the topic locked, and opens no choice", () => {
		const { root } = drawn(givenOnTheChoice(), false);
		const button = root.querySelector<HTMLButtonElement>(".mt-topic-button");
		if (button === null) {
			throw new Error("the card has no button of the topic");
		}

		press(button);

		expect(button.getAttribute("aria-disabled")).toBe("true");
		expect(button.getAttribute("aria-expanded")).toBe("false");
	});

	test("turns away the cross of the topic chosen: nothing saved, nothing asked", async () => {
		const { root, heard } = drawn(givenOnTheChoice(), false);

		press(cross(root));
		await act(async () => {});

		expect(heard).toEqual({ saved: [], sent: [] });
	});
});

test("a card in a chat gives the choice back to the coach when its cross is pressed", async () => {
	const { root, heard } = drawn(givenOnTheChoice(), true);

	press(cross(root));
	await act(async () => {});

	expect(heard.saved).toEqual([{ lesson_topic: "" }]);
});
