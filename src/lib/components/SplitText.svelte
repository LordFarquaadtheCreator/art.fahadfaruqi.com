<script lang="ts">
	let {
		text,
		segments = 16,
		delay = 0,
		duration = 1.15,
		stagger = 0.03
	}: { text: string; segments?: number; delay?: number; duration?: number; stagger?: number } =
		$props();

	let spacer = $state<HTMLElement | null>(null);

	// Where each grapheme cluster starts, in px from the first cluster's left edge, as it
	// was last measured. `null` until then — the first paint (and every prerender) draws
	// the plain text, so nothing is ever animated from a geometry it has not read.
	let edges = $state<number[] | null>(null);
	let total = $state(0);

	// The slices are built from measured glyph geometry, not from a percentage of the
	// run's width. A percentage assumes every glyph advances by the same amount, which is
	// only true of unshaped Latin: Devanagari conjuncts, Arabic joining forms and Thai
	// reordering all move ink by unequal amounts, so a window edge computed as "1/16 of
	// the total" lands inside a cluster and tears it. Measuring puts the boundaries on
	// real cluster edges instead.
	function measure() {
		const node = spacer;
		if (!node || typeof window === 'undefined' || !('Segmenter' in Intl)) return;

		const textNode = node.firstChild;
		if (!textNode || textNode.nodeType !== Node.TEXT_NODE) return;

		const segmenter = new Intl.Segmenter(undefined, { granularity: 'grapheme' });
		const clusters = [...segmenter.segment(text)].map((s) => s.segment);
		if (clusters.length < segments) return;

		// One rect per cluster, in logical order, on the spacer that already sits in the
		// flow. `visibility: hidden` hides ink without skipping layout or shaping, so the
		// advances read here are the ones the visible copies lay out with. RTL text
		// measures right-to-left, so edges are sorted by true horizontal position.
		const range = document.createRange();
		const points: { start: number; end: number }[] = [];
		let offset = 0;

		for (const cluster of clusters) {
			range.setStart(textNode, offset);
			range.setEnd(textNode, offset + cluster.length);
			offset += cluster.length;

			const rects = range.getClientRects();
			if (rects.length === 0) continue;
			points.push({
				start: Math.min(...Array.from(rects, (r) => r.left)),
				end: Math.max(...Array.from(rects, (r) => r.right))
			});
		}

		if (points.length < segments) return;

		const origin = Math.min(...points.map((p) => p.start));
		const span = Math.max(...points.map((p) => p.end)) - origin;
		if (span <= 0) return;

		// One edge per band boundary, rounded to a whole cluster: a boundary only ever
		// falls between two clusters, never through one. The zero edge and the final edge
		// bound the first and last window.
		const next: number[] = [];
		for (let band = 0; band < segments; band++) {
			const target = (band / segments) * span;
			let edge = origin;
			for (const point of points) {
				if (point.start - origin <= target) edge = point.start;
				else break;
			}
			next.push(edge - origin);
		}
		next.push(span);

		edges = next;
		total = span;
	}

	$effect(() => {
		measure();

		if (typeof window === 'undefined') return;
		// Re-measure on resize: the advance widths are font-size dependent, and `.large` is
		// a `clamp()` on viewport width, so a resize moves every edge.
		const observer = new ResizeObserver(() => measure());
		if (spacer) observer.observe(spacer);
		return () => observer.disconnect();
	});

	const slices = $derived(
		Array.from({ length: segments }, (_, index) => {
			// Measured: window and copy are both in px, so the copy offset is exact.
			if (edges && total > 0) {
				const left = edges[index];
				const right = edges[index + 1];
				return {
					left: `${left}px`,
					width: `${right - left}px`,
					copy: `${total}px`,
					offset: `${-left}px`,
					from: `${(index % 2 === 0 ? -1 : 1) * (14 + index * 4)}%`,
					delay: `${delay + index * stagger}s`
				};
			}

			// Before measurement: equal percentage bands, the original behaviour. The copy
			// is widened so each window still shows its share of the same run.
			const width = 100 / segments + 0.01;
			const share = width / 100;
			const step = 100 / segments / share;
			return {
				left: `${(index / segments) * 100}%`,
				width: `${width}%`,
				copy: `${100 / share}%`,
				offset: `${index * -step}%`,
				from: `${(index % 2 === 0 ? -1 : 1) * (14 + index * 4)}%`,
				delay: `${delay + index * stagger}s`
			};
		})
	);
</script>

<span class="split title large" aria-hidden="true">
	<span class="split__spacer" bind:this={spacer}>{text}</span>
	{#each slices as slice, index (index)}
		<span class="split__window" style="left: {slice.left}; width: {slice.width}">
			<span
				class="split__run"
				style="left: {slice.offset}; width: {slice.copy}; --from: {slice.from}; animation-delay: {slice.delay}; animation-duration: {duration}s"
			>
				{text}
			</span>
		</span>
	{/each}
</span>

<style>
	.split {
		position: relative;
		display: block;
	}

	.split__spacer {
		visibility: hidden;
	}

	/* The window is deliberately taller than the line box the text sits in, and centred
	   on it. The `.title` line-height (0.82) is tight enough on purpose for uppercase
	   Latin, but Devanagari hangs matras and the shirorekha above the Latin ascent, so a
	   window clipped to the line height cuts the tops off `ि`/`ी`/`े` and the nukta. The
	   overhang is expressed in `em` so it scales with the `clamp()`ed display size. */
	.split__window {
		position: absolute;
		top: 50%;
		translate: 0 -50%;
		height: 1.6em;
		overflow: hidden;
	}

	.split__run {
		position: absolute;
		top: 0;
		/* One line box, tall enough to hold the glyphs with their overhead marks. */
		height: 1.6em;
		line-height: 1.6;
		display: block;
		will-change: transform;
		animation-name: assemble;
		animation-timing-function: cubic-bezier(0.16, 0.84, 0.28, 1);
		animation-fill-mode: both;
	}

	@keyframes assemble {
		from {
			transform: translate3d(var(--from), 0, 0);
		}
		to {
			transform: translate3d(0, 0, 0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.split__window {
			display: none;
		}

		.split__spacer {
			visibility: visible;
		}
	}
</style>
