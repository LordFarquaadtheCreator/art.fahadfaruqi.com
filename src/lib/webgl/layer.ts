import { browser } from '$app/environment';
import { PlateLayer } from './PlateLayer';

let layer: PlateLayer | null = null;
let registrations = 0;
let unavailable = false;

/** Svelte action: hands a plate frame to the WebGL layer. */
export function registerPlate(frame: HTMLElement) {
	return { destroy: register(frame) };
}

/**
 * Registers a plate frame with the WebGL layer, which owns a single canvas appended to
 * the body. The layer is created on first use and torn down when the last plate goes.
 */
function register(frame: HTMLElement): () => void {
	if (!browser || unavailable || !supported()) return () => {};

	layer ??= create();
	if (!layer) return () => {};

	registrations++;
	const unregister = layer.register(frame);

	return () => {
		unregister();
		if (--registrations === 0) {
			layer?.dispose();
			layer = null;
		}
	};
}

/** A dead-end feature should be silent, not a per-plate error loop. */
function supported(): boolean {
	if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
		unavailable = true;
		return false;
	}

	try {
		const probe = document.createElement('canvas');
		const context = probe.getContext('webgl2') ?? probe.getContext('webgl');
		if (!context) unavailable = true;
		return Boolean(context);
	} catch {
		unavailable = true;
		return false;
	}
}

function create(): PlateLayer | null {
	try {
		const canvas = document.createElement('canvas');
		canvas.className = 'plate-canvas';
		canvas.setAttribute('aria-hidden', 'true');
		document.body.appendChild(canvas);
		return new PlateLayer(canvas);
	} catch (error) {
		unavailable = true;
		console.warn('[plate-layer] could not start; photographs stay as plain images', error);
		return null;
	}
}
