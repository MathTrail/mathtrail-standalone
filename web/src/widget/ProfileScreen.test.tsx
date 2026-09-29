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
const inPortuguese = {
	...profileRead,
	profile: { ...profileRead.profile, ui_language: "pt-BR" },
};

// fields are the fields of the section labelled label.
function fields(root: HTMLElement, label: string): string[][] {
	const section = [...root.querySelectorAll(".mt-fields")].find(
		(found) => found.querySelector(".mt-section-label")?.textContent === label,
	);
	return [...(section?.querySelectorAll("dl > div") ?? [])].map((field) =>
		[...field.children].map((part) => part.textContent ?? ""),
	);
}

describe("the profile's card", () => {
	test("shows the child's details, and never the parent's notes", async () => {
		const { root } = await draw(inPortuguese);

		expect(root.querySelector("article")?.getAttribute("aria-label")).toBe(
			"Profile",
		);
		expect(root.querySelector(".mt-head .mt-name")?.textContent).toBe("Comet");
		expect(fields(root, "Profile · for the parent")).toEqual([
			["Grade", "3", "Only a label: changing it moves no rating."],
			["Interests", "space, animals, football"],
			["Not at school yet", "Division with a remainder"],
			["Language of the cards", "Brazilian Portuguese"],
		]);
		expect(root.textContent).not.toContain("Loses heart");
		expect(root.querySelector(".mt-bar")).toBeNull();
	});

	test("says where the profile's file is, and what the parent can do with it", async () => {
		const { root } = await draw(inPortuguese);

		expect(fields(root, "Your data")).toEqual([
			[
				"Export",
				"The whole profile is one file in your Google Drive: mathtrail-profile.json, in the folder MathTrail. Download or copy it there; the chat can give you its link.",
				"Other files with a profile: mathtrail-profile (1).json. MathTrail reads only the newest; you can delete the others.",
			],
			[
				"Delete",
				"Delete the file and empty Drive's bin: nothing of the profile stays anywhere.",
			],
			[
				"Cut off access",
				"In your Google account, under third-party access, remove MathTrail: it can then reach nothing.",
			],
			["Remove the app", "Disconnect MathTrail in your chat's settings."],
		]);
		// A card opens no link: the chat gives it.
		expect(root.querySelector("a")).toBeNull();
		expect(root.textContent).not.toContain("drive.google.com");
	});

	test("names the file alone when it lies in no folder, and no other file when there is none", async () => {
		const { root } = await draw({
			...inPortuguese,
			location: { ...inPortuguese.location, folder: "", others: [] },
		});

		expect(fields(root, "Your data")[0]).toEqual([
			"Export",
			"The whole profile is one file in your Google Drive: mathtrail-profile.json. Download or copy it there; the chat can give you its link.",
		]);
	});

	test("says nothing of the data when its tool does not say where the file is", async () => {
		const { location: _, ...saved } = inPortuguese;
		const { root } = await draw(saved);

		expect(fields(root, "Your data")).toEqual([]);
		expect(root.querySelectorAll(".mt-fields")).toHaveLength(1);
	});

	test("after a change refused, says nothing was saved over the profile as it stays", async () => {
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

	test("asks the chat to edit the profile", async () => {
		const { root, heard } = await draw(inPortuguese);

		press(buttonIn(root, "Edit profile"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Edit profile"]));
		expect(heard.calls).toEqual([]);
	});
});
