<script lang="ts">
	import { pb } from '$lib/pb';
	import { serverClock } from '$lib/serverclock.svelte';
	import { api, type LeagueSummary } from '$lib/api';
	import { language } from '$lib/language.svelte';

	let when = $state('');
	let busy = $state(false);
	let msg = $state('');
	let msgTone = $state<'ok' | 'error'>('error');

	let botCount = $state(3);
	let botLeague = $state('');
	let chatCount = $state(6);
	let chatLeague = $state('');
	let leagues = $state<LeagueSummary[]>([]);
	const isEnglish = $derived(language.isEnglish);

	$effect(() => {
		if (serverClock.dev)
			api
				.myLeagues()
				.then(
					(r) =>
						(leagues = r.leagues.filter((league) => league.inviteCode !== 'GLOBAL'))
				)
				.catch(() => {});
	});

	async function genBots() {
		busy = true;
		msg = '';
		try {
			await pb.send('/api/dev/bots', {
				method: 'POST',
				body: { count: botCount, leagueId: botLeague }
			});
			location.reload();
		} catch (e: unknown) {
			msgTone = 'error';
			msg = (e as { message?: string })?.message ?? (isEnglish ? 'Failed' : 'Feila');
			busy = false;
		}
	}

	async function sendBotChat() {
		busy = true;
		msg = '';
		try {
			const result = await pb.send<{ sent: number }>('/api/dev/bot-chat', {
				method: 'POST',
				body: { count: chatCount, leagueId: chatLeague }
			});
			msgTone = 'ok';
			msg = isEnglish
				? `Sent ${result.sent} bot message${result.sent === 1 ? '' : 's'}.`
				: `Sendte ${result.sent} botmelding${result.sent === 1 ? '' : 'ar'}.`;
		} catch (e: unknown) {
			msgTone = 'error';
			msg = (e as { message?: string })?.message ?? (isEnglish ? 'Failed' : 'Feila');
		} finally {
			busy = false;
		}
	}

	$effect(() => {
		serverClock.refresh();
	});

	// Seed the input from the current sim time (or now).
	$effect(() => {
		if (!when) {
			const base = serverClock.simTime
				? new Date(serverClock.simTime)
				: new Date(serverClock.now());
			when = base.toISOString().slice(0, 16);
		}
	});

	const presets: { label: string; ts: string }[] = [
		{ label: 'opening', ts: '2026-06-11T20:00' },
		{ label: 'group-md2-live', ts: '2026-06-15T21:30' },
		{ label: 'after-groups', ts: '2026-06-25T06:00' },
		{ label: 'after-r32', ts: '2026-07-04T06:00' },
		{ label: 'after-qf', ts: '2026-07-12T06:00' },
		{ label: 'after-final', ts: '2026-07-20T00:00' }
	];

	function presetLabel(label: string) {
		const labels: Record<string, [string, string]> = {
			opening: ['Opningskamp', 'Opening match'],
			'group-md2-live': ['Gruppe MD2 live', 'Group MD2 live'],
			'after-groups': ['Etter gruppene', 'After groups'],
			'after-r32': ['Etter 32-delsfinalar', 'After R32'],
			'after-qf': ['Etter kvartfinalar', 'After QF'],
			'after-final': ['Etter finalen', 'After final']
		};
		const [nn, en] = labels[label] ?? [label, label];
		return isEnglish ? en : nn;
	}

	async function advance(ts: string) {
		busy = true;
		msg = '';
		try {
			await pb.send('/api/dev/advance', {
				method: 'POST',
				body: { timestamp: ts }
			});
			location.reload(); // re-pull all stores against the new clock
		} catch (e: unknown) {
			msgTone = 'error';
			msg = (e as { message?: string })?.message ?? (isEnglish ? 'Failed' : 'Feila');
			busy = false;
		}
	}

	async function reset() {
		busy = true;
		msg = '';
		try {
			await pb.send('/api/dev/reset', { method: 'POST', body: {} });
			location.reload();
		} catch (e: unknown) {
			msgTone = 'error';
			msg = (e as { message?: string })?.message ?? (isEnglish ? 'Failed' : 'Feila');
			busy = false;
		}
	}
</script>

<p class="kicker">{isEnglish ? 'Test harness' : 'Testverktøy'}</p>
<h1>{isEnglish ? 'Dev tools' : 'Utviklarverktøy'}</h1>

{#if !serverClock.loaded}
	<p class="muted">…</p>
{:else if !serverClock.dev}
	<section class="card">
		<p class="muted">
			{isEnglish ? 'Disabled. Start the server with' : 'Avslått. Start serveren med'} <code>WMP_DEV=1</code>
			{isEnglish ? 'to simulate the tournament.' : 'for å simulere turneringa.'}
		</p>
	</section>
{:else}
	<section class="card">
		<div class="state">
			<span class="kicker">{isEnglish ? 'Simulated clock' : 'Simulert klokke'}</span>
			<b class="digits"
				>{serverClock.simulated
					? new Date(serverClock.now()).toLocaleString()
					: isEnglish ? 'live (real time)' : 'live (sanntid)'}</b
			>
		</div>
	</section>

	<section class="card">
			<h3>{isEnglish ? 'Jump to' : 'Hopp til'}</h3>
		<p class="muted small">
				{isEnglish
					? 'Matches before this time are simulated (finished, or live if in the middle of the match); later matches are reset. Locks, friends\' match tips, and the Forecast deadline follow this clock.'
					: 'Kampar før dette tidspunktet blir simulerte (ferdige, eller live viss dei er midt i kampen); seinare kampar blir nullstilte. Låsing, venetips og VM-tipsfristen følgjer denne klokka.'}
		</p>
		<div class="field">
			<input class="input" type="datetime-local" bind:value={when} />
		</div>
		<button
			class="btn"
			disabled={busy || !when}
			onclick={() => advance(when)}>{isEnglish ? 'Advance' : 'Køyr fram'}</button
		>

		<div class="presets">
			{#each presets as p (p.ts)}
				<button
					class="chip"
					disabled={busy}
					onclick={() => advance(p.ts)}>{presetLabel(p.label)}</button
				>
			{/each}
		</div>
	</section>

	<section class="card">
		<h3>{isEnglish ? 'Generate bot players' : 'Lag bot-spelarar'}</h3>
		<p class="muted small">
			{isEnglish
				? 'Each bot gets a fully random Forecast and a match tip for every match, and joins the selected league (or all your private leagues) - a live leaderboard race.'
				: 'Kvar bot får eit heilt tilfeldig VM-tips og kamptips for kvar kamp, og blir med i vald liga (eller alle private ligaene dine) - eit live tabelløp.'}
		</p>
		<div class="field">
			<label for="bc">{isEnglish ? 'How many' : 'Kor mange'}</label>
			<input
				id="bc"
				class="input"
				type="number"
				min="1"
				max="20"
				bind:value={botCount}
			/>
		</div>
		<div class="field">
			<label for="bl">{isEnglish ? 'League' : 'Liga'}</label>
			<select id="bl" class="input" bind:value={botLeague}>
				<option value="">{isEnglish ? 'All my private leagues' : 'Alle dei private ligaene mine'}</option>
				{#each leagues as l (l.id)}
					<option value={l.id}>{l.name}</option>
				{/each}
			</select>
		</div>
		<button class="btn" disabled={busy} onclick={genBots}>
			{isEnglish ? `Generate ${botCount} bot${botCount === 1 ? '' : 's'}` : `Lag ${botCount} bot${botCount === 1 ? '' : 'ar'}`}
		</button>
	</section>

	<section class="card">
		<h3>{isEnglish ? 'Send bot chat' : 'Send bot-chat'}</h3>
		<p class="muted small">
			{isEnglish
				? 'Use existing test bots to post live messages into league chat. Generate bots first if the league has none. If no league is chosen, messages are sent in each of your private leagues that already has bots.'
				: 'Bruk eksisterande testbotar til å poste live meldingar i liga-chatten. Lag botar først viss ligaen ikkje har nokon. Viss du ikkje vel liga, blir meldingar sende i kvar av dei private ligaene dine som allereie har botar.'}
		</p>
		<div class="field">
			<label for="cc">{isEnglish ? 'How many messages' : 'Kor mange meldingar'}</label>
			<input
				id="cc"
				class="input"
				type="number"
				min="1"
				max="50"
				bind:value={chatCount}
			/>
		</div>
		<div class="field">
			<label for="cl">{isEnglish ? 'League' : 'Liga'}</label>
			<select id="cl" class="input" bind:value={chatLeague}>
				<option value="">{isEnglish ? 'All my private leagues with bots' : 'Alle private ligaene mine med botar'}</option>
				{#each leagues as l (l.id)}
					<option value={l.id}>{l.name}</option>
				{/each}
			</select>
		</div>
		<button class="btn" disabled={busy} onclick={sendBotChat}>
			{isEnglish ? `Send ${chatCount} bot message${chatCount === 1 ? '' : 's'}` : `Send ${chatCount} botmelding${chatCount === 1 ? '' : 'ar'}`}
		</button>
	</section>

	<section class="card">
		<h3>{isEnglish ? 'Reset' : 'Nullstill'}</h3>
		<p class="muted small">
			{isEnglish ? 'Clear all results and the simulated clock (back to real time).' : 'Tøm alle resultat og den simulerte klokka (tilbake til sanntid).'}
		</p>
		<button class="btn secondary" disabled={busy} onclick={reset}
			>{isEnglish ? 'Reset all' : 'Nullstill alt'}</button
		>
	</section>

	{#if msg}<p class:error={msgTone === 'error'} class:notice={msgTone === 'ok'}>{msg}</p>{/if}
{/if}

<style>
	h1 {
		margin: 0.1rem 0 1rem;
	}
	.small {
		font-size: 0.85rem;
	}
	.state {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}
	.state b {
		font-size: 1.2rem;
	}
	.presets {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-top: 0.9rem;
	}
	.chip {
		padding: 0.5rem 0.8rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		color: var(--text);
		font:
			700 0.78rem var(--font);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		cursor: pointer;
	}
	.chip:hover {
		border-color: var(--accent);
	}
	code {
		font-family: var(--font-mono);
		color: var(--accent);
	}
	.notice {
		color: var(--accent);
		font-size: 0.9rem;
	}
</style>
