<script lang="ts">
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
	 * keeps them pointed there while they travel.
	 */

	/**
	 * Where a line sits in the page, in layout terms. Read from `offsetTop`/`offsetHeight`
	 * and the offset parent's box rather than from the line's own rect, because that rect
	 * carries the flight's transform: a return to the top of the page and back down would
	 * otherwise fold the previous flight into the next one's measurement.
	 *
	 * The centre is the edge the flight aligns on, not the bottom — `onPinned` has the
	 * reason.
	 */
	function anchor(node: HTMLElement, card: HTMLElement) {
		const parent = (node.offsetParent as HTMLElement | null) ?? card;
		const box = parent.getBoundingClientRect();

		return {
			left: box.left + node.offsetLeft,
			centre: box.top + node.offsetTop + node.offsetHeight / 2
		};
	}

	// The flight's own aim loop. The header keeps moving until it pins, so a landing measured
	// once is only correct at the instant it was measured: stop mid-flight and the lines land
	// on the slot the header would have reached, which is not where it is. Both ends are read
	// live instead, from layout geometry that no transform can distort. The window outlasts
	// the transition itself, so the last frame is aimed at a settled header; re-writing an
	// unchanged value starts no new transition.
	let aim = 0;

	function stopAim() {
		if (aim) cancelAnimationFrame(aim);
		aim = 0;
	}

	function keepAim(pairs: { node: HTMLElement; target: HTMLElement }[], card: HTMLElement) {
		const until = performance.now() + 1500;
		const legs = pairs.map((pair) => ({
			...pair,
			from: { left: 0, centre: 0 },
			to: { left: 0, centre: 0 }
		}));

		const step = () => {
			// Every read first, then every write: interleaving them forces a synchronous
			// layout per leg, and this runs for 1.5s of frames.
			for (const leg of legs) {
				leg.from = anchor(leg.node, card);
				const slot = leg.target.getBoundingClientRect();
				leg.to = { left: slot.left, centre: slot.top + slot.height / 2 };
			}

			for (const leg of legs) {
				leg.node.style.setProperty('--dx', `${leg.to.left - leg.from.left}px`);
				leg.node.style.setProperty('--dy', `${leg.to.centre - leg.from.centre}px`);
			}

			aim = performance.now() < until ? requestAnimationFrame(step) : 0;
		};

		step();
	}

	$effect(() => () => stopAim());

	function onPinned(value: boolean) {
		stopAim();

		if (!value) {
			pinned = false;
			return;
		}

		pinned = true;

		const legs: { node: HTMLElement; target: HTMLElement }[] = [];

		[cardIndex, cardName, cardCount].forEach((node, i) => {
			const target = [headIndex, headName, headCount][i];
			if (!node || !target) return;

			const to = target.getBoundingClientRect();
			if (!to.height) return;

			// The scale is the two type sizes, not the two boxes: a line set at 96px landing on a
			// slot set at 44px has to arrive at 44px, and the boxes only agree by accident —
			// their line heights come from different rules. Scaling about the line's centre lands
			// the baselines too, exactly, while both lines are set in the same face: the
			// half-leading cancels and only the line heights differ.
			node.style.setProperty('--ds', `${size(target) / size(node)}`);
			legs.push({ node, target });
		});

		if (card && legs.length) keepAim(legs, card);
	}

	const pad = (value: number) => String(value).padStart(2, '0');

	// The size a line is actually set at, whatever its clamp resolved to.
	const size = (node: HTMLElement) => parseFloat(getComputedStyle(node).fontSize);

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
		<h3 class="card__name title medium" bind:this={cardName}>{set.name}</h3>
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
	   onto it, and the deltas arrive as --dx/--dy/--ds. The band itself stays in the flow.
	   Reclaiming its height here — a negative margin the height of the card — is what moved
	   every photograph below it, because the space being given back is above the fold and
	   cannot be reclaimed without moving what is under it. */
	.card--out {
		overflow: visible;
	}

	.card--out .card__index,
	.card--out .card__name,
	.card--out .card__meta {
		/* Over the header's own panel, which is otherwise where the flight ends up. */
		position: relative;
		z-index: 6;
		/* Centre, so the ghost arrives with its baseline on the slot's — see the ratio the
		   script writes into --ds. */
		transform-origin: left center;
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
			box-shadow 0.4s ease;
	}

	/* While the card is in frame it *is* the header, so the header waits — and shows itself
	   the moment the card starts handing over. Only ever hidden once the sentinel is being
	   watched, so a visitor whose scripts never run still sees a set label. */
	.set__head--waiting {
		opacity: 0;
	}

	/* Pinned, the header firms up, so a set label that stops moving still reads as attached to
	   the photographs passing under it. The pinned state must not change the box: the pin
	   happens while somebody is reading the grid below it, and a panel 8px taller for being
	   pinned moves every plate under it at the same instant. */
	.set__head--pinned {
		background: color-mix(in srgb, var(--bg) 94%, transparent);
		box-shadow: 0 1px 0 var(--line);
	}
</style>
