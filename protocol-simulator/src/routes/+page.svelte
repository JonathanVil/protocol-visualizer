<script>
    import MonacoEditer from "$lib/MonacoEditer.svelte";
    import Graph  from "$lib/Graph.svelte";
    import {Queue} from '$lib/Queue.js';
    import ManualMessageComponent from "$lib/ManualMessageComponent.svelte";
    import {getTransitTime, setTransitbounds, getNextMessageId, parseProtocolCode, getStepSize, setStepSize} from "$lib/protocolUtils.js";
    import Icon from '@iconify/svelte';
    import {onMount} from "svelte";

    /** @typedef {import('$lib/types.js').Message} Message */
    /** @typedef {import('$lib/types.js').ActorConstructor} ActorConstructor */
    /** @typedef {import('$lib/types.js').Actor} Actor */


    /**@type {{ protocols: { name: string; content: string }[] }}*/
    export let data; // props from +page.server.js
    let predefinedProtocols = data.protocols;
    /**
	 * @type {{ name: string; content: string; } | null}
	 */
    let selectedProtocol = null;
    
    let sourceCode = "// Write your code here...";

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

    // the actual values of these is not stored here, so these are not important until updated by user
    let transitLowerInput = 8
    let transitUpperInput = 12

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

    function setTransitTimeInput() {
        setTransitbounds(transitUpperInput, transitLowerInput);
    }

    function resetSimulation() {
        clearInterval(intervalId);
        messages = new Queue();
        timeouts = new Queue();
        actors = [];
        id = 0;
        stepSizeUpdated = false;
        paused = true;
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

    //Frontend functions & variables'
    /** @type {HTMLElement | null} */
    let codepanel;
    function enableCodeEditor() {
        if (codepanel != null){
            codepanel.classList.toggle("hidden");
        }
    }

    onMount(() => {
        /** @type {HTMLElement | null} */
        const codeButton = document.getElementById("btn-code");

        codepanel = document.getElementById("codepanel")

        if (codepanel != null){ //toggle code block by default
            codepanel.classList.toggle("hidden");
        }

        /** @type {HTMLElement | null} */
        const settingsButton = document.getElementById("btn-settings");
        /** @type {HTMLElement | null} */
        const settingspanel = document.getElementById("settingspanel");

        if (codeButton && codepanel) {
            codeButton.addEventListener("click", () => { codepanel?.classList.toggle("hidden"); });
        }
        if (settingspanel && settingsButton) {
            settingsButton.addEventListener("click", () => { settingspanel.classList.toggle("hidden"); });
        }
    })
</script>

<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">
<!--Top navigation bar-->
<header class="h-14 bg-white shadow flex items-center justify-between px-6">
    <div class="flex items-center gap-3">
        <div class="w-6 h-6 bg-gray-300 rounded"></div>
        <h1 class="text-lg font-semibold">Protocol Simulator</h1>
    </div>
    <div class="flex items-center justify-center gap-1">

        <select class=" p-2 border border-gray-300 rounded w-fit" bind:value={selectedProtocol}>
            {#each predefinedProtocols as protocol}
                <option value={protocol}>{protocol.name}</option>
            {/each}
        </select>

        <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700" on:click={() => {enableCodeEditor(); sourceCode = selectedProtocol?.content ?? ""}}>Load</button>
    </div>

    <a href="https://github.com/JonathanVil/protocol-visualizer" aria-label="GitHub profile" target="_blank">
        <i class="fa fa-github" style="font-size:36px"></i>
    </a>



</header>

<!--Dotted graph (background)-->
<div class="cy-wrapper">
    <Graph bind:this={graphRef} nodes={actors} />
</div>


<!--Code block-->
<div id="codepanel" class="hidden absolute top-22 left-1 rounded-lg w-9/20 h-4/5">
    <MonacoEditer bind:sourceCode={sourceCode} />
    <button class="absolute bottom-6 right-2 py-3 bg-blue-600 text-white rounded hover:bg-blue-700 w-20 text-xs"
            on:click={spawnActor}>
        Spawn actor
    </button>
</div>


<!--Burger menu's-->
<div class="flex flex-col absolute top-14 ">
    <button id="btn-code" class="p-1 rounded-lg hover:bg-blue-200">
        <Icon icon="mdi:menu" class="w-6 h-6 text-black" />
    </button>
</div>

<div class="flex flex-col absolute top-14 right-5 ">

    <button id="btn-settings" class="p-1 rounded-lg hover:bg-blue-200">
        <Icon icon="mdi:menu" class="w-6 h-6 text-black" />
    </button>
</div>


<!--Settings block-->
<div id="settingspanel" class="hidden absolute top-22 right-1 rounded-lg w-1/7 h-2/5 bg-[#91B7C7]/16 border">
    <div class="flex flex-col gap-3">
        <div class="font-medium">
            <p class="">Step size</p>
        </div>
        <div>
            <input type="range" class="w-3/4" min="10" max="1000" bind:value={stepSizeInput} on:input={setStepSizeInput}>
            <p>{stepSizeInput}</p>
        </div>
        <div class="font-medium">
            <p>Transit time</p>
        </div>
        <div class="flex">
            <div class="flex flex-col">
                <p>Min</p>
                <input type="range" class="w-1/2" min="1" max="{transitUpperInput - 1}" bind:value={transitLowerInput} on:input={setTransitTimeInput}>
                <p>{transitLowerInput}</p>
            </div>
            <div>
                <p>Max</p>
                <input type="range" class="w-1/2" min="{transitLowerInput}" max="50" bind:value={transitUpperInput} on:input={setTransitTimeInput}>
                <p>{transitUpperInput}</p>
            </div>

        </div>
    </div>

</div>

<!--Message block-->
<div class="absolute bottom-2 left-1 rounded-lg w-2/5 h-1/12 bg-white border">
    <ManualMessageComponent messages={messages} />
</div>


<!-- 🔹 Bottom Right Buttons -->
<div class="absolute bottom-5 right-6 flex flex-col gap-3 text-xs">
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
    <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            on:click={resetSimulation}>
        Reset simulation
    </button>
</div>



<style>
    .cy-wrapper {
        width: 100vw;
        height: 95vh;

        /* dots*/
        background-color: #ffffff;
        background-image: radial-gradient(#d1d5db 1px, transparent 1px);
        background-size: 30px 30px;
    }

    /* Sørg for at komponenten indeni rent faktisk fylder wrapperen - google AI */
    :global(.cy-wrapper > *) {
        width: 100% !important;
        height: 100% !important;
    }
</style>