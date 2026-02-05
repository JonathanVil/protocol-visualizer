<script>
    import { onMount } from 'svelte';
    import cytoscape from 'cytoscape';

    /** @typedef {import('$lib/types.js').Actor} Actor */

    /** @type {Actor[]} */
    export let nodes = [];

    /** @type {{ source: number, target: number, label: string }[]} */
    export let edges = [];

    /** @type {HTMLElement} */
    let cyContainer;

    /** @type {any} */
    let cyInstance;

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
                }
            ],
            layout: { name: 'circle', animate: true}
        });
    });

    //use svelte reactive statement
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
    }


</script>

<div bind:this={cyContainer} class="w-full h-96 border border-gray-300 rounded-md"></div>
