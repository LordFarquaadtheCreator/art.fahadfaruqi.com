<script lang="ts">
	import { turn } from '$lib/content/about-reel';
	import { reel } from '$lib/utils/reel';
	import type { ReelAnimation } from './animation';
	import { sectionComponent } from './sections';

	// The About page's mechanic: one card at a time turns through the centre. The turn's
	// length scales with how many sections it holds, and every frame it hands each
	// section's slot two numbers:
	//
	//   --c     the beat's signed distance from the centre (0 = in the gate)
	//   --gate  how close to the centre it is, 0..1
	//
	// Each section's own component decides what to do with them.
	const perFrame = turn.params.perFrame ?? 0.6;

	const animation: ReelAnimation = {
		measure: ({ vh, marks }) => perFrame * vh * marks.length,

		tick: ({ p, marks, show }) => {
			const count = marks.length;
			if (!count) return;

			const head = p * count;

			marks.forEach((mark, i) => {
				const c = head - i;
				const gate = Math.max(0, 1 - Math.abs(c));

				mark.el.style.setProperty('--c', c.toFixed(4));
				mark.el.style.setProperty('--gate', gate.toFixed(3));
			});

			const current = marks[Math.min(count - 1, Math.max(0, Math.floor(head)))];
			if (current) show(current.index);
		}
	};

	const options = $derived({ ...turn.progression, animation });
	const pad = (value: number) => String(value).padStart(2, '0');
</script>

<section class="reel" use:reel={options} aria-label="About — the turn">
	<div class="reel__viewport" data-reel-viewport>
		<div class="reel__chrome" aria-hidden="true">
			<span class="label">About — the turn</span>
			<span class="label">
				Turn <span class="num reel__count" data-reel-count>01</span> /
				{pad(turn.sections.length)}
			</span>
		</div>

		<div class="stage">
			{#each turn.sections as section (section.id)}
				{@const Beat = sectionComponent(section.component)}
				<div class="slot" data-reel-frame data-index={section.index}>
					<Beat {section} />
				</div>
			{/each}
		</div>

		<div class="reel__foot" aria-hidden="true">
			<span class="label">{turn.sections.length} turns</span>
			<span class="label">scroll</span>
		</div>
	</div>
</section>

<style>
	.reel {
		--frame-h: min(58vh, calc(100dvh - var(--header-h) - 9rem));
		position: relative;
		margin-inline: calc(var(--gutter) * -1);
	}

	.reel__viewport {
		position: relative;
	}

	/* Fallback: the sections are a readable column; each slot keeps its own order. */
	.stage {
		display: grid;
		gap: clamp(2rem, 6vh, 4rem);
		justify-items: center;
		padding: 1rem var(--gutter);
	}

	.slot {
		display: flex;
		justify-content: center;
		width: 100%;
	}

	.reel__chrome,
	.reel__foot {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 1rem;
		padding-inline: var(--gutter);
	}

	.reel__chrome {
		padding-block: 0.75rem;
		border-bottom: 1px solid var(--line);
	}

	.reel__foot {
		padding-block: 1rem;
	}

	.reel__count {
		color: var(--accent);
	}

	/* Reel mode. Declared after the base rules on purpose: the two tie on specificity, so
	   source order decides which wins. */

	:global(.reel[data-reel='on']) {
		height: calc(100dvh + var(--pin, 100vh));
	}

	:global(.reel[data-reel='on']) .reel__viewport {
		position: sticky;
		top: 0;
		height: 100dvh;
		display: flex;
		flex-direction: column;
		justify-content: center;
		overflow: clip;
	}

	:global(.reel[data-reel='on']) .stage {
		display: block;
		padding: 0;
	}

	/* Every section occupies the same slot; the section's component moves itself inside it. */
	:global(.reel[data-reel='on']) .slot {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
	}

	:global(.reel[data-reel='on']) .reel__chrome,
	:global(.reel[data-reel='on']) .reel__foot {
		position: absolute;
		left: 0;
		right: 0;
		z-index: 1;
	}

	:global(.reel[data-reel='on']) .reel__chrome {
		top: 0;
		padding-top: calc(var(--header-h) + 1rem);
		border-bottom: 0;
		opacity: var(--entry, 1);
	}

	:global(.reel[data-reel='on']) .reel__foot {
		bottom: 0;
		padding-bottom: 1.5rem;
		opacity: calc(1 - var(--exit, 0));
	}
</style>
