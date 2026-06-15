<script lang="ts">
	import { page } from '$app/stores';
	import { tipsStore, type FriendTip, type LiveMatchScore } from '$lib/tips.svelte';
	import { teamDisplayName } from '$lib/teamNames';
	import { matchStageLabel } from '$lib/stageLabels';
	import Avatar from '$lib/components/Avatar.svelte';
	import Flag from '$lib/components/Flag.svelte';
	import { ArrowLeft } from '@lucide/svelte';

	let matchId = $derived($page.params.matchId ?? '');
	let tips = $state<FriendTip[]>([]);
	let liveScore = $state<LiveMatchScore | null>(null);
	let loaded = $state(false);
	let err = $state('');

	async function refresh(id: string) {
		try {
			const r = await tipsStore.friendsLive(id);
			tips = r.tips;
			liveScore = r.match;
		} catch (e) {
			err = (e as Error)?.message ?? 'Could not load this match.';
		} finally {
			loaded = true;
		}
	}

	$effect(() => {
		void tipsStore.load();
	});

	$effect(() => {
		const id = matchId;
		if (!id) return;
		loaded = false;
		void refresh(id);
		const t = setInterval(() => void refresh(id), 60_000);
		return () => clearInterval(t);
	});

	let match = $derived(tipsStore.matches.find((m) => m.id === matchId) ?? null);
	let isKO = $derived(!!match && match.stage !== 'group');

	// Prefer the freshly-polled score; fall back to the match record.
	let score = $derived<LiveMatchScore | null>(
		liveScore ??
			(match
				? {
						ftHome: match.ftHome,
						ftAway: match.ftAway,
						etHome: match.etHome,
						etAway: match.etAway,
						penHome: match.penHome,
						penAway: match.penAway,
						status: match.status,
						finalizedAt: match.finalizedAt
					}
				: null)
	);
	let finished = $derived(!!score && !!score.finalizedAt);

	function team(id: string) {
		return tipsStore.team(id);
	}
	function teamLabel(side: 'h' | 'a') {
		if (!match) return '';
		const id = side === 'h' ? match.homeTeam : match.awayTeam;
		const label = side === 'h' ? match.homeLabel : match.awayLabel;
		return teamDisplayName(team(id), label);
	}
	function code(id: string) {
		return team(id)?.fifaCode ?? '';
	}
	function scoreText(s: LiveMatchScore) {
		let t = `${s.ftHome}–${s.ftAway}`;
		if (s.etHome || s.etAway) t = `${s.etHome}–${s.etAway} aet`;
		if (s.penHome || s.penAway) t += ` (${s.penHome}–${s.penAway} pens)`;
		return t;
	}
	function tipText(f: FriendTip) {
		let t = `${f.ftHome}–${f.ftAway}`;
		if (f.etHome || f.etAway) t = `${f.etHome}–${f.etAway} aet`;
		if (f.penWinner) t += ` · ${code(f.penWinner)} on pens`;
		return t;
	}
	function pointsText(p: number) {
		return p > 0 ? `+${p}` : `${p}`;
	}

	let sortedTips = $derived(
		[...tips].sort((a, b) => b.points - a.points || a.name.localeCompare(b.name, 'en-US'))
	);
</script>

<a class="back" href="/"><ArrowLeft size={16} /> Home</a>

{#if match}
	<section class="card live-head">
		<div class="live-head-top">
			<span class="stage">{matchStageLabel(match)}</span>
			<span class="state" class:finished>{finished ? 'Full time' : 'Live'}</span>
		</div>
		<div class="matchup">
			<span class="side">
				{#if team(match.homeTeam)}
					<Flag iso2={team(match.homeTeam)?.iso2 ?? ''} code={code(match.homeTeam)} size={28} />
				{/if}
				<b>{teamLabel('h')}</b>
			</span>
			<strong class="big-score digits">{score ? scoreText(score) : 'vs'}</strong>
			<span class="side away">
				<b>{teamLabel('a')}</b>
				{#if team(match.awayTeam)}
					<Flag iso2={team(match.awayTeam)?.iso2 ?? ''} code={code(match.awayTeam)} size={28} />
				{/if}
			</span>
		</div>
	</section>
{/if}

<section class="card friends">
	<div class="hd">
		<h3>How everyone is doing</h3>
	</div>
	{#if !finished}
		<p class="muted note">
			Points below are provisional — they update as the score changes and only count
			toward the league table once the match is finished.
		</p>
	{/if}

	{#if err}
		<p class="muted">{err}</p>
	{:else if !loaded}
		<p class="muted">Loading…</p>
	{:else if sortedTips.length === 0}
		<p class="muted">No one in your leagues tipped this match.</p>
	{:else}
		<ul class="friend-list">
			{#each sortedTips as f (f.userId)}
				<li class:me={f.isMe}>
					<span class="who">
						<Avatar name={f.name} size={28} />
						<b>{f.name}</b>
						{#if f.isMe}<i>you</i>{/if}
					</span>
					<span class="their-tip digits">{tipText(f)}</span>
					<strong class="pts" class:plus={f.points > 0} class:zero={f.points === 0}>
						{pointsText(f.points)} pts
					</strong>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style>
	.back {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		color: var(--muted);
		font-weight: 600;
		margin-bottom: 0.9rem;
	}
	.live-head {
		display: grid;
		gap: 0.9rem;
		margin-bottom: 1rem;
	}
	.live-head-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}
	.stage {
		color: var(--muted);
		font-weight: 600;
		font-size: 0.85rem;
	}
	.state {
		font-size: 0.7rem;
		font-weight: 800;
		letter-spacing: 0.1em;
		color: #fff;
		background: var(--danger, #e5484d);
		padding: 0.2rem 0.5rem;
		border-radius: var(--radius-pill);
	}
	.state.finished {
		background: var(--muted);
	}
	.matchup {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
		align-items: center;
		gap: 1rem;
	}
	.side {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		min-width: 0;
	}
	.side.away {
		justify-content: flex-end;
	}
	.side b {
		font-size: 1.1rem;
		font-weight: 700;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.big-score {
		font-size: 1.8rem;
		font-weight: 800;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}
	.friends .hd h3 {
		font-size: 1.15rem;
		margin: 0 0 0.5rem;
	}
	.note {
		font-size: 0.85rem;
		margin: 0 0 0.8rem;
	}
	.friend-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: 0.4rem;
	}
	.friend-list li {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto auto;
		align-items: center;
		gap: 0.8rem;
		padding: 0.55rem 0.7rem;
		border: 1px solid var(--border);
		border-radius: var(--radius);
		background: var(--surface);
	}
	.friend-list li.me {
		border-color: color-mix(in srgb, var(--accent) 45%, var(--border));
		background: color-mix(in srgb, var(--accent) 5%, var(--surface));
	}
	.who {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-width: 0;
	}
	.who b {
		font-weight: 650;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.who i {
		font-style: normal;
		color: var(--muted);
		font-size: 0.78rem;
	}
	.their-tip {
		color: var(--muted);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}
	.pts {
		font-weight: 800;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		color: var(--muted);
	}
	.pts.plus {
		color: var(--success);
	}
</style>
