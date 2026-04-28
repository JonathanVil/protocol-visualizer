<script>
    import InfoToolTip from "$lib/components/InfoToolTip.svelte";

    /** @type {number} */
    export let tickSpeed;

    /** @type {number} */
    export let transitLower = 20;

    /** @type {number} */
    export let transitUpper = 20;
</script>

<div class="absolute top-22 right-1 rounded-lg w-1/7 h-2/5 bg-[#91B7C7]/16 border">
    <div class="flex flex-col gap-3">
        <div class="flex flex-row gap-2 font-medium">
            <p class="">Ticks / Second</p>
            <InfoToolTip text="The number of simulation ticks processed per second. A tick is the smallest unit in the system"></InfoToolTip>
        </div>
        <div>
            <input type="range" class="w-3/4" min="1" max="40" bind:value={tickSpeed}>
            <p>{tickSpeed}</p>

        </div>
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
    </div>
</div>