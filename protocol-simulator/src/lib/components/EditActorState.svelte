<script>
    // --- inline edit popup state ---
    import ActorStatePopper from "$lib/components/ActorStatePopper.svelte";

    /** @type {string | null} */
    export let editingKey = null;

    /** @type {string} */
    export let editText = '';

    /** @type {any} */
    export let editOriginalValue;

    /**
     * @param {string} newValue
     */
    export let save = (newValue) => {}

    /**
     * @param {event?} event
     */
    function closeEdit(event) {
        event?.stopPropagation?.();
        editingKey = null;
        editText = '';
        editOriginalValue = undefined;
    }

    /** @param {string} text @param {any} original
     * @param original
     */
    function parseEditedValue(text, original) {
        const t = text.trim();

        if (t === 'undefined') return undefined;
        if (t === 'null') return null;

        // Preserve strings as strings (no JSON parsing surprises).
        if (typeof original === 'string') return text;

        // Booleans/numbers/objects: try JSON.parse first.
        try {
            return JSON.parse(t);
        } catch {
            // Fallback: if original was a number, allow plain numeric input.
            if (typeof original === 'number') {
                const n = Number(t);
                if (!Number.isNaN(n)) return n;
            }
            // Otherwise: keep as text (last resort).
            return text;
        }
    }

    /**
     * @param {event} event
     */
    function saveEdit(event) {
        event?.stopPropagation?.();
        if (!editingKey) return;

        const nextValue = parseEditedValue(editText, editOriginalValue);
        console.log(nextValue);

        save(nextValue);

        closeEdit(null);
    }

    /**
     * @param {KeyboardEvent} event
     */
    function onEditKeydown(event) {
        if (event.key === 'Escape') closeEdit(event);
        if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') saveEdit(event);
    }

    /** @param {HTMLElement} el */
    function init(el){
        el.focus()
    }
</script>

{#if editingKey}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div
            class="absolute inset-0 z-20 flex items-start justify-start p-2"
            aria-roledescription="cancel"
    >
        <div
                class="w-[320px] rounded-md border border-white/10 bg-slate-950/95 p-2 shadow-[0_12px_30px_rgba(0,0,0,0.55)]"
        >
            <div class="mb-1 flex items-center justify-between">
                <div class="text-[12px] font-semibold opacity-90">Edit: {editingKey}</div>
                <button
                        type="button"
                        class="inline-flex h-6 w-6 items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        aria-label="Close edit popup"
                        title="Close"
                        on:click={closeEdit}
                >
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>


            <textarea use:init
                    class="w-full resize-y rounded bg-white/5 px-2 py-1 font-mono text-[12px] leading-[1.2] text-white outline-none ring-1 ring-white/10 focus:ring-2 focus:ring-white/30"
                    rows="4"
                    bind:value={editText}
                    on:keydown={onEditKeydown}></textarea>

            <div class="mt-2 flex items-center justify-end gap-2">
                <div class="mr-auto text-[11px] opacity-70">
                    <span>Esc</span> to cancel, <span>Ctrl/⌘+Enter</span> to save
                </div>

                <button
                        type="button"
                        class="rounded px-2 py-1 text-[12px] text-white/80 hover:text-white hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        on:click={closeEdit}
                >
                    Cancel
                </button>
                <button
                        type="button"
                        class="rounded bg-white/10 px-2 py-1 text-[12px] text-white hover:bg-white/15 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/30"
                        on:click={saveEdit}
                >
                    Save
                </button>
            </div>
        </div>
    </div>
{/if}
