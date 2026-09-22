<script lang="ts">
	import { exifLine } from '$lib/utils/exif';
	import { cursor } from '$lib/utils/cursor.svelte';
	import { registerPlate } from '$lib/webgl/layer';
	import type { Photo } from '$lib/utils/metadata';

	let { photo, onOpen }: { photo: Photo; onOpen: (photo: Photo) => void } = $props();

	let loaded = $state(false);
	let ratio = $state<number | null>(null);

	// The pointer carries the plate's number while it is over the grid, so you always
	// know where in the set you are without reading the caption.
	function carryIndex() {
		cursor.label = String(photo.number).padStart(2, '0');
		cursor.active = true;
	}

	function dropIndex() {
		cursor.active = false;
	}

	// The placeholder carries the photo's aspect ratio, so space is reserved before
	// the full image arrives and nothing shifts.
	function captureRatio(event: Event) {
		const image = event.currentTarget as HTMLImageElement;
		if (image.naturalWidth > 0 && image.naturalHeight > 0) {
			ratio = image.naturalWidth / image.naturalHeight;
		}
	}

	const caption = $derived(exifLine(photo.exif));
	const portrait = $derived(ratio !== null && ratio < 1);
</script>

<figure class="plate" class:plate--portrait={portrait}>
	<button
		class="plate__frame"
		class:plate__frame--loaded={loaded}
		type="button"
		use:registerPlate
		style={ratio
			? `aspect-ratio: ${ratio}; --ratio: ${ratio}`
			: `--ratio: ${3 / 2}`}
		onclick={() => onOpen(photo)}
		onpointerenter={carryIndex}
		onpointerleave={dropIndex}
		onfocus={carryIndex}
		onblur={dropIndex}
	>
		<img
			class="plate__lqip"
			src={photo.lqip}
			alt=""
			aria-hidden="true"
			decoding="async"
			onload={captureRatio}
		/>
		<img
			class="plate__image"
			src={photo.grid}
			srcset="{photo.grid} 800w, {photo.wide} 1600w, {photo.display} 2200w"
			sizes="(min-width: 1024px) 55vw, 100vw"
			alt={photo.alt}
			loading="lazy"
			decoding="async"
			onload={() => (loaded = true)}
		/>
	</button>

	<figcaption class="plate__caption">
		<span class="plate__number num">{String(photo.number).padStart(2, '0')}</span>
		<span class="plate__title">{photo.title}</span>
		<span class="plate__exif num">{caption}</span>
	</figcaption>
</figure>

<style>
	.plate {
		margin: 0;
	}

	.plate__frame {
		position: relative;
		display: block;
		width: 100%;
		aspect-ratio: 3 / 2;
		overflow: hidden;
		background: var(--bg-elev);
	}

	/* Below the 12-column breakpoint every plate is full-width, which makes a portrait
	   photograph several screens tall. Cap the width against the viewport height instead. */
	@media (max-width: 1023px) {
		.plate__frame {
			max-width: min(100%, calc(var(--ratio, 1.5) * 74vh));
		}
	}

	.plate__lqip,
	.plate__image {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.plate__lqip {
		filter: blur(14px);
		transform: scale(1.08);
	}

	.plate__image {
		opacity: 0;
		transition:
			opacity 0.7s ease,
			transform 0.9s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.plate__frame--loaded .plate__image {
		opacity: 1;
	}

	/* The quad draws over this image rather than replacing it. The image stays the
	   photograph: if the layer cannot start, cannot fetch a plate's bytes, or loses its
	   context, there is simply no effect — never a missing picture.

	   This scale is the hover affordance only while the layer is off (see the escape
	   hatch in app.css); with the layer running, the quad covers it and the zoom is
	   done in the shader's UV space instead. */
	.plate__frame:hover .plate__image,
	.plate__frame:focus-visible .plate__image {
		transform: scale(1.025);
	}

	.plate__caption {
		position: relative;
		display: grid;
		grid-template-columns: auto 1fr;
		column-gap: 0.75rem;
		align-items: baseline;
		margin-top: 0.625rem;
		padding-top: 0.5rem;
		border-top: 1px solid var(--line);
	}

	/* The caption's rule draws itself from the left as the pointer arrives, so the
	   hover state has a direction instead of just a colour change. */
	.plate__caption::after {
		content: '';
		position: absolute;
		top: -1px;
		left: 0;
		right: 0;
		height: 1px;
		background: var(--line-2);
		transform: scaleX(0);
		transform-origin: left center;
		transition: transform 0.5s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.plate:hover .plate__caption::after,
	.plate__frame:focus-visible + .plate__caption::after {
		transform: scaleX(1);
	}

	.plate__number {
		font-size: 0.6875rem;
		color: var(--faint);
	}

	.plate__title {
		font-size: 0.9375rem;
		font-weight: 500;
		letter-spacing: -0.01em;
	}

	.plate__exif {
		grid-column: 2;
		font-size: 0.6875rem;
		color: var(--faint);
		transition: color 0.3s ease;
	}

	.plate:hover .plate__exif,
	.plate__frame:focus-visible + .plate__caption .plate__exif {
		color: var(--muted);
	}
</style>
