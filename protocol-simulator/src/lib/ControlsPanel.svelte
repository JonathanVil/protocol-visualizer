<script>
    export let paused, startSimulation, pauseSimulation, resetSimulation;
</script>

<div class="absolute bottom-5 right-6 flex flex-col gap-3 text-xs">
    {#if paused}
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
                on:click={startSimulation}>
            <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#1f1f1f"><path d="M320-200v-560l440 280-440 280Z"/></svg>
            <span>Start Simulator</span>
        </button>
    {:else}
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
                on:click={pauseSimulation}>
            <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#1f1f1f"><path d="M560-200v-560h160v560H560Zm-320 0v-560h160v560H240Z"/></svg>
            <span>Pause Simulator</span>
        </button>
    {/if}
    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            on:click={resetSimulation}>
        Reset simulation
    </button>
</div>