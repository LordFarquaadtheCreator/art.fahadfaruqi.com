<script lang="ts">
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

	const photo = $derived(photos[index]);
	const hasPrevious = $derived(index > 0);
	const hasNext = $derived(index < photos.length - 1);

	function previous() {
		if (hasPrevious) onNavigate(index - 1);
	}

	function next() {
		if (hasNext) onNavigate(index + 1);
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

	function handleTouchStart(event: TouchEvent) {
		touchStartX = event.changedTouches[0]?.clientX ?? null;
	}

	function handleTouchEnd(event: TouchEvent) {
		if (touchStartX === null) return;

		const delta = (event.changedTouches[0]?.clientX ?? touchStartX) - touchStartX;
		touchStartX = null;

		if (Math.abs(delta) < 45) return;
		if (delta > 0) previous();
		else next();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen && photo}
	<div class="viewer" role="dialog" aria-modal="true" aria-label="{photo.title} — image viewer">
		<button class="viewer__scrim" type="button" onclick={onClose} aria-label="Close viewer"></button>

		<div
			class="viewer__surface"
			bind:this={surface}
			tabindex="-1"
			role="group"
			aria-label="Photograph"
			ontouchstart={handleTouchStart}
			ontouchend={handleTouchEnd}
		>
			<figure class="viewer__figure">
				<img class="viewer__image" src={photo.display} alt={photo.alt} decoding="async" />
			</figure>

			<div class="viewer__panel">
				<div class="viewer__row">
					<span class="label num"
						>{String(photo.number).padStart(2, '0')} / {String(photos.length).padStart(2, '0')}</span
					>
					<span class="label">{photo.set}</span>
				</div>

				<h2 class="viewer__title">{photo.title}</h2>

				{#if photo.description}
					<p class="viewer__description">{photo.description}</p>
				{/if}

				<MetadataDisplay exif={photo.exif} />
			</div>
		</div>

		<div class="viewer__controls">
			<button
				class="viewer__control label"
				type="button"
				onclick={previous}
				disabled={!hasPrevious}
				aria-label="Previous photograph"
			>
				← Prev
			</button>
			<button
				class="viewer__control label"
				type="button"
				onclick={next}
				disabled={!hasNext}
				aria-label="Next photograph"
			>
				Next →
			</button>
			<button class="viewer__control label" type="button" onclick={onClose} aria-label="Close viewer">
				Close ✕
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
		animation: viewer-in 0.28s ease both;
	}

	.viewer__scrim {
		position: absolute;
		inset: 0;
		background: var(--scrim);
		backdrop-filter: blur(22px);
	}

	.viewer__surface {
		position: relative;
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(14rem, 20rem);
		gap: clamp(1rem, 3vw, 2.5rem);
		width: 100%;
		max-width: 100rem;
		max-height: 100%;
	}

	.viewer__figure {
		margin: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 0;
	}

	.viewer__image {
		max-width: 100%;
		max-height: min(84vh, 100rem);
		object-fit: contain;
		animation: image-in 0.5s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.viewer__panel {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		align-self: center;
		padding-bottom: 0.5rem;
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
		right: clamp(0.75rem, 3vw, 2.5rem);
		bottom: clamp(0.75rem, 2vw, 1.5rem);
		display: flex;
		justify-content: space-between;
		gap: 1rem;
	}

	.viewer__control {
		color: var(--muted);
		transition: color 0.2s ease;
	}

	.viewer__control:hover:not(:disabled) {
		color: var(--fg);
	}

	.viewer__control:disabled {
		opacity: 0.35;
		cursor: default;
	}

	@keyframes viewer-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes image-in {
		from {
			opacity: 0;
			transform: scale(0.985);
		}
		to {
			opacity: 1;
			transform: none;
		}
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
