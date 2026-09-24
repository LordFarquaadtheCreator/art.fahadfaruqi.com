<svelte:head>
	<title>About — Fahad Faruqi</title>
	<meta name="description" content="About Fahad Faruqi." />
</svelte:head>

<section class="about shell">
	<h1 class="title-attention">About</h1>
	<p class="about__line label">Nothing written here yet.</p>
</section>

<style>
	.about {
		padding-block: clamp(3rem, 12vh, 9rem) var(--block);
	}

	.about__line {
		margin: clamp(2.5rem, 8vh, 6rem) 0 0;
		padding-top: 1rem;
		border-top: 1px solid var(--line);
		animation: about-in 0.8s cubic-bezier(0.16, 0.84, 0.28, 1) both;
		animation-delay: 0.24s;
	}

	@keyframes about-in {
		from {
			opacity: 0;
			transform: translate3d(0, 0.6rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
</style>
