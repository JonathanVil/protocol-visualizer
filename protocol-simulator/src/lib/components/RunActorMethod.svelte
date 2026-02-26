<script>

    /** @type {string} */
    export let methodName;

    /** @type {string[]} */
    export let argumentNames = [];

    /** @type {(args: string[]) => void} */
    export let submit;

    /** @type {() => void} */
    export let cancel;

    /** @type {any[]} */
    let args = [];

    function run() {
        if (args.length !== argumentNames.length) {
            return;
        }

        submit(args);
    }
</script>

{#if methodName}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div
            class="absolute inset-0 z-20 flex items-start justify-start p-2"
            aria-roledescription="cancel"
    >
        <div
                class="w-[320px] rounded-md border border-white/10 bg-slate-950/95 p-2 shadow-[0_12px_30px_rgba(0,0,0,0.55)]"
        >
            <div class="mb-1 flex items-center justify-between">
                <div class="text-[12px] font-semibold opacity-90">Run: {methodName}</div>
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



            <div class="mt-2 flex items-center justify-end gap-2">
                <div class="mr-auto text-[11px] opacity-70">
                    <span>Esc</span> to cancel, <span>Ctrl/⌘+Enter</span> to save
                </div>

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
                        on:click={run}
                >
                    Save
                </button>
            </div>
        </div>
    </div>
{/if}