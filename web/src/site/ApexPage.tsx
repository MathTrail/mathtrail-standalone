import { siteName } from "./brand";
import { type Head, PageHead } from "./Layout";

/**
 * ApexPage is the page the bare domain serves: it sends the reader on at once
 * to the front page at to, whose header switches the language. The host serves
 * files and answers no redirect of its own, so the page refreshes to the front
 * page, which needs no script. A refresh comes due once the page has loaded,
 * so it loads no stylesheet, no font and no picture: it is gone as soon as it
 * is read, and its icon, which a browser fetches aside, keeps nothing waiting.
 * Its one link is all a reader whose browser holds the refresh back needs. Its
 * head says what the front page says, for a chat or a directory that previews
 * the bare domain without following the refresh, and names the front page as
 * the page it stands for.
 */
export function ApexPage({ head, to }: { head: Head; to: string }) {
	return (
		<html lang={head.lang} dir={head.dir}>
			<head>
				<meta charset="utf-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1" />
				<meta http-equiv="refresh" content={`0; url=${to}`} />
				<PageHead head={head} />
			</head>
			<body>
				<a href={to}>{siteName}</a>
			</body>
		</html>
	);
}
