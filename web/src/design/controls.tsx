import type { ComponentChildren, Ref } from "preact";
import { useEffect, useId, useRef } from "preact/hooks";
import { classes } from "./classes";
import { Icon, type IconName } from "./icons";

/**
 * Button is the design's button, secondary unless it is the one thing to do
 * next. One that shows and hides something says whether it is shown, and, when
 * what it shows is apart from it, names it by its id. A disabled button is out
 * of use; a locked one ignores presses but keeps its place in the tab order
 * and the focus it has, so that a press answered in a moment does not throw
 * the focus away. A class given is the button's own, beside the design's.
 */
export function Button({
	variant = "secondary",
	expanded,
	controls,
	disabled = false,
	locked = false,
	onClick,
	buttonRef,
	className,
	children,
}: {
	variant?: "primary" | "secondary";
	expanded?: boolean;
	controls?: string;
	disabled?: boolean;
	locked?: boolean;
	onClick?: () => void;
	buttonRef?: Ref<HTMLButtonElement>;
	className?: string;
	children: ComponentChildren;
}) {
	return (
		<button
			type="button"
			ref={buttonRef}
			class={classes(
				"mt-btn",
				variant === "primary" && "mt-btn-primary",
				className,
			)}
			disabled={disabled}
			aria-disabled={locked ? "true" : undefined}
			aria-expanded={expanded}
			aria-controls={controls}
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

/**
 * Told is what is said under a field: a line of explanation, and what is wrong
 * with what it holds. Each is read out with the field, by the ids it is given.
 */
type Told = { note?: string; problem?: string };

// toldUnder are the lines said under the field id, and the ids a control
// names them by for a screen reader, or undefined when there is nothing to
// say.
function toldUnder(
	id: string,
	{ note, problem }: Told,
): { ids: string | undefined; lines: ComponentChildren } {
	const ids = [
		note !== undefined && `${id}-note`,
		problem !== undefined && `${id}-problem`,
	].filter((named): named is string => named !== false);
	return {
		ids: ids.length > 0 ? ids.join(" ") : undefined,
		lines: (
			<>
				{note !== undefined && (
					<p id={`${id}-note`} class="mt-field-note">
						{note}
					</p>
				)}
				{problem !== undefined && (
					<p id={`${id}-problem`} class="mt-field-problem">
						{problem}
					</p>
				)}
			</>
		),
	};
}

/**
 * TextField is a line of text a person types, under its name. Its text is held
 * by whoever draws it, and is never longer than most characters, each counted
 * as withinLength counts it. Enter does what onEnter says, unless it only
 * closes a word being composed, as it does in the input methods of Chinese and
 * Japanese.
 */
export function TextField({
	label,
	value,
	most,
	note,
	problem,
	onInput,
	onEnter,
	inputRef,
}: Told & {
	label: string;
	value: string;
	most: number;
	onInput: (value: string) => void;
	onEnter?: () => void;
	inputRef?: Ref<HTMLInputElement>;
}) {
	const id = useId();
	const told = toldUnder(id, { note, problem });
	return (
		<div class="mt-form-field">
			<label for={id} class="mt-form-label">
				{label}
			</label>
			<input
				id={id}
				ref={inputRef}
				type="text"
				class="mt-input"
				value={value}
				aria-invalid={problem !== undefined ? "true" : undefined}
				aria-describedby={told.ids}
				onInput={(event) =>
					onInput(withinLength(event.currentTarget.value, most))
				}
				onKeyDown={(event) => {
					if (event.key === "Enter" && !event.isComposing) {
						event.preventDefault();
						onEnter?.();
					}
				}}
			/>
			{told.lines}
		</div>
	);
}

/**
 * withinLength is text cut to at most most characters, a character counted as
 * one however many units of the page's strings it takes — an emoji is one —
 * as a profile counts it. A box's own limit counts those units instead, and
 * would cut an emoji pseudonym the profile takes at half its length.
 */
export function withinLength(text: string, most: number): string {
	const characters = [...text];
	return characters.length > most ? characters.slice(0, most).join("") : text;
}

/** Choice is one of the values a list offers: the value, and its words. */
export type Choice = { value: string; label: string };

/**
 * SelectField is one value chosen from a list, under its name: the platform's
 * own list, which a phone opens as a list of its own the width of the screen.
 */
export function SelectField({
	label,
	value,
	choices,
	note,
	problem,
	onChange,
}: Told & {
	label: string;
	value: string;
	choices: readonly Choice[];
	onChange: (value: string) => void;
}) {
	const id = useId();
	const told = toldUnder(id, { note, problem });
	return (
		<div class="mt-form-field">
			<label for={id} class="mt-form-label">
				{label}
			</label>
			<select
				id={id}
				class="mt-input mt-select"
				value={value}
				aria-invalid={problem !== undefined ? "true" : undefined}
				aria-describedby={told.ids}
				onChange={(event) => onChange(event.currentTarget.value)}
			>
				{choices.map((choice) => (
					<option key={choice.value} value={choice.value}>
						{choice.label}
					</option>
				))}
			</select>
			{told.lines}
		</div>
	);
}

/**
 * ChipsField is a list of short texts a person adds to and takes from, under
 * its name: each text drawn apart with a button that takes it out, and a line
 * to type the next one into, added with its button or with Enter. A list as
 * long as it may be takes no more, and says so. The focus stays where the next text
 * is typed once one is added or taken out — the button pressed for it may have
 * gone, or been switched off — or, the list full and the line shut, on the
 * button that takes the last one out.
 */
export function ChipsField({
	label,
	chips,
	typed,
	most,
	longest,
	typeLabel,
	addLabel,
	removeLabel,
	fullNote,
	problem,
	onType,
	onAdd,
	onRemove,
}: {
	label: string;
	chips: readonly string[];
	typed: string;
	most: number;
	longest: number;
	typeLabel: string;
	addLabel: string;
	removeLabel: (chip: string) => string;
	fullNote: string;
	problem?: string;
	onType: (typed: string) => void;
	onAdd: () => void;
	onRemove: (chip: string) => void;
}) {
	const id = useId();
	const full = chips.length >= most;
	const told = toldUnder(id, { note: full ? fullNote : undefined, problem });
	const input = useRef<HTMLInputElement>(null);
	const list = useRef<HTMLUListElement>(null);
	const moving = useRef(false);
	useEffect(() => {
		if (!moving.current) {
			return;
		}
		moving.current = false;
		const last = list.current?.querySelector<HTMLButtonElement>(
			"li:last-child .mt-chip-remove",
		);
		(full ? last : input.current)?.focus({ preventScroll: true });
	});
	const add = () => {
		// Only an add that changes something draws the list again.
		moving.current = !full && typed.trim() !== "";
		onAdd();
	};
	const remove = (chip: string) => {
		moving.current = true;
		onRemove(chip);
	};
	return (
		<fieldset class="mt-form-field" aria-describedby={told.ids}>
			<legend class="mt-form-label">{label}</legend>
			{chips.length > 0 && (
				<ul class="mt-chips" ref={list}>
					{chips.map((chip) => (
						<li key={chip} class="mt-chip mt-chip-removable">
							<span class="mt-chip-text">{chip}</span>
							<button
								type="button"
								class="mt-chip-remove"
								aria-label={removeLabel(chip)}
								onClick={() => remove(chip)}
							>
								<Icon name="cross" size={14} />
							</button>
						</li>
					))}
				</ul>
			)}
			<div class="mt-add">
				<label for={id} class="mt-vh">
					{typeLabel}
				</label>
				<input
					id={id}
					ref={input}
					type="text"
					class="mt-input"
					value={typed}
					disabled={full}
					onInput={(event) =>
						onType(withinLength(event.currentTarget.value, longest))
					}
					onKeyDown={(event) => {
						if (event.key === "Enter" && !event.isComposing) {
							event.preventDefault();
							add();
						}
					}}
				/>
				<Button disabled={full || typed.trim() === ""} onClick={add}>
					{addLabel}
				</Button>
			</div>
			{told.lines}
		</fieldset>
	);
}

/**
 * CheckGroup is a list of statements under its name, each ticked on its own:
 * the names of what is ticked are value.
 */
export function CheckGroup({
	legend,
	choices,
	value,
	problem,
	onChange,
}: {
	legend: string;
	choices: readonly Choice[];
	value: readonly string[];
	problem?: string;
	onChange: (value: string[]) => void;
}) {
	const id = useId();
	const told = toldUnder(id, { problem });
	const ticked = new Set(value);
	return (
		<fieldset class="mt-form-field" aria-describedby={told.ids}>
			<legend class="mt-form-label">{legend}</legend>
			<div class="mt-checks">
				{choices.map((choice) => (
					<Checkbox
						key={choice.value}
						label={choice.label}
						checked={ticked.has(choice.value)}
						onChange={(checked) =>
							onChange(
								checked
									? [...value, choice.value]
									: value.filter((named) => named !== choice.value),
							)
						}
					/>
				))}
			</div>
			{told.lines}
		</fieldset>
	);
}

/** SwitchOption is one view a switch offers: what it is, and its name. */
export type SwitchOption<Value extends string> = {
	value: Value;
	label: string;
};

/**
 * ViewSwitch is a choice of one of a few views of the same thing, drawn as a
 * row of segments with the one chosen pressed in. Its options are radio
 * buttons under a legend a screen reader alone hears: a view gives no answer
 * and sends nothing, so the arrows that move through a group of radio buttons
 * change nothing anyone would regret, and the group is one stop of the
 * keyboard, as it should be.
 */
export function ViewSwitch<Value extends string>({
	legend,
	options,
	value,
	onChange,
}: {
	legend: string;
	options: readonly SwitchOption<Value>[];
	value: Value;
	onChange: (value: Value) => void;
}) {
	const name = useId();
	return (
		<fieldset class="mt-switch">
			<legend class="mt-vh">{legend}</legend>
			{options.map((option) => (
				<label key={option.value} class="mt-switch-option">
					<input
						type="radio"
						class="mt-switch-input"
						name={name}
						value={option.value}
						checked={option.value === value}
						onChange={() => onChange(option.value)}
					/>
					<span class="mt-switch-label">{option.label}</span>
				</label>
			))}
		</fieldset>
	);
}
