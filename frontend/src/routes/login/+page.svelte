<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { strings } from '$lib/strings';

	let identity = $state('');
	let password = $state('');
	let error = $state('');
	let busy = $state(false);
	const t = strings;

	// After signing in, resume an invite if one was carried in the URL.
	let invite = $derived($page.url.searchParams.get('invite'));
	function dest() {
		return invite ? `/join/${invite}` : '/';
	}
	let registerHref = $derived(
		invite ? `/register?invite=${encodeURIComponent(invite)}` : '/register'
	);

	async function submit(e: Event) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			await auth.login(identity, password);
			goto(dest());
		} catch {
			error = t.auth.wrongCredentials;
		} finally {
			busy = false;
		}
	}

</script>

<div class="auth">
	<div class="brand-intro">
		<Logo variant="hero" tagline={t.auth.tagline} />
		<p class="muted brand-copy">{t.auth.subtitle}</p>
	</div>

	<form class="card" onsubmit={submit}>
		<div class="field">
			<label for="id">{t.auth.emailLabel}</label>
			<input
				id="id"
				class="input"
				type="email"
				placeholder={t.auth.emailPlaceholder}
				bind:value={identity}
				autocomplete="email"
				required
			/>
		</div>
		<div class="field">
			<div class="lblrow">
				<label for="pw">{t.auth.passwordLabel}</label>
				<a class="forgot" href="/forgot-password">{t.auth.forgotPassword}</a>
			</div>
			<input
				id="pw"
				class="input"
				type="password"
				bind:value={password}
				autocomplete="current-password"
				required
			/>
		</div>
		{#if error}<p class="error">{error}</p>{/if}
		<button class="btn" disabled={busy}>{busy ? `${t.auth.login}…` : t.auth.login}</button>
		<p class="muted switch">
			{t.auth.newHere} <a href={registerHref}>{t.auth.createAccount}</a>
		</p>
	</form>
</div>

<style>
	.auth {
		max-width: 420px;
		margin: 8dvh auto 0;
		padding: 0 0.2rem;
	}
	.brand-intro {
		display: grid;
		gap: 0.8rem;
		margin-bottom: 1.25rem;
	}
	.brand-copy {
		max-width: 34ch;
		margin: 0;
		font-size: 0.98rem;
		line-height: 1.5;
	}
	@media (max-width: 520px) {
		.auth {
			margin-top: 5dvh;
		}
		.brand-intro {
			gap: 0.7rem;
			margin-bottom: 1rem;
		}
	}
	.switch {
		text-align: center;
		margin: 1rem 0 0;
	}
	.lblrow {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.75rem;
	}
	.forgot {
		font-size: 0.8rem;
		color: var(--muted);
	}
	.forgot:hover {
		color: var(--accent);
	}
</style>
