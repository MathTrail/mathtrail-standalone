import type { Details, EditOutcome, Problem } from "./payload";

/**
 * profileLimits are the limits the form holds the details to before anything
 * is sent: the ones the profile's file holds them to, so that what the form
 * lets through is never refused for its size.
 */
export const profileLimits = {
	/** pseudonym is how many characters a pseudonym may have. */
	pseudonym: 32,
	/** minGrade and maxGrade are the school years the lessons are for. */
	minGrade: 1,
	maxGrade: 6,
	/** interests is how many interests a profile holds. */
	interests: 10,
	/** interest is how many characters one interest may have. */
	interest: 40,
} as const;

/**
 * Draft is the form as the adult fills it in: each detail the form changes,
 * the interest still being typed, the language of the lessons, the chat's
 * being the empty text, and the country and the region, none being the empty
 * text.
 */
export type Draft = {
	pseudonym: string;
	grade: number;
	interests: readonly string[];
	interest: string;
	skills: readonly string[];
	language: string;
	country: string;
	region: string;
};

/** draftOf is the form as it opens: the details as they stand. */
export function draftOf(details: Details): Draft {
	return {
		pseudonym: details.pseudonym,
		grade: details.grade,
		interests: details.interests,
		interest: "",
		skills: details.excluded_skills,
		language: details.ui_language ?? "",
		country: details.country ?? "",
		region: details.region ?? "",
	};
}

/**
 * withCountry is the draft with the country chosen. A family that moves to
 * another country leaves the old one's region behind, as the service does.
 */
export function withCountry(draft: Draft, country: string): Draft {
	return {
		...draft,
		country,
		region: country === draft.country ? draft.region : "",
	};
}

/**
 * withInterestAdded is the draft with the interest being typed added to the
 * others, as itself without the spaces around it, and the typing begun
 * afresh. One already there is not added twice, and its typing is begun afresh
 * too; an empty one, or one past what a profile holds, is not added, and the
 * typing is left as it is.
 */
export function withInterestAdded(draft: Draft): Draft {
	const typed = draft.interest.trim();
	if (draft.interests.includes(typed)) {
		return { ...draft, interest: "" };
	}
	if (typed === "" || draft.interests.length >= profileLimits.interests) {
		return draft;
	}
	return { ...draft, interests: [...draft.interests, typed], interest: "" };
}

/**
 * changesOf are the details a draft says otherwise than those it was opened
 * on, named as edit_profile takes them — and no others, so that a change made
 * in the chat while the form was open is kept. A pseudonym is compared without
 * the spaces around it; the interests in their order, since tasks are dressed
 * in them in turn, the one still being typed among them; the skills as a set;
 * and the chat's language, no country and no region are each the empty text.
 */
export function changesOf(
	details: Details,
	draft: Draft,
): Record<string, unknown> {
	const changes: Record<string, unknown> = {};
	if (draft.pseudonym.trim() !== details.pseudonym) {
		changes.pseudonym = draft.pseudonym;
	}
	if (draft.grade !== details.grade) {
		changes.grade = draft.grade;
	}
	const { interests } = withInterestAdded(draft);
	if (!sameList(interests, details.interests)) {
		changes.interests = interests;
	}
	if (!sameSet(draft.skills, details.excluded_skills)) {
		changes.excluded_skills = draft.skills;
	}
	if (draft.language !== (details.ui_language ?? "")) {
		changes.ui_language = draft.language;
	}
	if (draft.country !== (details.country ?? "")) {
		changes.country = draft.country;
	}
	if (draft.region !== (details.region ?? "")) {
		changes.region = draft.region;
	}
	return changes;
}

// sameList says whether two lists hold the same texts in the same order.
function sameList(one: readonly string[], other: readonly string[]): boolean {
	return (
		one.length === other.length && one.every((text, at) => text === other[at])
	);
}

// sameSet says whether two lists hold the same texts, in any order.
function sameSet(one: readonly string[], other: readonly string[]): boolean {
	const set = new Set(one);
	return set.size === new Set(other).size && other.every((t) => set.has(t));
}

/**
 * Editing is where the form stands: closed, with what became of the change
 * last saved; open, on the details it was opened on, with what the service
 * refused of the draft or that it could not be saved; on its way to the
 * service; or closed for good, the profile not being there any more.
 */
export type Editing =
	| { state: "closed"; said: "saved" | "saved_language" | undefined }
	| {
			state: "open";
			from: Details;
			draft: Draft;
			problems: readonly Problem[];
			failed: boolean;
	  }
	| { state: "saving"; from: Details; draft: Draft }
	| { state: "gone" };

/**
 * EditingEvent is what moves the form on: opened on the details as they
 * stand, typed into, closed without saving, sent, and answered.
 */
export type EditingEvent =
	| { type: "opened"; details: Details }
	| { type: "typed"; draft: Draft }
	| { type: "closed" }
	| { type: "sent" }
	| { type: "answered"; outcome: EditOutcome };

/** notEditing is the form before it is first opened. */
export const notEditing: Editing = { state: "closed", said: undefined };

/**
 * editingAfter is where the form stands after event. Nothing is typed into a
 * form on its way, and nothing sent from one not open: an event out of its
 * turn changes nothing.
 */
export function editingAfter(editing: Editing, event: EditingEvent): Editing {
	switch (event.type) {
		case "opened":
			return editing.state === "closed"
				? {
						state: "open",
						from: event.details,
						draft: draftOf(event.details),
						problems: [],
						failed: false,
					}
				: editing;
		case "typed":
			return editing.state === "open"
				? { ...editing, draft: event.draft }
				: editing;
		case "closed":
			return editing.state === "open" ? notEditing : editing;
		case "sent":
			return editing.state === "open"
				? { state: "saving", from: editing.from, draft: editing.draft }
				: editing;
		case "answered":
			return editing.state === "saving"
				? answered(editing, event.outcome)
				: editing;
	}
}

// answered is where a form on its way stands once the service has answered:
// closed when the change was saved — saying the language chosen comes with the
// next task when the adult chose another — open again when it was refused or
// not saved, and closed for good when the profile is gone.
function answered(
	saving: Extract<Editing, { state: "saving" }>,
	outcome: EditOutcome,
): Editing {
	const { from, draft } = saving;
	switch (outcome.kind) {
		case "saved":
			return {
				state: "closed",
				said:
					draft.language === (from.ui_language ?? "")
						? "saved"
						: "saved_language",
			};
		case "refused":
			return {
				state: "open",
				from,
				draft,
				problems: outcome.problems,
				failed: false,
			};
		case "gone":
			return { state: "gone" };
		case "failed":
			return { state: "open", from, draft, problems: [], failed: true };
	}
}
