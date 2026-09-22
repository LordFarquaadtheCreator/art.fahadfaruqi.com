<script lang="ts">
	import { fade } from 'svelte/transition';
	import type { Photo } from '$lib/utils/metadata';
	import MetadataDisplay from './MetadataDisplay.svelte';

	let {
		photos,
		index,
		isOpen,
		onClose,
		onNavigate
	}: {
		photos: Photo[];
		index: number;
		isOpen: boolean;
		onClose: () => void;
		onNavigate: (index: number) => void;
	} = $props();

	let surface = $state<HTMLElement | null>(null);
	let touchStartX = $state<number | null>(null);
	let drag = $state(0);
	let dragging = $state(false);
	let imageReady = $state(false);
	// Which way the viewer is travelling, so a swap slides in from the side you came from.
	let direction = $state(1);

	const photo = $derived(photos[index]);
	const hasPrevious = $derived(index > 0);
	const hasNext = $derived(index < photos.length - 1);

	const reduced =
		typeof window !== 'undefined' &&
		window.matchMedia('(prefers-reduced-motion: reduce)').matches;

	const fadeMs = (duration: number) => ({ duration: reduced ? 0 : duration });

	function previous() {
		if (!hasPrevious) return;
		direction = -1;
		onNavigate(index - 1);
	}

	function next() {
		if (!hasNext) return;
		direction = 1;
		onNavigate(index + 1);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!isOpen) return;

		if (event.key === 'Escape') onClose();
		else if (event.key === 'ArrowLeft') previous();
		else if (event.key === 'ArrowRight') next();
	}

	// Lock the page behind the viewer, focus the viewer on open, and warm the
	// neighbouring display images so arrow navigation is instant.
	$effect(() => {
		if (!isOpen) return;

		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = 'hidden';
		surface?.focus();

		for (const offset of [1, -1]) {
			const neighbour = photos[index + offset];
			if (neighbour) new Image().src = neighbour.display;
		}

		return () => {
			document.body.style.overflow = previousOverflow;
		};
	});

	// A new photograph starts unresolved, so its pixels fade in rather than pop.
	$effect(() => {
		photo?.key;
		imageReady = false;
	});

	function handleTouchStart(event: TouchEvent) {
		touchStartX = event.changedTouches[0]?.clientX ?? null;
		dragging = true;
	}

	// The photograph follows the finger; the release decides whether it travels.
	function handleTouchMove(event: TouchEvent) {
		if (touchStartX === null) return;
		const delta = (event.changedTouches[0]?.clientX ?? touchStartX) - touchStartX;
		drag = Math.max(-140, Math.min(140, delta));
	}

	function handleTouchEnd() {
		const travelled = drag;
		dragging = false;
		drag = 0;
		touchStartX = null;

		if (Math.abs(travelled) < 45) return;
		if (travelled > 0) previous();
		else next();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen && photo}
	<div
		class="viewer"
		role="dialog"
		aria-modal="true"
		aria-label="{photo.title} — image viewer"
		transition:fade={fadeMs(220)}
	>
		<button class="viewer__scrim" type="button" onclick={onClose} aria-label="Close viewer"></button>

		<!-- Field framing: brackets in the viewer's own margin, never over the photograph,
		     and the load status. An <img> reports no byte progress, so the status is the
		     word alone — there is no percentage to show and none is invented. -->
		<div class="viewer__hud" aria-hidden="true">
			<span class="viewer__bracket viewer__bracket--tl"></span>
			<span class="viewer__bracket viewer__bracket--tr"></span>
			<span class="viewer__bracket viewer__bracket--bl"></span>
			<span class="viewer__bracket viewer__bracket--br"></span>
			<span class="viewer__decode label" class:viewer__decode--done={imageReady}>Decoding</span>
		</div>

		<div
			class="viewer__surface"
			class:viewer__surface--dragging={dragging}
			bind:this={surface}
			tabindex="-1"
			role="group"
			aria-label="Photograph"
			style="--drag: {drag}px; --lean: {drag / 60}"
			ontouchstart={handleTouchStart}
			ontouchmove={handleTouchMove}
			ontouchend={handleTouchEnd}
		>
			<figure class="viewer__figure">
				{#key photo.key}
					<img
						class="viewer__lqip"
						class:viewer__lqip--faded={imageReady}
						src={photo.grid}
						alt=""
						aria-hidden="true"
						decoding="async"
					/>
					<img
						class="viewer__image"
						class:viewer__image--ready={imageReady}
						style="--from: {direction * 2.5}%"
						src={photo.display}
						alt={photo.alt}
						decoding="async"
						onload={() => (imageReady = true)}
						onerror={() => (imageReady = true)}
					/>
				{/key}
			</figure>

			{#key photo.key}
				<div class="viewer__panel">
					<div class="viewer__row" style="--i: 0">
						<span class="label num viewer__count"
							>Plate {String(index + 1).padStart(2, '0')} / {String(photos.length).padStart(2, '0')}</span
						>
						<span class="label">{photo.set}</span>
					</div>

					<h2 class="viewer__title" style="--i: 1">{photo.title}</h2>

					{#if photo.description}
						<p class="viewer__description" style="--i: 2">{photo.description}</p>
					{/if}

					<div class="viewer__exif" style="--i: 3">
						<MetadataDisplay exif={photo.exif} size={photo.size} />
					</div>
				</div>
			{/key}
		</div>

		<div class="viewer__controls">
			<button
				class="viewer__control label"
				type="button"
				onclick={previous}
				disabled={!hasPrevious}
				aria-label="Previous photograph"
			>
				<span class="viewer__arrow" aria-hidden="true">←</span>
				<span class="viewer__word">Prev</span>
			</button>
			<button
				class="viewer__control label"
				type="button"
				onclick={next}
				disabled={!hasNext}
				aria-label="Next photograph"
			>
				<span class="viewer__word">Next</span>
				<span class="viewer__arrow" aria-hidden="true">→</span>
			</button>
			<button class="viewer__control label" type="button" onclick={onClose} aria-label="Close viewer">
				<span class="viewer__word">Close</span>
				<span class="viewer__arrow" aria-hidden="true">✕</span>
			</button>
		</div>
	</div>
{/if}

<style>
	.viewer {
		position: fixed;
		inset: 0;
		z-index: 100;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: clamp(0.75rem, 3vw, 2.5rem);
	}

	.viewer__scrim {
		position: absolute;
		inset: 0;
		background: var(--scrim);
		backdrop-filter: blur(22px);
	}

	/* The HUD lives in the viewer's own margin — the same inset as the layout padding —
	   so the framing never draws over the photograph itself. */
	.viewer__hud {
		position: absolute;
		inset: clamp(0.75rem, 3vw, 2.5rem);
		pointer-events: none;
	}

	.viewer__bracket {
		position: absolute;
		width: clamp(0.75rem, 2vw, 1.5rem);
		height: clamp(0.75rem, 2vw, 1.5rem);
		border: 0 solid var(--accent-line);
	}

	.viewer__bracket--tl {
		top: 0;
		left: 0;
		border-top-width: 1px;
		border-left-width: 1px;
	}

	.viewer__bracket--tr {
		top: 0;
		right: 0;
		border-top-width: 1px;
		border-right-width: 1px;
	}

	.viewer__bracket--bl {
		bottom: 0;
		left: 0;
		border-bottom-width: 1px;
		border-left-width: 1px;
	}

	.viewer__bracket--br {
		bottom: 0;
		right: 0;
		border-bottom-width: 1px;
		border-right-width: 1px;
	}

	/* Status, not decoration: it says what is happening and nothing more. */
	.viewer__decode {
		position: absolute;
		left: calc(clamp(0.75rem, 2vw, 1.5rem) + 0.4rem);
		bottom: 0.15rem;
		color: var(--accent);
		opacity: 1;
		transition: opacity 0.35s ease;
	}

	.viewer__decode--done {
		opacity: 0;
	}

	/* Where you are in the run: live, so it is amber, like every other readout. */
	.viewer__count {
		color: var(--accent);
	}

	.viewer__surface {
		position: relative;
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(14rem, 20rem);
		gap: clamp(1rem, 3vw, 2.5rem);
		width: 100%;
		max-width: 100rem;
		max-height: 100%;
		transform: translate3d(var(--drag, 0), 0, 0);
		transition: transform 0.34s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	/* While the finger is down the surface tracks it exactly, tips with the gesture and
	   pulls in slightly, so the release reads as travel. */
	.viewer__surface--dragging {
		transition: none;
		transform: translate3d(var(--drag, 0), 0, 0) rotate(calc(var(--lean, 0) * 0.5deg))
			scale(0.985);
	}

	.viewer__figure {
		position: relative;
		margin: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 0;
	}

	/* Stands in while the 2200px file decodes. This is the grid's own derivative, not the
	   24px LQIP: the grid already fetched it, so it is usually decoded and waiting, and
	   at this size a 24px source would be an unrecognisable wash. */
	.viewer__lqip {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		filter: blur(5px);
		transform: scale(1.01);
		transition: opacity 0.45s ease;
	}

	.viewer__lqip--faded {
		opacity: 0;
	}

	.viewer__image {
		max-width: 100%;
		max-height: min(84vh, 100rem);
		object-fit: contain;
		opacity: 0;
		transition: opacity 0.32s ease;
		animation: image-swap 0.52s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.viewer__image--ready {
		opacity: 1;
	}

	/* Arrives from the side it was navigated from, and outruns nothing: the outgoing
	   photograph is gone, so the movement reads as travel rather than a cross-fade. */
	@keyframes image-swap {
		from {
			transform: translate3d(var(--from, 2.5%), 0, 0);
		}
		to {
			transform: translate3d(0, 0, 0);
		}
	}

	.viewer__panel {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		align-self: center;
		padding-bottom: 0.5rem;
	}

	/* The caption block catches up with the photograph, one line at a time. */
	.viewer__panel > * {
		animation: panel-in 0.42s cubic-bezier(0.16, 0.84, 0.28, 1) both;
		animation-delay: calc(var(--i, 0) * 55ms + 90ms);
	}

	@keyframes panel-in {
		from {
			opacity: 0;
			transform: translate3d(0, 0.5rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	.viewer__row {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		border-top: 1px solid var(--line);
		padding-top: 0.5rem;
	}

	.viewer__title {
		margin: 0;
		font-size: clamp(1.25rem, 2.4vw, 1.9rem);
		font-weight: 500;
		letter-spacing: -0.02em;
	}

	.viewer__description {
		margin: 0;
		font-size: 0.875rem;
		color: var(--muted);
		max-width: 34ch;
	}

	.viewer__controls {
		position: absolute;
		left: clamp(0.75rem, 3vw, 2.5rem);
		right: clamp(0.75rem, 2vw, 1.5rem);
		bottom: clamp(0.75rem, 2vw, 1.5rem);
		display: flex;
		justify-content: space-between;
		gap: 1rem;
	}

	.viewer__control {
		display: inline-flex;
		align-items: baseline;
		gap: 0.4rem;
		color: var(--muted);
		transition: color 0.2s ease;
	}

	.viewer__control:hover:not(:disabled) {
		color: var(--fg);
	}

	/* The arrow leans the way it is about to take you. */
	.viewer__arrow {
		display: inline-block;
		transition: transform 0.28s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.viewer__control:hover:not(:disabled) .viewer__arrow {
		transform: translate3d(-0.18rem, 0, 0);
	}

	.viewer__control:last-child:hover .viewer__arrow {
		transform: translate3d(0.18rem, 0, 0) rotate(45deg);
	}

	.viewer__control:disabled {
		opacity: 0.35;
		cursor: default;
	}

	@media (max-width: 900px) {
		.viewer__surface {
			grid-template-columns: minmax(0, 1fr);
			align-content: center;
			gap: 1rem;
		}

		.viewer__panel {
			align-self: start;
		}

		.viewer__image {
			max-height: 58vh;
		}
	}
</style>
