<script lang="ts">
	import { count } from '$lib/actions/count';
	import type { PhotoSet } from '$lib/utils/group-images';

	let {
		sets,
		active,
		total,
		onSelect
	}: {
		sets: PhotoSet[];
		active: string;
		total: number;
		onSelect: (slug: string) => void;
	} = $props();

	let items = $state<HTMLButtonElement[]>([]);
	// The underline is one element that travels, not a border that blinks on and off.
	let mark = $state({ left: 0, width: 0 });

	function measure() {
		const index = active === 'all' ? 0 : sets.findIndex((set) => set.slug === active) + 1;
		const node = items[index];
		if (!node) {
			mark = { left: 0, width: 0 };
			return;
		}
		mark = { left: node.offsetLeft, width: node.offsetWidth };
	}

	$effect(() => {
		// Re-measure when the selection changes and when the row rewraps.
		active;
		sets;
		measure();
		window.addEventListener('resize', measure);
		return () => window.removeEventListener('resize', measure);
	});
</script>

<nav class="index" aria-label="Filter by set">
	<span class="index__mark" style="--left: {mark.left}px; --width: {mark.width}px" aria-hidden="true"
	></span>

	<button
		class="index__item label"
		class:index__item--active={active === 'all'}
		type="button"
		aria-current={active === 'all' ? 'true' : undefined}
		bind:this={items[0]}
		onclick={() => onSelect('all')}
	>
		All <span class="index__count num" use:count={total}></span>
	</button>

	{#each sets as set, index (set.slug)}
		<button
			class="index__item label"
			class:index__item--active={active === set.slug}
			type="button"
			aria-current={active === set.slug ? 'true' : undefined}
			bind:this={items[index + 1]}
			onclick={() => onSelect(set.slug)}
		>
			{set.name} <span class="index__count num" use:count={set.photos.length}></span>
		</button>
	{/each}
</nav>

<style>
	.index {
		position: relative;
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem 1.25rem;
	}

	.index__mark {
		position: absolute;
		bottom: 0;
		height: 1px;
		width: var(--width, 0);
		background: var(--line-2);
		transform: translate3d(var(--left, 0), 0, 0);
		transition:
			transform 0.45s cubic-bezier(0.16, 0.84, 0.28, 1),
			width 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.index__item {
		display: inline-flex;
		align-items: baseline;
		gap: 0.4rem;
		padding-block: 0.125rem;
		color: var(--muted);
		transition: color 0.25s ease;
	}

	.index__item:hover {
		color: var(--fg);
	}

	.index__item--active {
		color: var(--fg);
	}

	.index__count {
		font-size: 0.625rem;
		color: var(--faint);
		transition: color 0.25s ease;
	}

	.index__item--active .index__count {
		color: var(--muted);
	}
</style>
