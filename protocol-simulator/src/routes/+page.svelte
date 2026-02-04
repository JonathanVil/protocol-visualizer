<h1>Protocol Simulator</h1>

<textarea
        bind:value={sourceCode}
        placeholder="Write your protocol here..."
></textarea>


<div class="flex-1">
    <h3>Actor code</h3>
</div>


<button on:click={compile}>
    Compile
</button>




<script lang="js">
    import "./layout.css"; // ensure this path matches your project structure
    let sourceCode = "";
    let ActorClass = null;

    function compile() {
        const code = sourceCode;
        try {
            const result = new Function(code)();

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





