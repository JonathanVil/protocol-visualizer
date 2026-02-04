<script lang="js">
    let sourceCode = "";

    function compile() {
      const code = document.getElementById("code").value
      try {
        const result = new Function(code)()

        // Enforce contract
        if (typeof result !== "function") {
          throw new Error("Code must return a class")
        }
        if (!result.prototype.receive) {
          throw new Error("Actor must implement receive(msg)")
        }

        ActorClass = result
        alert("Compiled successfully!")
      } catch (e) {
        alert("Compile error: " + e.message)
      }
    }
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://svelte.dev/docs/kit">svelte.dev/docs/kit</a> to read the documentation</p>

<textarea
  bind:value={sourceCode}
  placeholder="Write your protocol here..."
></textarea>

<button on:click={compile}>
  Compile
</button>
