<script lang="ts">
	let {
		text,
		segments = 16,
		delay = 0,
		duration = 1.15,
		stagger = 0.03
	}: { text: string; segments?: number; delay?: number; duration?: number; stagger?: number } =
		$props();

	const slices = $derived(
		Array.from({ length: segments }, (_, index) => ({
			index,
			left: (index / segments) * 100,
			width: 100 / segments + 0.01,
			from: `${(index % 2 === 0 ? -1 : 1) * (14 + index * 4)}%`,
			animationDelay: `${delay + index * stagger}s`
		}))
	);

	const geometry = $derived.by(() => {
		const share = (100 / segments + 0.01) / 100;
		return { step: 100 / segments / share, copyWidth: 100 / share };
	});
</script>

<span class="split title-attention" aria-hidden="true">
	<span class="split__spacer">{text}</span>
	{#each slices as slice (slice.index)}
		<span class="split__window" style="left: {slice.left}%; width: {slice.width}%">
			<span
				class="split__run"
				style="left: {slice.index * -geometry.step}%; width: {geometry.copyWidth}%; --from: {slice.from}; animation-delay: {slice.animationDelay}; animation-duration: {duration}s"
			>
				{text}
			</span>
		</span>
	{/each}
</span>

<style>
	.split {
		position: relative;
		display: block;
	}

	.split__spacer {
		visibility: hidden;
	}

	.split__window {
		position: absolute;
		top: 0;
		height: 100%;
		overflow: hidden;
	}

	.split__run {
		position: absolute;
		top: 0;
		display: block;
		will-change: transform;
		animation-name: assemble;
		animation-timing-function: cubic-bezier(0.16, 0.84, 0.28, 1);
		animation-fill-mode: both;
	}

	@keyframes assemble {
		from {
			transform: translate3d(var(--from), 0, 0);
		}
		to {
			transform: translate3d(0, 0, 0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.split__window {
			display: none;
		}

		.split__spacer {
			visibility: visible;
		}
	}
</style>
