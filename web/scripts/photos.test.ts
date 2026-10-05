// @vitest-environment node
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterAll, beforeAll, describe, expect, test } from "vitest";
import { photoDirectory, photoPath, photos } from "../src/site/brand.ts";
import { cropOf, originalOf, pictureOnly } from "./photos.ts";
import { keptPhotoOf } from "./prerender-site.ts";

const repository = join(import.meta.dirname, "..", "..");

// Chunk is one chunk of a WebP file: its name and what it holds.
type Chunk = { name: string; data: Buffer };

// chunksOf are the chunks of a WebP file after its header, in order.
function chunksOf(webp: Buffer): Chunk[] {
	const chunks: Chunk[] = [];
	for (let at = 12; at + 8 <= webp.length; ) {
		const size = webp.readUInt32LE(at + 4);
		chunks.push({
			name: webp.toString("latin1", at, at + 4),
			data: webp.subarray(at + 8, at + 8 + size),
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
				chunks.map(({ name }) => name),
				photo.name,
			).toEqual(["VP8 "]);
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
});

describe("the originals of the photographs", () => {
	let dir = "";

	beforeAll(async () => {
		dir = await mkdtemp(join(tmpdir(), "originals-"));
		for (const file of ["dad.JPG", "mum.png", "mum.txt", "family.png"]) {
			await writeFile(join(dir, file), "");
		}
		await writeFile(join(dir, "family.webp"), "");
	});

	afterAll(async () => {
		await rm(dir, { recursive: true, force: true });
	});

	test("are found by the photograph's name, whatever kind of picture each is", async () => {
		expect(await originalOf(dir, "dad")).toBe(join(dir, "dad.JPG"));
		expect(await originalOf(dir, "mum")).toBe(join(dir, "mum.png"));
	});

	test("are refused when a photograph has none, or more than one", async () => {
		await expect(originalOf(dir, "older-son")).rejects.toThrow(
			"no original of older-son",
		);
		await expect(originalOf(dir, "family")).rejects.toThrow(
			"2 originals of family",
		);
	});
});
