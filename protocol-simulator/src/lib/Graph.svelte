<script>
    import { onMount } from 'svelte';
    import cytoscape from 'cytoscape';

    /** @type {{ id: string, label: string }[]} */
    export let nodes = [];

    /** @type {{ source: string, target: string, label: string }[]} */
    export let edges = [];

    /** @type {HTMLElement} */
    let cyContainer;

    /** @type {any} */
    let cyInstance;

    onMount(() => {
        cyInstance = cytoscape({
            container: cyContainer,
            elements: [
                ...nodes.map(n => ({ data: { id: n.id, label: n.label } })),
                ...edges.map(e => ({ data: { source: e.source, target: e.target, label: e.label } }))
            ],
            style: [
                {
                    selector: 'node',
                    style: {
                        'background-color': '#1d4ed8',
                        label: 'data(label)',
                        color: '#fff',
                        'text-valign': 'center',
                        'text-halign': 'center',
                        'font-size': '12px'
                    }
                },
                {
                    selector: 'edge',
                    style: {
                        width: 2,
                        'line-color': '#9ca3af',
                        'target-arrow-color': '#9ca3af',
                        'target-arrow-shape': 'triangle',
                        'curve-style': 'bezier',
                        label: 'data(label)',
                        'font-size': '10px',
                        'text-rotation': 'autorotate',
                        'text-margin-y': -5
                    }
                }
            ],
            layout: { name: 'cose', animate: true }
        });
    });


    /**
     * @param {{ id: string, label: string }[]} newNodes
     * @param {{ source: string, target: string, label: string }[]} newEdges
     */
    export function updateGraph(newNodes, newEdges) {
        if (!cyInstance) return;

        cyInstance.elements().remove();

        cyInstance.add([
            ...newNodes.map(n => ({ data: { id: n.id, label: n.label } })),
            ...newEdges.map(e => ({ data: { source: e.source, target: e.target, label: e.label } }))
        ]);

        cyInstance.layout({ name: 'cose', animate: true }).run();
    }
</script>

<div bind:this={cyContainer} class="w-full h-96 border border-gray-300 rounded-md"></div>
