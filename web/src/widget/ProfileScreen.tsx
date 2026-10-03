import { Verdict } from "../design/blocks";
import { Button } from "../design/controls";
import { type Field, ProfileFields } from "../design/progress";
import { MessageHeader } from "../design/thread";
import type { Words } from "../i18n/words";
import type { Host } from "./bridge";
import { CardRoot } from "./CardRoot";
import { RequestNote, useChatRequest } from "./ChatRequest";
import { languageName, listed, skillName } from "./names";
import type { Details, Location, ProfileReport } from "./payload";
import { type Key, useWords } from "./words";

/**
 * ProfileCard is the card of the child's profile, drawn when the model reads
 * the profile or saves a change to it.
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
// was, the details as they stand, and — where the tool that drew the card says
// where the file is — what the parent can do with the data.
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
	const { details } = profile;
	return (
		<article aria-label={words.text("profile.card_label")}>
			<MessageHeader
				author="person"
				name={details.pseudonym}
				badge={words.text("child.grade", { grade: details.grade })}
				wide={wide}
			/>
			<div class="mt-progress">
				{profile.refused && (
					<Verdict detail={words.text("profile.not_saved_detail")}>
						{words.text("profile.not_saved")}
					</Verdict>
				)}
				<ParentProfile details={details} host={host} />
				{profile.location !== undefined && (
					<ParentData location={profile.location} />
				)}
			</div>
		</article>
	);
}

/**
 * ParentProfile is the child's profile as the parent reads it on a card: the
 * grade, which is only a label once the child has started, the interests,
 * what the child has not met at school yet, and the language of the lessons —
 * with the one way to change them, which is to ask in the chat, at its head.
 */
export function ParentProfile({
	details,
	host,
}: {
	details: Details;
	host: Host;
}) {
	const words = useWords();
	const request = useChatRequest(host);
	return (
		<ProfileFields
			label={words.text("profile.label")}
			fields={detailFields(words, details)}
			action={
				<Button
					locked={request.state === "sent"}
					onClick={() => request.send(words.text("profile.edit"))}
				>
					{words.text("profile.edit")}
				</Button>
			}
			status={<RequestNote state={request.state} />}
		/>
	);
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
			value: new Intl.NumberFormat(words.locale).format(details.grade),
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
