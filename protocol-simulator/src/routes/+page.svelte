<script>
    import MonacoEditer from "$lib/MonacoEditer.svelte";
    import Graph  from "$lib/Graph.svelte";
    import {Queue} from '$lib/Queue.js';
    import ManualMessageComponent from "$lib/ManualMessageComponent.svelte";
    import {getTransitTime, setLowerTransitTime, setUpperTransitTime, getNextMessageId, parseProtocolCode, getStepSize, setStepSize} from "$lib/protocolUtils.js";

    /** @typedef {import('$lib/types.js').Message} Message */
    /** @typedef {import('$lib/types.js').ActorConstructor} ActorConstructor */
    /** @typedef {import('$lib/types.js').Actor} Actor */

    let sourceCode = "";

    //reference to graph instance
    /** @type {import('$lib/Graph.svelte').default} */
    let graphRef;

    /** @type {Actor[]} */
    let actors = [];

    let messages = new Queue();
    let timeouts = new Queue();
    let id = 0;


    /** @type number */
    let intervalId;

    let stepSizeInput = getStepSize();
    let stepSizeUpdated = false;
    let paused = true;

    function spawnActor() {
        /** @type {ActorConstructor} */
        const actorClass = parseProtocolCode(sourceCode, send, getActors, createQueue, timeout); // we need to give send here so the actor "knows" it

        if (actorClass == null) {
          console.error("Actor class not defined");
          return;
        }

        //  svelte automatically updates them in the Graph.svelte
        /** @type {Actor} */
        let actor = new actorClass(id++);
        actors = [...actors, actor];
        console.log("Adding actor");
        console.log(actors);

    }


    /**
     * Deliver message to destination actor. Transform message to lightweight msg. Lastly invoke actors 'receive' method
     * @param {Message} message
     */
    function deliverMessage(message) {
        let actor = actors[message.destination];
        let msg = {type: message.type, from: message.source, data: message.data};
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

    function resetSimulation() {
        clearInterval(intervalId);
        messages = new Queue();
        timeouts = new Queue();
        actors = [];
        id = 0;
        stepSizeUpdated = false;
        paused = false;
        graphRef.resetGraph();
    }



    function step() {
        if (paused) {
            return
        }
        //handle messages
        messageStep()

        //handle timeouts
        timeoutStep()

        //handle updating stepsize
        if (stepSizeUpdated) { // We need to reboot the simulation loop in order to update stepsize
            paused = true;
            startSimulation();
        }
    }

    function messageStep() {
        let n = messages.length;
        for (let i = 0; i < n; i++) {
            let message = messages.pop()
            if (message != null){
                message.elapsedSteps++
                //Animate messages
                graphRef.animateMessage(message);

                if (message.elapsedSteps === message.transitSteps){
                    deliverMessage(message)
                } else {
                    messages.push(message)
                }
            }
        }
    }

    function timeoutStep() {
        let n = timeouts.length;
        for (let i = 0; i < n; i++) {
            let timer = timeouts.pop()
            if (timer != null){
                if (timer.steps === 0){
                    timer.reaction()
                } else {
                    timer.steps -= 1
                    timeouts.push(timer);
                }

            }
        }
    }


    // These are the functions we export into the Actors
    // TODO: put these somewhere nice :)

    /** @param {number} from
     *  @param {number} to
     *  @param {any} data
     *  @param {string} type
     * */
    function send(from, to, type, data) { //Example of use: send(this.id, from.id, "PING", "Hello")
        console.log(from, "send to", to);
        let transitTime = getTransitTime();
        messages.push({id: getNextMessageId(), source: from, destination: to, type: type, transitSteps: transitTime, elapsedSteps: 0, data: data})
    }


    function getActors() { //Example of use: let total actors = getActors()
        return actors.length;
    }

    function createQueue() { //Example of use: let q = createQueue(); q.push("hey"); let hey = q.pop();
        return new Queue();
    }

    /**
     * @param {Actor} actor
     * @param {number} steps
     * @param {function} reaction
     */
    function timeout(actor, steps, reaction) { //Example of use: timeout(this, 10, fart); function fart() { console.log("fart") }
        timeouts.push({
            steps,
            reaction: reaction.bind(actor)
        });
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
