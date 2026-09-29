import type { ComponentChildren, Ref } from "preact";
import { classes } from "./classes";
import { Avatar, Icon, Mark } from "./icons";

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
			<Avatar size={24} />
			<span class="mt-bar-name">{props.name}</span>{" "}
			<span class="mt-bar-action">{props.action}</span>
			<Icon name="chevron-right" size={16} className="mt-chevron" />
		</button>
	);
}

/** Badge is a short label in a pill: the child's grade in a card's header. */
export function Badge({ children }: { children: ComponentChildren }) {
	return <span class="mt-badge">{children}</span>;
}

/**
 * MessageHeader says who speaks — MathTrail, with its mark, or the child, with
 * the avatar — with a badge under the name, or beside it on a wide card, and a
 * short note beside the name when there is one. A compact header heads a
 * reply below the task.
 */
export function MessageHeader({
	author = "app",
	name,
	badge,
	meta,
	compact = false,
	wide = false,
}: {
	author?: "app" | "person";
	name: string;
	badge?: string;
	meta?: string;
	compact?: boolean;
	wide?: boolean;
}) {
	const picture =
		author === "app" ? (
			<Mark size={compact ? 28 : 32} />
		) : (
			<Avatar size={compact ? 24 : 32} />
		);
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
	if (compact) {
		text = <div class="mt-head-text">{nameLine}</div>;
	} else if (wide) {
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
				compact && "mt-head-compact",
				wide && "mt-head-wide",
			)}
		>
			{picture}
			{text}
		</header>
	);
}

/**
 * ReplyCard is a reply below the task — MathTrail's result, or the child's
 * question — under a compact header.
 */
export function ReplyCard({
	author,
	name,
	meta,
	children,
}: {
	author: "app" | "person";
	name: string;
	meta?: string;
	children: ComponentChildren;
}) {
	return (
		<article class="mt-reply">
			<MessageHeader author={author} name={name} meta={meta} compact />
			<div class="mt-reply-body">{children}</div>
		</article>
	);
}
