// Animates an element in the first time it comes into view. An element that is already
// on screen is left completely alone, and the hidden state lives behind
// `[data-revealed='false']` — an attribute only this action ever sets. So if the script
// never runs, or dies, everything is visible instead of nothing: an entrance may never
// be the reason a photograph is missing.
export function reveal(node: HTMLElement, margin = '-8%') {
	// Without an observer there is no way to un-hide what this would hide, so hide nothing.
	if (typeof IntersectionObserver === 'undefined') {
		return { destroy: () => {} };
	}

	const box = node.getBoundingClientRect();

	if (box.top < window.innerHeight && box.bottom > 0) {
		return { destroy: () => {} };
	}

	node.dataset.revealed = 'false';

	const observer = new IntersectionObserver(
		(entries) => {
			for (const entry of entries) {
				if (entry.isIntersecting) {
					node.dataset.revealed = 'true';
					observer.disconnect();
				}
			}
		},
		{ rootMargin: `0px 0px ${margin} 0px` }
	);

	observer.observe(node);

	return {
		destroy: () => {
			observer.disconnect();
		}
	};
}
