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
 * markPath is where the product's mark is served. One file is both the icon a
 * browser shows on the tab and the mark beside the name in the header, so the
 * two cannot drift apart.
 */
export const markPath = "/assets/favicon.svg";

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
