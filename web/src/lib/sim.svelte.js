/**
 * Client for the Go simulator's WebSocket protocol (see PROTOCOL.md).
 *
 * State is rebuilt from `snapshot` frames (sent on connect, every tick, and after
 * commands that run actor code) and kept current in between by applying events.
 * Event application is idempotent, so events that overlap a snapshot are harmless.
 */

/** @typedef {{ name: string, args: string[] }} MethodInfo */

/**
 * @typedef {{
 *   id: number,
 *   typeName: string,
 *   fields: Record<string, unknown>,
 *   methods: MethodInfo[]
 * }} Actor
 */

/**
 * @typedef {{
 *   id: string,
 *   from: number,
 *   to: number,
 *   payload: unknown,
 *   sentTick: number,
 *   deliverAtTick: number
 * }} InTransitMsg
 */

/** @typedef {{ seq: number, tick: number, type: string, payload: any }} SimEvent */

/** @typedef {{ tickDurationMs: number, transitTicks: number }} Settings */

/**
 * @typedef {{
 *   tick: number,
 *   running: boolean,
 *   settings: Settings,
 *   actors: Actor[] | null,
 *   inTransit: InTransitMsg[] | null
 * }} Snapshot
 */

export const DEFAULT_URL = 'ws://localhost:8067/ws';

/** Maximum number of events kept in the log. */
const MAX_LOG_EVENTS = 2000;

class Sim {
	connected = $state(false);
	/** @type {string | null} */
	error = $state(null);
	tick = $state(0);
	running = $state(false);
	/** @type {Actor[]} */
	actors = $state([]);
	/** @type {InTransitMsg[]} */
	inTransit = $state([]);
	/** @type {string[]} */
	actorTypes = $state([]);
	/** @type {SimEvent[]} */
	eventLog = $state([]);
	/** @type {Settings} */
	settings = $state({ tickDurationMs: 1000, transitTicks: 1 });

	/** @type {WebSocket | null} */
	#ws = null;
	#cmdId = 0;
	/** @type {Map<string, { resolve: (r: any) => void, reject: (e: Error) => void }>} */
	#pending = new Map();

	connect(url = DEFAULT_URL) {
		this.error = null;
		this.#ws?.close();
		const ws = new WebSocket(url);
		this.#ws = ws;

		ws.onopen = () => {
			if (this.#ws !== ws) return;
			this.connected = true;
			// The snapshot that follows is the whole world; events from an earlier
			// connection (possibly to a since-restarted program) no longer apply.
			this.eventLog = [];
			this.send('requestTypes')
				.then((types) => {
					this.actorTypes = [...types].sort();
				})
				.catch(() => {});
		};
		ws.onclose = () => {
			if (this.#ws !== ws) return;
			this.connected = false;
			this.#rejectPending(new Error('connection closed'));
		};
		ws.onerror = () => {
			if (this.#ws !== ws) return;
			this.error = `Cannot reach the simulator at ${url}. Is your Go program running?`;
		};
		ws.onmessage = (e) => {
			try {
				this.#dispatch(JSON.parse(e.data));
			} catch (err) {
				console.error('Failed to handle server frame', err, e.data);
			}
		};
	}

	disconnect() {
		const ws = this.#ws;
		this.#ws = null;
		ws?.close();
		this.connected = false;
		this.#rejectPending(new Error('disconnected'));
	}

	// --- Commands ---

	/**
	 * Sends a command and resolves with the ack result, or rejects with the error frame.
	 * @param {string} type
	 * @param {unknown} [payload]
	 * @returns {Promise<any>}
	 */
	send(type, payload) {
		const id = `c-${++this.#cmdId}`;
		return new Promise((resolve, reject) => {
			if (!this.#ws || this.#ws.readyState !== WebSocket.OPEN) {
				reject(new Error('Not connected to the simulator'));
				return;
			}
			this.#pending.set(id, { resolve, reject });
			this.#ws.send(JSON.stringify({ id, type, ...(payload !== undefined && { payload }) }));
		});
	}

	start() {
		return this.send('start').then(() => {
			this.running = true;
		});
	}

	stop() {
		return this.send('stop').then(() => {
			this.running = false;
		});
	}

	/** @param {string} name */
	spawn(name) {
		return this.send('spawn', { name });
	}

	/** @param {number} from @param {number} to @param {unknown} payload */
	sendMessage(from, to, payload) {
		return this.send('message.send', { from, to, payload });
	}

	/** @param {string} messageId */
	dropMessage(messageId) {
		return this.send('message.drop', { messageId });
	}

	/** @param {string} messageId @param {number} ticks */
	delayMessage(messageId, ticks) {
		return this.send('message.delay', { messageId, ticks });
	}

	/** @param {string} messageId */
	deliverNow(messageId) {
		return this.send('message.deliverNow', { messageId });
	}

	/** @param {number} actorId @param {string} field @param {unknown} value */
	setField(actorId, field, value) {
		return this.send('actor.setField', { actorId, field, value });
	}

	/** @param {number} actorId @param {string} method @param {unknown[]} args */
	invoke(actorId, method, args) {
		return this.send('actor.invoke', { actorId, method, args });
	}

	/** @param {number} tickDurationMs */
	setTickDuration(tickDurationMs) {
		return this.send('sim.setSpeed', { tickDurationMs });
	}

	/** @param {number} ticks */
	setTransitTicks(ticks) {
		return this.send('sim.setTransitTime', { ticks });
	}

	// --- Inbound frames ---

	/** @param {any} frame */
	#dispatch(frame) {
		switch (frame.type) {
			case 'snapshot':
				this.#applySnapshot(frame.payload);
				break;
			case 'ack':
				this.#settle(frame.replyTo)?.resolve(frame.result);
				break;
			case 'error':
				this.#settle(frame.replyTo)?.reject(new Error(frame.message ?? frame.code));
				break;
			default: {
				/** @type {SimEvent} */
				const event = { seq: frame.seq, tick: frame.tick, type: frame.type, payload: frame.payload };
				this.#applyEvent(event);
				this.eventLog.push(event);
				if (this.eventLog.length > MAX_LOG_EVENTS) {
					this.eventLog.splice(0, this.eventLog.length - MAX_LOG_EVENTS);
				}
				if (event.tick > this.tick) this.tick = event.tick;
			}
		}
	}

	/** @param {Snapshot} p */
	#applySnapshot(p) {
		this.tick = p.tick;
		this.running = p.running;
		this.settings = p.settings;
		this.actors = p.actors ?? [];
		this.inTransit = p.inTransit ?? [];
	}

	/** @param {SimEvent} event */
	#applyEvent({ type, payload: p }) {
		switch (type) {
			case 'actor.spawned':
				if (!this.actors.some((a) => a.id === p.actorId)) {
					this.actors.push({ id: p.actorId, typeName: p.typeName, fields: {}, methods: [] });
				}
				break;
			case 'actor.fieldChanged': {
				const actor = this.actors.find((a) => a.id === p.actorId);
				if (actor) actor.fields[p.field] = p.value;
				break;
			}
			case 'message.sent':
				if (!this.inTransit.some((m) => m.id === p.messageId)) {
					this.inTransit.push({
						id: p.messageId,
						from: p.from,
						to: p.to,
						payload: p.payload,
						sentTick: p.sentTick,
						deliverAtTick: p.deliverAtTick
					});
				}
				break;
			case 'message.delivered':
			case 'message.dropped':
				this.inTransit = this.inTransit.filter((m) => m.id !== p.messageId);
				break;
			case 'message.delayed': {
				const msg = this.inTransit.find((m) => m.id === p.messageId);
				if (msg) msg.deliverAtTick = p.newDeliverTick;
				break;
			}
			case 'sim.settingsChanged':
				this.settings = { tickDurationMs: p.tickDurationMs, transitTicks: p.transitTicks };
				break;
		}
	}

	/** @param {string} replyTo */
	#settle(replyTo) {
		const pending = this.#pending.get(replyTo);
		this.#pending.delete(replyTo);
		return pending;
	}

	/** @param {Error} err */
	#rejectPending(err) {
		for (const { reject } of this.#pending.values()) reject(err);
		this.#pending.clear();
	}
}

export const sim = new Sim();
