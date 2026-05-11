<script>
    import InfoToolTip from "$lib/components/InfoToolTip.svelte";
    import Icon from "@iconify/svelte";

    /** @type {number} */
    export let tickSpeed;

    /** @type {number} */
    export let transitLower = 20;

    /** @type {number} */
    export let transitUpper = 20;

    /** @type {number} */
    export let dropChance = 0;

    export let close;
</script>
<!--Tickspeed -->
<div class="absolute top-18 right-4 rounded-lg bg-white border border-gray-300 p-2">
    <button on:click={() => close()} class="absolute top-2 right-2 rounded-lg hover:bg-blue-200">
        <Icon icon="mdi:close" class="w-6 h-6 text-black" />
    </button>

    <div class="flex flex-col gap-3">
        <div class="flex flex-row gap-2 font-medium">
            <p class="">Ticks / Second</p>
            <InfoToolTip text="The number of simulation ticks processed per second. A tick is the smallest unit in the system"></InfoToolTip>
        </div>
        <div>
            <input type="range" class="w-3/4" min="1" max="40" bind:value={tickSpeed}>
            <p>{tickSpeed}</p>

        </div>

        <!--Transit bounds-->
        <div class="flex flex-row gap-2 font-medium">
            <p>Transit time</p>
            <InfoToolTip text="The time it takes for a message to travel from the sender to the receiver, measured in simulation units."></InfoToolTip>
        </div>
        <div class="flex">
            <div class=" w-2/7">
                <p>Min</p>
                <p>20</p>
            </div>
            <div>
                <p>Max</p>
                <input type="range" class="w-1/2" min="{transitLower}" max="50" bind:value={transitUpper}>
                <p>{transitUpper}</p>
            </div>
        </div>

        <!-- Drop chance-->
        <div class="flex flex-row gap-2 font-medium">
            <p class="">Drop chance</p>
            <InfoToolTip text="The chance that any message is dropped when it would otherwise be delivered"></InfoToolTip>
        </div>
        <div>
            <input type="range" class="w-3/4" min="0" max="100" bind:value={dropChance}>
            <p>{dropChance}</p>

        </div>
    </div>
</div>