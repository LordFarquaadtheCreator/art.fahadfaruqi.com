<script lang="ts">
	import { onMount } from 'svelte';
	import GalleryGrid from '$lib/components/GalleryGrid.svelte';
	import Hero from '$lib/components/Hero.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import SetIndex from '$lib/components/SetIndex.svelte';
	import { groupBySet } from '$lib/utils/group-images';
	import { fetchPhotos, type Photo } from '$lib/utils/metadata';

	type Status = 'loading' | 'ready' | 'error';

	// Matches the departures in this file's styles. The node goes on the timer either way, so
	// a stalled animation cannot leave a panel over the gallery.
	const DISSOLVE_MS = 1400;

	// What the band reports: the state of the read, never a number — the count is the hero's.
	const READOUT: Record<Status, string> = {
		loading: 'Receiving index',
		ready: 'Index received',
		error: 'Index unavailable'
	};

	let photos = $state<Photo[]>([]);
	let status = $state<Status>('loading');
	// The panel being replaced, held on screen while it dissolves. Every hand-off ends in the
	// gallery, so the gallery is never the one leaving.
	let leaving = $state<Exclude<Status, 'ready'> | null>(null);
	let leavingTimer: ReturnType<typeof setTimeout> | undefined;
	let failure = $state('');

	let activeSet = $state('all');
	let viewerOpen = $state(false);
	let viewerIndex = $state(0);

	const readout = $derived(READOUT[status]);
	const sets = $derived(groupBySet(photos));
	const visibleSets = $derived(
		activeSet === 'all' ? sets : sets.filter((set) => set.slug === activeSet)
	);
	const visiblePhotos = $derived(visibleSets.flatMap((set) => set.photos));

	onMount(() => {
		const controller = new AbortController();
		load(controller.signal);
		return () => {
			controller.abort();
			clearTimeout(leavingTimer);
		};
	});

	async function load(signal?: AbortSignal) {
		setStatus('loading');

		try {
			photos = await fetchPhotos(signal);
			setStatus('ready');
		} catch (error) {
			if (signal?.aborted) return;
			failure = error instanceof Error ? error.message : String(error);
			setStatus('error');
		}
	}

	// A panel hands over instead of cutting: the one being replaced stays on screen as a ghost
	// while it dissolves. Reduced motion gets none — there is no dissolve to watch.
	function setStatus(next: Status) {
		if (next === status) return;

		clearTimeout(leavingTimer);

		if (status !== 'ready' && !reducedMotion()) {
			leaving = status;
			leavingTimer = setTimeout(() => (leaving = null), DISSOLVE_MS);
		} else {
			leaving = null;
		}

		status = next;
	}

	function openViewer(photo: Photo) {
		const index = visiblePhotos.findIndex((candidate) => candidate.key === photo.key);
		viewerIndex = index === -1 ? 0 : index;
		viewerOpen = true;
	}

	let pass = $state(0);

	const reducedMotion = () =>
		typeof window !== 'undefined' &&
		window.matchMedia('(prefers-reduced-motion: reduce)').matches;

	// Every filter change bumps the pass, which restarts the gallery's wipe without
	// re-mounting a single plate — the photographs stay decoded and the WebGL layer
	// keeps the textures it already uploaded.
	function setFilter(slug: string) {
		if (slug === activeSet) return;

		activeSet = slug;
		pass += 1;

		const filters = document.querySelector('.filters');
		if (!filters) return;

		// Only travel when the filter is still below the fold, so re-filtering while
		// reading the grid does not yank the page out from under you.
		if (filters.getBoundingClientRect().top > window.innerHeight * 0.5) {
			filters.scrollIntoView({
				behavior: reducedMotion() ? 'auto' : 'smooth',
				block: 'start'
			});
		}
	}
</script>

<svelte:head>
	<title>Fahad's Art</title>
	<meta
		name="description"
		content="Selected photographs by Fahad Faruqi. All human-made."
	/>
</svelte:head>

<Hero {status} photoCount={photos.length} setCount={sets.length} />

<div class="stage">
	{#if status === 'error' || leaving === 'error'}
		<section class="state" class:leaving={leaving === 'error'} inert={leaving === 'error'}>
			<p class="state__title fade-in-animation">The index could not be loaded.</p>
			<p class="state__detail num fade-in-animation">{failure}</p>
			<button class="state__retry label fade-in-animation" type="button" onclick={() => load()}>Retry</button>
		</section>
	{/if}

	{#if status === 'loading' || leaving === 'loading'}
		<section
			class="loading"
			class:leaving={leaving === 'loading'}
			inert={leaving === 'loading'}
			role="status"
			aria-live="polite"
		>
			<span class="loading__hairline" aria-hidden="true"></span>
			<div class="loading__inner">
				<p class="label num loading__readout fade-in-animation">
					{readout}<span class="loading__dot" aria-hidden="true"></span>
				</p>
				<p class="label num loading__source fade-in-animation">assets.fahadfaruqi.com</p>
			</div>
			<div class="loading__frames" aria-hidden="true">
				{#each [0, 1, 2, 3] as frame (frame)}
					<span class="loading__frame" style="--i: {frame}"></span>
				{/each}
			</div>
		</section>
	{/if}

	{#if status === 'ready' && visibleSets.length > 0}
		<div class="stage__in">
			<SetIndex sets={sets} active={activeSet} total={photos.length} onSelect={setFilter} />

			<GalleryGrid sets={visibleSets} {pass} onOpen={openViewer} />
		</div>
	{/if}
</div>

<Lightbox
	photos={visiblePhotos}
	index={viewerIndex}
	isOpen={viewerOpen}
	onClose={() => (viewerOpen = false)}
	onNavigate={(next) => (viewerIndex = next)}
/>

<style>
	.stage {
		position: relative;
	}

	.state {
		padding-block: var(--block);
	}

	.state__title {
		margin: 0 0 0.5rem;
	}

	.state__detail {
		margin: 0 0 1.5rem;
		font-size: 0.75rem;
		color: var(--muted);
		overflow-wrap: anywhere;
	}

	.state__retry {
		border-bottom: 1px solid var(--line-2);
		padding-bottom: 0.15rem;
		transition:
			color 0.25s ease,
			border-color 0.25s ease,
			transform 0.3s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.state__retry:hover {
		color: var(--fg);
		border-bottom-color: var(--accent);
		transform: translate3d(0, -1px, 0);
	}

	.state__retry:active {
		transform: translate3d(0, 0, 0) scale(0.98);
	}

	.loading {
		position: relative;
		padding-block: 0 var(--block);
	}

	.loading__hairline {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 1px;
		background: var(--line);
		overflow: hidden;
	}

	.loading__hairline::after {
		content: '';
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		width: 28%;
		background: var(--accent);
		animation: loading-run 1.7s cubic-bezier(0.5, 0, 0.5, 1) infinite;
	}

	@keyframes loading-run {
		from {
			transform: translate3d(-100%, 0, 0);
		}
		to {
			transform: translate3d(460%, 0, 0);
		}
	}

	.loading__inner {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding-block: 1.4rem;
	}

	.loading__readout {
		color: var(--accent);
	}

	.loading__dot {
		display: inline-block;
		vertical-align: middle;
		width: 0.3rem;
		height: 0.3rem;
		margin-left: 0.5rem;
		border-radius: 50%;
		background: currentColor;
		animation: loading-pulse 1.4s ease-in-out infinite;
	}

	@keyframes loading-pulse {
		0%,
		100% {
			opacity: 0.2;
		}
		50% {
			opacity: 1;
		}
	}

	.loading__frames {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: clamp(0.75rem, 2vw, 2rem);
	}

	@media (min-width: 700px) {
		.loading__frames {
			grid-template-columns: repeat(4, minmax(0, 1fr));
		}
	}

	.loading__frame {
		position: relative;
		aspect-ratio: 3 / 2;
		overflow: hidden;
		background: var(--bg-elev);
		animation: meta-in 0.7s cubic-bezier(0.16, 0.84, 0.28, 1) both;
		animation-delay: calc(var(--i) * 80ms);
	}

	.loading__frame::after {
		content: '';
		position: absolute;
		inset: 0;
		background: linear-gradient(
			100deg,
			transparent 25%,
			color-mix(in srgb, var(--fg) 7%, transparent) 50%,
			transparent 75%
		);
		animation: loading-sheen 1.9s ease-in-out infinite;
		animation-delay: calc(var(--i) * 130ms);
	}

	@keyframes loading-sheen {
		from {
			transform: translate3d(-100%, 0, 0);
		}
		to {
			transform: translate3d(100%, 0, 0);
		}
	}

	/* Carries the strip above the grid; the grid's own stagger carries the sets. */
	.stage__in {
		animation: state-in 0.72s cubic-bezier(0.16, 0.84, 0.28, 1) 780ms both;
	}

	@keyframes state-in {
		from {
			opacity: 0;
			transform: translate3d(0, 1.25rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	/* The panel being replaced holds no space: it is out of flow in the same frame the next
	   state takes the room, so nothing below it moves. Its parts leave in the order they were
	   there for — the hairline drawn once (0–520ms), the report put away (520–1020ms), the
	   frames folded to their top edge (520–1230ms) — and it is clear by 1400ms, when the timer
	   drops it. Declared after the panels on purpose: the selectors tie on specificity, so
	   source order is what decides. */
	.leaving {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		pointer-events: none;
		animation: band-out 1.4s cubic-bezier(0.4, 0, 0.9, 0.55) forwards;
	}

	@keyframes band-out {
		from {
			opacity: 1;
			transform: translate3d(0, 0, 0);
		}
		to {
			opacity: 0;
			transform: translate3d(0, -2.5rem, 0);
		}
	}

	/* The report goes before the arriving nav reads under it. */
	.leaving .loading__inner,
	.leaving .loading__hairline {
		animation: strip-out 0.5s ease 520ms both;
	}

	@keyframes strip-out {
		to {
			opacity: 0;
		}
	}

	/* A new name, because an animation does not restart by itself. */
	.leaving .loading__hairline::after {
		width: 100%;
		animation: hairline-draw 0.52s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	@keyframes hairline-draw {
		from {
			transform: translate3d(-100%, 0, 0);
		}
		to {
			transform: none;
		}
	}

	.leaving .loading__frame {
		transform-origin: top;
		animation: frame-release 0.5s cubic-bezier(0.16, 0.84, 0.28, 1) both;
		animation-delay: calc(520ms + var(--i, 0) * 70ms);
	}

	@keyframes frame-release {
		from {
			transform: scaleY(1);
		}
		to {
			transform: scaleY(0);
		}
	}

	/* The dot is a pulse while something is being waited on. Nothing is, so it holds. */
	.leaving .loading__dot {
		animation: none;
		opacity: 1;
	}
</style>
