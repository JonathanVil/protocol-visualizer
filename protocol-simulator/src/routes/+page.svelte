<script>
    import MonacoEditer from "$lib/MonacoEditer.svelte";
    import Graph  from "$lib/Graph.svelte";
    import { parseProtocolCode } from '$lib/protocolUtils.js';
    import {LinkedList} from '$lib/LinkedList.js';
    import ManualMessageComponent from "$lib/ManualMessageComponent.svelte";
    import {transitTime} from "$lib/protocolUtils.js";
    import { getStepSize, setStepSize } from "$lib/protocolUtils.js";
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
    let paused = false;

    function spawnActor() {
        /** @type {ActorConstructor} */
        const actorClass = parseProtocolCode(sourceCode, send); // we need to give send here so the actor "knows" it

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
     *  @param {string} type
     * */
    function send(from, to, type) {
        messages.append({id: Date.now(), source: from, destination: to, type: type, transitSteps: transitTime, elapsedSteps: 0})
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

<button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        on:click={spawnActor}>
    Spawn actor
</button>

<button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        on:click={startSimulation}>
    Start Simulator
</button>

<button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        on:click={pauseSimulation}>
    Pause Simulator
</button>


<input
        type="number"
        bind:value={stepSizeInput}
        placeholder="100"
/>

<button class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        on:click={setStepSizeInput}>
    Set step size (ms)
</button>

<ManualMessageComponent messages={messages} />

<Graph bind:this={graphRef} nodes={actors}/>

