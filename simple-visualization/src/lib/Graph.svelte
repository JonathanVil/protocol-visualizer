<script lang="ts">
    import {onMount} from 'svelte';
    import cytoscape from 'cytoscape';
    import {sim} from '$lib/sim.svelte';

    let container: HTMLElement;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let cy = $state<any>(null);

    onMount(() => {
        const instance = cytoscape({
            container,
            elements: [],
            style: [
                {
                    selector: 'node.actor',
                    style: {
                        'background-color': '#1d4ed8',
                        label: 'data(label)',
                        color: '#fff',
                        'text-valign': 'center',
                        'text-halign': 'center',
                        'font-size': '10px',
                        width: 50,
                        height: 50,
                    },
                },
                {
                    selector: 'edge',
                    style: {
                        width: 2,
                        'line-color': '#9ca3af',
                        'target-arrow-color': '#9ca3af',
                        'target-arrow-shape': 'triangle',
                        'curve-style': 'bezier',
                    },
                },
                {
                    selector: 'node.message',
                    style: {
                        'background-color': '#f59e0b',
                        label: 'data(label)',
                        color: '#000',
                        'text-valign': 'center',
                        'text-halign': 'center',
                        'font-size': '8px',
                        width: 24,
                        height: 24,
                        shape: 'ellipse',
                    },
                },
            ],
            layout: {name: 'circle'},
        });

        cy = instance;
        return () => instance.destroy();
    });

    // Sync actors → Cytoscape nodes + bidirectional edges
    $effect(() => {
        if (!cy) return;

        const actors = sim.actors;
        const existing = new Set<string>(cy.nodes('.actor').map((n: any) => n.id()));
        const changed = existing.size != actors.length;

        for (const actor of actors) {
            const id = String(actor.id);
            if (existing.has(id)) continue;

            cy.add({group: 'nodes', data: {id, label: `${actor.typeName}[${actor.id}]`}, classes: 'actor'});

            for (const existingId of existing) {
                cy.add({group: 'edges', data: {source: existingId, target: id}});
            }

            existing.add(id);
        }

        if (changed) {
            cy.nodes('.actor').layout({name: 'circle', animate: true}).run();
        }
    });

    // Sync inTransit → message nodes at midpoint
    $effect(() => {
        if (!cy) return;

        const live = new Set(sim.inTransit.map((m) => m.id));

        // Remove departed messages
        cy.nodes('.message').forEach((n: any) => {
            if (!live.has(n.id())) cy.remove(n);
        });

        // Add new messages positioned at midpoint between source and destination
        for (const msg of sim.inTransit) {
            const src = cy.getElementById(String(msg.from));
            const tgt = cy.getElementById(String(msg.to));
            if (src.empty() || tgt.empty()) continue;

            const sp = src.position();
            const tp = tgt.position();

            let node = cy.getElementById(msg.id);
            if (node.empty()) {
                cy.add({
                    group: 'nodes',
                    data: {id: msg.id, label: String(msg.payload ?? '')},
                    position: {x: sp.x, y: sp.y},
                    classes: 'message',
                });
            }

            let elapsedTicks = (sim.tick - msg.sentTick);
            let transitTicks = (msg.deliverAtTick - msg.sentTick) - 1;
            let targetPosThisTickX = sp.x + (((tp.x - sp.x) * elapsedTicks) / transitTicks);
            let targetPosThisTickY = sp.y + (((tp.y - sp.y) * elapsedTicks) / transitTicks)

            if (elapsedTicks >= transitTicks) {
                targetPosThisTickX = tp.x;
                targetPosThisTickY = tp.y;
            }

            node.animate({
                position: {x: targetPosThisTickX, y: targetPosThisTickY}
            }, {
                duration: sim.settings.tickDurationMs,
                easing: 'linear',
                queue: false,
                complete: () => {

                }
            })
        }
    });
</script>

<div
        bind:this={container}
        style="width: 100%; height: 400px; border: 1px solid var(--pico-muted-border-color, #d1d5db); border-radius: 6px;"
></div>
