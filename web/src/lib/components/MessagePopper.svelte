<script lang="ts">
    import Icon from "@iconify/svelte";
    import {sim, type InTransitMsg} from "$lib/sim.svelte";
    import {formatValue} from "$lib/format";
    import {notifyError} from "$lib/notifications.svelte";

    interface Props {
        message: InTransitMsg | undefined;
        close: () => void;
    }

    let {message, close}: Props = $props();

    let delay = $state(10);

    function closeAfter(command: Promise<unknown>) {
        command.then(close).catch(notifyError);
    }
</script>

{#if message}
<div class="pointer-events-auto relative z-50 whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 px-2 py-1.5 text-[16px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]">
    <!-- Message data -->
    <div class="flex flex-col items-center text-center">
        <p class="items-center font-bold">Message</p>

        <button class="absolute right-0 top-0 bg-white/20 rounded-full p-1 m-1 hover:bg-white/30"
                aria-label="Close"
                onclick={close}>
            <Icon icon="mdi:close" class="w-6 h-6 text-white" />
        </button>

        <div>
            <div class="text-xs">ID: {message.id}</div>
            <div>{message.from} ⟶ {message.to}</div>
            <div class="font-mono text-sm">{formatValue(message.payload)}</div>
            <div class="text-xs opacity-70">Sent tick {message.sentTick}, arrives tick {message.deliverAtTick}</div>
        </div>
    </div>

    <hr class="h-px my-2 bg-white border-0">
    <!-- Methods -->
    <div class="flex flex-row gap-2 items-center text-center">
        <button class="bg-blue-600 text-white rounded hover:bg-blue-700 w-25 h-10 text-base flex text-center justify-center items-center"
                onclick={() => closeAfter(sim.deliverNow(message.id))}>
            Deliver
        </button>
        <button class="bg-blue-600 text-white rounded hover:bg-blue-700 w-25 h-10 text-base flex text-center justify-center items-center"
                onclick={() => closeAfter(sim.dropMessage(message.id))}>
            Drop
        </button>
    </div>

    <div class="mt-2 flex flex-row items-center text-center gap-2">
        <button class="bg-blue-600 text-white rounded hover:bg-blue-800 w-25 h-10 text-base flex text-center justify-center items-center"
                onclick={() => sim.delayMessage(message.id, delay).catch(notifyError)}>
            Delay
        </button>
        <input
                class="border p-1 h-10 w-12 rounded bg-blue-800"
                type="number"
                min="1"
                aria-label="Delay in ticks"
                bind:value={delay}
                onpointerdown={(e) => e.stopPropagation()}
                onmousedown={(e) => e.stopPropagation()}
        />
        <p>ticks</p>
    </div>
</div>
{/if}
