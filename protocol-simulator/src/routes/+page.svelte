<h1 class="text-5xl">Protocol Simulator</h1>


<div class="flex flex-col gap-10">
    <div>
        <div>
            <h3 class="font-bold text-1xl">Output</h3>
        </div>
        <div class="container-fluid border-2 h-100 border-gray-500 ">

        </div>
    </div>


    <div>
        <div >
            <h3 class="font-bold text-1xl">Actor code</h3>
        </div>

        <!--Monaco Editor-->
        <div bind:this={editorDiv} class="w-full h-120 border border-gray-300 rounded-md"></div>

        <button on:click={compile}>
            Compile
        </button>
    </div>
</div>




<script lang="js">
    import {onMount} from "svelte";
    import "./layout.css"; // ensure this path matches your project structure

    let sourceCode = "";
    let ActorClass = null;


    /*!!!                MONACO EDITOR                  !!!*/
    //JSDOC comment for type
    /** @type {HTMLElement} */
    let editorDiv;

    /** @type {import('monaco-editor').editor.IStandaloneCodeEditor} */
    let editorInstance;
    onMount(async () => {
        // dynamic import only runs in the browser
        const monaco = await import('monaco-editor');

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


    /*!!!             COMPILING ACTOR                   !!!*/
    function compile() {
        const code = sourceCode;
        try {
            const result = new Function(code)();
            alert(code)

            // Enforce contract
            if (typeof result !== "function") {
                throw new Error("Code must return a class");
            }
            if (!result.prototype.receive) {
                throw new Error("Actor must implement receive(msg)");
            }

            ActorClass = result;
            alert("Compiled successfully!");
        } catch (e) {
            const msg = e instanceof Error ? e.message : String(e);
            alert("Compile error: " + msg);
        }
    }
</script>





