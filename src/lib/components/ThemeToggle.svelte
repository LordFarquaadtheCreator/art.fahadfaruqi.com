<script lang="ts">
	import { browser } from '$app/environment';

	type Theme = 'dark' | 'light';

	let theme = $state<Theme>('dark');

	// The theme is applied by the inline script in app.html before first paint;
	// this only reads it back and persists changes.
	$effect(() => {
		if (browser) {
			theme = document.documentElement.dataset.theme === 'light' ? 'light' : 'dark';
		}
	});

	function toggle() {
		theme = theme === 'dark' ? 'light' : 'dark';
		document.documentElement.dataset.theme = theme;
		localStorage.setItem('theme', theme);
	}
</script>

<button
	class="toggle label"
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
	}
</style>
