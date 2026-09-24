<script lang="ts">
    import InfoToolTip from "$lib/components/InfoToolTip.svelte";
    import {sim} from "$lib/sim.svelte";
    import {notifyError} from "$lib/notifications.svelte";

    let selectedType = $state("");

    // Default to the first registered type once the types arrive.
    $effect(() => {
        if (!sim.actorTypes.includes(selectedType)) selectedType = sim.actorTypes[0] ?? "";
    });
</script>

<div class="bg-white border border-gray-300 flex flex-row items-end gap-4 p-2 rounded-md">
    <div class="flex flex-col w-40 gap-1">
        <div class="flex flex-row items-end justify-between">
            <label for="actor-type" class="text-xs font-medium text-slate-600">Actor type</label>
            <InfoToolTip text="Actor types registered by your Go program with RegisterActorType" />
        </div>
        <select
            id="actor-type"
            class="h-9 rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
            bind:value={selectedType}
        >
            {#each sim.actorTypes as type (type)}
                <option value={type}>{type}</option>
            {:else}
                <option value="" disabled>No types registered</option>
            {/each}
        </select>
    </div>

    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={!sim.connected || !selectedType}
            onclick={() => sim.spawn(selectedType).catch(notifyError)}>
        Spawn
    </button>
</div>
