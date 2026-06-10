<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api, type LeaderboardRow, type LeagueInvite, type LeagueInviteUser } from '$lib/api';
	import { auth } from '$lib/auth.svelte';
	import Avatar from '$lib/components/Avatar.svelte';
	import LeagueChatCard from '$lib/components/LeagueChatCard.svelte';
	import {
		Eye,
		EyeOff,
		Copy,
		Share2,
		ChevronDown,
		Telescope,
		Mail,
		Search,
		UserPlus
	} from '@lucide/svelte';

	interface Cfg {
		match: {
			tendency: number;
			exact: number;
			totalGoals: number;
			goalDiff: number;
		};
		forecast: {
			groupPosition: number;
			perfectGroupBonus: number;
			advance: number;
			round: Record<string, number>;
		};
		tiebreakers: string[];
	}
	let cfg = $state<Cfg | null>(null);

	const tbLabel: Record<string, string> = {
		points: 'Total points',
		exactScores: 'Most exact scores',
		correctWinners: 'Most correct winners',
		goalDiffDeviation: 'Smallest goal-difference error',
		fewestTips: 'Fewest submitted tips',
		earliestEdit: 'Earliest last edit (submitted first)'
	};
	const roundLabel: Record<string, string> = {
		R32: 'Round of 32',
		R16: 'Round of 16',
		QF: 'Quarter-final',
		SF: 'Semi-final',
		FINAL: 'Final',
		CHAMPION: 'Winner'
	};

	let revealed = $state(false);
	let openRow = $state<string | null>(null);

	let id = $derived($page.params.id ?? '');
	let league = $state<{ id: string; name: string } | null>(null);
	let role = $state('');
	let isAdmin = $state(false);
	let hideForecast = $state(false);
	let rows = $state<LeaderboardRow[]>([]);
	let invite = $state('');
	let loaded = $state(false);
	let error = $state('');
	let tab = $state<'total' | 'tipsPoints' | 'forecastPoints'>('total');
	let deleteConfirm = $state('');
	let deleteBusy = $state(false);
	let deleteError = $state('');
	let inviteAdmin = $state(false);
	let inviteQuery = $state('');
	let inviteCandidates = $state<LeagueInviteUser[]>([]);
	let pendingInvites = $state<LeagueInvite[]>([]);
	let inviteSearchBusy = $state(false);
	let inviteSendBusy = $state('');
	let inviteError = $state('');
	let settingsBusy = $state(false);
	let settingsError = $state('');
	let renameName = $state('');
	let renameBusy = $state(false);
	let renameError = $state('');
	let renameSuccess = $state(false);

	$effect(() => {
		const lid = id;
		loaded = false;
		cfg = null;
		inviteAdmin = false;
		inviteQuery = '';
		inviteCandidates = [];
		pendingInvites = [];
		inviteError = '';
		settingsError = '';
		renameName = '';
		renameError = '';
		renameSuccess = false;
		Promise.all([api.leaderboard(lid), api.myLeagues()])
			.then(([lb, mine]) => {
				league = lb.league;
				rows = lb.rows;
				cfg = (lb.scoring as Cfg | undefined) ?? null;
				hideForecast = lb.hideForecast ?? false;
				isAdmin = lb.isAdmin ?? false;
				const mineLeague = mine.leagues.find((l) => l.id === lid);
				invite = mineLeague?.inviteCode ?? '';
				role = mineLeague?.role ?? '';
				if (invite && invite !== 'GLOBAL') {
					void loadInviteManager(lid);
				}
			})
			.catch(() => (error = 'Could not load this league.'))
			.finally(() => (loaded = true));
	});

	$effect(() => {
		const lid = id;
		const q = inviteQuery.trim();
		if (!inviteAdmin) {
			inviteCandidates = [];
			inviteSearchBusy = false;
			return;
		}
		inviteSearchBusy = true;
		let cancelled = false;
		const delay = q.length === 0 ? 0 : 220;
		const timer = setTimeout(() => {
			api.inviteCandidates(lid, q)
				.then((result) => {
					if (!cancelled) inviteCandidates = result.users;
				})
				.catch(() => {
					if (!cancelled) inviteCandidates = [];
				})
				.finally(() => {
					if (!cancelled) inviteSearchBusy = false;
				});
		}, delay);
		return () => {
			cancelled = true;
			clearTimeout(timer);
			inviteSearchBusy = false;
		};
	});

	let sorted = $derived(
		[...rows].sort((a, b) => b[tab] - a[tab])
	);
	let fcView = $derived(tab === 'forecastPoints');

	function copyInvite() {
		navigator.clipboard?.writeText(invite);
	}

	let linkCopied = $state(false);
	let copyTimer: ReturnType<typeof setTimeout>;
	async function shareInvite() {
		const url = new URL(`/join/${encodeURIComponent(invite)}`, window.location.origin).toString();
		const title = 'Join my World Cup prediction league on Midttunet!';
		const text = 'Tap here to challenge me.';
		try {
			if (navigator.share) {
				await navigator.share({ title, text, url });
				return;
			}
		} catch (e: unknown) {
			if ((e as { name?: string })?.name === 'AbortError') return;
		}
		await navigator.clipboard?.writeText(url);
		linkCopied = true;
		clearTimeout(copyTimer);
		copyTimer = setTimeout(() => (linkCopied = false), 1800);
	}

	async function toggleHideForecast() {
		if (!league) return;
		settingsBusy = true;
		settingsError = '';
		const next = !hideForecast;
		try {
			await api.updateLeagueSettings(league.id, { hideForecast: next });
			hideForecast = next;
		} catch {
			settingsError = 'Could not save settings.';
		} finally {
			settingsBusy = false;
		}
	}

	async function renameLeague() {
		if (!league || !renameName.trim()) return;
		renameBusy = true;
		renameError = '';
		renameSuccess = false;
		try {
			await api.updateLeagueSettings(league.id, { name: renameName.trim() });
			league = { ...league, name: renameName.trim() };
			renameName = '';
			renameSuccess = true;
			setTimeout(() => (renameSuccess = false), 2500);
		} catch {
			renameError = 'Could not rename the league.';
		} finally {
			renameBusy = false;
		}
	}

	async function togglePaid(userId: string, current: boolean) {
		try {
			await api.setMemberPaid(id, userId, !current);
			rows = rows.map((r) => (r.userId === userId ? { ...r, paid: !current } : r));
		} catch {
			// silently ignore — paid state stays as-is
		}
	}

	async function deleteLeague() {
		if (!league || deleteConfirm.trim() !== league.name.trim()) return;
		deleteBusy = true;
		deleteError = '';
		try {
			await api.deleteLeague(league.id);
			await goto('/leagues');
		} catch {
			deleteError = 'Could not delete the league.';
		} finally {
			deleteBusy = false;
		}
	}

	async function loadInviteManager(leagueId: string) {
		try {
			const result = await api.leagueInvites(leagueId);
			if (id !== leagueId) return;
			pendingInvites = result.invites;
			inviteAdmin = true;
		} catch {
			if (id !== leagueId) return;
			pendingInvites = [];
			inviteAdmin = false;
		}
	}

	async function sendInvite(user: LeagueInviteUser) {
		inviteSendBusy = user.id;
		inviteError = '';
		try {
			const result = await api.createLeagueInvite(id, user.id);
			pendingInvites = [result.invite, ...pendingInvites];
			inviteCandidates = inviteCandidates.filter((candidate) => candidate.id !== user.id);
			inviteQuery = '';
		} catch {
			inviteError = 'Could not send invite.';
		} finally {
			inviteSendBusy = '';
		}
	}

	function inviteDate(iso: string) {
		const date = new Date(iso);
		if (!Number.isFinite(date.getTime())) return '';
		return new Intl.DateTimeFormat('en-US', {
			day: '2-digit',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23'
		}).format(date);
	}
</script>

<a href="/leagues" class="muted back">← Leagues</a>

{#if error}
	<p class="error">{error}</p>
{:else if !loaded}
	<p class="muted">Loading…</p>
{:else if league}
	<p class="kicker">League</p>
	<h1>{league.name}</h1>

	<section class="card">
		<div class="tabs">
			<button class:active={tab === 'total'} onclick={() => (tab = 'total')}>Total</button>
			<button class:active={tab === 'tipsPoints'} onclick={() => (tab = 'tipsPoints')}>Match tips</button>
			<button class:active={tab === 'forecastPoints'} onclick={() => (tab = 'forecastPoints')}>Forecast</button>
		</div>

		<table class="lb">
			<thead>
				<tr>
					<th>#</th>
					<th>Player</th>
					{#if fcView}
						<th class="num ext" title="Correct group placement">Grp</th>
						<th class="num ext" title="Teams that advanced from group stage">Adv</th>
						<th class="num ext" title="Predicted team that reached Round of 32">R32</th>
						<th class="num ext" title="Predicted team that reached Round of 16">R16</th>
						<th class="num ext" title="Predicted team that reached quarter-final">QF</th>
						<th class="num ext" title="Predicted team that reached semi-final">SF</th>
						<th class="num ext" title="Predicted team that reached final">F</th>
						<th class="num ext" title="Correct winner predicted">Win</th>
					{:else}
						<th class="num ext" title="Matches tipped">Tips</th>
						<th class="num ext" title="Forecast points">WC</th>
						<th class="num ext" title="Exact scores (tiebreaker 1)">Exact</th>
						<th class="num ext" title="Correct winners (tiebreaker 2)">Win</th>
						<th class="num ext" title="Goal-difference error (tiebreaker 3, lower is better)">GD&Delta;</th>
					{/if}
					<th class="num pts">Points</th>
				</tr>
			</thead>
			<tbody>
				{#each sorted as r, i (r.userId)}
					{@const f = r.forecast ?? {}}
					<tr
						class:lead={r.userId === auth.user?.id}
						class="main"
						class:open={openRow === r.userId}
						onclick={() =>
							(openRow = openRow === r.userId ? null : r.userId)}
					>
						<td class="rank">
							{#if i === 0}🥇
							{:else if i === 1}🥈
							{:else if i === 2}🥉
							{:else}{i + 1}{/if}
							{#if r.rankDelta > 0}<span class="delta up">↑{r.rankDelta}</span>
							{:else if r.rankDelta < 0}<span class="delta dn">↓{Math.abs(r.rankDelta)}</span>{/if}
						</td>
						<td class="player">
							<div class="pwrap">
								<Avatar name={r.name} src={r.avatarUrl} size={28} />
								<span class="pname">{r.name}</span>
								{#if !hideForecast || r.userId === auth.user?.id}
									<a
										class="fclink"
										href={`/forecast/${r.userId}`}
										title={`View ${r.name}'s Forecast`}
										onclick={(e) => e.stopPropagation()}
									>
										<Telescope size={15} />
									</a>
								{/if}
								{#if isAdmin}
									<button
										class="paid-toggle"
										class:active={r.paid}
										title={r.paid ? 'Mark as unpaid' : 'Mark as paid'}
										onclick={(e) => { e.stopPropagation(); togglePaid(r.userId, r.paid ?? false); }}
									>💲</button>
								{/if}
								<ChevronDown size={14} class="rx" />
							</div>
						</td>
						{#if fcView}
							<td class="num ext digits">{f.groups ?? 0}</td>
							<td class="num ext digits">{f.advance ?? 0}</td>
							<td class="num ext digits">{f.R32 ?? 0}</td>
							<td class="num ext digits">{f.R16 ?? 0}</td>
							<td class="num ext digits">{f.QF ?? 0}</td>
							<td class="num ext digits">{f.SF ?? 0}</td>
							<td class="num ext digits">{f.FINAL ?? 0}</td>
							<td class="num ext digits">{f.champion ? '✓' : '–'}</td>
						{:else}
							<td class="num ext digits">{r.predicted}</td>
							<td class="num ext digits">{r.forecastPoints}</td>
							<td class="num ext digits">{r.exactScores}</td>
							<td class="num ext digits">{r.correctWinners}</td>
							<td class="num ext digits">{r.gdDeviation}</td>
						{/if}
						<td class="num pts digits">{r[tab]}</td>
					</tr>
					{#if openRow === r.userId}
						<tr class="detail">
							<td colspan="3">
								{#if fcView}
									<div class="stats">
										<span><i>Correct group placement</i><b>{f.groups ?? 0}</b></span>
										<span><i>Advanced team</i><b>{f.advance ?? 0}</b></span>
										<span><i>Reached Round of 32</i><b>{f.R32 ?? 0}</b></span>
										<span><i>Reached Round of 16</i><b>{f.R16 ?? 0}</b></span>
										<span><i>Reached quarter-final</i><b>{f.QF ?? 0}</b></span>
										<span><i>Reached semi-final</i><b>{f.SF ?? 0}</b></span>
										<span><i>Reached final</i><b>{f.FINAL ?? 0}</b></span>
										<span><i>Correct winner</i><b>{f.champion ? 'Yes' : 'No'}</b></span>
									</div>
								{:else}
									<div class="stats">
										<span><i>Matches tipped</i><b>{r.predicted}</b></span>
										<span><i>Match tip points</i><b>{r.tipsPoints}</b></span>
										<span><i>Forecast points</i><b>{r.forecastPoints}</b></span>
										<span><i>Exact scores</i><b>{r.exactScores}</b></span>
										<span><i>Correct winners</i><b>{r.correctWinners}</b></span>
										<span><i>Goal-difference error</i><b>{r.gdDeviation}</b></span>
									</div>
								{/if}
							</td>
						</tr>
					{/if}
				{/each}
			</tbody>
		</table>

		<p class="muted small note">
			Points update automatically as results come in.
		</p>
	</section>

	{#if invite}
		<LeagueChatCard leagueId={league.id} />
	{/if}

	{#if invite && invite !== 'GLOBAL'}
		<section class="card invite">
			<div class="invite-head">
				<h3>Share league</h3>
				<p class="muted small">Share the code or link with the people you want to invite.</p>
			</div>
			<div class="irow">
				<div class="ic">
					<div class="muted small">Invite code</div>
					<div class="code" class:masked={!revealed}>
						{revealed ? invite : '•'.repeat(invite.length || 6)}
					</div>
				</div>
				<div class="spacer"></div>
				<button
					class="btn secondary eye"
					aria-label={revealed ? 'Hide code' : 'Show code'}
					onclick={() => (revealed = !revealed)}
				>
					{#if revealed}<EyeOff size={18} />{:else}<Eye size={18} />{/if}
				</button>
				<button class="btn secondary copy" onclick={copyInvite}>
					<Copy size={16} /> Copy
				</button>
			</div>
			<button class="btn share" onclick={shareInvite}>
				<Share2 size={16} />
				{linkCopied ? 'Link copied!' : 'Share invite link'}
			</button>
		</section>
	{/if}

	{#if invite && invite !== 'GLOBAL' && inviteAdmin}
		<section class="card invite-manager">
			<div class="invite-head">
				<h3><Mail size={17} /> Invite people</h3>
				<p class="muted small">Send an in-app request to a registered user.</p>
			</div>

			<label class="field invite-search">
				<span class="muted small">Search users</span>
				<span class="search-shell">
					<Search size={16} />
					<input
						class="input"
						bind:value={inviteQuery}
						placeholder="Name"
						autocomplete="off"
					/>
				</span>
			</label>

			{#if inviteSearchBusy}
				<p class="muted small invite-note">Searching...</p>
			{:else if !inviteSearchBusy && inviteCandidates.length === 0}
				<p class="muted small invite-note">No available users found.</p>
			{/if}

			{#if inviteCandidates.length > 0}
				<div class="candidate-list">
					{#each inviteCandidates as candidate (candidate.id)}
						<div class="candidate-row">
							<Avatar name={candidate.name} src={candidate.avatarUrl} size={38} />
							<span class="candidate-main">
								<b>{candidate.name}</b>
							</span>
							<button
								class="btn secondary invite-person"
								disabled={!!inviteSendBusy}
								onclick={() => sendInvite(candidate)}
							>
								<UserPlus size={16} /> {inviteSendBusy === candidate.id ? 'Sending...' : 'Invite'}
							</button>
						</div>
					{/each}
				</div>
			{/if}

			{#if pendingInvites.length > 0}
				<div class="pending-list">
					<p class="kicker">Pending</p>
					{#each pendingInvites as pending (pending.id)}
						<div class="pending-row">
							<Avatar name={pending.invitedUser.name} src={pending.invitedUser.avatarUrl} size={34} />
							<span>
								<b>{pending.invitedUser.name}</b>
							</span>
							<em>{inviteDate(pending.created)}</em>
						</div>
					{/each}
				</div>
			{/if}

			{#if inviteError}<p class="error">{inviteError}</p>{/if}
		</section>
	{/if}

	{#if isAdmin}
		<section class="card settings-zone">
			<h3>League settings</h3>

			<label class="toggle-row">
				<span class="toggle-label">
					<b>Hide forecasts from members</b>
					<span class="muted small">
						When on, members cannot view each other's pre-tournament bracket predictions.
					</span>
				</span>
				<button
					class="toggle-btn"
					class:on={hideForecast}
					disabled={settingsBusy}
					aria-pressed={hideForecast}
					aria-label="Toggle hide forecasts"
					onclick={toggleHideForecast}
				>
					<span class="toggle-thumb"></span>
				</button>
			</label>

			{#if settingsError}<p class="error">{settingsError}</p>{/if}

		{#if invite !== 'GLOBAL'}
			<div class="rename-row">
				<label class="field rename-field">
					<span class="muted small">Rename league</span>
					<div class="rename-input-row">
						<input
							class="input"
							bind:value={renameName}
							placeholder={league.name}
							maxlength={80}
						/>
						<button
							class="btn secondary"
							disabled={renameBusy || !renameName.trim()}
							onclick={renameLeague}
						>
							{renameBusy ? 'Saving…' : 'Rename'}
						</button>
					</div>
				</label>
				{#if renameError}<p class="error">{renameError}</p>{/if}
				{#if renameSuccess}<p class="success-msg">League renamed.</p>{/if}
			</div>
		{/if}
	</section>
	{/if}

	{#if role === 'owner' && invite !== 'GLOBAL'}
		<section class="card danger-zone">
			<h3>Delete league</h3>
			<p class="muted">
				This permanently deletes the league and removes all memberships. Type the league name to confirm.
			</p>
			<label class="field">
				<span class="muted small">Type {league.name}</span>
				<input class="input" bind:value={deleteConfirm} placeholder={league.name} />
			</label>
			<button
				class="btn danger"
				disabled={deleteBusy || deleteConfirm.trim() !== league.name.trim()}
				onclick={deleteLeague}
			>
				{deleteBusy ? 'Deleting…' : 'Delete league permanently'}
			</button>
			{#if deleteError}<p class="error">{deleteError}</p>{/if}
		</section>
	{/if}

	{#if cfg}
		<details class="card legend">
			<summary>How points work</summary>

			<h4>Per match (match tips) — max {cfg.match.tendency +
					cfg.match.exact +
					cfg.match.totalGoals +
					cfg.match.goalDiff} p</h4>
			<ul class="leg">
				<li>
					<span>Correct result - group stage: H / D / A; knockout: the team that advances</span><b>{cfg.match.tendency} p</b>
				</li>
				<li><span>Exact score</span><b>+{cfg.match.exact} p</b></li>
				<li><span>Correct total goals</span><b>+{cfg.match.totalGoals} p</b></li>
				<li><span>Correct goal difference</span><b>+{cfg.match.goalDiff} p</b></li>
			</ul>
			<p class="muted small">
				Knockout matches cannot end in a draw - the result points go to the team that advances. If a knockout match is decided in extra time, the score after extra time is used for points.
			</p>

			<h4>Forecast for the tournament</h4>
			<ul class="leg">
				<li><span>Each team in the correct group position</span><b>{cfg.forecast.groupPosition} p</b></li>
				<li><span>The full group in the correct order (bonus)</span><b>+{cfg.forecast.perfectGroupBonus} p</b></li>
				<li>
					<span>Each team you picked to advance (top 2 in a group or a best third) that actually goes through</span
					><b>{cfg.forecast.advance} p</b>
				</li>
			</ul>
			<p class="muted small">
				Reached knockout round (per correctly predicted team):
			</p>
			<ul class="leg">
				{#each Object.entries(roundLabel) as [k, lbl] (k)}
					{#if cfg.forecast.round[k] != null}
						<li><span>{lbl}</span><b>{cfg.forecast.round[k]} p</b></li>
					{/if}
				{/each}
			</ul>

			<h4>Tiebreakers (in order)</h4>
			<ol class="tiebreak">
				{#each cfg.tiebreakers as t (t)}
					<li>{tbLabel[t] ?? t}</li>
				{/each}
			</ol>
		</details>
	{/if}
{/if}

<style>
	.paid-toggle {
		all: unset;
		cursor: pointer;
		font-size: 1rem;
		line-height: 1;
		opacity: 0.2;
		transition: opacity 0.15s;
	}
	.paid-toggle.active {
		opacity: 1;
	}
	.back {
		display: inline-block;
		margin: 0.5rem 0 0.75rem;
	}
	h1 {
		margin: 0 0 1rem;
	}
	.irow {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.share {
		margin-top: 0.85rem;
	}
	.invite-head {
		display: grid;
		gap: 0.25rem;
		margin-bottom: 0.9rem;
	}
	.invite-head h3,
	.invite-head p {
		margin: 0;
	}
	.invite-head h3 {
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}
	.invite-manager {
		display: grid;
		gap: 0.75rem;
	}
	.invite-search {
		display: grid;
		gap: 0.35rem;
	}
	.search-shell {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: 0.45rem;
		padding: 0 0.75rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		background: var(--surface);
		color: var(--muted);
	}
	.search-shell .input {
		border: 0;
		padding-left: 0;
		background: transparent;
	}
	.search-shell .input:focus {
		outline: none;
	}
	.invite-note {
		margin: -0.25rem 0 0;
	}
	.candidate-list,
	.pending-list {
		display: grid;
		gap: 0.5rem;
	}
	.candidate-row,
	.pending-row {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.65rem;
		padding: 0.65rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		background: var(--surface);
	}
	.candidate-main,
	.pending-row span {
		display: grid;
		gap: 0.15rem;
		min-width: 0;
	}
	.candidate-main b,
	.pending-row b {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.pending-row em {
		color: var(--muted);
		font-size: 0.78rem;
		font-style: normal;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.invite-person {
		width: auto;
		padding: 0.6rem 0.75rem;
	}
	.pending-list .kicker {
		margin: 0.25rem 0 0;
	}
	.settings-zone h3 {
		margin: 0 0 1rem;
	}
	.rename-row {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
		display: grid;
		gap: 0.35rem;
	}
	.rename-field {
		display: grid;
		gap: 0.35rem;
	}
	.rename-input-row {
		display: flex;
		gap: 0.5rem;
	}
	.rename-input-row .input {
		flex: 1;
	}
	.rename-input-row .btn {
		flex: none;
		width: auto;
	}
	.success-msg {
		color: var(--success);
		font-size: 0.85rem;
		margin: 0;
	}
	.toggle-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		cursor: pointer;
	}
	.toggle-label {
		display: grid;
		gap: 0.2rem;
		flex: 1;
	}
	.toggle-label b {
		font-weight: 600;
	}
	.toggle-btn {
		flex: none;
		position: relative;
		width: 44px;
		height: 26px;
		border-radius: 999px;
		border: 2px solid var(--border);
		background: var(--surface-2);
		cursor: pointer;
		transition: background 0.18s, border-color 0.18s;
		padding: 0;
	}
	.toggle-btn.on {
		background: var(--accent);
		border-color: var(--accent);
	}
	.toggle-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.toggle-thumb {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: var(--text);
		transition: transform 0.18s;
	}
	.toggle-btn.on .toggle-thumb {
		transform: translateX(18px);
		background: var(--bg);
	}
	.danger-zone {
		border-color: color-mix(in srgb, var(--danger) 35%, var(--border));
		background: color-mix(in srgb, var(--danger) 6%, var(--surface-1));
	}
	.danger-zone h3 {
		margin: 0 0 0.4rem;
		color: var(--danger);
	}
	.danger-zone .field {
		display: grid;
		gap: 0.35rem;
		margin: 0.9rem 0;
	}
	.btn.danger {
		background: var(--danger);
		color: var(--bg);
		border-color: var(--danger);
	}
	.btn.danger:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
	.ic {
		min-width: 0;
	}
	.small {
		font-size: 0.8rem;
	}
	.code {
		font-family: var(--font-mono);
		font-weight: 700;
		letter-spacing: 0.2em;
		font-size: 1.3rem;
	}
	.code.masked {
		color: var(--muted);
		letter-spacing: 0.15em;
	}
	.eye {
		width: auto;
		padding: 0.7rem;
	}
	.copy {
		width: auto;
	}
	.tabs {
		display: flex;
		gap: 0.4rem;
		margin-bottom: 0.75rem;
	}
	.tabs button {
		flex: 1;
		padding: 0.5rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--muted);
		font-weight: 600;
	}
	.tabs button.active {
		color: var(--bg);
		background: var(--text);
		border-color: var(--text);
	}
	.lb {
		width: 100%;
		border-collapse: collapse;
	}
	.lb th,
	.lb td {
		text-align: left;
		padding: 0.6rem 0.4rem;
		border-bottom: 1px solid var(--border);
	}
	.lb th {
		color: var(--muted);
		font-size: 0.8rem;
		font-weight: 600;
	}
	.num {
		text-align: right;
	}
	.rank {
		width: 2rem;
		color: var(--muted);
		font-family: var(--font-mono);
	}
	.delta {
		display: block;
		font-size: 0.6rem;
		font-weight: 700;
		line-height: 1;
		margin-top: 0.15rem;
		letter-spacing: 0.01em;
	}
	.delta.up { color: var(--success); }
	.delta.dn { color: var(--danger); }
	tr.lead td {
		background: color-mix(in srgb, var(--accent) 9%, transparent);
	}
	tr.lead .rank {
		color: var(--accent);
		font-weight: 800;
	}
	.lb th.num,
	.lb td.num {
		text-align: right;
	}

	/* Pts is the focus — set it apart from the stat columns. */
	.lb th.pts,
	.lb td.pts {
		padding-left: 1.15rem;
		border-left: 1px solid var(--border);
		font-size: 1.02rem;
	}
	.lb th.pts {
		font-size: 0.8rem;
	}

	/* Extra tiebreaker columns: desktop only. */
	.ext {
		display: none;
	}
	.player {
		width: 100%;
		min-width: 0;
	}
	.pwrap {
		display: flex;
		align-items: center;
		gap: 0.55rem;
		min-width: 0;
		width: 100%;
	}
	.pname {
		flex: 1;
		min-width: 0;
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.fclink {
		display: inline-grid;
		place-items: center;
		color: var(--muted);
		flex: none;
	}
	.fclink:hover {
		color: var(--accent);
	}
	:global(.lb .rx) {
		color: var(--muted);
		transition: transform 0.15s ease;
		margin-left: auto;
		flex: none;
	}
	tr.main.open :global(.rx) {
		transform: rotate(180deg);
	}
	tr.main {
		cursor: pointer;
	}
	.detail td {
		padding: 0 0.4rem 0.7rem;
	}
	.stats {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.4rem 1rem;
	}
	.stats span {
		display: flex;
		justify-content: space-between;
		gap: 0.6rem;
		padding: 0.35rem 0;
		border-bottom: 1px solid var(--border);
	}
	.stats i {
		color: var(--muted);
		font-style: normal;
		font-size: 0.85rem;
	}
	.stats b {
		font-family: var(--font-mono);
	}

	@media (min-width: 760px) {
		.ext {
			display: table-cell;
		}
		:global(.lb .rx) {
			display: none;
		}
		tr.main {
			cursor: default;
		}
		.detail {
			display: none;
		}
	}
	@media (max-width: 759px) {
		.lb {
			table-layout: fixed;
		}
		.lb th:first-child,
		.lb td.rank {
			width: 3rem;
		}
		.lb th,
		.lb td {
			padding-left: 0.3rem;
			padding-right: 0.3rem;
		}
		.lb th.pts,
		.lb td.pts {
			width: 4rem;
			padding-left: 0.6rem;
		}
	}
	@media (max-width: 560px) {
		.stats {
			grid-template-columns: 1fr;
		}
		.candidate-row,
		.pending-row {
			grid-template-columns: auto minmax(0, 1fr);
		}
		.invite-person,
		.pending-row em {
			grid-column: 1 / -1;
		}
		.invite-person {
			width: 100%;
		}
	}
	@media (max-width: 360px) {
		.pwrap {
			gap: 0.35rem;
		}
		.fclink,
		:global(.lb .rx) {
			display: none;
		}
	}
	.note {
		margin: 0.75rem 0 0;
	}
	.legend summary {
		cursor: pointer;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		font-size: 0.85rem;
		color: var(--accent);
	}
	.legend h4 {
		margin: 1rem 0 0.5rem;
		font-size: 0.95rem;
	}
	.legend .small {
		margin: 0.4rem 0 0;
	}
	ul.leg {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	ul.leg li {
		display: flex;
		align-items: baseline;
		gap: 0.75rem;
		padding: 0.4rem 0;
		border-bottom: 1px solid var(--border);
	}
	ul.leg li:last-child {
		border-bottom: none;
	}
	ul.leg li span {
		flex: 1;
	}
	ul.leg li b {
		font-family: var(--font-mono);
		color: var(--accent);
		white-space: nowrap;
	}
	ol.tiebreak {
		margin: 0.5rem 0 0;
		padding-left: 1.3rem;
		line-height: 1.8;
	}
	ol.tiebreak li {
		padding-left: 0.3rem;
	}
</style>
