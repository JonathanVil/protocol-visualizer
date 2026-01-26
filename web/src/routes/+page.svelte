<script lang="ts">
    let sourceCode = "";

    function compile() {
      console.log("Compiling:", sourceCode);
      alert("Compiling:\n" + sourceCode);
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
