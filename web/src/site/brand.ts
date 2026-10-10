/** siteName is the product's name, as the header and a sharing preview spell it. */
export const siteName = "MathTrail";

/** sourceURL is where the product's code and content are published. */
export const sourceURL = "https://github.com/MathTrail/mathtrail-standalone";

/**
 * connectorURL is the address a chat connects MathTrail by: the service's MCP
 * endpoint, which a parent pastes into the chat's connectors.
 */
export const connectorURL = "https://mcp.mathtrail.app/mcp";

/**
 * claudeConnectorsURL is the page of Claude's settings where a parent adds a
 * connector of their own.
 */
export const claudeConnectorsURL = "https://claude.ai/customize/connectors";

/**
 * chatGPTDeveloperModeURL is where OpenAI tells how ChatGPT's developer mode
 * adds an app by its address.
 */
export const chatGPTDeveloperModeURL =
	"https://developers.openai.com/apps-sdk/deploy/connect-chatgpt";

/** issuesURL is where anybody reports a mistake or proposes an idea, in public. */
export const issuesURL = `${sourceURL}/issues`;

/**
 * contactAddress is where questions about the product and a child's data go:
 * the address the privacy policy names.
 */
export const contactAddress = "altedtech.info@gmail.com";

/**
 * markPath is where the product's logo is served. One file is both the icon a
 * browser shows on the tab and the mark beside the name in the header, so the
 * two cannot drift apart. The address never changes: an icon shown away from
 * the site is kept by its address and fetched again only seldom, so a new
 * address would leave the old icon showing.
 */
export const markPath = "/assets/favicon.svg";

/**
 * coachPrototypePath is where the prototype of the coach is served: the coach
 * is a product of its own, still in the making, and its prototype is the
 * mockup a reader can try, the export of the tool it was drawn in. The page
 * about the coach frames it, and it is kept in the site's own assets under the
 * same name.
 */
export const coachPrototypePath = "/assets/coach-prototype.html";

/**
 * coachScreenPath is where a screen of the coach's prototype is served: the
 * child choosing a companion, as the author photographed it for the first
 * screen of the page about the coach. It is kept in the site's own assets
 * under the same name.
 */
export const coachScreenPath = "/assets/coach-screen.png";

/**
 * coachScreenSize is the size of the prototype's screen, in pixels, which the
 * page states so that a browser keeps its room before it arrives.
 */
export const coachScreenSize = { width: 909, height: 540 } as const;

/**
 * sharingPicturePath is where the picture a shared link to a page of locale
 * shows is served, one for each language of the site. It is kept in the
 * site's own assets under the same name.
 */
export function sharingPicturePath(locale: string): string {
	return `/assets/og-${locale}.png`;
}

/** sharingPictureSize is the size of a sharing picture, in pixels. */
export const sharingPictureSize = { width: 1200, height: 630 } as const;

/**
 * photoDirectory is where the site's photographs are served from, and the
 * directory of the site's own assets that keeps them. They are the family's
 * own: the licence of the code does not cover them, and no crawler is let in.
 */
export const photoDirectory = "/assets/photos/";

/** photoPath is where the photograph called name is served. */
export function photoPath(name: string): string {
	return `${photoDirectory}${name}.webp`;
}

/**
 * Photo is one of the site's photographs: the name it is served by, and the
 * size it is kept at, in pixels, which a page states so that a browser keeps
 * its room before it arrives.
 */
export type Photo = {
	readonly name: string;
	readonly width: number;
	readonly height: number;
};

/**
 * familyPhoto is the four of the family together, beside the first words of
 * the page "About": 3:4, as it was taken.
 */
export const familyPhoto: Photo = { name: "family", width: 900, height: 1200 };

/**
 * familyMembers are the family who make MathTrail, in the order the page
 * "About" shows their cards, each by the name their words and their portrait
 * go by: their place in the family, never their own name.
 */
export const familyMembers = [
	"mum",
	"dad",
	"older-son",
	"younger-son",
] as const;

/** portraitSize is the size of each of the family's portraits: 4:5. */
export const portraitSize = { width: 960, height: 1200 } as const;

/** photos are every photograph of the site: the family's, then each portrait. */
export const photos: readonly Photo[] = [
	familyPhoto,
	...familyMembers.map((name) => ({ name, ...portraitSize })),
];

/**
 * historyDirectory is where the pictures of the history the page "Why" tells
 * are served from, and the directory of the site's own assets that keeps
 * them: old paintings, a few photographs under licences of their own, which
 * their captions name, and screens of the coach's prototype.
 */
export const historyDirectory = "/assets/history/";

/** historyPicturePath is where the picture of the history called name is served. */
export function historyPicturePath(name: string): string {
	return `${historyDirectory}${name}.webp`;
}

/** historyPictureSize is the size every picture of the history is kept at: 4:3. */
export const historyPictureSize = { width: 1200, height: 900 } as const;

/**
 * profileFileName is the name of the file in a parent's Google Drive that
 * holds the child's profile, as the service writes it.
 */
export const profileFileName = "mathtrail-profile.json";
