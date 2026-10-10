import type { TopicArt } from "./art";
import { art as arithmeticWithATrick } from "./arithmetic-with-a-trick";
import { art as calendarAndAge } from "./calendar-and-age";
import { art as clocks } from "./clocks";
import { art as divisibilityAndRemainders } from "./divisibility-and-remainders";
import { art as enumeration } from "./enumeration";
import { art as figuresOnAGrid } from "./figures-on-a-grid";
import { art as gapsAndBoundaries } from "./gaps-and-boundaries";
import { art as knightsAndLiars } from "./knights-and-liars";
import { art as ordering } from "./ordering";
import { art as overlappingGroups } from "./overlapping-groups";
import { art as parityAndAlternation } from "./parity-and-alternation";
import { art as partsAndShares } from "./parts-and-shares";

/**
 * topicArt are the drawings of the topics' pages, by the slug of the topic. A
 * topic with none here draws its page's first screen from the site's data.
 */
export const topicArt: ReadonlyMap<string, TopicArt> = new Map([
	["arithmetic-with-a-trick", arithmeticWithATrick],
	["calendar-and-age", calendarAndAge],
	["clocks", clocks],
	["divisibility-and-remainders", divisibilityAndRemainders],
	["enumeration", enumeration],
	["figures-on-a-grid", figuresOnAGrid],
	["gaps-and-boundaries", gapsAndBoundaries],
	["knights-and-liars", knightsAndLiars],
	["ordering", ordering],
	["overlapping-groups", overlappingGroups],
	["parity-and-alternation", parityAndAlternation],
	["parts-and-shares", partsAndShares],
]);
