<script lang="ts">
	import './layout.css';

	import type { Snippet } from 'svelte';

	const { children }: { children: Snippet } = $props();
</script>

{@render children()}
