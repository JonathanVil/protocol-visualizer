<script>
    import { onMount } from "svelte";
    import cytoscape from "cytoscape";

    /** @type {HTMLElement} */
    let cyContainer;

    onMount(() => {
        const cy = cytoscape({
            container: cyContainer,

            // ALL ACTORS GOES HERE
            elements: [
                { data: { id: "A", label: "Actor A" } },
                { data: { id: "B", label: "Actor B" } },
                { data: { id: "C", label: "Actor C" } },
                { data: { id: "D", label: "Actor D" } },

                { data: { source: "A", target: "B", label: "" } },
                { data: { source: "A", target: "C", label: "" } },
                { data: { source: "A", target: "D", label: "" } },
            ],

            // STYLING
            style: [
                {
                    selector: "node",
                    style: {
                        "background-color": "#1d4ed8",
                        label: "data(label)",
                        color: "#fff",
                        "text-valign": "center",
                        "text-halign": "center",
                        /** @type {*} */
                        shape: "circle",
                        "font-size": "5px",

                    },
                },
                {
                    selector: "edge",
                    style: {
                        width: 1,
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
            layout: { name: "circle", animate: true },
        });
    });
</script>

<h1 class="text-4xl font-bold mb-4">Cytoscape Example</h1>

<div bind:this={cyContainer} class="w-full h-200 border border-gray-300 rounded-md"></div>