<script>
    export let paused, startSimulation, pauseSimulation, resetSimulation, tickByOne;
</script>

<div class="absolute bottom-5 right-6 flex flex-col gap-3 text-xs">
    {#if paused}
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
                on:click={tickByOne}>
            <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px"  viewBox="0 0 640 640" width="24px" fill="#1f1f1f"><path d="M566.6 342.6C579.1 330.1 579.1 309.8 566.6 297.3L406.6 137.3C394.1 124.8 373.8 124.8 361.3 137.3C348.8 149.8 348.8 170.1 361.3 182.6L466.7 288L96 288C78.3 288 64 302.3 64 320C64 337.7 78.3 352 96 352L466.7 352L361.3 457.4C348.8 469.9 348.8 490.2 361.3 502.7C373.8 515.2 394.1 515.2 406.6 502.7L566.6 342.7z"/></svg>
            <span>Tick by one</span>
        </button>
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
                on:click={startSimulation}>
            <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#1f1f1f"><path d="M320-200v-560l440 280-440 280Z"/></svg>
            <span>Start Simulation</span>
        </button>
    {:else}
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
                on:click={pauseSimulation}>
            <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#1f1f1f"><path d="M560-200v-560h160v560H560Zm-320 0v-560h160v560H240Z"/></svg>
            <span>Pause Simulation</span>
        </button>
    {/if}
    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
            on:click={resetSimulation}>
        <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 512 512" width="24px" fill="#1f1f1f"><path d="M436.7 74.7L448 85.4 448 32c0-17.7 14.3-32 32-32s32 14.3 32 32l0 128c0 17.7-14.3 32-32 32l-128 0c-17.7 0-32-14.3-32-32s14.3-32 32-32l47.9 0-7.6-7.2c-.2-.2-.4-.4-.6-.6-75-75-196.5-75-271.5 0s-75 196.5 0 271.5 196.5 75 271.5 0c8.2-8.2 15.5-16.9 21.9-26.1 10.1-14.5 30.1-18 44.6-7.9s18 30.1 7.9 44.6c-8.5 12.2-18.2 23.8-29.1 34.7-100 100-262.1 100-362 0S-25 175 75 75c99.9-99.9 261.7-100 361.7-.3z"/></svg>
        <span>Clear Simulation</span>
    </button>
</div>