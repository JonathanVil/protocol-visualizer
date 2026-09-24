# Protocol — Bidirectional WebSocket

All frames are JSON objects sent over a single WebSocket connection at `/ws`.

**Naming convention**: commands are imperative (`message.drop`), events are past tense
(`message.dropped`). Commands originate in the frontend; events originate in the simulator.

---

## Frame shapes

### Frontend → Backend: command

```json
{ "id": "c-1042", "type": "message.drop", "payload": { "messageId": "m-87" } }
```

| Field | Type | Notes |
|---|---|---|
| `id` | string | Client-generated, unique per connection. Echoed in the corresponding ack or error. |
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

#### Snapshot — the whole world

```json
{ "type": "snapshot", "seq": 0, "tick": 5, "payload": {
    "tick": 5,
    "running": true,
    "settings": { "tickDurationMs": 1000, "transitTicks": 5 },
    "actors": [
      { "id": 0, "typeName": "ping",
        "fields": { "PingsReceived": 3 },
        "methods": [ { "name": "SendPing", "args": ["int"] } ] }
    ],
    "inTransit": [
      { "id": "m-87", "from": 0, "to": 1, "payload": "ping", "sentTick": 3, "deliverAtTick": 8 }
    ]
} }
```

A snapshot is sent once on connect, before any events, and fetched at a tick boundary so it is
internally consistent. The simulator also emits snapshots as events (with a `seq`) after every
tick, on start/stop, after a spawn, and after `message.deliverNow` / `actor.invoke`, since actor
code may change state outside a tick.

The frontend replaces its world with every snapshot and applies events in between. Events emitted
just before the connect-time snapshot may arrive after it, so applying an event must be idempotent.

`fields` holds the actor's exported struct fields, keyed by their `json` tag name if present
(fields tagged `json:"-"` are omitted). `methods` lists exported methods other than those of the
`Actor` interface, with their parameter types.

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
| `message.send` | `from: int`, `to: int`, `payload: any` | null | `ACTOR_NOT_FOUND` |
| `message.drop` | `messageId: string` | null | `MESSAGE_NOT_FOUND`, `MESSAGE_ALREADY_DELIVERED` |
| `message.delay` | `messageId: string`, `ticks: int` | null | `MESSAGE_NOT_FOUND`, `MESSAGE_ALREADY_DELIVERED`, `INVALID_ARGUMENT` |
| `message.deliverNow` | `messageId: string` | null | `MESSAGE_NOT_FOUND`, `MESSAGE_ALREADY_DELIVERED` |

### Actor manipulation

| Command | Payload fields | Result | Error codes |
|---|---|---|---|
| `actor.setField` | `actorId: int`, `field: string`, `value: any` | null | `ACTOR_NOT_FOUND`, `FIELD_NOT_FOUND`, `INVALID_ARGUMENT` |
| `actor.invoke` | `actorId: int`, `method: string`, `args: [any, ...]` | method return value | `ACTOR_NOT_FOUND`, `METHOD_NOT_FOUND`, `INVALID_ARGUMENT` |

`field` is a key of the snapshot's `fields`. `value` and `args` are converted to the Go field or
parameter type; string, bool and numeric types are supported (a number must be integral for an
integer type). For `actor.invoke`, a trailing `error` result is returned as an error frame when it
is non-nil; otherwise the result is the single return value, an array of them, or null.

### Sim settings

| Command | Payload fields | Result | Error codes |
|---|---|---|---|
| `sim.setSpeed` | `tickDurationMs: int` | null | `INVALID_ARGUMENT` |
| `sim.setTransitTime` | `ticks: int` | null | `INVALID_ARGUMENT` |

Unrecognized `type` → `UNKNOWN_COMMAND` error.
Malformed JSON or unparseable payload → `BAD_PAYLOAD` error.

---

## Event reference

| Event type | Payload fields |
|---|---|
| `message.sent` | `messageId`, `from`, `to`, `payload`, `sentTick`, `deliverAtTick` |
| `message.delivered` | `messageId`, `from`, `to`, `payload` |
| `message.dropped` | `messageId` |
| `message.delayed` | `messageId`, `newDeliverTick` |
| `actor.spawned` | `actorId`, `typeName` |
| `actor.fieldChanged` | `actorId`, `field`, `value` |
| `sim.settingsChanged` | `tickDurationMs`, `transitTicks` |
| `snapshot` | see [Snapshot](#snapshot--the-whole-world) |

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
5. **Snapshot before stream**: each connection subscribes to events, then fetches the snapshot at
   a tick boundary and sends it before any events. Every connection receives every event.
