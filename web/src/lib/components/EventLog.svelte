<script lang="ts">
    import {sim, type SimEvent} from "$lib/sim.svelte";
    import {formatValue} from "$lib/format";

    /** Human-readable description of a simulation event. */
    function describe(event: SimEvent): string {
        switch (event.type) {
            case 'message.sent': {
                const p = event.payload;
                return `Actor ${p.from} sent ${formatValue(p.payload)} to actor ${p.to} (arrives at tick ${p.deliverAtTick})`;
            }
            case 'message.delivered': {
                const p = event.payload;
                return `Delivered ${formatValue(p.payload)} from actor ${p.from} to actor ${p.to}`;
            }
            case 'message.dropped':
                return `Dropped message ${event.payload.messageId}`;
            case 'message.delayed':
                return `Delayed message ${event.payload.messageId} until tick ${event.payload.newDeliverTick}`;
            case 'actor.spawned':
                return `Spawned ${event.payload.typeName} actor ${event.payload.actorId}`;
            case 'actor.fieldChanged': {
                const p = event.payload;
                return `Set ${p.field} = ${formatValue(p.value)} on actor ${p.actorId}`;
            }
            case 'sim.settingsChanged': {
                const p = event.payload;
                return `Settings changed: ${Math.round(1000 / p.tickDurationMs)} ticks/s, transit time ${p.transitTicks} ticks`;
            }
            default: {
                // An event type this client doesn't know about yet.
                const unknown = event as {type: string, payload: unknown};
                return `${unknown.type} ${formatValue(unknown.payload)}`;
            }
        }
    }

    /** Events grouped by tick, newest first. */
    const ticks = $derived.by(() => {
        const groups: {tick: number, lines: string[]}[] = [];
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
