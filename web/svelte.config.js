import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Transpile <script lang="ts"> with Vite (esbuild) rather than relying on the
	// Svelte compiler's built-in type stripping, which only handles a subset of TS.
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter()
	}
};

export default config;
