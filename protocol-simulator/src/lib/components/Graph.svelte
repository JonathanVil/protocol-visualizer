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
    import ActorStatePopper from "$lib/components/ActorStatePopper.svelte";
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
            data: { id: String(actor.id), color: color } // cytoscape needs string IDs
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
            const el = document.createElement('div');
            el.style.position = 'absolute'; // critical
            cyContainer.appendChild(el);

            const actorStore = writable(actor);

            const component = mount(ActorStatePopper, {
                target: el,
                props: { store: actorStore }
            });

            const popper = node.popper({
                content: () => el
            });

            const update = () => popper.update();
            node.on('position', update);
            cyInstance.on('pan zoom resize', update);

            entry = { popper, actorStore, el, component, node, update };
            poppers.set(id, entry);
        } else {
            entry.actorStore.set(actor);
        }

        entry.popper.update();
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
                        'line-color': '#9ca3af',
                        'target-arrow-color': '#9ca3af',
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
                createMessagerPopper(evt);
            })
    });

    onDestroy(() => {
        // If Graph component is removed, ensure poppers/components don’t leak
        removeAllPoppers();
    });


    /**
     * @param {import('cytoscape').EventObject} evt - The Cytoscape event object
     */
    function createMessagerPopper(evt) {

        /** @type {import('cytoscape').NodeSingular} */
        const messageNode = evt.target;

        let messagePopUp = messageNode.scratch('messagePopup');

        //If the popup does not already exists
        if (!messagePopUp) {
            const container = document.createElement('div');
            container.style.position = 'absolute';
            container.style.zIndex = '9999';
            container.style.pointerEvents = 'auto';

            cyContainer.appendChild(container);

            const messageObject = messages.find( /** @param {Message} m */ m => String(m.id) === messageNode.id() ); // find the message that matches this graph node

            //We then mount a new svelte component (MessagePopper) to the DOM
            const component = mount(MessagePopper, {
                target: container,
                props: {
                    message: messageObject,
                    delayMessage: delayMessage,
                    deliverMessage: deliverGraphMessage,
                    dropMessage: dropMessage,
                }
            });

            //Anchor / reference for messageNode, that the popper can use for position
            const popperReference = messageNode.popperRef();

            const messagePopper = createPopper(popperReference, container, {
                placement: 'right',
            });

            //Update postion off popper when the message node change positon or zoom
            const update = () => messagePopper.update();
            messageNode.on('position', update);
            cyInstance.on('pan zoom resize', update);

            //Bundle the "Popper": popper instance, svelte component and DOM element
            messagePopUp = {messagePopper, component, container};

            //Save it in the scratch of the node.
            messageNode.scratch('messagePopup', messagePopUp);
        }
    }

    /** @param {import('cytoscape').NodeSingular} messageNode */
    function removeMessagePopper(messageNode) {
        const messagePopUp = messageNode.scratch('messagePopup')

        // Remove event listeners (must match original handler references)
        messageNode?.off?.('position', messagePopUp.update);
        cyInstance?.off?.('pan zoom resize', messagePopUp.update);

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
    export function dropMessage(message) {
        const id = message.id;
        const messageNode = messagesToNodes.get(id);

        if (messageNode) {
            //remove messageNode (and popper) from graph
            removeMessagePopper(messageNode)
            cyInstance.remove(messageNode);
            graphMessageNodes.splice(graphMessageNodes.indexOf(messageNode), 1);
            messagesToNodes.delete(id);

            //remove message from logic message
            messages.remove(/** @param {Message} m */ m => m.id === id)
            console.log("Dropped message", messageNode)
        } else {
            console.log("Could not find message node to drop", messageNode)
        }

    }

    /**
     * @param {Message} message
     * @param {Number} delay
     * */
    export function delayMessage(message, delay) {
        messages.remove(/** @param {Message} m */ m => m.id === message.id)
        message.transitTicks = (message.transitTicks - message.elapsedTicks) + delay;
        messages.push(message)
        animateMessage(message)
    }

    /** @type {(msg: Message) => void} */
    export let deliverMessage;

    /**
     * //Wrapper function between parent and child to remove the message from logic and graph
     * @param {Message} message */
    function deliverGraphMessage(message) {
        deliverMessage(message);
        dropMessage(message);
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
                cyInstance.layout({name: 'circle', radius: 120, avoidOverlap: true, fit: true}).run();
            }
        });
    }

    /** @param {Message} message */
    export function animateMessage(message) {
        const source = cyInstance.getElementById(message.source).position();
        const target = cyInstance.getElementById(message.destination).position();

        const targetPosThisTickX = source.x + ((target.x - source.x) * message.elapsedTicks) / message.transitTicks
        const targetPosThisTickY = source.y + ((target.y - source.y) * message.elapsedTicks) / message.transitTicks

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

        msg.animate({
            position: {x: targetPosThisTickX, y: targetPosThisTickY}
        }, {
            duration: tickSize,
            easing: 'linear',
            queue: false,
            complete: () => {
                if (message.elapsedTicks >= message.transitTicks) {
                    // remove message node from graph & array

                    if (msg.scratch('messagePopup')) {
                        removeMessagePopper(msg);
                    }
                    messagesToNodes.delete(message.id);
                    cyInstance.remove(msg);
                    graphMessageNodes.splice(graphMessageNodes.indexOf(msg), 1);

                }
            }
        });

    }
</script>

<div bind:this={cyContainer} class="w-full h-96 border border-gray-300 rounded-md relative overflow-hidden"></div>
