<script>
    /** @type {{ tick: number, lines: string[] }[]} */
    export let eventLog = [];
    export let restoreState;
</script>


<div class="overflow-scroll max-h-[calc(100vh-20rem)] shadow-xl p-2 rounded-md">
	<table class="w-full border-collapse">
		<thead class="sticky top-0 bg-white">
		</thead>
		<tbody>
			{#each [...eventLog].toReversed() as tick}
				<tr class="hover:bg-blue-100 transition-colors">
					<td class="p-2 font-semibold align-top">{tick.tick}</td>
					<td class="p-2 align-top">
						<i
							class="fa fa-fast-backward cursor-pointer"
							aria-hidden="true"
							on:click={() => restoreState(tick.tick)}
						></i>
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