<script lang="ts">
	let { progress }: { progress: number } = $props();

	const reducedMotion = () =>
		typeof window !== 'undefined' &&
		window.matchMedia('(prefers-reduced-motion: reduce)').matches;

	const visible = $derived(progress > 0.08);
</script>

<button
	class="to-top label"
	class:to-top--visible={visible}
	type="button"
	tabindex={visible ? 0 : -1}
	onclick={() => window.scrollTo({ top: 0, behavior: reducedMotion() ? 'auto' : 'smooth' })}
>
	Top <span class="to-top__arrow" aria-hidden="true">↑</span>
</button>

<style>
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
		border-color: var(--accent);
	}

	.to-top__arrow {
		display: inline-block;
		transition: transform 0.35s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.to-top:hover .to-top__arrow {
		transform: translate3d(0, -0.2rem, 0);
	}
</style>
