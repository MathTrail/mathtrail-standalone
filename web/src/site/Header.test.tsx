import { describe, expect, test } from "vitest";
import { inPickerOrder, type LanguageLink } from "./Header";

// offered is a language the page is offered in, under its own name.
function offered(locale: string, name: string): LanguageLink {
	return { locale, href: `/${locale}/`, name, current: false };
}

describe("the language picker", () => {
	test("lists the languages by their own names: the Latin ones first, then the Cyrillic, the Arabic and the Han", () => {
		const languages = [
			offered("zh-Hans", "简体中文"),
			offered("ar", "العربية"),
			offered("ja", "日本語"),
			offered("ru", "Русский"),
			offered("pt", "Português"),
			offered("fr", "Français"),
			offered("es", "Español"),
			offered("en", "English"),
			offered("de", "Deutsch"),
		];

		expect(inPickerOrder(languages).map(({ locale }) => locale)).toEqual([
			"de",
			"en",
			"es",
			"fr",
			"pt",
			"ru",
			"ar",
			"ja",
			"zh-Hans",
		]);
	});
});
