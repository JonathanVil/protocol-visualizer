<script lang="ts">
    import {onMount} from "svelte";
    import Icon from '@iconify/svelte';
    import SettingsPanel from "$lib/components/SettingsPanel.svelte";
    import NavigationBar from "$lib/components/NavigationBar.svelte";
    import ControlsPanel from "$lib/components/ControlsPanel.svelte";
    import Graph from "$lib/components/Graph.svelte";
    import ManualMessageComponent from "$lib/components/ManualMessageComponent.svelte";
    import SpawnActor from "$lib/components/SpawnActor.svelte";
    import EventLog from "$lib/components/EventLog.svelte";
    import DocumentationViewer from "$lib/components/DocumentationViewer.svelte";
    import {sim} from '$lib/sim.svelte';
    import {notifications} from '$lib/notifications.svelte';
    import docs from '../../docs/BASICS.md?raw';

    onMount(() => {
        sim.connect();
        return () => sim.disconnect();
    });

    const LeftPanelOptions = {
        DOCS: "docs",
        LOG: "log",
        NONE: "none"
    } as const;
    type LeftPanel = typeof LeftPanelOptions[keyof typeof LeftPanelOptions];

    let leftPanel = $state<LeftPanel>(LeftPanelOptions.DOCS);

    function toggleLeftPanel(panel: LeftPanel) {
        leftPanel = leftPanel === panel ? LeftPanelOptions.NONE : panel;
    }

    let settingsPanelOpen = $state(false);
</script>

<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">

<!-- Exactly one viewport tall: the graph fills what the nav bar leaves, and nothing scrolls. -->
<div class="relative flex h-dvh flex-col overflow-hidden">
    <!--Top navigation bar-->
    <NavigationBar/>

    <aside
            class="fixed left-0 top-14 z-50 flex h-full w-14 flex-col items-center gap-2 border-r border-slate-200 bg-white/80 py-3"
    >
        <button
                class="p-2 text-slate-300 rounded-md hover:bg-blue-200 aria-pressed:bg-blue-200 border-blue-500"
                aria-label="Guide"
                aria-pressed={leftPanel === LeftPanelOptions.DOCS}
                onclick={() => toggleLeftPanel(LeftPanelOptions.DOCS)}
        >
            <Icon icon="mdi:help" class="w-6 h-6 text-black" />
        </button>

        <button
                class="p-2 text-slate-300 rounded-md hover:bg-blue-200 aria-pressed:bg-blue-200 border-blue-500"
                aria-label="Log"
                aria-pressed={leftPanel === LeftPanelOptions.LOG}
                onclick={() => toggleLeftPanel(LeftPanelOptions.LOG)}
        >
            <Icon icon="mdi:clipboard-text-outline" class="w-6 h-6 text-black" />
        </button>
    </aside>

    <!--Dotted graph (background)-->
    <div class="cy-wrapper min-h-0 flex-1">
        <Graph />
    </div>

    {#if sim.error && !sim.connected}
        <div class="absolute top-18 left-1/2 -translate-x-1/2 rounded-md border border-red-300 bg-red-50 px-4 py-2 text-sm text-red-800 shadow">
            {sim.error}
        </div>
    {/if}

    <div class="fixed bottom-4 left-1/2 -translate-x-1/2 z-50 flex flex-col gap-2">
        {#each notifications as notification (notification.id)}
            <div class="rounded-md border border-red-300 bg-red-50 px-4 py-2 text-sm text-red-800 shadow">
                {notification.text}
            </div>
        {/each}
    </div>

    {#if leftPanel === LeftPanelOptions.DOCS}
        <div class="absolute top-14 left-14 rounded-lg w-9/20 h-4/5">
            <DocumentationViewer source={docs} />
        </div>
    {:else if leftPanel === LeftPanelOptions.LOG}
        <div class="absolute top-14 left-14 rounded-lg w-9/20 max-h-4/5">
            <EventLog />
        </div>
    {/if}

    {#if !settingsPanelOpen}
        <button onclick={() => settingsPanelOpen = true} class="absolute top-18 right-4 p-1 rounded-lg hover:bg-blue-200" aria-label="Open settings">
            <Icon icon="mdi:cog-outline" class="w-8 h-8 text-black" />
        </button>
    {:else}
        <SettingsPanel close={() => settingsPanelOpen = false} />
    {/if}

    <!--Spawn and message block-->
    <div class="absolute bottom-2 left-14 flex flex-row items-end gap-2">
        <SpawnActor />
        <ManualMessageComponent />
    </div>

    <!-- Bottom right buttons -->
    <ControlsPanel />
</div>

<style>
    .cy-wrapper {
        /* dots*/
        background-color: #ffffff;
        background-image: radial-gradient(#d1d5db 1px, transparent 1px);
        background-size: 30px 30px;
    }

    /* Make the graph fill the wrapper */
    :global(.cy-wrapper > .cy-graph) {
        width: 100% !important;
        height: 100% !important;
    }
</style>
