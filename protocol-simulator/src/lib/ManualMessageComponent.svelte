<script>
    import {LinkedList} from '$lib/LinkedList.js';
    import {transitTime} from "$lib/protocolUtils.js";

    /** @typedef {import('$lib/types.js').Message} Message */
    let to = 0;
    let from = 0;
    let type = "";
    export let messages = new LinkedList();

    function sendMessageManual(){
        /** @type {Message} */
        let message = {source: from, destination: to, type: type, transitSteps: transitTime, elapsedSteps: 0}
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