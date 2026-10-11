import type { CallToolResult } from "@modelcontextprotocol/client";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { lessonLanguages } from "./dictionaries";
import {
	buttonIn,
	type Drawn,
	drawCard,
	foldIn,
	press,
	takeDown,
	unfold,
} from "./testing/card";
import type { ToolCall } from "./testing/host";
import {
	editGone,
	editRefused,
	editSaved,
	failure,
	profileRefused,
	savedWords,
	standing,
} from "./testing/lesson";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
	vi.restoreAllMocks();
});

// opened draws the progress of payload, answered by tools, and opens the form
// in its profile's section, as the parent does. The profile's own card, of an
// earlier chat, folds nothing away, and its form is opened straight from it.
async function opened(
	tools: (call: ToolCall) => CallToolResult | Promise<CallToolResult> = () => {
		throw new Error("no tool is answered here");
	},
	options: Omit<Parameters<typeof drawCard>[1] & object, "tools"> = {},
	payload: object = standing,
): Promise<Drawn> {
	drawn = await drawCard(payload, { tools, ...options });
	if (drawn.root.querySelector(".mt-fold") !== null) {
		unfold(drawn.root, "Profile");
	}
	press(buttonIn(drawn.root, "Edit"));
	await vi.waitFor(() =>
		expect(drawn?.root.querySelector(".mt-form")).not.toBeNull(),
	);
	return drawn;
}

// form is the form of the card.
function form(root: HTMLElement): HTMLFormElement {
	const found = root.querySelector<HTMLFormElement>(".mt-form");
	if (found === null) {
		throw new Error("the card shows no form");
	}
	return found;
}

// field is the box or list of the form named label.
function field<E extends HTMLElement>(root: HTMLElement, label: string): E {
	const named = [...form(root).querySelectorAll("label")].find(
		(found) => found.textContent === label,
	);
	const control = named?.htmlFor
		? root.ownerDocument.getElementById(named.htmlFor)
		: null;
	if (control === null) {
		throw new Error(`the form has no field ${label}`);
	}
	return control as E;
}

// typed types text into the box named label, as a person does.
function typed(root: HTMLElement, label: string, text: string): void {
	const box = field<HTMLInputElement>(root, label);
	act(() => {
		box.value = text;
		box.dispatchEvent(new Event("input", { bubbles: true }));
	});
}

// chosen chooses value in the list named label, as a person does.
function chosen(root: HTMLElement, label: string, value: string): void {
	const list = field<HTMLSelectElement>(root, label);
	act(() => {
		list.value = value;
		list.dispatchEvent(new Event("change", { bubbles: true }));
	});
}

// ticked presses the tick named label: a skill's, or another of the form's.
function ticked(root: HTMLElement, label: string): void {
	const tick = [...form(root).querySelectorAll(".mt-check")].find(
		(found) => found.textContent === label,
	);
	act(() => tick?.querySelector<HTMLElement>("input")?.click());
}

// fields are the form's fields, locked together while it is on its way.
function fields(root: HTMLElement): HTMLFieldSetElement {
	const found =
		form(root).querySelector<HTMLFieldSetElement>(".mt-form-fields");
	if (found === null) {
		throw new Error("the form has no fields");
	}
	return found;
}

// problems are what the form says under each field that has something wrong,
// by the field's name.
function problems(root: HTMLElement): Record<string, string> {
	return Object.fromEntries(
		[...form(root).querySelectorAll(".mt-form-field")].flatMap((group) => {
			const problem = group.querySelector(".mt-field-problem")?.textContent;
			const name = group.querySelector(".mt-form-label")?.textContent ?? "";
			return problem === undefined || problem === null ? [] : [[name, problem]];
		}),
	);
}

const note = (root: HTMLElement) =>
	root.querySelector(".mt-form .mt-action-note")?.textContent;

const saved = (root: HTMLElement) =>
	root.querySelector(".mt-fields .mt-action-note")?.textContent;

describe("the form of the profile", () => {
	test("opens in place of the profile, filled in from it, with the focus on the pseudonym", async () => {
		const { root, heard } = await opened();

		expect(form(root).getAttribute("aria-label")).toBe("Change the profile");
		expect(field<HTMLInputElement>(root, "Pseudonym").value).toBe("Comet");
		expect(document.activeElement).toBe(field(root, "Pseudonym"));
		expect(field<HTMLSelectElement>(root, "Grade").value).toBe("3");
		expect(
			[...form(root).querySelectorAll(".mt-chip-text")].map(
				(chip) => chip.textContent,
			),
		).toEqual(["space", "animals", "football"]);
		expect(
			[...form(root).querySelectorAll(".mt-check")]
				.filter(
					(tick) => tick.querySelector<HTMLInputElement>("input")?.checked,
				)
				.map((tick) => tick.textContent),
		).toEqual(["Division with a remainder"]);
		expect(form(root).querySelectorAll(".mt-checks .mt-check")).toHaveLength(
			25,
		);
		expect(
			field<HTMLSelectElement>(root, "Language of the lessons").value,
		).toBe("");
		expect(form(root).closest(".mt-fields")?.querySelector("dl")).toBeNull();
		expect(heard.calls).toEqual([]);
	});

	// The notes are the parent's words to the model, said in the chat.
	test("has no field for the parent's notes", async () => {
		const { root } = await opened();

		expect(form(root).querySelector("textarea")).toBeNull();
		expect(form(root).querySelectorAll("input[type=text]")).toHaveLength(2);
	});

	test("offers every language of the lessons by name, after the chat's", async () => {
		const { root } = await opened();

		const choices = [
			...field<HTMLSelectElement>(root, "Language of the lessons").options,
		];
		expect(choices[0]?.textContent).toBe("The chat's language");
		expect(choices).toHaveLength(lessonLanguages.length + 1);
		expect(choices.map((choice) => choice.textContent)).toContain("French");
	});

	// Some two hundred and fifty countries are named and put in order for the
	// list, which a phone feels when it is done at every key typed.
	test("names the countries once, and not again at every key typed", async () => {
		const names = vi.spyOn(Intl, "DisplayNames");
		const { root } = await opened();
		const countriesNamed = () =>
			names.mock.calls.filter(([, options]) => options?.type === "region")
				.length;
		const once = countriesNamed();

		for (const pseudonym of ["N", "No", "Nov", "Nova"]) {
			typed(root, "Pseudonym", pseudonym);
		}

		expect(once).toBeGreaterThan(0);
		expect(countriesNamed()).toBe(once);
	});

	test("shows a language the chat chose that no card speaks, as it is kept, and keeps it to choose again", async () => {
		const { root } = await opened(
			undefined,
			{},
			{
				...standing,
				profile: { ...standing.profile, ui_language: "so" },
			},
		);
		const language = () =>
			field<HTMLSelectElement>(root, "Language of the lessons");

		expect(language().value).toBe("so");
		chosen(root, "Language of the lessons", "fr");

		expect(language().value).toBe("fr");
		expect([...language().options].map((choice) => choice.value)).toContain(
			"so",
		);
	});

	test("shows a skill left out that the catalog no longer has, by its id, and keeps it once unticked", async () => {
		const { root } = await opened(
			undefined,
			{},
			{
				...standing,
				profile: { ...standing.profile, excluded_skills: ["juggling"] },
			},
		);

		ticked(root, "juggling");

		const juggling = [...form(root).querySelectorAll(".mt-check")].find(
			(tick) => tick.textContent === "juggling",
		);
		expect(juggling?.querySelector<HTMLInputElement>("input")?.checked).toBe(
			false,
		);
		expect(form(root).querySelectorAll(".mt-checks .mt-check")).toHaveLength(
			26,
		);
	});

	test("saves only what changed, by the card's own tool, shows it at once and tells the model", async () => {
		const { root, heard } = await opened(() =>
			editSaved({ pseudonym: "Nova", grade: 4 }),
		);

		typed(root, "Pseudonym", "Nova");
		chosen(root, "Grade", "4");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(saved(root)).toBe("Saved."));
		expect(heard.calls).toEqual([
			{ name: "edit_profile", arguments: { pseudonym: "Nova", grade: 4 } },
		]);
		expect(root.querySelector(".mt-form")).toBeNull();
		expect(root.querySelector(".mt-head .mt-name")?.textContent).toBe("Nova");
		expect(root.querySelector(".mt-head .mt-badge")?.textContent).toBe(
			"Grade 4",
		);
		await vi.waitFor(() => expect(heard.modelLines).toEqual([savedWords]));
		expect(heard.messages).toEqual([]);
		// The focus goes back once the card has drawn itself without the form,
		// which need not have come before the model was told.
		await vi.waitFor(() =>
			expect(document.activeElement).toBe(buttonIn(root, "Edit")),
		);
	});

	test("says a new language comes with the next task, and the card goes on in its own words", async () => {
		const { root, heard } = await opened(() =>
			editSaved({ ui_language: "fr" }),
		);

		chosen(root, "Language of the lessons", "fr");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() =>
			expect(saved(root)).toBe(
				"Saved. The new language starts with the next task.",
			),
		);
		expect(heard.calls[0]?.arguments).toEqual({ ui_language: "fr" });
		expect(buttonIn(root, "Edit")).toBeDefined();
	});

	test("adds an interest with its button or with Enter, takes one out, and sends them in their order", async () => {
		const { root, heard } = await opened(() => editSaved());

		typed(root, "A new interest", "chess");
		act(() => {
			field(root, "A new interest").dispatchEvent(
				new KeyboardEvent("keydown", { key: "Enter", bubbles: true }),
			);
		});
		typed(root, "A new interest", "maps");
		press(buttonIn(root, "Add"));
		press(buttonIn(root, "Remove animals"));
		typed(root, "A new interest", "kites");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		// An interest typed and not yet added is sent with the rest.
		expect(heard.calls[0]?.arguments).toEqual({
			interests: ["space", "football", "chess", "maps", "kites"],
		});
	});

	test("ticks and unticks the skills left out, and sends them", async () => {
		const { root, heard } = await opened(() => editSaved());

		ticked(root, "Division with a remainder");
		ticked(root, "Simple fractions");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(heard.calls[0]?.arguments).toEqual({
			excluded_skills: ["fractions"],
		});
	});

	test("leaves the country of the sign-in out of the counts when ticked, and sends it", async () => {
		const { root, heard } = await opened(() => editSaved());

		ticked(root, "Don't count the country I sign in from");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(heard.calls[0]?.arguments).toEqual({ signin_country_off: true });
	});

	// The places the form draws on list the states in the order of their
	// codes, which is not the order of their names: Alaska's code comes before
	// Alabama's. The form offers them by their names.
	test("asks for a state once the United States is chosen, none first and then the states by their names, and sends the one chosen with it", async () => {
		const { root, heard } = await opened(() =>
			editSaved({ country: "US", region: "US-CA" }),
		);
		expect(() => field(root, "State")).toThrow("no field State");

		chosen(root, "Country", "US");

		const states = [...field<HTMLSelectElement>(root, "State").options];
		expect(field<HTMLSelectElement>(root, "State").value).toBe("");
		expect(states).toHaveLength(52);
		expect(
			states.slice(0, 5).map((state) => [state.value, state.textContent]),
		).toEqual([
			["", "Not set"],
			["US-AL", "Alabama"],
			["US-AK", "Alaska"],
			["US-AZ", "Arizona"],
			["US-AR", "Arkansas"],
		]);

		chosen(root, "State", "US-CA");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(heard.calls[0]?.arguments).toEqual({
			country: "US",
			region: "US-CA",
		});
	});

	// A family that moves leaves its old state behind, as the service does: a
	// state of the United States says nothing of a family in France.
	test("leaves the state behind when another country is chosen, and asks for none there", async () => {
		const { root, heard } = await opened(
			() => editSaved({ country: "FR", region: null }),
			{},
			{
				...standing,
				profile: { ...standing.profile, country: "US", region: "US-TX" },
			},
		);
		expect(field<HTMLSelectElement>(root, "State").value).toBe("US-TX");

		chosen(root, "Country", "FR");

		expect(() => field(root, "State")).toThrow("no field State");
		press(buttonIn(root, "Save"));
		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(heard.calls[0]?.arguments).toEqual({ country: "FR", region: "" });
	});

	// A page's form, submitted, loads another page in its place: in a chat's
	// frame that would put something else where the card was. The form is
	// sent by its Save button alone.
	test("is never submitted as a page's form is, and sends nothing when something submits it", async () => {
		const { root, heard } = await opened();
		const submitted = new Event("submit", { bubbles: true, cancelable: true });

		act(() => {
			form(root).dispatchEvent(submitted);
		});

		expect(submitted.defaultPrevented).toBe(true);
		expect(heard.calls).toEqual([]);
		expect(root.querySelector(".mt-form")).not.toBeNull();
	});

	test("keeps what was typed while its section is folded and opened again", async () => {
		const { root } = await opened();

		typed(root, "Pseudonym", "Nova");
		press(foldIn(root, "Profile"));

		expect(form(root).closest("[hidden]")).not.toBeNull();

		press(foldIn(root, "Profile"));

		expect(form(root).closest("[hidden]")).toBeNull();
		expect(field<HTMLInputElement>(root, "Pseudonym").value).toBe("Nova");
	});

	test("says what became of a change saved while its section was folded, once it is opened again", async () => {
		const { root, heard } = await opened(() =>
			editSaved({ pseudonym: "Nova" }),
		);

		typed(root, "Pseudonym", "Nova");
		press(buttonIn(root, "Save"));
		press(foldIn(root, "Profile"));
		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		await vi.waitFor(() => expect(root.querySelector(".mt-form")).toBeNull());
		press(foldIn(root, "Profile"));

		expect(saved(root)).toBe("Saved.");
	});

	test("with nothing changed closes, and sends nothing", async () => {
		const { root, heard } = await opened();

		press(buttonIn(root, "Save"));

		expect(root.querySelector(".mt-form")).toBeNull();
		expect(saved(root)).toBe("");
		expect(heard.calls).toEqual([]);
	});

	test("cancelled closes, and keeps the profile as it was", async () => {
		const { root, heard } = await opened();

		typed(root, "Pseudonym", "Nova");
		press(buttonIn(root, "Cancel"));

		expect(root.querySelector(".mt-form")).toBeNull();
		expect(root.querySelector(".mt-head .mt-name")?.textContent).toBe("Comet");
		expect(heard.calls).toEqual([]);
		expect(document.activeElement).toBe(buttonIn(root, "Edit"));
	});

	test("is sent once when Save is pressed twice before the card redraws, and holds while on its way", async () => {
		const { root, heard } = await opened(() => new Promise(() => {}));
		typed(root, "Pseudonym", "Nova");
		const save = buttonIn(root, "Save");

		act(() => {
			save.click();
			save.click();
		});

		await vi.waitFor(() => expect(note(root)).toBe("Saving…"));
		expect(heard.calls).toHaveLength(1);
		expect(save.getAttribute("aria-disabled")).toBe("true");
		expect(buttonIn(root, "Cancel").getAttribute("aria-disabled")).toBe("true");
		expect(form(root).getAttribute("aria-busy")).toBe("true");
		// What is typed while it is on its way would not be what it sends.
		expect(fields(root).disabled).toBe(true);
		press(save);
		press(buttonIn(root, "Cancel"));
		expect(heard.calls).toHaveLength(1);
		expect(root.querySelector(".mt-form")).not.toBeNull();
	});

	test("refused, says under each field what to change, in the card's words, and stays open", async () => {
		const { root, heard } = await opened(() => editRefused);

		typed(root, "Pseudonym", " ");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() =>
			expect(note(root)).toBe("Not saved: see the fields marked."),
		);
		expect(problems(root)).toEqual({
			Pseudonym: "Fill this in.",
			Grade: "Choose one of the values offered.",
			Interests: "One of them is empty or too long.",
			"Not at school yet":
				"One of these is not a skill MathTrail knows: untick it.",
			"Language of the lessons": "Choose a language from the list.",
		});
		const pseudonym = field(root, "Pseudonym");
		expect(pseudonym.getAttribute("aria-invalid")).toBe("true");
		expect(
			pseudonym
				.getAttribute("aria-describedby")
				?.split(" ")
				.map((id) => document.getElementById(id)?.textContent),
		).toContain("Fill this in.");
		// The service's words for the model are not the card's.
		expect(root.textContent).not.toContain("is required");
		expect(heard.modelLines).toEqual([]);
	});

	test("says a rule it has no words for in words that fit any", async () => {
		const { root } = await opened(() => ({
			content: [],
			structuredContent: {
				screen: "profile",
				status: "rejected",
				code: "invalid_profile",
				problems: [{ field: "pseudonym", code: "brand_new", rule: "new" }],
				changed: false,
				profile: standing.profile,
			},
		}));

		typed(root, "Pseudonym", "Nova");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() =>
			expect(problems(root)).toEqual({ Pseudonym: "Check this field." }),
		);
	});

	test("not saved, says so, keeps what was typed, and saves on the next press", async () => {
		const answers = [failure, editSaved({ pseudonym: "Nova" })];
		const { root, heard } = await opened(() => answers.shift() ?? failure);

		typed(root, "Pseudonym", "Nova");
		press(buttonIn(root, "Save"));
		await vi.waitFor(() => expect(note(root)).toBe("Not saved — try again."));
		expect(field<HTMLInputElement>(root, "Pseudonym").value).toBe("Nova");
		expect(fields(root).disabled).toBe(false);

		press(buttonIn(root, "Save"));
		await vi.waitFor(() => expect(saved(root)).toBe("Saved."));
		expect(heard.calls).toHaveLength(2);
	});

	test("that never reaches the service is not saved either", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const { root } = await opened(() => {
			throw new Error("the host lost the call");
		});

		typed(root, "Pseudonym", "Nova");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(note(root)).toBe("Not saved — try again."));
	});

	test("finds the profile gone, says so, and offers no form", async () => {
		const { root, heard } = await opened(() => editGone);

		typed(root, "Pseudonym", "Nova");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-fields-said")?.textContent).toBe(
				"This profile is no longer there. Ask in the chat for a new one.",
			),
		);
		expect(root.querySelector(".mt-form")).toBeNull();
		expect(() => buttonIn(root, "Edit")).toThrow();
		expect(document.activeElement).toBe(root.querySelector(".mt-fields-said"));
		expect(heard.modelLines).toEqual([]);
	});

	test("saved on a host that keeps no line for the model is saved all the same", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const { root, heard } = await opened(() => editSaved({ grade: 5 }), {
			refuseModelLines: true,
		});

		chosen(root, "Grade", "5");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(saved(root)).toBe("Saved."));
		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		expect(console.error).toHaveBeenCalled();
	});

	test("says what became of a save from a place heard once something is said", async () => {
		const { root } = await opened();

		expect(note(root)).toBe("");
		expect(
			root.querySelector(".mt-form .mt-action-note")?.getAttribute("aria-live"),
		).toBe("polite");
	});
});

describe("the profile's card of an earlier chat", () => {
	test("opens the same form, and a save clears the refusal it was drawn with", async () => {
		const { root, heard } = await opened(
			() => editSaved({ pseudonym: "Nova" }),
			{},
			profileRefused,
		);

		typed(root, "Pseudonym", "Nova");
		press(buttonIn(root, "Save"));

		await vi.waitFor(() => expect(saved(root)).toBe("Saved."));
		expect(root.querySelector(".mt-verdict-line")).toBeNull();
		expect(root.querySelector(".mt-head .mt-name")?.textContent).toBe("Nova");
		expect(heard.calls[0]?.name).toBe("edit_profile");
	});
});
