<script lang="ts">
    import { onMount } from 'svelte';
    import {requested} from "$app/server";
    interface Message {
        From: number;
        To: number;
        Payload: unknown;
    }

    interface WebSocketMessage {
        type: string;
        payload: any;
    }

    interface Event {
        Tick: number;
        Message: Message;
    }

    let events = $state<Event[]>([]);
    let ws: WebSocket;
    let connected = $state(false);
    let error = $state<string | null>(null);
    function connect() {
        error = null;
        ws = new WebSocket('ws://localhost:8067/ws');
        ws.onopen = () => {
            connected = true;
            requestInitialState();
        };
        ws.onclose = () => { connected = false; };
        ws.onerror = () => { error = 'WebSocket error. Is the backend running?'; connected = false; };
        ws.onmessage = (e: MessageEvent) => {
            try {
                let msg: WebSocketMessage = JSON.parse(e.data);
                switch (msg.type) {
                    case 'event':
                        events = [...events, msg.payload as Event];
                        break;
                    case 'actorTypes':
                        actorTypes = msg.payload.actors
                        break;
                }

            } catch {
                error = 'Failed to parse message';
            }
        };
    }
    onMount(() => {
        connect();
        return () => ws.close();
    });
    function requestInitialState() {
        requestActorTypes();
    }
    function requestActorTypes() { ws.send(JSON.stringify({ type: 'requestTypes' })); }
    function start() { ws.send(JSON.stringify({ type: 'start' })); }
    function stop() { ws.send(JSON.stringify({ type: 'stop' })); }
    function spawn() { ws.send(JSON.stringify({ type: 'spawn', payload: { 'name': selectedActorType } })); }

    let actorTypes = $state<string[]>([]);
    let selectedActorType = $state<string | null>(null);
</script>

<main class="container">
    <h1>Protocol Simulator</h1>

    {#if error}
        <p style="color: var(--pico-del-color)">{error}</p>
    {/if}

    <div>
        <button onclick={connect}>Reconnect</button>
        {#if connected}
            <button onclick={start}>Start</button>
            <button onclick={stop} class="secondary">Stop</button>

            <div role="group">
                <select name="actorTypes" bind:value={selectedActorType}>
                    {#each actorTypes as type}
                        <option>{type}</option>
                    {/each}
                </select>
                <button onclick={() => spawn()} type="button">Spawn</button>
            </div>
        {/if}
    </div>

    <table>
        <thead>
        <tr>
            <th>Tick</th>
            <th>From</th>
            <th>To</th>
            <th>Payload</th>
        </tr>
        </thead>
        <tbody>
        {#each events as event}
            <tr>
                <td>{event.Tick}</td>
                <td>Node {event.Message.From}</td>
                <td>Node {event.Message.To}</td>
                <td>{String(event.Message.Payload)}</td>
            </tr>
        {/each}
        </tbody>
    </table>
</main>