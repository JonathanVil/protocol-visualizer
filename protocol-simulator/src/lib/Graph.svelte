<script>
    import { onMount } from 'svelte';
    import cytoscape from 'cytoscape';
    import {transitTime} from "$lib/protocolUtils.js";
    import {getStepSize} from "$lib/protocolUtils.js";

    /** @typedef {import('$lib/types.js').Actor} Actor */
    /** @typedef {import('$lib/types.js').Message} Message */

    /** @type {Actor[]} */
    export let nodes = [];

    /** @type {{ source: number, target: number, label: string }[]} */
    export let edges = [];

    /** @type {import('cytoscape').NodeSingular[]} */
    let graphMessages = [];

    /** @type {HTMLElement} */
    let cyContainer;

    /** @type {any} */
    let cyInstance;

    /**
     * Used to reset the visuals in the graph
     */
    export function resetGraph() {
        nodes = [];
        graphMessages = [];
        edges = [];
    }

    // Helper: convert Actor → cytoscape node
    /**
     * @param {Actor} actor
     */
    function actorToNode(actor) {
        return {
            data: { id: String(actor.id) } // cytoscape needs string IDs
        };
    }

    onMount(() => {
        cyInstance = cytoscape({
            container: cyContainer,
            elements: [],
            style: [
                {
                    selector: 'node',
                    style: {
                        'background-color': '#1d4ed8',
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
    });

    //Adding Nodes
    $: if (cyInstance) { //if the instance of the graph is created
        cyInstance.elements().remove();
        cyInstance.add([ //convert out data to cytoscape elements
            ...nodes.map(n => ({ data: { id: n.id} })), //goes through evert node and converts them
        ]);

        for (let i = 0; i < nodes.length - 1; i++) {
            const nodeA = nodes[i];
            const nodeB = nodes[nodes.length - 1];
            edges = [ ...edges, { source: nodeA.id, target: nodeB.id, label: "" } ];
        }
        cyInstance.add([
            ...edges.map(e => ({ data: { source: e.source, target: e.target, label: e.label } })),
        ])


        //run it again
        cyInstance.layout({name: 'circle', radius: 120, avoidOverlap: true, fit: true}).run();

        //Add messages back
        cyInstance.add(
            graphMessages.map(e => ({
                group: 'nodes',
                data: {id: e.id(), type: e.data('type')},
                position: e.position(),
                classes: 'message'
            }))
        )
    }

    /** @param {Message} message */
    export function animateMessage(message) {

        const source = cyInstance.getElementById(message.source).position();
        const target = cyInstance.getElementById(message.destination).position();


        const targetPosThisStepX = source.x + ((target.x - source.x) * message.elapsedSteps) / transitTime
        const targetPosThisStepY = source.y + ((target.y - source.y) * message.elapsedSteps) / transitTime

        //Create the message node in graph, if it does not exists.
        let msg = cyInstance.getElementById(message.id)
        if (msg.empty()) {
            msg = cyInstance.add({
                group: 'nodes',
                data: {id: message.id, type: message.type},
                position: {x: source.x, y: source.y},
                classes: 'message'
            });
            graphMessages.push(msg);
        }

        msg.animate({
            position: {x: targetPosThisStepX, y: targetPosThisStepY}
        }, {
            duration: getStepSize(),
            easing: 'linear',
            queue: false,
            complete: () => {
                if (message.elapsedSteps >= message.transitSteps) {
                    // remove message node from graph
                    cyInstance.remove(msg);
                    //remove the message from array
                    //console.log("Removed message", msg);
                    graphMessages.splice(graphMessages.indexOf(msg), 1);

                }
            }
        });

    }

</script>

<div bind:this={cyContainer} class="w-full h-96 border border-gray-300 rounded-md"></div>
