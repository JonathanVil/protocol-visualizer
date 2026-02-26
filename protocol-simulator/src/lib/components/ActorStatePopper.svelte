<script>
    /** @typedef {import('$lib/types.js').Actor} Actor */
    /** @typedef {import('svelte/store').Readable<Actor>} ActorReadable */

    import EditActorState from "$lib/components/EditActorState.svelte";
    import RunActorMethod from "$lib/components/RunActorMethod.svelte";

    /** @type {ActorReadable} */
    export let store;

    /** @param {any} v */
    export function formatValue(v) {
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

    let stateCollapsed = false;

    /**
     * @param {event} event
     */
    function toggleShowState(event) {
        event?.stopPropagation?.();
        stateCollapsed = !stateCollapsed;
    }

    /** @type {string | null} */
    let editingKey = null;

    /** @type {string} */
    let editText = '';

    /** @type {any} */
    let editOriginalValue;

    /** @param {event} event
     *  @param {string} key
     * */
    function openEdit(event, key) {
        event?.stopPropagation?.();
        editingKey = key;
        const value = Reflect.get(actor, key);
        editText = typeof value === 'string' ? value : formatValue(value);
        editOriginalValue = value;
    }

    /** @param {any} newValue */
    function saveEdit(newValue) {
        if (!actor || !editingKey) return;
        Reflect.set(actor, editingKey, newValue);
        console.log(`Saving ${editingKey} = ${newValue}`);
    }

    let methodsListOpen = false;

    /** @type {[string, string[]] | null} */
    let selectedMethod = null;
    /**
    * @param {Object} obj
    */
    function getAllMethods(obj) {
        if (!obj) return new Map();

        /**
         * Extract argument names from a function.
         * Handles: `foo(a,b)`, `function foo(a,b)`, `(a,b)=>`, `a=>`
         * Not perfect for every JS edge case, but good enough for UI display.
         * @param {Function} fn
         * @returns {string[]}
         */
        function getArgNames(fn) {
            const src = Function.prototype.toString.call(fn).trim();

            // Arrow: single param without parens:  x => ...
            const singleArrow = src.match(/^([A-Za-z_$][\w$]*)\s*=>/);
            if (singleArrow) return [singleArrow[1]];

            // Anything with (...) up front: function/method/arrow with parens
            const paren = src.match(/^[^(]*\(\s*([^)]*)\)/);
            const raw = (paren?.[1] ?? '').trim();
            if (!raw) return [];

            return raw
                .split(',')
                .map((s) => s.trim())
                .filter(Boolean)
                // strip default values: `a = 1` -> `a`
                .map((s) => s.replace(/\s*=.*$/, '').trim())
                // strip rest: `...args` -> `args`
                .map((s) => s.replace(/^\.\.\./, '').trim());
        }

        /** @type {Map<string, string[]>} */
        const namesToArgs = new Map();

        let proto = Object.getPrototypeOf(obj);
        for (const key of Reflect.ownKeys(proto)) {
            if (key === 'constructor') continue;

            const desc = Object.getOwnPropertyDescriptor(proto, key);
            if (!desc) continue;

            if (typeof desc.value === 'function') {
                namesToArgs.set(String(key), getArgNames(desc.value));
            }
        }

        return namesToArgs;
    }

    /**
     * @param {event} event
     */
    function toggleShowMethods(event) {
        event?.stopPropagation?.();
        methodsListOpen = !methodsListOpen;
    }
</script>

<div
    class="pointer-events-auto relative whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 pl-2 pr-6 py-1.5 text-[12px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]"
>
    <div class="absolute flex flex-row right-1 top-1">
        <button
            type="button"
            class="inline-flex h-5 w-5 items-center justify-center rounded text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
            aria-label={stateCollapsed ? 'Show popper' : 'Hide popper'}
            title={stateCollapsed ? 'Show' : 'Hide'}
            on:click={toggleShowState}
        >
            {#if stateCollapsed}
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

        <button
                type="button"
                class="inline-flex h-5 w-5 items-center justify-center rounded text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                aria-label="Show methods"
                title="Show methods"
                on:click={toggleShowMethods}
        >
            <i class="fa fa-play"></i>
        </button>
    </div>

    <div class="pr-6">
        <div class="mb-0.5 font-semibold opacity-90">Actor {actor?.id}</div>

        {#if !stateCollapsed}
            {#if !actor}
                <div class="font-mono opacity-95"><span class="opacity-90">Actor</span>: <span>null</span></div>
            {:else}
                {#each entries as [key, value] (key)}
                    <div class="font-mono opacity-95 flex items-center gap-1">
                        <span class="opacity-90">{key}</span>:
                        {#key versionByKey.get(key) ?? 0}
                            <span class="flash">{formatValue(value)}</span>
                        {/key}

                        <button
                            type="button"
                            class="ml-1 inline-flex h-5 w-5 items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                            aria-label={"Edit " + key}
                            title={"Edit " + key}
                            on:click={(e) => openEdit(e, key)}
                        >
                            <!-- pencil icon -->
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M12 20h9" />
                                <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
                            </svg>
                        </button>
                    </div>
                {/each}

                {#if editingKey}
                    <EditActorState save={v => saveEdit(v)} bind:editingKey={editingKey} bind:editText={editText} bind:editOriginalValue={editOriginalValue} />
                {/if}

                {#if selectedMethod}
                    <RunActorMethod methodName={selectedMethod[0]} argumentNames={selectedMethod[1]}  />
                {/if}
            {/if}
        {/if}

        <div class="font-mono">
            {#if methodsListOpen}
                {#each getAllMethods(actor).entries() as [name, val]}
                    <div class="font-mono opacity-95 flex items-center gap-1">
                        <span class="opacity-90">{name}({val.join(', ')})</span>

                        <button
                                type="button"
                                class="ml-1 inline-flex h-5 w-5 items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                                aria-label={"Run " + name}
                                title={"Run " + name}
                                on:click={() => selectedMethod = [name, val]}
                        >
                            <i class="fa fa-play"></i>
                        </button>
                    </div>
                {:else}
                    <p>No functions found</p>
                {/each}
            {/if}
        </div>
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