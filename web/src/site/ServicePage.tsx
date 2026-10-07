import { Fragment } from "preact";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";
import { ServiceFlow } from "./ServiceFlow";
import { ServiceMap } from "./ServiceMap";
import { ServiceProfile } from "./ServiceProfile";
import { ServiceSection } from "./ServiceSection";
import {
	coreParts,
	hosts,
	type Lifetime,
	lifeScale,
	lifetimes,
	logShare,
	profileFile,
	purposes,
	reachOf,
	type Store,
	stores,
	tokenSample,
	wires,
} from "./service";

/**
 * ServicePage is the page of how the service works, a technical page rather
 * than one written for parents: the service in three tiers, the map of its
 * parts and what they talk to, a lesson call by call, where its state lives
 * when it keeps none, the profile's file, and what it seals and for how long.
 * It reads in full without a script: what its choices show is in the page,
 * and the rules of its head show it.
 */
export function ServicePage({ page }: PageProps) {
	return (
		<div class="s-service">
			<Hero page={page} />
			<ServiceMap page={page} />
			<ServiceFlow page={page} />
			<State page={page} />
			<ServiceProfile page={page} />
			<Seal page={page} />
		</div>
	);
}

// Hero is the first screen: the heading, what the service is, the chips of
// what it is built as, and beside them the service in three tiers.
function Hero({ page }: { page: PageReader }) {
	return (
		<section class="s-wrap s-hero s-service-hero">
			<div class="s-hero-copy">
				<h1>{page.text("hero.title")}</h1>
				<p class="s-lead">{page.text("hero.lead")}</p>
				<ul class="s-chips s-service-chips">
					{page.list("hero.chips").map((key) => (
						<li key={key} class="s-chip">
							{page.text(key)}
						</li>
					))}
				</ul>
			</div>
			<Tiers page={page} />
		</section>
	);
}

// Tiers are the service in three tiers, one over another, with what joins
// each two: the chat a family uses, the service with its core of pure
// computation, and the parent's Drive, where the one file lives.
function Tiers({ page }: { page: PageReader }) {
	return (
		<figure class="s-service-tiers" aria-label={page.plain("hero.figure")}>
			<div class="s-service-tier s-service-tier-client">
				<TierHead page={page} tier="client" />
				<ul class="s-service-hosts">
					{hosts.map((host) => (
						<li key={host}>{host}</li>
					))}
					<li class="s-service-hosts-more">
						{page.text("hero.client.others")}
					</li>
				</ul>
				<p class="s-service-tier-note">{page.text("hero.client.note")}</p>
			</div>
			<Wire page={page} words="hero.mcp" via={wires.chat} />
			<div class="s-service-tier s-service-tier-cloud">
				<TierHead page={page} tier="cloud" />
				<p class="s-service-tier-sub">{page.text("hero.cloud.sub")}</p>
				<div class="s-service-core">
					<p>{page.text("hero.cloud.core")}</p>
					<ul>
						{coreParts.map((part) => (
							<li key={part}>{page.text(`map.parts.${part}.name`)}</li>
						))}
					</ul>
				</div>
				<p class="s-service-tier-note">{page.text("hero.cloud.around")}</p>
			</div>
			<Wire page={page} words="hero.drive" via={wires.drive} />
			<div class="s-service-tier s-service-tier-storage">
				<TierHead page={page} tier="storage" />
				<p>
					<code class="s-service-file">
						{profileFile.folder}/{profileFile.name}
					</code>
				</p>
				<p class="s-service-tier-note">{page.text("hero.storage.note")}</p>
			</div>
		</figure>
	);
}

// TierHead is a tier's name, and beside it its number and what it is.
function TierHead({ page, tier }: { page: PageReader; tier: string }) {
	return (
		<p class="s-service-tier-head">
			<span class="s-service-tier-title">
				{page.text(`hero.${tier}.title`)}
			</span>
			<span class="s-service-tier-tag">{page.text(`hero.${tier}.tag`)}</span>
		</p>
	);
}

// Wire is what joins two tiers: an arrow both ways, and the protocol that
// carries their talk, by its name, with how it is carried.
function Wire({
	page,
	words,
	via,
}: {
	page: PageReader;
	words: string;
	via: string;
}) {
	return (
		<p class="s-service-wire">
			<span class="s-service-wire-arrow" aria-hidden="true">
				<span />
			</span>
			<span>
				{page.text(words, { via: <span class="s-service-via">{via}</span> })}
			</span>
		</p>
	);
}

// State is where the state lives that a server with state would keep: the
// service itself, which keeps nothing, then every place that keeps a piece of
// it, and what v1 has none of on purpose.
function State({ page }: { page: PageReader }) {
	return (
		<ServiceSection page={page} id="state">
			<div class="s-service-stores">
				<div class="s-service-store s-service-store-self">
					<h3>{page.text("state.self.label")}</h3>
					<p class="s-service-nothing">{page.text("state.self.value")}</p>
					<p class="s-service-store-note">{page.text("state.self.note")}</p>
				</div>
				{stores.map((store) => (
					<StoreCard key={store.id} page={page} store={store} />
				))}
			</div>
			<div class="s-service-absent">
				<h3>{page.text("state.absent.title")}</h3>
				<ul class="s-chips s-unneeded">
					{page.list("state.absent.items").map((key) => (
						<li key={key} class="s-chip">
							{page.text(key)}
						</li>
					))}
				</ul>
			</div>
		</ServiceSection>
	);
}

// StoreCard is one place the state lives: who keeps it, a line on how, and
// what it keeps for how long, with a note where it has one.
function StoreCard({ page, store }: { page: PageReader; store: Store }) {
	const words = `state.stores.${store.id}`;
	return (
		<div class={`s-service-store s-service-store-${store.tone}`}>
			<div>
				<h3>{page.text(`${words}.title`)}</h3>
				<p class="s-service-store-sub">{page.text(`${words}.sub`)}</p>
			</div>
			<div class="s-service-store-body">
				<dl class="s-service-store-rows">
					{store.rows.map((row) => {
						const at = `${words}.rows.${row.id}`;
						return (
							<div key={row.id}>
								<dt>
									{row.code === undefined ? (
										<>
											<span>{page.text(`${at}.name`)}</span>
											{page.has(`${at}.sub`) && (
												<span class="s-service-store-row-sub">
													{page.text(`${at}.sub`)}
												</span>
											)}
										</>
									) : (
										<code>{row.code}</code>
									)}
								</dt>
								<dd>{page.text(`${at}.life`)}</dd>
							</div>
						);
					})}
				</dl>
				{page.has(`${words}.note`) && (
					<p class="s-service-store-note">{page.text(`${words}.note`)}</p>
				)}
			</div>
		</div>
	);
}

// Seal is what the service seals and for how long: a token taken apart, and
// the lifetimes of the sign-in's pieces on a scale of their own.
function Seal({ page }: { page: PageReader }) {
	return (
		<ServiceSection page={page} id="seal">
			<div class="s-service-seal">
				<Token page={page} />
				<Lives page={page} />
			</div>
		</ServiceSection>
	);
}

// Token is a sealed access token taken apart: its version, its purpose, the
// id of its key and its ciphertext, each told under it.
function Token({ page }: { page: PageReader }) {
	const { version, purpose, key, sealed } = tokenSample;
	return (
		<div class="s-service-card s-service-token">
			<h3 class="s-service-label">{page.text("seal.token.title")}</h3>
			<p class="s-service-token-parts" dir="ltr">
				<code class="s-service-token-version">{version}</code>
				<span>.</span>
				<code class="s-service-token-purpose">{purpose}</code>
				<span>.</span>
				<code class="s-service-token-key">{key}</code>
				<span>.</span>
				<code class="s-service-token-sealed">{sealed}</code>
			</p>
			<dl class="s-service-token-legend">
				<div>
					<dt>{version}</dt>
					<dd>{page.text("seal.token.version")}</dd>
				</div>
				<div class="s-service-token-purpose">
					<dt>{page.text("seal.token.purpose")}</dt>
					<dd>
						{purposes.map((letter, at) => (
							<Fragment key={letter}>
								{at > 0 && " · "}
								{letter} {page.text(`seal.token.purposes.${letter}`)}
							</Fragment>
						))}
					</dd>
				</div>
				<div class="s-service-token-key">
					<dt>{page.text("seal.token.key")}</dt>
					<dd>{page.text("seal.token.key-note")}</dd>
				</div>
				<div>
					<dt>{page.text("seal.token.cipher")}</dt>
					<dd>{page.text("seal.token.cipher-note")}</dd>
				</div>
			</dl>
			<p class="s-service-small">{page.text("seal.token.note")}</p>
		</div>
	);
}

// Lives are the lifetimes on the scale of the logarithm of the time: its marks
// over them, then each piece by its name and how long it lives, beside a bar
// as long as that, and for a token that slides, as far as it can be renewed.
function Lives({ page }: { page: PageReader }) {
	const marks = lifeScale.ticks.map(logShare);
	const between = marks.slice(1, -1);
	return (
		<div class="s-service-card s-service-lives-card">
			<h3 class="s-service-label">{page.text("seal.life.title")}</h3>
			<ul class="s-service-lives" dir="ltr">
				<li class="s-service-lives-axis" aria-hidden="true">
					<span />
					<span class="s-service-marks">
						{marks.map((share, at) => (
							<span
								key={share}
								style={
									at === 0 || at === marks.length - 1
										? undefined
										: { insetInlineStart: `${share}%` }
								}
							>
								{page.text(`seal.life.ticks.${at + 1}`)}
							</span>
						))}
					</span>
				</li>
				{lifetimes.map((life) => (
					<Life key={life.id} page={page} life={life} marks={between} />
				))}
			</ul>
			<p class="s-service-small">{page.text("seal.life.note")}</p>
		</div>
	);
}

// Life is one lifetime: its piece's name and how long it lives, and its bar
// over a hairline at each of the scale's marks between its ends.
function Life({
	page,
	life,
	marks,
}: {
	page: PageReader;
	life: Lifetime;
	marks: readonly number[];
}) {
	const words = `seal.life.items.${life.id}`;
	const { lives, more } = reachOf(life);
	const bar = [
		"s-service-life-bar",
		life.key === true ? "s-service-life-key" : "",
		more > 0 ? "s-service-life-slides" : "",
	]
		.filter((name) => name !== "")
		.join(" ");
	return (
		<li class="s-service-life">
			<span class="s-service-life-name">
				{page.text(`${words}.name`)}
				<br />
				<span class="s-service-life-sub">{page.text(`${words}.life`)}</span>
			</span>
			<span class="s-service-life-track" aria-hidden="true">
				{marks.map((share) => (
					<span
						key={share}
						class="s-service-life-mark"
						style={{ insetInlineStart: `${share}%` }}
					/>
				))}
				<span class={bar} style={{ "--lives": `${lives}%` }} />
				{more > 0 && (
					<span class="s-service-life-more" style={{ "--more": `${more}%` }} />
				)}
			</span>
		</li>
	);
}
