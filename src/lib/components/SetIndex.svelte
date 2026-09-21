<script lang="ts">
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
</script>

<nav class="index" aria-label="Filter by set">
	<button
		class="index__item label"
		class:index__item--active={active === 'all'}
		type="button"
		aria-current={active === 'all' ? 'true' : undefined}
		onclick={() => onSelect('all')}
	>
		All <span class="index__count num">{total}</span>
	</button>

	{#each sets as set (set.slug)}
		<button
			class="index__item label"
			class:index__item--active={active === set.slug}
			type="button"
			aria-current={active === set.slug ? 'true' : undefined}
			onclick={() => onSelect(set.slug)}
		>
			{set.name} <span class="index__count num">{set.photos.length}</span>
		</button>
	{/each}
</nav>

<style>
	.index {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem 1.25rem;
	}

	.index__item {
		display: inline-flex;
		align-items: baseline;
		gap: 0.4rem;
		padding-block: 0.125rem;
		color: var(--muted);
		border-bottom: 1px solid transparent;
		transition:
			color 0.25s ease,
			border-color 0.25s ease;
	}

	.index__item:hover {
		color: var(--fg);
	}

	.index__item--active {
		color: var(--fg);
		border-bottom-color: var(--line-2);
	}

	.index__count {
		font-size: 0.625rem;
		color: var(--faint);
	}

	.index__item--active .index__count {
		color: var(--muted);
	}
</style>
