import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		// `fallback` is what puts a 404.html in the deploy: GitHub Pages serves that file
		// for any path that is not a file in the artifact, and the router renders the root
		// error page (src/routes/+error.svelte) for whatever address was asked for.
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: '404.html',
			precompress: false
		})
	}
};

export default config;
