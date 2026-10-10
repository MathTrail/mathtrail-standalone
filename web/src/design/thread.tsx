import type { ComponentChildren, Ref } from "preact";
import { classes } from "./classes";
import { Icon, Mark } from "./icons";

/**
 * ThreadBar is the line at the top of a card. Over a task it says whose card
 * it is and leads to the progress; over the progress it leads back to the
 * task.
 */
export function ThreadBar(
	props: (
		| { variant?: "person"; name: string; action: string }
		| { variant: "back"; label: string }
	) & { onClick: () => void; buttonRef?: Ref<HTMLButtonElement> },
) {
	if (props.variant === "back") {
		return (
			<button
				type="button"
				ref={props.buttonRef}
				class="mt-bar mt-bar-back"
				onClick={props.onClick}
			>
				<Icon name="chevron-left" size={16} className="mt-chevron" />
				<span>{props.label}</span>
			</button>
		);
	}
	return (
		<button
			type="button"
			ref={props.buttonRef}
			class="mt-bar"
			onClick={props.onClick}
		>
			<Icon name="avatar" size={24} />
			<span class="mt-bar-name">{props.name}</span>{" "}
			<span class="mt-bar-action">{props.action}</span>
			<Icon name="chevron-right" size={16} className="mt-chevron" />
		</button>
	);
}

/**
 * NameBar is the line at the top of a card that leads nowhere: whose card it
 * is, on a card drawn where there is no progress to open.
 */
export function NameBar({ name }: { name: string }) {
	return (
		<div class="mt-bar">
			<Icon name="avatar" size={24} />
			<span class="mt-bar-name">{name}</span>
		</div>
	);
}

/** Badge is a short label in a pill: the child's grade in a card's header. */
export function Badge({ children }: { children: ComponentChildren }) {
	return <span class="mt-badge">{children}</span>;
}

/**
 * Version is what a header says of the build a card came with: the word for a
 * version, in the card's language, and the number.
 */
export type Version = { label: string; number: string };

/**
 * MessageHeader says who speaks — MathTrail, with its logo, or the child, with
 * the avatar — with a badge under the name, or beside it on a wide card, and a
 * short note beside the name when there is one. A version, when there is one,
 * closes the line at its far end.
 */
export function MessageHeader({
	author = "app",
	name,
	badge,
	meta,
	version,
	wide = false,
}: {
	author?: "app" | "person";
	name: string;
	badge?: string;
	meta?: string;
	version?: Version;
	wide?: boolean;
}) {
	const picture =
		author === "app" ? <Mark size={32} /> : <Icon name="avatar" size={32} />;
	const nameLine = (
		<>
			<span class="mt-name">{name}</span>
			{meta !== undefined && (
				<>
					{" "}
					<span class="mt-meta">{meta}</span>
				</>
			)}
		</>
	);
	let text: ComponentChildren;
	if (wide) {
		text = (
			<div class="mt-head-text">
				<span class="mt-name">{name}</span>
				{badge !== undefined && <Badge>{badge}</Badge>}
				{meta !== undefined && <span class="mt-meta">{meta}</span>}
			</div>
		);
	} else {
		text = (
			<div class="mt-head-text">
				<div class="mt-head-line">{nameLine}</div>
				{badge !== undefined && <Badge>{badge}</Badge>}
			</div>
		);
	}
	return (
		<header
			class={classes(
				"mt-head",
				wide && "mt-head-wide",
				version !== undefined && "mt-head-versioned",
			)}
		>
			{picture}
			{text}
			{version !== undefined && (
				// The word runs the way the card's language does; the number is
				// Latin letters and digits in any language, and reads left to
				// right inside a card that reads the other way. The version is
				// for a grown-up reading the card off the screen, so a screen
				// reader does not read it out to the child before every task.
				<span class="mt-version" aria-hidden="true">
					<span class="mt-version-label">{version.label}</span>{" "}
					<span class="mt-version-number" dir="ltr">
						{version.number}
					</span>
				</span>
			)}
		</header>
	);
}
