// published are the addresses the site has handed out: the consent screen of
// the sign-in points at two of them, and a listing points at the apex. A build
// that dropped one would break a link somebody else holds, and nothing on the
// site need still link to it for the link rule to notice. An anchor linked from
// outside its page is held the same way, written after its page's address, and
// its page must still hold an element with that id. The list is written by
// hand, address by address, and never worked out from the site's texts, so that
// a page going missing from them cannot take its address off the list.
export const published: readonly string[] = [
	"/",
	"/en/",
	"/en/privacy/",
	"/en/terms/",
	"/en/topics/",
	"/ru/",
	"/ru/privacy/",
	"/ru/terms/",
	"/ru/topics/",
];
