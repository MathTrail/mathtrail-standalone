// The site's photographs, made from their originals: each cropped to the
// shape its page shows it in, scaled to the size it is kept at, and written as
// WebP by the browser's own encoder from a canvas, which holds the picture and
// nothing else — no place, no time, no camera. The originals stay with the
// family, outside the repository.
//
//	node scripts/photos.ts <directory of the originals>
//
// The directory holds an original of each photograph under the photograph's
// name — family.png, dad.jpg — as PNG, JPEG or WebP. It runs where the
// browsers are, inside the image of the pinned Playwright release, and writes
// over the photographs in site/assets/photos/, to be looked at before they are
// kept.

import { mkdir, readdir, readFile, writeFile } from "node:fs/promises";
import { basename, dirname, extname, join, parse } from "node:path";
import { chromium } from "playwright-core";
import { photoPath, photos } from "../src/site/brand.ts";
import { keptPhotoOf } from "./prerender-site.ts";

/** quality is how near a photograph's WebP stays to its original, from 0 to 1. */
export const quality = 0.8;

// originalTypes are the kinds of file an original may be, by its extension.
const originalTypes: Readonly<Record<string, string>> = {
	".jpeg": "image/jpeg",
	".jpg": "image/jpeg",
	".png": "image/png",
	".webp": "image/webp",
};

/** Size is a width and a height, in pixels. */
export type Size = { readonly width: number; readonly height: number };

/** Crop is the part of an original a photograph keeps, in the original's pixels. */
export type Crop = Size & { readonly x: number; readonly y: number };

/**
 * cropOf is the part of an original that a photograph of size keeps: the
 * largest part with the photograph's proportions, centred across and held to
 * the top, since faces are nearer a portrait's top than its foot. An original
 * smaller than its photograph is refused: blown up, it would only blur.
 */
export function cropOf(original: Size, size: Size): Crop {
	const scale = Math.min(
		original.width / size.width,
		original.height / size.height,
	);
	if (scale < 1) {
		throw new Error(
			`photos: an original of ${original.width} by ${original.height} is smaller than its photograph, ${size.width} by ${size.height}`,
		);
	}
	const width = Math.round(size.width * scale);
	const height = Math.round(size.height * scale);
	return { x: Math.floor((original.width - width) / 2), y: 0, width, height };
}

/**
 * originalOf is the original of the photograph called name in dir: the one
 * file of that name, of a kind an original may be.
 */
export async function originalOf(dir: string, name: string): Promise<string> {
	const found = (await readdir(dir)).filter((file) => {
		const { name: stem, ext } = parse(file);
		return stem === name && Object.hasOwn(originalTypes, ext.toLowerCase());
	});
	const [only] = found;
	if (only === undefined) {
		throw new Error(
			`photos: ${dir} has no original of ${name}: ${name}.png, ${name}.jpg or ${name}.webp`,
		);
	}
	if (found.length > 1) {
		throw new Error(
			`photos: ${dir} has ${found.length} originals of ${name}: ${found.join(", ")}`,
		);
	}
	return join(dir, only);
}

/**
 * pictureOnly is a WebP file with the picture alone: the one image chunk the
 * encoder wrote, under a header of its own, without the extended header and
 * whatever rides beside it — a colour profile, EXIF, XMP. A file whose picture
 * is not one image chunk is refused, since dropping the rest would change what
 * it shows.
 */
export function pictureOnly(webp: Buffer): Buffer {
	if (
		webp.toString("latin1", 0, 4) !== "RIFF" ||
		webp.toString("latin1", 8, 12) !== "WEBP"
	) {
		throw new Error("photos: the browser wrote no WebP");
	}
	const images: Buffer[] = [];
	for (let at = 12; at + 8 <= webp.length; ) {
		const name = webp.toString("latin1", at, at + 4);
		const size = webp.readUInt32LE(at + 4);
		// A chunk of an odd size is followed by a byte that pads it.
		const end = at + 8 + size + (size % 2);
		if (name === "VP8 " || name === "VP8L") {
			images.push(webp.subarray(at, end));
		} else if (name === "ALPH" || name === "ANIM" || name === "ANMF") {
			throw new Error(
				`photos: the picture has a chunk ${name}, which a photograph has no use for`,
			);
		}
		at = end;
	}
	const [image] = images;
	if (image === undefined || images.length > 1) {
		throw new Error(
			`photos: the browser wrote ${images.length} images where a photograph is one`,
		);
	}
	const header = Buffer.alloc(12);
	header.write("RIFF", 0, "latin1");
	header.writeUInt32LE(4 + image.length, 4);
	header.write("WEBP", 8, "latin1");
	return Buffer.concat([header, image]);
}

// measure is the size of the picture data holds, as the browser decodes it.
// It runs in the browser.
async function measure(data: string): Promise<Size> {
	const image = new Image();
	image.src = data;
	await image.decode();
	return { width: image.naturalWidth, height: image.naturalHeight };
}

// draw is the part crop of the picture data holds, scaled to width by height
// on a canvas with no transparency and encoded as WebP of quality: a data URL
// of the picture alone. It runs in the browser.
async function draw(job: {
	data: string;
	crop: Crop;
	width: number;
	height: number;
	quality: number;
}): Promise<string> {
	const image = new Image();
	image.src = job.data;
	await image.decode();
	const canvas = document.createElement("canvas");
	canvas.width = job.width;
	canvas.height = job.height;
	const context = canvas.getContext("2d", { alpha: false });
	if (context === null) {
		throw new Error("photos: the browser has no canvas to draw on");
	}
	context.imageSmoothingQuality = "high";
	const { crop } = job;
	context.drawImage(
		image,
		crop.x,
		crop.y,
		crop.width,
		crop.height,
		0,
		0,
		job.width,
		job.height,
	);
	return canvas.toDataURL("image/webp", job.quality);
}

// webpData is the prefix of a data URL that holds WebP, which a browser that
// cannot write WebP replaces with PNG's.
const webpData = "data:image/webp;base64,";

async function main(dir: string): Promise<void> {
	const browser = await chromium.launch();
	try {
		const page = await browser.newPage();
		for (const photo of photos) {
			const original = await originalOf(dir, photo.name);
			const type = originalTypes[extname(original).toLowerCase()];
			const data = `data:${type};base64,${(await readFile(original)).toString("base64")}`;
			const crop = cropOf(await page.evaluate(measure, data), photo);
			const drawn = await page.evaluate(draw, {
				data,
				crop,
				width: photo.width,
				height: photo.height,
				quality,
			});
			if (!drawn.startsWith(webpData)) {
				throw new Error("photos: the browser wrote no WebP");
			}
			const kept = keptPhotoOf(photo);
			await mkdir(dirname(kept), { recursive: true });
			await writeFile(
				kept,
				pictureOnly(Buffer.from(drawn.slice(webpData.length), "base64")),
			);
			console.log(
				`photos: ${photoPath(photo.name)} from ${basename(original)}`,
			);
		}
	} finally {
		await browser.close();
	}
}

if (import.meta.main) {
	const [dir] = process.argv.slice(2);
	if (dir === undefined) {
		console.error("usage: node scripts/photos.ts <directory of the originals>");
		process.exitCode = 2;
	} else {
		await main(dir);
	}
}
