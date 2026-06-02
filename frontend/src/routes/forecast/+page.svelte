<script lang="ts">
	import DeadlineCountdown from '$lib/components/DeadlineCountdown.svelte';
	import { forecastStore as fs, koKey, type KOMatch } from '$lib/forecast.svelte';
	import Flag from '$lib/components/Flag.svelte';
	import { vibrate } from '$lib/haptics';
	import { flip } from 'svelte/animate';
	import { teamDisplayName } from '$lib/teamNames';
	import {
		ChevronUp,
		ChevronDown,
		Lock,
		Check,
		CircleCheck,
		X,
		Trophy
	} from '@lucide/svelte';
	import { collapseOnScroll } from '$lib/actions';
	import { stageName as knockoutStageName } from '$lib/stageLabels';

	let section = $state<'groups' | 'thirds' | 'bracket'>('groups');
	let saveState = $state<'idle' | 'saving' | 'saved' | 'error'>('idle');
	let err = $state('');

	$effect(() => {
		if (!fs.loaded) fs.load().catch((e) => (err = e?.message ?? 'Load failed'));
	});

	// Debounced autosave. The Forecast is a living prediction edited until
	// lock, so changes persist automatically ~1s after the last edit.
	let primed = false;
	let timer: ReturnType<typeof setTimeout>;
	$effect(() => {
		// Track every part of the prediction.
		const snapshot = JSON.stringify([
			fs.groupOrder,
			fs.thirds,
			fs.bracket
		]);
		if (!fs.loaded || fs.locked) return;
		if (!primed) {
			primed = true; // skip the initial hydrate
			return;
		}
		void snapshot;
		clearTimeout(timer);
		timer = setTimeout(async () => {
			saveState = 'saving';
			err = '';
			try {
				await fs.save();
				saveState = 'saved';
			} catch (e: unknown) {
				saveState = 'error';
				err =
					(e as { message?: string })?.message ??
					'Could not save — your changes were not saved.';
			}
		}, 1000);
		return () => clearTimeout(timer);
	});

	const stages = ['R32', 'R16', 'QF', 'SF', '3RD', 'FINAL'];
	let byStage = $derived(
		stages.map((s) => ({
			stage: s,
			matches: fs.knockout.filter((m) => m.stage === s)
		}))
	);

	let finalMatch = $derived(fs.knockout.find((m) => m.stage === 'FINAL'));
	let champion = $derived(
		finalMatch ? fs.bracket[koKey(finalMatch)] : ''
	);
	let actualThirds = $derived(fs.actualBestThirds());

	function tname(id: string) {
		return teamDisplayName(fs.team(id));
	}
	const ord = (n: number) =>
		n === 1 ? '1.' : n === 2 ? '2.' : n === 3 ? '3.' : `${n}.`;

	function sideLabel(m: KOMatch, side: 'home' | 'away') {
		const [h, a] = fs.sides(m);
		const id = side === 'home' ? h : a;
		if (id) return { id, name: tname(id), team: fs.team(id) };
		return {
			id: '',
			name: side === 'home' ? m.homeLabel : m.awayLabel,
			team: undefined
		};
	}
</script>

<div class="stickyhead" use:collapseOnScroll>
	<p class="kicker">Whole tournament</p>
	<div class="sh-expand">
		<div class="sh-inner">
			<h1>Forecast</h1>
			<p class="muted desc">
				Your Forecast for groups, best thirds, and the road to the final.
				{#if fs.locked}<b>Locked.</b
						>{:else}Locks at kickoff.{/if}
			</p>
				{#if !fs.locked && fs.tournamentStart}
					<DeadlineCountdown
						deadline={fs.tournamentStart}
						label="Locks"
						compact
					/>
				{/if}
		</div>
	</div>
	{#if fs.loaded}
		<div class="seg">
			<button class:on={section === 'groups'} onclick={() => (section = 'groups')}>Groups</button>
			<button class:on={section === 'thirds'} onclick={() => (section = 'thirds')}>Best thirds</button>
			<button class:on={section === 'bracket'} onclick={() => (section = 'bracket')}>Knockout</button>
		</div>
	{/if}
</div>

{#if err}<p class="error">{err}</p>{/if}

{#if !fs.loaded}
	<p class="muted">Loading…</p>
{:else}
	{#if fs.locked}
		<div class="card lockbar"><Lock size={16} /> The tournament has started - the Forecast is final.</div>
	{/if}

	{#if section === 'groups'}
		<p class="muted small">
			Rank each group from 1st to 4th. The top 2 advance; 3rd place can advance as a best third.
		</p>
		{#each fs.groups as g (g.letter)}
			<section class="card grp">
				<h3>Group {g.letter}</h3>
				{#each fs.groupOrder[g.letter] as id, i (id)}
					{@const ao = fs.actualOrder(g.letter)}
					{@const apos = ao ? ao.indexOf(id) + 1 : 0}
					{@const exact = ao ? ao[i] === id : null}
					{@const advanced =
						ao &&
						(apos <= 2 ||
							(apos === 3 && (actualThirds?.has(id) ?? false)))}
					{@const scoredAdv =
						advanced && (i < 2 || (i === 2 && !!fs.thirds[g.letter]))}
					{@const state =
						exact === null
							? 'pending'
							: exact
								? 'ok'
								: scoredAdv
									? 'half'
									: 'miss'}
					{@const rank = fs.team(id)?.fifaRanking}
					<div
						class="trow"
						class:rwin={state === 'ok'}
						class:rhalf={state === 'half'}
						class:rmiss={state === 'miss'}
						animate:flip={{ duration: 240 }}
					>
						<span class="pos">{i + 1}</span>
						<Flag iso2={fs.team(id)?.iso2 ?? ''} code={fs.team(id)?.fifaCode ?? ''} />
						<span class="nmwrap">
							<span class="nm">{tname(id)}</span>
							{#if rank}<span class="wpct">#{rank}</span>{/if}
						</span>
						<span class="tag">
							{#if state === 'ok'}<span class="ind ok"><Check size={15} /></span>
							{:else if state === 'half'}
								<span class="apos half">actual {ord(apos)}</span>
								<span class="ind half"><CircleCheck size={15} /></span>
							{:else if state === 'miss'}
								<span class="apos">actual {ord(apos)}</span>
								<span class="ind no"><X size={15} /></span>
								{:else if i < 2}<span class="pill ok">through</span>
								{:else if i === 2}<span class="pill">3rd place</span>{/if}
						</span>
						{#if !fs.locked}
							<span class="ord">
								<button aria-label="Move up" disabled={i === 0} onclick={() => { fs.move(g.letter, i, -1); vibrate(15); }}><ChevronUp size={16} /></button>
								<button aria-label="Move down" disabled={i === 3} onclick={() => { fs.move(g.letter, i, 1); vibrate(15); }}><ChevronDown size={16} /></button>
							</span>
						{/if}
					</div>
				{/each}
			</section>
		{/each}
	{:else if section === 'thirds'}
		<div class="thead">
			<p class="muted small">
				Choose the 8 of 12 group thirds you think will advance. The teams come from your group rankings.
			</p>
			<span class="cnt" class:full={fs.chosenThirdLetters.length === 8}>
				{fs.chosenThirdLetters.length} / 8
			</span>
		</div>
		<section class="card tlist">
			{#each fs.groups as g (g.letter)}
				{@const tid = fs.groupThird(g.letter)}
				{@const on = !!fs.thirds[g.letter]}
				{@const adv = actualThirds ? actualThirds.has(tid) : null}
				<label class="trow" class:on>
					<input
						type="checkbox"
						checked={on}
						disabled={fs.locked ||
							(!on && fs.chosenThirdLetters.length >= 8)}
						onchange={() => fs.toggleThird(g.letter)}
					/>
					<span class="gl">{g.letter}</span>
					<Flag iso2={fs.team(tid)?.iso2 ?? ''} code={fs.team(tid)?.fifaCode ?? ''} />
					<span class="nm">{tname(tid) || '—'}</span>
					<span class="spacer"></span>
					{#if on && adv === true}<span class="ind ok"><Check size={15} /></span>
					{:else if on && adv === false}<span class="ind no"><X size={15} /></span>
					{:else if adv === true}<span class="ind dim"><Check size={14} /></span>{/if}
				</label>
			{/each}
		</section>
	{:else}
		{#if champion}
			<div class="card champ">
				<Trophy size={20} />
				<span class="lbl">Predicted winner</span>
				<Flag
					iso2={fs.team(champion)?.iso2 ?? ''}
					code={fs.team(champion)?.fifaCode ?? ''}
					size={26}
				/>
				<b>{tname(champion)}</b>
			</div>
		{/if}
		{#each byStage as col (col.stage)}
			<h3 class="rname">{knockoutStageName(col.stage)}</h3>
			{#each col.matches as m (koKey(m))}
				{@const H = sideLabel(m, 'home')}
				{@const A = sideLabel(m, 'away')}
				{@const w = fs.bracket[koKey(m)]}
				{@const actAdv =
					m.num > 0
						? fs.advancerOf(m.num)
						: (fs.results.find(
								(r) => r.stage === m.stage && r.finished
							)?.advancer ?? '')}
				{@const bok = actAdv ? w === actAdv : null}
				<div class="bm card" class:rwin={bok === true} class:rmiss={bok === false}>
					<button
						class="bteam"
						class:win={w && w === H.id}
						disabled={fs.locked || !H.id}
						onclick={() => fs.pick(m, H.id)}
					>
						{#if H.team}<Flag iso2={H.team.iso2} code={H.team.fifaCode} />{/if}
						<span class="bn" class:ph={!H.id}>{H.name}</span>
					</button>
					<span class="vs">vs</span>
					<button
						class="bteam"
						class:win={w && w === A.id}
						disabled={fs.locked || !A.id}
						onclick={() => fs.pick(m, A.id)}
					>
						{#if A.team}<Flag iso2={A.team.iso2} code={A.team.fifaCode} />{/if}
						<span class="bn" class:ph={!A.id}>{A.name}</span>
					</button>
					{#if bok === true}<span class="ind ok"><Check size={15} /></span>
					{:else if bok === false}<span class="ind no"><X size={15} /></span>{/if}
				</div>
			{/each}
		{/each}
	{/if}

	{#if !fs.locked}
		<div class="savebar">
			<span class="savestat" class:err={saveState === 'error'}>
				{#if saveState === 'saving'}
						Saving…
				{:else if saveState === 'error'}
						{err || 'Save failed'}
				{:else if saveState === 'saved'}
						<Check size={15} /> Saved · changes are saved automatically
				{:else}
						Changes are saved automatically
				{/if}
			</span>
		</div>
	{/if}
{/if}

<style>
	h1 {
		margin: 0.25rem 0 0.2rem;
	}
	.small {
		font-size: 0.85rem;
	}
	.lockbar {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--warning);
	}
	.stickyhead {
		position: sticky;
		top: var(--topbar-h);
		z-index: 20;
		margin: 0 -1rem;
		padding: 0.6rem 1rem 0.75rem;
		background: var(--bg);
		border-bottom: 1px solid var(--border);
	}
	.stickyhead h1 {
		margin: 0.1rem 0 0;
	}
	.stickyhead .desc {
		margin: 0.3rem 0 0;
		font-size: 0.9rem;
	}
	@media (min-width: 900px) {
		.stickyhead {
			top: 0;
			margin: 0 -2rem;
			padding: 0.75rem 2rem 0.85rem;
		}
	}
	.grp h3 {
		margin: 0 0 0.6rem;
	}
	.trow {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.45rem 0;
		border-top: 1px solid var(--border);
	}
	.trow:nth-child(2) {
		border-top: none;
	}
	.pos {
		width: 1.2rem;
		text-align: center;
		font-weight: 800;
		color: var(--muted);
	}
	.nmwrap {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 0.35rem;
		overflow: hidden;
	}
	.nm {
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.wpct {
		font-size: 0.72rem;
		color: var(--muted);
		font-family: var(--font-mono);
		white-space: nowrap;
	}
	.pill.ok {
		color: var(--success);
		border-color: var(--success);
	}
	.ord button {
		background: var(--surface-2);
		border: 1px solid var(--border);
		color: var(--text);
		border-radius: 7px;
		width: 30px;
		height: 26px;
		margin-left: 2px;
	}
	.ord button:disabled {
		color: var(--muted);
		opacity: 0.5;
	}
	.rname {
		margin: 1.2rem 0 0.5rem;
		color: var(--muted);
		font-size: 0.95rem;
	}
	.bm {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.7rem;
	}
	.bm + .bm {
		margin-top: 0.5rem;
	}
	.bteam {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.55rem 0.6rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-sm);
		color: var(--text);
		min-width: 0;
	}
	.bteam:disabled {
		opacity: 0.7;
	}
	.bteam.win {
		background: var(--text);
		border-color: var(--text);
		color: var(--bg);
	}
	.bn {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-weight: 600;
		font-size: 0.9rem;
	}
	.bn.ph {
		color: var(--muted);
		font-weight: 500;
	}
	.vs {
		color: var(--muted);
		font-size: 0.8rem;
	}
	.champ {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		color: var(--gold);
		border-color: var(--border-strong);
		background: var(--surface);
		text-shadow: none;
	}
	.champ .lbl {
		text-transform: uppercase;
		letter-spacing: 0.14em;
		font-size: 0.78rem;
		font-weight: 700;
	}
	.champ b {
		font-family: var(--font-display);
		font-size: 1.15rem;
		letter-spacing: 0.02em;
	}
	.savebar {
		position: sticky;
		bottom: calc(var(--nav-h) + 0.5rem);
		display: flex;
		justify-content: center;
		margin-top: 1.5rem;
		pointer-events: none;
	}
	.savestat {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		font-size: 0.8rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--muted);
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		padding: 0.4rem 0.85rem;
	}
	.savestat.err {
		color: var(--danger);
		border-color: var(--danger);
		text-transform: none;
		letter-spacing: 0;
	}
	.thead {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		margin-bottom: 0.6rem;
	}
	.thead .small {
		flex: 1;
	}
	.cnt {
		font-family: var(--font-mono);
		font-weight: 700;
		padding: 0.2rem 0.6rem;
		border-radius: var(--radius-pill);
		border: 1px solid var(--border);
		color: var(--muted);
		white-space: nowrap;
	}
	.cnt.full {
		color: var(--bg);
		background: var(--text);
		border-color: var(--text);
	}
	.tlist {
		padding: 0.3rem 0.9rem;
	}
	.trow {
		display: flex;
		align-items: center;
		gap: 0.7rem;
		padding: 0.8rem 0;
		min-height: 44px;
		border-top: 1px solid var(--border);
		cursor: pointer;
	}
	.trow:first-child {
		border-top: none;
	}
	.trow input {
		width: 20px;
		height: 20px;
		accent-color: var(--accent);
	}
	.trow .gl {
		display: grid;
		place-items: center;
		width: 24px;
		height: 24px;
		border-radius: 6px;
		background: var(--surface-2);
		font-family: var(--font-display);
		font-size: 0.85rem;
		color: var(--muted);
	}
	.trow.on {
		color: var(--text);
	}
	.trow.on .gl {
		background: var(--text);
		color: var(--bg);
	}
	.trow .nm {
		font-weight: 600;
	}
	.ind {
		display: inline-grid;
		place-items: center;
	}
	.ind.ok {
		color: var(--success);
	}
	.ind.no {
		color: var(--danger);
	}
	.ind.half {
		color: var(--gold);
	}
	.apos.half {
		color: var(--gold);
	}
	.tag {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
	}
	.apos {
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--muted);
	}
	.ind.dim {
		color: var(--muted);
		opacity: 0.7;
	}
	.trow.rwin,
	.bm.rwin {
		border-color: color-mix(in srgb, var(--success) 45%, var(--border));
	}
	.trow.rhalf {
		border-color: color-mix(in srgb, var(--gold) 45%, var(--border));
	}
	.trow.rmiss,
	.bm.rmiss {
		border-color: color-mix(in srgb, var(--danger) 40%, var(--border));
	}
	.bm.rwin,
	.bm.rmiss {
		border-style: solid;
	}
	:global(:root[data-theme='worldcup']) .grp,
	:global(:root[data-theme='worldcup']) .tlist,
	:global(:root[data-theme='worldcup']) .bm,
	:global(:root[data-theme='worldcup']) .champ,
	:global(:root[data-theme='worldcup']) .lockbar {
		background:
			radial-gradient(circle at 14% 0%, rgba(143, 197, 143, 0.075), transparent 32%),
			linear-gradient(180deg, rgba(13, 34, 40, 0.96), rgba(7, 17, 25, 0.98)),
			var(--surface);
		border-color: color-mix(in srgb, var(--accent) 12%, var(--border));
		box-shadow: 0 16px 42px -34px rgba(0, 0, 0, 0.9), inset 0 1px 0 rgba(255, 255, 255, 0.035);
	}
	:global(:root[data-theme='worldcup']) .grp::before,
	:global(:root[data-theme='worldcup']) .tlist::before,
	:global(:root[data-theme='worldcup']) .bm::before,
	:global(:root[data-theme='worldcup']) .champ::before,
	:global(:root[data-theme='worldcup']) .lockbar::before {
		display: none;
	}
	:global(:root[data-theme='worldcup']) .trow {
		border-top-color: color-mix(in srgb, var(--accent) 11%, var(--border));
	}
	:global(:root[data-theme='worldcup']) .trow .gl,
	:global(:root[data-theme='worldcup']) .bteam,
	:global(:root[data-theme='worldcup']) .cnt,
	:global(:root[data-theme='worldcup']) .savestat {
		background: color-mix(in srgb, var(--surface-2) 78%, transparent);
		border-color: color-mix(in srgb, var(--accent) 12%, var(--border));
	}
	:global(:root[data-theme='worldcup']) .bteam.win,
	:global(:root[data-theme='worldcup']) .trow.on .gl,
	:global(:root[data-theme='worldcup']) .cnt.full {
		background: linear-gradient(180deg, color-mix(in srgb, var(--accent) 42%, var(--surface-2)), var(--surface-2));
		border-color: color-mix(in srgb, var(--accent) 36%, var(--border));
		color: var(--text);
	}
</style>
