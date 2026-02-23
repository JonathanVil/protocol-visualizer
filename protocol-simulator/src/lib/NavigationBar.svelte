<script>
    export let predefinedProtocols;

    /**
     * @type {{ name: string; content: string; } | null}
     */
    let selectedProtocol = null;

    export let sourceCode;

    export let codePanelOpen;

</script>

<header class="h-14 bg-white shadow flex items-center justify-between px-6">
    <div class="flex items-center gap-3">
        <div class="w-6 h-6 bg-gray-300 rounded"></div>
        <h1 class="text-lg font-semibold">Protocol Simulator</h1>
    </div>
    <div class="flex items-center justify-center gap-1">

        <select class=" p-2 border border-gray-300 rounded w-fit" bind:value={selectedProtocol}>
            {#each predefinedProtocols as protocol}
                <option value={protocol}>{protocol.name}</option>
            {/each}
        </select>

        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700" on:click={() => {codePanelOpen = true; sourceCode = selectedProtocol?.content ?? ""}}>Load</button>
    </div>

    <a href="https://github.com/JonathanVil/protocol-visualizer" aria-label="GitHub">
        <i class="fa fa-github" style="font-size:36px"></i>
    </a>
</header>