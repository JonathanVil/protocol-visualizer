/**
 * Client for the Go simulator's WebSocket protocol (see PROTOCOL.md).
 *
 * State is rebuilt from `snapshot` frames (sent on connect, every tick, and after
 * commands that run actor code) and kept current in between by applying events.
 * Event application is idempotent, so events that overlap a snapshot are harmless.
 */

export interface MethodInfo {
	name: string;
	/** Go parameter types, e.g. "int", "string". */
	args: string[];
}

export interface Actor {
	id: number;
	typeName: string;
	fields: Record<string, unknown>;
	methods: MethodInfo[];
}

export interface InTransitMsg {
	id: string;
	from: number;
	to: number;
	payload: unknown;
	sentTick: number;
	deliverAtTick: number;
}

export interface Settings {
	tickDurationMs: number;
	transitTicks: number;
}

export interface Snapshot {
	tick: number;
	running: boolean;
	settings: Settings;
	actors: Actor[] | null;
	inTransit: InTransitMsg[] | null;
}

/** A simulation event, discriminated by `type`. */
export type SimEvent = { seq: number; tick: number } & (
	| {
			type: 'message.sent';
			payload: {
				messageId: string;
				from: number;
				to: number;
				payload: unknown;
				sentTick: number;
				deliverAtTick: number;
			};
	  }
	| {
			type: 'message.delivered';
			payload: { messageId: string; from: number; to: number; payload: unknown };
	  }
	| { type: 'message.dropped'; payload: { messageId: string } }
	| { type: 'message.delayed'; payload: { messageId: string; newDeliverTick: number } }
	| { type: 'actor.spawned'; payload: { actorId: number; typeName: string } }
	| { type: 'actor.fieldChanged'; payload: { actorId: number; field: string; value: unknown } }
	| { type: 'sim.settingsChanged'; payload: Settings }
);

type ServerFrame =
	| { type: 'snapshot'; seq: number; tick: number; payload: Snapshot }
	| { type: 'ack'; replyTo: string; result: unknown }
	| { type: 'error'; replyTo: string; code: string; message: string }
	| SimEvent;

interface PendingCommand {
	resolve: (result: unknown) => void;
	reject: (err: Error) => void;
}

export const DEFAULT_URL = 'ws://localhost:8067/ws';

/** Maximum number of events kept in the log. */
const MAX_LOG_EVENTS = 2000;

class Sim {
	connected = $state(false);
	error = $state<string | null>(null);
	tick = $state(0);
	running = $state(false);
	actors = $state<Actor[]>([]);
	inTransit = $state<InTransitMsg[]>([]);
	actorTypes = $state<string[]>([]);
	eventLog = $state<SimEvent[]>([]);
	settings = $state<Settings>({ tickDurationMs: 1000, transitTicks: 1 });

	#ws: WebSocket | null = null;
	#cmdId = 0;
	#pending = new Map<string, PendingCommand>();

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
			this.send<string[]>('requestTypes')
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
		ws.onmessage = (e: MessageEvent<string>) => {
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

	/** Sends a command and resolves with the ack result, or rejects with the error frame. */
	send<T = unknown>(type: string, payload?: unknown): Promise<T> {
		const id = `c-${++this.#cmdId}`;
		return new Promise<T>((resolve, reject) => {
			if (!this.#ws || this.#ws.readyState !== WebSocket.OPEN) {
				reject(new Error('Not connected to the simulator'));
				return;
			}
			this.#pending.set(id, { resolve: resolve as (result: unknown) => void, reject });
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

	/** Resolves with the new actor's id. */
	spawn(name: string) {
		return this.send<number>('spawn', { name });
	}

	sendMessage(from: number, to: number, payload: unknown) {
		return this.send('message.send', { from, to, payload });
	}

	dropMessage(messageId: string) {
		return this.send('message.drop', { messageId });
	}

	delayMessage(messageId: string, ticks: number) {
		return this.send('message.delay', { messageId, ticks });
	}

	deliverNow(messageId: string) {
		return this.send('message.deliverNow', { messageId });
	}

	setField(actorId: number, field: string, value: unknown) {
		return this.send('actor.setField', { actorId, field, value });
	}

	/** Resolves with the method's return value(s). */
	invoke(actorId: number, method: string, args: unknown[]) {
		return this.send('actor.invoke', { actorId, method, args });
	}

	setTickDuration(tickDurationMs: number) {
		return this.send('sim.setSpeed', { tickDurationMs });
	}

	setTransitTicks(ticks: number) {
		return this.send('sim.setTransitTime', { ticks });
	}

	// --- Inbound frames ---

	#dispatch(frame: ServerFrame) {
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
			default:
				this.#applyEvent(frame);
				this.eventLog.push(frame);
				if (this.eventLog.length > MAX_LOG_EVENTS) {
					this.eventLog.splice(0, this.eventLog.length - MAX_LOG_EVENTS);
				}
				if (frame.tick > this.tick) this.tick = frame.tick;
		}
	}

	#applySnapshot(p: Snapshot) {
		this.tick = p.tick;
		this.running = p.running;
		this.settings = p.settings;
		this.actors = p.actors ?? [];
		this.inTransit = p.inTransit ?? [];
	}

	#applyEvent(event: SimEvent) {
		switch (event.type) {
			case 'actor.spawned': {
				const p = event.payload;
				if (!this.actors.some((a) => a.id === p.actorId)) {
					this.actors.push({ id: p.actorId, typeName: p.typeName, fields: {}, methods: [] });
				}
				break;
			}
			case 'actor.fieldChanged': {
				const p = event.payload;
				const actor = this.actors.find((a) => a.id === p.actorId);
				if (actor) actor.fields[p.field] = p.value;
				break;
			}
			case 'message.sent': {
				const p = event.payload;
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
			}
			case 'message.delivered':
			case 'message.dropped': {
				const { messageId } = event.payload;
				this.inTransit = this.inTransit.filter((m) => m.id !== messageId);
				break;
			}
			case 'message.delayed': {
				const p = event.payload;
				const msg = this.inTransit.find((m) => m.id === p.messageId);
				if (msg) msg.deliverAtTick = p.newDeliverTick;
				break;
			}
			case 'sim.settingsChanged':
				this.settings = { ...event.payload };
				break;
		}
	}

	#settle(replyTo: string) {
		const pending = this.#pending.get(replyTo);
		this.#pending.delete(replyTo);
		return pending;
	}

	#rejectPending(err: Error) {
		for (const { reject } of this.#pending.values()) reject(err);
		this.#pending.clear();
	}
}

export const sim = new Sim();
