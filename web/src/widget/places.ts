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
 * regionName is the name in English of a region of the country, or undefined
 * for a code that is none of that country's regions — a region of another
 * country among them, which a file edited by hand can pair with any country.
 * Only the list's own codes are looked at, never what every object has.
 */
export function regionName(
	country: string,
	region: string,
): string | undefined {
	const named = Object.hasOwn(regions, country) ? regions[country] : undefined;
	return named !== undefined && Object.hasOwn(named, region)
		? named[region]
		: undefined;
}
