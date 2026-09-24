// Hands .nav-button the side the pointer arrived from, so the line draws out of the edge
// the visit started at. CSS cannot see that. Left is the answer for a keyboard visit.
export function navLine(node: HTMLElement) {
	function enter(event: PointerEvent) {
		const box = node.getBoundingClientRect();
		const from = event.clientX - box.left < box.width / 2 ? 'left' : 'right';
		node.style.setProperty('--nav-from', from);
	}

	node.addEventListener('pointerenter', enter);

	return {
		destroy: () => {
			node.removeEventListener('pointerenter', enter);
		}
	};
}
