<script>
    import { onMount } from "svelte";
    import cytoscape from "cytoscape";

    /** @type {HTMLElement} */
    let cyContainer;

    onMount(() => {
        const cy = cytoscape({
            container: cyContainer,
            elements: [
                { data: { id: "A", label: "Actor A" } },
                { data: { id: "B", label: "Actor B" } },
                { data: { source: "A", target: "B", label: "msg1" } },
            ],
            style: [
                {
                    selector: "node",
                    style: {
                        "background-color": "#1d4ed8",
                        label: "data(label)",
                        color: "#fff",
                        "text-valign": "center",
                        "text-halign": "center",
                    },
                },
                {
                    selector: "edge",
                    style: {
                        width: 2,
                        "line-color": "#9ca3af",
                        "target-arrow-color": "#9ca3af",
                        "target-arrow-shape": "triangle",
                        "curve-style": "bezier",
                        label: "data(label)",
                        "font-size": "10px",
                        "text-rotation": "autorotate",
                    },
                },
            ],
            layout: { name: "cose", animate: true },
        });
    });
</script>

<h1 class="text-4xl font-bold mb-4">Cytoscape Example</h1>

<div bind:this={cyContainer} class="w-full h-96 border border-gray-300 rounded-md"></div>