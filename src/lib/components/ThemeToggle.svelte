<script lang="ts">
	import { browser } from '$app/environment';

	type Theme = 'dark' | 'light';

	let theme = $state<Theme>('dark');

	// The canvas colour, duplicated from --bg: it has to exist before the stylesheets do.
	const CANVAS: Record<Theme, string> = { dark: '#0a0a0a', light: '#f2f1ed' };

	/** Everything the palette does not reach: the properties the user agent paints itself —
	 *  the canvas behind the document, the scrollbars, and the browser chrome on mobile. */
	function paintAgentSurfaces(next: Theme) {
		const root = document.documentElement;
		root.dataset.theme = next;
		root.style.colorScheme = next;
		root.style.backgroundColor = CANVAS[next];
		document
			.querySelector('meta[name="theme-color"]')
			?.setAttribute('content', CANVAS[next]);
	}

	// The theme is applied by the inline script in app.html before first paint;
	// this reads it back, and reconciles the surfaces that script could not know about.
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
