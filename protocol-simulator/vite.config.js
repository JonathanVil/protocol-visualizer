import { defineConfig } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite'; // Import the Vite plugin

export default defineConfig({
    plugins: [
        sveltekit(),
        tailwindcss(), // Add the Tailwind CSS Vite plugin here!
    ],
    css: {
        // Make sure there is no 'postcss' section configured for Tailwind integration here.
        // If a 'postcss' section exists (e.g., added by the CLI or for other plugins like Autoprefixer),
        // ensure it does NOT include the Tailwind CSS plugin.
        // postcss: { plugins: [] },

        // A community recommendation (might help with Vite plugin issues):
        // transformer: 'lightningcss'
    },
});