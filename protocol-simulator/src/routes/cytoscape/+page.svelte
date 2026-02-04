<script>
    import MonacoEditer from "$lib/MonacoEditer.svelte";
    import  Graph  from "$lib/Graph.svelte";
    import { parseProtocolCode } from '$lib/protocolUtils.js';

    let sourceCode = "";

    /** @type {{ id: string, label: string }[]} */
    let actors = [];
    /** @type {{ source: string, target: string, label: string }[]} */
    let messages = [];

    function compile() {
        const result = parseProtocolCode(sourceCode);


        //note: svelte automatically updates them in the Graph.svelte!
        actors = result.actors;
        messages = result.messages;
        //alert(actors.pop().label);

    }
</script>

<h1 class="text-5xl font-bold mb-6">Protocol Simulator</h1>

<!--Connect to the MonacoEditor and gets the written sourceCode-->
<MonacoEditer bind:sourceCode={sourceCode} />

<button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        on:click={compile}>
    Compile
</button>

<Graph nodes={actors} edges={messages} />

