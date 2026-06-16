<script lang="ts">
	import { onMount } from 'svelte';
	import { sim } from '$lib/sim.svelte';
	import Graph from '$lib/Graph.svelte';

	let selectedType = $state('');

	onMount(() => {
		sim.connect();
		return () => sim.disconnect();
	});
</script>

<main class="container">
	<h1>Protocol Simulator</h1>

	{#if sim.error}
		<p style="color: var(--pico-del-color)">{sim.error}</p>
	{/if}

	<div style="display: flex; gap: 0.75rem; align-items: center; flex-wrap: wrap; margin-bottom: 1rem;">
		<span>Tick: <strong>{sim.tick}</strong></span>
		<span>{sim.running ? 'Running' : 'Stopped'}</span>

		<button onclick={() => sim.connect()} class="secondary" style="width: auto">Reconnect</button>

		{#if sim.connected}
			<button onclick={() => sim.start()} style="width: auto">Start</button>
			<button onclick={() => sim.stop()} class="secondary" style="width: auto">Stop</button>

			<div role="group">
				<select bind:value={selectedType}>
					{#each sim.actorTypes as t}
						<option value={t}>{t}</option>
					{/each}
				</select>
				<button onclick={() => selectedType && sim.spawn(selectedType)} type="button">
					Spawn
				</button>
			</div>
		{/if}
	</div>

	<Graph />

	<details open style="margin-top: 1rem;">
		<summary>Event log ({sim.eventLog.length})</summary>
		<div style="overflow-x: auto; max-height: 300px; overflow-y: auto;">
			<table>
				<thead>
					<tr>
						<th>Seq</th>
						<th>Tick</th>
						<th>Type</th>
						<th>Payload</th>
					</tr>
				</thead>
				<tbody>
					{#each sim.eventLog.slice().reverse() as event (event.seq)}
						<tr>
							<td>{event.seq}</td>
							<td>{event.tick}</td>
							<td>{event.type}</td>
							<td style="font-size: 0.75rem; font-family: monospace">
								{JSON.stringify(event.payload)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</details>
</main>
