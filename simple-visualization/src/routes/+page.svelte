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

    onMount(() => {
        ws = new WebSocket('ws://localhost:8067/ws');
        ws.onmessage = (e: MessageEvent) => {
            events = [...events, JSON.parse(e.data) as Event];
        };
        return () => ws.close();
    });

    function start() { ws.send('start'); }
    function stop() { ws.send('stop'); }
</script>

<button onclick={start}>Start</button>
<button onclick={stop}>Stop</button>

<ul>
    {#each events as event}
        <li>Tick {event.Tick}: Node {event.Message.From} → Node {event.Message.To} — {String(event.Message.Payload)}</li>
    {/each}
</ul>