import type { ComponentChildren } from "preact";

/**
 * SwitchOption is one choice of a switch: the value of its radio button, its
 * words, and whether it is the one chosen when the page opens.
 */
export type SwitchOption = {
	readonly value: string;
	readonly id?: string;
	readonly label: ComponentChildren;
	readonly chosen: boolean;
};

/**
 * ServiceSwitch is a choice of one of a few views of the same part of the
 * page, drawn as a row of segments with the one chosen filled in. Its options
 * are radio buttons under a legend a screen reader alone hears, as the card's
 * own switch has them: a view gives no answer and sends nothing, and what each
 * one shows is in the page's markup already, shown or hidden by the stylesheet
 * as the buttons are checked, with no script.
 */
export function ServiceSwitch({
	legend,
	name,
	options,
	small = false,
}: {
	legend: ComponentChildren;
	name: string;
	options: readonly SwitchOption[];
	small?: boolean;
}) {
	return (
		<fieldset
			class={
				small ? "s-service-switch s-service-switch-small" : "s-service-switch"
			}
		>
			<legend class="s-hidden">{legend}</legend>
			{options.map((option) => (
				<label key={option.value} class="s-service-option">
					<input
						type="radio"
						class="s-service-option-pick"
						name={name}
						value={option.value}
						id={option.id}
						checked={option.chosen}
					/>
					<span class="s-service-option-face">{option.label}</span>
				</label>
			))}
		</fieldset>
	);
}
