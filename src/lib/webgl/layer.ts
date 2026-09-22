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

let probeResult: boolean | null = null;

/**
 * Whether this browser can give us a context — asked once per document.
 *
 * Asking per plate costs a context per plate: a detached canvas holds onto its context
 * until the browser gets round to collecting it, so 27 plates meant 27 contexts, well
 * over the page's cap, and the browser began killing the oldest contexts — including the
 * one the layer actually draws into. A lost context composites as an opaque white
 * rectangle over the viewport, which is what the white flash was.
 */
function supported(): boolean {
	if (probeResult !== null) return probeResult;

	if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
		probeResult = false;
		unavailable = true;
		return false;
	}

	try {
		const probe = document.createElement('canvas');
		const context = probe.getContext('webgl2') ?? probe.getContext('webgl');
		if (!context) {
			probeResult = false;
			unavailable = true;
			return false;
		}
		// Hand the probe's context straight back: it existed only to answer this question.
		context.getExtension('WEBGL_lose_context')?.loseContext();
		probeResult = true;
		return true;
	} catch {
		probeResult = false;
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

/**
 * Vite replaces this module on every edit up its graph, which resets `layer` to null while
 * the previous instance still holds a canvas and a GL context. Hand the old one back rather
 * than leaking a context per edit.
 */
if (import.meta.hot) {
	import.meta.hot.dispose(() => {
		layer?.dispose();
		layer = null;
		registrations = 0;
		probeResult = null;
	});
}
