<script>
    import {getTickSpeed, setTickSpeed, setTransitbounds} from "$lib/protocolUtils.js";


    // the actual values of these is not stored here, so these are not important until updated by user
    let transitLowerInput = 8
    let transitUpperInput = 12

    let tickSpeedInput = getTickSpeed();
    export let tickSpeedUpdated = false;
    function setTickSpeedInput() {
        setTickSpeed(tickSpeedInput);
        tickSpeedUpdated = true;
    }

    function setTransitTimeInput() {
        setTransitbounds(transitUpperInput, transitLowerInput);
    }
</script>

<div id="settingspanel" class="hidden absolute top-22 right-1 rounded-lg w-1/7 h-2/5 bg-[#91B7C7]/16 border">
    <div class="flex flex-col gap-3">
        <div class="font-medium">
            <p class="">Ticks / Second</p>
        </div>
        <div>
            <input type="range" class="w-3/4" min="1" max="40" bind:value={tickSpeedInput} on:input={setTickSpeedInput}>
            <p>{tickSpeedInput}</p>
        </div>
        <div class="font-medium">
            <p>Transit time</p>
        </div>
        <div class="flex">
            <div class="flex flex-col">
                <p>Min</p>
                <input type="range" class="w-1/2" min="1" max="{transitUpperInput - 1}" bind:value={transitLowerInput} on:input={setTransitTimeInput}>
                <p>{transitLowerInput}</p>
            </div>
            <div>
                <p>Max</p>
                <input type="range" class="w-1/2" min="{transitLowerInput}" max="50" bind:value={transitUpperInput} on:input={setTransitTimeInput}>
                <p>{transitUpperInput}</p>
            </div>

        </div>
    </div>
</div>