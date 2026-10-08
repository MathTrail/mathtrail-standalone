import type { ComponentChildren } from "preact";

/**
 * Phone is a phone drawn around a chat, as a parent holds one: a dark edge
 * around a screen with the status bar on top, the chat in the middle, and the
 * field a message is typed in and the bar that leads home at the bottom. The
 * chat scrolls inside the screen. All but the chat is a drawing, which a
 * screen reader skips.
 */
export function Phone({ children }: { children: ComponentChildren }) {
	return (
		<div class="s-phone">
			<div class="s-phone-screen">
				<div class="s-phone-status" aria-hidden="true">
					<span>9:41</span>
					<span class="s-phone-island" />
					<span class="s-phone-system">
						<span class="s-phone-signal">
							<span class="s-phone-bar" />
							<span class="s-phone-bar" />
							<span class="s-phone-bar" />
							<span class="s-phone-bar" />
						</span>
						<span class="s-phone-battery" />
					</span>
				</div>
				{children}
				<div class="s-phone-composer" aria-hidden="true">
					<span class="s-phone-field" />
					<span class="s-phone-send">↑</span>
				</div>
				<div class="s-phone-home" aria-hidden="true" />
			</div>
		</div>
	);
}
