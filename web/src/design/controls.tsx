import type { ComponentChildren, Ref } from "preact";
import { useId } from "preact/hooks";
import { classes } from "./classes";
import { Icon, type IconName } from "./icons";

/**
 * Button is the design's button, secondary unless it is the one thing to do
 * next. One that shows and hides something says whether it is shown. A
 * disabled button is out of use; a locked one ignores presses but keeps its
 * place in the tab order and the focus it has, so that a press answered in a
 * moment does not throw the focus away.
 */
export function Button({
	variant = "secondary",
	expanded,
	disabled = false,
	locked = false,
	onClick,
	buttonRef,
	children,
}: {
	variant?: "primary" | "secondary";
	expanded?: boolean;
	disabled?: boolean;
	locked?: boolean;
	onClick?: () => void;
	buttonRef?: Ref<HTMLButtonElement>;
	children: ComponentChildren;
}) {
	return (
		<button
			type="button"
			ref={buttonRef}
			class={classes("mt-btn", variant === "primary" && "mt-btn-primary")}
			disabled={disabled}
			aria-disabled={locked ? "true" : undefined}
			aria-expanded={expanded}
			onClick={() => {
				if (!locked) {
					onClick?.();
				}
			}}
		>
			{children}
		</button>
	);
}

/**
 * OptionState is how an answer option stands: waiting to be chosen, chosen and
 * being checked, the right answer, the child's wrong answer, or one of the
 * others once the answer is known.
 */
export type OptionState =
	| "default"
	| "selected"
	| "correct"
	| "wrong"
	| "muted";

// The icon beside each state's words, for the states that have words.
const stateIcons: Partial<Record<OptionState, IconName>> = {
	selected: "spinner",
	correct: "check",
	wrong: "cross",
};

/** Option is one answer the child can give: its letter and its text. */
export type Option<L extends string = string> = {
	letter: L;
	value: string;
	state?: OptionState;
	status?: string;
};

/**
 * Said names the language some words are in, and the way they run, where they
 * may differ from the card's own: a task is written in the chat's language,
 * and the card speaks the one the parent chose for it.
 */
export type Said = { lang?: string; dir?: "ltr" | "rtl" };

/**
 * OptionRow is one answer as a button: its letter, its text and, once it has
 * one, its status — checking, the right answer, the child's. A press gives the
 * answer, so the row is a button rather than one of a group of radio buttons,
 * whose arrow keys would give an answer by moving through them. A locked row
 * keeps the focus it has and ignores presses.
 */
export function OptionRow<L extends string>({
	letter,
	value,
	state = "default",
	status,
	locked = false,
	onSelect,
	said = {},
}: Option<L> & {
	locked?: boolean;
	onSelect?: (letter: L) => void;
	said?: Said;
}) {
	const icon = stateIcons[state];
	// The spaces between the parts cost nothing in a flex row, and keep a
	// screen reader from reading the letter and the text as one word.
	return (
		<button
			type="button"
			class="mt-option"
			data-state={state}
			aria-disabled={locked ? "true" : undefined}
			onClick={() => {
				if (!locked) {
					onSelect?.(letter);
				}
			}}
		>
			<span class="mt-option-letter">{letter}</span>{" "}
			<span class="mt-option-value" lang={said.lang} dir={said.dir}>
				{value}
			</span>
			{status !== undefined && (
				<>
					{" "}
					<span class="mt-option-status">
						{icon !== undefined && (
							<Icon
								name={icon}
								size={16}
								track={state === "selected" ? "accent-tint" : undefined}
							/>
						)}
						{status}
					</span>
				</>
			)}
		</button>
	);
}

/**
 * OptionList is the five answers under a legend that says what to do with
 * them, or what they now show. When the list is locked, every row is. The
 * answers' texts are in the language said, the legend in the card's.
 */
export function OptionList<L extends string>({
	legend,
	options,
	locked = false,
	onSelect,
	said,
}: {
	legend: string;
	options: readonly Option<L>[];
	locked?: boolean;
	onSelect?: (letter: L) => void;
	said?: Said;
}) {
	return (
		<fieldset class="mt-options">
			<legend>{legend}</legend>
			<div class="mt-options-list">
				{options.map((option) => (
					<OptionRow
						key={option.letter}
						{...option}
						locked={locked}
						onSelect={onSelect}
						said={said}
					/>
				))}
			</div>
		</fieldset>
	);
}

/**
 * ReplyField is the one-line field a question is typed into, with its send
 * button. Its text is held by whoever draws it, so that the words typed
 * survive the field being taken off the page and put back. Enter sends, unless
 * it only closes a word being composed, as it does in the input methods of
 * Chinese and Japanese; blank text is never sent.
 */
export function ReplyField({
	value,
	placeholder,
	label,
	sendLabel,
	disabled = false,
	onInput,
	onSend,
}: {
	value: string;
	placeholder: string;
	label: string;
	sendLabel: string;
	disabled?: boolean;
	onInput: (value: string) => void;
	onSend: (words: string) => void;
}) {
	const id = useId();
	const words = value.trim();
	const send = () => {
		if (!disabled && words !== "") {
			onSend(words);
		}
	};
	return (
		<div class="mt-field" data-disabled={disabled ? "true" : undefined}>
			<label for={id} class="mt-vh">
				{label}
			</label>
			<input
				id={id}
				type="text"
				placeholder={placeholder}
				disabled={disabled}
				value={value}
				onInput={(event) => onInput(event.currentTarget.value)}
				onKeyDown={(event) => {
					if (event.key === "Enter" && !event.isComposing) {
						event.preventDefault();
						send();
					}
				}}
			/>
			<button
				type="button"
				class="mt-icon-btn"
				aria-label={sendLabel}
				disabled={disabled || words === ""}
				onClick={send}
			>
				<Icon name="send" size={18} className="mt-send" />
			</button>
		</div>
	);
}

/**
 * Checkbox is a statement the reader ticks to confirm it, the whole line a
 * target for a finger.
 */
export function Checkbox({
	label,
	checked,
	onChange,
}: {
	label: string;
	checked: boolean;
	onChange: (checked: boolean) => void;
}) {
	return (
		<label class="mt-check">
			<input
				type="checkbox"
				checked={checked}
				onChange={(event) => onChange(event.currentTarget.checked)}
			/>
			<span>{label}</span>
		</label>
	);
}
