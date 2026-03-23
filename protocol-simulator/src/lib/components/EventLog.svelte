<script>
    /** @type {{ tick: number, lines: string[] }[]} */
    export let eventLog = [];
    export let restoreState;
</script>


<div class="overflow-scroll max-h-[calc(100vh-20rem)] shadow-xl p-2 rounded-md bg-white border border-gray-300">
	<table class="w-full border-collapse">
		<tbody>
			{#if eventLog.length === 0}
				<tr>
					<td colspan="3" class="p-2 text-center">No events yet</td>
				</tr>
			{/if}
			{#each [...eventLog].toReversed() as tick}
				<tr class="hover:bg-blue-100 transition-colors">
					<td class="p-2 font-semibold align-top">{tick.tick}</td>
					<td class="p-2 align-top">
						<button
							class="fa fa-fast-backward cursor-pointer"
							aria-label="Rewind"
							title="Rewind"
							on:click={() => restoreState(tick.tick)}
						></button>
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