import { describe, expect, test } from "vitest";
import schema from "../../../internal/domain/profile/profile.schema.json";
import {
	changesOf,
	type Draft,
	draftOf,
	type Editing,
	editingAfter,
	notEditing,
	profileLimits,
	withCountry,
	withInterestAdded,
} from "./editing";
import type { Details } from "./payload";
import { countryCodes, regionName, regionsOf } from "./places";

const details: Details = {
	pseudonym: "Comet",
	grade: 3,
	interests: ["space", "animals"],
	excluded_skills: ["fractions", "division"],
	ui_language: null,
	country: null,
	region: null,
	signin_country_off: false,
};

// drafted is the draft of the details with what a case changes in it.
function drafted(change: Partial<Draft> = {}): Draft {
	return { ...draftOf(details), ...change };
}

describe("the form's limits", () => {
	// A limit the form holds a parent to that the file does not would refuse
	// what the service takes; one the file holds that the form does not would
	// let the parent type what the service refuses.
	test("are the profile's own", () => {
		const student = schema.$defs.student.properties;
		expect(profileLimits.pseudonym).toBe(student.pseudonym.maxLength);
		expect(profileLimits.minGrade).toBe(student.grade.minimum);
		expect(profileLimits.maxGrade).toBe(student.grade.maximum);
		expect(profileLimits.interests).toBe(student.interests.maxItems);
		expect(profileLimits.interest).toBe(student.interests.items.maxLength);
	});
});

describe("a draft", () => {
	test("opens on the details as they stand, the chat's language as no text", () => {
		expect(draftOf(details)).toEqual({
			pseudonym: "Comet",
			grade: 3,
			interests: ["space", "animals"],
			interest: "",
			skills: ["fractions", "division"],
			language: "",
			country: "",
			region: "",
			signInCountryOff: false,
		});
		expect(draftOf({ ...details, ui_language: "fr" }).language).toBe("fr");
		expect(
			draftOf({ ...details, signin_country_off: true }).signInCountryOff,
		).toBe(true);
		expect(
			draftOf({ ...details, country: "US", region: "US-TX" }),
		).toMatchObject({ country: "US", region: "US-TX" });
	});

	test("keeps the state while the country stays, and leaves it behind for another country", () => {
		const inTexas = drafted({ country: "US", region: "US-TX" });
		expect(withCountry(inTexas, "US")).toEqual(inTexas);
		expect(withCountry(inTexas, "FR")).toEqual(
			drafted({ country: "FR", region: "" }),
		);
		expect(withCountry(inTexas, "")).toEqual(
			drafted({ country: "", region: "" }),
		);
	});

	test("adds the interest being typed, without its spaces, and begins the typing afresh", () => {
		expect(withInterestAdded(drafted({ interest: "  chess " }))).toEqual(
			drafted({ interests: ["space", "animals", "chess"], interest: "" }),
		);
	});

	test("adds an interest it has once, and begins the typing afresh", () => {
		expect(withInterestAdded(drafted({ interest: "space" }))).toEqual(
			drafted({ interest: "" }),
		);
	});

	test.each([
		["nothing typed", drafted({ interest: "   " })],
		[
			"as many interests as a profile holds",
			drafted({
				interests: Array.from({ length: profileLimits.interests }, (_, at) =>
					String(at),
				),
				interest: "one more",
			}),
		],
	])("adds nothing with %s", (_, draft) => {
		expect(withInterestAdded(draft)).toBe(draft);
	});
});

describe("the changes of a draft", () => {
	test("are none when it says what the details say", () => {
		expect(changesOf(details, drafted())).toEqual({});
	});

	test("are only the details it says otherwise, named as the service takes them", () => {
		expect(
			changesOf(details, drafted({ pseudonym: "Nova", language: "fr" })),
		).toEqual({ pseudonym: "Nova", ui_language: "fr" });
		expect(changesOf(details, drafted({ grade: 5 }))).toEqual({ grade: 5 });
	});

	test("compare a pseudonym without the spaces around it", () => {
		expect(changesOf(details, drafted({ pseudonym: "  Comet " }))).toEqual({});
	});

	test("take the skills as a set, and the interests in their order", () => {
		expect(
			changesOf(details, drafted({ skills: ["division", "fractions"] })),
		).toEqual({});
		expect(
			changesOf(details, drafted({ interests: ["animals", "space"] })),
		).toEqual({ interests: ["animals", "space"] });
		expect(changesOf(details, drafted({ skills: [] }))).toEqual({
			excluded_skills: [],
		});
	});

	test("count the interest still being typed as added", () => {
		expect(changesOf(details, drafted({ interest: "chess" }))).toEqual({
			interests: ["space", "animals", "chess"],
		});
	});

	test("give the language back to the chat as no text", () => {
		expect(
			changesOf({ ...details, ui_language: "fr" }, drafted({ language: "" })),
		).toEqual({ ui_language: "" });
	});

	test("leave the country of the sign-in out, and count it again", () => {
		expect(changesOf(details, drafted({ signInCountryOff: true }))).toEqual({
			signin_country_off: true,
		});
		expect(
			changesOf(
				{ ...details, signin_country_off: true },
				drafted({ signInCountryOff: false }),
			),
		).toEqual({ signin_country_off: false });
	});
});

describe("the form", () => {
	const open: Editing = {
		state: "open",
		from: details,
		draft: drafted(),
		problems: [],
		failed: false,
	};
	const saving: Editing = { state: "saving", from: details, draft: drafted() };

	test("opens on the details, and closes unsaved", () => {
		expect(editingAfter(notEditing, { type: "opened", details })).toEqual(open);
		expect(editingAfter(open, { type: "closed" })).toEqual(notEditing);
	});

	test("takes what is typed while it is open", () => {
		const draft = drafted({ pseudonym: "Nova" });
		expect(editingAfter(open, { type: "typed", draft })).toEqual({
			...open,
			draft,
		});
	});

	test("goes on its way when sent", () => {
		expect(editingAfter(open, { type: "sent" })).toEqual(saving);
	});

	test.each<[string, Editing, Editing]>([
		["saved", saving, { state: "closed", said: "saved" }],
		[
			"saved with another language",
			{ ...saving, draft: drafted({ language: "fr" }) },
			{ state: "closed", said: "saved_language" },
		],
	])("closes once %s, and says so", (_, before, after) => {
		expect(
			editingAfter(before, {
				type: "answered",
				outcome: { kind: "saved", details, told: undefined },
			}),
		).toEqual(after);
	});

	test("opens again on what was typed when it is refused, with what was refused", () => {
		const problems = [{ field: "pseudonym", code: "required" }];
		expect(
			editingAfter(saving, {
				type: "answered",
				outcome: { kind: "refused", problems },
			}),
		).toEqual({ ...open, problems });
	});

	test("opens again on what was typed when it is not saved, and says so", () => {
		expect(
			editingAfter(saving, { type: "answered", outcome: { kind: "failed" } }),
		).toEqual({ ...open, failed: true });
	});

	test("closes for good when the profile is gone", () => {
		expect(
			editingAfter(saving, { type: "answered", outcome: { kind: "gone" } }),
		).toEqual({ state: "gone" });
	});

	test.each<[string, Editing, Parameters<typeof editingAfter>[1]]>([
		[
			"typed into while on its way",
			saving,
			{ type: "typed", draft: drafted() },
		],
		["closed while on its way", saving, { type: "closed" }],
		["sent while on its way", saving, { type: "sent" }],
		["sent while closed", notEditing, { type: "sent" }],
		["opened while open", open, { type: "opened", details }],
		[
			"answered while open",
			open,
			{ type: "answered", outcome: { kind: "gone" } },
		],
		["opened once gone", { state: "gone" }, { type: "opened", details }],
	])("does not change when %s", (_, before, event) => {
		expect(editingAfter(before, event)).toBe(before);
	});
});

describe("the places the form offers", () => {
	// The form offers the codes of the file the service holds a place to, read
	// from that file: a country or a state the service would refuse is never
	// offered, and none the service takes is missing.
	test("are the service's own list", () => {
		expect(countryCodes).toHaveLength(250);
		expect(countryCodes).toContain("US");
		expect(countryCodes).toContain("XK");
		expect(regionsOf("US")).toHaveLength(51);
		expect(regionsOf("FR")).toEqual([]);
		expect(regionName("US", "US-TX")).toBe("Texas");
		expect(regionName("US", "US-ZZ")).toBeUndefined();
		expect(regionName("FR", "US-TX")).toBeUndefined();
		// The codes come from a file a person can edit, and a code that names
		// something every object has names no region.
		expect(regionName("constructor", "name")).toBeUndefined();
		expect(regionName("US", "valueOf")).toBeUndefined();
	});

	test("are sent as the codes the adult changed, and no others", () => {
		const inTexas = { ...details, country: "US", region: "US-TX" };
		expect(changesOf(inTexas, draftOf(inTexas))).toEqual({});
		expect(changesOf(inTexas, withCountry(draftOf(inTexas), "FR"))).toEqual({
			country: "FR",
			region: "",
		});
		expect(
			changesOf(details, {
				...draftOf(details),
				country: "US",
				region: "US-CA",
			}),
		).toEqual({ country: "US", region: "US-CA" });
	});
});
