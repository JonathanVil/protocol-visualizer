<script module lang="ts">
    import cytoscape from 'cytoscape';
    import cytoscapePopper from 'cytoscape-popper';
    import {computePosition, flip, shift, limitShift, type ComputePositionConfig} from '@floating-ui/dom';

    /** What our popper factory returns; cytoscape-popper types it as an empty interface. */
    interface FloatingPopper {
        update(): void;
    }

    const popperFactory: cytoscapePopper.PopperFactory = (ref, content, options): FloatingPopper => {
        // see https://floating-ui.com/docs/computePosition#options
        const popperOptions: Partial<ComputePositionConfig> = {
            middleware: [
                flip(),
                shift({limiter: limitShift()})
            ],
            ...(options as Partial<ComputePositionConfig>),
        };

        function update() {
            computePosition(ref, content, popperOptions).then(({x, y}) => {
                Object.assign(content.style, {
                    left: `${x}px`,
                    top: `${y}px`,
                });
            });
        }
        update();
        return {update};
    };

    cytoscape.use(cytoscapePopper(popperFactory));
</script>

<script lang="ts">
    import {mount, unmount, onMount, untrack, type Component} from 'svelte';
    import type {Core, NodeSingular} from 'cytoscape';
    import ActorPopper from "$lib/components/ActorPopper.svelte";
    import MessagePopper from "$lib/components/MessagePopper.svelte";
    import {sim} from "$lib/sim.svelte";
    import {payloadLabel} from "$lib/format";

    let cyContainer: HTMLElement;

    /** Layer that holds the actor and message poppers. */
    let uiLayer: HTMLElement;

    let cy = $state<Core | null>(null);

    /** Bumped after every layout, so message positions are recomputed. */
    let layoutVersion = $state(0);

    /** A Svelte component mounted next to a graph node. */
    interface Popper<Exports> {
        el: HTMLElement;
        component: Exports;
        node: NodeSingular;
        update: () => void;
    }

    type ActorPopperExports = {
        setStateCollapsed(val: boolean): void;
        setMethodsCollapsed(val: boolean): void;
    };

    const actorPoppers = new Map<number, Popper<ActorPopperExports>>();
    const messagePoppers = new Map<string, Popper<unknown>>();

    /**
     * Mounts a component in a popper anchored to a graph node.
     * `makeProps` receives the popper's reposition function; returned getters stay reactive.
     */
    function mountPopper<Props extends Record<string, any>, Exports extends Record<string, any>>(
        node: NodeSingular,
        Component: Component<Props, Exports>,
        makeProps: (reposition: () => void) => Props,
        options?: Partial<ComputePositionConfig>,
    ): Popper<Exports> {
        const el = document.createElement('div');
        el.style.position = 'absolute';
        uiLayer.appendChild(el);

        const popper = node.popper({content: () => el, popper: options}) as FloatingPopper;
        const update = () => popper.update();
        const component = mount(Component, {target: el, props: makeProps(update)});
        node.on('position', update);
        cy?.on('pan zoom resize', update);
        return {el, component, node, update};
    }

    function unmountPopper(entry: Popper<unknown> | undefined) {
        if (!entry) return;
        entry.node.off('position', undefined, entry.update);
        cy?.off('pan zoom resize', entry.update);
        unmount(entry.component as Record<string, any>);
        entry.el.remove();
    }

    function setActorStateCollapsed(collapsed: boolean) {
        for (const popper of actorPoppers.values()) {
            popper.component.setStateCollapsed(collapsed);
            popper.update();
        }
    }

    function setActorMethodsCollapsed(collapsed: boolean) {
        for (const popper of actorPoppers.values()) {
            popper.component.setMethodsCollapsed(collapsed);
            popper.update();
        }
    }

    function closeMessagePopper(id: string) {
        unmountPopper(messagePoppers.get(id));
        messagePoppers.delete(id);
    }

    function openMessagePopper(node: NodeSingular) {
        const id = node.id();
        if (messagePoppers.has(id)) return;

        messagePoppers.set(id, mountPopper(node, MessagePopper, () => ({
            get message() { return sim.inTransit.find(m => m.id === id); },
            close: () => closeMessagePopper(id),
        }), {placement: 'right'}));
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
                    selector: 'node.actor.dead',
                    style: {
                        'background-color': '#525252',
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

            for (const actor of sim.actors) {
                graph.getElementById(String(actor.id)).toggleClass('dead', !actor.alive);
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
                actorPoppers.set(actorId, mountPopper(graph.getElementById(String(actorId)), ActorPopper, (reposition) => ({
                    get actor() { return sim.actors.find(a => a.id === actorId); },
                    setStateCollapsedGlobal: setActorStateCollapsed,
                    setMethodsCollapsedGlobal: setActorMethodsCollapsed,
                    reposition,
                })));
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
            const positionAt = (ticksElapsed: number) => {
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
