<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { auth } from '$lib/auth.svelte';
	import { strings } from '$lib/strings';

	let code = $derived($page.params.code ?? '');
	let leagueName = $state('');
	let phase = $state<'loading' | 'invite' | 'joining' | 'invalid' | 'error'>(
		'loading'
	);
	const t = strings;

	// Resolve the code once, then either auto-join (authed) or show the
	// sign-in / create-account choice (carrying the invite code through).
	$effect(() => {
		const c = code;
		if (!c) {
			phase = 'invalid';
			return;
		}
		let cancelled = false;
		(async () => {
			try {
				const lg = await api.invitePreview(c);
				if (cancelled) return;
				leagueName = lg.name;
				if (auth.isAuthed) {
					phase = 'joining';
					const r = await api.joinLeague(c);
					if (!cancelled) goto(`/leagues/${r.id}`);
				} else {
					phase = 'invite';
				}
			} catch {
				if (!cancelled) phase = phase === 'joining' ? 'error' : 'invalid';
			}
		})();
		return () => {
			cancelled = true;
		};
	});
</script>

<div class="auth">
	<h1>WC Pool</h1>
	<p class="muted">{t.auth.tagline}</p>

	<div class="card">
		{#if phase === 'loading'}
			<p class="muted">Checking invitation…</p>
		{:else if phase === 'joining'}
			<p class="muted">Joining <strong>{leagueName}</strong>…</p>
		{:else if phase === 'invite'}
			<p class="kicker">You are invited</p>
			<h2 class="lname">{leagueName}</h2>
			<p class="muted">
				Log in or create an account to join this league.
			</p>
			<a class="btn" href={`/register?invite=${encodeURIComponent(code)}`}>
				Create account
			</a>
			<a
				class="btn secondary"
				href={`/login?invite=${encodeURIComponent(code)}`}
			>
				Log in
			</a>
		{:else if phase === 'error'}
			<p class="error">Could not join the league. Try again.</p>
			<a class="btn secondary" href="/leagues">Go to leagues</a>
		{:else}
			<p class="error">This invite link is invalid or expired.</p>
			<a class="btn secondary" href="/">Go to home</a>
		{/if}
	</div>
</div>

<style>
	.auth {
		max-width: 380px;
		margin: 12dvh auto 0;
	}
	h1 {
		margin: 0;
		font-size: 2rem;
	}
	.muted {
		margin: 0.25rem 0 1.5rem;
	}
	.lname {
		margin: 0.1rem 0 0.6rem;
		font-size: 1.7rem;
	}
	.card .btn + .btn {
		margin-top: 0.6rem;
	}
</style>
