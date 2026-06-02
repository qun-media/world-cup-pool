<script lang="ts">
	import Flag from '$lib/components/Flag.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { fly } from 'svelte/transition';
	import {
		ArrowRight,
		Clock,
		MessageCircle,
		ShieldCheck,
		Trophy,
		Users
	} from '@lucide/svelte';

	let targetIndex = $state(0);
	const landingTargets = ['coworker', 'friend'];
	const landingVerb = 'Log in and beat your';
	const landingTarget = $derived(
		landingTargets[targetIndex] ?? landingTargets[0] ?? ''
	);

	$effect(() => {
		targetIndex = 0;
	});

	$effect(() => {
		const words = landingTargets;
		if (words.length < 2) return;
		const timer = setInterval(() => {
			targetIndex = (targetIndex + 1) % words.length;
		}, 2200);
		return () => clearInterval(timer);
	});

</script>

<svelte:head>
	<title>WC Pool</title>
	<meta
		name="description"
		content="Join friends for World Cup match tips, leagues, points and chat."
	/>
</svelte:head>

<section class="public-landing" aria-labelledby="landing-title">
	<div class="landing-shell">
		<section class="landing-hero">
			<div class="hero-copy">
				<Logo variant="hero" tagline="Forecast with friends" />
				<h1 id="landing-title" class="landing-headline">
					<span class="landing-verb">{landingVerb}</span>
					<span class="landing-target-slot">
						{#key `en-${landingTarget}`}
							<span
								class="landing-target"
								in:fly={{ y: 18, duration: 220, opacity: 0.15 }}
								out:fly={{ y: -18, duration: 180, opacity: 0.05 }}
							>
								{landingTarget}
							</span>
						{/key}
					</span>
				</h1>
				<p class="lead">
					Tip every match, build your World Cup bracket, and follow the league drama as the points land.
				</p>

				<div class="hero-actions" aria-label="Sign in actions">
					<div class="secondary-actions">
						<a class="btn secondary" href="/login">Use email</a>
						<a class="btn ghost" href="/register">Create account</a>
					</div>
				</div>
			</div>
		</section>

		<section class="showcase" aria-labelledby="showcase-title">
			<div class="section-head">
				<p class="kicker">Inside the app</p>
				<h2 id="showcase-title">Everything feels like match day.</h2>
			</div>

			<div class="mock-grid">
				<article class="card mock-card table-card">
					<div class="mock-head"><Trophy size={18} /><h3>Group table</h3></div>
					<table>
						<thead><tr><th>#</th><th>Team</th><th>P</th><th>GD</th><th>Pts</th></tr></thead>
						<tbody>
							<tr><td>1</td><td><span class="team-cell"><Flag iso2="nl" code="NED" size={17} /> Netherlands</span></td><td>3</td><td>+5</td><td>7</td></tr>
							<tr><td>2</td><td><span class="team-cell"><Flag iso2="se" code="SWE" size={17} /> Sweden</span></td><td>3</td><td>+1</td><td>5</td></tr>
							<tr><td>3</td><td><span class="team-cell"><Flag iso2="jp" code="JPN" size={17} /> Japan</span></td><td>3</td><td>-2</td><td>3</td></tr>
							<tr><td>4</td><td><span class="team-cell"><Flag iso2="tn" code="TUN" size={17} /> Tunisia</span></td><td>3</td><td>-4</td><td>1</td></tr>
						</tbody>
					</table>
				</article>

				<article class="card mock-card score-card">
					<div class="mock-head"><ShieldCheck size={18} /><h3>Points system</h3></div>
					<div class="score-total"><strong class="digits">6</strong><span>max per match</span></div>
					<ul>
						<li><span>Correct outcome</span><b>3 p</b></li>
						<li><span>Exact score</span><b>+1 p</b></li>
						<li><span>Total goals</span><b>+1 p</b></li>
						<li><span>Goal difference</span><b>+1 p</b></li>
					</ul>
				</article>

				<article class="card mock-card chat-card">
					<div class="mock-head"><MessageCircle size={18} /><h3>League chat</h3></div>
					<div class="bubble theirs"><b>Anna</b><span>That 90th minute goal changed everything.</span></div>
					<div class="bubble mine"><b>You</b><span>I had 2-1. Six points!</span></div>
					<div class="chat-meta"><Users size={15} /> Private leagues, live reactions</div>
				</article>
			</div>
		</section>

		<section class="bottom-cta card">
			<div>
				<p class="kicker">Ready before kickoff</p>
				<h2>Make the first pick now.</h2>
			</div>
			<a class="btn" href="/login"><Clock size={16} /> Log in <ArrowRight size={16} /></a>
		</section>
	</div>
</section>

<style>
	:global(.app-shell.public-shell) {
		max-width: none;
		padding-top: 0;
		padding-inline: 0;
		padding-bottom: 0;
	}
	.public-landing {
		--accent: #8fc58f;
		--accent-2: #d8b86c;
		--bg: #071019;
		--surface: #0b171f;
		--surface-2: #10242b;
		--surface-3: #18343a;
		--border: rgba(214, 190, 128, 0.12);
		--border-strong: rgba(214, 190, 128, 0.22);
		--text: #f3f6ee;
		--muted: #a7b7ae;
		--gold: #d9bb72;
		min-height: 100dvh;
		overflow-x: clip;
		color: var(--text);
		background:
			radial-gradient(80% 42% at 50% 0%, rgba(217, 187, 114, 0.12), transparent 62%),
			radial-gradient(52% 32% at 80% 12%, rgba(143, 197, 143, 0.13), transparent 70%),
			linear-gradient(180deg, #081824 0%, #071019 52%, #050a0f 100%);
	}
	.landing-shell {
		width: 100%;
		max-width: 1120px;
		margin: 0 auto;
		padding: clamp(1rem, 4vw, 2rem);
	}
	.landing-hero {
		display: grid;
		gap: 1rem;
		padding: min(8dvh, 4rem) 0 1.5rem;
		min-width: 0;
	}
	.hero-copy {
		display: grid;
		gap: 0.8rem;
		min-width: 0;
	}
	h1 {
		max-width: 10ch;
		font-size: clamp(2.45rem, 15vw, 5.8rem);
		line-height: 0.92;
		letter-spacing: 0;
	}
	.landing-headline {
		display: grid;
		gap: 0.08em;
		max-width: 11ch;
		min-width: 0;
	}
	.landing-verb {
		color: color-mix(in srgb, var(--text) 84%, var(--muted));
		font-weight: 500;
		letter-spacing: -0.02em;
	}
	.landing-target-slot {
		display: grid;
		min-height: 1.02em;
		min-width: 0;
		width: 100%;
	}
	.landing-target {
		grid-area: 1 / 1;
		font-weight: 800;
		background: linear-gradient(110deg, #f3dfae 8%, var(--gold) 44%, var(--accent) 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
		will-change: transform, opacity;
	}
	.lead {
		max-width: 52ch;
		min-width: 0;
		margin: 0;
		color: var(--muted);
		font-size: 1.03rem;
		line-height: 1.55;
	}
	.hero-actions {
		display: grid;
		gap: 0.75rem;
		max-width: 430px;
		min-width: 0;
		margin-top: 0.35rem;
	}
	.secondary-actions {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.55rem;
	}
	.secondary-actions .btn {
		min-height: 44px;
		padding: 0.75rem 0.8rem;
	}
	.bottom-cta,
	.mock-card {
		background:
			radial-gradient(circle at 18% 0%, rgba(217, 187, 114, 0.08), transparent 32%),
			linear-gradient(180deg, rgba(13, 34, 40, 0.95), rgba(8, 21, 30, 0.98));
		border-color: var(--border-strong);
	}
	.mock-head,
	.bottom-cta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		min-width: 0;
		flex-wrap: wrap;
	}
	.chat-meta {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		min-width: 0;
		flex-wrap: wrap;
		color: var(--accent);
		font-size: 0.8rem;
		font-weight: 800;
	}
	.showcase {
		display: grid;
		gap: 1rem;
		padding: 1.2rem 0 2rem;
		min-width: 0;
	}
	.section-head {
		display: grid;
		gap: 0.25rem;
		min-width: 0;
	}
	.section-head h2,
	.bottom-cta h2 {
		font-size: clamp(1.4rem, 7vw, 2.15rem);
		letter-spacing: 0;
	}
	.mock-grid {
		display: grid;
		gap: 0.85rem;
		min-width: 0;
	}
	.mock-grid .card + .card {
		margin-top: 0;
	}
	.mock-card {
		display: grid;
		gap: 0.9rem;
		min-width: 0;
	}
	.mock-head {
		justify-content: flex-start;
		color: var(--gold);
	}
	.mock-head h3 {
		font-size: 1.05rem;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.86rem;
	}
	th,
	td {
		padding: 0.48rem 0.2rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.07);
		text-align: right;
		overflow-wrap: anywhere;
		vertical-align: middle;
	}
	th:nth-child(2),
	td:nth-child(2) {
		text-align: left;
	}
	.table-card table {
		table-layout: auto;
	}
	.table-card th,
	.table-card td {
		white-space: nowrap;
		overflow-wrap: normal;
	}
	.table-card th:nth-child(1),
	.table-card td:nth-child(1),
	.table-card th:nth-child(3),
	.table-card td:nth-child(3),
	.table-card th:nth-child(4),
	.table-card td:nth-child(4),
	.table-card th:nth-child(5),
	.table-card td:nth-child(5) {
		width: 2.4rem;
	}
	td:nth-child(2) {
		font-weight: 800;
	}
	.team-cell {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
	}
	.score-total {
		display: flex;
		align-items: baseline;
		gap: 0.55rem;
	}
	.score-total strong {
		font-size: 3rem;
		color: var(--accent);
	}
	.score-card ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		gap: 0.45rem;
	}
	.score-card li {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		min-width: 0;
		flex-wrap: wrap;
		padding: 0.55rem 0;
		border-bottom: 1px solid rgba(255, 255, 255, 0.07);
	}
	.score-card b {
		color: var(--accent);
		font-family: var(--font-mono);
	}
	.bubble {
		display: grid;
		gap: 0.2rem;
		max-width: 82%;
		padding: 0.75rem;
		border-radius: 16px;
		background: rgba(255, 255, 255, 0.07);
	}
	.bubble.mine {
		justify-self: end;
		background: color-mix(in srgb, var(--accent) 18%, transparent);
	}
	.bubble b {
		font-size: 0.8rem;
		color: var(--gold);
	}
	.chat-meta {
		color: var(--muted);
	}
	.bottom-cta {
		align-items: stretch;
		margin-bottom: 2rem;
		min-width: 0;
	}
	.bottom-cta .btn {
		width: auto;
		min-width: 190px;
	}

	@media (min-width: 780px) {
		.mock-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (min-width: 980px) {
		.mock-grid {
			grid-template-columns: 1.05fr 0.85fr 1fr;
		}
	}

	@media (min-width: 521px) and (max-width: 1099px) {
		h1 {
			font-size: clamp(2.15rem, 8.6vw, 4rem);
			max-width: 10.8ch;
		}
		.lead {
			font-size: 0.98rem;
			max-width: 46ch;
		}
		.section-head h2,
		.bottom-cta h2 {
			font-size: clamp(1.25rem, 4vw, 1.8rem);
		}
		.mock-head h3 {
			font-size: 1rem;
		}
		table {
			font-size: 0.82rem;
		}
	}

	@media (min-width: 1100px) {
		h1 {
			font-size: clamp(2.85rem, 5vw, 4.95rem);
			max-width: 11.5ch;
		}
		.lead {
			font-size: 0.98rem;
			max-width: 48ch;
		}
		.section-head h2,
		.bottom-cta h2 {
			font-size: clamp(1.3rem, 2.4vw, 1.95rem);
		}
		.mock-head h3 {
			font-size: 1rem;
		}
		table {
			font-size: 0.82rem;
		}
	}

	@media (max-width: 520px) {
		.landing-shell {
			padding-top: 4.6rem;
		}
		h1 {
			font-size: clamp(2rem, 11vw, 3.1rem);
		}
		.landing-headline,
		.bottom-cta > div {
			max-width: 100%;
			min-width: 0;
		}
		.secondary-actions {
			grid-template-columns: 1fr;
		}
		.bottom-cta {
			flex-direction: column;
		}
		.bottom-cta .btn {
			width: 100%;
		}
	}
</style>