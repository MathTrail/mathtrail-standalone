import type { VNode } from "preact";
import { AboutPage } from "./AboutPage";
import { CoachPage } from "./CoachPage";
import { frontPage } from "./content";
import type { SiteData } from "./data";
import { HomePage } from "./HomePage";
import { ResearchPage } from "./ResearchPage";
import type { PageReader } from "./reader";
import { ServicePage } from "./ServicePage";
import { serviceRules } from "./service";
import { TechniquesPage } from "./TechniquesPage";
import { topicPage } from "./TopicPage";
import { TopicsPage, topicsStyle } from "./TopicsPage";
import { historyRules } from "./WhyHistory";
import { WhyPage } from "./WhyPage";

/**
 * PageProps are what a page's component draws: its words, in the page's
 * language, and the site's data, which no language changes.
 */
export type PageProps = { readonly page: PageReader; readonly data: SiteData };

/**
 * Page is a page a component draws: the component, which draws the part of the
 * page that is its own, between the site's header and its footer; whether it
 * draws a card of the widget or a picture as a card draws one, either of which
 * loads the card's stylesheet; and the rules of style its head carries, written
 * from the site's data when the site is built.
 */
export type Page = {
	readonly draw: (props: PageProps) => VNode;
	readonly card?: boolean;
	readonly style?: (data: SiteData) => string;
};

/**
 * sitePages are the pages a component draws, by the name their words' file
 * has in every language: index for index.yaml, the front page, why for
 * why.yaml, topics for topics.yaml, about for about.yaml, coach for coach.yaml,
 * the page of a product of its own still in the making, research for
 * research.yaml, the page of the numbers behind the product, service for
 * service.yaml, the page of how the service works, and topics/<slug>
 * for the page of a topic of the catalog, which one template draws for every
 * topic.
 * Which topics have a page is for their words to say, and for the catalog to
 * agree with; a file of words that no component draws stops the build.
 */
export function sitePages(data: SiteData): ReadonlyMap<string, Page> {
	return new Map<string, Page>([
		[frontPage, { draw: HomePage, card: true }],
		[
			"why",
			{
				draw: WhyPage,
				card: true,
				style: (data) =>
					data.why === undefined ? "" : historyRules(data.why.history),
			},
		],
		["topics", { draw: TopicsPage, card: true, style: topicsStyle }],
		["techniques", { draw: TechniquesPage }],
		["about", { draw: AboutPage }],
		["coach", { draw: CoachPage }],
		["research", { draw: ResearchPage }],
		["service", { draw: ServicePage, style: serviceRules }],
		...data.topics.all.map((topic): [string, Page] => [
			`topics/${topic.slug}`,
			topicPage(topic, data),
		]),
	]);
}
