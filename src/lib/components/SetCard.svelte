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

			([entry]) => notify(entry.isIntersecting && entry.boundingClientRect.top > 0),
			{ threshold: 0.4 }
		);
		observer.observe(node);

		return { destroy: () => observer.disconnect() };
	}


	function sticky(node: HTMLElement, notify: (value: boolean) => void) {
		if (typeof IntersectionObserver === 'undefined') {
			return { destroy: () => {} };
		}

		const root = document.documentElement;
		const gap = parseFloat(getComputedStyle(root).fontSize) || 16;
		// A px length, not a number in root-em: the token is registered as a <length>, so this
		// resolves the calc() rather than handing back its source text.
		const header = parseFloat(getComputedStyle(root).getPropertyValue('--header-h')) || 3 * gap;


		const LEAD = 3.5;

		const observer = new IntersectionObserver(([entry]) => notify(!entry.isIntersecting), {
			rootMargin: `-${header + LEAD * gap}px 0px 0px 0px`
		});

		observer.observe(node);
		measuring = true;

		return { destroy: () => observer.disconnect() };
	}


	function anchor(node: HTMLElement, card: HTMLElement) {
		const parent = (node.offsetParent as HTMLElement | null) ?? card;
		const box = parent.getBoundingClientRect();

		return {
			left: box.left + node.offsetLeft,
			centre: box.top + node.offsetTop + node.offsetHeight / 2
		};
	}


	let aim = 0;

	function stopAim() {
		if (aim) cancelAnimationFrame(aim);
		aim = 0;
	}

	function keepAim(pairs: { node: HTMLElement; target: HTMLElement }[], card: HTMLElement) {
		const until = performance.now() + 1500;
		const bar = document.querySelector<HTMLElement>('.header__inner');
		const edge = bar
			? bar.getBoundingClientRect().right - (parseFloat(getComputedStyle(bar).paddingRight) || 0)
			: window.innerWidth;
		const legs = pairs.map((pair) => ({
			...pair,
			from: { left: 0, centre: 0 },
			to: { left: 0, centre: 0 }
		}));

		const step = () => {

			for (const leg of legs) {
				leg.from = anchor(leg.node, card);
				const slot = leg.target.getBoundingClientRect();
				leg.to = { left: slot.left, centre: slot.top + slot.height / 2 };
			}

			for (const leg of legs) {
				const scale = parseFloat(leg.node.style.getPropertyValue('--ds')) || 1;
				const room = edge - leg.from.left - leg.node.offsetWidth * scale;
				const dx = Math.min(leg.to.left - leg.from.left, room);
				leg.node.style.setProperty('--dx', `${dx}px`);
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


			node.style.setProperty('--ds', `${size(target) / size(node)}`);
			legs.push({ node, target });
		});

		if (card && legs.length) keepAim(legs, card);
	}

	const pad = (value: number) => String(value).padStart(2, '0');


	const size = (node: HTMLElement) => parseFloat(getComputedStyle(node).fontSize);


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
		<span class="label  card__index" bind:this={cardIndex}>Set {pad(index + 1)}</span>
		<h3 class="card__name title medium" bind:this={cardName}>{set.name}</h3>
		<span class="label  card__meta" bind:this={cardCount}>
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
	<span class="label " bind:this={headIndex}>{pad(index + 1)}</span>
	<h2 class="set__name" bind:this={headName}>{set.name}</h2>
	<span class="label " bind:this={headCount}>{set.photos.length} photographs</span>
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
		justify-items: start;
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

	.card--out {
		overflow: visible;
	}

	.card--out .card__index,
	.card--out .card__name,
	.card--out .card__meta {
		position: relative;
		z-index: 6;
		transform-origin: left center;
		transform: translate3d(var(--dx, 0px), var(--dy, 0px), 0) scale(var(--ds, 1));
		opacity: 0;
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

	.set__head--waiting {
		opacity: 0;
	}

	.set__head--pinned {
		background: color-mix(in srgb, var(--bg) 94%, transparent);
		box-shadow: 0 1px 0 var(--line);
	}
</style>
