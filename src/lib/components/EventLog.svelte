<script>
    /** @type {{ tick: number, lines: string[] }[]} */
    export let eventLog = [];
    export let restoreState;
	/** @type {number | null} */
	export let previewingTick;
	/** @type {number} */
	export let currentTick = 0;
</script>


<div class="flex h-full flex-col">
	<div class="flex items-center justify-between gap-3 rounded-t-md border border-b-0 border-gray-300 bg-gray-100 px-3 py-2">
		<h1 class="text-lg font-medium">Log</h1>
		<p class="text-sm font-semibold text-gray-700">Current tick: {currentTick}</p>
	</div>

	<div class="overflow-scroll max-h-[calc(100vh-20rem)] shadow-xl p-2 rounded-b-md bg-white border border-gray-300">
		<table class="w-full border-collapse">
			<tbody>
				{#if eventLog.length === 0}
					<tr>
						<td colspan="3" class="p-2 text-center">No events yet</td>
					</tr>
				{/if}
				{#each [...eventLog].toReversed() as tick, i}
					<tr
						class={`transition-colors hover:bg-blue-100 ${
							i % 2 === 0 ? 'bg-gray-100' : 'bg-white'
						}`}
					>
						<td class="p-2 font-semibold align-top">{tick.tick}</td>
						<td class="p-2 align-top">
							{#if previewingTick === tick.tick}
								<p>{previewingTick === tick.tick ? 'Previewing' : ''}</p>
							{:else}
								<button
									class="fa fa-fast-backward cursor-pointer"
									aria-label="Rewind"
									title="Rewind"
									on:click={() => restoreState(tick.tick)}
								></button>
							{/if}
						</td>
						<td class="p-2">
							{#each tick.lines.toReversed() as line}
								<p>{line}</p>
							{/each}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>


<style>

    .row {
        padding: 0.5rem;
    }

    .row:nth-child(odd) {
        background-color: white;
    }

    .row:nth-child(even) {
        background-color: #f3f4f6; /* light grey */
    }

</style>