<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { auth } from '$lib/auth.svelte';
	import { language } from '$lib/language.svelte';
	import { strings } from '$lib/strings';

	let code = $derived($page.params.code ?? '');
	let leagueName = $state('');
	let phase = $state<'loading' | 'invite' | 'joining' | 'invalid' | 'error'>(
		'loading'
	);
	const t = $derived(strings[language.resolved]);
	const isEnglish = $derived(language.isEnglish);

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
			<p class="muted">{isEnglish ? 'Checking invitation…' : 'Sjekkar invitasjonen…'}</p>
		{:else if phase === 'joining'}
			<p class="muted">{isEnglish ? 'Joining' : 'Blir med i'} <strong>{leagueName}</strong>…</p>
		{:else if phase === 'invite'}
			<p class="kicker">{isEnglish ? 'You are invited' : 'Du er invitert'}</p>
			<h2 class="lname">{leagueName}</h2>
			<p class="muted">
				{isEnglish ? 'Log in or create an account to join this league.' : 'Logg inn eller opprett konto for å bli med i denne ligaen.'}
			</p>
			<a class="btn" href={`/register?invite=${encodeURIComponent(code)}`}>
				{isEnglish ? 'Create account' : 'Opprett konto'}
			</a>
			<a
				class="btn secondary"
				href={`/login?invite=${encodeURIComponent(code)}`}
			>
				{isEnglish ? 'Log in' : 'Logg inn'}
			</a>
		{:else if phase === 'error'}
			<p class="error">{isEnglish ? 'Could not join the league. Try again.' : 'Kunne ikkje bli med i ligaen. Prøv igjen.'}</p>
			<a class="btn secondary" href="/leagues">{isEnglish ? 'Go to leagues' : 'Gå til ligaer'}</a>
		{:else}
			<p class="error">{isEnglish ? 'This invite link is invalid or expired.' : 'Invitasjonslenka er ugyldig eller utløpt.'}</p>
			<a class="btn secondary" href="/">{isEnglish ? 'Go to home' : 'Gå til heim'}</a>
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
