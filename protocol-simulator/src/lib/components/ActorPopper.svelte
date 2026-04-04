<script>
    /** @typedef {import('$lib/types.js').Actor} Actor */
    /** @typedef {import('svelte/store').Readable<Actor>} ActorReadable */

    import EditActorState from "$lib/components/EditActorState.svelte";
    import RunActorMethod from "$lib/components/RunActorMethod.svelte";
    import { timeoutsStore } from "$lib/stores.js";

    /** @typedef {import('$lib/types.js').TimeoutEntry} TimeOutEntry */

    /** @type {ActorReadable} */
    export let store;

    /** @type {(actor: Actor, color: string) => void} */
    export let toggleAlive;

    /** @type {(c: boolean) => void} */
    export let setStateCollapsedGlobal;

    /** @type {(c: boolean) => void} */
    export let setMethodsCollapsedGlobal;

    /** @type {() => void} */
    export let reposition;

    /** @type {TimeOutEntry[]} */
    $: actorTimeouts = $timeoutsStore.toArray().filter(t => t.actorId === actor?.id);

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
    const excludedAttributes = ['id', 'nodeColor', 'alive', 'protocolName']

    $: actor = $store;

    /** @type {string} */
    $: originalColor = actor.nodeColor;

    $: entries =
        actor
            ? Object.entries(actor).filter(([k, v]) => typeof v !== 'function' && !excludedAttributes.includes(k))
            : [];

    $: methods = getAllMethods(actor);

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
     * @param {MouseEvent} event
     */
    function toggleShowState(event) {
        event?.stopPropagation?.();

        // if user holds shift
        if (event.shiftKey) {
            setStateCollapsedGlobal(!stateCollapsed);
        } else {
            stateCollapsed = !stateCollapsed;
            reposition();
        }
    }

    /**
     * Used by its parent to toggle all state
     * @param {boolean} val
     */
    export function setStateCollapsed(val) {
        stateCollapsed = val;
    }

    /**
     * Used by its parent to toggle all methods
     * @param {boolean} val
     */
    export function setMethodsCollapsed(val) {
        methodsListCollapsed = val;
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

    let methodsListCollapsed = false;

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
     * @param {MouseEvent} event
     */
    function toggleShowMethods(event) {
        event?.stopPropagation?.();

        if(event.shiftKey) {
            setMethodsCollapsedGlobal(!methodsListCollapsed);
        } else {
            methodsListCollapsed = !methodsListCollapsed;
            reposition();
        }
    }

    /**
     * @param name {string}
     * @param args {any[]}
     */
    function runMethod(name, args) {
        if (!actor) return;
        if (!methods.has(name)) return;

        selectedMethod = null;
        console.log(`Running ${name}(${args.join(', ')})`);

        const method = Reflect.get(actor, name);
        if (typeof method === 'function') {
            /** @type {Record<string, any>} */
            const jank = actor; // reassign to please the type checker
            jank[method.name](args);
        } else {
            console.error(`Method ${name} is not a function`);
        }
    }
</script>

<div class="flex flex-col gap-3">
    <div
            class="pointer-events-auto relative whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 pl-2 pr-6 py-1.5 text-[12px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]"
    >
        <div class="absolute flex flex-row right-1 top-1">
            <button
                    type="button"
                    class="inline-flex h-5 w-5 items-center justify-center rounded text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                    aria-label={stateCollapsed ? 'Show popper' : 'Hide popper'}
                    title={stateCollapsed ? 'Show state (Shift + Click to show all)' : 'Hide state (Shift + Click to hide all)'}
                    on:click={toggleShowState}
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
                    aria-label="Show methods"
                    title={methodsListCollapsed ? 'Show methods (Shift + Click to show all)' : 'Hide methods (Shift + Click to hide all)'}
                    on:click={toggleShowMethods}
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
                <div class="mb-0.5 font-semibold opacity-90">Actor {actor?.id} ({actor.protocolName})</div>
                <button class=" bg-blue-600 text-white rounded hover:bg-blue-700 w-13 h-5 text-xs flex text-center justify-center items-center"
                        on:click={() =>
                    {
                       toggleAlive(actor, originalColor)
                    }}>
                    {#if actor.alive}
                        <p>Kill</p>
                    {:else}
                        <p>Resurrect</p>
                    {/if}
                </button>
            </div>

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
                {/if}
            {/if}

            <div class="font-mono">
                {#if !methodsListCollapsed}
                    {#each methods.entries() as [name, val]}
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

                    {#if selectedMethod}
                        <RunActorMethod run={args => selectedMethod && runMethod(selectedMethod[0], args)} cancel={() => selectedMethod = null} methodName={selectedMethod[0]} argumentNames={selectedMethod[1]}  />
                    {/if}
                {/if}
            </div>
        </div>


        <p class="text-white text-xs font-mono mt-2" style="text-shadow: 0 0 3px black, 0 0 3px black;">Timeouts</p>
        <div class="flex flex-row items-end">
            {#if actorTimeouts.length > 0}
                {#each actorTimeouts as t, i}
                    {@const radius = 12}
                    {@const circumference = 2 * Math.PI * radius}
                    {@const progress = t.ticks / t.totalTicks}
                    {@const color = "#e24b4a"}
                    <svg width="40" height="40" viewBox="0 0 40 40">
                        <circle cx="18" cy="18" r={radius}
                                fill="none"
                                stroke="rgba(255,255,255,0.08)"
                                stroke-width="3"
                        />
                        <!-- countdown arc -->
                        <circle cx="18" cy="18" r={radius}
                                fill="none"
                                stroke={color}
                                stroke-width="3"
                                stroke-linecap="round"
                                stroke-dasharray={circumference}
                                stroke-dashoffset={circumference * (1 - progress)}
                                transform="rotate(-90 18 18)"
                        />
                    </svg>
                    <span class="text-white font-mono"
                          style="font-size: 9px; text-shadow: 0 0 3px #000">
                {t.ticks} ticks
            </span>
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