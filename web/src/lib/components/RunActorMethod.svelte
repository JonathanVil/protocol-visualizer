<script>
    import {parseLoose} from "$lib/format.js";

    /** @type {string} */
    export let methodName;

    /** Go parameter types, e.g. "int", "string", "bool". @type {string[]} */
    export let argumentTypes = [];

    /** @type {(args: unknown[]) => void} */
    export let run;

    /** @type {() => void} */
    export let cancel;

    /** @type {string[]} */
    let values = [];
    $: values = argumentTypes.map((_, i) => values[i] ?? '');

    /** @param {string} type */
    function isNumeric(type) {
        return /^(u?int(8|16|32|64)?|float(32|64)|uintptr|byte|rune)$/.test(type);
    }

    /**
     * @param {string} type
     * @param {string} value
     */
    function isInvalid(type, value) {
        if (value.trim() === '') return false;
        if (isNumeric(type)) return Number.isNaN(Number(value));
        if (type === 'bool') return value !== 'true' && value !== 'false';
        return false;
    }

    /**
     * Converts the text input to the JSON value the backend expects for the Go type.
     * @param {string} type
     * @param {string} value
     */
    function parseArg(type, value) {
        if (type === 'string') return value;
        if (isNumeric(type)) return Number(value);
        if (type === 'bool') return value === 'true';
        return parseLoose(value);
    }

    function submit() {
        if (argumentTypes.some((type, i) => (values[i].trim() === '' && type !== 'string') || isInvalid(type, values[i]))) {
            return;
        }
        run(argumentTypes.map((type, i) => parseArg(type, values[i])));
    }
</script>

{#if methodName}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div
            class="absolute inset-0 z-20 flex items-start justify-start p-2"
            aria-roledescription="cancel"
    >
        <div
                class="rounded-md border border-white/10 bg-slate-950/95 p-2"
        >
            <div class="mb-1 flex items-center justify-between">
                <div class="text-[12px] font-semibold opacity-90">Run: <code>{methodName}({argumentTypes.join(', ')})</code></div>
                <button
                        type="button"
                        class="inline-flex h-6 w-6 items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        aria-label="Close run popup"
                        title="Close"
                        on:click={cancel}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            {#if argumentTypes.length > 0}
                <div class="mt-3 space-y-2">
                    {#each argumentTypes as type, i}
                        <div class="flex flex-row items-center justify-between gap-2">
                            <label class="text-xs text-white/80" for={"arg-" + i}>
                                arg{i} <span class="opacity-60">{type}</span>
                            </label>

                            <input
                                id={"arg-" + i}
                                class="min-w-0 rounded border bg-white/5 px-2 py-1 text-xs text-white outline-none placeholder:text-white/35 focus:border-white/30 {isInvalid(type, values[i] ?? '') ? 'border-red-500' : 'border-white/10'}"
                                bind:value={values[i]}
                                placeholder={type === 'bool' ? 'true / false' : 'Enter ' + type}
                            />
                        </div>
                    {/each}
                </div>
            {/if}

            <div class="mt-2 flex items-center justify-end gap-2">
                <button
                        type="button"
                        class="rounded px-2 py-1 text-[12px] text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        on:click={cancel}
                >
                    Cancel
                </button>
                <button
                        type="button"
                        class="rounded bg-white/10 px-2 py-1 text-[12px] text-white hover:bg-white/15 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        on:click={submit}
                >
                    Run
                </button>
            </div>
        </div>
    </div>
{/if}
