import type { ComponentChildren } from "preact";
import type { Alternate } from "./addresses";
import { markPath, sharingPictureSize, siteName } from "./brand";

/**
 * Head is what a page tells a browser, a search engine and a chat about
 * itself.
 */
export type Head = {
	readonly lang: string;
	readonly dir: "ltr" | "rtl";
	readonly title: string;
	readonly description: string;
	readonly canonical: string;
	readonly alternates: readonly Alternate[];
	/** image is the address of the picture a shared link to the page shows. */
	readonly image: string;
	/** imageAlt says what the picture shows, for whoever cannot see it. */
	readonly imageAlt: string;
};

/**
 * cardStylesheet is where the styles of the widget's cards are served: the
 * very styles a chat draws a card with, loaded only by a page that draws one.
 */
export const cardStylesheet = "/assets/card.css";

/**
 * Layout is a whole page: its head, and a body drawn from its children. The
 * page is always light, whatever the reader's system prefers: the site has one
 * look. It loads the design's tokens before its own styles, which read them,
 * then the cards' styles when it draws a card, then the rules of its own a
 * page's data writes, and it runs no script, so every page reads the same
 * with JavaScript off. The rules are the site's own, written when it is
 * built, and carried as they are.
 */
export function Layout({
	head,
	card = false,
	style,
	children,
}: {
	head: Head;
	card?: boolean;
	style?: string;
	children: ComponentChildren;
}) {
	return (
		<html lang={head.lang} dir={head.dir} data-theme="light">
			<head>
				<meta charset="utf-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1" />
				<meta name="color-scheme" content="light" />
				<title>{head.title}</title>
				<meta name="description" content={head.description} />
				<link rel="canonical" href={head.canonical} />
				{head.alternates.map(({ hreflang, url }) => (
					<link key={hreflang} rel="alternate" hreflang={hreflang} href={url} />
				))}
				<meta property="og:type" content="website" />
				<meta property="og:site_name" content={siteName} />
				<meta property="og:title" content={head.title} />
				<meta property="og:description" content={head.description} />
				<meta property="og:url" content={head.canonical} />
				<meta property="og:image" content={head.image} />
				<meta property="og:image:alt" content={head.imageAlt} />
				<meta
					property="og:image:width"
					content={String(sharingPictureSize.width)}
				/>
				<meta
					property="og:image:height"
					content={String(sharingPictureSize.height)}
				/>
				<meta name="twitter:card" content="summary_large_image" />
				<link rel="icon" href={markPath} type="image/svg+xml" />
				<link rel="stylesheet" href="/assets/tokens.css" />
				<link rel="stylesheet" href="/assets/style.css" />
				{card && <link rel="stylesheet" href={cardStylesheet} />}
				{style !== undefined && (
					<style dangerouslySetInnerHTML={{ __html: style }} />
				)}
			</head>
			<body>{children}</body>
		</html>
	);
}
