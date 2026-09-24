<script lang="ts">
	import { reveal } from '$lib/actions/reveal';
	import { registerParallax } from '$lib/utils/parallax';
	import type { PhotoSet } from '$lib/utils/group-images';
	import type { Photo } from '$lib/utils/metadata';
	import PhotoPlate from './PhotoPlate.svelte';
	import SetCard from './SetCard.svelte';

	let {
		sets,
		pass = 0,
		onOpen
	}: { sets: PhotoSet[]; pass?: number; onOpen: (photo: Photo, set: PhotoSet) => void } = $props();

	// Two identical entrances alternating by pass, because a CSS animation only restarts
	// when its name changes — this is what re-plays the wipe on every filter change
	// without re-mounting a single plate.
	const wipe = $derived(pass % 2 === 0 ? 'a' : 'b');
</script>

<div class="gallery" data-pass={wipe}>
	{#each sets as set, setIndex (set.slug)}
		<section class="set" id={set.slug} style="--set: {setIndex}">
			<SetCard {set} index={setIndex} />

			<div class="set__grid shell">
				{#each set.photos as photo, index (photo.key)}
					<div
						class="cell cell--{index % 8}"
						style="--order: {Math.min(index, 8)}"
						use:reveal
					>
						<div class="cell__slide" use:registerParallax>
							<PhotoPlate {photo} onOpen={(opened) => onOpen(opened, set)} />
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/each}
</div>

<style>
	.gallery {
		padding-block: clamp(2rem, 6vh, 5rem) var(--block);
		/* The header is sticky, so a scroll to the grid has to clear it. */
		scroll-margin-top: calc(var(--header-h) + 0.5rem);
	}

	/* Filtering re-plays the incoming sets, one after another. The two names are
	   identical on purpose; see the wipe derivation in the script. */
	.gallery[data-pass='a'] > .set {
		animation: set-in-a 0.62s cubic-bezier(0.16, 0.84, 0.28, 1) both;
		animation-delay: calc(var(--set, 0) * 80ms);
	}

	.gallery[data-pass='b'] > .set {
		animation: set-in-b 0.62s cubic-bezier(0.16, 0.84, 0.28, 1) both;
		animation-delay: calc(var(--set, 0) * 80ms);
	}

	@keyframes set-in-a {
		from {
			opacity: 0;
			transform: translate3d(0, 1.25rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	@keyframes set-in-b {
		from {
			opacity: 0;
			transform: translate3d(0, 1.25rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	.set + .set {
		margin-top: var(--block);
	}

	/* The parallax carrier. Keeping the drift on a wrapper means the cell's own
	   entrance transition and the scroll offset never fight over `transform`. */
	.cell__slide {
		transform: translate3d(0, var(--parallax, 0), 0);
	}

	.set__name {
		flex: 1;
		margin: 0;
		font-size: clamp(1.5rem, 3.2vw, 2.75rem);
		font-weight: 500;
		letter-spacing: -0.03em;
		text-transform: lowercase;
	}

	.set__grid {
		display: grid;
		grid-template-columns: repeat(12, minmax(0, 1fr));
		gap: clamp(0.75rem, 1.5vw, 1.75rem);
		row-gap: 0;
	}

	.cell {
		grid-column: 1 / -1;
		opacity: 1;
		transition:
			opacity 0.9s ease,
			transform 0.9s cubic-bezier(0.16, 0.84, 0.28, 1);
		transition-delay: calc(var(--order) * 70ms);
	}

	/* Set by the reveal action at runtime, so the compiler cannot see the attribute.
	   Only a cell the script has explicitly claimed is hidden — if the script never
	   runs, or dies, every photograph is on screen instead of none. */
	:global(.cell[data-revealed='false']) {
		opacity: 0;
		transform: translate3d(0, 1.75rem, 0);
	}

	.cell + .cell {
		margin-top: clamp(2rem, 6vh, 4.5rem);
	}

	/* Offset, asymmetric placement — the Swiss-print grid the reference is built on.
	   Explicit column starts with auto rows leave deliberate holes. */
	@media (min-width: 1024px) {
		.cell--0 {
			grid-column: 1 / span 7;
		}

		.cell--1 {
			grid-column: 9 / span 4;
			margin-top: clamp(2rem, 8vh, 6rem);
		}

		.cell--2 {
			grid-column: 3 / span 5;
		}

		.cell--3 {
			grid-column: 9 / span 4;
		}

		.cell--4 {
			grid-column: 1 / span 4;
			margin-top: clamp(1rem, 5vh, 4rem);
		}

		.cell--5 {
			grid-column: 6 / span 7;
		}

		.cell--6 {
			grid-column: 2 / span 6;
			margin-top: clamp(2rem, 9vh, 7rem);
		}

		.cell--7 {
			grid-column: 9 / span 4;
		}

		/* A portrait frame in a wide span would tower over its neighbours, so those
		   plates get a narrow column and sit within the rhythm instead. */
		.cell:has(:global(.plate--portrait)) {
			grid-column: 3 / span 4;
			justify-self: start;
			width: 100%;
		}

		.cell:nth-child(even):has(:global(.plate--portrait)) {
			grid-column: 8 / span 4;
		}
	}
</style>
