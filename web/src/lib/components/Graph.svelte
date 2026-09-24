<script module>
    import cytoscape from 'cytoscape';
    import cytoscapePopper from 'cytoscape-popper';
    import {computePosition, flip, shift, limitShift} from '@floating-ui/dom';

    /**
     * @param {cytoscapePopper.RefElement} ref
     * @param {HTMLElement} content
     * @param {cytoscapePopper.PopperOptions|undefined} options
     */
    function popperFactory(ref, content, options) {
        // see https://floating-ui.com/docs/computePosition#options
        const popperOptions = {
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
</script>

<script>
    import {mount, unmount, onMount, untrack} from 'svelte';
    import ActorPopper from "$lib/components/ActorPopper.svelte";
    import MessagePopper from "$lib/components/MessagePopper.svelte";
    import {sim} from "$lib/sim.svelte.js";
    import {payloadLabel} from "$lib/format.js";

    /** @typedef {import('$lib/sim.svelte.js').InTransitMsg} InTransitMsg */

    /** @type {HTMLElement} */
    let cyContainer;

    /** Layer that holds the actor and message poppers. @type {HTMLElement} */
    let uiLayer;

    /** @type {import('cytoscape').Core | null} */
    let cy = $state(null);

    /** Bumped after every layout, so message positions are recomputed. */
    let layoutVersion = $state(0);

    /**
     * A Svelte component mounted next to a graph node.
     * @typedef {{ el: HTMLElement, component: any, node: any, update: () => void }} Popper
     */

    /** @type {Map<number, Popper>} */
    const actorPoppers = new Map();

    /** @type {Map<string, Popper>} */
    const messagePoppers = new Map();

    /**
     * Mounts a component in a popper anchored to a graph node.
     * @param {any} node
     * @param {any} Component
     * @param {Record<string, any>} props may contain getters, which stay reactive
     * @param {object} [options]
     * @returns {Popper}
     */
    function mountPopper(node, Component, props, options) {
        const el = document.createElement('div');
        el.style.position = 'absolute';
        uiLayer.appendChild(el);

        const popper = node.popper({content: () => el, popper: options});
        const update = () => popper.update();
        // Object.assign rather than spreading, which would evaluate (and freeze) the getters.
        const component = mount(Component, {target: el, props: Object.assign(props, {reposition: update})});
        node.on('position', update);
        cy?.on('pan zoom resize', update);
        return {el, component, node, update};
    }

    /** @param {Popper | undefined} entry */
    function unmountPopper(entry) {
        if (!entry) return;
        entry.node.off('position', entry.update);
        cy?.off('pan zoom resize', entry.update);
        unmount(entry.component);
        entry.el.remove();
    }

    /** @param {boolean} collapsed */
    function setActorStateCollapsed(collapsed) {
        for (const popper of actorPoppers.values()) {
            popper.component.setStateCollapsed(collapsed);
            popper.update();
        }
    }

    /** @param {boolean} collapsed */
    function setActorMethodsCollapsed(collapsed) {
        for (const popper of actorPoppers.values()) {
            popper.component.setMethodsCollapsed(collapsed);
            popper.update();
        }
    }

    /** @param {string} id */
    function closeMessagePopper(id) {
        unmountPopper(messagePoppers.get(id));
        messagePoppers.delete(id);
    }

    /** @param {any} node */
    function openMessagePopper(node) {
        const id = node.id();
        if (messagePoppers.has(id)) return;

        messagePoppers.set(id, mountPopper(node, MessagePopper, {
            get message() { return sim.inTransit.find(m => m.id === id); },
            close: () => closeMessagePopper(id),
        }, {placement: 'right'}));
    }

    onMount(() => {
        const instance = cytoscape({
            container: cyContainer,
            elements: [],
            style: [
                {
                    selector: 'node.actor',
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
                        'line-color': '#707075',
                        'curve-style': 'bezier',
                    }
                },
                {
                    selector: 'node.message',
                    style: {
                        'background-color': '#253478',
                        label: 'data(label)',
                        'font-size': '10px',
                        width: 25,
                        height: 25
                    }
                }
            ],
        });

        instance.on('tap', 'node.message', (evt) => openMessagePopper(evt.target));

        cy = instance;
        return () => {
            for (const popper of [...actorPoppers.values(), ...messagePoppers.values()]) {
                unmountPopper(popper);
            }
            actorPoppers.clear();
            messagePoppers.clear();
            instance.destroy();
        };
    });

    // Sync actors → nodes (fully connected, as the backend has no network topology yet).
    $effect(() => {
        if (!cy) return;
        const graph = cy;
        const ids = new Set(sim.actors.map(a => String(a.id)));
        let changed = false;

        graph.batch(() => {
            graph.nodes('.actor').forEach(node => {
                if (ids.has(node.id())) return;
                unmountPopper(actorPoppers.get(Number(node.id())));
                actorPoppers.delete(Number(node.id()));
                node.remove(); // also removes its edges
                changed = true;
            });

            for (const actor of sim.actors) {
                const id = String(actor.id);
                if (!graph.getElementById(id).empty()) continue;

                const others = graph.nodes('.actor');
                graph.add({group: 'nodes', data: {id}, classes: 'actor'});
                others.forEach(other => {
                    graph.add({group: 'edges', data: {id: `${other.id()}-${id}`, source: other.id(), target: id}});
                });
                changed = true;
            }
        });

        if (!changed) return;

        graph.nodes('.actor')
            .layout({name: 'circle', radius: 120, avoidOverlap: true, fit: true})
            .run();

        const newIds = sim.actors.map(a => a.id).filter(id => !actorPoppers.has(id));
        untrack(() => {
            layoutVersion++;

            // Poppers are mounted after layout so they are placed next to their final node position.
            for (const actorId of newIds) {
                actorPoppers.set(actorId, mountPopper(graph.getElementById(String(actorId)), ActorPopper, {
                    get actor() { return sim.actors.find(a => a.id === actorId); },
                    setStateCollapsedGlobal: setActorStateCollapsed,
                    setMethodsCollapsedGlobal: setActorMethodsCollapsed,
                }));
            }
        });
    });

    // Sync in-transit messages → message nodes, animated along their edge.
    $effect(() => {
        if (!cy) return;
        const graph = cy;
        const {tick, running} = sim;
        const duration = sim.settings.tickDurationMs;
        const version = layoutVersion;
        const live = new Set(sim.inTransit.map(m => m.id));

        graph.nodes('.message').forEach(node => {
            if (live.has(node.id())) return;
            untrack(() => closeMessagePopper(node.id()));
            node.remove();
        });

        for (const msg of sim.inTransit) {
            const src = graph.getElementById(String(msg.from));
            const dst = graph.getElementById(String(msg.to));
            if (src.empty() || dst.empty()) continue;

            let node = graph.getElementById(msg.id);
            if (node.empty()) {
                node = graph.add({
                    group: 'nodes',
                    data: {id: msg.id, label: payloadLabel(msg.payload)},
                    position: {...src.position()},
                    classes: 'message',
                });
            }

            // Only restart the animation when something affecting it changed, so
            // unrelated updates (e.g. another message being sent) don't make it jump.
            const key = `${tick}:${msg.deliverAtTick}:${running}:${version}`;
            if (node.scratch('animKey') === key) continue;
            node.scratch('animKey', key);

            const sp = src.position();
            const dp = dst.position();
            const total = Math.max(1, msg.deliverAtTick - msg.sentTick);
            /** @param {number} ticksElapsed */
            const positionAt = (ticksElapsed) => {
                const t = Math.min(1, Math.max(0, ticksElapsed / total));
                return {x: sp.x + (dp.x - sp.x) * t, y: sp.y + (dp.y - sp.y) * t};
            };

            node.stop();
            node.position(positionAt(tick - msg.sentTick));
            if (running) {
                // Head for where the message will be at the next tick.
                node.animate({position: positionAt(tick + 1 - msg.sentTick)}, {duration, easing: 'linear'});
            }
        }
    });
</script>

<div bind:this={cyContainer} class="cy-graph w-full h-96 border border-gray-300 rounded-md relative overflow-hidden"></div>
<div bind:this={uiLayer}></div>
