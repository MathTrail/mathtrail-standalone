import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	Button,
	Checkbox,
	CheckGroup,
	ChipsField,
	type Option,
	OptionList,
	OptionRow,
	type OptionState,
	SelectField,
	TextField,
	ViewSwitch,
	withinLength,
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

describe("a checkbox", () => {
	test("is ticked by pressing anywhere on its line, and says so", () => {
		const ticks: boolean[] = [];
		draw(
			<Checkbox
				label="I am the child's parent or tutor"
				checked={false}
				onChange={(checked) => ticks.push(checked)}
			/>,
		);

		const line = root.querySelector("label.mt-check");
		expect(line?.textContent).toBe("I am the child's parent or tutor");
		act(() => {
			line
				?.querySelector("span")
				?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
		});
		expect(ticks).toEqual([true]);
	});
});

// typed types text into a box, as a person does.
function typed(box: Element | null, text: string): void {
	if (!(box instanceof HTMLInputElement)) {
		throw new Error("no box to type into");
	}
	act(() => {
		box.value = text;
		box.dispatchEvent(new Event("input", { bubbles: true }));
	});
}

// keyed presses key in element, as a person does, while a word is composed in
// an input method or not.
function keyed(element: Element | null, key: string, isComposing = false) {
	act(() => {
		element?.dispatchEvent(
			new KeyboardEvent("keydown", { key, isComposing, bubbles: true }),
		);
	});
}

// describedBy are the texts of what an element is described by, in order.
function describedBy(element: Element | null): string[] {
	return (element?.getAttribute("aria-describedby") ?? "")
		.split(" ")
		.filter((id) => id !== "")
		.map((id) => document.getElementById(id)?.textContent ?? "");
}

describe("a text cut to its length", () => {
	test.each([
		["shorter than its most, as it is", "Comet", 32, "Comet"],
		["longer, cut", "abcdef", 4, "abcd"],
		["of emoji, cut by characters", "🐱🐶🐭🐹", 2, "🐱🐶"],
	])("%s", (_, text, most, cut) => {
		expect(withinLength(text, most)).toBe(cut);
	});
});

describe("a text field", () => {
	test("is named by its label, takes no more than its most, and says what is wrong", () => {
		const input = vi.fn();
		draw(
			<TextField
				label="Pseudonym"
				value="Comet"
				most={32}
				note="Never the child's real name."
				problem="Fill this in."
				onInput={input}
			/>,
		);

		const box = root.querySelector("input");
		const label = root.querySelector("label");
		expect(label?.textContent).toBe("Pseudonym");
		expect(label?.getAttribute("for")).toBe(box?.id);
		expect(box?.value).toBe("Comet");
		expect(box?.getAttribute("aria-invalid")).toBe("true");
		expect(describedBy(box)).toEqual([
			"Never the child's real name.",
			"Fill this in.",
		]);
		typed(box, "Nova");
		expect(input).toHaveBeenCalledWith("Nova");
		typed(box, "x".repeat(40));
		expect(input).toHaveBeenLastCalledWith("x".repeat(32));
		// An emoji is one character, as the profile counts it, though a
		// string of the page takes two units for it.
		typed(box, "🐱".repeat(20));
		expect(input).toHaveBeenLastCalledWith("🐱".repeat(20));
	});

	test("with nothing wrong is not marked, and with nothing to say is described by nothing", () => {
		draw(<TextField label="Pseudonym" value="" most={32} onInput={() => {}} />);

		const box = root.querySelector("input");
		expect(box?.hasAttribute("aria-invalid")).toBe(false);
		expect(box?.hasAttribute("aria-describedby")).toBe(false);
		expect(root.querySelector(".mt-field-problem")).toBeNull();
	});

	test("does its thing on Enter, unless Enter only ends a word being composed", () => {
		const enter = vi.fn();
		draw(
			<TextField
				label="Pseudonym"
				value="Comet"
				most={32}
				onInput={() => {}}
				onEnter={enter}
			/>,
		);

		keyed(root.querySelector("input"), "Enter", true);
		expect(enter).not.toHaveBeenCalled();
		keyed(root.querySelector("input"), "Enter");
		expect(enter).toHaveBeenCalledOnce();
	});
});

describe("a list to choose from", () => {
	test("offers its choices in their order under its label, and gives the one chosen", () => {
		const changed = vi.fn();
		draw(
			<SelectField
				label="Grade"
				value="3"
				choices={[
					{ value: "1", label: "1" },
					{ value: "3", label: "3" },
				]}
				problem="Choose one of the values offered."
				onChange={changed}
			/>,
		);

		const list = root.querySelector("select");
		expect(root.querySelector("label")?.getAttribute("for")).toBe(list?.id);
		expect([...(list?.options ?? [])].map((option) => option.value)).toEqual([
			"1",
			"3",
		]);
		expect(list?.value).toBe("3");
		expect(describedBy(list)).toEqual(["Choose one of the values offered."]);
		act(() => {
			if (list !== null) {
				list.value = "1";
				list.dispatchEvent(new Event("change", { bubbles: true }));
			}
		});
		expect(changed).toHaveBeenCalledWith("1");
	});
});

describe("a list of chips", () => {
	// chips draws the list with what a case gives it, and the rest of it
	// answering nothing.
	function chips(
		props: Partial<Parameters<typeof ChipsField>[0]> = {},
	): Parameters<typeof ChipsField>[0] {
		return {
			label: "Interests",
			chips: ["space", "cats"],
			typed: "",
			most: 3,
			longest: 40,
			typeLabel: "A new interest",
			addLabel: "Add",
			removeLabel: (chip: string) => `Remove ${chip}`,
			fullNote: "That is as many interests as a profile holds.",
			onType: () => {},
			onAdd: () => {},
			onRemove: () => {},
			...props,
		};
	}

	test("draws each apart with a button that takes it out, named by it", () => {
		const removed = vi.fn();
		draw(<ChipsField {...chips({ onRemove: removed })} />);

		expect(root.querySelector("legend")?.textContent).toBe("Interests");
		expect(
			[...root.querySelectorAll(".mt-chip-text")].map(
				(chip) => chip.textContent,
			),
		).toEqual(["space", "cats"]);
		pressed(root.querySelector('[aria-label="Remove cats"]'));
		expect(removed).toHaveBeenCalledWith("cats");
	});

	test("adds what is typed with its button or with Enter, and adds nothing blank", () => {
		const added = vi.fn();
		const { rerender } = drawn(<ChipsField {...chips({ onAdd: added })} />);
		const add = [...root.querySelectorAll("button")].find(
			(button) => button.textContent === "Add",
		);
		expect(add?.disabled).toBe(true);
		expect(root.querySelector(".mt-add label")?.textContent).toBe(
			"A new interest",
		);

		rerender(<ChipsField {...chips({ typed: "maps", onAdd: added })} />);
		expect(add?.disabled).toBe(false);
		pressed(add ?? null);
		keyed(root.querySelector(".mt-add input"), "Enter", true);
		keyed(root.querySelector(".mt-add input"), "Enter");
		expect(added).toHaveBeenCalledTimes(2);
	});

	test("keeps the focus where the next one is typed once one is added or taken out", () => {
		const { rerender } = drawn(<ChipsField {...chips({ typed: "maps" })} />);
		const box = root.querySelector<HTMLInputElement>(".mt-add input");

		pressed(
			[...root.querySelectorAll("button")].find(
				(button) => button.textContent === "Add",
			) ?? null,
		);
		rerender(
			<ChipsField {...chips({ chips: ["space", "cats"], typed: "" })} />,
		);
		expect(document.activeElement).toBe(box);

		pressed(root.querySelector('[aria-label="Remove cats"]'));
		rerender(<ChipsField {...chips({ chips: ["space"] })} />);
		expect(document.activeElement).toBe(box);
	});

	test("full once one is added, keeps the focus on the button that takes the last one out", () => {
		const { rerender } = drawn(<ChipsField {...chips({ typed: "maps" })} />);

		keyed(root.querySelector(".mt-add input"), "Enter");
		rerender(<ChipsField {...chips({ chips: ["space", "cats", "maps"] })} />);

		expect(document.activeElement).toBe(
			root.querySelector('[aria-label="Remove maps"]'),
		);
	});

	test("as long as it may be takes no more, and says so", () => {
		draw(<ChipsField {...chips({ chips: ["a", "b", "c"], typed: "d" })} />);

		expect(
			root.querySelector<HTMLInputElement>(".mt-add input")?.disabled,
		).toBe(true);
		expect(describedBy(root.querySelector("fieldset"))).toEqual([
			"That is as many interests as a profile holds.",
		]);
	});
});

describe("a group of ticks", () => {
	test("gives what is ticked, in the order it was ticked, and takes out what is unticked", () => {
		const changed = vi.fn();
		draw(
			<CheckGroup
				legend="Not at school yet"
				choices={[
					{ value: "fractions", label: "Fractions" },
					{ value: "division", label: "Division" },
				]}
				value={["division"]}
				problem="Check this field."
				onChange={changed}
			/>,
		);

		const boxes = [...root.querySelectorAll<HTMLInputElement>("input")];
		expect(boxes.map((box) => box.checked)).toEqual([false, true]);
		expect(describedBy(root.querySelector("fieldset"))).toEqual([
			"Check this field.",
		]);
		act(() => boxes[0]?.click());
		expect(changed).toHaveBeenLastCalledWith(["division", "fractions"]);
		act(() => boxes[1]?.click());
		expect(changed).toHaveBeenLastCalledWith([]);
	});
});

// drawn draws an element and gives a way to draw it again with other props.
function drawn(element: preact.JSX.Element) {
	draw(element);
	return { rerender: (next: preact.JSX.Element) => draw(next) };
}

describe("a switch between views", () => {
	const options = [
		{ value: "last_task", label: "Last task" },
		{ value: "week", label: "Past week" },
	] as const;

	test("is a group of radio buttons under a legend a screen reader alone hears, the view shown chosen", () => {
		draw(
			<ViewSwitch
				legend="Show the change"
				options={options}
				value="week"
				onChange={() => {}}
			/>,
		);

		const group = root.querySelector("fieldset.mt-switch");
		expect(group?.querySelector("legend.mt-vh")?.textContent).toBe(
			"Show the change",
		);
		const radios = [
			...root.querySelectorAll<HTMLInputElement>("input[type=radio]"),
		];
		expect(
			radios.map((radio) => [
				radio.value,
				radio.checked,
				radio.closest("label")?.textContent,
			]),
		).toEqual([
			["last_task", false, "Last task"],
			["week", true, "Past week"],
		]);
		// One name, one group: the keyboard stops at it once.
		expect(new Set(radios.map((radio) => radio.name)).size).toBe(1);
		expect(radios[0]?.name).not.toBe("");
	});

	test("hands on the view chosen", () => {
		const onChange = vi.fn();
		draw(
			<ViewSwitch
				legend="Show the change"
				options={options}
				value="week"
				onChange={onChange}
			/>,
		);

		pressed(root.querySelector('input[value="last_task"]'));

		expect(onChange).toHaveBeenCalledWith("last_task");
	});

	test("names two switches on one page apart, so that each is a group of its own", () => {
		draw(
			<div>
				<ViewSwitch
					legend="One"
					options={options}
					value="week"
					onChange={() => {}}
				/>
				<ViewSwitch
					legend="Two"
					options={options}
					value="week"
					onChange={() => {}}
				/>
			</div>,
		);

		const names = [...root.querySelectorAll("fieldset.mt-switch")].map(
			(group) => group.querySelector("input")?.name,
		);
		expect(new Set(names).size).toBe(2);
	});
});
