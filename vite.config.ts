import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';
import { resolve } from 'path'

export default defineConfig({
	plugins: [sveltekit()],

	resolve: {
		alias: {
		  	'@': resolve(__dirname, 'src'),
		},
	},

	test: {
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
