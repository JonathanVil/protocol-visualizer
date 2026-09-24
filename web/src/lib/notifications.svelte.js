/** @typedef {{ id: number, text: string }} Notification */

let nextId = 0;

/** Transient error messages shown to the user, e.g. rejected commands. */
export const notifications = $state(/** @type {Notification[]} */ ([]));

/**
 * Shows an error for a few seconds.
 * @param {unknown} err
 */
export function notifyError(err) {
	const id = nextId++;
	const text = err instanceof Error ? err.message : String(err);
	notifications.push({ id, text });
	setTimeout(() => {
		const i = notifications.findIndex((n) => n.id === id);
		if (i !== -1) notifications.splice(i, 1);
	}, 5000);
}
