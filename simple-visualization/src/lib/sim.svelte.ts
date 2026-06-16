export interface Actor {
	id: number;
	typeName: string;
}

export interface InTransitMsg {
	id: string;
	from: number;
	to: number;
	payload: unknown;
	deliverAtTick: number;
}

export interface SimEvent {
	seq: number;
	tick: number;
	type: string;
	payload: unknown;
}

class Sim {
	connected = $state(false);
	error = $state<string | null>(null);
	tick = $state(0);
	running = $state(false);
	actors = $state<Actor[]>([]);
	inTransit = $state<InTransitMsg[]>([]);
	actorTypes = $state<string[]>([]);
	eventLog = $state<SimEvent[]>([]);

	#ws: WebSocket | null = null;
	#cmdId = 0;
	#pending = new Map<string, { resolve: (r: unknown) => void; reject: (e: Error) => void }>();

	connect(url = 'ws://localhost:8067/ws') {
		this.error = null;
		this.#ws?.close();
		const ws = new WebSocket(url);
		this.#ws = ws;

		ws.onopen = () => {
			this.connected = true;
			this.send('requestTypes')
				.then((r) => { this.actorTypes = r as string[]; })
				.catch(() => {});
		};
		ws.onclose = () => { this.connected = false; };
		ws.onerror = () => { this.error = 'WebSocket error — is the backend running?'; };
		ws.onmessage = (e: MessageEvent) => {
			try { this.#dispatch(JSON.parse(e.data as string)); }
			catch { this.error = 'Failed to parse server message'; }
		};
	}

	disconnect() {
		this.#ws?.close();
		this.#ws = null;
	}

	start() {
		this.send('start')
			.then(() => { this.running = true; })

	}
	stop()  {
		this.send('stop')
			.then(() => { this.running = false; })
	}
	spawn(name: string) {
		this.#fire('spawn', { name });
	}

	// Correlated send — resolves with ack result or rejects on error frame
	send(type: string, payload?: unknown): Promise<unknown> {
		const id = `c-${++this.#cmdId}`;
		return new Promise((resolve, reject) => {
			if (!this.#ws || this.#ws.readyState !== WebSocket.OPEN) {
				reject(new Error('not connected'));
				return;
			}
			this.#pending.set(id, { resolve, reject });
			this.#ws.send(JSON.stringify({ id, type, ...(payload !== undefined && { payload }) }));
		});
	}

	#fire(type: string, payload?: unknown) {
		if (!this.#ws || this.#ws.readyState !== WebSocket.OPEN) {
			this.error = 'Not connected';
			return;
		}
		const id = `c-${++this.#cmdId}`;
		this.#ws.send(JSON.stringify({ id, type, ...(payload !== undefined && { payload }) }));
	}

	#dispatch(frame: Record<string, unknown>) {
		switch (frame.type) {
			case 'snapshot': {
				const p = frame.payload as {
					Tick: number;
					Running: boolean;
					Messages: Array<{ Id: string; From: number; To: number; Payload: unknown; DeliverAtTick: number }>;
				};
				this.tick = p.Tick;
				this.running = p.Running;
				this.inTransit = (p.Messages ?? []).map((m) => ({
					id: m.Id, from: m.From, to: m.To,
					payload: m.Payload, deliverAtTick: m.DeliverAtTick,
				}));
				break;
			}
			case 'ack': {
				const pending = this.#pending.get(frame.replyTo as string);
				if (pending) {
					this.#pending.delete(frame.replyTo as string);
					pending.resolve(frame.result);
				}
				break;
			}
			case 'error': {
				const pending = this.#pending.get(frame.replyTo as string);
				if (pending) {
					this.#pending.delete(frame.replyTo as string);
					pending.reject(new Error(`${frame.code}: ${frame.message}`));
				}
				console.warn('sim error:', frame.code, frame.message);
				break;
			}
			default: {
				// Simulation event
				const seq = frame.seq as number;
				const tick = frame.tick as number;
				this.eventLog = [
					...this.eventLog,
					{ seq, tick, type: frame.type as string, payload: frame.payload },
				];
				if (tick > this.tick) this.tick = tick;
				if (frame.payload) {
					this.#applyEvent(frame.type as string, frame.payload as Record<string, unknown>);
				}
			}
		}
	}

	#applyEvent(type: string, p: Record<string, unknown>) {
		switch (type) {
			case 'actor.spawned':
				this.actors = [...this.actors, { id: p.actorId as number, typeName: p.typeName as string }];
				break;
			case 'message.sent':
				// Backend will emit this once the plan is implemented
				this.inTransit = [...this.inTransit, {
					id: p.messageId as string,
					from: p.from as number,
					to: p.to as number,
					payload: p.payload,
					deliverAtTick: p.deliverTick as number,
				}];
				break;
			case 'message.delivered':
			case 'message.dropped':
				this.inTransit = this.inTransit.filter((m) => m.id !== (p.messageId as string));
				break;
			case 'message.delayed':
				this.inTransit = this.inTransit.map((m) =>
					m.id === (p.messageId as string)
						? { ...m, deliverAtTick: p.newDeliverTick as number }
						: m
				);
				break;
			case 'clock.tick':
				this.tick = (p.tick as number) ?? this.tick;
				break;
		}
	}
}

export const sim = new Sim();
