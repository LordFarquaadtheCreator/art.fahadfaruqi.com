<script lang="ts">
	import { tick } from 'svelte';
	import type { Photo } from '$lib/utils/metadata';
	import type { PhotoSet } from '$lib/utils/group-images';

	let { set, index }: { set: PhotoSet; index: number } = $props();

	let entered = $state(false);
	let pinned = $state(false);
	let measuring = $state(false);
	let card = $state<HTMLElement>();
	let cardIndex = $state<HTMLElement>();
	let cardName = $state<HTMLElement>();
	let cardCount = $state<HTMLElement>();
	let headIndex = $state<HTMLElement>();
	let headName = $state<HTMLElement>();
	let headCount = $state<HTMLElement>();

	function onEnter(node: HTMLElement, notify: (value: boolean) => void) {
		if (typeof IntersectionObserver === 'undefined') {
			notify(true);
			return { destroy: () => {} };
		}

		const observer = new IntersectionObserver(
			// Only an arrival from below replays the entrance. Coming back up into the card
			// is a return, not an entrance, and replaying it there is what looked like the
			// card resetting itself every time it left the viewport.
			([entry]) => notify(entry.isIntersecting && entry.boundingClientRect.top > 0),
			{ threshold: 0.4 }
		);
		observer.observe(node);

		return { destroy: () => observer.disconnect() };
	}

	// The sentinel stands where the header pins, so one observation answers both questions:
	// the header is pinned, and the card is on its way out.
	function sticky(node: HTMLElement, notify: (value: boolean) => void) {
		if (typeof IntersectionObserver === 'undefined') {
			return { destroy: () => {} };
		}

		const root = document.documentElement;
		const gap = parseFloat(getComputedStyle(root).fontSize) || 16;
		const header = parseFloat(getComputedStyle(root).getPropertyValue('--header-h')) || 3;

		// How far above the header line the hand-off starts, in root-em. The card's content
		// sits above its own bottom edge — the band's bottom padding — so waiting for the
		// sentinel to clear the line means the lines are already above the top of the frame
		// and the entire flight happens off-screen. The lead pulls it forward to where the
		// shrink is still watchable. 0 is the strict reading of "scrolled out of frame".
		const LEAD = 3.5;

		const observer = new IntersectionObserver(([entry]) => notify(!entry.isIntersecting), {
			rootMargin: `-${(header + LEAD) * gap}px 0px 0px 0px`
		});

		observer.observe(node);
		measuring = true;

		return { destroy: () => observer.disconnect() };
	}

	/**
	 * Where the card's lines have to travel to sit on the header's own, and the aim that
	 * keeps them pointed there while they travel. The card's height is recorded first,
	 * because it is what the band closes by.
	 */

	/**
	 * Where a line sits in the page, in layout terms. Read from `offsetTop`/`offsetHeight`
	 * and the offset parent's box rather than from the line's own rect, because that rect
	 * carries the flight's transform: a second hand-off (the closing band re-crosses the
	 * sentinel) would otherwise fold the previous flight into the next one's measurement.
	 */
	function anchor(node: HTMLElement, card: HTMLElement) {
		const parent = (node.offsetParent as HTMLElement | null) ?? card;
		const box = parent.getBoundingClientRect();

		return {
			left: box.left + node.offsetLeft,
			bottom: box.top + node.offsetTop + node.offsetHeight
		};
	}

	// The flight's own aim loop. The header keeps moving until it pins, and the band closes
	// underneath while the lines travel, so a landing measured once is only correct at the
	// instant it was measured: stop mid-flight and the lines land on the slot the header
	// would have reached, which is not where it is. Both ends are read live instead, from
	// layout geometry that no transform can distort. The window outlasts the collapse's
	// 0.5s delay plus the transition, so the last frame is aimed at a settled header;
	// re-writing an unchanged value starts no new transition.
	let aim = 0;

	function stopAim() {
		if (aim) cancelAnimationFrame(aim);
		aim = 0;
	}

	function keepAim(pairs: { node: HTMLElement; target: HTMLElement }[], card: HTMLElement) {
		const until = performance.now() + 1500;
		const legs = pairs.map((pair) => ({
			...pair,
			from: { left: 0, bottom: 0 },
			to: pair.target.getBoundingClientRect()
		}));

		const step = () => {
			// Every read first, then every write: interleaving them forces a synchronous
			// layout per leg, and this runs for 1.5s of frames.
			for (const leg of legs) {
				leg.from = anchor(leg.node, card);
				leg.to = leg.target.getBoundingClientRect();
			}

			for (const leg of legs) {
				leg.node.style.setProperty('--dx', `${leg.to.left - leg.from.left}px`);
				leg.node.style.setProperty('--dy', `${leg.to.bottom - leg.from.bottom}px`);
			}

			aim = performance.now() < until ? requestAnimationFrame(step) : 0;
		};

		step();
	}

	$effect(() => () => stopAim());

	async function onPinned(value: boolean) {
		stopAim();

		if (!value) {
			pinned = false;
			return;
		}

		if (card) card.style.setProperty('--h', `${card.offsetHeight}px`);

		pinned = true;
		await tick();

		// The header tightens into its pinned padding as the class lands, and measuring a
		// moving box bakes that drift into the landing. Hold its transition for the read.
		const head = headName?.closest('.set__head') as HTMLElement | null;
		const held = head?.style.transition ?? '';
		if (head) head.style.transition = 'none';

		const legs: { node: HTMLElement; target: HTMLElement }[] = [];

		[cardIndex, cardName, cardCount].forEach((node, i) => {
			const target = [headIndex, headName, headCount][i];
			if (!node || !target) return;

			const to = target.getBoundingClientRect();
			if (!to.height) return;

			// The scale is the one value a transform cannot corrupt on either side: the
			// source's height is its layout height and the target's is untransformed.
			node.style.setProperty('--ds', `${to.height / node.offsetHeight}`);
			legs.push({ node, target });
		});

		if (head) head.style.transition = held;

		if (card && legs.length) keepAim(legs, card);
	}

	const pad = (value: number) => String(value).padStart(2, '0');

	// The most recent capture date in the set, from the listing's own timestamps.
	const latest = (photos: Photo[]) =>
		photos.reduce((newest, photo) => (photo.uploaded > newest ? photo.uploaded : newest), '').slice(0, 10);
</script>

<section
	class="card"
	class:card--in={entered && !pinned}
	class:card--out={pinned}
	aria-hidden="true"
	use:onEnter={(value) => (entered = value)}
	bind:this={card}
>
	<div class="card__inner fade-in-animation">
		<span class="label num card__index" bind:this={cardIndex}>Set {pad(index + 1)}</span>
		<h3 class="title medium" bind:this={cardName}>{set.name}</h3>
		<span class="label num card__meta" bind:this={cardCount}>
			{set.photos.length} photos &middot; {latest(set.photos)}
		</span>
	</div>
	<span class="card__rule"></span>
</section>

<span class="set__sentinel" aria-hidden="true" use:sticky={onPinned}></span>
<header
	class="set__head"
	class:set__head--pinned={pinned}
	class:set__head--waiting={measuring && !pinned}
>
	<span class="label num" bind:this={headIndex}>{pad(index + 1)}</span>
	<h2 class="set__name" bind:this={headName}>{set.name}</h2>
	<span class="label num" bind:this={headCount}>{set.photos.length} photographs</span>
</header>

<style>
	.card {
		position: relative;
		display: flex;
		align-items: flex-end;
		min-height: clamp(8rem, 24vh, 15rem);
		padding-block: clamp(1.5rem, 6vh, 4rem) clamp(1rem, 3vh, 2rem);
		overflow: hidden;
		/* Opening the band back up is not delayed: the lines come home into a box that is
		   already the right size. Closing it is, so they travel first. */
		transition: margin-bottom 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.card__inner {
		display: grid;
		gap: 0.5rem;
		width: 100%;
	}

	.card__index,
	.card__meta {
		color: var(--muted);
	}

	/* The rule is the projection: it draws itself across the band as the set arrives.
	   Its default is drawn, so a visitor whose animations never run still sees it. */
	.card__rule {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 1px;
		background: var(--accent);
		transform: scaleX(1);
		transform-origin: left center;
	}

	.card--in .card__index,
	.card--in .card__name,
	.card--in .card__meta {
		animation: card-in 0.75s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.card--in .card__name {
		animation-delay: 0.06s;
	}

	.card--in .card__meta {
		animation-delay: 0.12s;
	}

	.card--in .card__rule {
		animation: card-rule 0.9s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	/* Out of frame, the card becomes the header: each line travels to that slot and shrinks
	   onto it, and the band closes behind them. The deltas arrive as --dx/--dy/--ds. The
	   box stays put while they travel, which is what keeps the landing exact — the space
	   itself closes afterwards, by the margin the card is no longer using. */
	.card--out {
		overflow: visible;
		margin-bottom: calc(-1 * var(--h, 0px));
		transition-delay: 0.5s;
	}

	.card--out .card__index,
	.card--out .card__name,
	.card--out .card__meta {
		/* Over the header's own panel, which is otherwise where the flight ends up. */
		position: relative;
		z-index: 6;
		transform-origin: left bottom;
		transform: translate3d(var(--dx, 0px), var(--dy, 0px), 0) scale(var(--ds, 1));
		opacity: 0;
		/* The fade waits for the flight. Letting it run first is what made this read as the
		   card vanishing rather than arriving on the header. */
		transition:
			transform 0.55s cubic-bezier(0.16, 0.84, 0.28, 1),
			opacity 0.28s ease 0.32s;
	}

	.card--out .card__rule {
		opacity: 0;
		transform: scaleX(0);
		transition:
			transform 0.45s ease,
			opacity 0.3s ease;
	}

	/* The hidden state lives inside the keyframes, so a bundle that never runs leaves the
	   card visible rather than blank. */
	@keyframes card-in {
		from {
			opacity: 0;
			transform: translate3d(0, 1.1rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	@keyframes card-rule {
		from {
			transform: scaleX(0);
		}
		to {
			transform: scaleX(1);
		}
	}

	.set__sentinel {
		display: block;
		height: 1px;
		margin-bottom: -1px;
	}

	.set__head {
		position: sticky;
		/* Above the WebGL canvas, so a sticky set label is never drawn over by a
		   photograph passing underneath it. */
		z-index: 5;
		top: var(--header-h);
		display: flex;
		align-items: baseline;
		gap: 1rem;
		padding-block: 0.75rem;
		border-bottom: 1px solid var(--line);
		background: color-mix(in srgb, var(--bg) 82%, transparent);
		backdrop-filter: blur(10px);
		margin-bottom: clamp(1.5rem, 5vh, 4rem);
		transition:
			opacity 0.35s ease,
			background-color 0.4s ease,
			padding 0.4s cubic-bezier(0.16, 0.84, 0.28, 1),
			box-shadow 0.4s ease;
	}

	/* While the card is in frame it *is* the header, so the header waits — and shows itself
	   the moment the card starts handing over. Only ever hidden once the sentinel is being
	   watched, so a visitor whose scripts never run still sees a set label. */
	.set__head--waiting {
		opacity: 0;
	}

	/* Pinned, the header tightens and firms up, so a set label that stops moving still
	   reads as attached to the photographs passing under it. */
	.set__head--pinned {
		background: color-mix(in srgb, var(--bg) 94%, transparent);
		padding-block: 0.5rem;
		box-shadow: 0 1px 0 var(--line);
	}
</style>
