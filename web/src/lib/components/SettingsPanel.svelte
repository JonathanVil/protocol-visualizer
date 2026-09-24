<script lang="ts">
    import InfoToolTip from "$lib/components/InfoToolTip.svelte";
    import Icon from "@iconify/svelte";
    import {sim} from "$lib/sim.svelte";
    import {notifyError} from "$lib/notifications.svelte";

    let {close}: {close: () => void} = $props();

    // Sliders show the backend's settings, and follow them when they change
    // (e.g. from another tab) unless the user is dragging.
    // Derived primitives only change when the value does, not on every snapshot.
    const backendTicksPerSecond = $derived(Math.round(1000 / sim.settings.tickDurationMs));
    const backendTransitTicks = $derived(sim.settings.transitTicks);

    let ticksPerSecond = $state(0);
    let transitTicks = $state(0);
    $effect(() => {
        ticksPerSecond = backendTicksPerSecond;
    });
    $effect(() => {
        transitTicks = backendTransitTicks;
    });
</script>

<div class="absolute top-18 right-4 rounded-lg bg-white border border-gray-300 p-4 min-w-60">
    <button onclick={close} class="absolute top-2 right-2 rounded-lg hover:bg-blue-200" aria-label="Close settings">
        <Icon icon="mdi:close" class="w-6 h-6 text-black" />
    </button>

    <div class="flex flex-col gap-3">
        <!-- Tick speed -->
        <div class="flex flex-row gap-2 font-medium">
            <p>Ticks / Second</p>
            <InfoToolTip text="The number of simulation ticks processed per second. A tick is the smallest unit in the system"></InfoToolTip>
        </div>
        <div class="flex flex-col">
            <input type="range" min="1" max="40" bind:value={ticksPerSecond}
                   disabled={!sim.connected}
                   onchange={() => sim.setTickDuration(Math.round(1000 / ticksPerSecond)).catch(notifyError)}>
            <p>{ticksPerSecond}</p>
        </div>

        <!-- Transit time -->
        <div class="flex flex-row gap-2 font-medium">
            <p>Transit time</p>
            <InfoToolTip text="The time it takes for a message to travel from the sender to the receiver, measured in ticks. Applies to messages sent from now on."></InfoToolTip>
        </div>
        <div class="flex flex-col">
            <input type="range" min="1" max="50" bind:value={transitTicks}
                   disabled={!sim.connected}
                   onchange={() => sim.setTransitTicks(transitTicks).catch(notifyError)}>
            <p>{transitTicks} ticks</p>
        </div>

        <!-- Drop chance -->
        <div class="flex flex-row gap-2 font-medium opacity-50">
            <p>Drop chance</p>
            <InfoToolTip text="Not supported by the Go simulator yet"></InfoToolTip>
        </div>
        <div class="flex flex-col opacity-50">
            <input type="range" min="0" max="100" value="0" disabled>
            <p>0%</p>
        </div>
    </div>
</div>
