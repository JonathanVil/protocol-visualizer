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

            // NOTE: Object.is is great for primitives; if you need deep detection for objects,
            // compare formatValue(prev) vs formatValue(value) instead (slower, but catches mutations).
            if (prevByKey.has(key) && !Object.is(prev, value)) {
                versionByKey.set(key, (versionByKey.get(key) ?? 0) + 1);
            }

            prevByKey.set(key, value);
        }
    }
</script>

<div
    class="pointer-events-none whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 px-2 py-1.5 text-[12px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]"
>
    <div class="mb-0.5 font-semibold opacity-90">Actor {actor.id}</div>

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