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
    /** @type {string[]} */
    export let tickLog = [];

    function sendMessageManual(){
        let logEntry = `Manually sent: Actor ${from} sent msg ${type} with data ${data} to Actor ${to}`
        console.log(logEntry);
        tickLog.push(logEntry);
        let transitTime = getTransitTime();
        /** @type {Message} */
        let message = {id: getNextMessageId(), source: from, destination: to, type: type, transitTicks: transitTime, elapsedTicks: 0, data: data}
        messages.push(message);
    }
</script>

<div class="absolute bottom-2 left-1 flex flex-row items-center gap-1 text-md">
    <div class="bg-white border border-gray-300 flex items-center gap-4 p-1 rounded-md h-14">
        <div class="flex flex-row gap-4">

            <div class="flex flex-col w-12">
                From
                <input class="border p-1 h-7" id="from" bind:value={from} placeholder="ID" />
            </div>

            <div class="flex flex-col w-12">
                To
                <input class="border p-1 h-7" id="to" bind:value={to} placeholder="ID" />
            </div>

            <div class="flex flex-col w-30">
                Message Type
                <input class="border p-1 h-7" id="type" bind:value={type} placeholder="Message type" />
            </div>

            <div class="flex flex-col w-30">
                Data
                <input class="border p-1 h-7" id="data" bind:value={data} placeholder="Data" />
            </div>

        </div>
    </div>

    <button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            on:click={sendMessageManual}>
        Send
    </button>

</div>
