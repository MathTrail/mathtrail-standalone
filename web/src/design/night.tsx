import { useScopedId } from "./ids";

// The night's own colours, the same in either theme, as a photograph's are:
// the sky, the ridges and the mountain, its snow and the stars, the trail up
// to the summit, the flag on it, and the shooting star with its glow. Each is
// the drawing's own attribute rather than a style, which a host's page may
// refuse.
const ink = {
	sky: "#0f1830",
	ridges: "#1c2c4c",
	mountain: "#2f4a78",
	snow: "#dfe6f2",
	trail: "#f0b040",
	pole: "#ecebe6",
	flag: "#3fbf86",
	star: "#fff6d8",
	glow: "#f0d27a",
} as const;

// The stars, each where it shines, how large and how bright: x, y, radius and
// opacity on the drawing's grid of 592 by 132.
const stars: readonly (readonly [number, number, number, number])[] = [
	[9.5, 33, 1.5, 0.61],
	[527, 73.5, 1, 0.8],
	[74.7, 46, 1.5, 0.57],
	[526.2, 86.6, 1.5, 0.6],
	[148.3, 52.3, 1, 0.77],
	[393.2, 47.1, 1, 0.73],
	[45.3, 24.5, 1.5, 0.67],
	[313.5, 43.1, 1, 0.83],
	[172.8, 13, 1, 0.8],
	[29.6, 78.9, 1, 0.39],
	[290.7, 15.5, 1, 0.64],
	[281.4, 36.4, 1, 0.81],
	[57.6, 70.8, 1, 0.83],
	[231.6, 79.5, 1, 0.71],
	[517.5, 30.4, 1, 0.39],
	[242.5, 73.5, 1.5, 0.71],
	[156.7, 11.9, 1, 0.64],
	[509.6, 81.6, 1, 0.67],
	[479.4, 24.9, 1, 0.56],
	[241.8, 36.2, 1, 0.65],
	[138, 39.7, 1.5, 0.7],
	[9.4, 20.8, 1, 0.85],
	[546.5, 12.1, 1.5, 0.43],
	[71.4, 21, 1.5, 0.47],
	[71.1, 70.3, 1, 0.5],
	[568.2, 10.4, 1, 0.41],
	[298.3, 62, 1, 0.63],
	[121.4, 33, 1, 0.58],
	[51.7, 44.2, 1, 0.35],
	[86.5, 62.1, 1, 0.69],
	[175.6, 56.3, 1, 0.37],
	[411.8, 23, 1, 0.83],
	[347.5, 10, 1, 0.66],
	[581.9, 5.2, 1.5, 0.72],
	[214.4, 58, 1, 0.65],
	[8.1, 87.1, 1, 0.35],
	[515.3, 70.4, 1, 0.37],
	[185.5, 6.6, 1, 0.74],
	[278.9, 13.3, 1, 0.38],
	[37.1, 18.7, 1, 0.68],
];

// The shooting star's tail, from its faint end to the head: each stretch
// wider and brighter than the one before it.
const tail: readonly (readonly [string, number, number])[] = [
	["M110 18 262 52", 1.6, 0.15],
	["M150 27 262 52", 1.8, 0.3],
	["M190 36 262 52", 2.2, 0.55],
	["M228 44.4 262 52", 2.6, 0.85],
];

/**
 * DayOver is what a card says once the day's tasks are over: a night over the
 * mountains, with a trail up to a flag on the summit and a shooting star to
 * make a wish on, and under it a wish for tomorrow and what happened. The
 * picture is decoration, the words say it all; it is night in either theme.
 */
export function DayOver({ title, detail }: { title: string; detail: string }) {
	return (
		<div class="mt-dayover">
			<NightTrail />
			<div class="mt-dayover-words">
				<h2 class="mt-dayover-title">{title}</h2>
				<p class="mt-dayover-line">{detail}</p>
			</div>
		</div>
	);
}

// NightTrail is the night itself, as high as the drawing and as wide as the
// card. It is cut, never squeezed: at its sides on a card narrower than the
// drawing, and above and below alike on a wider one. The summit, the flag and
// the star stand near its middle, so a card of any width a chat gives keeps
// them. Its glow is named apart from that of any other night on the page.
function NightTrail() {
	const glow = useScopedId();
	return (
		<div class="mt-night">
			<svg
				viewBox="0 0 592 132"
				preserveAspectRatio="xMidYMid slice"
				aria-hidden="true"
			>
				<defs>
					<radialGradient id={glow}>
						<stop offset="0" stop-color={ink.glow} stop-opacity="0.55" />
						<stop offset="0.45" stop-color={ink.glow} stop-opacity="0.18" />
						<stop offset="1" stop-color={ink.glow} stop-opacity="0" />
					</radialGradient>
				</defs>
				<rect width="592" height="132" fill={ink.sky} />
				{stars.map(([x, y, radius, opacity]) => (
					<circle
						key={`${x},${y}`}
						cx={x}
						cy={y}
						r={radius}
						fill={ink.snow}
						opacity={opacity}
					/>
				))}
				<polygon
					points="0,132 0,100 60,78 120,96 180,70 230,90 300,74 360,98 430,80 500,100 592,86 592,132"
					fill={ink.ridges}
				/>
				<polygon points="200,132 372,46 544,132" fill={ink.mountain} />
				<polygon
					points="372,46 351,57 361,55 372,62 383,55 393,57"
					fill={ink.snow}
				/>
				<path
					d="M330 132 404 110 352 90 392 72 362 58 372 48"
					fill="none"
					stroke={ink.trail}
					stroke-width="2.5"
					stroke-dasharray="6 6"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
				<path d="M372 46V24" stroke={ink.pole} stroke-width="2" />
				<polygon points="373,24 392,29 373,34" fill={ink.flag} />
				{tail.map(([d, width, opacity]) => (
					<path
						key={d}
						d={d}
						stroke={ink.star}
						stroke-width={width}
						stroke-linecap="round"
						opacity={opacity}
					/>
				))}
				<circle cx="262" cy="52" r="9" fill={`url(#${glow})`} />
				<circle cx="262" cy="52" r="3.2" fill={ink.star} />
			</svg>
		</div>
	);
}
