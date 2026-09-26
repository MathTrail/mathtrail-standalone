import { useEffect, useState } from "preact/hooks";
import type { Bridge, ToolResult } from "./bridge";
import { StubCard } from "./StubCard";

/**
 * WidgetApp is the card a tool's result is drawn as. Until the first result
 * arrives it draws nothing.
 */
export function WidgetApp({ bridge }: { bridge: Bridge }) {
	const result = useLatestResult(bridge);
	if (result === undefined) {
		return null;
	}
	return <StubCard payload={result.structuredContent} />;
}

// useLatestResult is the bridge's latest result, kept current as new ones come.
function useLatestResult(bridge: Bridge): ToolResult | undefined {
	const [result, setResult] = useState(bridge.result());
	useEffect(() => {
		// A result may have arrived between the first render and this
		// subscription; it is read once more rather than missed.
		setResult(bridge.result());
		return bridge.subscribe(() => setResult(bridge.result()));
	}, [bridge]);
	return result;
}
