// The reel engine. One action per section: it owns the pin — how long the section holds
// the viewport — and splits that scroll into four phases from the section's progression:
//
//   entry   arriving; the crank is parked at 0
//   work    the section's own animation runs, scrubbed 0..1 over the length it measured
//   hold    complete; nothing moves
//   exit    the departure gesture, then the next section threads in
//
// It publishes --pin, --p, --entry and --exit on the section element and calls the
// section's animation once per frame. Nothing hijacks the wheel: the pin is
// position: sticky and the browser's own scrolling drives it, so trackpad, wheel,
// keyboard and scrollbar all keep working.
//
// Nothing here knows what a section looks like — see animation.ts for the contract.

import type { Mark, ReelAnimation, ReelContext } from '$lib/components/reel/animation';
import type { Progression } from '$lib/components/reel/types';

export interface ReelOptions extends Progression {
	animation: ReelAnimation;
}

interface Item {
	node: HTMLElement;
	options: ReelOptions;
	marks: Mark[];
	show: (index: number) => void;
	entryPx: number;
	workPx: number;
	exitPx: number;
	pin: number;
	active: boolean;
}

const items = new Set<Item>();
const watched = new Map<Element, Item>();
let observer: IntersectionObserver | null = null;
let raf = 0;

const clamp = (value: number, low: number, high: number) => Math.min(high, Math.max(low, value));
const pad = (value: number) => String(value).padStart(2, '0');

const reduced = () =>
	typeof window !== 'undefined' &&
	window.matchMedia('(prefers-reduced-motion: reduce)').matches;

function measure(item: Item) {
	const context: ReelContext = { node: item.node, marks: item.marks, vh: window.innerHeight };
	const work = Math.max(1, item.options.animation.measure(context));

	item.entryPx = Math.max(1, item.options.entry * context.vh);
	item.workPx = work;
	item.exitPx = Math.max(1, item.options.exit * context.vh);
	item.pin = item.entryPx + item.workPx + Math.max(0, item.options.hold * context.vh) + item.exitPx;

	item.node.style.setProperty('--pin', `${item.pin.toFixed(1)}px`);
}

function step(item: Item) {
	const box = item.node.getBoundingClientRect();
	const s = clamp(-box.top, 0, item.pin);
	const p = clamp((s - item.entryPx) / item.workPx, 0, 1);
	const entry = clamp(s / item.entryPx, 0, 1);
	const exit = clamp((s - (item.pin - item.exitPx)) / item.exitPx, 0, 1);

	item.node.style.setProperty('--p', p.toFixed(4));
	item.node.style.setProperty('--entry', entry.toFixed(3));
	item.node.style.setProperty('--exit', exit.toFixed(3));

	item.options.animation.tick({
		p,
		entry,
		exit,
		vh: window.innerHeight,
		marks: item.marks,
		show: item.show
	});
}

function tick() {
	raf = requestAnimationFrame(tick);

	for (const item of items) {
		if (item.active) step(item);
	}
}

function start() {
	if (!raf) raf = requestAnimationFrame(tick);
}

function stop() {
	cancelAnimationFrame(raf);
	raf = 0;
}

function watch(item: Item) {
	if (!observer && typeof IntersectionObserver !== 'undefined') {
		observer = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) {
					const entryItem = watched.get(entry.target);
					if (entryItem) entryItem.active = entry.isIntersecting;
				}
			},
			{ rootMargin: '100% 0px 100% 0px' }
		);
	}

	watched.set(item.node, item);
	observer?.observe(item.node);
}

/** Svelte action for one pinned section. The DOM must carry the data-reel-* hooks. */
export function reel(node: HTMLElement, options: ReelOptions) {
	const viewport = node.querySelector<HTMLElement>('[data-reel-viewport]');

	// Reduced motion, or markup that cannot be driven: the section stays a plain, readable
	// document and nothing is hidden.
	if (!viewport || reduced()) {
		return { update: () => {}, destroy: () => {} };
	}

	const count = node.querySelector<HTMLElement>('[data-reel-count]');
	const show = (index: number) => {
		if (count) count.textContent = pad(index);
	};

	const item: Item = {
		node,
		options,
		marks: [...node.querySelectorAll<HTMLElement>('[data-reel-frame]')].map((el) => ({
			el,
			index: Number(el.dataset.index ?? '0')
		})),
		show,
		entryPx: 1,
		workPx: 1,
		exitPx: 1,
		pin: 0,
		active: true
	};

	// Engage before measuring: the section layout only takes its pinned shape under
	// [data-reel='on'].
	node.dataset.reel = 'on';

	items.add(item);
	watch(item);
	measure(item);
	step(item);
	start();

	const resize = new ResizeObserver(() => measure(item));
	resize.observe(viewport);

	return {
		update(next: ReelOptions) {
			item.options = next;
			measure(item);
			step(item);
		},
		destroy() {
			resize.disconnect();
			observer?.unobserve(node);
			watched.delete(node);
			items.delete(item);
			if (items.size === 0) stop();
		}
	};
}
