<script lang="ts">
	import { onMount } from 'svelte';
	import Atmosphere from '$lib/components/Atmosphere.svelte';
	import GalleryGrid from '$lib/components/GalleryGrid.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import SetIndex from '$lib/components/SetIndex.svelte';
	import SplitText from '$lib/components/SplitText.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { count } from '$lib/actions/count';
	import { reveal } from '$lib/actions/reveal';
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
	let scrolled = $state(false);
	let progress = $state(0);

	// The masthead's scroll state, as component state rather than an attribute an action
	// sets: Svelte can see a bound class, so it keeps the rules and scopes them. The read
	// is a plain pair of numbers (neither forces layout while scrolling), and Svelte
	// batches the updates, so this stays cheaper than deferring to a frame.
	$effect(() => {
		const read = () => {
			const y = window.scrollY;
			const span = document.documentElement.scrollHeight - window.innerHeight;
			scrolled = y > 24;
			progress = span > 0 ? Math.min(1, y / span) : 0;
		};

		read();
		window.addEventListener('scroll', read, { passive: true });
		window.addEventListener('resize', read, { passive: true });

		return () => {
			window.removeEventListener('scroll', read);
			window.removeEventListener('resize', read);
		};
	});

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

<header class="masthead" class:masthead--scrolled={scrolled}>
	<div class="masthead__inner shell">
		<a class="masthead__mark" href="#top">Fahad Faruqi</a>
		<SetIndex sets={sets} active={activeSet} total={photos.length} onSelect={setFilter} />
		<ThemeToggle />
	</div>
	<span class="masthead__progress" aria-hidden="true" style="--progress: {progress}"></span>
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
		<section class="sets shell" aria-label="Sets" use:reveal>
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

	{#if status === 'ready' && sets.length > 0}
		<section class="marquee" aria-hidden="true">
			<div class="marquee__track">
				{#each [0, 1] as run (run)}
					<div class="marquee__run">
						<span class="marquee__word">Fahad Faruqi</span>
						<span class="marquee__dot"></span>
						<span class="marquee__word">Photographs</span>
						<span class="marquee__dot"></span>
						<span class="marquee__word">Queens, New York</span>
						<span class="marquee__dot"></span>
						<span class="marquee__word">{photos.length} Photographs</span>
						<span class="marquee__dot"></span>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	{#if status === 'ready' && visibleSets.length > 0}
		<GalleryGrid sets={visibleSets} {pass} onOpen={openViewer} />
	{/if}
</main>

<footer class="footer shell" use:reveal>
	<div class="footer__row">
		<span class="label">&copy; {new Date().getFullYear()} Fahad Faruqi</span>
		<span class="label">All photographs by Fahad Faruqi</span>
		<span class="label num">art.fahadfaruqi.com</span>
	</div>
</footer>

<button
	class="to-top label"
	class:to-top--visible={progress > 0.08}
	type="button"
	tabindex={progress > 0.08 ? 0 : -1}
	onclick={() => window.scrollTo({ top: 0, behavior: reducedMotion() ? 'auto' : 'smooth' })}
>
	Top <span class="to-top__arrow" aria-hidden="true">↑</span>
</button>

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
		transition:
			background-color 0.45s ease,
			border-color 0.45s ease,
			backdrop-filter 0.45s ease;
	}

	/* Scrolled, the bar sinks a little: tighter, dimmer, and with a firmer rule, so
	   the page reads as having moved rather than the bar just sticking. */
	.masthead--scrolled {
		background: color-mix(in srgb, var(--bg) 72%, transparent);
		backdrop-filter: blur(18px);
		border-bottom-color: var(--line-2);
	}

	/* The glow lives in a fixed layer behind the page, so the page itself has to be
	   lifted above it for text and photographs to sit on the light, not under it. */
	main,
	.footer {
		position: relative;
		z-index: 1;
	}

	.masthead__inner {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
		padding-block: 0.85rem;
		transition: padding 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.masthead--scrolled .masthead__inner {
		padding-block: 0.6rem;
	}

	/* One hairline of how far through the document you are. */
	.masthead__progress {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 1px;
		background: var(--line-2);
		transform: scaleX(var(--progress, 0));
		transform-origin: left center;
	}

	.masthead__inner > * {
		animation: masthead-in 0.7s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.masthead__inner > :nth-child(1) {
		animation-delay: 0.04s;
	}

	.masthead__inner > :nth-child(2) {
		animation-delay: 0.11s;
	}

	.masthead__inner > :nth-child(3) {
		animation-delay: 0.18s;
	}

	@keyframes masthead-in {
		from {
			opacity: 0;
			transform: translate3d(0, -0.5rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	.masthead__mark {
		position: relative;
		font-size: 0.8125rem;
		font-weight: 600;
		letter-spacing: 0.02em;
		white-space: nowrap;
	}

	/* The only link in the masthead draws its own underline on approach. */
	.masthead__mark::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: -0.15rem;
		height: 1px;
		background: var(--line-2);
		transform: scaleX(0);
		transform-origin: left center;
		transition: transform 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.masthead__mark:hover::after,
	.masthead__mark:focus-visible::after {
		transform: scaleX(1);
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

	/* The strip catches up with the headline rather than appearing with it. */
	.hero__meta > * {
		animation: meta-in 0.8s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.hero__meta > :nth-child(1) {
		animation-delay: 0.24s;
	}

	.hero__meta > :nth-child(2) {
		animation-delay: 0.32s;
	}

	.hero__meta > :nth-child(3) {
		animation-delay: 0.4s;
	}

	.hero__meta > :nth-child(4) {
		animation-delay: 0.48s;
	}

	@keyframes meta-in {
		from {
			opacity: 0;
			transform: translate3d(0, 0.6rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	.sets {
		padding-bottom: clamp(2rem, 6vh, 4rem);
	}

	/* The table arrives row by row, whether it is the first paint or the skeleton
	   being replaced. */
	.sets__row {
		animation: row-in 0.6s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.sets__row:nth-child(2) {
		animation-delay: 0.09s;
	}

	.sets__row:nth-child(3) {
		animation-delay: 0.18s;
	}

	.sets__row:nth-child(4) {
		animation-delay: 0.27s;
	}

	.sets__row:nth-child(n + 5) {
		animation-delay: 0.36s;
	}

	@keyframes row-in {
		from {
			opacity: 0;
			transform: translate3d(0, 0.75rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
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

	/* The error block arrives the same way everything else does. */
	.state > * {
		animation: meta-in 0.7s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.state > :nth-child(2) {
		animation-delay: 0.08s;
	}

	.state > :nth-child(3) {
		animation-delay: 0.16s;
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
		border-bottom-color: var(--fg);
		transform: translate3d(0, -1px, 0);
	}

	.state__retry:active {
		transform: translate3d(0, 0, 0) scale(0.98);
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
		transition:
			opacity 0.9s ease,
			transform 0.9s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	/* Set by the reveal action; see the note on .cell in GalleryGrid. */
	:global(.footer[data-revealed='false']) {
		opacity: 0;
		transform: translate3d(0, 1rem, 0);
	}

	.footer__row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem 2rem;
		justify-content: space-between;
	}

	/* Off the top of the document, and only then. Hidden it is not focusable, so it
	   never appears in the tab order as an invisible stop. */
	.to-top {
		position: fixed;
		right: var(--gutter);
		bottom: clamp(1rem, 3vh, 2rem);
		z-index: 15;
		display: inline-flex;
		align-items: baseline;
		gap: 0.35rem;
		padding: 0.4rem 0.7rem;
		border: 1px solid var(--line);
		background: color-mix(in srgb, var(--bg) 80%, transparent);
		backdrop-filter: blur(10px);
		color: var(--muted);
		opacity: 0;
		transform: translate3d(0, 0.75rem, 0);
		pointer-events: none;
		transition:
			opacity 0.4s ease,
			transform 0.5s cubic-bezier(0.16, 0.84, 0.28, 1),
			color 0.25s ease,
			border-color 0.25s ease;
	}

	.to-top--visible {
		opacity: 1;
		transform: none;
		pointer-events: auto;
	}

	.to-top:hover {
		color: var(--fg);
		border-color: var(--line-2);
	}

	.to-top__arrow {
		display: inline-block;
		transition: transform 0.35s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.to-top:hover .to-top__arrow {
		transform: translate3d(0, -0.2rem, 0);
	}

	/* A band of type running under the index, so the page announces itself once between
	   the list of sets and the photographs. Two identical runs translated by exactly half
	   the track make the loop seamless; hover holds it still. */
	.marquee {
		position: relative;
		overflow: hidden;
		margin-top: clamp(2rem, 8vh, 6rem);
		padding-block: clamp(0.75rem, 2vh, 1.5rem);
		border-block: 1px solid var(--line);
		-webkit-mask-image: linear-gradient(90deg, transparent, #000 7%, #000 93%, transparent);
		mask-image: linear-gradient(90deg, transparent, #000 7%, #000 93%, transparent);
	}

	.marquee__track {
		display: flex;
		width: max-content;
		animation: marquee-run 44s linear infinite;
	}

	.marquee:hover .marquee__track {
		animation-play-state: paused;
	}

	.marquee__run {
		display: flex;
		align-items: center;
		gap: clamp(1.5rem, 4vw, 3.5rem);
		padding-right: clamp(1.5rem, 4vw, 3.5rem);
	}

	.marquee__word {
		font-size: clamp(1.75rem, 5.5vw, 4.5rem);
		font-weight: 600;
		letter-spacing: -0.03em;
		line-height: 1;
		text-transform: uppercase;
		white-space: nowrap;
		color: var(--muted);
		transition: color 0.45s ease;
	}

	.marquee:hover .marquee__word {
		color: var(--fg);
	}

	.marquee__dot {
		flex: none;
		width: 0.4rem;
		height: 0.4rem;
		border-radius: 50%;
		background: var(--line-2);
	}

	@keyframes marquee-run {
		from {
			transform: translate3d(0, 0, 0);
		}
		to {
			transform: translate3d(-50%, 0, 0);
		}
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
