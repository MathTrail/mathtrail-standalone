import { afterEach, describe, expect, test, vi } from "vitest";
import {
	buttonIn,
	type Drawn,
	drawCard,
	press,
	takeDown,
} from "./testing/card";
import { profileRead, profileRefused } from "./testing/lesson";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
});

// draw draws the profile's card of payload, in a chat whose language is
// English.
async function draw(payload: object): Promise<Drawn> {
	drawn = await drawCard(payload, { context: { locale: "en-US" } });
	return drawn;
}

// A profile whose cards speak a language the widget has no words for yet:
// the card speaks the chat's, and names the chosen one in it.
const inSwahili = {
	...profileRead,
	profile: { ...profileRead.profile, ui_language: "sw" },
};

// fields are the fields of the group labelled label: each its name, what it
// holds — a list's names joined — and the note under it, when there is one.
function fields(root: HTMLElement, label: string): string[][] {
	const group = [...root.querySelectorAll(".mt-fields")].find(
		(found) =>
			found.querySelector(".mt-fields-head .mt-section-label")?.textContent ===
			label,
	);
	return [...(group?.querySelectorAll("dl > div") ?? [])].map((field) => {
		const told = field.querySelector("dd");
		const names = [...(told?.querySelectorAll(".mt-chip") ?? [])].map(
			(chip) => chip.textContent ?? "",
		);
		const note = told?.querySelector(".mt-field-note")?.textContent;
		return [
			field.querySelector("dt")?.textContent ?? "",
			names.length > 0
				? names.join(", ")
				: (told?.firstChild?.textContent ?? ""),
			...(note === undefined || note === null ? [] : [note]),
		];
	});
}

describe("the profile's card", () => {
	test("shows the child's details, and never the parent's notes", async () => {
		const { root } = await draw(inSwahili);

		expect(root.querySelector("article")?.getAttribute("aria-label")).toBe(
			"Profile",
		);
		expect(root.querySelector(".mt-head .mt-name")?.textContent).toBe("Comet");
		expect(fields(root, "Profile · for the parent")).toEqual([
			["Grade", "3", "Only a label: changing it moves no rating."],
			["Interests", "space, animals, football"],
			["Not at school yet", "Division with a remainder"],
			["Language of the lessons", "Swahili"],
			[
				"Country",
				"Not set",
				"Optional. Used only to count, without names, how many families each country has.",
			],
		]);
		expect(root.textContent).not.toContain("Loses heart");
		expect(root.querySelector(".mt-bar")).toBeNull();
	});

	test("names the country by the card's words for it, and a state of the United States by its name", async () => {
		const { root } = await draw({
			...inSwahili,
			profile: { ...inSwahili.profile, country: "US", region: "US-TX" },
		});

		expect(fields(root, "Profile · for the parent").slice(-2)).toEqual([
			[
				"Country",
				"United States",
				"Optional. Used only to count, without names, how many families each country has.",
			],
			["State", "Texas"],
		]);
	});

	test("says where the profile's file is, and what the parent can do with it, each under the question it answers", async () => {
		const { root } = await draw(inSwahili);

		expect(fields(root, "Your data")).toEqual([
			[
				"Where the profile is",
				"The whole profile is one file in your Google Drive: mathtrail-profile.json, in the folder MathTrail. Download or copy it there; the chat can give you its link.",
				"Other files with a profile: mathtrail-profile (1).json. MathTrail reads only the newest; you can delete the others.",
			],
			[
				"How to delete it",
				"Delete the file and empty Drive's bin: nothing of the profile stays anywhere.",
			],
			[
				"How to cut off access",
				"In your Google account, under third-party access, remove MathTrail: it can then reach nothing.",
			],
			[
				"How to remove the app",
				"Disconnect MathTrail in your chat's settings.",
			],
		]);
		// A card opens no link: the chat gives it.
		expect(root.querySelector("a")).toBeNull();
		expect(root.textContent).not.toContain("drive.google.com");
	});

	test("names the file alone when it lies in no folder, and no other file when there is none", async () => {
		const { root } = await draw({
			...inSwahili,
			location: { ...inSwahili.location, folder: "", others: [] },
		});

		expect(fields(root, "Your data")[0]).toEqual([
			"Where the profile is",
			"The whole profile is one file in your Google Drive: mathtrail-profile.json. Download or copy it there; the chat can give you its link.",
		]);
	});

	test("says nothing of the data when its tool does not say where the file is", async () => {
		const { location: _, ...saved } = inSwahili;
		const { root } = await draw(saved);

		expect(fields(root, "Your data")).toEqual([]);
		expect(root.querySelectorAll(".mt-fields")).toHaveLength(1);
	});

	test("of an earlier chat, after a change refused, says nothing was saved over the profile as it stays", async () => {
		const { root } = await draw(profileRefused);

		expect(root.querySelector(".mt-verdict-line")?.textContent).toBe(
			"Nothing was saved.",
		);
		expect(root.querySelector(".mt-verdict-detail")?.textContent).toBe(
			"The chat says what to change.",
		);
		expect(fields(root, "Profile · for the parent")[0]?.[1]).toBe("3");
		expect(root.textContent).not.toContain("40 given");
	});

	test("opens the form in its place from the button at its head, and asks the chat for nothing", async () => {
		const { root, heard } = await draw(inSwahili);
		const edit = buttonIn(root, "Edit");
		expect(edit.closest(".mt-fields-head")).not.toBeNull();

		press(edit);

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-form")).not.toBeNull(),
		);
		expect(heard.messages).toEqual([]);
		expect(heard.calls).toEqual([]);
	});
});
