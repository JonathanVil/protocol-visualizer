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

    /** @type {{ tick: number, lines: string[], state: any }[]} */
    let eventLog = [{ tick: 0, lines: [], state: null}]
    let messages = new Queue();
    let timeouts = new Queue();
    let nextActorId = 0;
    let tick = 0;

    let restoringState = false;
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
        let actor = watchActor(new actorClass(nextActorId++));
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
        console.log("Starting simulation");

        paused = false;
        handleTick();
    }

    function pauseSimulation() {
        console.log("Pausing simulation");
        paused = true;
    }

    /** @type {() => void} */
    let resetGraph;

    function resetSimulation() {
        messages = new Queue();
        timeouts = new Queue();
        eventLog = [];
        actors = [];
        nextActorId = 0;
        nextMessageId = -1;
        tick = 0;
        paused = true;
        resetGraph();
    }

    function tickByOne() {
        handleTick();

    }

    function handleTick() {
        let startTime = Date.now()
        const entry = eventLog.find(e => e.tick === tick);
        if (entry) { //if last tick had an event, we save its state for rewinding
            saveState()
            console.log(entry)
        }

        tick++

        //update messages by one tick
        handleMessages()

        //update timeouts by one tick
        handleTimeouts()



        if (!paused) {
            let elapsedTime = Date.now() - startTime;
            if (elapsedTime > tickSize) {
                console.log(`--------------------TIME TO HANDLE TICK HIGHER THAN TICKSIZE--------------------`);
            }
            setTimeout(handleTick, tickSize - elapsedTime); //we get tick size, not speed, since we want the interval at which we tick, not the frequency of ticks
        }
    }

    /**
     * @param {string} line
     */
    export function addLogEntry(line) {
        const entry = eventLog.find(e => e.tick === tick);
        if (!entry) {
            eventLog = [...eventLog, {tick, lines: [line], state: null}];
        } else {
            entry.lines = [...entry.lines, line];
            eventLog = [...eventLog.slice(0, eventLog.length - 1), entry]
        }
    }

    function saveState() {
        let actorsState = actors.map(actor => {
            /** @type {Record<string, any>} */ // we save the fields in a dict of field name to field value
            let snapshot = {};

            /** @type {Record<string, any>} */ //this pattern basically just allows us to access fields and methods that are not in the actor definition
            const actorObj = actor;

            for (let key of Object.keys(actorObj)) { //these are the properties of our actor
                if (typeof actorObj[key] === "function") continue; //filter out functions

                if (typeof actorObj[key] === "object" && actorObj[key] instanceof Queue) {
                    // save queues
                    snapshot[key] = actorObj[key].toArray().map(/** @param {any} e */ e => structuredClone(e));
                } else {
                    // save normal fields
                    snapshot[key] = structuredClone(actorObj[key]);
                }
            }

            return snapshot;
        });

        let messagesState = messages.toArray().map(m => structuredClone(m)); //we lose methods on clone, so we need an iterable copy in order to restore the queue
        let timeoutsState = timeouts.toArray().map(t => structuredClone(t));


        let state = {actorsState: actorsState, messagesState: messagesState, timeoutsState: timeoutsState};

        const entry = eventLog.find(e => e.tick === tick);
        if (entry){
            entry.state = state; //add the saved state to the entry
            eventLog = [...eventLog.slice(0, eventLog.length - 1), entry] //add the entry to the log
        }

    }
    /** @type {(message: Message) => void} */
    let removeMessageNode;

    /** @type {(actor: Actor) => void} */
    let removeActorNode;
    /** @param {number} restoredTick **/
    function restoreState(restoredTick) {
        if (tick === restoredTick) return; //cant rewind to current tick

        const entry = eventLog.find(e => e.tick === restoredTick);
        if (!entry) return;

        // restore actors
        restoringState = true;

        let actorsState = entry.state.actorsState;

        if (actorsState.length < actors.length) {
            for (let i = actorsState.length; i < actors.length; i++) {
                removeActorNode(actors[i]);
            }
            actors = actors.slice(0, actorsState.length);
        }
        nextActorId = actors.length;

        for (let i = 0; i < actorsState.length; i++) {
            const savedActor = actorsState[i];
            /** @type {Record<string, any>} */
            const actor = actors[i];

            // Restore saved properties
            for (let key of Object.keys(savedActor)) {
                if (typeof actor[key] === "object" && actor[key] instanceof Queue) {
                    // we need to restore fields that are queues a little differently
                    let restoredQueue = new Queue();
                    for (let e of savedActor[key]) { // we saved an array, now we make it a queue
                        restoredQueue.push(e);
                    }
                    actor[key] = restoredQueue;
                } else {
                    //we restore regular fields
                    actor[key] = structuredClone(savedActor[key]);
                }
            }
        }
        for (let actor of actors) {
            updateActorStatePopper(actor); // reflect the updated fields
        }

        // restore messages (note: we dont restore the nextMessageId)
        let restoredMessages = new Queue();
        for (let m of entry.state.messagesState) { // we saved an array, now we make it a queue
            restoredMessages.push(m);
            animateMessage(m)
        }
        for (let m of messages.toArray()) {
            if (!restoredMessages.find(/** @param {Message} msg */ msg => msg.id === m.id)) {
                console.log(m.id)
                removeMessageNode(m);
            }
        }
        messages = restoredMessages

        // restore timeouts
        timeouts = new Queue();
        for (let t of entry.state.timeoutsState) {
            timeouts.push(t);
        }

        tick = restoredTick;

        // clear eventlog entries that happened after where we restored to
        let index = eventLog.indexOf(entry)
        eventLog = eventLog.slice(0, index + 1);

        saveState(); // ensure the copy of state is clean for next rewind

        restoringState = false;
    }

    /** @type {(msg: Message) => void} */
    let animateMessage;
    function handleMessages() {
        let n = messages.length;
        /** @type {Map<number, Message[]>} */
        const deliverableMessages = new Map(); // map of actorID to messages waiting to be delivered to them
        for (let i = 0; i < n; i++) {
            let message = messages.pop()

            if (message == null) continue;

            messages.push(message);

            animateMessage(message)



            if (message.arrivalTick > tick) { // we only look at messages that should be delivered
                continue;
            }

            // group messages by receiver id
            let list = deliverableMessages.get(message.destination);
            if (!list) {
                list = [];
            }

            list.push(message);
            deliverableMessages.set(message.destination, list);
        }

        // deliver messages
        for (const msgs of deliverableMessages.values()) {
            if (msgs.length === 1) {  // if there is only one message scheduled, deliver it
                deliverMessage(msgs[0])
                messages.remove(/** @param {Message} m */ m => m.id === msgs[0].id);
                continue
            }

            // we need to find the messages that have waited longest
            let lowestArrivalTick = msgs[0].arrivalTick;
            for (const msg of msgs) {
                if (msg.arrivalTick < lowestArrivalTick) {
                    lowestArrivalTick = msg.arrivalTick;
                }
            }

            let oldestMsgs = []
            for (const msg of msgs) {
                if (msg.arrivalTick === lowestArrivalTick) {
                    oldestMsgs.push(msg);
                }
            }
            //finally deliver a random message
            let randomIndex = Math.round(Math.random() * (oldestMsgs.length - 1))

            deliverMessage(oldestMsgs[randomIndex]);
            messages.remove(/** @param {Message} m */ m => m.id === oldestMsgs[randomIndex].id);
        }

    }

    function handleTimeouts() {
        let n = timeouts.length;
        for (let i = 0; i < n; i++) {
            let timer = timeouts.pop()
            if (timer != null){
                if (timer.ticks === 0){
                    /** @type {Record<string, any>} */
                    const actor = actors[timer.actorId];

                    actor[timer.reaction]();
                } else {
                    timer.ticks -= 1
                    timeouts.push(timer);
                }

            }
        }
    }

    /** @param {Message} message */
    function removeMessage(message) {
        messages.remove(/** @param {Message} m */ m => m.id === message.id)
    }

    /** @type {(actor: Actor) => void} */
    let updateActorStatePopper;

    /** @param {Actor} actor */
    function watchActor(actor) {
        return new Proxy(actor, {
            set(target, prop, value, receiver) {
                const prev = Reflect.get(target, prop);
                const success = Reflect.set(target, prop, value);

                if (success && prev !== value && !restoringState) { //make sure restoring state is false, so we dont log rewinding
                    let logEntry = `Actor ${target.id} ${String(prop)} changed from ${prev} to ${value}`;
                    console.log(logEntry);
                    addLogEntry(logEntry);
                    updateActorStatePopper(receiver);

                    // reflect nodeColor in graph
                    if (prop === "nodeColor" && changeColor) {
                        changeColor(value, target);
                    }
                }
                return true;
            }
        });
    }

    /** @type {(color: any, actor: Actor) => void} */
    export let changeColor;

    // These are the functions we export into the Actors
    // TODO: put these somewhere nice :)

    /** @param {number} from
     *  @param {number} to
     *  @param {any} data
     *  @param {string} type
     * */
    function send(from, to, type, data) { //Example of use: send(this.id, msg.id, "PING", "Hello")
        if (actors.length - 1 < to) return; // cant send messages to freaks who are not real

        let logEntry = `Actor ${from} sent msg ${type} to Actor ${to}`
        if (data){
            logEntry = `Actor ${from} sent msg ${type} with data ${data} to Actor ${to}`
        }
        console.log(logEntry);
        addLogEntry(logEntry);
        let transitTime = getTransitTime();
        let arrivalTick = tick + transitTime;
        let message = {id: getNextMessageId(), source: from, destination: to, type: type, sentTick: tick, arrivalTick: arrivalTick, data: data}
        messages.push(message)
        animateMessage(message);
    }


    function getActors() { //Example of use: let total actors = getActors()
        return actors.length;
    }

    function createQueue() { //Example of use: let q = createQueue(); q.push("hey"); let hey = q.pop();
        return new Queue()
    }

    /**
     * @param {Actor} actor
     * @param {number} ticks
     * @param {function} reaction
     */
    function timeout(actor, ticks, reaction) { //Example of use: timeout(this, 10, this.fart); function fart() { console.log("fart") }
        timeouts.push({
            ticks,
            actorId: actor.id,
            reaction: reaction.name
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

    /**
     * @param {Message} message
     * @param {Number} delay
     * @returns void
     * */
    function delayMessage(message, delay) {
        if (message.arrivalTick - tick + delay <= 0) {
            deliverMessage(message);
        } else {
            let logEntry = `Message ${message.type} delayed by ${delay} ticks`
            console.log(logEntry);
            addLogEntry(logEntry);
            message.arrivalTick = Number(message.arrivalTick) + Number(delay);
            animateMessage(message);
        }
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
    let nextMessageId = -1;

    /**
     * @return {number} The next message id
     */
    export function getNextMessageId() {
        if (nextMessageId < -1000) {nextMessageId = -1}
        return nextMessageId--;
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
            bind:removeMessageNode={removeMessageNode}
            bind:removeActorNode={removeActorNode}
            bind:messages={messages}
            deliverMessage={deliverMessage}
            delayMessage={delayMessage}
            addLogEntry={addLogEntry}
            removeMessage={removeMessage}
            bind:changeColor={changeColor}
            actors={actors}
            tickSize={tickSize}
            tick={tick}
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
        <EventLog
            eventLog={eventLog}
            restoreState={restoreState}
        />

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



<div id="ui-layer"></div>


{#if settingsPanelOpen}
    <!--Settings block-->
    <SettingsPanel bind:tickSpeed={
        () => Math.floor(1000 / tickSize),
        (v) => {
            tickSize = Math.floor(1000 / v);
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