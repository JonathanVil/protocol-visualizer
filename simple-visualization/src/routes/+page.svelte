<script lang="ts">
    import { onMount } from 'svelte';
    interface Message {
        From: number;
        To: number;
        Payload: unknown;
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
        ws.onopen = () => { connected = true; };
        ws.onclose = () => { connected = false; };
        ws.onerror = () => { error = 'WebSocket error. Is the backend running?'; connected = false; };
        ws.onmessage = (e: MessageEvent) => {
            try {
                events = [...events, JSON.parse(e.data) as Event];
            } catch {
                error = 'Failed to parse message';
            }
        };
    }
    onMount(() => {
        connect();
        return () => ws.close();
    });
    function start() { ws.send(JSON.stringify({ type: 'start' })); }
    function stop() { ws.send(JSON.stringify({ type: 'stop' })); }
    function spawn(name: string) { ws.send(JSON.stringify({ type: 'spawn', payload: { name } })); }
</script>

<main class="container">
    <h1>Protocol Simulator</h1>

    {#if error}
        <p style="color: var(--pico-del-color)">{error}</p>
    {/if}

    <div role="group">
        {#if !connected}
            <button onclick={connect}>Reconnect</button>
        {:else}
            <button onclick={start}>Start</button>
            <button onclick={stop} class="secondary">Stop</button>
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