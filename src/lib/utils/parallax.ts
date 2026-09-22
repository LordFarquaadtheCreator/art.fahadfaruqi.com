// Scroll parallax for the plates. One loop drives every registered node, because
// twenty-seven independent loops would each read layout on their own schedule; here the
// reads all happen together, once per frame.
const nodes = new Set<HTMLElement>();
let frame = 0;

const reduced = () =>
	typeof window !== 'undefined' &&
	window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function tick() {
	const viewport = window.innerHeight;

	for (const node of nodes) {
		const box = node.getBoundingClientRect();

		// Only the plates on or near the screen are worth measuring.
		if (box.bottom < -200 || box.top > viewport + 200) continue;

		// -1 at the bottom of the viewport, +1 at the top: the plate drifts against the
		// page by a few pixels, and settles to nothing as it reaches the centre.
		const centre = (box.top + box.height / 2 - viewport / 2) / (viewport / 2 + box.height / 2);
		node.style.setProperty('--parallax', `${(-centre * 16).toFixed(2)}px`);
	}

	frame = requestAnimationFrame(tick);
}

function start() {
	if (!frame) frame = requestAnimationFrame(tick);
}

function stop() {
	cancelAnimationFrame(frame);
	frame = 0;
}

export function registerParallax(node: HTMLElement) {
	if (reduced()) {
		node.style.setProperty('--parallax', '0px');
		return { destroy: () => {} };
	}

	nodes.add(node);
	start();

	return {
		destroy: () => {
			nodes.delete(node);
			if (nodes.size === 0) stop();
		}
	};
}
