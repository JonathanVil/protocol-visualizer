<script>
    import InfoToolTip from "$lib/components/InfoToolTip.svelte";

    let to = 0;
    let from = 0;
    let type = "";

    /** @type {any} */
    let data = null;

    /** @type {(from: number, to: number, data: any, type: string) => void} */
    export let send;

    function submit() {
        send(from, to, type, data);
    }
</script>

<div class="bg-white border border-gray-300 flex flex-row items-end gap-4 p-2 rounded-md">

    <div class="flex flex-col w-12 gap-1">
        <label for="from" class="text-xs font-medium text-slate-600">From</label>
        <input
            class="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
            id="from"
            bind:value={from}
            placeholder="ID"
        />
    </div>

    <div class="flex flex-col w-12 gap-1">
        <label for="to" class="text-xs font-medium text-slate-600">To</label>
        <input
            class="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
            id="to"
            bind:value={to}
            placeholder="ID"
        />
    </div>

    <div class="flex flex-col w-30 gap-1">
        <label for="type" class="text-xs font-medium text-slate-600">Message Type</label>
        <input
            class="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
            id="type"
            bind:value={type}
            placeholder="Message type"
        />
    </div>

    <div class="flex flex-col w-30 gap-1">
        <div class="flex flex-row items-end justify-between">
            <label for="data" class="text-xs font-medium text-slate-600">Data</label>
            <InfoToolTip text="Enter message content. Click 'Send' to put the message in transit in the network" />
        </div>

        <input
            class="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
            id="data"
            bind:value={data}
            placeholder="Data"
        />
    </div>

    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center text-sm font-medium" on:click={submit}>
        Send
    </button>

</div>
