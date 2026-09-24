export interface Notification {
	id: number;
	text: string;
}

let nextId = 0;

/** Transient error messages shown to the user, e.g. rejected commands. */
export const notifications = $state<Notification[]>([]);

/** Shows an error for a few seconds. */
export function notifyError(err: unknown) {
	const id = nextId++;
	const text = err instanceof Error ? err.message : String(err);
	notifications.push({ id, text });
	setTimeout(() => {
		const i = notifications.findIndex((n) => n.id === id);
		if (i !== -1) notifications.splice(i, 1);
	}, 5000);
}
