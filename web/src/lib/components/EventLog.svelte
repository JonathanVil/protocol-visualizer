<script>
    import {sim} from "$lib/sim.svelte.js";
    import {formatValue} from "$lib/format.js";

    /** @typedef {import('$lib/sim.svelte.js').SimEvent} SimEvent */

    /**
     * Human-readable description of a simulation event.
     * @param {SimEvent} event
     */
    function describe({type, payload: p}) {
        switch (type) {
            case 'message.sent':
                return `Actor ${p.from} sent ${formatValue(p.payload)} to actor ${p.to} (arrives at tick ${p.deliverAtTick})`;
            case 'message.delivered':
                return `Delivered ${formatValue(p.payload)} from actor ${p.from} to actor ${p.to}`;
            case 'message.dropped':
                return `Dropped message ${p.messageId}`;
            case 'message.delayed':
                return `Delayed message ${p.messageId} until tick ${p.newDeliverTick}`;
            case 'actor.spawned':
                return `Spawned ${p.typeName} actor ${p.actorId}`;
            case 'actor.fieldChanged':
                return `Set ${p.field} = ${formatValue(p.value)} on actor ${p.actorId}`;
            case 'sim.settingsChanged':
                return `Settings changed: ${Math.round(1000 / p.tickDurationMs)} ticks/s, transit time ${p.transitTicks} ticks`;
            default:
                return `${type} ${formatValue(p)}`;
        }
    }

    /** Events grouped by tick, newest first. */
    const ticks = $derived.by(() => {
        /** @type {{ tick: number, lines: string[] }[]} */
        const groups = [];
        for (const event of sim.eventLog) {
            const last = groups.at(-1);
            if (last?.tick === event.tick) {
                last.lines.push(describe(event));
            } else {
                groups.push({tick: event.tick, lines: [describe(event)]});
            }
        }
        return groups.reverse();
    });
</script>

<div class="flex h-full flex-col">
	<div class="flex items-center justify-between gap-3 rounded-t-md border border-b-0 border-gray-300 bg-gray-100 px-3 py-2">
		<h1 class="text-lg font-medium">Log</h1>
		<p class="text-sm font-semibold text-gray-700">Current tick: {sim.tick}</p>
	</div>

	<div class="overflow-scroll max-h-[calc(100vh-20rem)] shadow-xl p-2 rounded-b-md bg-white border border-gray-300">
		<table class="w-full border-collapse">
			<tbody>
				{#each ticks as tick, i}
					<tr class={`transition-colors hover:bg-blue-100 ${i % 2 === 0 ? 'bg-gray-100' : 'bg-white'}`}>
						<td class="p-2 font-semibold align-top">{tick.tick}</td>
						<td class="p-2">
							{#each tick.lines.toReversed() as line}
								<p>{line}</p>
							{/each}
						</td>
					</tr>
				{:else}
					<tr>
						<td colspan="2" class="p-2 text-center">No events yet</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
