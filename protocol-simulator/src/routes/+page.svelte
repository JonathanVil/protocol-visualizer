<script>
    import MonacoEditor from "$lib/components/MonacoEditor.svelte";
    import SettingsPanel from "$lib/components/SettingsPanel.svelte";
    import NavigationBar from "$lib/components/NavigationBar.svelte";
    import ControlsPanel from "$lib/components/ControlsPanel.svelte";
    import Graph  from "$lib/components/Graph.svelte";
    import {Queue} from '$lib/datastructures/Queue.js';
    import ManualMessageComponent from "$lib/components/ManualMessageComponent.svelte";
    import Icon from '@iconify/svelte';
    import EventLog from "$lib/components/EventLog.svelte";

    /** @typedef {import('$lib/types.js').Message} Message */
    /** @typedef {import('$lib/types.js').ActorConstructor} ActorConstructor */
    /** @typedef {import('$lib/types.js').Actor} Actor */


    /**@type {{ protocols: { name: string; content: string }[] }}*/
    export let data; // props from +page.server.js
    let predefinedProtocols = data.protocols;

    let sourceCode = "// Write your code here...";

    /** @type {Actor[]} */
    let actors = [];

    /** @type {{ tick: number, lines: string[] }[]} */
    let eventLog = [{ tick: 0, lines: []}]

    let messages = new Queue();
    let timeouts = new Queue();
    let id = 0;
    let tick = 0;
    /** @type number */
    let intervalId;

    let tickSpeedUpdated = false;
    let paused = true;

    function spawnActor() {
        /** @type {ActorConstructor|null} */
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
        addLogEntry(logEntry);
    }


    /**
     * Deliver message to destination actor. Transform message to lightweight msg. Lastly invoke actors 'receive' method
     * @param {Message} message
     */
    function deliverMessage(message) {
        let logEntry = `Actor ${message.destination} recieved msg ${message.type} from Actor ${message.source}`
        if (message.data) {
            logEntry = `Actor ${message.destination} recieved msg ${message.type} with data ${message.data} from Actor ${message.source}`
        }
        console.log(logEntry);
        addLogEntry(logEntry);
        let actor = actors[message.destination];
        let msg = {type: message.type, from: message.source, data: message.data};
        actor.receive(msg)
    }

    function startSimulation() {
        if (tickSpeedUpdated) {
            console.log("Updating simulation tickspeed");
        } else {
            console.log("Starting simulation");
        }

        paused = false;
        clearInterval(intervalId);
        tickSpeedUpdated = false;
        intervalId = setInterval(handleTick, tickSize); //we get tick size, not speed, since we want the interval at which we tick, not the frequency of ticks
    }

    function pauseSimulation() {
        console.log("Pausing simulation");
        paused = true;
    }

    /** @type {() => void} */
    let resetGraph;

    function resetSimulation() {
        clearInterval(intervalId);
        messages = new Queue();
        timeouts = new Queue();
        eventLog = [];
        actors = [];
        id = 0;
        tickSpeedUpdated = false;
        tick = 0;
        paused = true;
        resetGraph();
    }

    function tickByOne() {
        if (paused){
            paused = false;
            handleTick();
            paused = true;
        }
    }

    function handleTick() {
        if (paused) {
            return
        }
        tick++

        //update messages by one tick
        handleMessages()

        //update timeouts by one tick
        handleTimeouts()

        //handle updating tickspeed
        if (tickSpeedUpdated) { // We need to reboot the simulation loop in order to update tickspeed
            paused = true;
            startSimulation();
        }
    }

    /**
     * @param {string} line
     */
    export function addLogEntry(line) {
        const entry = eventLog.find(e => e.tick === tick);
        if (!entry) {
            eventLog = [...eventLog, {tick, lines: [line]}];
        } else {
            entry.lines = [...entry.lines, line];
            eventLog = [...eventLog.slice(0, eventLog.length - 1), entry]
        }
    }

    /** @type {(msg: Message) => void} */
    let animateMessage;
    function handleMessages() {
        let n = messages.length;
        for (let i = 0; i < n; i++) {
            let message = messages.pop()
            if (message != null){
                message.elapsedTicks++
                //Animate messages
                animateMessage(message);

                if (message.elapsedTicks=== message.transitTicks){
                    deliverMessage(message)
                } else {
                    messages.push(message)
                }
            }
        }
    }

    function handleTimeouts() {
        let n = timeouts.length;
        for (let i = 0; i < n; i++) {
            let timer = timeouts.pop()
            if (timer != null){
                if (timer.ticks === 0){
                    timer.reaction()
                } else {
                    timer.ticks -= 1
                    timeouts.push(timer);
                }

            }
        }
    }

    /** @type {(actor: Actor) => void} */
    let updateActorStatePopper;

    /** @param {Actor} actor */
    function watchActor(actor) {
        return new Proxy(actor, {
            set(target, prop, value, receiver) {
                const prev = Reflect.get(target, prop);
                const success = Reflect.set(target, prop, value);

                if (success && prev !== value) {
                    let logEntry = `Actor ${target.id} ${String(prop)} changed from ${prev} to ${value}`;
                    console.log(logEntry);
                    addLogEntry(logEntry);
                    updateActorStatePopper(receiver);
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
        let logEntry = `Actor ${from} sent msg ${type} to Actor ${to}`
        if (data){
            logEntry = `Actor ${from} sent msg ${type} with data ${data} to Actor ${to}`
        }
        console.log(logEntry);
        console.log(tickSize);
        addLogEntry(logEntry);
        let transitTime = getTransitTime();
        messages.push({id: getNextMessageId(), source: from, destination: to, type: type, transitTicks: transitTime, elapsedTicks: 0, data: data})
    }


    function getActors() { //Example of use: let total actors = getActors()
        return actors.length;
    }

    function createQueue() { //Example of use: let q = createQueue(); q.push("hey"); let hey = q.pop();
        return new Queue();
    }

    /**
     * @param {Actor} actor
     * @param {number} ticks
     * @param {function} reaction
     */
    function timeout(actor, ticks, reaction) { //Example of use: timeout(this, 10, fart); function fart() { console.log("fart") }
        timeouts.push({
            ticks,
            reaction: reaction.bind(actor)
        });
    }

    let settingsPanelOpen = false;

    export const LeftPanelOptions = {
        CODE: "code",
        LOG: "log",
        NONE: "none"
    };
    let leftPanel = LeftPanelOptions.CODE;

    /** @param {string} panel */
    function toggleLeftPanel(panel) {
        leftPanel = leftPanel === panel ? LeftPanelOptions.NONE : panel;
    }


    /** PROTOCOL UTILS */
    /**
     * @param {string} codeString
     * @param {function} send
     * @param {function} getActors
     * @param {function} createQueue
     * @param {function} timeout
     * @returns {ActorConstructor|null}
     */
    export function parseProtocolCode(codeString, send, getActors, createQueue, timeout) {

        try {

            return new Function(
                "send",
                "getActors",
                "createQueue",
                "timeout",
                codeString
            )(send, getActors, createQueue, timeout);

        } catch (e) {
            console.error('Error parsing code:', e);
            return null;
        }
    }

    let transitTimeUpperBound = 20;
    let transitTimeLowerBound = 20;

    /**
     * @return {number} The transit time in ticks
     */
    export function getTransitTime() {
        return Math.floor(Math.random() * (transitTimeUpperBound - transitTimeLowerBound + 1)) + transitTimeLowerBound;
    }

    let tickSize = 100;

    // id's for messages
    /** @type {number} */
    let id_messages = -1;

    /**
     * @return {number} The next message id
     */
    export function getNextMessageId() {
        if (id_messages < -1000) {id_messages = -1}
        return id_messages--;
    }
</script>

<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">

<!--Top navigation bar-->
<NavigationBar
        bind:predefinedProtocols={predefinedProtocols}
        bind:leftPanel={leftPanel}
        bind:sourceCode={sourceCode}
></NavigationBar>

<!--Dotted graph (background)-->
<div class="cy-wrapper">
    <Graph
            bind:resetGraph={resetGraph}
            bind:animateMessage={animateMessage}
            bind:updateActorStatePopper={updateActorStatePopper}
            bind:messages={messages}
            actors={actors}
            tickSize={tickSize}
    />
</div>


{#if leftPanel === LeftPanelOptions.CODE}
    <!--Code block-->
    <div class="absolute top-24 left-1 rounded-lg w-9/20 h-4/5">
        <MonacoEditor bind:sourceCode={sourceCode} />
    </div>
{:else if leftPanel === LeftPanelOptions.LOG}
    <!--Log block-->
    <div class="absolute top-24 left-1 rounded-lg w-9/20">
        <EventLog eventLog={eventLog} />
    </div>
{/if}


<!--Send actor button-->
<button class="absolute bottom-2 left-120 py-3 bg-blue-600 text-white rounded hover:bg-blue-700 w-25 h-10 text-base flex text-center justify-center items-center"
        on:click={spawnActor}>
    Spawn actor
</button>

<!--Left panel selector-->
<div class="absolute top-14 left-5 flex items-center gap-1 rounded-lg bg-white/80 backdrop-blur p-1 shadow">
    <button
            type="button"
            class="p-1 rounded-md hover:bg-blue-200 aria-[pressed=true]:bg-blue-200"
            aria-label="Toggle Source Code panel"
            title="Source Code"
            aria-pressed={leftPanel === LeftPanelOptions.CODE}
            on:click={() => toggleLeftPanel(LeftPanelOptions.CODE)}
    >
        <Icon icon="mdi:code-tags" class="w-6 h-6 text-black" />
    </button>

    <button
            type="button"
            class="p-1 rounded-md hover:bg-blue-200 aria-[pressed=true]:bg-blue-200"
            aria-label="Toggle Log panel"
            title="Log"
            aria-pressed={leftPanel === LeftPanelOptions.LOG}
            on:click={() => toggleLeftPanel(LeftPanelOptions.LOG)}
    >
        <Icon icon="mdi:clipboard-text-outline" class="w-6 h-6 text-black" />
    </button>
</div>

<button on:click={() => settingsPanelOpen = !settingsPanelOpen} class="absolute top-14 right-5 p-1 rounded-lg hover:bg-blue-200">
    <Icon icon="mdi:menu" class="w-6 h-6 text-black" />
</button>


{#if settingsPanelOpen}
    <!--Settings block-->
    <SettingsPanel bind:tickSpeed={
        () => Math.floor(1000 / tickSize),
        (v) => {
            tickSize = Math.floor(1000 / v);
            tickSpeedUpdated = true;
        }}

       bind:transitLower={transitTimeLowerBound}
       bind:transitUpper={transitTimeUpperBound}>
    </SettingsPanel>
{/if}

<!--Message block-->
<ManualMessageComponent
        send={send}
/>

<!-- 🔹 Bottom Right Buttons -->
<ControlsPanel
        paused={paused}
        startSimulation={startSimulation}
        tickByOne={tickByOne}
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