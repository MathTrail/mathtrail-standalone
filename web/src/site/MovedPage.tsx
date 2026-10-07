import { siteName } from "./brand";
import { type Head, PageHead } from "./Layout";

/**
 * MovedPage is what an address the site handed out serves once its page has
 * moved: it sends the reader on at once to the page at to. The host serves
 * files and answers no redirect of its own, so the page refreshes to where its
 * page is now, which needs no script. A refresh comes due once the page has
 * loaded, so it loads no stylesheet, no font and no picture: it is gone as
 * soon as it is read, and its icon, which a browser fetches aside, keeps
 * nothing waiting. Its one link is all a reader whose browser holds the
 * refresh back needs. Its head says what the page it stands for says, for a
 * chat or a directory that previews the old address without following the
 * refresh, and names that page's address as its own.
 */
export function MovedPage({ head, to }: { head: Head; to: string }) {
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
