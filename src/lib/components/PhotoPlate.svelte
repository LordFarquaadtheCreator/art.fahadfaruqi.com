<script lang="ts">
	import { exifLine } from '$lib/utils/exif';
	import { cursor } from '$lib/utils/cursor.svelte';
	import { registerPlate } from '$lib/webgl/layer';
	import type { Photo } from '$lib/utils/metadata';

	let { photo, onOpen }: { photo: Photo; onOpen: (photo: Photo) => void } = $props();

	let loaded = $state(false);

	// From the listing, so the frame holds its shape before the image arrives.
	const metaRatio = $derived(
		photo.width > 0 && photo.height > 0 ? photo.width / photo.height : null
	);
	let measuredRatio = $state<number | null>(null);
	const ratio = $derived(metaRatio ?? measuredRatio);

	const DWELL_MS = 500;
	const WARM_MS = 300;

	let revealed = $state(false);
	let focused = $state(false);
	let timer: ReturnType<typeof setTimeout> | null = null;

	// The master is megabytes; it is fetched only once a pointer settles here.
	let masterWarmed = false;
	let warmTimer: ReturnType<typeof setTimeout> | null = null;

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

	function warmMaster() {
		if (masterWarmed) return;
		masterWarmed = true;
		const image = new Image();
		image.decoding = 'async';
		image.src = photo.master;
	}

	function warmSoon() {
		if (masterWarmed || warmTimer !== null) return;
		warmTimer = setTimeout(() => {
			warmTimer = null;
			warmMaster();
		}, WARM_MS);
	}

	function cancelWarm() {
		if (warmTimer !== null) {
			clearTimeout(warmTimer);
			warmTimer = null;
		}
	}

	$effect(() => () => {
		stopTimer();
		cancelWarm();
	});

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

	function carryIndex() {
		cursor.label = String(photo.number).padStart(2, '0');
		cursor.active = true;
	}

	function dropIndex() {
		cursor.active = false;
	}

	function measureRatio(event: Event) {
		if (ratio !== null) return;
		const image = event.currentTarget as HTMLImageElement;
		if (image.naturalWidth > 0 && image.naturalHeight > 0) {
			measuredRatio = image.naturalWidth / image.naturalHeight;
		}
	}

	const caption = $derived(exifLine(photo.exif));
	const portrait = $derived(ratio !== null && ratio < 1);
</script>

<figure
	class="plate"
	class:plate--portrait={portrait}
	style="--ratio: {ratio ?? 1.5}"
	use:dwellWhileVisible
>
	<div class="plate__media" class:plate__media--focused={focused}>
		<button
			class="plate__frame"
			class:plate__frame--loaded={loaded}
			type="button"
			use:registerPlate
			style={ratio ? `aspect-ratio: ${ratio}` : null}
			onclick={() => {
				disarm();
				onOpen(photo);
			}}
			onpointerenter={(event) => {
				carryIndex();
				if (canHover && event.pointerType === 'mouse') {
					arm();
					warmSoon();
				}
			}}
			onpointerleave={() => {
				dropIndex();
				disarm();
				cancelWarm();
			}}
			onfocus={() => {
				focused = true;
				carryIndex();
				arm(0);
				warmMaster();
			}}
			onblur={() => {
				focused = false;
				dropIndex();
				disarm();
				cancelWarm();
			}}
		>
			<img
				class="plate__image"
				src={photo.compressed}
				alt={photo.alt}
				loading="lazy"
				decoding="async"
				onload={(event) => {
					loaded = true;
					measureRatio(event);
				}}
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
		<span class="plate__number">{String(photo.number).padStart(2, '0')}</span>
		<span class="plate__title">{photo.title}</span>
		<span class="plate__exif">{caption}</span>
	</figcaption>
</figure>

<style>
	.plate {
		margin: 0;
		max-width: min(100%, calc(var(--ratio, 1.5) * 74vh));
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

	.plate__image {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		opacity: 0;
		transition:
			opacity 0.7s ease,
			transform 0.9s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.plate__frame--loaded .plate__image {
		opacity: 1;
	}

	.plate__frame:hover .plate__image,
	.plate__frame:focus-visible .plate__image {
		transform: scale(1.025);
	}

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
		font-size: var(--label-size);
		color: var(--faint);
		transition: color 0.3s ease;
	}

	.plate:hover .plate__number,
	.plate__media--focused + .plate__caption .plate__number {
		color: var(--accent);
	}

	.plate__title {
		font-size: 0.9375rem;
		font-weight: 500;
		letter-spacing: -0.01em;
	}

	.plate__exif {
		grid-column: 2;
		font-size: var(--label-size);
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
