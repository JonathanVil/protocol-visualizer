/** Formats an arbitrary value for display. */
export function formatValue(v: unknown): string {
	if (v === null) return 'null';
	if (v === undefined) return 'undefined';
	if (typeof v === 'string') return JSON.stringify(v);
	try {
		return typeof v === 'object' ? JSON.stringify(v) : String(v);
	} catch {
		return String(v);
	}
}

/** Short label for a message payload, e.g. on a message node in the graph. */
export function payloadLabel(payload: unknown, max = 12): string {
	const text = typeof payload === 'string' ? payload : formatValue(payload);
	return text.length > max ? text.slice(0, max - 1) + '…' : text;
}

/**
 * Parses user input as JSON, falling back to the raw string, so that `5` is a
 * number, `true` a boolean, `{"a":1}` an object, and `ping` a string.
 */
export function parseLoose(text: string): unknown {
	try {
		return JSON.parse(text);
	} catch {
		return text;
	}
}
