<script lang="ts">
	import PocketBase from 'pocketbase';
	import { browser } from '$app/environment';

	// Isolated PB client so superuser login doesn't touch the regular user session.
	const adminPb = new PocketBase(browser ? window.location.origin : '/');
	adminPb.autoCancellation(false);

	// Superuser auth state (separate from regular user auth).
	let loggedIn = $state(false);
	let email = $state('');
	let password = $state('');
	let loginBusy = $state(false);
	let loginError = $state('');

	// Per-action state.
	let rankingsBusy = $state(false);
	let rankingsMsg = $state('');
	let rankingsTone = $state<'ok' | 'error'>('ok');

	let syncBusy = $state(false);
	let syncMsg = $state('');
	let syncTone = $state<'ok' | 'error'>('ok');

	async function loginSuperuser() {
		loginBusy = true;
		loginError = '';
		try {
			await adminPb.collection('_superusers').authWithPassword(email, password);
			loggedIn = true;
			password = '';
		} catch {
			loginError = 'Invalid credentials.';
		} finally {
			loginBusy = false;
		}
	}

	async function refreshRankings() {
		rankingsBusy = true;
		rankingsMsg = '';
		try {
			await adminPb.send('/api/admin/rankings/refresh', { method: 'POST' });
			rankingsTone = 'ok';
			rankingsMsg = 'Rankings refreshed.';
		} catch (e: unknown) {
			rankingsTone = 'error';
			rankingsMsg = (e as { message?: string })?.message ?? 'Failed.';
		} finally {
			rankingsBusy = false;
		}
	}

	async function syncResults() {
		syncBusy = true;
		syncMsg = '';
		try {
			await adminPb.send('/api/sync/refresh', { method: 'POST' });
			syncTone = 'ok';
			syncMsg = 'Sync complete.';
		} catch (e: unknown) {
			syncTone = 'error';
			syncMsg = (e as { message?: string })?.message ?? 'Failed.';
		} finally {
			syncBusy = false;
		}
	}

	function logout() {
		adminPb.authStore.clear();
		loggedIn = false;
	}
</script>

<div class="page">
	<h1>Admin</h1>

	{#if !loggedIn}
		<div class="card login-card">
			<h2>Superuser login</h2>
			<p class="muted">Log in with your PocketBase admin credentials.</p>
			<form onsubmit={(e) => { e.preventDefault(); loginSuperuser(); }}>
				<label>
					Email
					<input type="email" bind:value={email} required autocomplete="username" />
				</label>
				<label>
					Password
					<input type="password" bind:value={password} required autocomplete="current-password" />
				</label>
				{#if loginError}<p class="msg error">{loginError}</p>{/if}
				<button type="submit" disabled={loginBusy} class="btn-primary">
					{loginBusy ? 'Logging in…' : 'Log in'}
				</button>
			</form>
		</div>
	{:else}
		<div class="actions stagger">
			<section class="card action-card">
				<h2>FIFA Rankings</h2>
				<p class="muted">
					Re-apply the FIFA rankings embedded in this build to the database. Run this after deploying a new build with updated rankings.
				</p>
				{#if rankingsMsg}<p class="msg" class:ok={rankingsTone === 'ok'} class:error={rankingsTone === 'error'}>{rankingsMsg}</p>{/if}
				<button onclick={refreshRankings} disabled={rankingsBusy} class="btn-primary">
					{rankingsBusy ? 'Refreshing…' : 'Refresh rankings'}
				</button>
			</section>

			<section class="card action-card">
				<h2>Sync results</h2>
				<p class="muted">
					Force an immediate results sync from the configured provider (API-Football or openfootball).
				</p>
				{#if syncMsg}<p class="msg" class:ok={syncTone === 'ok'} class:error={syncTone === 'error'}>{syncMsg}</p>{/if}
				<button onclick={syncResults} disabled={syncBusy} class="btn-primary">
					{syncBusy ? 'Syncing…' : 'Sync now'}
				</button>
			</section>
		</div>

		<button onclick={logout} class="btn-ghost logout">
			Log out
		</button>
	{/if}
</div>

<style>
	.page {
		max-width: 540px;
		margin: 0 auto;
		padding: 1.5rem 0 3rem;
	}
	h1 {
		margin-bottom: 1.5rem;
	}
	h2 {
		margin: 0 0 0.4rem;
		font-size: 1rem;
		font-weight: 700;
	}
	.login-card,
	.action-card {
		padding: 1.5rem;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
		margin-top: 1rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.85rem;
		font-weight: 600;
	}
	input {
		padding: 0.55rem 0.75rem;
		border: 1px solid var(--border);
		border-radius: var(--radius);
		background: var(--surface);
		color: var(--text);
		font: inherit;
		font-size: 0.95rem;
	}
	input:focus {
		outline: 2px solid var(--accent);
		outline-offset: 1px;
	}
	.actions {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
	}
	.action-card {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.btn-primary {
		align-self: flex-start;
		padding: 0.55rem 1.1rem;
		background: var(--accent);
		color: var(--bg);
		border: none;
		border-radius: var(--radius);
		font: 700 0.875rem var(--font);
		cursor: pointer;
		transition: opacity 0.1s;
	}
	.btn-primary:disabled {
		opacity: 0.5;
		cursor: default;
	}
	.btn-ghost {
		background: none;
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 0.45rem 1rem;
		color: var(--muted);
		font: inherit;
		font-size: 0.85rem;
		cursor: pointer;
	}
	.logout {
		margin-top: 1.5rem;
	}
	.msg {
		font-size: 0.85rem;
		padding: 0.4rem 0.7rem;
		border-radius: var(--radius);
	}
	.msg.ok {
		background: color-mix(in srgb, var(--accent) 12%, transparent);
		color: var(--accent);
	}
	.msg.error {
		background: color-mix(in srgb, var(--error, #e53e3e) 12%, transparent);
		color: var(--error, #e53e3e);
	}
</style>
