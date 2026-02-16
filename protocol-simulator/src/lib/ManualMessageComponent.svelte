<script>
    import {Queue} from '$lib/Queue.js';
    import {getNextMessageId, getTransitTime} from "$lib/protocolUtils.js";

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

<div class="flex flex-row gap-12">


    <div class="flex flex-row gap-3 h-full w-full">
        <div class="basis-[5%]">
            From
            <input class="border max-w-15" id="from" bind:value={from} placeholder="From actor ID" />
        </div>
        <div class="basis-[10%]">
            To
            <input class="border max-w-15" id="to" bind:value={to} placeholder="To actor ID" />
        </div>
        <div class="basis-[33%]">
            Message Type
            <input class="border" id="type" bind:value={type} placeholder="Message type" />
        </div>
        <div class="basis-[33%] ">
            <p>Data</p>
            <input class="border" id="data" bind:value={data} placeholder="Data" />

        </div>



    </div>

    <button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            on:click={sendMessageManual}>
        Send
    </button>




</div>