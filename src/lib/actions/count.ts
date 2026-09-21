// Counts a number from its previous value to the new one. Paired with `use:count={n}`
// it is the only motion in the site that writes to an element's text.
const reduced = () =>
	typeof window !== 'undefined' &&
	window.matchMedia('(prefers-reduced-motion: reduce)').matches;

export function count(node: HTMLElement, value: number) {
	let from = value;
	let frame = 0;

	node.textContent = String(value);

	return {
		update(next: number) {
			if (next === from) return;

			const start = from;
			from = next;
			cancelAnimationFrame(frame);

			// A hidden document produces no frames, so write the answer rather than
			// leaving an intermediate number on screen.
			if (reduced() || document.hidden) {
				node.textContent = String(next);
				return;
			}

			const at = performance.now();
			const span = 520;

			const step = (now: number) => {
				const t = Math.min(1, (now - at) / span);
				// Same easing family the rest of the site uses, so counts settle, not stop.
				const eased = 1 - Math.pow(1 - t, 3);
				node.textContent = String(Math.round(start + (next - start) * eased));
				if (t < 1) frame = requestAnimationFrame(step);
			};

			frame = requestAnimationFrame(step);
		},
		destroy() {
			cancelAnimationFrame(frame);
		}
	};
}
