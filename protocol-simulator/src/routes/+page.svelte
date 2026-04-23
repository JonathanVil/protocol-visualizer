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
    import DocumentationViewer from "$lib/components/DocumentationViewer.svelte";
    import {timeoutsStore} from "$lib/stores.js";

    /** @typedef {import('$lib/types.js').Message} Message */
    /** @typedef {import('$lib/types.js').ActorConstructor} ActorConstructor */
    /** @typedef {import('$lib/types.js').Actor} Actor */
    /** @typedef {import('$lib/types.js').TimeoutEntry} TimeOutEntry */
    /** @typedef {import('$lib/types.js').EditorTab} EditorTab */


    /**@type {{ protocols: { name: string; content: string }[], docs: string }}*/
    export let data; // props from +page.server.js
    let predefinedProtocols = data.protocols;
    let docs = data.docs;

    /** @type {EditorTab[]} */
    let editorTabs = [];

    /** @type {EditorTab | null} */
    let selectedEditorTab = null;

    /** @type {(name: string | null | undefined, code: string | null | undefined) => void} */
    let openNewEditorTab;

    /** @type {Actor[]} */
    let actors = []; // the list of currently active actors

    /** @type {boolean[][]} */
    let actorConnections = []

    /** @type {{ tick: number, lines: string[], state: any }[]} */
    let eventLog = []
    let messages = new Queue();

    let tick = 0;

    let nextTimeoutId = 0;

    let restoringState = false;
    let paused = true;
    let previewingRewind = false;
    /** @type {Actor[]} */
    let cachedActors = [] // used when previewing and rewinding, this is the list of all actors at the latest point in the eventlog

    /** @param {string|null} protocolName */
    function spawnActor(protocolName) {
        if (selectedEditorTab?.model.getValue() == null) return;

        /** @type {ActorConstructor|null} */
        const actorClass = parseProtocolCode(selectedEditorTab?.model.getValue(), send, getActors, deleteTimeout, createQueue, timeout); // we need to give send here so the actor "knows" it

        if (actorClass == null) {
          console.error("Actor class not defined");
          return;
        }

        let nextId = actors.length

        //  svelte automatically updates them in the Graph.svelte
        /** @type {Actor} */
        let newActor = new actorClass(nextId);
        newActor.alive = true;
        newActor.protocolName = protocolName;

        let actor = watchActor(newActor);
        actors = [...actors, actor]; // Must be this way to be reactive in the UI

        actorConnections.push([true])

        for (let i = 0; i < actorConnections.length - 1; i++) {
            actorConnections[i].push(true); // push new actor to other lists
            actorConnections[nextId].push(true); // push other actors to new actor
        }
        console.log(actorConnections)

        let event = "Adding " + protocolName + " actor"
        console.log(event);
        logEvent(event);
    }


    /**
     * Deliver message to destination actor. Transform message to lightweight msg. Lastly invoke actors 'receive' method
     * @param {Message} message
     */
    function deliverMessage(message) {
        // if the message is delivered to a inactive Actor, ignore it
        if (!actors[message.destination].alive) {
            let event = `Actor ${message.destination} would have recieved msg ${message.type} from Actor ${message.source}, but is dead`
            console.log(event);
            logEvent(event);
            return;
        }

        if (!(actorConnections[message.source][message.destination])) {
            let event = `Actor ${message.destination} would have recieved msg ${message.type} from Actor ${message.source}, but connection is severed`
            console.log(event);
            logEvent(event);
            return;
        }

        let event = `Actor ${message.destination} received msg ${message.type} from Actor ${message.source}`
        if (message.data) {
            event = `Actor ${message.destination} received msg ${message.type} with data ${message.data} from Actor ${message.source}`
        }
        console.log(event);
        logEvent(event);
        let actor = actors[message.destination];
        let msg = {type: message.type, from: message.source, data: message.data};
        actor.receive(msg)
    }

    /**
     * @param {number} source
     * @param {number} target
     * @returns {boolean}
     */
    function toggleConnection(source, target) {
        let newState = !(actorConnections[source][target]);
        actorConnections[source][target] = newState;
        actorConnections[target][source] = newState;

        let status = "Connected to";
        if (!newState) {status = "Disconnected from"; }

        let event = `Actor ${source} ${status} Actor ${target}`
        console.log(event);
        logEvent(event);

        return newState;
    }

    /** @type {(source: number, target: number, state: boolean) => void} */
    let setEdgeState;

    function startSimulation() {
        console.log("Starting simulation");
        if (previewingRewind){
            finalizeRewind()
        }

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
        $timeoutsStore = new Queue();
        eventLog = [];
        actors = [];
        cachedActors = [];
        actorRelations = []
        previewingRewind = false;
        nextMessageId = -1;
        tick = 0;
        paused = true;
        resetGraph();
    }

    function tickByOne() {
        if (previewingRewind){
            finalizeRewind()
        }
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
            setTimeout(() => !paused && handleTick(), tickSize - elapsedTime); //we get tick size, not speed, since we want the interval at which we tick, not the frequency of ticks
        }
    }

    /**
     * @param {string} event
     */
    export function logEvent(event) {
        if (previewingRewind) {
            finalizeRewind()
        }
        const entry = eventLog.find(e => e.tick === tick);
        if (!entry) {
            eventLog = [...eventLog, {tick, lines: [event], state: null}];
        } else {
            entry.lines = [...entry.lines, event];
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
        let timeoutsState = $timeoutsStore.toArray().map(t => structuredClone(t));

        let actorConnectionsState = structuredClone(actorConnections);
        let state = {actorsState: actorsState, actorConnectionsState: actorConnectionsState, messagesState: messagesState, timeoutsState: timeoutsState };

        const entry = eventLog.find(e => e.tick === tick);
        if (entry){
            entry.state = state; //add the saved state to the entry
        }

    }
    /** @type {(message: Message) => void} */
    let removeMessageNode;

    /** @type {() => void} */
    let clearMessageNodes;

    /** @type {(actor: Actor) => void} */
    let removeActorNode;
    /** @param {number} restoredTick **/
    function restoreState(restoredTick) {
        if (tick === restoredTick) return; //cant rewind to current tick

        const entry = eventLog.find(e => e.tick === restoredTick);
        if (!entry) return;

        // restore actors
        restoringState = true;
        paused = true;

        previewingRewind = true;

        saveState()

        tick = restoredTick;

        let actorsState = entry.state.actorsState;
        if (actors.length >= cachedActors.length) {
            cachedActors = actors
        }


        if (actorsState.length < actors.length) {
            for (let i = actorsState.length; i < actors.length; i++) {
                removeActorNode(actors[i]);
            }
            actors = actors.slice(0, actorsState.length);
        } else if (actorsState.length > actors.length) {
            // add missing actors back
            for (let i = (actors.length); i < actorsState.length; i++){
                actors.push(cachedActors[i])

                addActorNodeManually(actors[i]);
            }

        }

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
        for (let i = 0; i < actors.length; i++) {
            updateActorStatePopper(actors[i]); // reflect the updated fields
        }

        clearMessageNodes();


        // restore messages (note: we dont restore the nextMessageId)
        let restoredMessages = new Queue();
        for (let m of entry.state.messagesState) { // we saved an array, now we make it a queue
            restoredMessages.push(m);
        }

        for (let m of messages.toArray()) {
            if (!restoredMessages.find(/** @param {Message} msg */ msg => msg.id === m.id)) {
                removeMessageNode(m);
            }
        }

        messages = restoredMessages
        for (let m of messages.toArray()) {
            animateMessage(m, true)
        }

        // restore timeouts
        $timeoutsStore = new Queue();
        for (let t of entry.state.timeoutsState) {
            $timeoutsStore.push(t);
        }

        //restore population state. First kill those who need to die and then revive the rest <3
        for (let actor of actors) {
            if (!actor.alive) {
                changeColor("#525252", actor);
            } else {
                changeColor(actor.nodeColor, actor);
            }
        }

        // restore actorConnections
        actorConnections = entry.state.actorConnectionsState;
        for (let i = 0; i < actorConnections.length; i++) {
            for (let j = i; j < actorConnections.length; j++) {
                if (i === j) continue;
                setEdgeState(i, j, actorConnections[i][j])
            }
        }

        saveState(); // ensure the copy of state is clean for next rewind

        restoringState = false;
    }

    function finalizeRewind() { // Switches from previewing a previous state, to actually executing from that state
        // clear eventlog entries that happened after where we restored to
        let index = eventLog.findIndex(e => e.tick === tick);
        if (index === -1) {
            index = eventLog.findIndex(e => e.tick === tick - 1); //sometimes the tick will be off by one, dont worry about it
        }
        if (index !== -1) {
            eventLog = eventLog.slice(0, index + 1);
        }


        cachedActors = []

        previewingRewind = false;

    }

    /** @type {(msg: Message, instant: boolean) => void} */
    let animateMessage;
    function handleMessages() {
        let n = messages.length;
        /** @type {Map<number, Message[]>} */
        const deliverableMessages = new Map(); // map of actorID to messages waiting to be delivered to them
        for (let i = 0; i < n; i++) {
            let message = messages.pop()

            if (message == null) continue;

            messages.push(message);

            animateMessage(message, false)



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
        let n = $timeoutsStore.length;
        for (let i = 0; i < n; i++) {
            let timer = $timeoutsStore.pop()
            if (timer != null){
                if (timer.ticks === 0){
                    /** @type {Record<string, any>} */
                    const actor = actors[timer.actorId];

                    actor[timer.reaction]();
                } else {
                    timer.ticks -= 1
                    $timeoutsStore.push(timer);
                }
            }
        }
        $timeoutsStore = $timeoutsStore;
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
                    let event = `Actor ${target.id} ${String(prop)} changed from ${prev} to ${value}`;
                    console.log(event);
                    logEvent(event);
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

    /** @type {(actor: Actor) => void} */
    export let addActorNodeManually;

    // These are the functions we export into the Actors

    /** @param {number} from
     *  @param {number} to
     *  @param {string} type
     *  @param {any} data
     * */
    function send(from, to, type, data) { //Example of use: send(this.id, msg.id, "PING", "Hello")
        if (!(to < actors.length && to >= 0)) return; // cant send messages to freaks who are not real
        if (!(from < actors.length && from >= 0)) return;


        let event = `Actor ${from} sent msg ${type} to Actor ${to}`
        if (data){
            event = `Actor ${from} sent msg ${type} with data ${data} to Actor ${to}`
        }
        console.log(event);
        logEvent(event);
        let transitTime = getTransitTime();
        let arrivalTick = tick + transitTime;
        let message = {id: getNextMessageId(), source: Number(from), destination: Number(to), type: type, sentTick: tick, arrivalTick: arrivalTick, data: data}
        messages.push(message)
        animateMessage(message, false);
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
        $timeoutsStore.push({
            id: nextTimeoutId,
            ticks,
            totalTicks: ticks,
            actorId: actor.id,
            reaction: reaction.name
        });
        return nextTimeoutId++;
    }

    /**
     * @param {number} timerId
     */
    function deleteTimeout(timerId) { //Example of use: timeout(this, 10, this.fart); function fart() { console.log("fart") }
        $timeoutsStore.remove(/** @param {TimeOutEntry} t */ t => t.id === timerId)
    }

    let settingsPanelOpen = false;

    export const LeftPanelOptions = {
        DOCS: "docs",
        CODE: "code",
        LOG: "log",
        NONE: "none"
    };
    let leftPanel = LeftPanelOptions.DOCS;

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
            let event = `Message ${message.type} delayed by ${delay} ticks`
            console.log(event);
            logEvent(event);
            message.arrivalTick = Number(message.arrivalTick) + Number(delay);
            animateMessage(message, true);
        }
    }

    /** Makes an actor inactive
     * It no longer can receive message and will remove all its timeouts
     * @param {Actor} actor
     * @returns void
     * */
    export function toggleAlive(actor) {
        if (actor.alive) {
            actor.alive = false;
            $timeoutsStore.remove(/** @param {TimeOutEntry} timeout */ timeout => timeout.actorId === actor.id)
            let event = `Actor ${actor.id} was killed`
            console.log(event);
            logEvent(event);

        } else
        {
            actor.alive = true;
            let event = `Actor ${actor.id} was revived`
            console.log(event);
            logEvent(event);
        }

    }


    /** PROTOCOL UTILS */
    /**
     * @param {string} codeString
     * @param {function} send
     * @param {function} getActors
     * @param {function} deleteTimeout
     * @param {function} createQueue
     * @param {function} timeout
     * @returns {ActorConstructor|null}
     */
    export function parseProtocolCode(codeString, send, getActors, deleteTimeout, createQueue, timeout) {

        try {

            return new Function(
                "send",
                "getActors",
                "deleteTimeout",
                "createQueue",
                "timeout",
                codeString
            )(send, getActors, deleteTimeout, createQueue, timeout);

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

<div class="min-h-screen">
    <!--Top navigation bar-->
    <NavigationBar/>

    <aside
            class="fixed left-0 top-14 z-50 flex h-full w-14 flex-col items-center gap-2 border-r border-slate-200 bg-white/80 py-3"
    >
        <button
                class:active={leftPanel === LeftPanelOptions.DOCS}
                class="p-2 text-slate-300 rounded-md hover:bg-blue-200 aria-pressed:bg-blue-200 border-blue-500"
                aria-label="Log"
                aria-pressed={leftPanel === LeftPanelOptions.DOCS}
                on:click={() => toggleLeftPanel(LeftPanelOptions.DOCS)}
        >
            <Icon icon="mdi:help" class="w-6 h-6 text-black" />
        </button>

        <button
                class:active={leftPanel === LeftPanelOptions.CODE}
                class="p-2 text-slate-300 rounded-md hover:bg-blue-200 aria-pressed:bg-blue-200 border-blue-500"
                aria-label="Code"
                aria-pressed={leftPanel === LeftPanelOptions.CODE}
                on:click={() => toggleLeftPanel(LeftPanelOptions.CODE)}
        >
            <Icon icon="mdi:code-tags" class="w-6 h-6 text-black" />
        </button>

        <button
                class:active={leftPanel === LeftPanelOptions.LOG}
                class="p-2 text-slate-300 rounded-md hover:bg-blue-200 aria-pressed:bg-blue-200 border-blue-500"
                aria-label="Log"
                aria-pressed={leftPanel === LeftPanelOptions.LOG}
                on:click={() => toggleLeftPanel(LeftPanelOptions.LOG)}
        >
            <Icon icon="mdi:clipboard-text-outline" class="w-6 h-6 text-black" />
        </button>
    </aside>

    <!--Dotted graph (background)-->
    <div class="cy-wrapper">
        <Graph
                bind:resetGraph={resetGraph}
                bind:animateMessage={animateMessage}
                bind:updateActorStatePopper={updateActorStatePopper}
                bind:removeMessageNode={removeMessageNode}
                bind:removeActorNode={removeActorNode}
                bind:messages={messages}
                toggleAlive={toggleAlive}
                toggleConnection={toggleConnection}
                deliverMessage={deliverMessage}
                delayMessage={delayMessage}
                logEvent={logEvent}
                removeMessage={removeMessage}
                bind:setEdgeState={setEdgeState}
                bind:changeColor={changeColor}
                bind:addActorNodeManually={addActorNodeManually}
                bind:clearMessageNodes={clearMessageNodes}
                actors={actors}
                tickSize={tickSize}
                tick={tick}
        />
    </div>

    <div id="ui-layer"></div>

    <div class="absolute top-14 left-14 rounded-lg w-9/20 h-4/5">
        {#if leftPanel === LeftPanelOptions.DOCS}
            <DocumentationViewer source={docs} />
        {:else if leftPanel === LeftPanelOptions.CODE}
            <!--Code block-->
            <MonacoEditor
                    bind:tabs={editorTabs}
                    bind:selectedTab={selectedEditorTab}
                    bind:openNewTab={openNewEditorTab}
                    bind:predefinedProtocols={predefinedProtocols}
                    spawnActor={spawnActor}
            />
        {:else if leftPanel === LeftPanelOptions.LOG}
            <!--Log block-->
            <EventLog
                    eventLog={eventLog}
                    restoreState={restoreState}
            />
        {/if}
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
            }}

           bind:transitLower={transitTimeLowerBound}
           bind:transitUpper={transitTimeUpperBound}>
        </SettingsPanel>
    {/if}

    <!--Message block-->
    <div class="absolute bottom-2 left-14">
        <ManualMessageComponent
                send={send}
        />
    </div>

    <!-- 🔹 Bottom Right Buttons -->
    <ControlsPanel
            paused={paused}
            startSimulation={startSimulation}
            tickByOne={tickByOne}
            pauseSimulation={pauseSimulation}
            resetSimulation={resetSimulation}
    />
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