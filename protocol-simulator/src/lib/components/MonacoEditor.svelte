<script>
    import {onMount} from "svelte";

    //sourcecode passed between parent and child
    export let sourceCode = "";

    /*!!!                MONACO EDITOR                  !!!*/
    /** @type {typeof import('monaco-editor')} */
    let monaco;

    //JSDOC comment for type
    /** @type {HTMLElement | null} */
    let editorDiv = null;

    /** @type {import('monaco-editor').editor.IStandaloneCodeEditor} */
    let editorInstance;

    let ignoreNextUpdate = false;


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
        // dynamic import only runs in the browser
        monaco = await import('monaco-editor');
        file = await fetch("built-in-functions.txt");
        builtInFunctions = await file.text();

        if (!editorDiv) return;

        editorInstance = monaco.editor.create(editorDiv, {
            value: sourceCode,
            language: "javascript",
            theme: "vs-dark",
            automaticLayout: true,
            minimap: { enabled: false },
            fontSize: 14,
        });

        editorInstance.onDidChangeModelContent(() => {
            ignoreNextUpdate = true;
            sourceCode = editorInstance.getValue();
        });
    });

    // Reactively update the editor when sourceCode changes from the parent
    $: if (
        editorInstance &&
        !ignoreNextUpdate &&
        editorInstance.getValue() !== sourceCode
    ) {
        editorInstance.setValue(sourceCode);
    }
    // Reset the ignore flag after the update cycle
    $: if (ignoreNextUpdate) {
        ignoreNextUpdate = false;
    }

    $: if (builtInFunctionsEditorDiv && showBuiltInFunctions) {
        if (monaco) {
            builtInEditorInstance = monaco.editor.create(builtInFunctionsEditorDiv, {
                value: builtInFunctions,
                language: "javascript",
                theme: "vs-dark",
                automaticLayout: true,
                minimap: { enabled: false },
                fontSize: 14,
            });
        }
    }


</script>

<div bind:this={editorDiv} class="w-full h-full border border-gray-300 rounded-md"></div>

<button
        class="absolute top-2 right-2 flex items-center gap-1 rounded-lg bg-white hover:bg-gray-200 focus:outline-none"
        aria-label="Built-in functions" title="See built-in functions"
        on:click={() => showBuiltInFunctions = !showBuiltInFunctions}>
    <img src="icon_func.png" class="w-8" alt="Italian Trulli">
</button>

{#if showBuiltInFunctions}
    awd
    <div bind:this={builtInFunctionsEditorDiv} class="absolute top-10 right-2 w-1/2 h-4/5 border-2 rounded-md "></div>
{/if}



