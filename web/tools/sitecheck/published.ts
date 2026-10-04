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
	"/en/topics/#logic",
	"/en/topics/#counting",
	"/en/topics/ordering/",
	"/en/topics/ordering/#traps",
	"/en/topics/ordering/#home",
	"/en/topics/knights-and-liars/",
	"/en/topics/knights-and-liars/#traps",
	"/en/topics/knights-and-liars/#home",
	"/en/topics/enumeration/",
	"/en/topics/enumeration/#traps",
	"/en/topics/enumeration/#home",
	"/en/topics/pigeonhole-principle/",
	"/en/topics/pigeonhole-principle/#traps",
	"/en/topics/pigeonhole-principle/#home",
	"/en/topics/overlapping-groups/",
	"/en/topics/overlapping-groups/#traps",
	"/en/topics/overlapping-groups/#home",
	"/ru/",
	"/ru/privacy/",
	"/ru/terms/",
	"/ru/topics/",
	"/ru/topics/#logic",
	"/ru/topics/#counting",
	"/ru/topics/ordering/",
	"/ru/topics/ordering/#traps",
	"/ru/topics/ordering/#home",
	"/ru/topics/knights-and-liars/",
	"/ru/topics/knights-and-liars/#traps",
	"/ru/topics/knights-and-liars/#home",
	"/ru/topics/enumeration/",
	"/ru/topics/enumeration/#traps",
	"/ru/topics/enumeration/#home",
	"/ru/topics/pigeonhole-principle/",
	"/ru/topics/pigeonhole-principle/#traps",
	"/ru/topics/pigeonhole-principle/#home",
	"/ru/topics/overlapping-groups/",
	"/ru/topics/overlapping-groups/#traps",
	"/ru/topics/overlapping-groups/#home",
];
