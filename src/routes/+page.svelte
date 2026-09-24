<script lang="ts">
	import { onMount } from 'svelte';
	import GalleryGrid from '$lib/components/GalleryGrid.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import SetIndex from '$lib/components/SetIndex.svelte';
	import SplitText from '$lib/components/SplitText.svelte';
	import { count } from '$lib/actions/count';
	import { groupBySet } from '$lib/utils/group-images';
	import { fetchPhotos, type Photo } from '$lib/utils/metadata';

	let photos = $state<Photo[]>([]);
	let status = $state<'loading' | 'ready' | 'error'>('loading');
	let failure = $state('');

	let activeSet = $state('all');
	let viewerOpen = $state(false);
	let viewerIndex = $state(0);

	const sets = $derived(groupBySet(photos));
	const visibleSets = $derived(
		activeSet === 'all' ? sets : sets.filter((set) => set.slug === activeSet)
	);
	const visiblePhotos = $derived(visibleSets.flatMap((set) => set.photos));

	onMount(() => {
		const controller = new AbortController();
		load(controller.signal);
		return () => controller.abort();
	});

	async function load(signal?: AbortSignal) {
		status = 'loading';

		try {
			photos = await fetchPhotos(signal);
			status = 'ready';
		} catch (error) {
			if (signal?.aborted) return;
			failure = error instanceof Error ? error.message : String(error);
			status = 'error';
		}
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

<!-- Hero -->
<section class="hero shell">
	<h1 class="title large" aria-label="Fahad Faruqi">
		<SplitText text="Fahad Faruqi" />
	</h1>

	<div class="hairline hero__meta">
		<p class="label fade-in-animation">Photography</p>
		<p class="label fade-in-animation">Nikon D3300</p>
		<p class="label fade-in-animation">Queens, New York</p>
		<p class="label num fade-in-animation">
			{#if status === 'ready'}
				<span use:count={photos.length}></span> photographs &middot; {sets.length} sets
			{:else if status === 'loading'}
				Loading photos
			{:else}
				Photos unavailable
			{/if}
		</p>
	</div>
</section>

<!-- Error State -->
{#if status === 'error'}
	<section class="state shell">
		<p class="state__title fade-in-animation">The index could not be loaded.</p>
		<p class="state__detail num fade-in-animation">{failure}</p>
		<button class="state__retry label fade-in-animation" type="button" onclick={() => load()}>Retry</button>
	</section>
{/if}

<!-- Loading State -->
{#if status === 'loading'}
	<section class="loading" role="status" aria-live="polite">
		<span class="loading__hairline" aria-hidden="true"></span>
		<div class="loading__inner shell">
			<p class="label num loading__readout">
				Receiving index<span class="loading__dot" aria-hidden="true"></span>
			</p>
			<p class="label num loading__source">assets.fahadfaruqi.com</p>
		</div>
		<div class="loading__frames shell" aria-hidden="true">
			{#each [0, 1, 2, 3] as frame (frame)}
				<span class="loading__frame" style="--i: {frame}"></span>
			{/each}
		</div>
	</section>
{/if}

<!-- Images -->
{#if status === 'ready' && visibleSets.length > 0}
    <SetIndex sets={sets} active={activeSet} total={photos.length} onSelect={setFilter} />

    <GalleryGrid sets={visibleSets} {pass} onOpen={openViewer} />
{/if}

<!-- Individual Photo Lightbox -->
<Lightbox
	photos={visiblePhotos}
	index={viewerIndex}
	isOpen={viewerOpen}
	onClose={() => (viewerOpen = false)}
	onNavigate={(next) => (viewerIndex = next)}
/>

<style>
	.hero {
		padding-block: clamp(3rem, 12vh, 9rem) clamp(1rem, 3vh, 3rem);
	}

	.hero__meta {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.5rem 1.5rem;
	}

	.hero__meta p {
		margin: 0;
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

	/* ---------------------------------------------------------------- the wait */

	.loading {
		position: relative;
		padding-block: 0 var(--block);
	}

	/* Indeterminate on purpose: the API reports no progress, so the line travels rather
	   than filling. With reduced motion it stops travelling and simply sits there. */
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

	/* --------------------------------------------------------------- the filter */

	.filters {
		border-bottom: 1px solid var(--line);
		scroll-margin-top: calc(var(--header-h) + 0.5rem);
	}


</style>
