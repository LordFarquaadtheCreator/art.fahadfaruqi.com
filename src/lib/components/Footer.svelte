<script lang="ts">
	import { reveal } from '$lib/actions/reveal';

	let copied = $state(false);
</script>

<footer class="footer shell" use:reveal>
	<div class="footer__row">
		<div class="footer__span">
			<span class="label">All photographs by Fahad Faruqi.</span>
		</div>
		<div class="footer__span">
			<span class="label">Human Made.</span>
		</div>
		<div class="footer__span">
			<button
				class="label"
				type="button"
				aria-pressed={copied}
				onclick={async () => {
					await navigator.clipboard.writeText('fahadfaruqi1@gmail.com');
					copied = true;
					setTimeout(() => {
						copied = false;
					}, 1000);
				}}
			>
				{!copied ? 'fahadfaruqi1@gmail.com' : 'Copied.'}
			</button>
		</div>
	</div>
</footer>

<style>
	.footer {
		position: relative;
		z-index: 1;
		padding-block: clamp(2rem, 6vh, 4rem);
		border-top: 1px solid var(--line);
		transition:
			opacity 0.9s ease,
			transform 0.9s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	:global(.footer[data-revealed='false']) {
		opacity: 0;
		transform: translate3d(0, 1rem, 0);
	}

	.footer__row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem 2rem;
		justify-content: space-evenly;
	}

	.footer__span {
		flex: 1;
		min-width: 0;
		display: flex;
		justify-content: center;
	}

	.footer__span .label {
		min-width: 12ch;
	}
</style>
