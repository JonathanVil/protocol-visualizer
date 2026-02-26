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

    let ignoreNextUpdate = false;

    onMount(async () => {
        // dynamic import only runs in the browser
        const monaco = await import('monaco-editor');

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
</script>

<div bind:this={editorDiv} class="w-full h-full border border-gray-300 rounded-md"></div>
