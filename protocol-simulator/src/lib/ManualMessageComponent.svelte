<script>
    import {Queue} from '$lib/Queue.js';
    import {getNextMessageId, getTransitTime, setUpperTransitTime, setLowerTransitTime} from "$lib/protocolUtils.js";

    /** @typedef {import('$lib/types.js').Message} Message */
    let to = 0;
    let from = 0;
    let type = "";

    /** @type {any} */
    let data = "";


    export let messages = new Queue();

    function sendMessageManual(){
        let transitTime = getTransitTime();
        /** @type {Message} */
        let message = {id: getNextMessageId(), source: from, destination: to, type: type, transitSteps: transitTime, elapsedSteps: 0, data: data}
        messages.push(message);
        console.log("Sending message manually");
        console.log(messages);
    }
</script>

<div>
    <input id="to" bind:value={to} placeholder="To actor ID" />
    <input id="from" bind:value={from} placeholder="From actor ID" />
    <input id="type" bind:value={type} placeholder="Message type" />
    <input id="data" bind:value={data} placeholder="Data" />
    <button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            on:click={sendMessageManual}>
        Send
    </button>

</div>