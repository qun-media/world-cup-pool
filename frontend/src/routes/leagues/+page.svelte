<script lang="ts">
	import { api, type LeagueSummary } from '$lib/api';
	import { goto } from '$app/navigation';
	import { language } from '$lib/language.svelte';
	import PendingInvites from '$lib/components/PendingInvites.svelte';
	import { ArrowRight, Crown, Globe2, LogIn, Plus, Users } from '@lucide/svelte';

	let leagues = $state<LeagueSummary[]>([]);
	let loaded = $state(false);
	let newName = $state('');
	let joinCode = $state('');
	let error = $state('');
	let busy = $state(false);
	const isEnglish = $derived(language.isEnglish);

	async function load() {
		try {
			leagues = (await api.myLeagues()).leagues;
		} catch {
			/* ignore */
		} finally {
			loaded = true;
		}
	}
	$effect(() => {
		load();
	});

	async function create(e: Event) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			const r = await api.createLeague(newName);
			newName = '';
			goto(`/leagues/${r.id}`);
		} catch {
			error = isEnglish ? 'Could not create league.' : 'Kunne ikkje opprette liga.';
		} finally {
			busy = false;
		}
	}

	async function join(e: Event) {
		e.preventDefault();
		error = '';
		busy = true;
		try {
			const r = await api.joinLeague(joinCode);
			joinCode = '';
			goto(`/leagues/${r.id}`);
		} catch {
			error = isEnglish ? 'Invalid invite code.' : 'Ugyldig invitasjonskode.';
		} finally {
			busy = false;
		}
	}

	function roleLabel(league: LeagueSummary) {
		if (league.inviteCode === 'GLOBAL') return 'Global';
		return league.role === 'owner'
			? isEnglish ? 'Owner' : 'Eigar'
			: isEnglish ? 'Member' : 'Medlem';
	}
</script>

<header class="league-hero">
	<p class="kicker">{isEnglish ? 'Play against your friends' : 'Spel mot venene dine'}</p>
	<h1>{isEnglish ? 'Leagues' : 'Ligaer'}</h1>
	<p class="muted">{isEnglish ? 'Pick a league, see the table, and jump straight to chat.' : 'Vel ei liga, sjå tabellen og hopp rett til chat.'}</p>
</header>

<PendingInvites compact />

<section class="league-section">
	<div class="section-head">
		<div>
			<p class="kicker">{isEnglish ? 'Overview' : 'Oversikt'}</p>
			<h2>{isEnglish ? 'Your leagues' : 'Ligaene dine'}</h2>
		</div>
		{#if loaded}<span class="count-pill">{leagues.length}</span>{/if}
	</div>
	{#if !loaded}
		<div class="league-grid">
			<div class="league-tile skeleton"></div>
			<div class="league-tile skeleton hide-mobile"></div>
		</div>
	{:else if leagues.length === 0}
		<div class="empty-state">
			<strong>{isEnglish ? 'No leagues yet' : 'Ingen ligaer enno'}</strong>
			<p class="muted">{isEnglish ? 'Create a league or join with an invite code.' : 'Opprett ei liga eller bli med med invitasjonskode.'}</p>
		</div>
	{:else}
		<div class="league-grid">
			{#each leagues as league (league.id)}
				<a
					class="league-tile"
					class:global={league.inviteCode === 'GLOBAL'}
					href={`/leagues/${league.id}`}
				>
					<span class="league-icon">
						{#if league.inviteCode === 'GLOBAL'}<Globe2 size={20} />
						{:else if league.role === 'owner'}<Crown size={20} />
						{:else}<Users size={20} />{/if}
					</span>
					<span class="league-main">
						<span class="league-topline">
							<b>{league.name}</b>
							<i>{roleLabel(league)}</i>
						</span>
						<span class="league-meta">
							<span><Users size={14} /> {league.members} {isEnglish ? (league.members === 1 ? 'member' : 'members') : (league.members === 1 ? 'medlem' : 'medlemer')}</span>
							{#if league.inviteCode !== 'GLOBAL'}<span>{isEnglish ? 'Code' : 'Kode'} {league.inviteCode}</span>{/if}
						</span>
					</span>
					<span class="go"><ArrowRight size={18} /></span>
				</a>
			{/each}
		</div>
	{/if}
</section>

<div class="action-grid">
	<section class="card action-card">
		<div class="action-title">
			<span class="action-icon"><Plus size={18} /></span>
			<div>
				<h3>{isEnglish ? 'Create league' : 'Opprett liga'}</h3>
				<p class="muted">{isEnglish ? 'Start a new private competition.' : 'Start ei ny privat tevling.'}</p>
			</div>
		</div>
		<form onsubmit={create}>
			<div class="field">
				<input class="input" placeholder={isEnglish ? 'League name' : 'Liganamn'} bind:value={newName} required />
			</div>
			<button class="btn" disabled={busy || !newName.trim()}>{isEnglish ? 'Create' : 'Opprett'}</button>
		</form>
	</section>

	<section class="card action-card">
		<div class="action-title">
			<span class="action-icon secondary"><LogIn size={18} /></span>
			<div>
				<h3>{isEnglish ? 'Join' : 'Bli med'}</h3>
				<p class="muted">{isEnglish ? 'Paste the code from the invite.' : 'Lim inn koden frå invitasjonen.'}</p>
			</div>
		</div>
		<form onsubmit={join}>
			<div class="field">
				<input
					class="input code"
					placeholder="INVITE CODE"
					bind:value={joinCode}
					required
				/>
			</div>
			<button class="btn secondary" disabled={busy || !joinCode.trim()}>{isEnglish ? 'Join' : 'Bli med'}</button>
		</form>
	</section>
</div>

{#if error}<p class="error">{error}</p>{/if}

<style>
	.league-hero {
		display: grid;
		gap: 0.35rem;
		margin: 0.2rem 0 1rem;
	}
	.league-hero h1,
	.section-head h2,
	.action-title h3 {
		margin: 0;
	}
	.muted {
		margin: 0;
	}
	.league-section {
		display: grid;
		gap: 0.75rem;
	}
	.section-head {
		display: flex;
		align-items: end;
		justify-content: space-between;
		gap: 1rem;
	}
	.section-head h2 {
		font-size: 1.2rem;
	}
	.count-pill {
		display: grid;
		place-items: center;
		min-width: 2.1rem;
		height: 2.1rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		background: var(--surface);
		font-family: var(--font-mono);
		font-weight: 800;
		color: var(--text);
	}
	.league-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 0.75rem;
	}
	.league-tile {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.8rem;
		min-height: 104px;
		padding: 1rem;
		border: 1px solid var(--border);
		border-radius: var(--radius);
		background: var(--surface);
		color: var(--text);
		box-shadow: var(--shadow-tile);
		transition: transform 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
	}
	.league-tile:hover,
	.league-tile:focus-visible {
		border-color: color-mix(in srgb, var(--accent) 38%, var(--border));
		box-shadow: var(--shadow-pop);
		transform: translateY(-2px);
		outline: none;
	}
	.league-tile.global {
		background: color-mix(in srgb, var(--accent) 6%, var(--surface));
	}
	.league-icon,
	.go,
	.action-icon {
		display: grid;
		place-items: center;
		border-radius: 14px;
		background: var(--surface-2);
		color: var(--muted);
	}
	.league-icon {
		width: 44px;
		height: 44px;
	}
	.league-tile.global .league-icon {
		background: color-mix(in srgb, var(--accent) 14%, var(--surface-2));
		color: var(--accent);
	}
	.league-main {
		display: grid;
		gap: 0.45rem;
		min-width: 0;
	}
	.league-topline {
		display: grid;
		gap: 0.35rem;
	}
	.league-topline b {
		font-size: 1.05rem;
		line-height: 1.1;
		overflow-wrap: anywhere;
	}
	.league-topline i {
		width: fit-content;
		padding: 0.16rem 0.45rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		background: var(--surface-2);
		color: var(--muted);
		font-style: normal;
		font-size: 0.7rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}
	.league-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem 0.6rem;
		color: var(--muted);
		font-size: 0.82rem;
		font-weight: 650;
	}
	.league-meta span {
		display: inline-flex;
		align-items: center;
		gap: 0.28rem;
	}
	.go {
		width: 34px;
		height: 34px;
		border-radius: var(--radius-pill);
		transition: transform 0.16s ease, color 0.16s ease, background 0.16s ease;
	}
	.league-tile:hover .go,
	.league-tile:focus-visible .go {
		background: var(--text);
		color: var(--bg);
		transform: translateX(2px);
	}
	.empty-state {
		padding: 1rem;
		border: 1px dashed var(--border-strong);
		border-radius: var(--radius);
		background: var(--surface);
	}
	.empty-state strong {
		display: block;
		margin-bottom: 0.25rem;
	}
	.skeleton {
		min-height: 104px;
		background: linear-gradient(90deg, var(--surface), var(--surface-2), var(--surface));
		background-size: 200% 100%;
		animation: pulse 1.4s ease-in-out infinite;
	}
	.action-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 0.85rem;
		margin-top: 0.95rem;
	}
	.action-card {
		display: grid;
		gap: 1rem;
	}
	.action-title {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.action-title .muted {
		font-size: 0.86rem;
	}
	.action-icon {
		width: 42px;
		height: 42px;
		background: color-mix(in srgb, var(--accent) 12%, var(--surface-2));
		color: var(--accent);
	}
	.action-icon.secondary {
		background: color-mix(in srgb, var(--accent-2) 12%, var(--surface-2));
		color: var(--accent-2);
	}
	.field {
		margin-bottom: 0.65rem;
	}
	.code {
		text-transform: uppercase;
		letter-spacing: 0.2em;
		font-weight: 700;
	}
	@keyframes pulse {
		from { background-position: 200% 0; }
		to { background-position: -200% 0; }
	}
	@media (max-width: 560px) {
		.league-grid,
		.action-grid {
			grid-template-columns: minmax(0, 1fr);
		}
		.league-tile {
			min-height: 96px;
			padding: 0.9rem;
			gap: 0.7rem;
		}
		.league-icon {
			width: 40px;
			height: 40px;
		}
		.go {
			width: 31px;
			height: 31px;
		}
		.hide-mobile {
			display: none;
		}
	}
</style>
