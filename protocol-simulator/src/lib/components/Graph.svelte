<script>
    import {mount, unmount, onMount, onDestroy} from 'svelte';
    import cytoscape from 'cytoscape';
    import cytoscapePopper from 'cytoscape-popper';
    import {
        computePosition,
        flip,
        shift,
        limitShift,
    } from '@floating-ui/dom';
    import ActorPopper from "$lib/components/ActorPopper.svelte";
    import {writable} from "svelte/store";
    import MessagePopper from "$lib/components/MessagePopper.svelte";
    import {createPopper} from "@popperjs/core";
    import {Queue} from '$lib/datastructures/Queue.js';



    /** @typedef {import('$lib/types.js').Actor} Actor */
    /** @typedef {import('$lib/types.js').Message} Message */

    /** @type {Actor[]} */
    export let actors = [];

    /** @type {number} */
    export let tickSize;

    /** @type {number} */
    export let tick;

    /** @type {(msg: Message) => void} */
    export let removeMessage;

    /** @type {Queue} */
    export let messages = new Queue();
    /** @type {{ source: number, target: number, label: string }[]} */
    let edges = [];

    /** @type {import('cytoscape').NodeSingular[]} */
    let graphMessageNodes = [];

    /** @type {HTMLElement} */
    let cyContainer;

    /** @type {any} */
    let cyInstance;

    /** Map from message ID to corresponding message graph node
     *  @type {Map<number, import('cytoscape').NodeSingular>} */
    const messagesToNodes = new Map();


    /** @typedef {import('svelte/store').Writable<Actor>} ActorStore */
    /**
     * @type {Map<number, {
     *   popper: any,
     *   actorStore: ActorStore,
     *   el: HTMLElement,
     *   component: any,
     *   node: any,
     *   update: () => void
     * }>}
     */
    const poppers = new Map();

    /**
     * @param {number} id
     */
    function removeActorStatePopper(id) {
        const entry = poppers.get(id);
        if (!entry) return;

        // Remove event listeners (must match original handler references)
        entry.node?.off?.('position', entry.update);
        cyInstance?.off?.('pan zoom resize', entry.update);

        // If popper implementation supports destroy, call it (guarded)
        try {
            entry.popper?.destroy?.();
        } catch {
            // ignore
        }

        // Unmount Svelte component + remove its container
        try {
            unmount(entry.component);
        } catch {
            // ignore
        }
        entry.el?.remove();

        poppers.delete(id);
    }

    function removeAllPoppers() {
        for (const id of poppers.keys()) {
            removeActorStatePopper(id);
        }
        poppers.clear();
    }

    /**
     * Used to reset the visuals in the graph
     */
    export function resetGraph() {
        for (const node of graphMessageNodes) {
            removeMessagePopper(node);
        }
        graphMessageNodes = [];
        edges = [];

        cyInstance.elements().remove(); // remove all nodes and edges
        removeAllPoppers();            // remove all poppers + unmount components
    }

    // Helper: convert Actor → cytoscape node
    /**
     * @param {Actor} actor
     */
    function actorToNode(actor) {
        //Check if color contains opacity & is hex
        let color = actor.nodeColor ?? '#1d4ed8';
        if (color.includes("#")) {color = color.slice(0, 7);}

        return {
            data: { id: String(actor.id), color: color }, // cytoscape needs string IDs
            classes: 'actor'
        };
    }

    // Helper: ensure an actor node exists
    /** @param {Actor} actor */
    function ensureActorNode(actor) {
        const id = String(actor.id);
        const existing = cyInstance.getElementById(id);
        if (!existing.empty()) return { added: false, id };

        cyInstance.add(actorToNode(actor));

        // add popper
        updateActorStatePopper(actor);

        return { added: true, id };
    }

    /** @param {Actor} actor */
    export function removeActorNode(actor) {
        const nodeId = String(actor.id);
        cyInstance.getElementById(nodeId).remove();
        removeActorStatePopper(actor.id);
        rearrangeGraph()
    }

    // Helper: ensure an edge exists
    /**
     * @param {{ source: number, target: number, label: string }} e
     */
    function ensureEdge(e) {
        const source = String(e.source);
        const target = String(e.target);
        const edgeId = `${source}->${target}`; // deterministic id prevents duplicates

        const existing = cyInstance.getElementById(edgeId);
        if (!existing.empty()) return false;

        cyInstance.add({
            group: 'edges',
            data: { id: edgeId, source, target, label: e.label ?? "" }
        });
        return true;
    }

    // Helper: adds or updates a "popper" displaying the actor state to cytoscape nodes
    /**
     * @param {Actor} actor
     */
    export function updateActorStatePopper(actor) {
        const id = actor.id;
        let entry = poppers.get(id);
        const node = cyInstance.getElementById(String(id));

        if (!entry) { // if no popper exists yet, we create one

            const uiLayer = document.getElementById("ui-layer");
            if (!uiLayer) {
                console.error("Could not find UI");
                return;
            }

            const el = document.createElement('div');
            el.style.position = 'absolute'; // critical
            uiLayer.appendChild(el);

            const actorStore = writable(actor);

            const update = () => popper.update();
            const component = mount(ActorPopper, {
                target: el,
                props: {
                    store: actorStore,
                    toggleAlive: (actor, originalColor) => {
                        //if actor is alive, it will be killed
                        if (actor.alive) {
                            changeColor("#525252", actor)
                            toggleAlive(actor)
                        } else {
                            changeColor(originalColor ? originalColor : '#1d4ed8', actor)
                            toggleAlive(actor)
                        }
                    },
                    setStateCollapsedGlobal: setActorStateCollapsed, setMethodsCollapsedGlobal: setActorMethodsCollapsed, reposition: update
                },
            });

            const popper = node.popper({
                content: () => el
            });
            node.on('position', update);
            cyInstance.on('pan zoom resize', update);

            entry = { popper, actorStore, el, component, node, update };
            poppers.set(id, entry);
        } else {
            entry.actorStore.set(actor);
        }

        entry.popper.update();
    }

    /** Toggle all Actor Poppers*/
    /** @param {boolean} collapsed */
    function setActorStateCollapsed(collapsed) {
        poppers.forEach(popper => {
            popper?.component.setStateCollapsed(collapsed);
            popper?.update();
        });
    }

    /** @param {boolean} collapsed */
    function setActorMethodsCollapsed(collapsed) {
        poppers.forEach(popper => {
            popper?.component.setMethodsCollapsed(collapsed);
            popper?.update();
        });
    }


    /**
     * @param {cytoscapePopper.RefElement} ref
     * @param {HTMLElement} content
     * @param {cytoscapePopper.PopperOptions|undefined} options
     */
    function popperFactory(ref, content, options) {
        // see https://floating-ui.com/docs/computePosition#options
        const popperOptions = {
            // matching the default behaviour from Popper@2
            // https://floating-ui.com/docs/migration#configure-middleware
            middleware: [
                flip(),
                shift({limiter: limitShift()})
            ],
            ...options,
        }

        function update() {
            computePosition(ref, content, popperOptions).then(({x, y}) => {
                Object.assign(content.style, {
                    left: `${x}px`,
                    top: `${y}px`,
                });
            });
        }
        update();
        return { update };
    }

    cytoscape.use(cytoscapePopper(popperFactory));

    onMount(() => {
        cyInstance = cytoscape({
            container: cyContainer,
            elements: [],
            style: [
                {
                    selector: 'node',
                    style: {
                        'background-color': 'data(color)',
                        label: 'data(id)',
                        color: '#fff',
                        'text-valign': 'center',
                        'text-halign': 'center',
                        'font-size': '8px'
                    }
                },
                {
                    selector: 'edge',
                    style: {
                        width: 2,
                        'line-color': '#707075',
                        'target-arrow-color': '#707075',
                        'curve-style': 'bezier',
                        label: 'data(label)',
                        'font-size': '10px',
                        'text-rotation': 'autorotate',
                        'text-margin-y': -5
                    }
                },
                {
                    selector: '.message',
                    style: {
                        'background-color': '#253478',
                        label: "data(type)",
                        'width': 25,
                        'height': 25
                    }
                }
            ],
            layout: { name: 'circle', animate: true}
        });

        //Listen for clicks on message nodes
        cyInstance.on('tap', '.message',
            /** @param {import('cytoscape').EventObject} evt - The Cytoscape event object*/
            (evt) => {
                createMessagePopper(evt);
            }
        )
        cyInstance.on('tap', 'edge',
            /** @param {import('cytoscape').EventObject} evt - The Cytoscape event object*/
            (evt) => {
                toggleEdge(evt);
            }
        )
    });

    onDestroy(() => {
        // If Graph component is removed, ensure poppers/components don’t leak
        removeAllPoppers();
    });


    /** @type {(event: String) => void}  */
    export let logEvent;

    /** Function used to kill an actor in page.svelte (from graph -> page.svelte)
     * @type {(actor: Actor) => void} */
    export let toggleAlive;

    /** @type {(source: number, target: number) => boolean}  */
    export let toggleConnection;

    /**
     * @param {import('cytoscape').EventObject} evt - The Cytoscape event object
     */
    function toggleEdge(evt) {
        let edge = evt.target;

        const source = Number(edge.source().id());
        const target = Number(edge.target().id());

        let active = toggleConnection(source, target);

        if (active) {
            edge.style('line-color', '#707075');
            edge.style('target-arrow-color', '#707075');
            edge.style('line-style', 'solid');
        } else {
            edge.style('line-color', '#b8b8bd');
            edge.style('target-arrow-color', '#b8b8bd');
            edge.style('line-style', 'dotted');
        }
    }

    /**@param {number} source
     * @param {number} target
     * @param {boolean} state */
    export function setEdgeState(source, target, state) {
        const edgeId = `${source}->${target}`;
        let edge = cyInstance.getElementById(edgeId);

        if (edge.empty()) {
            edge = cyInstance.getElementById(`${target}->${source}`);
        }

        if (edge.empty()) {
            console.warn("Edge not found:", edgeId);
            return;
        }

        if (state) {
            edge.style('line-color', '#707075');
            edge.style('target-arrow-color', '#707075');
            edge.style('line-style', 'solid');
        } else {
            edge.style('line-color', '#b8b8bd');
            edge.style('target-arrow-color', '#b8b8bd');
            edge.style('line-style', 'dotted');
        }
    }


    /**
     * @param {import('cytoscape').EventObject} evt - The Cytoscape event object
     */
    function createMessagePopper(evt) {

        /** @type {import('cytoscape').NodeSingular} */
        const messageNode = evt.target;

        let messagePopUp = messageNode.scratch('messagePopup');

        //If the popup does not already exists
        if (messagePopUp) {console.log("Message popper already exists"); return}

        const uiLayer = document.getElementById("ui-layer");
        if (!uiLayer) {
            console.error("Could not find UI");
            return;
        }

        const messageObject = messages.find( /** @param {Message} m */ m => String(m.id) === messageNode.id() ); // find the message that matches this graph node

        const popperContainer = document.createElement("div");
        popperContainer.style.position = 'absolute';
        uiLayer.appendChild(popperContainer);

        //We then mount a new svelte component (MessagePopper) to the DOM
        const component = mount(MessagePopper, {
            target: popperContainer,
            props: {
                message: messageObject,
                delayMessage: delayMessage,
                deliverMessage: deliverGraphMessage,
                dropMessage: dropMessage,
                closePopper: closePopper,
            }
        });


        //Anchor / reference for messageNode, that the popper can use for position
        const popperReference = messageNode.popperRef();

        const messagePopper = createPopper(popperReference, popperContainer, {
            placement: 'right',
        });

        //Update postion off popper when the message node change positon or zoom
        const update = () => messagePopper.update();
        messageNode.on('position', update);
        cyInstance.on('pan zoom resize', update);

        //Bundle the "Popper": popper instance, svelte component and DOM element
        messagePopUp = {messagePopper, component, popperContainer, update};

        messagesToNodes.set(messageObject.id, messageNode);

        //Save it in the scratch of the node.
        messageNode.scratch('messagePopup', messagePopUp);

    }

    /** @param {Message} message */

    function closePopper(message) {
        const messageNode = messagesToNodes.get(message.id);
        if (messageNode) {
            removeMessagePopper(messageNode);
        } else {
            console.warn("Tried to close popper for message without graph node", message);
        }
        messagesToNodes.delete(message.id);
    }

    /** @param {import('cytoscape').NodeSingular} messageNode */
    function removeMessagePopper(messageNode) {
        const messagePopUp = messageNode.scratch('messagePopup')
        if (!messagePopUp) {return}

        messageNode.removeScratch('messagePopup');

        // Remove event listeners (must match original handler references)
        if (messagePopUp.update) {
            messageNode?.off?.('position', messagePopUp.update);
            cyInstance?.off?.('pan zoom resize', messagePopUp.update);
        }


        // If popper implementation supports destroy, call it (guarded)
        try {
            messagePopUp.popper?.destroy?.();
        } catch {
            // ignore
        }

        // Unmount Svelte component + remove its container
        try {
            unmount(messagePopUp.component);
        } catch (e) {
            console.error(e);
        }
        messagePopUp.container?.remove();

    }

    /** @param {Message} message  */
    function dropMessage(message) {

            removeMessageNode(message)
            let event = `Dropped message ${message.type} from ${message.source} to ${message.destination}`;
            console.log(event);
            logEvent(event);

            //remove message from logic message
            removeMessage(message);


    }

    /** @param {Message} message  */
    export function removeMessageNode (message) {
        const id = message.id;
        const messageNode = messagesToNodes.get(id);
        if (!messageNode) return;


        //remove messageNode (and popper) from graph
        if (messageNode.scratch('messagePopup')) {
            removeMessagePopper(messageNode)
        }
        cyInstance.remove(messageNode);
        graphMessageNodes.splice(graphMessageNodes.indexOf(messageNode), 1);
        messagesToNodes.delete(id);

    }

    export function clearMessageNodes() {
        graphMessageNodes?.forEach(node => {
            cyInstance.remove(node);
        });
    }

    /**
     * @type {(message: Message, delay: number) => void}
     */
    export let delayMessage;

    /** @type {(msg: Message) => void} */
    export let deliverMessage;

    /**
     * //Wrapper function between parent and child to remove the message from logic and graph
     * @param {Message} message */
    function deliverGraphMessage(message) {
        deliverMessage(message);
        dropMessage(message);
    }

    /**
     * @param {any} color
     * @param {Actor} actor
     * */
    export function changeColor(color, actor){
        if (typeof color === 'string') {
            if (color.includes("#")) {color = color.slice(0, 7);}
            const node = cyInstance.getElementById(String(actor.id));
            node.style("background-color", color);
        } else {
            console.error("color not a string", color);
        }
    }

    /** @param {Actor} actor **/
    export function addActorNodeManually(actor){
        if (!actor) {
            console.warn("addActorNodeManually called with undefined actor");
            return;
        }

        cyInstance.batch(() => {

            const { added } = ensureActorNode(actor);
            if (!added) return;

            const newId = String(actor.id);

            const existingActors = cyInstance.nodes(`.actor[id != "${newId}"]`);

            for (const node of existingActors) {

                const otherNum = Number(node.id());
                const newNum = Number(newId);

                const source = otherNum
                const target = newNum


                const newEdge = { source, target, label: "" };

                const edgeAdded = ensureEdge(newEdge);

                if (edgeAdded) {
                    edges = [...edges, newEdge];
                }
            }

            rearrangeGraph();
        });
    }

    //Adding Nodes (incrementally)
    $: if (cyInstance) {

        console.log("Updating graph");

        // Only add what’s new; do NOT remove existing nodes/edges/messages
        cyInstance.batch(() => {
            let addedSomething = false;

            // 1) Ensure all actor nodes exist
            for (const actor of actors) {
                const { added } = ensureActorNode(actor);
                if (added) addedSomething = true;
            }

            // 2) Keep your “connect everyone to newest node” behavior,
            //    but only create missing edges (no duplicates)
            if (actors.length >= 2) {
                const newest = actors[actors.length - 1];
                for (let i = 0; i < actors.length - 1; i++) {
                    const nodeA = actors[i];
                    const newEdge = { source: nodeA.id, target: newest.id, label: "" };

                    // keep your local edges array updated (optional but consistent)
                    // and ensure the edge exists in Cytoscape
                    const edgeAdded = ensureEdge(newEdge);
                    if (edgeAdded) {
                        edges = [...edges, newEdge];
                        addedSomething = true;
                    }
                }
            }

            // 3) Only re-run layout when we actually added nodes/edges
            if (addedSomething) {
                rearrangeGraph()
            }
        });
    }

    function rearrangeGraph() {
        cyInstance.nodes('.actor')
            .layout({name: 'circle', radius: 120, avoidOverlap: true, fit: true})
            .run();

        let n = messages.length;
        for (let i = 0; i < n; i++) {
            let message = messages.pop()
            if (message == null) continue;
            messages.push(message);
            animateMessage(message, true)
        }
    }

    /** @param {Message} message
     *  @param {boolean} instant */
    export function animateMessage(message, instant) {
        const source = cyInstance.getElementById(message.source).position();
        const target = cyInstance.getElementById(message.destination).position();

        if (source == null || target == null) return;

        //Create the message node in graph, if it does not exist.
        let msg = cyInstance.getElementById(message.id)
        if (msg.empty()) {
            msg = cyInstance.add({
                group: 'nodes',
                data: {id: message.id, type: message.type},
                position: {x: source.x, y: source.y},
                classes: 'message'
            });
            graphMessageNodes.push(msg);
            messagesToNodes.set(message.id, msg);
        }

        let elapsedTicks = (tick - message.sentTick)
        let transitTicks = (message.arrivalTick - message.sentTick) - 1
        if (tick > message.arrivalTick){ // dont overshoot if a message is in queue
            elapsedTicks = message.arrivalTick - message.sentTick;
        }

        // elapsedticks             transitTicks
        let targetPosThisTickX = source.x + (((target.x - source.x) * elapsedTicks) / transitTicks)
        let targetPosThisTickY = source.y + (((target.y - source.y) * elapsedTicks) / transitTicks)

        if (elapsedTicks >= transitTicks) {
            targetPosThisTickX = target.x;
            targetPosThisTickY = target.y;
        }

        if (instant) {
            msg.position({ x: targetPosThisTickX, y: targetPosThisTickY });
        } else {
            msg.animate({
                position: {x: targetPosThisTickX, y: targetPosThisTickY}
            }, {
                duration: tickSize,
                easing: 'linear',
                queue: false,
                complete: () => {
                    if (!(messages.find(/** @param {Message} msg */msg => msg.id === message.id))) {
                        // remove message node from graph & array

                        if (msg.scratch('messagePopup')) {
                            removeMessagePopper(msg);
                            messagesToNodes.delete(message.id);
                        }
                        cyInstance.remove(msg);
                        graphMessageNodes.splice(graphMessageNodes.indexOf(msg), 1);

                    }
                }
            });
        }
    }

</script>

<div bind:this={cyContainer} class="w-full h-96 border border-gray-300 rounded-md relative overflow-hidden"></div>
