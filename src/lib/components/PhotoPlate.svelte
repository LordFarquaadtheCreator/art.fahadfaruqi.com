<script lang="ts">
	import { exifLine } from '$lib/utils/exif';
	import { cursor } from '$lib/utils/cursor.svelte';
	import { registerPlate } from '$lib/webgl/layer';
	import type { Photo } from '$lib/utils/metadata';

	let { photo, onOpen }: { photo: Photo; onOpen: (photo: Photo) => void } = $props();

	let loaded = $state(false);
	let ratio = $state<number | null>(null);

	// The delay is the feature: pass over a plate and nothing happens, rest on it and the
	// note arrives. Nothing about the photograph is behind it — the caption below still
	// carries the title, and the viewer carries everything — so this is an addition to
	// what is already readable, never the only way to read it.
	const DWELL_MS = 2000;

	let revealed = $state(false);
	// The keyboard's focus state is component state, not something a selector decides:
	// `:focus-visible` cannot be verified while the page is not the focused window, and the
	// caption's readout has to answer to the keyboard the same way it answers to a pointer.
	let focused = $state(false);
	let timer: ReturnType<typeof setTimeout> | null = null;

	// Hover is a desktop affordance. On a touch device there is no pointer to rest, so the
	// equivalent of hovering is dwelling on the plate while it sits in view.
	const canHover =
		typeof window !== 'undefined' &&
		window.matchMedia('(hover: hover) and (pointer: fine)').matches;

	function stopTimer() {
		if (timer !== null) {
			clearTimeout(timer);
			timer = null;
		}
	}

	function arm(delay = DWELL_MS) {
		if (!photo.description) return;
		stopTimer();
		timer = setTimeout(() => (revealed = true), delay);
	}

	function disarm() {
		stopTimer();
		revealed = false;
	}

	$effect(() => stopTimer);

	/** Touch path: reveal once the plate has been sitting in view for the same dwell. */
	function dwellWhileVisible(node: HTMLElement) {
		if (canHover || typeof IntersectionObserver === 'undefined') return;

		const observer = new IntersectionObserver(
			([entry]) => {
				if (entry.isIntersecting) arm();
				else disarm();
			},
			{ threshold: 0.65 }
		);
		observer.observe(node);

		return {
			destroy() {
				observer.disconnect();
				stopTimer();
			}
		};
	}

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

<figure class="plate" class:plate--portrait={portrait} use:dwellWhileVisible>
	<div class="plate__media" class:plate__media--focused={focused}>
		<button
			class="plate__frame"
			class:plate__frame--loaded={loaded}
			type="button"
			use:registerPlate
			style={ratio
				? `aspect-ratio: ${ratio}; --ratio: ${ratio}`
				: `--ratio: ${3 / 2}`}
			onclick={() => {
				disarm();
				onOpen(photo);
			}}
			onpointerenter={(event) => {
				carryIndex();
				if (canHover && event.pointerType === 'mouse') arm();
			}}
			onpointerleave={() => {
				dropIndex();
				disarm();
			}}
			onfocus={() => {
				focused = true;
				carryIndex();
				// A keyboard visitor has already committed to the plate; there is nothing
				// to wait for, so the note arrives at once.
				arm(0);
			}}
			onblur={() => {
				focused = false;
				dropIndex();
				disarm();
			}}
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

		{#if photo.description}
			<div class="plate__note" class:plate__note--in={revealed}>
				<span class="plate__note-title">{photo.title}</span>
				<span class="plate__note-text">{photo.description}</span>
			</div>
		{/if}
	</div>

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

	.plate__media {
		position: relative;
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
	   photograph several screens tall. Cap the width against the viewport height instead.
	   The cap belongs on the wrapper, so the note stays inside the photograph's edges. */
	@media (max-width: 1023px) {
		.plate__media {
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

	/* The note lands on the photograph's lower edge, over a scrim, rather than in the
	   caption below it: the caption sits inside the grid, so growing it would shove every
	   row beneath it down and shift the page while it is being read. */
	.plate__note {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		display: grid;
		gap: 0.3rem;
		padding: 3.5rem 1rem 1rem;
		background: linear-gradient(
			to top,
			rgba(8, 8, 8, 0.93),
			rgba(8, 8, 8, 0.7) 55%,
			rgba(8, 8, 8, 0)
		);
		color: #f2f1ee;
		opacity: 0;
		transform: translateY(0.5rem);
		transition:
			opacity 0.45s ease,
			transform 0.55s cubic-bezier(0.16, 0.84, 0.28, 1);
		pointer-events: none;
	}

	.plate__note--in {
		opacity: 1;
		transform: translateY(0);
	}

	.plate__note-title {
		font-size: 0.9375rem;
		font-weight: 500;
		letter-spacing: -0.01em;
	}

	.plate__note-text {
		font-size: 0.8125rem;
		line-height: 1.5;
		color: rgba(242, 241, 238, 0.84);
		text-wrap: pretty;
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
	   hover state has a direction instead of just a colour change. It draws in amber:
	   this is the active plate. */
	.plate__caption::after {
		content: '';
		position: absolute;
		top: -1px;
		left: 0;
		right: 0;
		height: 1px;
		background: var(--accent);
		transform: scaleX(0);
		transform-origin: left center;
		transition: transform 0.5s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.plate:hover .plate__caption::after,
	.plate__media--focused + .plate__caption::after {
		transform: scaleX(1);
	}

	.plate__number {
		font-size: 0.6875rem;
		color: var(--faint);
		transition: color 0.3s ease;
	}

	/* The plate's number under the pointer is a live readout, so it is one of the few
	   pieces of text that goes amber. */
	.plate:hover .plate__number,
	.plate__media--focused + .plate__caption .plate__number {
		color: var(--accent);
	}

	.plate__title {
		font-size: 0.9375rem;
		font-weight: 500;
		letter-spacing: -0.01em;
	}

	/* The caption's readout is set as a field instrument reports it: capitalised, with
	   the units attached and the figures tabular so the columns line up between plates. */
	.plate__exif {
		grid-column: 2;
		font-size: 0.6875rem;
		text-transform: uppercase;
		letter-spacing: 0.02em;
		color: var(--faint);
		transition: color 0.3s ease;
	}

	.plate:hover .plate__exif,
	.plate__media--focused + .plate__caption .plate__exif {
		color: var(--muted);
	}
</style>
