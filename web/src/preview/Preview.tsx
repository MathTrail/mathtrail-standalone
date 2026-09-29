import { useMemo, useState } from "preact/hooks";
import { HostedCard } from "./HostedCard";
import { type Language, scenesIn } from "./scenes";

// The widths a card is looked at in: the narrowest it has to fit, the
// design's narrow card, the design's wide card, and a chat on the web.
const widths = [320, 360, 640, 736] as const;

// The host's locale for each language the preview shows.
const locales: Record<Language, string> = { en: "en-US", ru: "ru-RU" };

// asked is what the page's address asks to be shown first — ?scene=wrong,
// ?theme=dark, ?lang=ru, ?widths=320,640 — so that one view can be opened, or
// sent, as a link.
const asked = new URLSearchParams(location.search);

/**
 * Preview is the scenes of a lesson on a card, each at the widths chosen, in
 * the theme and the language chosen — to be set beside the approved design
 * and looked at. Every scene at once is many frames; one scene at a time is
 * quicker.
 */
export function Preview() {
	const [theme, setTheme] = useState<"light" | "dark">(
		asked.get("theme") === "dark" ? "dark" : "light",
	);
	const [language, setLanguage] = useState<Language>(
		asked.get("lang") === "ru" ? "ru" : "en",
	);
	const [shown, setShown] = useState<ReadonlySet<number>>(
		new Set(
			(asked.get("widths") ?? "320,360,640")
				.split(",")
				.map(Number)
				.filter((width) => widths.some((known) => known === width)),
		),
	);
	const [only, setOnly] = useState(asked.get("scene") ?? "");
	const scenes = useMemo(() => scenesIn(language), [language]);
	const toggle = (width: number) => {
		const next = new Set(shown);
		if (!next.delete(width)) {
			next.add(width);
		}
		setShown(next);
	};

	return (
		<main class={`preview preview-${theme}`}>
			<header class="preview-bar">
				<strong>MathTrail widget</strong>
				<select
					aria-label="Scene"
					value={only}
					onChange={(event) => setOnly(event.currentTarget.value)}
				>
					<option value="">every scene</option>
					{scenes.map((scene) => (
						<option key={scene.name} value={scene.name}>
							{scene.name}
						</option>
					))}
				</select>
				<select
					aria-label="Theme"
					value={theme}
					onChange={(event) =>
						setTheme(event.currentTarget.value === "dark" ? "dark" : "light")
					}
				>
					<option value="light">light</option>
					<option value="dark">dark</option>
				</select>
				<select
					aria-label="Language"
					value={language}
					onChange={(event) =>
						setLanguage(event.currentTarget.value === "ru" ? "ru" : "en")
					}
				>
					<option value="en">en</option>
					<option value="ru">ru</option>
				</select>
				{widths.map((width) => (
					<label key={width}>
						<input
							type="checkbox"
							checked={shown.has(width)}
							onChange={() => toggle(width)}
						/>
						{width}px
					</label>
				))}
			</header>
			{scenes
				.filter((scene) => only === "" || scene.name === only)
				.map((scene) => (
					<section key={scene.name} class="preview-scene">
						<h2>{scene.name}</h2>
						<div class="preview-frames">
							{widths
								.filter((width) => shown.has(width))
								.map((width) => (
									<HostedCard
										key={`${scene.name}-${theme}-${language}-${width}`}
										scene={scene}
										theme={theme}
										locale={locales[language]}
										width={width}
									/>
								))}
						</div>
					</section>
				))}
		</main>
	);
}
