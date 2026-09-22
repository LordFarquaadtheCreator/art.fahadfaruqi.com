<script lang="ts">
	import type { Photo } from '$lib/utils/metadata';
	import type { PhotoSet } from '$lib/utils/group-images';

	let { set, index }: { set: PhotoSet; index: number } = $props();

	// The card replays every time the set comes back into view, which is why the class is
	// toggled and not left on: an animation only restarts when it is re-applied.
	let entered = $state(false);

	function onEnter(node: HTMLElement, notify: (value: boolean) => void) {
		if (typeof IntersectionObserver === 'undefined') {
			// Nothing to observe with: show the card rather than leave it waiting.
			notify(true);
			return { destroy: () => {} };
		}

		const observer = new IntersectionObserver(([entry]) => notify(entry.isIntersecting), {
			threshold: 0.4
		});
		observer.observe(node);

		return { destroy: () => observer.disconnect() };
	}

	const pad = (value: number) => String(value).padStart(2, '0');

	// The most recent capture date in the set, from the listing's own timestamps.
	const latest = (photos: Photo[]) =>
		photos.reduce((newest, photo) => (photo.uploaded > newest ? photo.uploaded : newest), '').slice(0, 10);
</script>

<!-- The header below already gives this set a heading and a count, so the card is a
     visual announcement of the same facts: nothing here is read out twice. -->
<section class="card" class:card--in={entered} aria-hidden="true" use:onEnter={(value) => (entered = value)}>
	<div class="card__inner shell">
		<span class="label num card__index">Set {pad(index + 1)}</span>
		<h3 class="card__name">{set.name}</h3>
		<span class="label num card__meta">
			{set.photos.length} plates &middot; {latest(set.photos)}
		</span>
	</div>
	<span class="card__rule"></span>
</section>

<style>
	.card {
		position: relative;
		display: flex;
		align-items: flex-end;
		min-height: clamp(8rem, 24vh, 15rem);
		padding-block: clamp(1.5rem, 6vh, 4rem) clamp(1rem, 3vh, 2rem);
		overflow: hidden;
	}

	.card__inner {
		display: grid;
		gap: 0.5rem;
		width: 100%;
	}

	.card__name {
		margin: 0;
		font-size: clamp(2.25rem, 8vw, 6rem);
		font-weight: 600;
		line-height: 0.95;
		letter-spacing: -0.03em;
		text-transform: uppercase;
	}

	.card__index,
	.card__meta {
		color: var(--muted);
	}

	/* The rule is the projection: it draws itself across the band as the set arrives.
	   Its default is drawn, so a visitor whose animations never run still sees it. */
	.card__rule {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 1px;
		background: var(--accent);
		transform: scaleX(1);
		transform-origin: left center;
	}

	.card--in .card__index,
	.card--in .card__name,
	.card--in .card__meta {
		animation: card-in 0.75s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.card--in .card__name {
		animation-delay: 0.06s;
	}

	.card--in .card__meta {
		animation-delay: 0.12s;
	}

	.card--in .card__rule {
		animation: card-rule 0.9s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	/* The hidden state lives inside the keyframes, so a bundle that never runs leaves the
	   card visible rather than blank. */
	@keyframes card-in {
		from {
			opacity: 0;
			transform: translate3d(0, 1.1rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	@keyframes card-rule {
		from {
			transform: scaleX(0);
		}
		to {
			transform: scaleX(1);
		}
	}
</style>
