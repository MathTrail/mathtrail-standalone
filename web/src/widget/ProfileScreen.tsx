import { useReducer, useRef, useState } from "preact/hooks";
import { Verdict } from "../design/blocks";
import { Button } from "../design/controls";
import { type Field, FieldsFrame, ProfileFields } from "../design/progress";
import { MessageHeader } from "../design/thread";
import type { Words } from "../i18n/words";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import {
	changesOf,
	type Draft,
	type Editing,
	editingAfter,
	notEditing,
} from "./editing";
import { sendEdit } from "./edits";
import { useFocusKeptOnTheCard } from "./focus";
import { countryName, languageName, listed, skillName } from "./names";
import { ProfileForm } from "./ProfileForm";
import type { Details, Location, ProfileReport } from "./payload";
import { regionName } from "./places";
import { countText, type Key, useWords } from "./words";

/**
 * ProfileCard is the card of the child's profile, as a card of an earlier chat
 * was drawn when the model read the profile or saved a change to it. Neither
 * tool draws one now; a host that still does draws this.
 */
export function ProfileCard({
	profile,
	host,
}: {
	profile: ProfileReport;
	host: Host;
}) {
	return (
		<CardRoot>
			{(wide) => <ProfileScreen profile={profile} wide={wide} host={host} />}
		</CardRoot>
	);
}

// ProfileScreen is the profile for the parent: a change just refused, if one
// was and nothing has been saved since, the details as they stand, and — where
// the tool that drew the card says where the file is — what the parent can do
// with the data.
function ProfileScreen({
	profile,
	wide,
	host,
}: {
	profile: ProfileReport;
	wide: boolean;
	host: Host;
}) {
	const words = useWords();
	const [details, setDetails] = useState(profile.details);
	const [saved, setSaved] = useState(false);
	return (
		<article aria-label={words.text("profile.card_label")}>
			<MessageHeader
				author="person"
				name={details.pseudonym}
				badge={words.text("child.grade", { grade: details.grade })}
				wide={wide}
			/>
			<div class="mt-progress">
				{profile.refused && !saved && (
					<Verdict detail={words.text("profile.not_saved_detail")}>
						{words.text("profile.not_saved")}
					</Verdict>
				)}
				<ParentProfile
					label={words.text("profile.label")}
					details={details}
					host={host}
					onSaved={(changed) => {
						setDetails(changed);
						setSaved(true);
					}}
				/>
				{profile.location !== undefined && (
					<ParentData location={profile.location} />
				)}
			</div>
		</article>
	);
}

/**
 * ParentProfile is the child's profile as the parent reads it on a card — the
 * grade, which is only a label once the child has started, the interests, what
 * the child has not met at school yet, and the language of the lessons — under
 * label, where what it stands in does not name it already, with the button at
 * its head that opens the form to change them in its place. The focus the
 * form had goes back to that button when it closes, and to what the card says
 * in its place when the profile is gone. A change saved is handed to onSaved,
 * and told to the model in the service's words, since the model is not called
 * by the form and would otherwise go on with what it was told before.
 */
export function ParentProfile({
	label,
	details,
	host,
	onSaved,
}: {
	label?: string;
	details: Details;
	host: Host;
	onSaved: (details: Details) => void;
}) {
	const words = useWords();
	const [editing, dispatch] = useReducer(editingAfter, notEditing);
	const sending = useRef(false);
	const edit = useRef<HTMLButtonElement>(null);
	const gone = useRef<HTMLParagraphElement>(null);
	useFocusKeptOnTheCard(editing.state === "closed", edit);
	useFocusKeptOnTheCard(editing.state === "gone", gone);

	async function save(from: Details, draft: Draft) {
		if (sending.current) {
			return;
		}
		const changes = changesOf(from, draft);
		if (Object.keys(changes).length === 0) {
			dispatch({ type: "closed" });
			return;
		}
		sending.current = true;
		dispatch({ type: "sent" });
		const outcome = await sendEdit(host, changes);
		sending.current = false;
		dispatch({ type: "answered", outcome });
		if (outcome.kind === "saved") {
			onSaved(outcome.details);
			if (outcome.told !== undefined) {
				host.tellModel(outcome.told).catch((error: unknown) => {
					console.error("widget: the model was not told of the change", error);
				});
			}
		}
	}

	switch (editing.state) {
		case "gone":
			return (
				<FieldsFrame label={label}>
					<p class="mt-fields-said" tabIndex={-1} ref={gone}>
						{words.text("profile.gone")}
					</p>
				</FieldsFrame>
			);
		case "open":
		case "saving":
			return (
				<FieldsFrame label={label}>
					<ProfileForm
						editing={editing}
						onChange={(draft) => dispatch({ type: "typed", draft })}
						onSave={() => save(editing.from, editing.draft)}
						onCancel={() => dispatch({ type: "closed" })}
					/>
				</FieldsFrame>
			);
		case "closed":
			return (
				<ProfileFields
					label={label}
					fields={detailFields(words, details)}
					action={
						<Button
							buttonRef={edit}
							onClick={() => dispatch({ type: "opened", details })}
						>
							{words.text("profile.edit")}
						</Button>
					}
					status={
						<p class="mt-action-note" aria-live="polite">
							{savedNote(words, editing)}
						</p>
					}
				/>
			);
	}
}

// savedNote is what became of the change last saved, in words: saved, and the
// language chosen comes with the next task when the adult chose another — the
// card goes on in the words it was drawn in.
function savedNote(
	words: Words<Key>,
	editing: Extract<Editing, { state: "closed" }>,
): string {
	switch (editing.said) {
		case "saved":
			return words.text("profile.saved");
		case "saved_language":
			return words.text("profile.saved_language");
		default:
			return "";
	}
}

/**
 * ParentData is what the parent can do with the child's data, each under the
 * question it answers rather than a word that reads as a button: where the
 * file is, which is the export, and the other files that hold a profile; how
 * to delete it; how to cut the service off from Drive; and how to remove the
 * app. It names the files and opens none of them.
 */
export function ParentData({ location }: { location: Location }) {
	const words = useWords();
	return (
		<ProfileFields
			label={words.text("data.label")}
			fields={dataFields(words, location)}
		/>
	);
}

// detailFields are the child's details as the parent's fields, each in the
// card's words: a list as its names, each drawn apart, or the word that it is
// not set when it is empty.
function detailFields(words: Words<Key>, details: Details): Field[] {
	const none = words.text("profile.none");
	const skills = details.excluded_skills.map((id) => skillName(words, id));
	return [
		{
			term: words.text("profile.grade"),
			value: countText(words, details.grade),
			note: words.text("profile.grade_note"),
		},
		{
			term: words.text("profile.interests"),
			value: details.interests.length > 0 ? details.interests : none,
		},
		{
			term: words.text("profile.not_yet"),
			value: skills.length > 0 ? skills : none,
		},
		{
			term: words.text("profile.language"),
			value:
				details.ui_language === null
					? words.text("profile.language_chat")
					: languageName(words, details.ui_language),
		},
		{
			term: words.text("profile.country"),
			value:
				details.country === null ? none : countryName(words, details.country),
			note: words.text("profile.country_note"),
		},
		...(details.region === null
			? []
			: [
					{
						term: words.text("profile.region"),
						value:
							regionName(details.country ?? "", details.region) ??
							details.region,
					},
				]),
	];
}

// dataFields are what the parent can do with the child's data, each under the
// question it answers.
function dataFields(words: Words<Key>, location: Location): Field[] {
	const others = location.others.map((other) => other.file);
	return [
		{
			term: words.text("data.where"),
			value:
				location.folder === ""
					? words.text("data.export_file", { file: location.file })
					: words.text("data.export_where", {
							file: location.file,
							folder: location.folder,
						}),
			note:
				others.length > 0
					? words.text("data.others", { files: listed(words, others) })
					: undefined,
		},
		{
			term: words.text("data.how_delete"),
			value: words.text("data.delete_how"),
		},
		{
			term: words.text("data.how_cut_off"),
			value: words.text("data.access_how"),
		},
		{
			term: words.text("data.how_remove"),
			value: words.text("data.remove_how"),
		},
	];
}
