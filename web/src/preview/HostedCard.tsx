import {
	AppBridge,
	PostMessageTransport,
} from "@modelcontextprotocol/ext-apps/app-bridge";
import { useEffect, useRef, useState } from "preact/hooks";
import { toolInfoOf } from "../widget/testing/host";
import type { Scene } from "./scenes";

/**
 * HostedCard is the widget's own page in a frame, driven the way a chat host
 * drives it: through the library's host side, over the messages between the
 * frame and this page. It is told the host's context, handed the scene's payload
 * — or, for a task caught being asked for, told the call as far as it got —
 * answered as the service would answer, its pages opened in a tab of their own
 * unless the scene says the host opens none or refuses them, and pressed as the
 * child would press.
 * A frame whose page scrolls sideways is marked: the card has to fit any width
 * from a phone's up.
 */
export function HostedCard({
	scene,
	theme,
	locale,
	width,
}: {
	scene: Scene;
	theme: "light" | "dark";
	locale: string;
	width: number;
}) {
	const frame = useRef<HTMLIFrameElement>(null);
	const [height, setHeight] = useState(160);
	const [overflows, setOverflows] = useState(false);

	useEffect(() => {
		const view = frame.current?.contentWindow;
		if (frame.current === null || view === null || view === undefined) {
			return;
		}
		const host = new AppBridge(
			null,
			{ name: "preview", version: "1.0.0" },
			scene.links === "none" ? {} : { openLinks: {} },
			{
				hostContext: {
					theme,
					locale,
					displayMode: "inline",
					platform: width < 640 ? "mobile" : "web",
					containerDimensions: { width, maxHeight: 5000 },
					...(scene.caught !== undefined && {
						toolInfo: toolInfoOf("next_task"),
					}),
					safeAreaInsets: scene.insets ?? {
						top: 0,
						right: 0,
						bottom: 0,
						left: 0,
					},
				},
			},
		);
		host.oncalltool = (params) =>
			scene.answers?.(params.name) ??
			Promise.reject(new Error("no service here"));
		host.onmessage = () =>
			Promise.resolve(scene.refuseMessages ? { isError: true } : {});
		host.onupdatemodelcontext = () => Promise.resolve({});
		host.onopenlink = ({ url }) => {
			if (scene.links === "refuse") {
				return Promise.resolve({ isError: true });
			}
			window.open(url, "_blank", "noopener,noreferrer");
			return Promise.resolve({});
		};
		host.addEventListener("sizechange", ({ height: drawn }) => {
			if (drawn !== undefined) {
				setHeight(drawn);
			}
			const page = view.document.documentElement;
			setOverflows(page.scrollWidth > page.clientWidth);
		});
		let taken = false;
		host.addEventListener("initialized", async () => {
			// A task caught being asked for is told its call's arguments — the
			// lesson asked for in the language the page is shown in — and as far
			// as the call got, and never its result.
			if (scene.caught !== undefined) {
				await host.sendToolInput({ arguments: { language: locale } });
				if (scene.caught === "cancelled") {
					await host.sendToolCancelled({});
				}
				return;
			}
			await host.sendToolInput({ arguments: {} });
			await host.sendToolResult({
				content: [],
				structuredContent: scene.payload as Record<string, unknown>,
			});
			whenDrawn(
				view.document,
				() => taken,
				() => scene.play?.(view.document),
			);
		});
		const loaded = frame.current;
		// The host listens before the page loads: the page starts the handshake
		// as soon as it runs.
		void host.connect(new PostMessageTransport(view, view)).then(() => {
			loaded.src = "/widget.html";
		});
		return () => {
			taken = true;
			void host.close();
		};
	}, [scene, theme, locale, width]);

	return (
		<iframe
			ref={frame}
			title={scene.name}
			style={{
				width: `${width}px`,
				height: `${height}px`,
				border: 0,
				outline: overflows ? "3px solid #d33" : "none",
				background: "transparent",
			}}
		/>
	);
}

// whenDrawn calls then once the card is drawn, which is when the child can
// first press anything — unless the frame is taken down first.
function whenDrawn(
	card: Document,
	takenDown: () => boolean,
	then: () => void,
): void {
	const check = () => {
		if (takenDown()) {
			return;
		}
		if (card.querySelector(".mt-widget") !== null) {
			then();
		} else {
			setTimeout(check, 50);
		}
	};
	check();
}
