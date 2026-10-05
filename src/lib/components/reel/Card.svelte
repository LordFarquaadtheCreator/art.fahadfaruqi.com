<script lang="ts">
	import type { Section } from './types';

	let { section }: { section: Section } = $props();

	const pad = (value: number) => String(value).padStart(2, '0');
	const ratio = (width: number, height: number) =>
		width > 0 && height > 0 ? `${width} / ${height}` : '3 / 2';
</script>

{#if section.kind === 'image'}
	<figure class="frame frame--image">
		<div class="frame__inner" style="--ratio: {ratio(section.width, section.height)}">
			<img src={section.src} alt={section.alt} loading="eager" decoding="async" draggable="false" />
		</div>
		<figcaption class="frame__cap label">
			<span class="num">{pad(section.index)}</span>
			<span class="frame__title">{section.title}</span>
		</figcaption>
	</figure>
{:else}
	<article
		class="frame"
		class:frame--note={section.kind === 'note'}
		class:frame--kept={section.kind === 'kept'}
	>
		<div class="frame__inner">
			{#if section.kind === 'kept'}
				<span class="frame__mark" aria-hidden="true"></span>
			{/if}
			<span class="label frame__tag">{section.label}</span>
			<p class="frame__text">{section.text}</p>
		</div>
	</article>
{/if}

<style>
	/* The card itself: presentation only. Every turn's component brings its own motion. */

	.frame {
		display: flex;
		flex-direction: column;
	}

	.frame__inner {
		position: relative;
		transform-origin: center center;
	}

	.frame--image .frame__inner {
		height: var(--frame-h, 58vh);
		aspect-ratio: var(--ratio, 3 / 2);
		overflow: hidden;
		background: var(--bg-elev);
	}

	.frame--image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.frame__cap {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		max-width: 100%;
		padding-top: 0.6rem;
		color: var(--muted);
	}

	.frame__title {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	/* Notes and archive slots are the same object: a card in the turn. */
	.frame--note .frame__inner,
	.frame--kept .frame__inner {
		height: var(--frame-h, 58vh);
		width: clamp(17rem, 30vw, 32rem);
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: flex-start;
		gap: 1rem;
		padding: clamp(1.25rem, 3vw, 2.25rem);
		background: var(--bg-elev);
		border: 1px solid var(--line);
	}

	.frame--kept .frame__inner {
		width: clamp(15rem, 25vw, 26rem);
		align-items: center;
		text-align: center;
		border-style: dashed;
		border-color: var(--line-2);
	}

	.frame__text {
		margin: 0;
		max-width: 30ch;
		font-size: clamp(1rem, 1.5vw, 1.35rem);
		line-height: 1.4;
	}

	.frame--kept .frame__text {
		color: var(--muted);
		font-size: var(--small-size);
	}

	.frame__mark {
		position: relative;
		width: 2.25rem;
		height: 2.25rem;
		border: 1px solid var(--line-2);
		border-radius: 50%;
	}

	.frame__mark::before,
	.frame__mark::after {
		content: '';
		position: absolute;
		background: var(--line-2);
	}

	.frame__mark::before {
		left: 50%;
		top: -0.4rem;
		bottom: -0.4rem;
		width: 1px;
	}

	.frame__mark::after {
		top: 50%;
		left: -0.4rem;
		right: -0.4rem;
		height: 1px;
	}
</style>
