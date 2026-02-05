<script>
    import {LinkedList} from '$lib/LinkedList.js';
    /** @typedef {import('$lib/types.js').Message} Message */
    /** {number} */
    let to = 0;
    let from = 0;
    let type = "";
    export let messages = new LinkedList();

    function sendMessageManual(){
        /** @type {Message} */
        let message = {source: from, destination: to, type: type, transitSteps: 100, elapsedSteps: 0}
        messages.append(message);
        console.log("Sending message manually");
        console.log(messages);
    }
</script>

<div>
    <input id="to" bind:value={to} placeholder="To actor ID" />
    <input id="from" bind:value={from} placeholder="From actor ID" />
    <input id="type" bind:value={type} placeholder="Message type" />
    <button on:click={sendMessageManual}>Send</button>

</div>