import { vi } from "vitest";

/**
 * onAMachineSpeaking has the platform's formats behave as they do on a
 * machine whose own language is language: asked for languages none of which
 * they have data for, they count, write numbers, list, name and sort in that
 * one. vi.restoreAllMocks puts them back.
 */
export function onAMachineSpeaking(language: string): void {
	const withOwn = (locales: Intl.LocalesArgument) => [
		...[locales ?? []].flat(),
		language,
	];
	const { Collator, DisplayNames, ListFormat, NumberFormat, PluralRules } =
		Intl;
	vi.spyOn(Intl, "Collator").mockImplementation(
		class extends Collator {
			constructor(
				locales?: Intl.LocalesArgument,
				options?: Intl.CollatorOptions,
			) {
				super(withOwn(locales), options);
			}
		},
	);
	vi.spyOn(Intl, "DisplayNames").mockImplementation(
		class extends DisplayNames {
			constructor(
				locales: Intl.LocalesArgument,
				options: Intl.DisplayNamesOptions,
			) {
				super(withOwn(locales), options);
			}
		},
	);
	vi.spyOn(Intl, "ListFormat").mockImplementation(
		class extends ListFormat {
			constructor(
				locales?: Intl.LocalesArgument,
				options?: Intl.ListFormatOptions,
			) {
				super(withOwn(locales), options);
			}
		},
	);
	vi.spyOn(Intl, "NumberFormat").mockImplementation(
		class extends NumberFormat {
			constructor(
				locales?: Intl.LocalesArgument,
				options?: Intl.NumberFormatOptions,
			) {
				super(withOwn(locales), options);
			}
		},
	);
	vi.spyOn(Intl, "PluralRules").mockImplementation(
		class extends PluralRules {
			constructor(
				locales?: Intl.LocalesArgument,
				options?: Intl.PluralRulesOptions,
			) {
				super(withOwn(locales), options);
			}
		},
	);
}
