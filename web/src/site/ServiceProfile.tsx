import type { PageReader } from "./reader";
import { ServiceSection } from "./ServiceSection";
import { ServiceSwitch } from "./ServiceSwitch";
import {
	type Block,
	blocks,
	shareOf,
	sizeScale,
	titleId,
	totalOf,
} from "./service";
import { TableFrame } from "./TableFrame";

/**
 * ServiceProfile is the profile's file: its size, as a file usually has it
 * or near its caps, which a switch chooses; the bar of its blocks against the
 * size it was meant to keep within; and the blocks one by one — what each
 * holds, the tools that write it and its size. Both sizes are in the page,
 * and the stylesheet shows the one chosen.
 */
export function ServiceProfile({ page }: { page: PageReader }) {
	return (
		<ServiceSection page={page} id="profile">
			<div class="s-service-card s-service-profile">
				<div class="s-service-profile-head">
					<div>
						<p class="s-service-label">{page.text("profile.size")}</p>
						<p class="s-service-total">
							<Sizes
								page={page}
								words="profile.total"
								typical={totalOf("typical")}
								cap={totalOf("cap")}
							/>
						</p>
					</div>
					<ServiceSwitch
						legend={page.text("profile.legend")}
						name="service-size"
						small
						options={[
							{
								value: "typical",
								label: page.text("profile.typical"),
								chosen: true,
							},
							{ value: "cap", label: page.text("profile.cap"), chosen: false },
						]}
					/>
				</div>
				<Bar page={page} />
				<Blocks page={page} />
				<p class="s-service-small">{page.text("profile.note")}</p>
			</div>
		</ServiceSection>
	);
}

// Sizes are a size as a file usually has it and near its caps, in words, one
// after the other: the stylesheet shows the one the switch chose, and a page
// read without it says both.
function Sizes({
	page,
	words,
	typical,
	cap,
}: {
	page: PageReader;
	words: string;
	typical: number;
	cap: number;
}) {
	return (
		<>
			<span class="s-service-typical">{page.text(words, { kb: typical })}</span>
			<span class="s-service-or"> / </span>
			<span class="s-service-cap">{page.text(words, { kb: cap })}</span>
		</>
	);
}

// Bar is the blocks of the file side by side, each as wide as its share of the
// scale, against the size the file was meant to keep within and the scale's
// marks. It says nothing the table under it does not, so a screen reader
// skips it.
function Bar({ page }: { page: PageReader }) {
	const marks = Array.from(
		{ length: sizeScale.most / sizeScale.step + 1 },
		(_, at) => at * sizeScale.step,
	);
	return (
		<div class="s-service-bar" dir="ltr" aria-hidden="true">
			<div class="s-service-bar-track">
				{blocks.map((block) => (
					<span
						key={block.id}
						class="s-service-segment"
						title={page.plain("profile.segment", {
							name: page.plain(`profile.blocks.${block.id}.name`),
							typical: block.typical,
							cap: block.cap,
						})}
						style={{
							"--typical": shareOf(block.typical),
							"--cap": shareOf(block.cap),
							"--tone": `var(${block.tone})`,
						}}
					/>
				))}
			</div>
			<div
				class="s-service-goal"
				style={{ insetInlineStart: shareOf(sizeScale.goal) }}
			>
				<span>{page.text("profile.goal", { kb: sizeScale.goal })}</span>
			</div>
			<p class="s-service-scale">
				{marks.map((kb) => (
					<span key={kb}>
						{kb === 0 ? kb : page.text("profile.kb", { kb })}
					</span>
				))}
			</p>
		</div>
	);
}

// Blocks are the blocks of the file in a table: the colour of each in the bar,
// its name, what it holds, the tools that write it, and its size.
function Blocks({ page }: { page: PageReader }) {
	return (
		<TableFrame labelledBy={titleId("profile")} class="s-service-blocks-frame">
			<table class="s-service-blocks">
				<thead>
					<tr>
						<td />
						<th scope="col">{page.text("profile.heads.block")}</th>
						<th scope="col">{page.text("profile.heads.holds")}</th>
						<th scope="col">{page.text("profile.heads.writers")}</th>
						<th scope="col">{page.text("profile.heads.size")}</th>
					</tr>
				</thead>
				<tbody>
					{blocks.map((block) => (
						<BlockRow key={block.id} page={page} block={block} />
					))}
				</tbody>
			</table>
		</TableFrame>
	);
}

// BlockRow is one block of the file as a row of its table.
function BlockRow({ page, block }: { page: PageReader; block: Block }) {
	const words = `profile.blocks.${block.id}`;
	return (
		<tr>
			<td>
				<span
					class="s-service-swatch"
					style={{ "--tone": `var(${block.tone})` }}
				/>
			</td>
			<th scope="row">{page.text(`${words}.name`)}</th>
			<td>{page.text(`${words}.holds`)}</td>
			<td>
				<span class="s-service-writers">
					{block.writers.map((tool) => (
						<code key={tool}>{tool}</code>
					))}
					{block.everyWrite && <span>{page.text("profile.every-write")}</span>}
				</span>
			</td>
			<td>
				<Sizes
					page={page}
					words="profile.kb"
					typical={block.typical}
					cap={block.cap}
				/>
			</td>
		</tr>
	);
}
