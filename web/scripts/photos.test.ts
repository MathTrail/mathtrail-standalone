// @vitest-environment node
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import type { Page } from "playwright-core";
import { afterAll, beforeAll, describe, expect, test } from "vitest";
import {
	type Photo,
	photoDirectory,
	photoPath,
	photos,
} from "../src/site/brand.ts";
import {
	cropOf,
	drawerIn,
	makePhotos,
	originalOf,
	pictureOnly,
	quality,
} from "./photos.ts";
import { keptPhotoOf } from "./prerender-site.ts";

const repository = join(import.meta.dirname, "..", "..");

// Chunk is one chunk of a WebP file: its name, what it holds, and whether the
// file holds all of it, as much as its size says.
type Chunk = { name: string; data: Buffer; whole: boolean };

// chunksOf are the chunks of a WebP file after its header, in order.
function chunksOf(webp: Buffer): Chunk[] {
	const chunks: Chunk[] = [];
	for (let at = 12; at + 8 <= webp.length; ) {
		const size = webp.readUInt32LE(at + 4);
		chunks.push({
			name: webp.toString("latin1", at, at + 4),
			data: webp.subarray(at + 8, at + 8 + size),
			whole: at + 8 + size <= webp.length,
		});
		at += 8 + size + (size % 2);
	}
	return chunks;
}

// sizeOf is the size of the picture a lossy image chunk holds, as the frame's
// own header gives it: after its tag and its start code, fourteen bits each.
function sizeOf(image: Buffer): { width: number; height: number } {
	expect(image.subarray(3, 6).toString("hex")).toBe("9d012a");
	return {
		width: image.readUInt16LE(6) & 0x3fff,
		height: image.readUInt16LE(8) & 0x3fff,
	};
}

// chunk is a chunk of a WebP file holding data, with the byte that pads a
// chunk of an odd size.
function chunk(name: string, data: Buffer): Buffer {
	const head = Buffer.alloc(8);
	head.write(name, 0, "latin1");
	head.writeUInt32LE(data.length, 4);
	return Buffer.concat([head, data, Buffer.alloc(data.length % 2)]);
}

// webp is a WebP file of chunks.
function webp(...chunks: Buffer[]): Buffer {
	const body = Buffer.concat(chunks);
	const head = Buffer.alloc(12);
	head.write("RIFF", 0, "latin1");
	head.writeUInt32LE(4 + body.length, 4);
	head.write("WEBP", 8, "latin1");
	return Buffer.concat([head, body]);
}

describe("the site's photographs", () => {
	test("are kept in the site's assets, under the directory they are served from", () => {
		for (const photo of photos) {
			expect(photoPath(photo.name).startsWith(photoDirectory)).toBe(true);
			expect(keptPhotoOf(photo)).toBe(
				join(repository, "site", photoPath(photo.name)),
			);
		}
	});

	// A page states a photograph's size, and a browser that trusts it stretches
	// a photograph of another. Anything in the file beside the picture — where
	// and when it was taken, the camera, a colour profile — is something nobody
	// chose to publish.
	test("are WebP files that hold the picture alone, at the size the page states", async () => {
		for (const photo of photos) {
			const file = await readFile(keptPhotoOf(photo));

			expect(file.toString("latin1", 0, 4), photo.name).toBe("RIFF");
			expect(file.readUInt32LE(4) + 8, photo.name).toBe(file.length);
			expect(file.toString("latin1", 8, 12), photo.name).toBe("WEBP");
			const chunks = chunksOf(file);
			expect(
				chunks.map(({ name, whole }) => [name, whole]),
				photo.name,
			).toEqual([["VP8 ", true]]);
			expect(sizeOf(chunks[0]?.data ?? Buffer.alloc(0)), photo.name).toEqual({
				width: photo.width,
				height: photo.height,
			});
		}
	});
});

describe("making a photograph", () => {
	test("keeps the largest part of the original with the photograph's proportions, held to its top", () => {
		expect(
			cropOf({ width: 1086, height: 1448 }, { width: 960, height: 1200 }),
		).toEqual({ x: 0, y: 0, width: 1086, height: 1358 });
		expect(
			cropOf({ width: 1086, height: 1448 }, { width: 900, height: 1200 }),
		).toEqual({ x: 0, y: 0, width: 1086, height: 1448 });
	});

	test("centres the part it keeps across an original wider than the photograph", () => {
		expect(
			cropOf({ width: 2001, height: 1200 }, { width: 960, height: 1200 }),
		).toEqual({ x: 520, y: 0, width: 960, height: 1200 });
	});

	test("refuses an original smaller than its photograph, which would only blur", () => {
		expect(() =>
			cropOf({ width: 600, height: 800 }, { width: 960, height: 1200 }),
		).toThrow("smaller than its photograph");
	});

	test("leaves the picture alone, dropping whatever the encoder wrote beside it", () => {
		const image = chunk("VP8 ", Buffer.from("a frame"));
		const written = webp(
			chunk("VP8X", Buffer.alloc(10)),
			chunk("ICCP", Buffer.from("a profile")),
			image,
			chunk("EXIF", Buffer.from("where")),
			chunk("XMP ", Buffer.from("who")),
		);

		expect(pictureOnly(written)).toEqual(webp(image));
	});

	test("refuses a file whose picture is not one image", () => {
		const image = chunk("VP8 ", Buffer.from("a frame"));

		expect(() => pictureOnly(Buffer.from("a PNG, perhaps"))).toThrow("no WebP");
		expect(() => pictureOnly(webp(chunk("VP8X", Buffer.alloc(10))))).toThrow(
			"0 images",
		);
		expect(() => pictureOnly(webp(image, image))).toThrow("2 images");
		expect(() =>
			pictureOnly(webp(chunk("ALPH", Buffer.from("alpha")), image)),
		).toThrow("ALPH");
	});

	test("refuses a file whose image is cut short of the size its chunk states", () => {
		const whole = webp(chunk("VP8 ", Buffer.from("a frame")));

		expect(() => pictureOnly(whole.subarray(0, whole.length - 3))).toThrow(
			"cut short",
		);
	});
});

// pageAnswering is a page whose browser answers what it is asked to run with
// answers, one after another — the size it decodes an original at, then the
// picture it writes — and keeps what it was given each time, in order.
function pageAnswering(...answers: unknown[]) {
	const given: unknown[] = [];
	const page: { evaluate: (run: unknown, arg: unknown) => Promise<unknown> } = {
		evaluate: async (_, arg) => {
			given.push(arg);
			return answers.shift();
		},
	};
	return { page: page as unknown as Page, given };
}

// asData is a data URL of a WebP file, as a browser's canvas writes one.
const asData = (file: Buffer) =>
	`data:image/webp;base64,${file.toString("base64")}`;

describe("drawing a photograph in a browser", () => {
	const original = "data:image/jpeg;base64,b3JpZ2luYWw=";
	const photo: Photo = { name: "dad", width: 960, height: 1200 };
	// decoded is wider than the photograph's proportions, so the part kept
	// lies across its middle.
	const decoded = { width: 2001, height: 1500 };
	const image = chunk("VP8 ", Buffer.from("a frame"));

	test("asks the browser to draw the part of the original kept at the size it decodes, at the photograph's size and quality", async () => {
		const { page, given } = pageAnswering(decoded, asData(webp(image)));

		await drawerIn(page)(original, photo);

		expect(given).toEqual([
			original,
			{
				data: original,
				crop: cropOf(decoded, photo),
				width: 960,
				height: 1200,
				quality,
			},
		]);
	});

	test("keeps of what the browser writes the picture alone", async () => {
		const written = webp(
			chunk("VP8X", Buffer.alloc(10)),
			image,
			chunk("EXIF", Buffer.from("where")),
		);
		const { page } = pageAnswering(decoded, asData(written));

		expect(await drawerIn(page)(original, photo)).toEqual(webp(image));
	});

	// A browser that cannot write WebP writes PNG in its place, and says so
	// only by the type at the head of its data.
	test("refuses a picture the browser wrote as PNG", async () => {
		const { page } = pageAnswering(
			decoded,
			"data:image/png;base64,iVBORw0KGgo=",
		);

		await expect(drawerIn(page)(original, photo)).rejects.toThrow(
			"the browser wrote no WebP",
		);
	});

	test("refuses an original the browser decodes smaller than its photograph, and asks it to draw nothing", async () => {
		const { page, given } = pageAnswering({ width: 600, height: 800 });

		await expect(drawerIn(page)(original, photo)).rejects.toThrow(
			"smaller than its photograph",
		);
		expect(given).toEqual([original]);
	});
});

describe("the originals of the photographs", () => {
	// mixed holds originals of two photographs, one of them twice, and a file
	// no original is; every holds one of each photograph; lacking, every one
	// but the last.
	let mixed = "";
	let every = "";
	let lacking = "";
	const last = photos.at(-1)?.name ?? "";

	beforeAll(async () => {
		mixed = await mkdtemp(join(tmpdir(), "originals-"));
		for (const file of [
			"dad.JPG",
			"mum.png",
			"mum.txt",
			"family.png",
			"family.webp",
		]) {
			await writeFile(join(mixed, file), "");
		}
		every = await mkdtemp(join(tmpdir(), "originals-"));
		lacking = await mkdtemp(join(tmpdir(), "originals-"));
		for (const { name } of photos) {
			await writeFile(join(every, `${name}.png`), "");
			if (name !== last) {
				await writeFile(join(lacking, `${name}.png`), "");
			}
		}
	});

	afterAll(async () => {
		for (const dir of [mixed, every, lacking]) {
			await rm(dir, { recursive: true, force: true });
		}
	});

	test("are found by the photograph's name, whatever kind of picture each is", async () => {
		expect(await originalOf(mixed, "dad")).toBe(join(mixed, "dad.JPG"));
		expect(await originalOf(mixed, "mum")).toBe(join(mixed, "mum.png"));
	});

	test("are refused when a photograph has none, or more than one", async () => {
		await expect(originalOf(mixed, "older-son")).rejects.toThrow(
			"no original of older-son",
		);
		await expect(originalOf(mixed, "family")).rejects.toThrow(
			"2 originals of family",
		);
	});

	test("make every photograph before any is kept, each as it was drawn", async () => {
		const kept: [string, string][] = [];
		const made = await makePhotos(
			every,
			async (_, photo) => Buffer.from(`drawn ${photo.name}`),
			async (photo: Photo, file: Buffer) => {
				kept.push([photo.name, file.toString()]);
			},
		);

		expect(kept).toEqual(photos.map(({ name }) => [name, `drawn ${name}`]));
		expect(made).toEqual(
			photos.map(({ name }) => `${photoPath(name)} from ${name}.png`),
		);
	});

	// A run that stops on the last photograph would otherwise leave the kept
	// ones half new and half old, made from two sets of originals.
	test("keep none of the photographs when the last of them cannot be made", async () => {
		const kept: string[] = [];
		const keep = async (photo: Photo) => {
			kept.push(photo.name);
		};
		const drawn = async (_: string, photo: Photo) => {
			if (photo.name === last) {
				throw new Error(`${last} cannot be drawn`);
			}
			return Buffer.from("drawn");
		};

		await expect(makePhotos(every, drawn, keep)).rejects.toThrow(
			"cannot be drawn",
		);
		await expect(
			makePhotos(lacking, async () => Buffer.from("drawn"), keep),
		).rejects.toThrow(`no original of ${last}`);
		expect(kept).toEqual([]);
	});
});
