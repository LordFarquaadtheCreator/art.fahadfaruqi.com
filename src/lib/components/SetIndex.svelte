<script lang="ts">
	import { count } from '$lib/actions/count';
	import { navLine } from '$lib/actions/nav-line';
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

<div class='filters'>
    <nav class="index" aria-label="Filter by set">
	<button
		class="nav-button label index__item"
		type="button"
		aria-current={active === 'all' ? 'true' : undefined}
		onclick={() => onSelect('all')}
		use:navLine
	>
		All <span class="index__count" use:count={total}></span>
	</button>

	{#each sets as set (set.slug)}
		<button
			class="nav-button label index__item"
			type="button"
			aria-current={active === set.slug ? 'true' : undefined}
			onclick={() => onSelect(set.slug)}
			use:navLine
		>
			{set.name} <span class="index__count" use:count={set.photos.length}></span>
		</button>
	{/each}
    </nav>
</div>

<style>
	.filters {
		border-bottom: 1px solid var(--line);
		scroll-margin-top: calc(var(--header-h) + 0.5rem);
	}

	.index {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem 1.25rem;
	}

	.index__count {
		font-size: 0.9em;
		color: var(--faint);
		transition: color 0.25s ease;
	}

	.index__item[aria-current] .index__count {
		color: var(--accent);
	}
</style>
