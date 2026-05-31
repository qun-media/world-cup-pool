<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { language } from '$lib/language.svelte';
	import {
		ArrowLeft,
		CheckCircle2,
		Clock,
		Info,
		ListChecks,
		Medal,
		Network,
		Telescope,
		Trophy,
		Users,
		Volleyball,
		X
	} from '@lucide/svelte';

	const isEnglish = $derived(language.isEnglish);

	let flow = $derived.by(() => [
		{
			icon: Telescope,
			title: isEnglish ? 'Forecast before kickoff' : 'VM-tips før avspark',
			text: isEnglish
				? 'Set the group order, best thirds, and the full knockout bracket before the first whistle.'
				: 'Set grupperekkjefølgje, beste trearar og heile sluttspelstreet før første avspark.'
		},
		{
			icon: Volleyball,
			title: isEnglish ? 'Match tips before every game' : 'Kamptips før kvar kamp',
			text: isEnglish
				? 'Pick the score for every match. You can change it right up until kickoff.'
				: 'Tipp resultatet for kvar kamp. Du kan endre heilt fram til avspark.'
		},
		{
			icon: Clock,
			title: isEnglish ? 'The tip locks' : 'Tipset låser seg',
			text: isEnglish
				? 'When the game starts, your tip locks and friends’ tips become visible in leagues.'
				: 'Når kampen startar, blir tipset låst, og tipsa til vener blir synlege i ligaene.'
		},
		{
			icon: Trophy,
			title: isEnglish ? 'Points along the way' : 'Poeng undervegs',
			text: isEnglish
				? 'Results, tables, and points update continuously through the group stage and knockout rounds.'
				: 'Resultat, tabellar og poeng blir oppdaterte gjennom gruppespel og sluttspel.'
		}
	]);

	let matchPoints = $derived.by(() => [
		{ label: isEnglish ? 'Correct outcome' : 'Rett utfall', value: '3', detail: isEnglish ? '1/X/2 in group stage, the team that advances in knockout' : '1/X/2 i gruppespel, laget som går vidare i sluttspel' },
		{ label: isEnglish ? 'Exact score' : 'Eksakt resultat', value: '+1', detail: isEnglish ? 'same score as the final result' : 'same resultat som sluttresultatet' },
		{ label: isEnglish ? 'Total goals' : 'Totalt mål', value: '+1', detail: isEnglish ? 'for example 2-1 and 3-0 both count as 3 goals' : 'til dømes tel både 2-1 og 3-0 som 3 mål' },
		{ label: isEnglish ? 'Correct goal difference' : 'Rett målforskjell', value: '+1', detail: isEnglish ? 'for example a one-goal win or a draw' : 'til dømes eittmålsiger eller uavgjort' }
	]);

	let forecastPoints = $derived.by(() => [
		{ label: isEnglish ? 'Correct group placement' : 'Rett gruppeplassering', value: '1' },
		{ label: isEnglish ? 'Perfect group' : 'Perfekt gruppe', value: '+2' },
		{ label: isEnglish ? 'Correct team through' : 'Rett lag vidare', value: '+1' },
		{ label: isEnglish ? 'R32 / R16 / QF' : '32-del / 16-del / kvart', value: '1 / 2 / 3' },
		{ label: isEnglish ? 'SF / Final / Winner' : 'Semi / finale / vinnar', value: '5 / 8 / 13' }
	]);

	let appFacts = $derived.by(() => [
		{ icon: Users, title: isEnglish ? 'Leagues' : 'Ligaer', text: isEnglish ? 'Create private leagues, share an invite, and follow the table together.' : 'Opprett private ligaer, del invitasjon og følg tabellen saman.' },
		{ icon: Network, title: isEnglish ? 'Tournament' : 'Turnering', text: isEnglish ? 'See groups, fixtures, and the knockout tree as the World Cup unfolds.' : 'Sjå grupper, kampar og sluttspelstreet medan VM går føre seg.' },
		{ icon: ListChecks, title: isEnglish ? 'Overview' : 'Oversikt', text: isEnglish ? 'The home page shows what is missing, the next deadline, and your standing.' : 'Framsida viser kva som manglar, neste frist og plasseringa di.' }
	]);

	function closeInfo() {
		if (browser && history.length > 1) {
			history.back();
			return;
		}
		void goto('/');
	}
</script>

<svelte:head>
	<title>{isEnglish ? 'About the game' : 'Info om spelet'} · WC Pool</title>
</svelte:head>

<div class="info-page">
	<button class="close" type="button" aria-label={isEnglish ? 'Close and go back' : 'Lukk og gå tilbake'} onclick={closeInfo}>
		<X size={18} />
		<span>{isEnglish ? 'Close' : 'Lukk'}</span>
	</button>

	<section class="hero" aria-labelledby="info-title">
		<div class="hero-copy">
			<p class="kicker">Info</p>
			<h1 id="info-title">{isEnglish ? 'How WC Pool works' : 'Slik fungerer WC Pool'}</h1>
			<p class="lead">
				{isEnglish
					? 'Pick the full World Cup before kickoff, enter match tips before every game, and compete with friends in leagues as the tournament rolls on.'
					: 'Tipp heile VM før avspark, legg inn kamptips før kvar kamp, og konkurrer med vener i ligaer gjennom turneringa.'}
			</p>
		</div>
		<div class="scoreboard" aria-label={isEnglish ? 'Quick overview' : 'Kort oversikt'}>
			<div><strong>104</strong><span>{isEnglish ? 'matches' : 'kampar'}</span></div>
			<div><strong>1</strong><span>{isEnglish ? 'Forecast' : 'VM-tips'}</span></div>
			<div><strong>6</strong><span>{isEnglish ? 'max per game' : 'maks per kamp'}</span></div>
		</div>
	</section>

	<section class="section-block" aria-labelledby="journey-title">
		<div class="section-head">
			<Info size={18} />
			<h2 id="journey-title">{isEnglish ? 'How it flows' : 'Slik går det føre seg'}</h2>
		</div>
		<div class="flow-grid">
			{#each flow as step, index}
				{@const Icon = step.icon}
				<article class="card flow-card">
					<div class="step-mark"><span>{index + 1}</span><Icon size={22} /></div>
					<h3>{step.title}</h3>
					<p>{step.text}</p>
				</article>
			{/each}
		</div>
	</section>

	<section class="section-block" aria-labelledby="app-title">
		<div class="section-head">
			<CheckCircle2 size={18} />
			<h2 id="app-title">{isEnglish ? 'The app and the game' : 'Appen og spelet'}</h2>
		</div>
		<div class="facts-grid">
			{#each appFacts as fact}
				{@const Icon = fact.icon}
				<article class="card fact-card">
					<Icon size={22} />
					<div>
						<h3>{fact.title}</h3>
						<p>{fact.text}</p>
					</div>
				</article>
			{/each}
		</div>
	</section>

	<section class="section-block scoring" aria-labelledby="score-title">
		<div class="section-head">
			<Medal size={18} />
			<h2 id="score-title">{isEnglish ? 'Scoring system' : 'Poengsystem'}</h2>
		</div>

		<div class="score-layout">
			<article class="card score-panel match-panel">
				<div class="panel-title">
					<Volleyball size={20} />
					<h3>{isEnglish ? 'Match tips' : 'Kamptips'}</h3>
				</div>
				<p>{isEnglish ? 'Max 6 points per match. In knockout, the advancing team counts as the correct outcome.' : 'Maks 6 poeng per kamp. I sluttspel tel laget som går vidare som rett utfall.'}</p>
				<div class="point-list">
					{#each matchPoints as point}
						<div class="point-row">
							<strong>{point.value}</strong>
							<div>
								<span>{point.label}</span>
								<small>{point.detail}</small>
							</div>
						</div>
					{/each}
				</div>
			</article>

			<article class="card score-panel forecast-panel">
				<div class="panel-title">
					<Telescope size={20} />
					<h3>{isEnglish ? 'Forecast' : 'VM-tips'}</h3>
				</div>
				<p>{isEnglish ? 'The Forecast locks at the first match and scores as groups and rounds are decided.' : 'VM-tipset låser seg ved første kamp og gir poeng etter kvart som grupper og rundar blir avgjorde.'}</p>
				<div class="forecast-grid">
					{#each forecastPoints as point}
						<div>
							<span>{point.label}</span>
							<strong>{point.value}</strong>
						</div>
					{/each}
				</div>
			</article>
		</div>

		<div class="card tie-break">
			<Medal size={18} />
			<p>
				{isEnglish
					? 'If points are tied, the table sorts by most exact scores, most correct winners, lowest goal-difference error, fewest submitted tips, and earliest submission.'
					: 'Ved poenglikskap blir tabellen sortert etter flest eksakte resultat, flest rette vinnarar, lågaste målforskjell-feil, færrast leverte tips og tidlegaste levering.'}
			</p>
		</div>
	</section>

	<button class="back-bottom" type="button" onclick={closeInfo}>
		<ArrowLeft size={18} />
		{isEnglish ? 'Back' : 'Tilbake'}
	</button>

	<footer class="copyright">
		<p>© 2026 Øyvind Hovden · <a href="mailto:oyvhov@gmail.com">oyvhov@gmail.com</a></p>
	</footer>
</div>

<style>
	.info-page {
		max-width: 1080px;
		margin: 0 auto;
		padding: 0 0 2rem;
	}
	:global(.info-page .card) {
		border-color: color-mix(in srgb, var(--border) 55%, transparent);
	}
	.info-page h1,
	.info-page h2,
	.info-page h3 {
		letter-spacing: 0;
	}
	.close {
		position: sticky;
		top: calc(var(--topbar-h) + 0.75rem);
		z-index: 8;
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: 0.4rem;
		width: fit-content;
		padding: 0.55rem 0.8rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		background: color-mix(in srgb, var(--surface) 86%, transparent);
		color: var(--text);
		font: inherit;
		font-weight: 800;
		box-shadow: var(--shadow-pop);
		backdrop-filter: blur(14px);
		cursor: pointer;
	}
	.hero {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 1rem;
		padding: 1.2rem 0 0.9rem;
	}
	.hero-copy {
		padding: 1rem 0 0;
	}
	.kicker {
		margin: 0 0 0.55rem;
	}
	h1 {
		font-size: 2rem;
		line-height: 1.05;
	}
	.lead {
		max-width: 680px;
		margin: 0.8rem 0 0;
		font-size: 1.02rem;
		line-height: 1.55;
		color: var(--muted);
	}
	.scoreboard {
		display: flex;
		flex-wrap: wrap;
		gap: 0.65rem;
		margin-top: 0.6rem;
	}
	.scoreboard div {
		display: inline-flex;
		align-items: center;
		gap: 0.45rem;
		padding: 0.45rem 1rem;
		background: var(--surface-2);
		border-radius: var(--radius-pill);
	}
	.scoreboard strong {
		font-size: 1.15rem;
		line-height: 1;
		color: var(--text);
	}
	.scoreboard span {
		font-size: 0.85rem;
		font-weight: 700;
		color: var(--muted);
	}
	.section-block {
		margin-top: 1.35rem;
	}
	.section-head {
		display: flex;
		align-items: center;
		gap: 0.55rem;
		margin-bottom: 0.75rem;
		color: var(--text);
	}
	:global(.section-head svg) {
		color: var(--accent-2);
	}
	.section-head h2 {
		font-size: 1.35rem;
	}
	.flow-grid,
	.facts-grid,
	.score-layout {
		display: grid;
		gap: 0.75rem;
	}
	.flow-grid :global(.card + .card),
	.facts-grid :global(.card + .card) {
		margin-top: 0;
	}
	.step-mark {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
		color: var(--accent);
	}
	.step-mark span {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		border-radius: 50%;
		background: color-mix(in srgb, var(--accent) 14%, transparent);
		font-weight: 900;
		color: var(--text);
	}
	.flow-card h3,
	.fact-card h3,
	.score-panel h3 {
		font-size: 1rem;
	}
	.flow-card p,
	.fact-card p,
	.score-panel p,
	.tie-break p {
		margin: 0.45rem 0 0;
		line-height: 1.48;
		color: var(--muted);
	}
	.fact-card {
		display: flex;
		gap: 0.75rem;
	}
	:global(.fact-card svg) {
		flex: none;
		color: var(--accent);
	}
	.panel-title {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	:global(.panel-title svg) {
		color: var(--accent-2);
	}
	.point-list {
		display: grid;
		gap: 0.6rem;
		margin-top: 0.9rem;
	}
	.point-row {
		display: grid;
		grid-template-columns: 3.25rem minmax(0, 1fr);
		gap: 0.75rem;
		align-items: center;
		padding: 0.72rem;
		border-radius: var(--radius-sm);
		background: var(--surface-2);
	}
	.point-row strong {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 3.25rem;
		height: 3.25rem;
		border-radius: 50%;
		background: var(--text);
		color: var(--bg);
		font-size: 1.05rem;
	}
	.point-row span,
	.forecast-grid span {
		display: block;
		font-weight: 900;
	}
	.point-row small {
		display: block;
		margin-top: 0.18rem;
		line-height: 1.35;
		color: var(--muted);
	}
	.forecast-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.55rem;
		margin-top: 0.9rem;
	}
	.forecast-grid div {
		min-height: 88px;
		padding: 0.75rem;
		border-radius: var(--radius-sm);
		background: var(--surface-2);
	}
	.forecast-grid strong {
		display: block;
		margin-top: 0.5rem;
		font-size: 1.35rem;
		color: var(--accent-2);
	}
	.tie-break {
		display: flex;
		gap: 0.7rem;
		align-items: flex-start;
		margin-top: 0.75rem;
		padding: 0.9rem 1rem;
	}
	:global(.tie-break svg) {
		flex: none;
		margin-top: 0.15rem;
		color: var(--gold);
	}
	.tie-break p {
		margin: 0;
	}
	.back-bottom {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		width: 100%;
		margin-top: 1.35rem;
		padding: 0.85rem 1rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-pill);
		background: var(--surface-2);
		color: var(--text);
		font: inherit;
		font-weight: 900;
		cursor: pointer;
	}
	@media (min-width: 560px) {
		.flow-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
	@media (min-width: 760px) {
		.info-page {
			padding-bottom: 3rem;
		}
		h1 {
			font-size: 2.75rem;
		}
		.hero {
			grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.8fr);
			align-items: end;
			gap: 1.5rem;
			padding-top: 2rem;
		}
		.facts-grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
		.score-layout {
			grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
		}
		.back-bottom {
			width: fit-content;
			padding-inline: 1.2rem;
		}
	}
	@media (min-width: 1020px) {
		.flow-grid {
			grid-template-columns: repeat(4, minmax(0, 1fr));
		}
	}
	@media (max-width: 759px) {
		.forecast-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	.copyright {
		margin-top: 2rem;
		text-align: center;
		font-size: 0.82rem;
		color: var(--muted);
	}
	.copyright a {
		color: var(--muted);
		text-decoration: underline;
		text-underline-offset: 3px;
	}
</style>