<script lang="ts">
	import { exifRows, type Exif } from '$lib/utils/exif';

	let { exif }: { exif: Exif } = $props();

	const rows = $derived(exifRows(exif));
</script>

{#if rows.length > 0}
	<dl class="meta">
		{#each rows as row (row.label)}
			<dt class="label">{row.label}</dt>
			<dd class="num">{row.value}</dd>
		{/each}
	</dl>
{/if}

<style>
	.meta {
		display: grid;
		grid-template-columns: max-content minmax(0, 1fr);
		gap: 0.4rem 1.25rem;
		margin: 0;
		border-top: 1px solid var(--line);
		padding-top: 0.75rem;
	}

	dt {
		color: var(--muted);
	}

	dd {
		margin: 0;
		font-size: 0.75rem;
		color: var(--fg);
	}
</style>
