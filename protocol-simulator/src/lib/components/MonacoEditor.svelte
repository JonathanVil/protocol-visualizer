<script>
    import {onMount} from "svelte";
    import EditorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
    import TsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

    /** @typedef {import('$lib/types.js').EditorTab} EditorTab */

    /** @type {EditorTab[]} */
    let tabs = [];

    /** @type {EditorTab | null} */
    export let selectedTab = null;

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

        tabs = [
            {
                id: crypto.randomUUID(),
                name: "Protocol 1",
                model: monaco.editor.createModel("// code here", "javascript"),
            },
            {
                id: crypto.randomUUID(),
                name: "Protocol 2",
                model: monaco.editor.createModel("// other code", "javascript")
            }
        ]

        setTab(tabs[0]);
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
     * @param {string} name
     * @param {string | null} code
     */
    export function openTab(name, code) {
        tabs = [...tabs, {
            id: crypto.randomUUID(),
            name,
            model: monaco.editor.createModel(code ?? "", "javascript")
        }]
    }
</script>

<div class="flex-row gap-2">
    {#each tabs as tab}
        <button class="border border-black" on:click={() => setTab(tab)}>
            {tab.name}
        </button>
    {/each}
</div>

<div bind:this={editorDiv} class="w-full h-full border border-gray-300 rounded-md"></div>
