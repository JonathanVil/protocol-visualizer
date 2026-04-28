<script>
    /** @type {string} */
    export let methodName;

    /** @type {string[]} */
    export let argumentNames = [];

    /** @type {(args: any[]) => void} */
    export let run;

    /** @type {() => void} */
    export let cancel;

    /** @type {{ name: string, value: string, type: 'String' | 'Number' }[]} */
    let args = [];

    $: args = argumentNames.map((name, index) => {
        const existing = args[index];
        return {
            name: existing?.name ?? name,
            value: existing?.value ?? '',
            type: existing?.type ?? 'String'
        };
    });

    /**
     * @param {{value: string, type: ("String"|"Number")}} arg
     */
    function isInvalidNumberArg(arg) {
        return arg?.type === 'Number' && arg.value.trim() !== '' && Number.isNaN(Number(arg.value));
    }

    function submit() {
        if (args.length !== argumentNames.length) {
            return;
        }

        if (args.some(isInvalidNumberArg)) {
            return;
        }

        const parsedArgs = args.map((arg) =>
            arg.type === 'Number' ? Number(arg.value) : arg.value
        );

        run(parsedArgs);
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
                <div class="text-[12px] font-semibold opacity-90">Run: <code>{methodName}({argumentNames.join(', ')})</code></div>
                <button
                        type="button"
                        class="inline-flex h-6 w-6 items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        aria-label="Close edit popup"
                        title="Close"
                        on:click={cancel}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            {#if args.length > 0}
                <div class="mt-3 space-y-2">
                    {#each args as arg}
                        <div class="flex flex-row items-center justify-between gap-2">
                            <label class="text-xs text-white/80" for={"arg-" + arg.name}>
                                {arg.name}
                            </label>

                            <input
                                id={"arg-" + arg.name}
                                class="min-w-0 rounded border bg-white/5 px-2 py-1 text-xs text-white outline-none placeholder:text-white/35 focus:border-white/30 {isInvalidNumberArg(arg) ? 'border-red-500' : 'border-white/10'}"
                                bind:value={arg.value}
                                placeholder={"Enter " + arg.name}
                            />

                            <select
                                id={"arg-type-" + arg.name}
                                class="rounded border border-white/10 bg-white/5 px-2 py-1 text-xs text-white outline-none focus:border-white/30"
                                bind:value={arg.type}
                            >
                                <option value="String" class="text-black">String</option>
                                <option value="Number" class="text-black">Number</option>
                            </select>
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