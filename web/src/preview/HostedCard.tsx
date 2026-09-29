import {
	AppBridge,
	PostMessageTransport,
} from "@modelcontextprotocol/ext-apps/app-bridge";
import { useEffect, useRef, useState } from "preact/hooks";
import type { Scene } from "./scenes";

/**
 * HostedCard is the widget's own page in a frame, driven the way a chat host
 * drives it: through the library's host side, over the messages between the
 * frame and this page. It is told the host's context, handed the scene's task,
 * answered as the service would answer, and pressed as the child would press.
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
			{},
			{
				hostContext: {
					theme,
					locale,
					displayMode: "inline",
					platform: width < 640 ? "mobile" : "web",
					containerDimensions: { width, maxHeight: 5000 },
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
		host.onmessage = async () =>
			scene.refuseMessages ? { isError: true } : {};
		host.onupdatemodelcontext = async () => ({});
		host.addEventListener("sizechange", ({ height: drawn }) => {
			if (drawn !== undefined) {
				setHeight(drawn);
			}
			const page = view.document.documentElement;
			setOverflows(page.scrollWidth > page.clientWidth);
		});
		let taken = false;
		host.addEventListener("initialized", async () => {
			await host.sendToolInput({ arguments: {} });
			await host.sendToolResult({
				content: [],
				structuredContent: scene.handed as unknown as Record<string, unknown>,
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

// whenDrawn calls then once the card has drawn its options, which is when the
// child can first press anything — unless the frame is taken down first.
function whenDrawn(
	card: Document,
	takenDown: () => boolean,
	then: () => void,
): void {
	const check = () => {
		if (takenDown()) {
			return;
		}
		if (card.querySelector(".mt-option") !== null) {
			then();
		} else {
			setTimeout(check, 50);
		}
	};
	check();
}
