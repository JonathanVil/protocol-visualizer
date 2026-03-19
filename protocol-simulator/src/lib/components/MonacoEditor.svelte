<script>
    import {onMount} from "svelte";
    import EditorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
    import TsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

    /** @typedef {import('$lib/types.js').EditorTab} EditorTab */

    /** @type {EditorTab[]} */
    export let tabs = [];

    /** @type {EditorTab | null} */
    export let selectedTab = null;

    /** @type {{ name: string; content: string }[] } */
    export let predefinedProtocols;

    /** @type {(name: string|null) => void} */
    export let spawnActor;

    let defaultProtocol = predefinedProtocols.find(/** @type {{ name: string; content: string }} */p => p.name === "Default");

    /** @type {string | null} */
    let editingTabId = null;

    /** @type {HTMLElement | null} */
    let editorDiv = null;

    /** @type {import('monaco-editor').editor.IStandaloneCodeEditor} */
    let editorInstance;

    /** @type {typeof import('monaco-editor')} */
    let monaco;

    let showLoadExistingMenu = false;

    /** @type {HTMLDivElement | null} */
    let loadExistingMenuWrapper = null;

    const newTabName = "Ping";


    /*!!!                Built-in functions                 !!!*/
    /** @type {boolean} */
    let showBuiltInFunctions = false;
    /** @type {HTMLElement | null} */
    let builtInFunctionsEditorDiv = null;
    /** @type {import('monaco-editor').editor.IStandaloneCodeEditor} */
    let builtInEditorInstance;

    let file;
    /** @type {String} */
    let builtInFunctions = ""



    onMount(async () => {
        self.MonacoEnvironment = {
            getWorker(_, label) {
                if (label === "typescript" || label === "javascript") {
                    return new TsWorker();
                }

                return new EditorWorker();
            }
        };

        /** @param {Event} event */
        const handleDocumentClick = (event) => {
            if (!showLoadExistingMenu) return;
            if (event.target instanceof Node && loadExistingMenuWrapper?.contains(event.target)) return;

            showLoadExistingMenu = false;
        };

        document.addEventListener("click", handleDocumentClick);

        monaco = await import("monaco-editor");
        // dynamic import only runs in the browser
        file = await fetch("BASICS.md");
        builtInFunctions = await file.text();

        if (!editorDiv) return;

        editorInstance = monaco.editor.create(editorDiv, {
            language: "javascript",
            theme: "vs-dark",
            automaticLayout: true,
            minimap: {enabled: false},
            fontSize: 14,
            model: selectedTab?.model ?? null
        });

        if (tabs.length === 0) {
            openNewTab(null, null);
        }
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
            const hasPlainNewTab = tabs.some((t) => t.name === newTabName);

            const maxNewTabNumber = Math.max(
                1,
                ...tabs.map(t => {
                    const newTabPattern = new RegExp(`^${newTabName} (\\d+)$`);
                    const match = t.name.match(newTabPattern);
                    return match ? Number.parseInt(match[1], 10) : 0;
                })
            );

            name = hasPlainNewTab
                ? `${newTabName} ${maxNewTabNumber + 1}`
                : newTabName;
        }

        let newTab = {
            id: crypto.randomUUID(),
            name,
            model: monaco.editor.createModel(code ?? defaultProtocol?.content ?? "", "javascript")
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

    $: if (builtInFunctionsEditorDiv && showBuiltInFunctions) {
        if (monaco) {
            builtInEditorInstance = monaco.editor.create(builtInFunctionsEditorDiv, {
                value: builtInFunctions,
                language: "markdown",
                theme: "vs-dark",
                automaticLayout: true,
                minimap: { enabled: false },
                fontSize: 14,
                readOnly: true
            });
        }
    }


</script>

<div class="flex h-full flex-col">
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
                                on:change={() => selectedTab = selectedTab}
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
                                title={`Double-click to rename ${tab.name}. Shift+click to close.`}
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
        <div class="relative flex shrink-0 items-center gap-2" bind:this={loadExistingMenuWrapper}>
            <button
                    type="button"
                    class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-gray-400 hover:bg-gray-50"
                    aria-label="Open tab"
                    title="Open tab"
                    on:click={() => openNewTab(null, null)}
            >
                + Open
            </button>
            <button
                    type="button"
                    class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-gray-400 hover:bg-gray-50"
                    aria-label="Load existing protocol"
                    title="Load existing protocol"
                    aria-expanded={showLoadExistingMenu}
                    on:click={() => showLoadExistingMenu = !showLoadExistingMenu}
            >
                Load existing
            </button>
            {#if showLoadExistingMenu}
                <div class="absolute right-0 top-full z-10 mt-2 w-52 rounded-md border border-gray-300 bg-white py-1 shadow-lg">
                    <div class="px-3 py-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
                        Select protocol
                    </div>
                    {#each predefinedProtocols.filter(p => p.name !== "Default") as protocol}
                        <button
                                type="button"
                                class="block w-full px-3 py-2 text-left text-sm text-gray-700 transition hover:bg-gray-100"
                                on:click={() => {
                                openNewTab(protocol.name, protocol.content);
                                showLoadExistingMenu = false;
                            }}
                        >
                            {protocol.name}
                        </button>
                    {/each}
                </div>
            {/if}
            <button
                    class="rounded-md border border-gray-300 bg-white px-1 text-sm font-medium text-gray-700 transition hover:border-gray-400 hover:bg-gray-50"
                    aria-label="Built-in functions" title="See built-in functions"
                    on:click={() => showBuiltInFunctions = !showBuiltInFunctions}>
                <img src="icon_func.png" class="w-8" alt="Italian Trulli">
            </button>
        </div>
    </div>

    <div bind:this={editorDiv} class="min-h-0 flex-1 border border-gray-300"></div>

    <div class="flex items-center justify-between gap-3 rounded-b-md border border-b-0 border-gray-300 bg-gray-100 px-3 py-2">
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center text-sm font-medium" on:click={() => spawnActor(selectedTab?.name ?? null)}>
            <i class="fa fa-plus-circle mr-2" aria-hidden="true"></i>
            <span>Spawn '{selectedTab?.name ?? "new"}' actor</span>
        </button>
    </div>

</div>

{#if showBuiltInFunctions}
    <div bind:this={builtInFunctionsEditorDiv} class="absolute top-15 right-2 w-2/3 h-4/5 border-2 rounded-md "></div>
{/if}



