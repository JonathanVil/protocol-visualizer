<script>
    import {onMount} from "svelte";
    import EditorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
    import TsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

    /** @typedef {import('$lib/types.js').EditorTab} EditorTab */

    /** @type {EditorTab[]} */
    export let tabs = [];

    /** @type {EditorTab | null} */
    export let selectedTab = null;

    /** @type {string | null} */
    let editingTabId = null;

    /** @type {HTMLElement | null} */
    let editorDiv = null;

    /** @type {import('monaco-editor').editor.IStandaloneCodeEditor} */
    let editorInstance;

    /** @type {typeof import('monaco-editor')} */
    let monaco;

    onMount(async () => {
        self.MonacoEnvironment = {
            getWorker(_, label) {
                if (label === "typescript" || label === "javascript") {
                    return new TsWorker();
                }

                return new EditorWorker();
            }
        };


        monaco = await import("monaco-editor");

        if (!editorDiv) return;

        editorInstance = monaco.editor.create(editorDiv, {
            language: "javascript",
            theme: "vs-dark",
            automaticLayout: true,
            minimap: { enabled: false },
            fontSize: 14,
            model: selectedTab?.model ?? null
        });
    });

    /** @param {EditorTab} tab */
    function setTab(tab) {
        if (selectedTab?.id === tab.id) return;

        if (selectedTab) {
            selectedTab.viewState = editorInstance.saveViewState();
        }

        selectedTab = tab;
        editorInstance.setModel(tab.model);
        editorInstance.restoreViewState(tab.viewState ?? null);
        editorInstance.focus();
    }

    /**
     * @param {string | null | undefined} name
     * @param {string | null | undefined} code
     */
    export function openNewTab(name, code) {
        if (!name) {
            const hasPlainNewTab = tabs.some((t) => t.name === "New tab");

            const maxNewTabNumber = Math.max(
                1,
                ...tabs.map(t => {
                    const match = t.name.match(/^New tab (\d+)$/);
                    return match ? Number.parseInt(match[1], 10) : 0;
                })
            );

            name = hasPlainNewTab
                ? `New tab ${maxNewTabNumber + 1}`
                : "New tab";
        }

        let newTab = {
            id: crypto.randomUUID(),
            name,
            model: monaco.editor.createModel(code ?? "", "javascript")
        }
        tabs = [...tabs, newTab]

        setTab(newTab);
        //editingTabId = newTab.id;
    }

    /** @param {EditorTab} tab */
    export function closeTab(tab) {
        if (tabs.length === 1) return;

        let i = tabs.indexOf(tab);
        tabs = tabs.filter(t => t.id !== tab.id);

        if (selectedTab?.id === tab.id) {
            setTab(tabs[i - 1] ?? tabs[0] ?? null);
        }
    }
</script>

<div class="flex items-center justify-between gap-3 rounded-t-md border border-b-0 border-gray-300 bg-gray-100 px-3 py-2">
    <div class="flex min-w-0 flex-1 items-center gap-2 overflow-x-auto">
        {#each tabs as tab}
            <div
                class={`flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm transition ${
                    selectedTab?.id === tab.id
                        ? "border-blue-500 bg-white text-blue-700 shadow-sm"
                        : "border-gray-300 bg-gray-50 text-gray-700 hover:bg-white"
                }`}
            >

                {#if editingTabId === tab.id}
                    <input
                        class="bg-transparent text-left font-medium outline-none placeholder:text-gray-400"
                        bind:value={tab.name}
                        aria-label="Tab name"
                        size={Math.max(tab.name.length, 1)}
                        autofocus
                        on:blur={() => editingTabId = null}
                        on:keydown={(event) => {
                            if (event.key === "Enter" || event.key === "Escape") {
                                editingTabId = null;
                            }
                        }}
                    />
                {:else}
                    <button
                        type="button"
                        class="truncate text-left font-medium flex items-center gap-2"
                        on:click={e => e.shiftKey ? closeTab(tab) : setTab(tab)}
                        on:dblclick={() => editingTabId = tab.id}
                    >
                        <span class="h-2 w-2 rounded-full bg-current opacity-70"></span>
                        {tab.name}
                    </button>
                {/if}

                <button
                    type="button"
                    class="flex h-5 w-5 items-center justify-center rounded text-gray-400 hover:bg-gray-200 hover:text-gray-700"
                    aria-label={`Close ${tab.name}`}
                    title={`Close ${tab.name}`}
                    on:click={() => closeTab(tab)}
                >
                    ×
                </button>
            </div>
        {/each}
    </div>

    <button
        type="button"
        class="shrink-0 rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-gray-400 hover:bg-gray-50"
        aria-label="Open tab"
        title="Open tab"
        on:click={() => openNewTab(null, null)}
    >
        + Open
    </button>
</div>

<div bind:this={editorDiv} class="h-full w-full rounded-b-md border border-gray-300"></div>
