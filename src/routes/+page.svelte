<script lang="ts">
	import { onMount } from 'svelte';
	import Atmosphere from '$lib/components/Atmosphere.svelte';
	import GalleryGrid from '$lib/components/GalleryGrid.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import SetIndex from '$lib/components/SetIndex.svelte';
	import SplitText from '$lib/components/SplitText.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
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

		const gallery = document.querySelector('.gallery');
		if (!gallery) return;

		// Only travel when the gallery is still below the fold, so filtering from the
		// masthead while reading the grid does not yank the page out from under you.
		if (gallery.getBoundingClientRect().top > window.innerHeight * 0.5) {
			gallery.scrollIntoView({
				behavior: reducedMotion() ? 'auto' : 'smooth',
				block: 'start'
			});
		}
	}

	const latest = (photos_: Photo[]) =>
		photos_.reduce((newest, photo) => (photo.uploaded > newest ? photo.uploaded : newest), '').slice(0, 10);
</script>

<svelte:head>
	<title>Fahad Faruqi — Photographs</title>
	<meta
		name="description"
		content="Selected photographs by Fahad Faruqi, grouped by set, with the camera and exposure data behind each frame."
	/>
</svelte:head>

<header class="masthead">
	<div class="masthead__inner shell">
		<a class="masthead__mark" href="#top">Fahad Faruqi</a>
		<SetIndex sets={sets} active={activeSet} total={photos.length} onSelect={setFilter} />
		<ThemeToggle />
	</div>
</header>

<main id="top">
	<section class="hero shell">
		<h1 class="hero__name" aria-label="Fahad Faruqi">
			<SplitText text="Fahad Faruqi" />
		</h1>

		<div class="hero__meta">
			<p class="label">Photography</p>
			<p class="label">Nikon D3300 &middot; 18–55mm</p>
			<p class="label">Queens, New York</p>
			<p class="label num">
				{#if status === 'ready'}
					<span use:count={photos.length}></span> photographs &middot; {sets.length} sets
				{:else if status === 'loading'}
					Loading index
				{:else}
					Index unavailable
				{/if}
			</p>
		</div>
	</section>

	{#if status === 'error'}
		<section class="state shell">
			<p class="state__title">The index could not be loaded.</p>
			<p class="state__detail num">{failure}</p>
			<button class="state__retry label" type="button" onclick={() => load()}>Retry</button>
		</section>
	{:else if status === 'loading'}
		<section class="sets shell" aria-hidden="true">
			<div class="skeleton">
				{#each Array(6) as _, index (index)}
					<span class="skeleton__cell" style="--order: {index}"></span>
				{/each}
			</div>
		</section>
	{:else}
		<section class="sets shell" aria-label="Sets">
			<div class="sets__head">
				<span class="label">Sets</span>
				<span class="label">Photographs</span>
				<span class="label">Latest</span>
			</div>

			{#each sets as set, index (set.slug)}
				<button
					class="sets__row"
					class:sets__row--active={activeSet === set.slug}
					type="button"
					onclick={() => setFilter(activeSet === set.slug ? 'all' : set.slug)}
				>
					<span class="sets__name">
						<span class="label num">{String(index + 1).padStart(2, '0')}</span>
						{set.name}
					</span>
					<span class="num sets__count" use:count={set.photos.length}></span>
					<span class="num sets__date">{latest(set.photos)}</span>
				</button>
			{/each}
		</section>
	{/if}

	{#if status === 'ready' && visibleSets.length > 0}
		<GalleryGrid sets={visibleSets} {pass} onOpen={openViewer} />
	{/if}
</main>

<footer class="footer shell">
	<div class="footer__row">
		<span class="label">&copy; {new Date().getFullYear()} Fahad Faruqi</span>
		<span class="label">All photographs by Fahad Faruqi</span>
		<span class="label num">art.fahadfaruqi.com</span>
	</div>
</footer>

<Lightbox
	photos={visiblePhotos}
	index={viewerIndex}
	isOpen={viewerOpen}
	onClose={() => (viewerOpen = false)}
	onNavigate={(next) => (viewerIndex = next)}
/>

<Atmosphere />

<style>
	.masthead {
		position: sticky;
		top: 0;
		z-index: 20;
		background: color-mix(in srgb, var(--bg) 86%, transparent);
		backdrop-filter: blur(10px);
		border-bottom: 1px solid var(--line);
	}

	/* The glow lives in a fixed layer behind the page, so the page itself has to be
	   lifted above it for text and photographs to sit on the light, not under it. */
	main,
	.footer {
		position: relative;
		z-index: 1;
	}

	.masthead__inner {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
		padding-block: 0.85rem;
	}

	.masthead__mark {
		font-size: 0.8125rem;
		font-weight: 600;
		letter-spacing: 0.02em;
		white-space: nowrap;
	}

	.hero {
		padding-block: clamp(3rem, 12vh, 9rem) clamp(4rem, 14vh, 11rem);
	}

	.hero__name {
		margin: 0;
		font-size: var(--display);
		font-weight: 600;
		letter-spacing: -0.045em;
		line-height: 0.86;
		text-transform: uppercase;
	}

	.hero__meta {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.5rem 1.5rem;
		margin-top: clamp(2.5rem, 8vh, 6rem);
		padding-top: 1rem;
		border-top: 1px solid var(--line);
	}

	.hero__meta p {
		margin: 0;
	}

	.sets {
		padding-bottom: clamp(2rem, 6vh, 4rem);
	}

	.sets__head,
	.sets__row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 8rem 8rem;
		align-items: baseline;
		gap: 1rem;
		width: 100%;
		text-align: left;
	}

	.sets__head {
		padding-bottom: 0.5rem;
		border-bottom: 1px solid var(--line);
	}

	.sets__head .label:not(:first-child),
	.sets__count,
	.sets__date {
		text-align: right;
	}

	.sets__row {
		position: relative;
		padding-block: 0.9rem;
		border-bottom: 1px solid var(--line);
		transition: color 0.25s ease;
	}

	/* The rule under a row draws itself in from the left on hover, and stays drawn
	   while that set is the one on screen. */
	.sets__row::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: -1px;
		height: 1px;
		background: var(--fg);
		transform: scaleX(0);
		transform-origin: left center;
		transition: transform 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.sets__row:hover::after,
	.sets__row--active::after {
		transform: scaleX(1);
	}

	.sets__row:hover .sets__name,
	.sets__row--active .sets__name {
		color: var(--fg);
	}

	.sets__row--active .sets__name {
		font-weight: 500;
	}

	.sets__name {
		display: flex;
		align-items: baseline;
		gap: 1rem;
		color: var(--muted);
		font-size: clamp(1.125rem, 2.4vw, 1.75rem);
		letter-spacing: -0.02em;
		text-transform: lowercase;
		transition: color 0.25s ease;
	}

	.sets__count,
	.sets__date {
		font-size: 0.75rem;
		color: var(--muted);
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
	}

	.skeleton {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
		gap: clamp(0.75rem, 1.5vw, 1.75rem);
	}

	.skeleton__cell {
		aspect-ratio: 3 / 2;
		background: var(--bg-elev);
		animation: pulse 2.4s ease-in-out infinite;
		animation-delay: calc(var(--order) * 140ms);
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 0.4;
		}
		50% {
			opacity: 0.85;
		}
	}

	.footer {
		padding-block: clamp(2rem, 6vh, 4rem);
		border-top: 1px solid var(--line);
	}

	.footer__row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem 2rem;
		justify-content: space-between;
	}

	@media (max-width: 700px) {
		.sets__head {
			display: none;
		}

		.sets__row {
			grid-template-columns: minmax(0, 1fr) auto;
		}

		.sets__date {
			display: none;
		}
	}
</style>
