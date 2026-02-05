<script>
    import MonacoEditer from "$lib/MonacoEditer.svelte";
    import  Graph  from "$lib/Graph.svelte";
    import { parseProtocolCode } from '$lib/protocolUtils.js';
    import {LinkedList} from '$lib/LinkedList.js';
    import ManualMessageComponent from "$lib/ManualMessageComponent.svelte";
    /** @typedef {import('$lib/types.js').Message} Message */
    /** @typedef {import('$lib/types.js').ActorConstructor} ActorConstructor */

    let sourceCode = "";

    /** @type {{ id: string}[]} */
    let actors = [];

    let messages = new LinkedList();
    let id = 0;



    function spawnActor() {
        /** @type {ActorConstructor} */
        const actorClass = parseProtocolCode(sourceCode);

        //note: svelte automatically updates them in the Graph.svelte!
        let actor = new actorClass(id++);
        actors = [...actors, actor];
        console.log("Adding actor");
        console.log(actors);

    }

    function startSimulation() {
        setInterval(spawnActor, 100);
    }





</script>

<h1 class="text-5xl font-bold mb-6">Protocol Simulator</h1>

<!--Connect to the MonacoEditor and gets the written sourceCode-->
<MonacoEditer bind:sourceCode={sourceCode} />

<button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        on:click={spawnActor}>
    Spawn actor
</button>

<button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        on:click={startSimulation}>
    Start Simulator
</button>

<ManualMessageComponent messages={messages} />

<Graph nodes={actors}/>

