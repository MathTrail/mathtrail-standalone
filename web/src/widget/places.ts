import places from "../../../internal/domain/country/places.json";

/**
 * countryCodes are the countries a parent can say the family lives in, by
 * their ISO 3166-1 codes: the list the service holds a country to, read from
 * the same file, so that the form offers no code the service would refuse.
 */
export const countryCodes: readonly string[] = places.countries;

/** Region is a part of a country a profile may name, by code and by name. */
export type Region = { code: string; name: string };

const regions: Readonly<Record<string, Readonly<Record<string, string>>>> =
	places.regions;

/**
 * regionsOf are the regions of a country a parent can name — for the United
 * States, its states — with their names in English, as the list writes them;
 * none for a country the list names no regions of.
 */
export function regionsOf(country: string): readonly Region[] {
	return Object.entries(regions[country] ?? {}).map(([code, name]) => ({
		code,
		name,
	}));
}

/**
 * regionName is the name of a region in English, or undefined for a code the
 * list does not have.
 */
export function regionName(region: string): string | undefined {
	for (const named of Object.values(regions)) {
		const name = named[region];
		if (name !== undefined) {
			return name;
		}
	}
	return undefined;
}
