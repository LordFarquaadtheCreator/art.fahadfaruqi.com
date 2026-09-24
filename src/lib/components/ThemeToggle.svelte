<script lang="ts">
	import { browser } from '$app/environment';
	import { CANVAS, type Theme } from '$lib/utils/theme';

	let theme = $state<Theme>('dark');

	function paintAgentSurfaces(next: Theme) {
		const root = document.documentElement;
		root.dataset.theme = next;
		root.style.colorScheme = next;
		root.style.backgroundColor = CANVAS[next];
		document
			.querySelector('meta[name="theme-color"]')
			?.setAttribute('content', CANVAS[next]);
	}

	$effect(() => {
		if (browser) {
			theme = document.documentElement.dataset.theme === 'light' ? 'light' : 'dark';
			paintAgentSurfaces(theme);
		}
	});

	function toggle() {
		theme = theme === 'dark' ? 'light' : 'dark';
		localStorage.setItem('theme', theme);
		paintAgentSurfaces(theme);
	}
</script>

<button
	class="toggle label"
	class:toggle--light={theme === 'light'}
	type="button"
	onclick={toggle}
	aria-label="Switch to {theme === 'dark' ? 'light' : 'dark'} theme"
>
	<span class="toggle__glyph" aria-hidden="true"></span>
	{theme === 'dark' ? 'Dark' : 'Light'}
</button>

<style>
	.toggle {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--muted);
		transition: color 0.2s ease;
	}

	.toggle:hover {
		color: var(--fg);
	}

	.toggle__glyph {
		width: 0.7rem;
		height: 0.7rem;
		border: 1px solid currentColor;
		border-radius: 50%;
		background: linear-gradient(90deg, currentColor 50%, transparent 50%);
		transition: transform 0.55s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	/* The half-filled disc turns over as the theme flips, so the switch reads as one
	   movement rather than a label swap. */
	.toggle--light .toggle__glyph {
		transform: rotate(180deg);
	}
</style>
