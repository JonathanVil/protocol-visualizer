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
        <textarea
                class="w-full border border-gray-300 rounded-md text-sm "
                bind:value={sourceCode}
                placeholder="Write your protocol here..."
        ></textarea>

        <button on:click={compile}>
            Compile
        </button>
    </div>
</div>







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





