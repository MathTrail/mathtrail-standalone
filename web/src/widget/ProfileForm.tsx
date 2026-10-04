import { useEffect, useMemo, useRef } from "preact/hooks";
import {
	Button,
	CheckGroup,
	ChipsField,
	type Choice,
	SelectField,
	TextField,
} from "../design/controls";
import type { Words } from "../i18n/words";
import {
	type Draft,
	type Editing,
	profileLimits,
	withCountry,
	withInterestAdded,
} from "./editing";
import { catalogSkills, countryName, languageName, skillName } from "./names";
import type { Problem } from "./payload";
import { countryCodes, regionsOf } from "./places";
import { isKey, type Key, lessonLanguages, useWords } from "./words";

/** Open is a form a parent is filling in, or one on its way to the service. */
type Open = Extract<Editing, { state: "open" | "saving" }>;

/**
 * ProfileForm is the form the adult changes the child's details with, on the
 * card: the pseudonym, the grade, the interests, what the child has not met at
 * school yet, the language of the lessons, and the country the family lives
 * in, with its state for the United States, which a parent may leave unsaid.
 * The parent's notes are not among them: they are said to the model, in the chat. What the service
 * refused is said under each field it refused, and what became of the save
 * under the buttons. A form on its way takes no second save and no cancel, and
 * its fields take nothing typed: what it sends is what it shows.
 */
export function ProfileForm({
	editing,
	onChange,
	onSave,
	onCancel,
}: {
	editing: Open;
	onChange: (draft: Draft) => void;
	onSave: () => void;
	onCancel: () => void;
}) {
	const words = useWords();
	const { draft } = editing;
	const saving = editing.state === "saving";
	const said = problemsSaid(
		words,
		editing.state === "open" ? editing.problems : [],
	);
	const first = useRef<HTMLInputElement>(null);
	useEffect(() => {
		first.current?.focus({ preventScroll: true });
	}, []);
	// Some two hundred and fifty countries are named and put in order for their
	// list, so it is drawn up again only when what it shows can change, and not
	// at every key typed into another field.
	const keptCountry = editing.from.country ?? "";
	const countries = useMemo(
		() => countryChoices(words, [keptCountry, draft.country]),
		[words, keptCountry, draft.country],
	);
	return (
		<form
			class="mt-form"
			aria-label={words.text("profile.form_label")}
			aria-busy={saving}
			onSubmit={(event) => event.preventDefault()}
		>
			<fieldset class="mt-form-fields" disabled={saving}>
				<TextField
					label={words.text("profile.pseudonym")}
					value={draft.pseudonym}
					most={profileLimits.pseudonym}
					note={words.text("profile.pseudonym_note")}
					problem={said.get("pseudonym")}
					inputRef={first}
					onInput={(pseudonym) => onChange({ ...draft, pseudonym })}
				/>
				<SelectField
					label={words.text("profile.grade")}
					value={String(draft.grade)}
					choices={gradeChoices(words)}
					note={words.text("profile.grade_note")}
					problem={said.get("grade")}
					onChange={(grade) => onChange({ ...draft, grade: Number(grade) })}
				/>
				<ChipsField
					label={words.text("profile.interests")}
					chips={draft.interests}
					typed={draft.interest}
					most={profileLimits.interests}
					longest={profileLimits.interest}
					typeLabel={words.text("profile.interest_new")}
					addLabel={words.text("profile.interest_add")}
					removeLabel={(interest) =>
						words.text("profile.interest_remove", { interest })
					}
					fullNote={words.text("profile.interests_full")}
					problem={said.get("interests")}
					onType={(interest) => onChange({ ...draft, interest })}
					onAdd={() => onChange(withInterestAdded(draft))}
					onRemove={(interest) =>
						onChange({
							...draft,
							interests: draft.interests.filter((kept) => kept !== interest),
						})
					}
				/>
				<CheckGroup
					legend={words.text("profile.not_yet")}
					choices={skillChoices(words, [
						...editing.from.excluded_skills,
						...draft.skills,
					])}
					value={draft.skills}
					problem={said.get("excluded_skills")}
					onChange={(skills) => onChange({ ...draft, skills })}
				/>
				<SelectField
					label={words.text("profile.language")}
					value={draft.language}
					choices={languageChoices(words, [
						editing.from.ui_language ?? "",
						draft.language,
					])}
					problem={said.get("ui_language")}
					onChange={(language) => onChange({ ...draft, language })}
				/>
				<SelectField
					label={words.text("profile.country")}
					value={draft.country}
					choices={countries}
					note={words.text("profile.country_note")}
					problem={said.get("country")}
					onChange={(country) => onChange(withCountry(draft, country))}
				/>
				{regionsOf(draft.country).length > 0 && (
					<SelectField
						label={words.text("profile.region")}
						value={draft.region}
						choices={regionChoices(words, draft.country, [
							editing.from.region ?? "",
							draft.region,
						])}
						problem={said.get("region")}
						onChange={(region) => onChange({ ...draft, region })}
					/>
				)}
			</fieldset>
			<div class="mt-form-actions">
				<div class="mt-btns">
					<Button variant="primary" locked={saving} onClick={onSave}>
						{words.text("profile.save")}
					</Button>
					<Button locked={saving} onClick={onCancel}>
						{words.text("profile.cancel")}
					</Button>
				</div>
				<p class="mt-action-note" aria-live="polite">
					{formNote(words, editing)}
				</p>
			</div>
		</form>
	);
}

// problemsSaid are the words for each field the service refused, by the field:
// the words of the first code given for it, or, for a code the card has no
// words for, words that fit any.
function problemsSaid(
	words: Words<Key>,
	problems: readonly Problem[],
): ReadonlyMap<string, string> {
	const said = new Map<string, string>();
	for (const { field, code } of problems) {
		if (!said.has(field)) {
			const key = `problem.${code}`;
			said.set(field, words.text(isKey(key) ? key : "problem.other"));
		}
	}
	return said;
}

// formNote is what became of the save, in words: on its way, not saved, or
// refused for the fields marked.
function formNote(words: Words<Key>, editing: Open): string {
	if (editing.state === "saving") {
		return words.text("profile.saving");
	}
	if (editing.failed) {
		return words.text("profile.save_failed");
	}
	return editing.problems.length > 0 ? words.text("profile.save_refused") : "";
}

// gradeChoices are the school years the lessons are for, each written as the
// card's language writes numbers.
function gradeChoices(words: Words<Key>): Choice[] {
	const numbers = new Intl.NumberFormat(words.locale);
	const choices: Choice[] = [];
	for (
		let grade = profileLimits.minGrade;
		grade <= profileLimits.maxGrade;
		grade++
	) {
		choices.push({ value: String(grade), label: numbers.format(grade) });
	}
	return choices;
}

// skillChoices are the skills a parent can leave out, in the catalog's order,
// each by the card's name for it — and after them any of kept that the catalog
// no longer has, by its id: what the profile leaves out is shown, and its row
// stays where it was once unticked. The service takes no skill the catalog
// does not have, and says so, by the field, for the card to ask for it to be
// unticked.
function skillChoices(words: Words<Key>, kept: readonly string[]): Choice[] {
	const unknown = new Set(
		kept.filter((skill) => !catalogSkills.includes(skill)),
	);
	return [...catalogSkills, ...unknown].map((skill) => ({
		value: skill,
		label: skillName(words, skill),
	}));
}

// countryChoices are the countries a parent can say the family lives in: none
// first, then each by the card's name for it and in the card's order of names —
// and among them one of kept the list does not have, by its code, so that what
// the profile says stays on the form.
function countryChoices(words: Words<Key>, kept: readonly string[]): Choice[] {
	const others = kept.filter(
		(code) => code !== "" && !countryCodes.includes(code),
	);
	const collator = new Intl.Collator(words.locale);
	const named = [...new Set([...countryCodes, ...others])]
		.map((code) => ({ value: code, label: countryName(words, code) }))
		.sort((one, other) => collator.compare(one.label, other.label));
	return [{ value: "", label: words.text("profile.none") }, ...named];
}

// regionChoices are the regions of the country a parent can name: none first,
// then each by its name in English, as the list names it, in the card's order
// of names — and among them one of kept the list does not have, by its code.
function regionChoices(
	words: Words<Key>,
	country: string,
	kept: readonly string[],
): Choice[] {
	const regions = regionsOf(country);
	const others = kept.filter(
		(code) => code !== "" && !regions.some((region) => region.code === code),
	);
	const collator = new Intl.Collator(words.locale);
	const named = [
		...regions.map(({ code, name }) => ({ value: code, label: name })),
		...others.map((code) => ({ value: code, label: code })),
	].sort((one, other) => collator.compare(one.label, other.label));
	return [{ value: "", label: words.text("profile.none") }, ...named];
}

// languageChoices are the languages of the lessons a parent can choose: the
// chat's first, then each the card speaks, by the card's name for it and in
// the card's order of names — and among them any of kept the card does not
// speak, such as one the chat set: the language the profile has stays on the
// form, to be chosen again once another was.
function languageChoices(words: Words<Key>, kept: readonly string[]): Choice[] {
	const others = kept.filter(
		(tag) => tag !== "" && !lessonLanguages.includes(tag),
	);
	const tags = [...new Set([...lessonLanguages, ...others])];
	const collator = new Intl.Collator(words.locale);
	const named = tags
		.map((tag) => ({ value: tag, label: languageName(words, tag) }))
		.sort((one, other) => collator.compare(one.label, other.label));
	return [{ value: "", label: words.text("profile.language_chat") }, ...named];
}
