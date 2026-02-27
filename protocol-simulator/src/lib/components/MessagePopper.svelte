<script>
    /** @typedef {import('$lib/types.js').Message} Message */



    /** @type {Message} */
    export let message;

    let delay = 100;

    export let dropMessage;

    export let delayMessage;

    export let deliverMessage;


</script>


<div class="pointer-events-auto relative whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 px-2 py-1.5 text-[16px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]">
    <!-- Message data -->
    <div class="flex flex-col items-center text-center">
        <p class="items-center font-bold">Message</p>
        <div>
            <div class="text-xs">ID: {message.id}</div>
            <div class="">{message.source} ⟶ {message.destination}</div>
        </div>
    </div>

    <hr class="h-px my-2 bg-white border-0">
    <!-- Methods -->
    <div class="flex flex-row gap-2 items-center text-center">
        <button class=" bg-blue-600 text-white rounded hover:bg-blue-700 w-25 h-10 text-base flex text-center justify-center items-center"
                on:click={() => deliverMessage(message)}>
            Deliver
        </button>
        <button class=" bg-blue-600 text-white rounded hover:bg-blue-700 w-25 h-10 text-base flex text-center justify-center items-center"
                on:click={() => dropMessage(message)}>
            Drop
        </button>
    </div>

    <div class="mt-2 flex flex-row items-center text-center gap-2">
        <button class="bg-blue-600 text-white rounded hover:bg-blue-800 w-25 h-10 text-base flex text-center justify-center items-center"
                on:click={() => delayMessage(message, delay)}>
            Delay
        </button>
        <input
                class="border p-1 h-10 w-12 rounded relative z-9999 bg-blue-800"
                id="from"
                bind:value={delay}
                placeholder="ID"
                on:pointerdown|stopPropagation
                on:mousedown|stopPropagation
        />
        <p>
            ticks
        </p>
    </div>


</div>