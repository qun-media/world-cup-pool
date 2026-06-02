import { describe, expect, it } from 'vitest';
import type { LeagueSummary } from './api';
import type { Match, Team } from './tips.svelte';
import { normalizeSearchText, searchApp } from './search';

const teams: Record<string, Team> = {
	norway: { id: 'norway', name: 'Norway', iso2: 'NO', fifaCode: 'NOR', fifaRanking: 0 },
	brazil: { id: 'brazil', name: 'Brazil', iso2: 'BR', fifaCode: 'BRA', fifaRanking: 0 },
	southKorea: { id: 'southKorea', name: 'South Korea', iso2: 'KR', fifaCode: 'KOR', fifaRanking: 0 },
	spain: { id: 'spain', name: 'Spain', iso2: 'ES', fifaCode: 'ESP', fifaRanking: 0 }
};

function match(partial: Partial<Match>): Match {
	return {
		id: 'm1',
		stage: 'group',
		groupLetter: 'A',
		roundLabel: 'Round 1',
		num: 1,
		kickoff: '2026-06-12T18:00:00Z',
		tvChannel: 'BBC1',
		status: 'scheduled',
		homeTeam: 'norway',
		awayTeam: 'brazil',
		homeLabel: '',
		awayLabel: '',
		ftHome: 0,
		ftAway: 0,
		etHome: 0,
		etAway: 0,
		penHome: 0,
		penAway: 0,
		advancer: '',
		finalizedAt: '',
		...partial
	};
}

const leagues: LeagueSummary[] = [
	{ id: 'l1', name: 'Family League', inviteCode: 'FAM123', role: 'owner', members: 8, hideForecast: false, isAdmin: true },
	{ id: 'l2', name: 'Work WC', inviteCode: 'WORK', role: 'member', members: 12, hideForecast: false, isAdmin: false }
];

describe('normalizeSearchText', () => {
	it('normalizes case and accents', () => {
		expect(normalizeSearchText('  SØR-Koréa  ')).toBe('sor korea');
		expect(normalizeSearchText('Æ Å Ø')).toBe('ae a o');
	});
});

describe('searchApp', () => {
	it('returns empty groups for empty queries', () => {
		expect(searchApp('', { matches: [], teams, leagues })).toEqual({
			matches: [],
			teams: [],
			groups: [],
			leagues: []
		});
	});

	it('matches fixtures by team names and builds a tips link', () => {
		const results = searchApp('brazil', {
			matches: [match({ id: 'm-brazil' })],
			teams,
			leagues: []
		});

		expect(results.matches[0]).toMatchObject({
			id: 'm-brazil',
			title: 'Norway - Brazil',
			href: '/tips?match=m-brazil'
		});
	});

	it('matches teams by name', () => {
		const results = searchApp('south korea', { matches: [], teams, leagues: [] });

		expect(results.teams[0]).toMatchObject({
			id: 'southKorea',
			title: 'South Korea',
			href: '/tips?team=southKorea'
		});
	});

	it('matches user leagues and links to the league page', () => {
		const results = searchApp('family', { matches: [], teams: {}, leagues });

		expect(results.leagues[0]).toMatchObject({
			id: 'l1',
			title: 'Family League',
			href: '/leagues/l1'
		});
	});

	it('matches groups and links to the group section on tips', () => {
		const results = searchApp('group a', {
			matches: [
				match({ id: 'm-a1', homeTeam: 'norway', awayTeam: 'brazil', groupLetter: 'A' }),
				match({ id: 'm-a2', homeTeam: 'southKorea', awayTeam: 'spain', groupLetter: 'A' })
			],
			teams,
			leagues: []
		});

		expect(results.groups[0]).toMatchObject({
			id: 'A',
			title: 'Group A',
			href: '/tips?group=A'
		});
	});

	it('limits each result group', () => {
		const manyMatches = Array.from({ length: 6 }, (_, index) =>
			match({ id: `m${index}`, num: index + 1 })
		);

		const results = searchApp('norway', { matches: manyMatches, teams, leagues }, 3);

		expect(results.matches).toHaveLength(3);
	});
});
