import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	Button,
	type Option,
	OptionList,
	OptionRow,
	type OptionState,
	ReplyField,
} from "./controls";

const root = document.createElement("div");
document.body.append(root);

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

function pressed(element: Element | null): void {
	if (!(element instanceof HTMLElement)) {
		throw new Error("nothing to press");
	}
	act(() => {
		element.focus();
		element.click();
	});
}

describe("an option", () => {
	test.each<[OptionState, string | undefined, string, string | null]>([
		["default", undefined, "B 4", null],
		["selected", "Checking…", "B 4 Checking…", "mt-spin"],
		["correct", "Correct answer", "B 4 Correct answer", "mt-icon"],
		["wrong", "Your answer", "B 4 Your answer", "mt-icon"],
		["muted", undefined, "B 4", null],
	])("in the state %s says %s", (state, status, name, icon) => {
		draw(<OptionRow letter="B" value="4" state={state} status={status} />);

		const row = root.querySelector("button.mt-option");
		expect(row?.getAttribute("data-state")).toBe(state);
		expect(row?.textContent).toBe(name);
		const mark = row?.querySelector(".mt-option-status svg");
		if (icon === null) {
			expect(mark).toBeNull();
		} else {
			expect(mark?.getAttribute("class")).toContain(icon);
			expect(mark?.getAttribute("aria-hidden")).toBe("true");
		}
	});

	test("gives its letter when pressed", () => {
		const chosen = vi.fn();
		draw(<OptionRow letter="C" value="5" onSelect={chosen} />);

		pressed(root.querySelector(".mt-option"));

		expect(chosen).toHaveBeenCalledWith("C");
	});

	test("that is locked ignores a press and keeps the focus", () => {
		const chosen = vi.fn();
		draw(<OptionRow letter="C" value="5" locked onSelect={chosen} />);
		const row = root.querySelector<HTMLButtonElement>(".mt-option");

		pressed(row);

		expect(chosen).not.toHaveBeenCalled();
		expect(row?.getAttribute("aria-disabled")).toBe("true");
		expect(row?.disabled).toBe(false);
		expect(document.activeElement).toBe(row);
	});

	test("is a button, which arrow keys do not answer with", () => {
		draw(<OptionRow letter="A" value="3" />);

		const row = root.querySelector(".mt-option");
		expect(row?.tagName).toBe("BUTTON");
		expect(row?.getAttribute("type")).toBe("button");
		expect(root.querySelector("input[type=radio]")).toBeNull();
	});
});

describe("the options", () => {
	const five: Option[] = ["A", "B", "C", "D", "E"].map((letter, at) => ({
		letter,
		value: String(at + 3),
	}));

	test("stand in their order under their legend", () => {
		draw(<OptionList legend="Pick one answer" options={five} />);

		expect(root.querySelector("fieldset legend")?.textContent).toBe(
			"Pick one answer",
		);
		expect(
			[...root.querySelectorAll(".mt-option-letter")].map(
				(letter) => letter.textContent,
			),
		).toEqual(["A", "B", "C", "D", "E"]);
	});

	test("give their texts in the language they are said in", () => {
		draw(
			<OptionList
				legend="Pick one answer"
				options={five}
				said={{ lang: "ar", dir: "rtl" }}
			/>,
		);

		for (const value of root.querySelectorAll(".mt-option-value")) {
			expect(value.getAttribute("lang")).toBe("ar");
			expect(value.getAttribute("dir")).toBe("rtl");
		}
		expect(root.querySelector("legend")?.hasAttribute("lang")).toBe(false);
	});

	test("that are locked are locked each", () => {
		const chosen = vi.fn();
		draw(
			<OptionList legend="Answers" options={five} locked onSelect={chosen} />,
		);

		for (const row of root.querySelectorAll(".mt-option")) {
			expect(row.getAttribute("aria-disabled")).toBe("true");
			pressed(row);
		}
		expect(chosen).not.toHaveBeenCalled();
	});
});

describe("a button", () => {
	test("is secondary unless it is primary", () => {
		draw(
			<>
				<Button>Hint</Button>
				<Button variant="primary">Another task</Button>
			</>,
		);

		const [secondary, primary] = root.querySelectorAll("button");
		expect(secondary?.className).toBe("mt-btn");
		expect(primary?.className).toBe("mt-btn mt-btn-primary");
	});

	test("says whether what it shows is shown", () => {
		draw(
			<>
				<Button expanded>Hide hint</Button>
				<Button expanded={false}>Hint</Button>
				<Button>I don't know</Button>
			</>,
		);

		expect(
			[...root.querySelectorAll("button")].map((shown) =>
				shown.getAttribute("aria-expanded"),
			),
		).toEqual(["true", "false", null]);
	});

	test("that is disabled or locked does nothing", () => {
		const done = vi.fn();
		draw(
			<>
				<Button disabled onClick={done}>
					Hint
				</Button>
				<Button locked onClick={done}>
					Hint
				</Button>
			</>,
		);

		for (const shown of root.querySelectorAll("button")) {
			pressed(shown);
		}
		expect(done).not.toHaveBeenCalled();
		const [disabled, locked] = root.querySelectorAll("button");
		expect(disabled?.disabled).toBe(true);
		expect(locked?.getAttribute("aria-disabled")).toBe("true");
	});

	test("that is pressed does its thing", () => {
		const done = vi.fn();
		draw(<Button onClick={done}>Hint</Button>);

		pressed(root.querySelector("button"));

		expect(done).toHaveBeenCalledOnce();
	});
});

describe("the question field", () => {
	function field(
		value: string,
		{
			onInput = vi.fn(),
			onSend = vi.fn(),
			disabled = false,
		}: {
			onInput?: (value: string) => void;
			onSend?: (words: string) => void;
			disabled?: boolean;
		} = {},
	) {
		draw(
			<ReplyField
				value={value}
				placeholder="Ask a question about the task…"
				label="Ask a question about the task"
				sendLabel="Send"
				disabled={disabled}
				onInput={onInput}
				onSend={onSend}
			/>,
		);
		const input = root.querySelector("input");
		const send = root.querySelector<HTMLButtonElement>("button");
		if (input === null || send === null) {
			throw new Error("the field is not drawn");
		}
		return { input, send };
	}

	test("hands on each word typed", () => {
		const typed = vi.fn();
		const { input } = field("", { onInput: typed });

		act(() => {
			input.value = "why";
			input.dispatchEvent(new Event("input", { bubbles: true }));
		});

		expect(typed).toHaveBeenCalledWith("why");
	});

	test("sends its words trimmed, with the button or with Enter", () => {
		const sent = vi.fn();
		const { input, send } = field("  why isn't it 6? ", { onSend: sent });

		pressed(send);
		act(() => {
			input.dispatchEvent(
				new KeyboardEvent("keydown", { key: "Enter", bubbles: true }),
			);
		});

		expect(sent.mock.calls).toEqual([["why isn't it 6?"], ["why isn't it 6?"]]);
	});

	test("sends nothing on an Enter that only closes a word being composed", () => {
		const sent = vi.fn();
		const { input } = field("为什么", { onSend: sent });

		act(() => {
			input.dispatchEvent(
				new KeyboardEvent("keydown", {
					key: "Enter",
					isComposing: true,
					bubbles: true,
				}),
			);
		});

		expect(sent).not.toHaveBeenCalled();
	});

	test("sends nothing blank", () => {
		const sent = vi.fn();
		const { input, send } = field("   ", { onSend: sent });

		act(() => {
			input.dispatchEvent(
				new KeyboardEvent("keydown", { key: "Enter", bubbles: true }),
			);
		});

		expect(send.disabled).toBe(true);
		expect(sent).not.toHaveBeenCalled();
	});

	test("sends nothing while it is disabled", () => {
		const sent = vi.fn();
		const { input, send } = field("why", { onSend: sent, disabled: true });

		act(() => {
			input.dispatchEvent(
				new KeyboardEvent("keydown", { key: "Enter", bubbles: true }),
			);
		});

		expect(input.disabled).toBe(true);
		expect(send.disabled).toBe(true);
		expect(sent).not.toHaveBeenCalled();
	});

	test("is named for a screen reader and its button too", () => {
		const { input, send } = field("");

		const label = root.querySelector("label");
		expect(label?.getAttribute("for")).toBe(input.id);
		expect(label?.textContent).toBe("Ask a question about the task");
		expect(send.getAttribute("aria-label")).toBe("Send");
	});
});
