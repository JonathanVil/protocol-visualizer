<script>
    /** @typedef {import('$lib/types.js').Actor} Actor */
    /** @typedef {import('svelte/store').Readable<Actor>} ActorReadable */

    /** @type {ActorReadable} */
    export let store;

    /** @param {any} v */
    function formatValue(v) {
        if (v === null) return 'null';
        if (v === undefined) return 'undefined';
        if (typeof v === 'string') return JSON.stringify(v);
        if (typeof v === 'function') return '[Function]';
        try {
            return typeof v === 'object' ? JSON.stringify(v) : String(v);
        } catch {
            return String(v);
        }
    }

    // these attributes are not shown in the popper
    const excludedAttributes = ['id', 'nodeColor']

    $: actor = $store;

    $: entries =
        actor
            ? Object.entries(actor).filter(([k, v]) => typeof v !== 'function' && !excludedAttributes.includes(k))
            : [];

    // --- flash-on-change bookkeeping ---
    /** @type {Map<string, any>} */
    let prevByKey = new Map();

    /** @type {Map<string, number>} */
    let versionByKey = new Map();

    $: if (actor) {
        for (const [key, value] of entries) {
            const prev = prevByKey.get(key);

            if (prevByKey.has(key) && !Object.is(prev, value)) {
                versionByKey.set(key, (versionByKey.get(key) ?? 0) + 1);
            }

            prevByKey.set(key, value);
        }
    }

    let collapsed = false;

    /** @param {event} event */
    function toggleCollapsed(event) {
        event?.stopPropagation?.();
        collapsed = !collapsed;
    }
</script>

<div
    class="pointer-events-auto relative whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 px-2 py-1.5 text-[12px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]"
>
    <button
        type="button"
        class="absolute right-1 top-1 inline-flex h-5 w-5 items-center justify-center rounded text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
        aria-label={collapsed ? 'Show popper' : 'Hide popper'}
        title={collapsed ? 'Show' : 'Hide'}
        on:click={toggleCollapsed}
    >
        {#if collapsed}
            <!-- "show" icon (eye) -->
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7Z" />
                <circle cx="12" cy="12" r="3" />
            </svg>
        {:else}
            <!-- "hide" icon (eye off) -->
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M10.58 10.58A2 2 0 0 0 12 14a2 2 0 0 0 1.42-.58" />
                <path d="M9.88 4.24A10.94 10.94 0 0 1 12 5c6.5 0 10 7 10 7a18.2 18.2 0 0 1-2.16 3.19" />
                <path d="M6.61 6.61C3.7 8.58 2 12 2 12s3.5 7 10 7c1.7 0 3.23-.32 4.58-.82" />
                <path d="M2 2l20 20" />
            </svg>
        {/if}
    </button>

    <div class="pr-6">
        <div class="mb-0.5 font-semibold opacity-90">Actor {actor?.id}</div>

        {#if !collapsed}
            {#if !actor}
                <div class="font-mono opacity-95"><span class="opacity-90">Actor</span>: <span>null</span></div>
            {:else}
                {#each entries as [key, value] (key)}
                    <div class="font-mono opacity-95">
                        <span class="opacity-90">{key}</span>:
                        {#key versionByKey.get(key) ?? 0}
                            <span class="flash">{formatValue(value)}</span>
                        {/key}
                    </div>
                {/each}
            {/if}
        {/if}
    </div>
</div>

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