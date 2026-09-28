import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import type { ProxyOptions } from 'vite';

const ASSETS_ORIGIN = 'https://assets.fahadfaruqi.com';

function busted(path: string): string {
	const url = new URL(path, ASSETS_ORIGIN);
	url.searchParams.set('cb', Date.now().toString(36));
	return url.pathname + url.search;
}

const noStore: NonNullable<ProxyOptions['configure']> = (proxy) => {
	proxy.on('proxyRes', (proxyRes) => {
		proxyRes.headers['cache-control'] = 'no-store';
	});
};

const devProxy: Record<string, ProxyOptions> = {
	'/cdn': {
		target: ASSETS_ORIGIN,
		changeOrigin: true,
		rewrite: (path) => busted(path.replace(/^\/cdn/, '')),
		configure: noStore
	},
	'/api': {
		target: ASSETS_ORIGIN,
		changeOrigin: true,
		rewrite: busted,
		configure: noStore
	}
};

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: { proxy: devProxy }
});
