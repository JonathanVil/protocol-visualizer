<script>
    import {onMount} from "svelte";

    //sourcecode passed between parent and child
    export let sourceCode = "";

    /*!!!                MONACO EDITOR                  !!!*/

    //JSDOC comment for type
    /** @type {HTMLElement | null} */
    let editorDiv = null;

    /** @type {import('monaco-editor').editor.IStandaloneCodeEditor} */
    let editorInstance;


    onMount(async () => {
        // dynamic import only runs in the browser
        const monaco = await import('monaco-editor');

        if (!editorDiv) return;

        editorInstance = monaco.editor.create(editorDiv, {
            value: sourceCode || "// Write your code here...",
            language: "javascript",
            theme: "vs-dark",
            automaticLayout: true,
            minimap: { enabled: false },
            fontSize: 14,
        });

        editorInstance.onDidChangeModelContent(() => {
            sourceCode = editorInstance.getValue();
        });
    });
</script>

<div bind:this={editorDiv} class="w-full h-80 border border-gray-300 rounded-md"></div>


