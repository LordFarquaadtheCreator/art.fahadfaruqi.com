<script lang="ts">
	import { exifRows, type Exif } from '$lib/utils/exif';
	import { formatBytes } from '$lib/utils/bytes';

	let { exif, size }: { exif: Exif; size?: number } = $props();

	// The file's own weight is the one piece of telemetry the readouts were missing, and
	// it comes from the listing, not from a guess.
	const bytes = $derived(size ? formatBytes(size) : null);
	const rows = $derived([
		...exifRows(exif),
		...(bytes ? [{ label: 'File', value: bytes }] : [])
	]);
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
		/* The rule above the panel is where the amber goes: the panel is a record, not a
		   live readout, so its text stays monochrome. */
		border-top: 1px solid var(--accent-line);
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
