<script lang="ts">
    import {untrack} from "svelte";
    import EditActorState from "$lib/components/EditActorState.svelte";
    import RunActorMethod from "$lib/components/RunActorMethod.svelte";
    import {sim, type Actor, type MethodInfo} from "$lib/sim.svelte";
    import {formatValue} from "$lib/format";
    import {notifyError} from "$lib/notifications.svelte";

    interface Props {
        actor: Actor | undefined;
        setStateCollapsedGlobal: (collapsed: boolean) => void;
        setMethodsCollapsedGlobal: (collapsed: boolean) => void;
        reposition: () => void;
    }

    let {actor, setStateCollapsedGlobal, setMethodsCollapsedGlobal, reposition}: Props = $props();

    const entries = $derived(actor ? Object.entries(actor.fields ?? {}) : []);
    const methods = $derived(actor?.methods ?? []);

    // --- flash-on-change bookkeeping ---
    const prevByKey = new Map<string, unknown>();
    let versionByKey = $state<Record<string, number>>({});

    $effect.pre(() => {
        for (const [key, value] of entries) {
            const prev = prevByKey.get(key);
            if (prevByKey.has(key) && JSON.stringify(prev) !== JSON.stringify(value)) {
                untrack(() => versionByKey[key] = (versionByKey[key] ?? 0) + 1);
            }
            prevByKey.set(key, value);
        }
    });

    let stateCollapsed = $state(true);
    let methodsListCollapsed = $state(true);

    function toggleShowState(event: MouseEvent) {
        event.stopPropagation();
        if (event.shiftKey) {
            setStateCollapsedGlobal(!stateCollapsed);
        } else {
            stateCollapsed = !stateCollapsed;
            reposition();
        }
    }

    function toggleShowMethods(event: MouseEvent) {
        event.stopPropagation();
        if (event.shiftKey) {
            setMethodsCollapsedGlobal(!methodsListCollapsed);
        } else {
            methodsListCollapsed = !methodsListCollapsed;
            reposition();
        }
    }

    /** Used by its parent to toggle all state */
    export function setStateCollapsed(val: boolean) {
        stateCollapsed = val;
    }

    /** Used by its parent to toggle all methods */
    export function setMethodsCollapsed(val: boolean) {
        methodsListCollapsed = val;
    }

    let editingKey = $state<string | null>(null);
    let editText = $state('');
    let editOriginalValue = $state<unknown>();

    function openEdit(event: MouseEvent, key: string, value: unknown) {
        event.stopPropagation();
        editingKey = key;
        editText = typeof value === 'string' ? value : formatValue(value);
        editOriginalValue = value;
    }

    function saveEdit(key: string, newValue: unknown) {
        if (!actor) return;
        sim.setField(actor.id, key, newValue).catch(notifyError);
    }

    let selectedMethod = $state<MethodInfo | null>(null);

    /** Result of the most recent method call, shown under the method list. */
    let lastResult = $state('');

    function runMethod(name: string, args: unknown[]) {
        if (!actor) return;
        selectedMethod = null;
        sim.invoke(actor.id, name, args)
            .then((result) => {
                lastResult = result === null || result === undefined
                    ? `${name}() done`
                    : `${name}() → ${formatValue(result)}`;
            })
            .catch(notifyError);
    }
</script>

{#if actor}
<div class="flex flex-col gap-3">
    <div
            class="pointer-events-auto relative whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 pl-2 pr-6 py-1.5 text-[12px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]"
    >
        <div class="absolute flex flex-row right-1 top-1">
            <button
                    type="button"
                    class="inline-flex h-5 w-5 items-center justify-center rounded text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                    aria-label={stateCollapsed ? 'Show state' : 'Hide state'}
                    title={stateCollapsed ? 'Show state (Shift + Click to show all)' : 'Hide state (Shift + Click to hide all)'}
                    onclick={toggleShowState}
            >
                <!-- "show" icon (eye) -->
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7Z" />
                    <circle cx="12" cy="12" r="3" />
                </svg>
                <span
                        class="absolute h-[2px] w-5 rounded bg-current transition-all duration-200 ease-in-out {stateCollapsed ? 'rotate-0 opacity-0 scale-75' : '-rotate-45 opacity-100 scale-100'}"
                        aria-hidden="true"
                ></span>
            </button>

            <button
                    type="button"
                    class="inline-flex h-5 w-5 items-center justify-center rounded text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                    aria-label={methodsListCollapsed ? 'Show methods' : 'Hide methods'}
                    title={methodsListCollapsed ? 'Show methods (Shift + Click to show all)' : 'Hide methods (Shift + Click to hide all)'}
                    onclick={toggleShowMethods}
            >
                <i class="fa fa-terminal"></i>
                <span
                        class="absolute h-[2px] w-5 rounded bg-current transition-all duration-200 ease-in-out {methodsListCollapsed ? 'rotate-0 opacity-0 scale-75' : '-rotate-45 opacity-100 scale-100'}"
                        aria-hidden="true"
                ></span>
            </button>
        </div>

        <div class="pr-6">
            <div class="flex flex-row items-center gap-28">
                <div class="mb-0.5 font-semibold opacity-90">Actor {actor.id} ({actor.typeName})</div>
                <button class="bg-blue-600 text-white rounded w-13 h-5 text-xs flex text-center justify-center items-center opacity-50 cursor-not-allowed"
                        disabled
                        title="Killing actors is not supported by the Go simulator yet">
                    Kill
                </button>
            </div>

            {#if !stateCollapsed}
                {#each entries as [key, value] (key)}
                    <div class="font-mono opacity-95 flex items-center gap-1">
                        <span class="opacity-90">{key}</span>:
                        {#key versionByKey[key] ?? 0}
                            <span class="flash">{formatValue(value)}</span>
                        {/key}

                        <button
                                type="button"
                                class="ml-1 inline-flex h-5 w-5 items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                                aria-label={"Edit " + key}
                                title={"Edit " + key}
                                onclick={(e) => openEdit(e, key, value)}
                        >
                            <!-- pencil icon -->
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M12 20h9" />
                                <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
                            </svg>
                        </button>
                    </div>
                {:else}
                    <p class="font-mono opacity-70">No exported fields</p>
                {/each}

                {#if editingKey}
                    <EditActorState save={(v) => editingKey && saveEdit(editingKey, v)} bind:editingKey bind:editText bind:editOriginalValue />
                {/if}
            {/if}

            {#if !methodsListCollapsed}
                <div class="font-mono">
                    {#each methods as method (method.name)}
                        <div class="font-mono opacity-95 flex items-center gap-1">
                            <span class="opacity-90">{method.name}({method.args.join(', ')})</span>

                            <button
                                    type="button"
                                    class="ml-1 inline-flex h-5 w-5 items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                                    aria-label={"Run " + method.name}
                                    title={"Run " + method.name}
                                    onclick={() => selectedMethod = method}
                            >
                                <i class="fa fa-play"></i>
                            </button>
                        </div>
                    {:else}
                        <p>No exported methods</p>
                    {/each}

                    {#if lastResult}
                        <p class="mt-1 opacity-70">{lastResult}</p>
                    {/if}

                    {#if selectedMethod}
                        {@const method = selectedMethod}
                        {#key method.name}
                            <RunActorMethod run={(args) => runMethod(method.name, args)} cancel={() => selectedMethod = null} methodName={method.name} argumentTypes={method.args} />
                        {/key}
                    {/if}
                </div>
            {/if}
        </div>
    </div>
</div>
{/if}

<style>
    @keyframes flash {
        0% {
            background: rgba(250, 204, 21, 0.55); /* amber-ish */
            color: white;
        }
        100% {
            background: transparent;
            color: inherit;
        }
    }

    .flash {
        animation: flash 450ms ease-out;
        border-radius: 0.25rem;
        padding: 0 0.15rem;
    }
</style>
