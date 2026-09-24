# Protocol — Bidirectional WebSocket

All frames are JSON objects sent over a single WebSocket connection at `/ws`.

**Naming convention**: commands are imperative (`message.drop`), events are past tense
(`message.dropped`). Commands originate in the frontend; events originate in the simulator.

---

## Frame shapes

### Frontend → Backend: command

```json
{ "Id": "c-1042", "type": "message.drop", "payload": { "messageId": "m-87" } }
```

| Field | Type | Notes |
|---|---|---|
| `Id` | string | Client-generated, unique per connection. Echoed in the corresponding ack or error. |
| `type` | string | Command name. |
| `payload` | object | Command-specific. May be omitted for payload-less commands. |

---

### Backend → Frontend: three kinds

#### Ack — command succeeded

```json
{ "type": "ack", "replyTo": "c-1042", "result": null }
```

`result` carries a return value where meaningful (e.g. `actor.invoke`), otherwise `null`.

#### Error — command rejected or failed

```json
{ "type": "error", "replyTo": "c-1042", "code": "MESSAGE_ALREADY_DELIVERED", "message": "m-87 was delivered at tick 312" }
```

`code` is a stable SCREAMING_SNAKE machine-readable token. `message` is human-readable detail.

#### Event — simulation fact, streamed continuously

```json
{ "seq": 5817, "tick": 312, "type": "message.delivered", "payload": { "messageId": "m-87", "from": 1, "to": 2, "payload": "ping" } }
```

| Field | Type | Notes |
|---|---|---|
| `seq` | int64 | Monotonically increasing, gap-free, per connection. Gaps indicate a backend bug. |
| `tick` | int | Simulation clock when the event occurred. |
| `type` | string | Event type (see table below). |
| `payload` | object | Event-specific. |

#### Snapshot — sent once on connect, before any events

```json
{ "type": "snapshot", "seq": 0, "tick": 5, "payload": { "actors": [...], "inTransit": [...], "settings": {...} } }
```

The snapshot is fetched at a tick boundary so it is internally consistent.
The frontend rebuilds its world from this frame and applies only events with `seq > snapshot.seq`.

---

## Command reference

All commands produce an `ack` or `error` frame. Payload may be omitted (or `null`) for
commands that don't require one.

### Simulation control

| Command | Payload fields | Result | Error codes |
|---|---|---|---|
| `start` | — | null | — |
| `stop` | — | null | — |
| `spawn` | `name: string` | actor Id (int) | `INVALID_ARGUMENT` (unknown type) |
| `requestTypes` | — | `[string, ...]` registered type names | — |

### Message manipulation

| Command | Payload fields | Result | Error codes |
|---|---|---|---|
| `message.drop` | `messageId: string` | null | `MESSAGE_NOT_FOUND`, `MESSAGE_ALREADY_DELIVERED` |
| `message.delay` | `messageId: string`, `ticks: int` | null | `MESSAGE_NOT_FOUND`, `MESSAGE_ALREADY_DELIVERED`, `INVALID_ARGUMENT` |
| `message.deliverNow` | `messageId: string` | null | `MESSAGE_NOT_FOUND`, `MESSAGE_ALREADY_DELIVERED` |

### Actor manipulation

| Command | Payload fields | Result | Error codes |
|---|---|---|---|
| `actor.setField` | `actorId: int`, `field: string`, `value: any` | null | `ACTOR_NOT_FOUND`, `FIELD_NOT_FOUND`, `INVALID_ARGUMENT` |
| `actor.invoke` | `actorId: int`, `method: string`, `args: any` | method return value | `ACTOR_NOT_FOUND`, `METHOD_NOT_FOUND`, `INVALID_ARGUMENT` |

### Sim settings

| Command | Payload fields | Result | Error codes |
|---|---|---|---|
| `sim.setSpeed` | `multiplier: float64` | null | `INVALID_ARGUMENT` |
| `sim.setTransitTime` | `ticks: int` | null | `INVALID_ARGUMENT` |

Unrecognized `type` → `UNKNOWN_COMMAND` error.
Malformed JSON or unparseable payload → `BAD_PAYLOAD` error.

---

## Event reference

| Event type | Payload fields |
|---|---|
| `message.delivered` | `messageId`, `from`, `to`, `payload` |
| `message.dropped` | `messageId` |
| `message.delayed` | `messageId`, `newDeliverTick` |
| `actor.spawned` | `actorId`, `typeName` |
| `actor.fieldChanged` | `actorId`, `field`, `value` |
| `sim.settingsChanged` | `speedMultiplier`, `transitTicks` |

Every successful mutation emits a corresponding event. The ack confirms the command was
accepted; the event is what the frontend renders from.

---

## Architectural invariants

1. **All V2 mutations are serialized through the simulation loop.** The WebSocket read
   goroutine enqueues a closure; the sim goroutine executes it between ticks.
2. **The sim package is wire-agnostic.** Simulation methods take plain Go values and return
   plain Go errors. They do not import the transport or protocol packages.
3. **Events are emitted only from the sim goroutine**, so `seq` assignment is race-free.
4. **One writer goroutine** owns the WebSocket write side; acks, errors, events, and
   snapshots all funnel through it.
5. **Snapshot before stream**: the snapshot is fetched at a tick boundary before the event
   stream begins, ensuring a consistent starting state.
