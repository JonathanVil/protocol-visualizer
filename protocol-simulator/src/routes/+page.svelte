<script>
    import MonacoEditer from "$lib/MonacoEditer.svelte";
    import SettingsPanel from "$lib/SettingsPanel.svelte";
    import NavigationBar from "$lib/NavigationBar.svelte";
    import ControlsPanel from "$lib/ControlsPanel.svelte";
    import Graph  from "$lib/Graph.svelte";
    import {Queue} from '$lib/Queue.js';
    import ManualMessageComponent from "$lib/ManualMessageComponent.svelte";
    import {getTransitTime, getNextMessageId, parseProtocolCode, getStepSize} from "$lib/protocolUtils.js";
    import Icon from '@iconify/svelte';
    import {onMount} from "svelte";
    import EventLog from "$lib/EventLog.svelte";

    /** @typedef {import('$lib/types.js').Message} Message */
    /** @typedef {import('$lib/types.js').ActorConstructor} ActorConstructor */
    /** @typedef {import('$lib/types.js').Actor} Actor */


    /**@type {{ protocols: { name: string; content: string }[] }}*/
    export let data; // props from +page.server.js
    let predefinedProtocols = data.protocols;

    let sourceCode = "// Write your code here...";

    //reference to graph instance
    /** @type {import('$lib/Graph.svelte').default} */
    let graphRef;

    /** @type {Actor[]} */
    let actors = [];
    /** @type {string[]} */
    let tickLog = ["tick 0"];

    /** @type {string[][]} */
    let eventLog = []

    let messages = new Queue();
    let timeouts = new Queue();
    let id = 0;
    let tick = 0;
    /** @type number */
    let intervalId;

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
        let actor = watchActor(new actorClass(id++));
        actors = [...actors, actor];
        let logEntry = "Adding actor"
        console.log(logEntry);
        tickLog.push(logEntry);
    }


    /**
     * Deliver message to destination actor. Transform message to lightweight msg. Lastly invoke actors 'receive' method
     * @param {Message} message
     */
    function deliverMessage(message) {
        let logEntry = `Actor ${message.destination} recieved msg ${message.type} with data ${message.data} from Actor ${message.source}`
        console.log(logEntry);
        tickLog.push(logEntry);
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



    function resetSimulation() {
        clearInterval(intervalId);
        messages = new Queue();
        timeouts = new Queue();
        actors = [];
        id = 0;
        tick = 0;
        stepSizeUpdated = false;
        paused = true;
        graphRef.resetGraph();
    }

    function step() {
        if (paused) {
            return
        }
        tick++;
        //handle messages
        messageStep()

        //handle timeouts
        timeoutStep()

        //handle log -- this should be the last step
        logStep()

        //handle updating stepsize
        if (stepSizeUpdated) { // We need to reboot the simulation loop in order to update stepsize
            paused = true;
            startSimulation();
        }

    }

    function logStep() {
        if (tickLog.length > 1) {
            eventLog = ([... eventLog, [... tickLog]])
        }
        tickLog = [`tick ${tick}`]
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

    /** @param {Actor} actor */
    function watchActor(actor) {
        return new Proxy(actor, {
            set(target, prop, value) {
                const prev = Reflect.get(target, prop);
                const success = Reflect.set(target, prop, value);

                if (success && prev !== value) {
                    let logEntry = `Actor field ${target.id} ${String(prop)} changed from ${prev} to ${value}`
                    console.log(logEntry);
                    tickLog.push(logEntry);
                    graphRef.updateActorStatePopper(target);
                }
                return true;
            }
        });
    }

    // These are the functions we export into the Actors
    // TODO: put these somewhere nice :)

    /** @param {number} from
     *  @param {number} to
     *  @param {any} data
     *  @param {string} type
     * */
    function send(from, to, type, data) { //Example of use: send(this.id, from.id, "PING", "Hello")
        let logEntry = `Actor ${from} sent msg ${type} with data ${data} to Actor ${to}`
        console.log(logEntry);
        tickLog.push(logEntry);
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
<NavigationBar
        bind:predefinedProtocols={predefinedProtocols}
        bind:codepanel={codepanel}
        bind:sourceCode={sourceCode}
></NavigationBar>

<!--Dotted graph (background)-->
<div class="cy-wrapper">
    <Graph bind:this={graphRef} nodes={actors} />
</div>


<!--Code block-->
<div id="codepanel" class="hidden absolute top-22 left-1 rounded-lg w-9/20 h-4/5">
    <MonacoEditer bind:sourceCode={sourceCode} />
</div>

<!--Log block-->
<div id="logpanel" class="absolute top-24 left-1 rounded-lg w-9/20 h-4/5">
    <EventLog bind:eventLog={eventLog} />
</div>



<!--Send actor button-->
<button class="absolute bottom-2 left-120 py-3 bg-blue-600 text-white rounded hover:bg-blue-700 w-25 h-10 text-base flex text-center justify-center items-center"
        on:click={spawnActor}>
    Spawn actor
</button>

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
<SettingsPanel bind:stepSizeUpdated={stepSizeUpdated}></SettingsPanel>

<!--Message block-->
<ManualMessageComponent
        messages={messages}
        bind:tickLog={tickLog}
/>

<!-- 🔹 Bottom Right Buttons -->
<ControlsPanel
        paused={paused}
        startSimulation={startSimulation}
        pauseSimulation={pauseSimulation}
        resetSimulation={resetSimulation}
/>

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