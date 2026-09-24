import type { Handle } from '@sveltejs/kit';
import { CANVAS } from '$lib/utils/theme';

export const handle: Handle = ({ event, resolve }) =>
	resolve(event, {
		transformPageChunk: ({ html }) =>
			html
				.replaceAll('%theme.canvas.dark%', CANVAS.dark)
				.replaceAll('%theme.canvas.light%', CANVAS.light)
	});
