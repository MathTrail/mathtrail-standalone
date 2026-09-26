/**
 * StubCard shows a result's payload as it arrived, until the screens that draw
 * it properly exist. The payload goes into the page as text, so nothing inside
 * it can become markup.
 */
export function StubCard({ payload }: { payload: unknown }) {
	return <pre class="stub">{JSON.stringify(payload ?? null, null, 2)}</pre>;
}
