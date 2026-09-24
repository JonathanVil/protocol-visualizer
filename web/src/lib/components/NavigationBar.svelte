<script>
    import {sim, DEFAULT_URL} from "$lib/sim.svelte.js";
</script>

<header class="h-14 bg-white shadow flex items-center justify-between px-6">
    <div class="flex items-center gap-3">
        <div class="w-6 h-6 bg-gray-300 rounded"></div>
        <h1 class="text-lg font-semibold">Protocol Simulator</h1>
    </div>

    <div class="flex items-center gap-4">
        <div class="flex items-center gap-2 text-sm" title={DEFAULT_URL}>
            <span class="inline-block w-2.5 h-2.5 rounded-full {sim.connected ? 'bg-green-500' : 'bg-red-500'}"></span>
            {#if sim.connected}
                <span>Connected · tick {sim.tick} · {sim.running ? 'running' : 'paused'}</span>
            {:else}
                <span>Disconnected</span>
                <button class="px-2 py-1 rounded border border-slate-300 hover:bg-blue-100" onclick={() => sim.connect()}>
                    Reconnect
                </button>
            {/if}
        </div>

        <a href="https://github.com/JonathanVil/protocol-visualizer" aria-label="GitHub">
            <i class="fa fa-github" style="font-size:36px"></i>
        </a>
    </div>
</header>
