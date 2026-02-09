<script>
    import MonacoEditer from "$lib/MonacoEditer.svelte";
    import Graph  from "$lib/Graph.svelte";
    import {LinkedList} from '$lib/LinkedList.js';
    import ManualMessageComponent from "$lib/ManualMessageComponent.svelte";
    import {transitTime, getNextMessageId, parseProtocolCode, getStepSize, setStepSize} from "$lib/protocolUtils.js";

    /** @typedef {import('$lib/types.js').Message} Message */
    /** @typedef {import('$lib/types.js').ActorConstructor} ActorConstructor */
    /** @typedef {import('$lib/types.js').Actor} Actor */

    let sourceCode = "";

    //reference to graph instance
    /** @type {import('$lib/Graph.svelte').default} */
    let graphRef;

    /** @type {Actor[]} */
    let actors = [];

    let messages = new LinkedList();
    let id = 0;


    /** @type number */
    let intervalId;

    let stepSizeInput = getStepSize();
    let stepSizeUpdated = false;
    let paused = true;

    function spawnActor() {
        /** @type {ActorConstructor} */
        const actorClass = parseProtocolCode(sourceCode, send, getActors); // we need to give send here so the actor "knows" it

        //  svelte automatically updates them in the Graph.svelte
        /** @type {Actor} */
        let actor = new actorClass(id++);
        actors = [...actors, actor];
        console.log("Adding actor");
        console.log(actors);

    }

    /** @param {Message} message */
    function deliverMessage(message) {
        let actor = actors[message.destination];
        let msg = {type: message.type, from: message.source};
        actor.receive(msg)
    }

    function startSimulation() {
        if (stepSizeUpdated) {
            console.log("Updating simulation step size");
        } else {
            console.log("Starting simulation");
        }

        paused = false;
        clearInterval(intervalId);
        stepSizeUpdated = false;
        intervalId = setInterval(step, getStepSize());
    }

    function pauseSimulation() {
        console.log("Pausing simulation");
        paused = true;
    }

    function setStepSizeInput() {
        setStepSize(stepSizeInput);
        stepSizeUpdated = true;
    }

    /** @param {number} from
     *  @param {number} to
     *  @param {any} data
     *  @param {string} type
     * */
    function send(from, to, type, data ) {
        console.log(from, "send to", to);
        messages.append({id: getNextMessageId(), source: from, destination: to, type: type, transitSteps: transitTime, elapsedSteps: 0, data: data})
    }

    function getActors() {
        return actors.length;
    }

    function step() {
        if (paused) {
            return
        }
        let n = messages.length;
        for (let i = 0; i < n; i++) {
            let message = messages.pop()
            if (message != null){
                message.elapsedSteps++
                //Animate messages
                graphRef.animateNewMessage(message);

                if (message.elapsedSteps === message.transitSteps){
                    deliverMessage(message)
                } else {
                    messages.append(message)
                }
            }
        }
        if (stepSizeUpdated) { // We need to reboot the simulation loop in order to update stepsize
            paused = true;
            startSimulation();
        }
    }



</script>

<h1 class="text-5xl font-bold mb-6">Protocol Simulator</h1>

<!--Connect to the MonacoEditor and gets the written sourceCode-->
<MonacoEditer bind:sourceCode={sourceCode} />

<div class="mt-4 flex-row space-x-2 pt-2">
    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            on:click={spawnActor}>
        Spawn actor
    </button>

    {#if paused}
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
                on:click={startSimulation}>
            <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#1f1f1f"><path d="M320-200v-560l440 280-440 280Z"/></svg>
            <span>Start Simulator</span>
        </button>
    {:else}
        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 inline-flex items-center"
                on:click={pauseSimulation}>
            <svg class="fill-current w-4 h-4 mr-2" xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#1f1f1f"><path d="M560-200v-560h160v560H560Zm-320 0v-560h160v560H240Z"/></svg>
            <span>Pause Simulator</span>
        </button>
    {/if}

    <input
        type="number"
        bind:value={stepSizeInput}
        placeholder="100"
    />

    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            on:click={setStepSizeInput}>
        Set step size (ms)
    </button>
</div>

<ManualMessageComponent messages={messages} />

<Graph bind:this={graphRef} nodes={actors}/>
