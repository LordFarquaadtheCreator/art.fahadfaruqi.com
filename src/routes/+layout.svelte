<script lang="ts">
	import '@fontsource-variable/inter';
	// The readout face. Two weights only: the readouts are set in 400, and italic is
	// never used, so nothing else ships.
	import '@fontsource/ibm-plex-mono/400.css';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<!-- Same mark as the main site at fahadfaruqi.com. -->
	<link rel="icon" href="/favicon.png" />
</svelte:head>

{@render children()}
