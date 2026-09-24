export interface Actor {
	id: number;
	typeName: string;
}

export interface InTransitMsg {
	id: number;
	from: number;
	to: number;
	payload: unknown;
	sentTick: number;
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
	actors = $state([]);
	inTransit = $state([]);
	actorTypes = $state([]);
	eventLog = $state([]);

	#ws = null;
	#cmdId = 0;
	#pending = new Map();

	connect(url = 'ws://localhost:8067/ws') {
		this.error = null;
		this.#ws?.close();
		const ws = new WebSocket(url);
		this.#ws = ws;

		ws.onopen = () => {
			this.connected = true;
			this.send('requestTypes')
				.then((r) => { this.actorTypes = r; })
				.catch(() => {});
		};
		ws.onclose = () => { this.connected = false; };
		ws.onerror = () => { this.error = 'WebSocket error — is the backend running?'; };
		ws.onmessage = (e) => {
			try { this.#dispatch(JSON.parse(e.data )); }
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
	spawn(name) {
		this.#fire('spawn', { name });
	}

	// Correlated send — resolves with ack result or rejects on error frame
	send(type, payload) {
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

	#fire(type, payload) {
		if (!this.#ws || this.#ws.readyState !== WebSocket.OPEN) {
			this.error = 'Not connected';
			return;
		}
		const id = `c-${++this.#cmdId}`;
		this.#ws.send(JSON.stringify({ id, type, ...(payload !== undefined && { payload }) }));
	}

	#dispatch(frame) {
		switch (frame.type) {
			case 'snapshot': {
				const p = frame.payload;
				this.tick = p.Tick;
				this.running = p.Running;
				this.actors = (p.Actors ?? []).map((a) => ({ id: a.id, typeName: a.typeName }));
				this.inTransit = (p.Messages ?? []).map((m) => ({
					id: m.Id, from: m.From, to: m.To,
					payload: m.Payload, sentTick: m.SentTick, deliverAtTick: m.DeliverAtTick,
				}));
				break;
			}
			case 'ack': {
				const pending = this.#pending.get(frame.replyTo );
				if (pending) {
					this.#pending.delete(frame.replyTo );
					pending.resolve(frame.result);
				}
				break;
			}
			case 'error': {
				const pending = this.#pending.get(frame.replyTo );
				if (pending) {
					this.#pending.delete(frame.replyTo );
					pending.reject(new Error(`${frame.code}: ${frame.message}`));
				}
				console.warn('sim error:', frame.code, frame.message);
				break;
			}
			default: {
				// Simulation event
				const seq = frame.seq ;
				const tick = frame.tick ;
				this.eventLog = [
					...this.eventLog,
					{ seq, tick, type: frame.type , payload: frame.payload },
				];
				if (tick > this.tick) this.tick = tick;
				/*
				if (frame.payload) {
					this.#applyEvent(frame.type , frame.payload as Record<string, unknown>);
				}
				*/
			}
		}
	}

	#applyEvent(type, p) {
		switch (type) {
			case 'actor.spawned':
				this.actors = [...this.actors, { id: p.actorId , typeName: p.typeName  }];
				break;
			case 'message.sent':
				// Backend will emit this once the plan is implemented
				this.inTransit = [...this.inTransit, {
					id: p.messageId,
					from: p.from,
					to: p.to,
					payload: p.payload,
					sentTick: p.sentTick,
					deliverAtTick: p.deliverTick,
				}];
				break;
			case 'message.delivered':
			case 'message.dropped':
				this.inTransit = this.inTransit.filter((m) => m.id !== (p.messageId ));
				break;
			case 'message.delayed':
				this.inTransit = this.inTransit.map((m) =>
					m.id === (p.messageId )
						? { ...m, deliverAtTick: p.newDeliverTick }
						: m
				);
				break;
			case 'clock.tick':
				this.tick = (p.tick ) ?? this.tick;
				break;
		}
	}
}

export const sim = new Sim();
