<script>
    import InfoToolTip from "$lib/components/InfoToolTip.svelte";
    import {sim} from "$lib/sim.svelte.js";
    import {parseLoose} from "$lib/format.js";
    import {notifyError} from "$lib/notifications.svelte.js";

    /** @type {number | undefined} */
    let from = $state();
    /** @type {number | undefined} */
    let to = $state();
    let payload = $state("");

    const canSend = $derived(sim.connected && from !== undefined && to !== undefined);

    function submit() {
        if (from === undefined || to === undefined) return;
        sim.sendMessage(from, to, parseLoose(payload)).catch(notifyError);
    }

    const selectClass = "h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20";
</script>

<div class="bg-white border border-gray-300 flex flex-row items-end gap-4 p-2 rounded-md">
    <div class="flex flex-col w-16 gap-1">
        <label for="from" class="text-xs font-medium text-slate-600">From</label>
        <select id="from" class={selectClass} bind:value={from}>
            {#each sim.actors as actor (actor.id)}
                <option value={actor.id}>{actor.id}</option>
            {/each}
        </select>
    </div>

    <div class="flex flex-col w-16 gap-1">
        <label for="to" class="text-xs font-medium text-slate-600">To</label>
        <select id="to" class={selectClass} bind:value={to}>
            {#each sim.actors as actor (actor.id)}
                <option value={actor.id}>{actor.id}</option>
            {/each}
        </select>
    </div>

    <div class="flex flex-col w-40 gap-1">
        <div class="flex flex-row items-end justify-between">
            <label for="payload" class="text-xs font-medium text-slate-600">Payload</label>
            <InfoToolTip text={'The message payload. Valid JSON (e.g. 5, true, {"type": "ping"}) is sent as that value; anything else is sent as a string.'} />
        </div>
        <input
            class={selectClass}
            id="payload"
            bind:value={payload}
            placeholder="ping"
        />
    </div>

    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={!canSend}
            onclick={submit}>
        Send
    </button>
</div>
